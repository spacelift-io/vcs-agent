package allowlist

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/spacelift-io/vcs-agent/nullable"
)

func TestAzureDevOpsValidation(t *testing.T) {
	type azureDevOpsTestCase struct {
		path         string
		matches      bool
		name, method string
		project      *string
	}

	testCases := []azureDevOpsTestCase{
		{
			path:    "/spacelift-development/backend/_apis/git/repositories/infra/diffs/commits?api-version=7.1-preview.1",
			matches: true,
			name:    "Get Commit Diff",
			project: nullable.String("spacelift-development/backend/infra"),
			method:  http.MethodGet,
		},
		{
			path:    "/spacelift-development/backend/_apis/git/repositories/infra/pullRequests/12345/threads?api-version=7.1-preview.1",
			matches: true,
			name:    "Create Pull Request Thread",
			project: nullable.String("spacelift-development/backend/infra"),
			method:  http.MethodPost,
		},
		{
			path:    "/spacelift-development/backend/_apis/git/repositories/infra/stats/branches?api-version=7.1-preview.1",
			matches: true,
			name:    "List Branch Stats",
			project: nullable.String("spacelift-development/backend/infra"),
			method:  http.MethodGet,
		},
		{
			path:    "/spacelift-development/backend/_apis/git/repositories/infra/items?path=%2FREADME.md&api-version=7.1-preview.1",
			matches: true,
			name:    "Get Item",
			project: nullable.String("spacelift-development/backend/infra"),
			method:  http.MethodGet,
		},
		{
			path:    "/spacelift-development/backend/_apis/git/repositories/infra/commits/a171eaa2c3790cbff77ec96f429e7aee5a3e724c/statuses?api-version=7.1-preview.1",
			matches: true,
			name:    "Create Commit Status",
			project: nullable.String("spacelift-development/backend/infra"),
			method:  http.MethodPost,
		},
		{
			path:    "/spacelift-development/backend/_apis/git/repositories/infra/pullRequests?api-version=7.1-preview.1",
			matches: true,
			name:    "List Pull Requests",
			project: nullable.String("spacelift-development/backend/infra"),
			method:  http.MethodGet,
		},
		{
			path:    "/spacelift-development/backend/_apis/git/repositories/infra/pullRequests/1234?api-version=7.1-preview.1",
			matches: true,
			name:    "Get Pull Request",
			project: nullable.String("spacelift-development/backend/infra"),
			method:  http.MethodGet,
		},
		{
			path:    "/spacelift-development/backend/_apis/git/repositories/infra/pullRequests/1234/labels?api-version=7.1-preview.1",
			matches: true,
			name:    "List Pull Request Labels",
			project: nullable.String("spacelift-development/backend/infra"),
			method:  http.MethodGet,
		},
		{
			path:    "/spacelift-development/backend/_apis/git/repositories?api-version=7.1-preview.1",
			matches: true,
			name:    "List Repositories",
			method:  http.MethodGet,
		},
		{
			path:    "/spacelift-development/_apis/git/repositories?api-version=7.1-preview.1",
			matches: true,
			name:    "List Repositories",
			method:  http.MethodGet,
		},
		{
			path:    "/spacelift-development/backend/_apis/git/repositories/infra/commits/a171eaa2c3790cbff77ec96f429e7aee5a3e724c?api-version=7.1-preview.1",
			matches: true,
			name:    "Get Commit",
			project: nullable.String("spacelift-development/backend/infra"),
			method:  http.MethodGet,
		},
		{
			path:    "/spacelift-development/_apis",
			matches: true,
			name:    "List Resource Locations",
			method:  http.MethodOptions,
		},
		{
			path:    "/spacelift-development/_apis/ResourceAreas",
			matches: true,
			name:    "List Resource Areas",
			method:  http.MethodGet,
		},
		{
			path:    "/spacelift-development/backend/_apis/policy/evaluations",
			matches: true,
			name:    "List Policy Evaluations",
			method:  http.MethodGet,
		},
		// Temporarily allow any request until we're sure we have the correct set of validation
		// rules. Once we do, the following test case can be removed, and the failing test
		// case can be uncommented.
		{
			path:    "/spacelift-development/_apis/UnknownResource",
			matches: true,
			name:    "Unknown Request",
			method:  http.MethodGet,
		},
		// {
		// 	path:    "/spacelift-development/backend/_apis/git/repositories/infra/pullRequests/1234/attachments/%2Fattachment.tar.gz?api-version=7.1-preview.1",
		// 	matches: false,
		// 	method:  http.MethodGet,
		// },
	}

	executeTestCase := func(testCase azureDevOpsTestCase) {
		request, err := http.NewRequest(testCase.method, "https://github.myorg.com"+testCase.path, nil)
		require.NoError(t, err, "could not create request")

		name, project, err := matchAzureDevOpsRequest(request)

		if testCase.matches {
			require.NoError(t, err, "could not find match for %q (%s)", testCase.name, testCase.path)
			require.Equal(t, testCase.name, name, "request name not correct for %q (%s)", testCase.name, testCase.path)

			if testCase.project != nil {
				require.Equal(t, *testCase.project, project, "project did not match for %q (%s)", testCase.name, testCase.path)
			}
		} else {
			require.ErrorIs(t, ErrNoMatch, err)
		}
	}

	for _, testCase := range testCases {
		executeTestCase(testCase)

		// If the test case has a name (i.e. we're expecting a match), check it also works when
		// the Azure DevOps instance is hosted as a sub-path rather than at the root of the domain.
		if testCase.name != "" {
			testCase.path = "/tfs" + testCase.path
			executeTestCase(testCase)
		}
	}
}
