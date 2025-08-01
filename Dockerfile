FROM alpine:3.22

RUN apk upgrade --update-cache --available && apk add ca-certificates && rm -rf /var/cache/apk/*
RUN adduser --disabled-password --no-create-home --uid=1983 spacelift

COPY spacelift-vcs-agent /usr/bin/spacelift-vcs-agent

RUN chmod +x /usr/bin/spacelift-vcs-agent

USER spacelift

CMD ["/usr/bin/spacelift-vcs-agent", "serve"]
