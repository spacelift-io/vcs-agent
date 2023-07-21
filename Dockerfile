FROM alpine:3.18


RUN apk add --no-cache ca-certificates
RUN apk upgrade --update-cache --available
RUN adduser --disabled-password --no-create-home --uid=1983 spacelift

# The reason we're using a wildcard on the copy is that goreleaser sets a _v1 suffix for the
# amd64 target, which isn't included in the docker TARGETARCH variable
COPY build/spacelift-vcs-agent /usr/bin/spacelift-vcs-agent
RUN chmod +x /usr/bin/spacelift-vcs-agent

CMD ["/usr/bin/spacelift-vcs-agent", "serve"]