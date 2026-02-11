// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

// Package legacybridge provides an HTTP transport bridge that delegates
// HTTP requests to saclient-go's Client.Do(). This allows legacy SDKs
// (api-client-go based) to use saclient-go's authentication mechanism
// without modifying the SDK source code.
//
// This is a transitional package. When each legacy SDK provides a native
// saclient-go entry point (e.g., NewClient(saclient.ClientAPI)), the
// corresponding bridge code should be removed and replaced with the
// SDK's native API.
//
// Migration targets:
//   - NewClient(saclient.ClientAPI) pattern (like SecretManager, SimpleMQ)
//   - NewClientFromSaclient(*saclient.Client) pattern (like IaaS)
//   - Saclient saclient.ClientAPI field pattern (like AppRun)
package legacybridge
