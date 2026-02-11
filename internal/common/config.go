// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	client "github.com/sacloud/api-client-go"
	"github.com/sacloud/api-client-go/profile"
	"github.com/sacloud/apigw-api-go"
	apigwapi "github.com/sacloud/apigw-api-go/apis/v1"
	"github.com/sacloud/apprun-api-go"
	dedicatedstorage "github.com/sacloud/dedicated-storage-api-go"
	dedicatedstorageapi "github.com/sacloud/dedicated-storage-api-go/apis/v1"
	"github.com/sacloud/eventbus-api-go"
	eventbus_api "github.com/sacloud/eventbus-api-go/apis/v1"
	saht "github.com/sacloud/go-http"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/query"
	"github.com/sacloud/iaas-api-go/trace"
	kms "github.com/sacloud/kms-api-go"
	kmsapi "github.com/sacloud/kms-api-go/apis/v1"
	nosql "github.com/sacloud/nosql-api-go"
	nosqlapi "github.com/sacloud/nosql-api-go/apis/v1"
	objectstorage "github.com/sacloud/object-storage-api-go"
	"github.com/sacloud/saclient-go"
	sm "github.com/sacloud/secretmanager-api-go"
	smapi "github.com/sacloud/secretmanager-api-go/apis/v1"
	seccon "github.com/sacloud/security-control-api-go"
	secconapi "github.com/sacloud/security-control-api-go/apis/v1"
	"github.com/sacloud/simplemq-api-go"
	"github.com/sacloud/simplemq-api-go/apis/v1/queue"
	"github.com/sacloud/terraform-provider-sakura/internal/defaults"
	"github.com/sacloud/terraform-provider-sakura/internal/legacybridge"
	ver "github.com/sacloud/terraform-provider-sakura/version"
)

const (
	traceHTTP = "http"
	traceAPI  = "api"
)

const uaEnvVar = "SAKURACLOUD_APPEND_USER_AGENT"

const (
	Zone                = "is1b"
	RetryMax            = 10
	APIRequestTimeout   = 300
	APIRequestRateLimit = 10
)

var (
	deletionWaiterTimeout            = 30 * time.Minute
	deletionWaiterPollingInterval    = 5 * time.Second
	databaseWaitAfterCreateDuration  = 1 * time.Minute
	vpcRouterWaitAfterCreateDuration = 2 * time.Minute
)

// Config type of SakuraCloud Config
type Config struct {
	Profile             string
	AccessToken         string
	AccessTokenSecret   string
	Zone                string
	Zones               []string
	DefaultZone         string
	TraceMode           string
	AcceptLanguage      string
	APIRootURL          string
	RetryMax            int
	RetryWaitMin        int
	RetryWaitMax        int
	APIRequestTimeout   int
	APIRequestRateLimit int
	TerraformVersion    string

	ServicePrincipalID             string
	ServicePrincipalKeyID          string
	ServicePrincipalPrivateKeyPath string
}

// APIClient for SakuraCloud API
type APIClient struct {
	iaas.APICaller
	defaultZone                      string // 各リソースでzone未指定の場合に利用するゾーン。iaas.APIDefaultZoneとは別物。
	zones                            []string
	deletionWaiterTimeout            time.Duration
	deletionWaiterPollingInterval    time.Duration
	databaseWaitAfterCreateDuration  time.Duration
	vpcRouterWaitAfterCreateDuration time.Duration
	CallerOptions                    *client.Options
	AppRunClient                     *apprun.Client
	KmsClient                        *kmsapi.Client
	SecretManagerClient              *smapi.Client
	SimpleMqClient                   *queue.Client
	EventBusClient                   *eventbus_api.Client
	ObjectStorageClient              *objectstorage.Client
	NosqlClient                      *nosqlapi.Client
	DedicatedStorageClient           *dedicatedstorageapi.Client
	ApigwClient                      *apigwapi.Client
	SecurityControlClient            *secconapi.Client
}

func (c *APIClient) CheckReferencedOption() query.CheckReferencedOption {
	return query.CheckReferencedOption{
		Tick:    c.deletionWaiterPollingInterval,
		Timeout: c.deletionWaiterTimeout,
	}
}

func (c *APIClient) GetZones() []string {
	return c.zones
}

