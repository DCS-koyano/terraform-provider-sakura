// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package common_test

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/sacloud/api-client-go/profile"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/defaults"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initTestProfileDir() func() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	os.Setenv("SAKURACLOUD_PROFILE_DIR", wd) //nolint
	profileDir := filepath.Join(wd, ".usacloud")
	if _, err := os.Stat(profileDir); err == nil {
		os.RemoveAll(profileDir) //nolint
	}

	return func() {
		os.RemoveAll(profileDir) //nolint
	}
}

func TestConfig_NewClient_loadFromProfile(t *testing.T) {
	defer initTestProfileDir()()

	defaultProfile := &profile.ConfigValue{
		AccessToken:          "token",
		AccessTokenSecret:    "secret",
		Zone:                 "dummy1",
		Zones:                []string{"dummy1", "dummy2"},
		UserAgent:            "dummy-ua",
		AcceptLanguage:       "ja-JP",
		RetryMax:             1,
		RetryWaitMin:         2,
		RetryWaitMax:         3,
		StatePollingTimeout:  4,
		StatePollingInterval: 5,
		HTTPRequestTimeout:   6,
		HTTPRequestRateLimit: 7,
		APIRootURL:           "dummy",
		TraceMode:            "dummy",
		FakeMode:             true,
		FakeStorePath:        "dummy",
	}
	testProfile := &profile.ConfigValue{
		AccessToken:          "testtoken",
		AccessTokenSecret:    "testsecret",
		Zone:                 "test",
		Zones:                []string{"test1", "test2"},
		UserAgent:            "test-ua",
		AcceptLanguage:       "ja-JP",
		RetryMax:             7,
		RetryWaitMin:         6,
		RetryWaitMax:         5,
		StatePollingTimeout:  4,
		StatePollingInterval: 3,
		HTTPRequestTimeout:   2,
		HTTPRequestRateLimit: 1,
		APIRootURL:           "test",
		TraceMode:            "test",
		FakeMode:             false,
		FakeStorePath:        "test",
	}

	// プロファイル指定なし & デフォルトプロファイルなし
	// プロファイル指定なし & デフォルトプロファイルあり
	// プロファイル指定あり & 指定プロファイルが存在しない
	// プロファイル指定あり 通常

	cases := []struct {
		scenario       string
		in             *common.Config
		profiles       map[string]*profile.ConfigValue
		expect         *common.Config
		currentProfile string
		err            error
	}{
		{
			scenario: "If profileName is not specified and profile is not exists, use default values",
			in: &common.Config{
				Profile:             "",
				Zone:                defaults.Zone,
				Zones:               iaas.SakuraCloudZones,
				RetryMax:            defaults.RetryMax,
				APIRequestTimeout:   defaults.APIRequestTimeout,
				APIRequestRateLimit: defaults.APIRequestRateLimit,
			},
			profiles: map[string]*profile.ConfigValue{},
			expect: &common.Config{
				Profile:             "default",
				Zone:                defaults.Zone,
				Zones:               iaas.SakuraCloudZones,
				RetryMax:            defaults.RetryMax,
				APIRequestTimeout:   defaults.APIRequestTimeout,
				APIRequestRateLimit: defaults.APIRequestRateLimit,
			},
		},
		{
			scenario: "If no profile is specified and a current profile exists, it is loaded from the current profile",
			in: &common.Config{
				Profile: "",
			},
			profiles: map[string]*profile.ConfigValue{
				"default": defaultProfile,
				"test":    testProfile,
			},
			currentProfile: "test",
			expect: &common.Config{
				Profile:             "test",
				AccessToken:         testProfile.AccessToken,
				AccessTokenSecret:   testProfile.AccessTokenSecret,
				Zone:                testProfile.Zone,
				Zones:               testProfile.Zones,
				TraceMode:           testProfile.TraceMode,
				AcceptLanguage:      testProfile.AcceptLanguage,
				APIRootURL:          testProfile.APIRootURL,
				RetryMax:            testProfile.RetryMax,
				RetryWaitMin:        testProfile.RetryWaitMin,
				RetryWaitMax:        testProfile.RetryWaitMax,
				APIRequestTimeout:   testProfile.HTTPRequestTimeout,
				APIRequestRateLimit: testProfile.HTTPRequestRateLimit,
			},
		},
		{
			scenario: "Values in the config are not overridden by the profile",
			in: &common.Config{
				Profile:           "",
				AccessToken:       "token",
				AccessTokenSecret: "secret",
				Zone:              "is1c",
			},
			profiles: map[string]*profile.ConfigValue{
				"default": defaultProfile,
				"test":    testProfile,
			},
			currentProfile: "test",
			expect: &common.Config{
				Profile:             "test",
				AccessToken:         "token",
				AccessTokenSecret:   "secret",
				Zone:                "is1c",
				Zones:               testProfile.Zones,
				TraceMode:           testProfile.TraceMode,
				AcceptLanguage:      testProfile.AcceptLanguage,
				APIRootURL:          testProfile.APIRootURL,
				RetryMax:            testProfile.RetryMax,
				RetryWaitMin:        testProfile.RetryWaitMin,
				RetryWaitMax:        testProfile.RetryWaitMax,
				APIRequestTimeout:   testProfile.HTTPRequestTimeout,
				APIRequestRateLimit: testProfile.HTTPRequestRateLimit,
			},
		},
		{
			scenario: "ProfileName is not specified and Profile is exists",
			in: &common.Config{
				Profile:             "",
				Zone:                defaults.Zone,
				Zones:               iaas.SakuraCloudZones,
				RetryMax:            defaults.RetryMax,
				APIRequestTimeout:   defaults.APIRequestTimeout,
				APIRequestRateLimit: defaults.APIRequestRateLimit,
			},
			profiles: map[string]*profile.ConfigValue{
				"default": defaultProfile,
			},
			expect: &common.Config{
				Profile:             "default",
				AccessToken:         defaultProfile.AccessToken,
				AccessTokenSecret:   defaultProfile.AccessTokenSecret,
				Zone:                defaults.Zone,
				Zones:               iaas.SakuraCloudZones,
				TraceMode:           defaultProfile.TraceMode,
				AcceptLanguage:      defaultProfile.AcceptLanguage,
				APIRootURL:          defaultProfile.APIRootURL,
				RetryMax:            defaults.RetryMax,
				RetryWaitMin:        defaultProfile.RetryWaitMin,
				RetryWaitMax:        defaultProfile.RetryWaitMax,
				APIRequestTimeout:   defaults.APIRequestTimeout,
				APIRequestRateLimit: defaults.APIRequestRateLimit,
			},
		},
		{
			scenario: "Empty Config and Profile is exists",
			in: &common.Config{
				Profile:             "",
				Zone:                "",
				Zones:               nil,
				RetryMax:            0,
				APIRequestTimeout:   0,
				APIRequestRateLimit: 0,
			},
			profiles: map[string]*profile.ConfigValue{
				"default": defaultProfile,
			},
			expect: &common.Config{
				Profile:             "default",
				AccessToken:         defaultProfile.AccessToken,
				AccessTokenSecret:   defaultProfile.AccessTokenSecret,
				Zone:                defaultProfile.Zone,
				Zones:               defaultProfile.Zones,
				TraceMode:           defaultProfile.TraceMode,
				AcceptLanguage:      defaultProfile.AcceptLanguage,
				APIRootURL:          defaultProfile.APIRootURL,
				RetryMax:            defaultProfile.RetryMax,
				RetryWaitMin:        defaultProfile.RetryWaitMin,
				RetryWaitMax:        defaultProfile.RetryWaitMax,
				APIRequestTimeout:   defaultProfile.HTTPRequestTimeout,
				APIRequestRateLimit: defaultProfile.HTTPRequestRateLimit,
			},
		},
		{
			scenario: "ProfileName is not specified with some values and Profile is exists",
			in: &common.Config{
				Profile:             "",
				AccessToken:         "from config",
				AccessTokenSecret:   "from config",
				Zone:                "from config",
				Zones:               []string{"zone1", "zone2"},
				TraceMode:           "from config",
				AcceptLanguage:      "from config",
				APIRootURL:          "from config",
				RetryMax:            8080,
				RetryWaitMin:        8080,
				RetryWaitMax:        8080,
				APIRequestTimeout:   8080,
				APIRequestRateLimit: 8080,
			},
			profiles: map[string]*profile.ConfigValue{
				"default": defaultProfile,
			},
			expect: &common.Config{
				Profile:             "default",
				AccessToken:         "from config",
				AccessTokenSecret:   "from config",
				Zone:                "from config",
				Zones:               []string{"zone1", "zone2"},
				TraceMode:           "from config",
				AcceptLanguage:      "from config",
				APIRootURL:          "from config",
				RetryMax:            8080,
				RetryWaitMin:        8080,
				RetryWaitMax:        8080,
				APIRequestTimeout:   8080,
				APIRequestRateLimit: 8080,
			},
		},
		{
			scenario: "Profile name specified but not exists",
			in: &common.Config{
				Profile: "test",
			},
			profiles: map[string]*profile.ConfigValue{
				"default": defaultProfile,
			},
			expect: &common.Config{
				Profile: "test",
			},
			err: errors.New(`loading profile "test" is failed: profile "test" is not exists`),
		},
		{
			scenario: "Profile name specified with normal profile",
			in: &common.Config{
				Profile:             "test",
				Zone:                defaults.Zone,
				Zones:               iaas.SakuraCloudZones,
				RetryMax:            defaults.RetryMax,
				APIRequestTimeout:   defaults.APIRequestTimeout,
				APIRequestRateLimit: defaults.APIRequestRateLimit,
			},
			profiles: map[string]*profile.ConfigValue{
				"default": defaultProfile,
				"test":    testProfile,
			},
			expect: &common.Config{
				Profile:             "test",
				AccessToken:         testProfile.AccessToken,
				AccessTokenSecret:   testProfile.AccessTokenSecret,
				Zone:                defaults.Zone,
				Zones:               iaas.SakuraCloudZones,
				TraceMode:           testProfile.TraceMode,
				AcceptLanguage:      testProfile.AcceptLanguage,
				APIRootURL:          testProfile.APIRootURL,
				RetryMax:            defaults.RetryMax,
				RetryWaitMin:        testProfile.RetryWaitMin,
				RetryWaitMax:        testProfile.RetryWaitMax,
				APIRequestTimeout:   defaults.APIRequestTimeout,
				APIRequestRateLimit: defaults.APIRequestRateLimit,
			},
		},
		{
			scenario: "only Profile name specified with normal profile",
			in: &common.Config{
				Profile: "test",
			},
			profiles: map[string]*profile.ConfigValue{
				"test": testProfile,
			},
			expect: &common.Config{
				Profile:             "test",
				AccessToken:         testProfile.AccessToken,
				AccessTokenSecret:   testProfile.AccessTokenSecret,
				Zone:                testProfile.Zone,
				Zones:               testProfile.Zones,
				TraceMode:           testProfile.TraceMode,
				AcceptLanguage:      testProfile.AcceptLanguage,
				APIRootURL:          testProfile.APIRootURL,
				RetryMax:            testProfile.RetryMax,
				RetryWaitMin:        testProfile.RetryWaitMin,
				RetryWaitMax:        testProfile.RetryWaitMax,
				APIRequestTimeout:   testProfile.HTTPRequestTimeout,
				APIRequestRateLimit: testProfile.HTTPRequestRateLimit,
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.scenario, func(t *testing.T) {
			initTestProfileDir()
			for profileName, profileValue := range tt.profiles {
				if err := profile.Save(profileName, profileValue); err != nil {
					t.Fatal(err)
				}
			}

			currentProfile := tt.currentProfile
			if tt.currentProfile == "" {
				currentProfile = profile.DefaultProfileName
			}
			if err := profile.SetCurrentName(currentProfile); err != nil {
				t.Fatal(err)
			}

			cfg, err := tt.in.LoadFromProfile()
			if err != nil {
				if tt.err.Error() != err.Error() {
					t.Errorf("got unexpected error: expected: %s got: %s", tt.err, err)
				}
			} else {
				tt.in.FillWith(cfg)
				require.EqualValues(t, tt.expect, tt.in)
			}
		})
	}
}

func TestFillWith_SPFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		base     common.Config
		other    common.Config
		wantSPID string
		wantKey  string
		wantPath string
	}{
		{
			name:     "empty base filled from other",
			base:     common.Config{},
			other:    common.Config{ServicePrincipalID: "sp-id", ServicePrincipalKeyID: "sp-key", ServicePrincipalPrivateKeyPath: "/key.pem"},
			wantSPID: "sp-id",
			wantKey:  "sp-key",
			wantPath: "/key.pem",
		},
		{
			name:     "base values not overridden",
			base:     common.Config{ServicePrincipalID: "base-id", ServicePrincipalKeyID: "base-key", ServicePrincipalPrivateKeyPath: "/base.pem"},
			other:    common.Config{ServicePrincipalID: "other-id", ServicePrincipalKeyID: "other-key", ServicePrincipalPrivateKeyPath: "/other.pem"},
			wantSPID: "base-id",
			wantKey:  "base-key",
			wantPath: "/base.pem",
		},
		{
			name:     "partial base filled partially",
			base:     common.Config{ServicePrincipalID: "base-id"},
			other:    common.Config{ServicePrincipalID: "other-id", ServicePrincipalKeyID: "other-key", ServicePrincipalPrivateKeyPath: "/other.pem"},
			wantSPID: "base-id",
			wantKey:  "other-key",
			wantPath: "/other.pem",
		},
		{
			name:     "both empty stays empty",
			base:     common.Config{},
			other:    common.Config{},
			wantSPID: "",
			wantKey:  "",
			wantPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.base.FillWith(&tt.other)
			assert.Equal(t, tt.wantSPID, tt.base.ServicePrincipalID)
			assert.Equal(t, tt.wantKey, tt.base.ServicePrincipalKeyID)
			assert.Equal(t, tt.wantPath, tt.base.ServicePrincipalPrivateKeyPath)
		})
	}
}

