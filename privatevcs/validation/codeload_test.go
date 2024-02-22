package validation_test

import (
	"net/http"
	"testing"

	"github.com/franela/goblin"
	"github.com/go-kit/log"
	. "github.com/onsi/gomega"
	"github.com/spacelift-io/spcontext"

	"github.com/spacelift-io/vcs-agent/privatevcs/validation"
)

func TestRewriteGitHubTarballRequest(t *testing.T) {
	g := goblin.Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("RewriteGitHubTarballRequest", func() {
		var ctx *spcontext.Context
		var req *http.Request
		var url string
		var vendor validation.Vendor

		g.BeforeEach(func() {
			ctx = spcontext.New(log.NewNopLogger())
		})

		g.JustBeforeEach(func() {
			var err error

			if req, err = http.NewRequest("GET", url, nil); err != nil {
				panic(err)
			}

			ctx = validation.RewriteGitHubTarballRequest(ctx, vendor, req)
		})

		g.Describe("when not a GitHub request", func() {
			g.BeforeEach(func() {
				vendor = validation.GitLab
				url = "https://gitlab.com/foo/bar"
			})

			g.It("should not rewrite the request", func() {
				Expect(req.URL.Host).To(Equal("gitlab.com"))
			})
		})

		g.Describe("when a GitHub request", func() {
			g.BeforeEach(func() { vendor = validation.GitHubEnterprise })

			g.Describe("when the request is not a tarball download request", func() {
				g.BeforeEach(func() { url = "https://github.corp.com/api/v3/repos/octocats/infra/deployments" })

				g.It("should not rewrite the request", func() {
					Expect(req.URL.Host).To(Equal("github.corp.com"))
				})
			})

			g.Describe("when the request is a tarball download request", func() {
				g.Describe("with codeload isolation disabled", func() {
					g.BeforeEach(func() {
						url = "https://github.corp.com/_codeload/octocats/infra/legacy.tar.gz/master"
					})

					g.It("should not rewrite the request", func() {
						Expect(req.URL.Host).To(Equal("github.corp.com"))
					})
				})

				g.Describe("with codeload isolation enabled", func() {
					g.BeforeEach(func() {
						url = "https://github.corp.com/octocats/infra/legacy.tar.gz/master"
					})

					g.It("should rewrite the request", func() {
						Expect(req.URL.Host).To(Equal("codeload.github.corp.com"))
					})
				})
			})
		})
	})
}
