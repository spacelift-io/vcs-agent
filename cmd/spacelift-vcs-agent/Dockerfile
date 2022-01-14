FROM ubuntu
RUN apt update && apt install -y net-tools ca-certificates
ADD spacelift-vcs-agent spacelift-vcs-agent
CMD ["./spacelift-vcs-agent"]
