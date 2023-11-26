# Spacelift VCS Agent

[![Publish](https://github.com/spacelift-io/vcs-agent/actions/workflows/deployment.yml/badge.svg?branch=main)](https://github.com/spacelift-io/vcs-agent/actions/workflows/deployment.yml)

The Spacelift VCS Agent provides a gateway to allow Spacelift to access VCS systems
that are not available via the public internet.

For more information visit <https://docs.spacelift.io/concepts/vcs-agent-pools>.

## ✨ Usage

You can either download the binary and run it directly or use the Docker image.

### Downloading the binary

The binary can be downloaded directly from the [Releases](https://github.com/spacelift-io/vcs-agent/releases) section or from Spacelift's CDN:

| URL                                                          | Architecture  |
| ------------------------------------------------------------ | ------------- |
| <https://downloads.spacelift.io/spacelift-vcs-agent-x86_64>  | Linux (amd64) |
| <https://downloads.spacelift.io/spacelift-vcs-agent-aarch64> | Linux (arm64) |

### Running via Docker

Use the following command to run the VCS Agent [via Docker](https://gallery.ecr.aws/spacelift/vcs-agent):

```shell
docker run -it --rm -e "SPACELIFT_VCS_AGENT_POOL_TOKEN=<VCS Token>" \
  -e "SPACELIFT_VCS_AGENT_TARGET_BASE_ENDPOINT=<https://github.mycompany.com>" \
  -e "SPACELIFT_VCS_AGENT_VENDOR=<bitbucket_datacenter>" \
  public.ecr.aws/spacelift/vcs-agent
```

To use this example, make sure to update the environment variables according to your environment as explained in the table below.

> If you want to pin the version of the VCS Agent, you can use the `public.ecr.aws/spacelift/vcs-agent:<tag>` image instead.

### Configuration

The configuration can be either provided as a command line argument or as an environment variable.

The VCS Agent requires the following settings to be configured to work:

| Command line flag        | Environment varariable                   | Description                                                                                                                  |
| ------------------------ | ---------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| `--token`                | SPACELIFT_VCS_AGENT_POOL_TOKEN           | The token downloaded from Spacelift when creating the pool. You can decode this from base64 to edit the VCS Gateway address. |
| `--target-base-endpoint` | SPACELIFT_VCS_AGENT_TARGET_BASE_ENDPOINT | The base endpoint address for the VCS integration. For example `https://github.mycompany.com`.                               |
| `--vendor`               | SPACELIFT_VCS_AGENT_VENDOR               | The VCS vendor to use. Possible values: `azure_devops`, `bitbucket_datacenter`, `github_enterprise` and `gitlab`.            |

In addition, when running locally, you may want to set `SPACELIFT_VCS_AGENT_DIAL_INSECURE=true`
to enable the VCS Agent to communicate with a Gateway instance that isn't using TLS.

Run the `spacelift-vcs-agent --help` command to see all available options.

### 🛠 Contributing

#### Running in VS Code

- Make a copy of the `.env.template` file and call it `.env` (see [here](#environment-variables)
  for more information on how to configure your environment variables).
- Press F5 / run the _Launch Package_ configuration.
  
#### Release (for maintainers)

Once you're ready to release a new version, bump a semver and push it:

```shell
git tag -a -m "Release v1.0.0" v1.0.0
git push origin v1.0.0
```
