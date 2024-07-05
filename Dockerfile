FROM alpine:3.20

RUN apk add --no-cache ca-certificates && apk upgrade --update-cache --available
RUN adduser --disabled-password --no-create-home --uid=1983 spacelift

COPY spacelift-vcs-agent /usr/bin/spacelift-vcs-agent

RUN chmod +x /usr/bin/spacelift-vcs-agent

USER spacelift

CMD ["/usr/bin/spacelift-vcs-agent", "serve"]