func (c *Config) FillWith(other *Config) {
	if c.AccessToken == "" {
		c.AccessToken = other.AccessToken
	}
	if c.AccessTokenSecret == "" {
		c.AccessTokenSecret = other.AccessTokenSecret
	}
	if c.Zone == "" {
		c.Zone = other.Zone
	}
	if c.DefaultZone == "" {
		c.DefaultZone = other.DefaultZone
	}
	if len(c.Zones) == 0 {
		c.Zones = other.Zones
	}
	if c.APIRootURL == "" {
		c.APIRootURL = other.APIRootURL
	}
	if c.TraceMode == "" {
		c.TraceMode = other.TraceMode
	}
	if c.AcceptLanguage == "" {
		c.AcceptLanguage = other.AcceptLanguage
	}
	if c.RetryMax == 0 && other.RetryMax > 0 {
		c.RetryMax = other.RetryMax
	}
	if c.RetryWaitMax == 0 {
		c.RetryWaitMax = other.RetryWaitMax
	}
	if c.RetryWaitMin == 0 {
		c.RetryWaitMin = other.RetryWaitMin
	}
	if c.APIRequestTimeout == 0 && other.APIRequestTimeout > 0 {
		c.APIRequestTimeout = other.APIRequestTimeout
	}
	if c.APIRequestRateLimit == 0 && other.APIRequestRateLimit > 0 {
		c.APIRequestRateLimit = other.APIRequestRateLimit
	}
	if c.ServicePrincipalID == "" {
		c.ServicePrincipalID = other.ServicePrincipalID
	}
	if c.ServicePrincipalKeyID == "" {
		c.ServicePrincipalKeyID = other.ServicePrincipalKeyID
	}
	if c.ServicePrincipalPrivateKeyPath == "" {
		c.ServicePrincipalPrivateKeyPath = other.ServicePrincipalPrivateKeyPath
	}
}

func (c *Config) FillWithDefault() {
	if c.Zone == "" {
		c.Zone = defaults.Zone
	}
	if len(c.Zones) == 0 {
		c.Zones = iaas.SakuraCloudZones
	}
	if c.RetryMax == 0 {
		c.RetryMax = defaults.RetryMax
	}
	if c.APIRequestTimeout == 0 {
		c.APIRequestTimeout = defaults.APIRequestTimeout
	}
	if c.APIRequestRateLimit == 0 {
		c.APIRequestRateLimit = defaults.APIRequestRateLimit
	}
}

func (c *Config) LoadFromProfile() (*Config, error) {
	if c.Profile == "" {
		if name, err := profile.CurrentName(); err != nil {
			c.Profile = profile.DefaultProfileName
		} else {
			c.Profile = name
		}
	}
	if c.Profile != profile.DefaultProfileName {
		log.Printf("[DEBUG] using profile %q", c.Profile)
	}

	pcv := &profile.ConfigValue{}
	if err := profile.Load(c.Profile, pcv); err != nil {
		return nil, fmt.Errorf("loading profile %q is failed: %s", c.Profile, err)
	}

	return &Config{
		AccessToken:         pcv.AccessToken,
		AccessTokenSecret:   pcv.AccessTokenSecret,
		Zone:                pcv.Zone,
		Zones:               pcv.Zones,
		TraceMode:           pcv.TraceMode,
		AcceptLanguage:      pcv.AcceptLanguage,
		APIRootURL:          pcv.APIRootURL,
		RetryMax:            pcv.RetryMax,
		RetryWaitMin:        pcv.RetryWaitMin,
		RetryWaitMax:        pcv.RetryWaitMax,
		APIRequestTimeout:   pcv.HTTPRequestTimeout,
		APIRequestRateLimit: pcv.HTTPRequestRateLimit,
	}, nil
}

