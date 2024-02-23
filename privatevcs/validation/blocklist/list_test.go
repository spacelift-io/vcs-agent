package blocklist_test

import (
	"net/http"
	"testing"

	"github.com/franela/goblin"
	"github.com/go-kit/log"
	. "github.com/onsi/gomega"
	"github.com/spacelift-io/spcontext"

	"github.com/spacelift-io/vcs-agent/privatevcs/validation/blocklist"
)

func TestList(t *testing.T) {
	g := goblin.Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("List", func() {
		var sut *blocklist.List

		g.Describe("Load", func() {
			var err error
			var path string

			g.JustBeforeEach(func() { sut, err = blocklist.Load(path) })

			g.Describe("with an invalid path", func() {
				g.BeforeEach(func() { path = "fixtures/not.there" })

				g.It("should return an error", func() {
					Expect(sut).To(BeNil())
					Expect(err).To(MatchError(`couldn't read the blocklist file "fixtures/not.there": open fixtures/not.there: no such file or directory`))
				})
			})

			g.Describe("with a duplicate rule", func() {
				g.BeforeEach(func() { path = "fixtures/duplicate.yaml" })

				g.It("should return an error", func() {
					Expect(sut).To(BeNil())
					Expect(err).To(MatchError(`invalid blocklist file "fixtures/duplicate.yaml": duplicate rule name "Duplicate"`))
				})
			})

			g.Describe("with an invalid rule", func() {
				g.BeforeEach(func() { path = "fixtures/invalid.yaml" })

				g.It("should return an error", func() {
					Expect(sut).To(BeNil())
					Expect(err).To(MatchError("invalid blocklist file \"fixtures/invalid.yaml\": invalid rule 0: could not compile rule \"Invalid\": invalid path matcher: error parsing regexp: missing closing ]: `[`"))
				})
			})

			g.Describe("with a valid blocklist", func() {
				g.BeforeEach(func() { path = "fixtures/valid.yaml" })

				g.It("should not return an error", func() {
					Expect(err).To(BeNil())
				})

				g.It("should have loaded the rule", func() {
					Expect(sut.Rules).To(HaveLen(2))
				})
			})
		})

		g.Describe("Validate", func() {
			var err error
			var ctx *spcontext.Context
			var req *http.Request

			g.BeforeEach(func() {
				ctx = spcontext.New(log.NewNopLogger())

				req, err = http.NewRequest("GET", "https://example.com/foo", nil)
				if err != nil {
					panic(err)
				}

				sut = new(blocklist.List)
			})

			g.JustBeforeEach(func() { _, err = sut.Validate(ctx, "", req) })

			g.Describe("with an empty list (default)", func() {
				g.It("should pass validation", func() {
					Expect(err).To(BeNil())
				})
			})

			g.Describe("with a non-matching rule", func() {
				g.BeforeEach(func() {
					sut.Rules = []*blocklist.Rule{{
						Name:   "NonMatching",
						Method: "POST",
						Path:   ".*",
					}}

					if err := sut.Compile(); err != nil {
						panic(err)
					}
				})

				g.It("should pass validation", func() {
					Expect(err).To(BeNil())
				})
			})

			g.Describe("with a matching rule", func() {
				g.BeforeEach(func() {
					sut.Rules = []*blocklist.Rule{{
						Name:   "Matching",
						Method: "GET",
						Path:   "/foo",
					}}

					if err := sut.Compile(); err != nil {
						panic(err)
					}
				})

				g.It("should fail validation", func() {
					Expect(err).To(MatchError(`request blocked by rule "Matching"`))
				})
			})
		})
	})
}
