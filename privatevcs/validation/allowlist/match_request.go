package allowlist

import (
	"fmt"
	"net/http"

	"github.com/spacelift-io/vcs-agent/privatevcs/validation"
)

// ErrNoMatch is returned when the request didn't match any planned API usage.
var ErrNoMatch = fmt.Errorf("vcs-agent: no match for request")

var vendorMatchers = map[validation.Vendor]func(r *http.Request) (name string, project string, err error){
	validation.AzureDevOps:         matchAzureDevOpsRequest,
	validation.BitbucketDatacenter: matchBitbucketDatacenterRequest,
	validation.GitHubEnterprise:    matchGitHubEnterpriseRequest,
	validation.GitLab:              matchGitLabRequest,
}

// MatchRequest matches the request based on the VCS vendor.
// It returns the API usage human-friendly name, as well as the target project.
// If the request isn't scope to a project (i.e. listing projects) it returns an empty string.
func MatchRequest(vendor validation.Vendor, r *http.Request) (name string, project string, err error) {
	return vendorMatchers[vendor](r)
}
