//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package reqres

import (
	"github.com/spiffe/spike-sdk-go/api/entity/data"
)

// PolicyCreateRequest for policy creation.
type PolicyCreateRequest struct {
	Name            string                  `json:"name"`
	SpiffeIdPattern string                  `json:"spiffedPattern"`
	PathPattern     string                  `json:"pathPattern"`
	Permissions     []data.PolicyPermission `json:"permissions"`
}

// PolicyCreateResponse for policy creation.
type PolicyCreateResponse struct {
	Id  string         `json:"id,omitempty"`
	Err data.ErrorCode `json:"err,omitempty"`
}

// PolicyReadRequest to read a policy.
type PolicyReadRequest struct {
	Id string `json:"id"`
}

// PolicyReadResponse to read a policy.
type PolicyReadResponse struct {
	data.Policy
	Err data.ErrorCode `json:"err,omitempty"`
}

// PolicyDeleteRequest to delete a policy.
type PolicyDeleteRequest struct {
	Id string `json:"id"`
}

// PolicyDeleteResponse to delete a policy.
type PolicyDeleteResponse struct {
	Err data.ErrorCode `json:"err,omitempty"`
}

// PolicyListRequest to list policies.
type PolicyListRequest struct {
	SpiffeIdPattern string `json:"spiffedPattern"`
	PathPattern     string `json:"pathPattern"`
}

// PolicyListResponse to list policies.
type PolicyListResponse struct {
	Policies []data.Policy  `json:"policies"`
	Err      data.ErrorCode `json:"err,omitempty"`
}

// PolicyAccessCheckRequest to validate policy access.
type PolicyAccessCheckRequest struct {
	SpiffeId string `json:"spiffeId"`
	Path     string `json:"path"`
	Action   string `json:"action"`
}

// PolicyAccessCheckResponse to validate policy access,.
type PolicyAccessCheckResponse struct {
	Allowed          bool           `json:"allowed"`
	MatchingPolicies []string       `json:"matchingPolicies"`
	Err              data.ErrorCode `json:"err,omitempty"`
}