func (c *Config) validate() error {
	hasToken := c.AccessToken != ""
	hasSecret := c.AccessTokenSecret != ""
	apiKeyComplete := hasToken && hasSecret

	hasSPID := c.ServicePrincipalID != ""
	hasSPKeyID := c.ServicePrincipalKeyID != ""
	hasSPKeyPath := c.ServicePrincipalPrivateKeyPath != ""
	spComplete := hasSPID && hasSPKeyID && hasSPKeyPath

	if apiKeyComplete || spComplete {
		// At least one auth method is complete; check for partial config in the other
		var err error
		if !apiKeyComplete && (hasToken || hasSecret) {
			if !hasToken {
				err = multierror.Append(err, errors.New("AccessToken is set but AccessTokenSecret is missing"))
			}
			if !hasSecret {
				err = multierror.Append(err, errors.New("AccessTokenSecret is set but AccessToken is missing"))
			}
		}
		if !spComplete && (hasSPID || hasSPKeyID || hasSPKeyPath) {
			if !hasSPID {
				err = multierror.Append(err, errors.New("ServicePrincipalKeyID or ServicePrincipalPrivateKeyPath is set but ServicePrincipalID is missing"))
			}
			if !hasSPKeyID {
				err = multierror.Append(err, errors.New("ServicePrincipalID or ServicePrincipalPrivateKeyPath is set but ServicePrincipalKeyID is missing"))
			}
			if !hasSPKeyPath {
				err = multierror.Append(err, errors.New("ServicePrincipalID or ServicePrincipalKeyID is set but ServicePrincipalPrivateKeyPath is missing"))
			}
		}
		return err
	}

	// Neither is complete
	var err error
	if hasToken || hasSecret {
		// Partial API key
		if !hasToken {
			err = multierror.Append(err, errors.New("AccessTokenSecret is set but AccessToken is missing"))
		}
		if !hasSecret {
			err = multierror.Append(err, errors.New("AccessToken is set but AccessTokenSecret is missing"))
		}
	}
	if hasSPID || hasSPKeyID || hasSPKeyPath {
		// Partial SP
		if !hasSPID {
			err = multierror.Append(err, errors.New("ServicePrincipalKeyID or ServicePrincipalPrivateKeyPath is set but ServicePrincipalID is missing"))
		}
		if !hasSPKeyID {
			err = multierror.Append(err, errors.New("ServicePrincipalID or ServicePrincipalPrivateKeyPath is set but ServicePrincipalKeyID is missing"))
		}
		if !hasSPKeyPath {
			err = multierror.Append(err, errors.New("ServicePrincipalID or ServicePrincipalKeyID is set but ServicePrincipalPrivateKeyPath is missing"))
		}
	}
	if err == nil {
		// Nothing set at all
		err = multierror.Append(err, errors.New("AccessToken/AccessTokenSecret or ServicePrincipal credentials are required"))
	}
	return err
}

