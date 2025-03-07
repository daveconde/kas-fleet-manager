/*
 * Connector Service Fleet Manager Admin APIs
 *
 * Connector Service Fleet Manager Admin is a Rest API to manage connector clusters.
 *
 * API version: 0.0.3
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package private

import (
	"time"
)

// ConnectorNamespace A connector namespace
type ConnectorNamespace struct {
	Id              string                                     `json:"id"`
	Kind            string                                     `json:"kind,omitempty"`
	Href            string                                     `json:"href,omitempty"`
	Owner           string                                     `json:"owner,omitempty"`
	CreatedAt       time.Time                                  `json:"created_at,omitempty"`
	ModifiedAt      time.Time                                  `json:"modified_at,omitempty"`
	Name            string                                     `json:"name"`
	Annotations     []ConnectorNamespaceRequestMetaAnnotations `json:"annotations,omitempty"`
	ResourceVersion int64                                      `json:"resource_version"`
	Quota           ConnectorNamespaceQuota                    `json:"quota,omitempty"`
	ClusterId       string                                     `json:"cluster_id"`
	// Namespace expiration timestamp in RFC 3339 format
	Expiration string                   `json:"expiration,omitempty"`
	Tenant     ConnectorNamespaceTenant `json:"tenant"`
	Status     ConnectorNamespaceStatus `json:"status"`
}
