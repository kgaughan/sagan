# Sagan

Sagan is a tool for helping run collections of Terraform projects. It aims to make the deployment and management of multiple projects significantly easier and faster. It's intended to supplement tools like [Atlantis](https://www.runatlantis.io/), especially in circumstances where they may not be feasible to use, such as when initially deploying a set of modules when doing an initial buildout.

Behind the scenes, it can manage tunnels, the automatic generation of credentials, &c. It also supports dependencies, rebuilds based on updates, and parallelism.

## Configuration file format

A Sagan configuration file is a YAML file that contains _helpers_, _workflows_, and _projects_.

A _helper_ is a task that runs at the start and end of a workflow. This can be a script that manages a tunnel, fetches some credentials to be used as part of a workflow, or any number of other tasks.

A _workflow_ is a list of commands that implements the various deployment phases of a project. In Terraform, this would be initialising a project, planning it, applying it, and cleanup.

A _project_ is a directory full of configuration, typically Terraform configuration, that a workflow describes how to manage.

```yaml
---
sagan:
  version: "1.0"
  helpers:
    sshuttle:
      # Helpers of type 'daemon' persist until they are no longer needed. The
      # duration of their persistence is decided by the values of the arguments
      # they expect. As long as there is at least one project that expects an
      # instance of the helper with a particular set of arguments, the helper
      # will keep running.
      type: daemon
      args:
        - name: cidr
          # This argument has a default value
          default: 192.168.0.0/16
          # This argument is used to prevent multiple instances of helper
          # running at once. This is useful for circumstances where you may,
          # for instances, reuse a private range in multiple environments and
          # want to avoid confusing sshuttle.
          exclusive: true
        - name: environment
        # The value of this argument is written to an environment variable
        - name: profile
          env: AWS_PROFILE
        - region: region
          # This argument also has a default 
          default: us-east-1
          env: REGION
      # Only the last command is kept running in the backgroun until it exits.
      # By default, it will send a SIGTERM to the last command upon shutdown.
      run:
        - cmd: aws ec2 describe-instances --filter "Name=tag:Env,Values=$environment" "Name=tag:Role,Values=bastion" --query "Reservations[].Instances[].InstanceId | [0]" --output text
          save_as: instance_id
        - cmd: aws ssm start-session --target $instance_id
        - cmd: sshuttle --ssh-cmd="ssh -o ProxyCommand='aws ssm start-session --target %h --document-name AWS-StartSSHSession --parameters portNumber=22'" --remote ec2-user@$instance_id $cidr
    vault:
      # A one-shot helper that may prompt for input.
      type: interactive
      # Is there a helper that needs to be running before this helper can be
      # used?
      requires:
        - sshuttle
      args:
        - name: environment
      run:
        - cmd: vault login -method=ldap -address=https://vault.$environment.infra.example.com -no-store
          save_as: VAULT_TOKEN
      # If this command will need to be re-executed after a while, how long
      # should the result be valid for?
      ttl: 2h
  workflows:
    default:
      temporaries:
        - name: plan
          type: file
      load:
        # This is basically the default behaviour i nothing
        - "*.auto.tfvars.json"
        - "terraform.tfvars.json"
      init:
        run:
          - cmd: terraform init
      plan:
        run:
          - cmd: terraform plan -plan $plan
        requires:
          - ".terraform": init
      apply:
        requires:
          - "$plan"
        run:
          - cmd: terraform apply $plan
      cleanup:
        run:
          - rm -rf .terraform $plan
  projects:
    # Every project has a path. The last element in the path is used as the
    # default project name.
    - path: fred
      # If you want to override the project name, you do it with 'name'
      name: frederick
      # The workflow to use.
      workflow: default
      helpers:
        # Will implicitly run the 'sshuttle' helper too.
        - vault
    - path: barney
      workflow: default
      # When rolling out this set of projects, 'frederick' must be deployed
      # before 'barney'
      requires:
        - frederick
      helpers:
        # This project doesn't require Vault, but it does require a tunnel.
        - sshuttle
      # We want to save some of this project's outputs to a configuration
      # file for the 'bamm-bamm' project.
      outputs:
        - write: ../bamm-bamm/terraform.tfvars.json
          # Overwrite the value of the field. Other actions are 'add'.
          action: replace
          # The name of the field to overwrite.
          name: thingy
    - path: betty
      workflow: default
    - path: bamm-bamm
      requires:
        - betty
      # This creates an indirect dependency on 'barney': if the field 'thingy'
      # in the named file changes, this project should be redeployed.
      redeploy_on:
        - file: terraform.tfvars.json
          name: thingy
```
