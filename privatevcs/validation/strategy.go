package validation

import (
	"net/http"

	"github.com/spacelift-io/spcontext"
)

// Strategy is a request validation strategy. The purpose of a strategy is to
// decide which requests should be allowed and which should be blocked.
// An error returned by a strategy is treated as a block.
type Strategy interface {
	// Validate validates the request and returns an error if the request should
	// be blocked.
	Validate(*spcontext.Context, Vendor, *http.Request) (*spcontext.Context, error)
}
