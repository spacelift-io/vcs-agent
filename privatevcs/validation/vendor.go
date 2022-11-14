package validation

// Vendor represents one of the supported VCS vendors.
type Vendor string

const (
	// AzureDevOps represents Azure DevOps Repos VCS vendor.
	AzureDevOps Vendor = "azure_devops"

	// BitbucketDatacenter represents Bitbucket Datacenter VCS vendor.
	BitbucketDatacenter Vendor = "bitbucket_datacenter"

	// GitHubEnterprise represents self-hosted GitHub Enterprise VCS vendor.
	GitHubEnterprise Vendor = "github_enterprise"

	// GitLab represents self-hosted GitLab VCS vendor.
	GitLab Vendor = "gitlab"
)