func TestFillWith_BackwardCompatibility_APIKeyFields(t *testing.T) {
	t.Parallel()

	// Verify existing API key fields still work correctly after SP field additions
	base := common.Config{
		AccessToken: "my-token",
	}
	other := common.Config{
		AccessToken:       "other-token",
		AccessTokenSecret: "other-secret",
		Zone:              "is1b",
	}

	base.FillWith(&other)

	assert.Equal(t, "my-token", base.AccessToken, "base AccessToken must not be overridden")
	assert.Equal(t, "other-secret", base.AccessTokenSecret, "empty AccessTokenSecret should be filled")
	assert.Equal(t, "is1b", base.Zone, "empty Zone should be filled")
	assert.Empty(t, base.ServicePrincipalID, "SP fields should remain empty when not set")
}

func TestFillWithDefault_DoesNotAffectSPFields(t *testing.T) {
	t.Parallel()

	cfg := common.Config{
		ServicePrincipalID:             "sp-id",
		ServicePrincipalKeyID:          "sp-key",
		ServicePrincipalPrivateKeyPath: "/key.pem",
	}
	cfg.FillWithDefault()

	assert.Equal(t, "sp-id", cfg.ServicePrincipalID, "FillWithDefault must not clear SP fields")
	assert.Equal(t, "sp-key", cfg.ServicePrincipalKeyID)
	assert.Equal(t, "/key.pem", cfg.ServicePrincipalPrivateKeyPath)
	assert.Equal(t, defaults.Zone, cfg.Zone, "Zone should get default")
	assert.Equal(t, defaults.RetryMax, cfg.RetryMax, "RetryMax should get default")
}
