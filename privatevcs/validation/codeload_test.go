package validation_test

import (
	"net/http"
	"testing"

	"github.com/go-kit/log"
	"github.com/spacelift-io/spcontext"
	"github.com/stretchr/testify/assert"

	"github.com/spacelift-io/vcs-agent/privatevcs/validation"
)

func TestRewriteGitHubTarballRequest(t *testing.T) {
	t.Run("when not a GitHub request", func(t *testing.T) {
		ctx := spcontext.New(log.NewNopLogger())
		vendor := validation.GitLab
		url := "https://gitlab.com/foo/bar"

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		validation.RewriteGitHubTarballRequest(ctx, vendor, req)

		assert.Equal(t, "gitlab.com", req.URL.Host)
	})

	t.Run("when a GitHub request", func(t *testing.T) {
		t.Run("when the request is not a tarball download request", func(t *testing.T) {
			ctx := spcontext.New(log.NewNopLogger())
			vendor := validation.GitHubEnterprise
			url := "https://github.corp.com/api/v3/repos/octocats/infra/deployments"

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			validation.RewriteGitHubTarballRequest(ctx, vendor, req)

			assert.Equal(t, "github.corp.com", req.URL.Host)
		})

		t.Run("when the request is a tarball download request", func(t *testing.T) {
			t.Run("with codeload isolation disabled", func(t *testing.T) {
				ctx := spcontext.New(log.NewNopLogger())
				vendor := validation.GitHubEnterprise
				url := "https://github.corp.com/_codeload/octocats/infra/legacy.tar.gz/master"

				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					t.Fatalf("failed to create request: %v", err)
				}

				validation.RewriteGitHubTarballRequest(ctx, vendor, req)

				assert.Equal(t, "github.corp.com", req.URL.Host)
			})

			t.Run("with codeload isolation enabled", func(t *testing.T) {
				ctx := spcontext.New(log.NewNopLogger())
				vendor := validation.GitHubEnterprise
				url := "https://github.corp.com/octocats/infra/legacy.tar.gz/master"

				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					t.Fatalf("failed to create request: %v", err)
				}

				validation.RewriteGitHubTarballRequest(ctx, vendor, req)

				assert.Equal(t, "codeload.github.corp.com", req.URL.Host)
			})
		})
	})
}
