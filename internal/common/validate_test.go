// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      Config
		wantErr     bool
		errContains []string
	}{
		{
			name: "API key only complete - no error",
			config: Config{
				AccessToken:       "token",
				AccessTokenSecret: "secret",
			},
			wantErr: false,
		},
		{
			name: "SP only complete - no error",
			config: Config{
				ServicePrincipalID:             "sp-id",
				ServicePrincipalKeyID:          "sp-key-id",
				ServicePrincipalPrivateKeyPath: "/path/to/key.pem",
			},
			wantErr: false,
		},
		{
			name: "both complete - no error",
			config: Config{
				AccessToken:                    "token",
				AccessTokenSecret:              "secret",
				ServicePrincipalID:             "sp-id",
				ServicePrincipalKeyID:          "sp-key-id",
				ServicePrincipalPrivateKeyPath: "/path/to/key.pem",
			},
			wantErr: false,
		},
		{
			name:    "neither set - error",
			config:  Config{},
			wantErr: true,
			errContains: []string{
				"AccessToken/AccessTokenSecret or ServicePrincipal credentials are required",
			},
		},
		{
			name: "API key partial - token only",
			config: Config{
				AccessToken: "token",
			},
			wantErr: true,
			errContains: []string{
				"AccessToken is set but AccessTokenSecret is missing",
			},
		},
		{
			name: "API key partial - secret only",
			config: Config{
				AccessTokenSecret: "secret",
			},
			wantErr: true,
			errContains: []string{
				"AccessTokenSecret is set but AccessToken is missing",
			},
		},
		{
			name: "SP partial - ID only",
			config: Config{
				ServicePrincipalID: "sp-id",
			},
			wantErr: true,
			errContains: []string{
				"ServicePrincipalKeyID is missing",
				"ServicePrincipalPrivateKeyPath is missing",
			},
		},
		{
			name: "SP partial - key ID and path without ID",
			config: Config{
				ServicePrincipalKeyID:          "sp-key-id",
				ServicePrincipalPrivateKeyPath: "/path/to/key.pem",
			},
			wantErr: true,
			errContains: []string{
				"ServicePrincipalID is missing",
			},
		},
		{
			name: "API key complete with partial SP - error on SP partial",
			config: Config{
				AccessToken:        "token",
				AccessTokenSecret:  "secret",
				ServicePrincipalID: "sp-id",
			},
			wantErr: true,
			errContains: []string{
				"ServicePrincipalKeyID is missing",
				"ServicePrincipalPrivateKeyPath is missing",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.config.validate()
			if tt.wantErr {
				require.Error(t, err)
				for _, s := range tt.errContains {
					assert.Contains(t, err.Error(), s)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHasBothAuthMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config Config
		want   bool
	}{
		{
			name: "both complete",
			config: Config{
				AccessToken:                    "token",
				AccessTokenSecret:              "secret",
				ServicePrincipalID:             "sp-id",
				ServicePrincipalKeyID:          "sp-key-id",
				ServicePrincipalPrivateKeyPath: "/path/to/key.pem",
			},
			want: true,
		},
		{
			name: "API key only",
			config: Config{
				AccessToken:       "token",
				AccessTokenSecret: "secret",
			},
			want: false,
		},
		{
			name: "SP only",
			config: Config{
				ServicePrincipalID:             "sp-id",
				ServicePrincipalKeyID:          "sp-key-id",
				ServicePrincipalPrivateKeyPath: "/path/to/key.pem",
			},
			want: false,
		},
		{
			name:   "neither",
			config: Config{},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.config.HasBothAuthMethods()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidate_errMessages(t *testing.T) {
	t.Parallel()

	// Verify that error messages do NOT contain private key path values
	cfg := Config{
		ServicePrincipalID: "sp-id",
	}
	err := cfg.validate()
	require.Error(t, err)
	assert.False(t, strings.Contains(err.Error(), "/path/to"), "error message should not contain private key paths")
}
