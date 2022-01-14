# Spacelift VCS Agent

The Spacelift VCS Agent provides a gateway to allow Spacelift to access VCS systems
that are not available via the public internet.

For more information see <https://docs.spacelift.io/concepts/vcs-agent-pools>.

## Running in vscode

- Make a copy of the `.env.template` file and call it `.env` (see [here](#environment-variables)
  for more information on how to configure your environment variables).
- Press F5 / run the _Launch Package_ configuration.

## Environment Variables

The VCS Agent requires the following environment variables to be configured to work:

| Name                                     | Description                                                                                                                  |
| ---------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| SPACELIFT_VCS_AGENT_POOL_TOKEN           | The token downloaded from Spacelift when creating the pool. You can decode this from base64 to edit the VCS Gateway address. |
| SPACELIFT_VCS_AGENT_TARGET_BASE_ENDPOINT | The base endpoint address for the VCS integration. For example `https://github.mycompany.com`.                               |
| SPACELIFT_VCS_AGENT_VENDOR               | The VCS vendor to use. For example `bitbucket_datacenter`                                                                    |

In addition, when running locally, you may want to set `SPACELIFT_VCS_AGENT_DIAL_INSECURE=true`
to enable the VCS Agent to communicate with a Gateway instance that isn't using TLS.
