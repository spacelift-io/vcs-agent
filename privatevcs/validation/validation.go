package validation

import (
	"fmt"
	"net/http"
)

// ErrNoMatch is returned when the request didn't match any planned API usage.
var ErrNoMatch = fmt.Errorf("no match for request")

var vendorMatchers = map[string]func(r *http.Request) (name string, project string, subdomain *string, err error){
	"azure_devops":         matchAzureDevOpsRequest,
	"bitbucket_datacenter": matchBitbucketDatacenterRequest,
	"github_enterprise":    matchGitHubEnterpriseRequest,
	"gitlab":               matchGitLabRequest,
}

// MatchRequest matches the request based on the VCS vendor.
// It returns the API usage human-friendly name, as well as the target project.
// If the request isn't scope to a project (i.e. listing projects) it returns an empty string.
func MatchRequest(vendor string, r *http.Request) (name string, project string, subdomain *string, err error) {
	return vendorMatchers[vendor](r)
}

// MatchAnyVendorRequest works like MatchRequest, but tries to match all available vendors.
func MatchAnyVendorRequest(r *http.Request) (name string, project string, subdomain *string, err error) {
	for _, matcher := range vendorMatchers {
		if name, project, updatedHostname, err := matcher(r); err == nil {
			return name, project, updatedHostname, nil
		}
	}
	return "", "", nil, ErrNoMatch
}
