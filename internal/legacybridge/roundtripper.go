// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package legacybridge

import (
	"errors"
	"net/http"
	"strings"

	"github.com/sacloud/saclient-go"
)

// SaclientRoundTripper delegates HTTP requests to saclient.ClientAPI.Do().
//
// RoundTrip clones the incoming request, conditionally removes the
// Authorization header if it has a "Basic " prefix (to strip the empty
// Basic Auth set by go-http), and then delegates to saclient.Client.Do().
// saclient's middleware chain sets the correct authentication header.
type SaclientRoundTripper struct {
	Client saclient.ClientAPI
}

func (rt *SaclientRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if rt.Client == nil {
		return nil, errors.New("legacybridge: saclient client is not initialized")
	}

	cloned := req.Clone(req.Context())
	if auth := cloned.Header.Get("Authorization"); strings.HasPrefix(auth, "Basic ") {
		cloned.Header.Del("Authorization")
	}

	return rt.Client.Do(cloned)
}

var _ http.RoundTripper = (*SaclientRoundTripper)(nil)

// NewHTTPClient creates an *http.Client backed by saclient.ClientAPI.
// The returned client is intended for injection into api-client-go's
// Options.HttpClient field.
func NewHTTPClient(sc saclient.ClientAPI) *http.Client {
	return &http.Client{
		Transport: &SaclientRoundTripper{Client: sc},
	}
}
