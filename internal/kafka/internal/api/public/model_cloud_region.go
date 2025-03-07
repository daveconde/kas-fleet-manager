/*
 * Kafka Service Fleet Manager
 *
 * Kafka Service Fleet Manager is a Rest API to manage Kafka instances.
 *
 * API version: 1.5.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package public

// CloudRegion Description of a region of a cloud provider.
type CloudRegion struct {
	// Indicates the type of this object. Will be 'CloudRegion'.
	Kind string `json:"kind,omitempty"`
	// Unique identifier of the object.
	Id string `json:"id,omitempty"`
	// Name of the region for display purposes, for example `N. Virginia`.
	DisplayName string `json:"display_name,omitempty"`
	// Whether the region is enabled for deploying an OSD cluster.
	Enabled bool `json:"enabled"`
	// The Kafka instance types supported by this region.  DEPRECATION NOTICE - instance_type will be deprecated
	// Deprecated
	DeprecatedSupportedInstanceTypes []string `json:"supported_instance_types"`
	// Indicates whether there is capacity left per instance type
	Capacity []RegionCapacityListItem `json:"capacity"`
}
