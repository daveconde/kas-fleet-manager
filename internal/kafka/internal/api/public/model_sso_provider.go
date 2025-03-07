/*
 * Kafka Service Fleet Manager
 *
 * Kafka Service Fleet Manager is a Rest API to manage Kafka instances.
 *
 * API version: 1.5.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package public

// SsoProvider SSO Provider
type SsoProvider struct {
	Id   string `json:"id,omitempty"`
	Kind string `json:"kind,omitempty"`
	Href string `json:"href,omitempty"`
	// base url
	BaseUrl     string `json:"base_url,omitempty"`
	TokenUrl    string `json:"token_url,omitempty"`
	Jwks        string `json:"jwks,omitempty"`
	ValidIssuer string `json:"valid_issuer,omitempty"`
}