func (c *Config) NewClient(envConf *Config, theClient *saclient.Client) (*APIClient, error) {
	if profileConf, err := c.LoadFromProfile(); err != nil {
		return nil, err
	} else {
		// 設定の優先度: tfファイル > 環境変数 > profile > プロバイダのデフォルト
		// ref: https://docs.usacloud.jp/terraform/provider/#api
		c.FillWith(envConf)
		c.FillWith(profileConf)
		c.FillWithDefault()
	}
	if err := c.validate(); err != nil {
		return nil, err
	}

	tfUserAgent := terraformUserAgent(c.TerraformVersion)
	providerUserAgent := fmt.Sprintf("%s/v%s", "terraform-provider-sakura", ver.Version)
	ua := fmt.Sprintf("%s %s", tfUserAgent, providerUserAgent)
	if add := os.Getenv(uaEnvVar); add != "" {
		ua += " " + add
		log.Printf("[DEBUG] Using modified User-Agent: %s", ua)
	}

	enableAPITrace := false
	enableHTTPTrace := false
	if c.TraceMode != "" {
		enableAPITrace = true
		enableHTTPTrace = true
		mode := strings.ToLower(c.TraceMode)
		switch mode {
		case traceAPI:
			enableHTTPTrace = false
		case traceHTTP:
			enableAPITrace = false
		}
	}

	// saclient-go: Populate to finalize settings
	if err := theClient.Populate(); err != nil {
		return nil, fmt.Errorf("populating saclient: %w", err)
	}

	// IaaS: saclient-go native via NewClientFromSaclient
	iaasClient := iaas.NewClientFromSaclient(theClient)

	// IaaS global side-effects (matching helper/api.newCaller behavior)
	if c.DefaultZone != "" {
		iaas.APIDefaultZone = c.DefaultZone
	}
	if len(c.Zones) > 0 {
		iaas.SakuraCloudZones = c.Zones
	}
	if c.APIRootURL != "" {
		apiRoot := c.APIRootURL
		if strings.HasSuffix(apiRoot, "/") {
			apiRoot = strings.TrimRight(apiRoot, "/")
		}
		iaas.SakuraCloudAPIRoot = apiRoot
	}
	if enableAPITrace {
		trace.AddClientFactoryHooks()
	}

	zones := c.Zones
	if len(zones) == 0 {
		zones = iaas.SakuraCloudZones
	}

	// EventBus: existing callerOptions (API key auth only, NOT bridged)
	callerOptions := &client.Options{
		AccessToken:          c.AccessToken,
		AccessTokenSecret:    c.AccessTokenSecret,
		AcceptLanguage:       c.AcceptLanguage,
		HttpRequestTimeout:   c.APIRequestTimeout,
		HttpRequestRateLimit: c.APIRequestRateLimit,
		RetryMax:             c.RetryMax,
		RetryWaitMax:         c.RetryWaitMax,
		RetryWaitMin:         c.RetryWaitMin,
		UserAgent:            ua,
		Trace:                enableHTTPTrace,
	}

	// Bridge for legacy SDKs (KMS, NoSQL, APIGW, ObjectStorage)
	bridgeHTTPClient := legacybridge.NewHTTPClient(theClient)
	bridgeCallerOptions := &client.Options{
		HttpClient:        bridgeHTTPClient,
		AccessToken:       "",
		AccessTokenSecret: "",
		RetryMax:          1, // 1 = no retry (0 resets to default 10)
		UserAgent:         ua,
		AcceptLanguage:    c.AcceptLanguage,
	}
	bridgeCallerOptionsWithoutBigInt := &client.Options{
		HttpClient:        bridgeHTTPClient,
		AccessToken:       "",
		AccessTokenSecret: "",
		RetryMax:          1,
		UserAgent:         ua,
		AcceptLanguage:    c.AcceptLanguage,
		RequestCustomizers: []saht.RequestCustomizer{
			func(req *http.Request) error {
				req.Header.Set("X-Sakura-Bigint-As-Int", "0")
				return nil
			}},
	}

	// Legacy SDKs: bridge
	kmsClient, err := kms.NewClient(client.WithOptions(bridgeCallerOptions))
	if err != nil {
		return nil, err
	}
	nosqlClient, err := nosql.NewClient(client.WithOptions(bridgeCallerOptions))
	if err != nil {
		return nil, err
	}
	apigwClient, err := apigw.NewClient(client.WithOptions(bridgeCallerOptions))
	if err != nil {
		return nil, err
	}

	// EventBus: existing callerOptions (SP認証ブリッジの対象外)
	eventbusClient, err := eventbus.NewClient(client.WithOptions(callerOptions))
	if err != nil {
		return nil, err
	}

	// saclient-go native SDKs
	smClient, err := sm.NewClient(theClient)
	if err != nil {
		return nil, err
	}
	simplemqClient, err := simplemq.NewQueueClient(theClient)
	if err != nil {
		return nil, err
	}
	dedicatedStorageClient, err := dedicatedstorage.NewClient(theClient)
	if err != nil {
		return nil, err
	}
	secconClient, err := seccon.NewClient(theClient)
	if err != nil {
		return nil, err
	}

	return &APIClient{
		APICaller:                        iaasClient,
		defaultZone:                      c.Zone,
		zones:                            zones,
		deletionWaiterTimeout:            deletionWaiterTimeout,
		deletionWaiterPollingInterval:    deletionWaiterPollingInterval,
		databaseWaitAfterCreateDuration:  databaseWaitAfterCreateDuration,
		vpcRouterWaitAfterCreateDuration: vpcRouterWaitAfterCreateDuration,
		CallerOptions:                    bridgeCallerOptions,
		KmsClient:                        kmsClient,
		SecretManagerClient:              smClient,
		SimpleMqClient:                   simplemqClient,
		EventBusClient:                   eventbusClient,
		AppRunClient:                     &apprun.Client{Saclient: theClient},
		ObjectStorageClient:              &objectstorage.Client{Options: bridgeCallerOptionsWithoutBigInt},
		NosqlClient:                      nosqlClient,
		DedicatedStorageClient:           dedicatedStorageClient,
		ApigwClient:                      apigwClient,
		SecurityControlClient:            secconClient,
	}, nil
}

// HasBothAuthMethods returns true when both API key and SP credentials are fully configured.
func (c *Config) HasBothAuthMethods() bool {
	apiKeyComplete := c.AccessToken != "" && c.AccessTokenSecret != ""
	spComplete := c.ServicePrincipalID != "" && c.ServicePrincipalKeyID != "" && c.ServicePrincipalPrivateKeyPath != ""
	return apiKeyComplete && spComplete
}

const tfUAEnvVar = "TF_APPEND_USER_AGENT"

func terraformUserAgent(version string) string {
	ua := fmt.Sprintf("HashiCorp Terraform/%s (+https://www.terraform.io)", version)

	if add := os.Getenv(tfUAEnvVar); add != "" {
		add = strings.TrimSpace(add)
		if len(add) > 0 {
			ua += " " + add
			log.Printf("[DEBUG] Using modified User-Agent: %s", ua)
		}
	}

	return ua
}
