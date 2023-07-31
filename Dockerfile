FROM alpine:3.18


RUN apk add --no-cache ca-certificates
RUN apk upgrade --update-cache --available
RUN adduser --disabled-password --no-create-home --uid=1983 spacelift

COPY build/spacelift-vcs-agent /usr/bin/spacelift-vcs-agent

RUN chmod +x /usr/bin/spacelift-vcs-agent

CMD ["/usr/bin/spacelift-vcs-agent", "serve"]

USER spacelift