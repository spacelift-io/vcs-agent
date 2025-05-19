package blocklist_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spacelift-io/vcs-agent/privatevcs/validation/blocklist"
)

func TestRuleValidate(t *testing.T) {
	t.Run("with no name (default)", func(t *testing.T) {
		sut := new(blocklist.Rule)

		err := sut.Validate()

		assert.EqualError(t, err, "rule name is required")
	})

	t.Run("with a name", func(t *testing.T) {
		t.Run("with invalid method regexp", func(t *testing.T) {
			sut := new(blocklist.Rule)
			sut.Name = "foo"
			sut.Method = "["

			err := sut.Validate()

			assert.EqualError(t, err, "could not compile rule \"foo\": invalid method matcher: error parsing regexp: missing closing ]: `[`")
		})

		t.Run("with valid method regexp", func(t *testing.T) {
			t.Run("with invalid path regexp", func(t *testing.T) {
				sut := new(blocklist.Rule)
				sut.Name = "foo"
				sut.Method = "GET"
				sut.Path = "["

				err := sut.Validate()

				assert.EqualError(t, err, "could not compile rule \"foo\": invalid path matcher: error parsing regexp: missing closing ]: `[`")
			})

			t.Run("with valid path regexp", func(t *testing.T) {
				sut := new(blocklist.Rule)
				sut.Name = "foo"
				sut.Method = "GET"
				sut.Path = ".*"

				err := sut.Validate()

				assert.NoError(t, err)
			})
		})
	})
}

func TestRuleMatches(t *testing.T) {
	// Create a common test request
	req, err := http.NewRequest("GET", "https://example.com/foo", nil)
	require.NoError(t, err, "failed to create request")

	t.Run("with a matching method", func(t *testing.T) {
		t.Run("with a matching path", func(t *testing.T) {
			sut := new(blocklist.Rule)
			sut.Name = "foo"
			sut.Method = "GET"
			sut.Path = ".*"

			err := sut.Validate()
			require.NoError(t, err, "failed to validate rule")

			match := sut.Matches(req)

			assert.True(t, match)
		})

		t.Run("with a non-matching path", func(t *testing.T) {
			sut := new(blocklist.Rule)
			sut.Name = "foo"
			sut.Method = "GET"
			sut.Path = "/bar"

			err := sut.Validate()
			require.NoError(t, err, "failed to validate rule")

			match := sut.Matches(req)

			assert.False(t, match)
		})
	})

	t.Run("with a non-matching method", func(t *testing.T) {
		t.Run("with a matching path", func(t *testing.T) {
			sut := new(blocklist.Rule)
			sut.Name = "foo"
			sut.Method = "POST"
			sut.Path = ".*"

			err := sut.Validate()
			require.NoError(t, err, "failed to validate rule")

			match := sut.Matches(req)

			assert.False(t, match)
		})

		t.Run("with a non-matching path", func(t *testing.T) {
			sut := new(blocklist.Rule)
			sut.Name = "foo"
			sut.Method = "POST"
			sut.Path = "/bar"

			err := sut.Validate()
			require.NoError(t, err, "failed to validate rule")

			match := sut.Matches(req)

			assert.False(t, match)
		})
	})
}