package blocklist_test

import (
	"net/http"
	"testing"

	"github.com/franela/goblin"
	. "github.com/onsi/gomega"

	"github.com/spacelift-io/vcs-agent/privatevcs/validation/blocklist"
)

func TestRule(t *testing.T) {
	g := goblin.Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Rule", func() {
		var sut *blocklist.Rule

		g.BeforeEach(func() {
			sut = new(blocklist.Rule)
		})

		g.Describe("Validate", func() {
			var err error

			g.JustBeforeEach(func() { err = sut.Validate() })

			g.Describe("with no name (default)", func() {
				g.It("should return an error", func() {
					Expect(err).To(MatchError("rule name is required"))
				})
			})

			g.Describe("with a name", func() {
				g.BeforeEach(func() { sut.Name = "foo" })

				g.Describe("with invalid method regexp", func() {
					g.BeforeEach(func() { sut.Method = "[" })

					g.It("should return an error", func() {
						Expect(err).To(MatchError("could not compile rule \"foo\": invalid method matcher: error parsing regexp: missing closing ]: `[`"))
					})
				})

				g.Describe("with valid method regexp", func() {
					g.BeforeEach(func() { sut.Method = "GET" })

					g.Describe("with invalid path regexp", func() {
						g.BeforeEach(func() { sut.Path = "[" })

						g.It("should return an error", func() {
							Expect(err).To(MatchError("could not compile rule \"foo\": invalid path matcher: error parsing regexp: missing closing ]: `[`"))
						})
					})

					g.Describe("with valid path regexp", func() {
						g.BeforeEach(func() { sut.Path = ".*" })

						g.It("should not return an error", func() {
							Expect(err).ToNot(HaveOccurred())
						})
					})
				})
			})
		})

		g.Describe("Matches", func() {
			var match bool
			var req *http.Request

			g.BeforeEach(func() {
				sut.Name = "foo"

				var err error
				req, err = http.NewRequest("GET", "https://example.com/foo", nil)

				if err != nil {
					panic(err)
				}
			})

			g.JustBeforeEach(func() {
				if err := sut.Validate(); err != nil {
					panic(err)
				}
				match = sut.Matches(req)
			})

			g.Describe("with a matching method", func() {
				g.BeforeEach(func() { sut.Method = "GET" })

				g.Describe("with a matching path", func() {
					g.BeforeEach(func() { sut.Path = ".*" })

					g.It("should return true", func() { Expect(match).To(BeTrue()) })
				})

				g.Describe("with a non-matching path", func() {
					g.BeforeEach(func() { sut.Path = "/bar" })

					g.It("should return false", func() { Expect(match).To(BeFalse()) })
				})
			})

			g.Describe("with a non-matching method", func() {
				g.BeforeEach(func() { sut.Method = "POST" })

				g.Describe("with a matching path", func() {
					g.BeforeEach(func() { sut.Path = ".*" })

					g.It("should return false", func() { Expect(match).To(BeFalse()) })
				})

				g.Describe("with a non-matching path", func() {
					g.BeforeEach(func() { sut.Path = "/bar" })

					g.It("should return false", func() { Expect(match).To(BeFalse()) })
				})
			})
		})
	})
}
