FROM alpine:latest

COPY sagan .
ENTRYPOINT ["/sagan"]
