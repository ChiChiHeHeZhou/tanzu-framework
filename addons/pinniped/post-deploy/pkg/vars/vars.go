// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package vars contains variables used throughout the codebase.
package vars

import "github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/constants"

var (
	// PinnipedAPIGroupSuffix is the API group suffix used to talk to Pinniped APIs
	PinnipedAPIGroupSuffix = constants.PinnipedDefaultAPIGroupSuffix

	// ConciergeIsClusterScoped indicates whether the Pinniped Concierge APIs are
	// cluster-scoped (as opposed to namespace-scoped).
	ConciergeIsClusterScoped = false

	// SupervisorNamespace is the supervisor service namespace.
	SupervisorNamespace = "pinniped-supervisor"

	// SupervisorSvcName is the supervisor service name.
	SupervisorSvcName = ""

	// ConciergeNamespace is the Concierge namespace.
	ConciergeNamespace = "pinniped-concierge"

	// DexNamespace is the Dex namespace.
	DexNamespace = "tanzu-system-auth"

	// DexSvcName is the Dex service name.
	DexSvcName = "dexsvc"

	// DexCertName is the Dex cert name.
	DexCertName = "dex-cert-tls"

	// DexConfigMapName is the Dex ConfigMap name.
	DexConfigMapName = "dex"

	// PinnipedOIDCProviderName is the Dex OIDC identity provider.
	PinnipedOIDCProviderName = "upstream-oidc-identity-provider"

	// PinnipedOIDCClientSecretName is the Dex client credentials.
	PinnipedOIDCClientSecretName = "upstream-idp-client-credentials" //nolint:gosec

	// SupervisorSvcEndpoint is the supervisor service endpoint.
	SupervisorSvcEndpoint = ""

	// FederationDomainName is the federation domain name.
	FederationDomainName = ""

	// JWTAuthenticatorName is the JWT authentication name.
	JWTAuthenticatorName = ""

	// SupervisorCertName is the supervisor cert name.
	SupervisorCertName = ""

	// SupervisorCABundleData is the supervisor CA bundle data.
	SupervisorCABundleData = ""

	// CustomTLSSecretName is the the name of custom provided TLS secret if they don't want to use the self-signed cert generated by cert manager
	CustomTLSSecretName = ""

	// IsDexRequired is flag indicates if configuring dex is required
	IsDexRequired = false
)
