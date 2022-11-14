package allowlist

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/spacelift-io/vcs-agent/nullable"
)

func TestGitHubEnterpriseValidation(t *testing.T) {
	testCases := []struct {
		path               string
		matches            bool
		name, method       string
		project, subdomain *string
	}{
		{
			path:    "/api/v3/repos/octocats/infra/compare/abc...123",
			matches: true,
			name:    "Compare Trees",
			project: nullable.String("octocats/infra"),
			method:  http.MethodGet,
		},
		{
			path:    "/repos/octocats/infra/compare/abc...123",
			matches: true,
			name:    "Compare Trees",
			project: nullable.String("octocats/infra"),
			method:  http.MethodGet,
		},
		{
			path:    "/api/v3/repos/octocats/infra/statuses/abc123",
			matches: true,
			name:    "Create Commit Status",
			project: nullable.String("octocats/infra"),
			method:  http.MethodPost,
		},
		{
			path:    "/repos/octocats/infra/statuses/abc123",
			matches: true,
			name:    "Create Commit Status",
			project: nullable.String("octocats/infra"),
			method:  http.MethodPost,
		},
		{
			path:    "/api/v3/repos/octocats/infra/check-runs",
			matches: true,
			name:    "Create Check Run",
			project: nullable.String("octocats/infra"),
			method:  http.MethodPost,
		},
		{
			path:    "/repos/octocats/infra/check-runs",
			matches: true,
			name:    "Create Check Run",
			project: nullable.String("octocats/infra"),
			method:  http.MethodPost,
		},
		{
			path:    "/api/v3/repos/octocats/infra/deployments",
			matches: true,
			name:    "Create Deployment",
			project: nullable.String("octocats/infra"),
			method:  http.MethodPost,
		},
		{
			path:    "/repos/octocats/infra/deployments",
			matches: true,
			name:    "Create Deployment",
			project: nullable.String("octocats/infra"),
			method:  http.MethodPost,
		},
		{
			path:    "/api/v3/repos/octocats/infra/deployments/123456/statuses",
			matches: true,
			name:    "Create Deployment Status",
			project: nullable.String("octocats/infra"),
			method:  http.MethodPost,
		},
		{
			path:    "/repos/octocats/infra/deployments/123456/statuses",
			matches: true,
			name:    "Create Deployment Status",
			project: nullable.String("octocats/infra"),
			method:  http.MethodPost,
		},
		{
			path:    "/api/v3/repos/octocats/infra/deployments/123456",
			matches: true,
			name:    "Delete Deployment",
			project: nullable.String("octocats/infra"),
			method:  http.MethodDelete,
		},
		{
			path:    "/repos/octocats/infra/deployments/123456",
			matches: true,
			name:    "Delete Deployment",
			project: nullable.String("octocats/infra"),
			method:  http.MethodDelete,
		},
		{
			path:    "/api/v3/repos/octocats/infra/commits/565958b65e14a5e06c1c467a66b446f2afcf87ef",
			matches: true,
			name:    "Get Individual Commit Details",
			project: nullable.String("octocats/infra"),
			method:  http.MethodGet,
		},
		{
			path:    "/repos/octocats/infra/commits/565958b65e14a5e06c1c467a66b446f2afcf87ef",
			matches: true,
			name:    "Get Individual Commit Details",
			project: nullable.String("octocats/infra"),
			method:  http.MethodGet,
		},
		{
			path:    "/_codeload/octocats/infra/legacy.tar.gz/565958b65e14a5e06c1c467a66b446f2afcf87ef",
			matches: true,
			name:    "Get Repository Tarball",
			project: nullable.String("octocats/infra"),
			method:  http.MethodGet,
		},
		{
			path:    "/codeload/octocats/infra/legacy.tar.gz/565958b65e14a5e06c1c467a66b446f2afcf87ef",
			matches: true,
			name:    "Get Repository Tarball",
			project: nullable.String("octocats/infra"),
			method:  http.MethodGet,
		},
		{
			path:      "/octocats/infra/legacy.tar.gz/565958b65e14a5e06c1c467a66b446f2afcf87ef",
			matches:   true,
			name:      "Get Repository Tarball",
			project:   nullable.String("octocats/infra"),
			method:    http.MethodGet,
			subdomain: nullable.String("codeload"),
		},
		{
			path:    "/api/graphql",
			matches: true,
			name:    "GraphQL Endpoint",
			method:  http.MethodPost,
		},
		{
			path:    "/api/v3/repos/octocats/infra/deployments",
			matches: true,
			name:    "List Deployments",
			project: nullable.String("octocats/infra"),
			method:  http.MethodGet,
		},
		{
			path:    "/repos/octocats/infra/deployments",
			matches: true,
			name:    "List Deployments",
			project: nullable.String("octocats/infra"),
			method:  http.MethodGet,
		},
		{
			path:    "/api/v3/repos/octocats/infra/pulls/123/files",
			matches: true,
			name:    "List Pull Request Files",
			project: nullable.String("octocats/infra"),
			method:  http.MethodGet,
		},
		{
			path:    "/repos/octocats/infra/pulls/123/files",
			matches: true,
			name:    "List Pull Request Files",
			project: nullable.String("octocats/infra"),
			method:  http.MethodGet,
		},
		{
			path:    "/api/v3/repos/octocats/infra/pulls/123",
			matches: true,
			name:    "Get Pull Request",
			project: nullable.String("octocats/infra"),
			method:  http.MethodGet,
		},
		{
			path:    "/repos/octocats/infra/pulls/123",
			matches: true,
			name:    "Get Pull Request",
			project: nullable.String("octocats/infra"),
			method:  http.MethodGet,
		},
		{
			path:    "/api/v3/app/installations/29/access_tokens",
			matches: true,
			name:    "Refresh Access Token",
			method:  http.MethodPost,
		},
		{
			path:    "/app/installations/29/access_tokens",
			matches: true,
			name:    "Refresh Access Token",
			method:  http.MethodPost,
		},
		{
			path:    "/api/v3/app/installations",
			matches: true,
			name:    "List Installations",
			method:  http.MethodGet,
		},
		{
			path:    "/app/installations",
			matches: true,
			name:    "List Installations",
			method:  http.MethodGet,
		},
		{
			path:    "/api/v3/app",
			matches: true,
			name:    "Get app details",
			method:  http.MethodGet,
		},
		{
			path:    "/app",
			matches: true,
			name:    "Get app details",
			method:  http.MethodGet,
		},
	}

	for i := range testCases {
		testCase := testCases[i]
		request, err := http.NewRequest(testCase.method, "https://github.myorg.com"+testCase.path, nil)
		require.NoError(t, err, "could not create request")

		name, project, err := matchGitHubEnterpriseRequest(request)

		if testCase.matches {
			require.NoError(t, err, "could not find match for %q", testCase.name)
			require.Equal(t, testCase.name, name)

			if testCase.project != nil {
				require.Equal(t, *testCase.project, project, "project did not match for %q", testCase.name)
			}
		} else {
			require.ErrorIs(t, ErrNoMatch, err)
		}
	}
}
