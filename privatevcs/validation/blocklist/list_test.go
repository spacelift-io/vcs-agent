package blocklist_test

import (
	"net/http"
	"testing"

	"github.com/go-kit/log"
	"github.com/spacelift-io/spcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spacelift-io/vcs-agent/privatevcs/validation/blocklist"
)

func TestListLoad(t *testing.T) {
	t.Run("with an invalid path", func(t *testing.T) {
		path := "fixtures/not.there"

		sut, err := blocklist.Load(path)

		assert.Nil(t, sut)
		assert.EqualError(t, err, `couldn't read the blocklist file "fixtures/not.there": open fixtures/not.there: no such file or directory`)
	})

	t.Run("with a duplicate rule", func(t *testing.T) {
		path := "fixtures/duplicate.yaml"

		sut, err := blocklist.Load(path)

		assert.Nil(t, sut)
		assert.EqualError(t, err, `invalid blocklist file "fixtures/duplicate.yaml": duplicate rule name "Duplicate"`)
	})

	t.Run("with an invalid rule", func(t *testing.T) {
		path := "fixtures/invalid.yaml"

		sut, err := blocklist.Load(path)

		assert.Nil(t, sut)
		assert.EqualError(t, err, "invalid blocklist file \"fixtures/invalid.yaml\": invalid rule 0: could not compile rule \"Invalid\": invalid path matcher: error parsing regexp: missing closing ]: `[`")
	})

	t.Run("with a valid blocklist", func(t *testing.T) {
		path := "fixtures/valid.yaml"

		sut, err := blocklist.Load(path)

		assert.NoError(t, err)
		assert.Len(t, sut.Rules, 2)
	})
}

func TestListValidate(t *testing.T) {
	t.Run("with an empty list (default)", func(t *testing.T) {
		ctx := spcontext.New(log.NewNopLogger())
		req, err := http.NewRequest("GET", "https://example.com/foo", nil)
		require.NoError(t, err, "failed to create request")

		sut := new(blocklist.List)

		_, err = sut.Validate(ctx, "", req)

		assert.NoError(t, err)
	})

	t.Run("with a non-matching rule", func(t *testing.T) {
		ctx := spcontext.New(log.NewNopLogger())
		req, err := http.NewRequest("GET", "https://example.com/foo", nil)
		require.NoError(t, err, "failed to create request")

		sut := new(blocklist.List)
		sut.Rules = []*blocklist.Rule{{
			Name:   "NonMatching",
			Method: "POST",
			Path:   ".*",
		}}

		err = sut.Compile()
		require.NoError(t, err, "failed to compile rules")

		_, err = sut.Validate(ctx, "", req)

		assert.NoError(t, err)
	})

	t.Run("with a matching rule", func(t *testing.T) {
		ctx := spcontext.New(log.NewNopLogger())
		req, err := http.NewRequest("GET", "https://example.com/foo", nil)
		require.NoError(t, err, "failed to create request")

		sut := new(blocklist.List)
		sut.Rules = []*blocklist.Rule{{
			Name:   "Matching",
			Method: "GET",
			Path:   "/foo",
		}}

		err = sut.Compile()
		require.NoError(t, err, "failed to compile rules")

		_, err = sut.Validate(ctx, "", req)

		assert.EqualError(t, err, `request blocked by rule "Matching"`)
	})
}