// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package legacybridge

import (
	"flag"
	"io"
	"net/http"
	"strings"
	"testing"

	old "github.com/sacloud/api-client-go"
	"github.com/sacloud/saclient-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSaclient implements saclient.ClientAPI for testing.
// Only Do() is meaningful; other methods are no-ops.
type mockSaclient struct {
	lastReq  *http.Request
	response *http.Response
	err      error
}

func (m *mockSaclient) Do(req *http.Request) (*http.Response, error) {
	m.lastReq = req
	if m.err != nil {
		return nil, m.err
	}
	if m.response != nil {
		return m.response, nil
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("")),
	}, nil
}

func (m *mockSaclient) Populate() error                                                   { return nil }
func (m *mockSaclient) Dup() saclient.ClientAPI                                           { return m }
func (m *mockSaclient) SetEnviron([]string) error                                         { return nil }
func (m *mockSaclient) SettingsFromTerraformProvider(saclient.TerraformProviderInterface) error {
	return nil
}
func (m *mockSaclient) CompatSettingsFromAPIClientParams(string, ...old.ClientParam) error { return nil }
func (m *mockSaclient) CompatSettingsFromAPIClientOptions(...*old.Options) error           { return nil }
func (m *mockSaclient) ServerURL() string                                                 { return "" }
func (m *mockSaclient) FlagSet(flag.ErrorHandling) *flag.FlagSet                          { return nil }
func (m *mockSaclient) Profile() (*saclient.Profile, error)                               { return nil, nil }
func (m *mockSaclient) ProfileName() (*string, *string)                                   { return nil, nil }
func (m *mockSaclient) ProfileOp() (saclient.ProfileAPI, error)                           { return nil, nil }
var _ saclient.ClientAPI = (*mockSaclient)(nil)

func TestSaclientRoundTripper_NilClient(t *testing.T) {
	t.Parallel()
	rt := &SaclientRoundTripper{Client: nil}
	req, _ := http.NewRequest("GET", "https://example.com", nil)

	resp, err := rt.RoundTrip(req)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "saclient client is not initialized")
}

func TestSaclientRoundTripper_BasicAuthRemoved(t *testing.T) {
	t.Parallel()
	doer := &mockSaclient{}
	rt := &SaclientRoundTripper{Client: doer}
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	req.Header.Set("Authorization", "Basic Og==")

	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Original request should NOT be modified
	assert.Equal(t, "Basic Og==", req.Header.Get("Authorization"), "original request must not be modified")

	// Cloned request sent to saclient should have Authorization removed
	assert.Empty(t, doer.lastReq.Header.Get("Authorization"), "Basic auth should be removed from cloned request")
}

func TestSaclientRoundTripper_BearerPreserved(t *testing.T) {
	t.Parallel()
	doer := &mockSaclient{}
	rt := &SaclientRoundTripper{Client: doer}
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	req.Header.Set("Authorization", "Bearer some-token")

	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Original request should NOT be modified
	assert.Equal(t, "Bearer some-token", req.Header.Get("Authorization"))

	// Bearer should be preserved in cloned request
	assert.Equal(t, "Bearer some-token", doer.lastReq.Header.Get("Authorization"), "Bearer auth should be preserved")
}

func TestSaclientRoundTripper_NoAuthHeader(t *testing.T) {
	t.Parallel()
	doer := &mockSaclient{}
	rt := &SaclientRoundTripper{Client: doer}
	req, _ := http.NewRequest("POST", "https://example.com", nil)

	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Empty(t, doer.lastReq.Header.Get("Authorization"))
}

func TestSaclientRoundTripper_DelegatesRequest(t *testing.T) {
	t.Parallel()
	doer := &mockSaclient{}
	rt := &SaclientRoundTripper{Client: doer}
	req, _ := http.NewRequest("POST", "https://api.example.com/v1/resource", nil)
	req.Header.Set("User-Agent", "test-ua")
	req.Header.Set("Accept-Language", "ja-JP")

	_, err := rt.RoundTrip(req)
	require.NoError(t, err)

	// Verify the cloned request preserves method, URL, and headers
	assert.Equal(t, "POST", doer.lastReq.Method)
	assert.Equal(t, "https://api.example.com/v1/resource", doer.lastReq.URL.String())
	assert.Equal(t, "test-ua", doer.lastReq.Header.Get("User-Agent"))
	assert.Equal(t, "ja-JP", doer.lastReq.Header.Get("Accept-Language"))
}

func TestNewHTTPClient_TransportType(t *testing.T) {
	t.Parallel()
	doer := &mockSaclient{}
	hc := NewHTTPClient(doer)

	require.NotNil(t, hc)
	_, ok := hc.Transport.(*SaclientRoundTripper)
	assert.True(t, ok, "Transport should be *SaclientRoundTripper")
}
