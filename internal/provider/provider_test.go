// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package sakura

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/saclient-go"
	"github.com/stretchr/testify/require"
)

var _ saclient.TerraformProviderInterface = (*sakuraProviderModel)(nil)

func TestLookupClientConfigServicePrincipalID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		field     types.String
		wantValue string
		wantOK    bool
	}{
		{
			name:      "null returns false",
			field:     types.StringNull(),
			wantValue: "",
			wantOK:    false,
		},
		{
			name:      "unknown returns false",
			field:     types.StringUnknown(),
			wantValue: "",
			wantOK:    false,
		},
		{
			name:      "valid value returns true",
			field:     types.StringValue("113702516320"),
			wantValue: "113702516320",
			wantOK:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &sakuraProviderModel{ServicePrincipalID: tt.field}
			got, ok := m.LookupClientConfigServicePrincipalID()
			require.Equal(t, tt.wantOK, ok)
			require.Equal(t, tt.wantValue, got)
		})
	}
}

func TestLookupClientConfigServicePrincipalKeyID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		field     types.String
		wantValue string
		wantOK    bool
	}{
		{
			name:      "null returns false",
			field:     types.StringNull(),
			wantValue: "",
			wantOK:    false,
		},
		{
			name:      "unknown returns false",
			field:     types.StringUnknown(),
			wantValue: "",
			wantOK:    false,
		},
		{
			name:      "valid value returns true",
			field:     types.StringValue("key-abc123"),
			wantValue: "key-abc123",
			wantOK:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &sakuraProviderModel{ServicePrincipalKeyID: tt.field}
			got, ok := m.LookupClientConfigServicePrincipalKeyID()
			require.Equal(t, tt.wantOK, ok)
			require.Equal(t, tt.wantValue, got)
		})
	}
}

func TestLookupClientConfigPrivateKeyPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		field     types.String
		wantValue string
		wantOK    bool
	}{
		{
			name:      "null returns false",
			field:     types.StringNull(),
			wantValue: "",
			wantOK:    false,
		},
		{
			name:      "unknown returns false",
			field:     types.StringUnknown(),
			wantValue: "",
			wantOK:    false,
		},
		{
			name:      "valid value returns true",
			field:     types.StringValue("/path/to/key.pem"),
			wantValue: "/path/to/key.pem",
			wantOK:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &sakuraProviderModel{ServicePrincipalPrivateKeyPath: tt.field}
			got, ok := m.LookupClientConfigPrivateKeyPath()
			require.Equal(t, tt.wantOK, ok)
			require.Equal(t, tt.wantValue, got)
		})
	}
}
