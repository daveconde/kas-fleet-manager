package config

import (
	"testing"
)

func TestKafkaSupportedSizesConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  SupportedKafkaInstanceTypesConfig
		wantErr bool
	}{
		{
			name: "Should not return an error with valid configuration",
			config: SupportedKafkaInstanceTypesConfig{
				SupportedKafkaInstanceTypes: []KafkaInstanceType{
					{
						Id:          "standard",
						DisplayName: "Standard",
						Sizes: []KafkaInstanceSize{
							{
								Id:                          "x1",
								IngressThroughputPerSec:     "30Mi",
								EgressThroughputPerSec:      "30Mi",
								TotalMaxConnections:         1000,
								MaxDataRetentionSize:        "100Gi",
								MaxPartitions:               1000,
								MaxDataRetentionPeriod:      "P14D",
								MaxConnectionAttemptsPerSec: 100,
								QuotaConsumed:               1,
								QuotaType:                   "rhosak",
								CapacityConsumed:            1,
							},
							{
								Id:                          "x2",
								IngressThroughputPerSec:     "60Mi",
								EgressThroughputPerSec:      "60Mi",
								TotalMaxConnections:         2000,
								MaxDataRetentionSize:        "200Gi",
								MaxPartitions:               2000,
								MaxDataRetentionPeriod:      "P14D",
								MaxConnectionAttemptsPerSec: 200,
								QuotaConsumed:               2,
								QuotaType:                   "rhosak",
								CapacityConsumed:            2,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Should fail because size was repeated",
			config: SupportedKafkaInstanceTypesConfig{
				SupportedKafkaInstanceTypes: []KafkaInstanceType{
					{
						Id:          "standard",
						DisplayName: "Standard",
						Sizes: []KafkaInstanceSize{
							{
								Id:                          "x1",
								IngressThroughputPerSec:     "30Mi",
								EgressThroughputPerSec:      "30Mi",
								TotalMaxConnections:         1000,
								MaxDataRetentionSize:        "100Gi",
								MaxPartitions:               1000,
								MaxDataRetentionPeriod:      "P14D",
								MaxConnectionAttemptsPerSec: 100,
								QuotaConsumed:               1,
								QuotaType:                   "rhosak",
								CapacityConsumed:            1,
							},
							{
								Id:                          "x1",
								IngressThroughputPerSec:     "30Mi",
								EgressThroughputPerSec:      "30Mi",
								TotalMaxConnections:         1000,
								MaxDataRetentionSize:        "100Gi",
								MaxPartitions:               1000,
								MaxDataRetentionPeriod:      "P14D",
								MaxConnectionAttemptsPerSec: 100,
								QuotaConsumed:               1,
								QuotaType:                   "rhosak",
								CapacityConsumed:            1,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Should fail because property TotalMaxConnections was not specified",
			config: SupportedKafkaInstanceTypesConfig{
				SupportedKafkaInstanceTypes: []KafkaInstanceType{
					{
						Id:          "standard",
						DisplayName: "Standard",
						Sizes: []KafkaInstanceSize{
							{
								Id:                          "x1",
								IngressThroughputPerSec:     "30Mi",
								EgressThroughputPerSec:      "30Mi",
								MaxDataRetentionSize:        "100Gi",
								MaxPartitions:               1000,
								MaxDataRetentionPeriod:      "P14D",
								MaxConnectionAttemptsPerSec: 100,
								QuotaConsumed:               1,
								QuotaType:                   "rhosak",
								CapacityConsumed:            1,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Should fail because property MaxPartitions was not specified",
			config: SupportedKafkaInstanceTypesConfig{
				SupportedKafkaInstanceTypes: []KafkaInstanceType{
					{
						Id:          "standard",
						DisplayName: "Standard",
						Sizes: []KafkaInstanceSize{
							{
								Id:                          "x1",
								IngressThroughputPerSec:     "30Mi",
								EgressThroughputPerSec:      "30Mi",
								TotalMaxConnections:         1000,
								MaxDataRetentionSize:        "100Gi",
								MaxDataRetentionPeriod:      "P14D",
								MaxConnectionAttemptsPerSec: 100,
								QuotaConsumed:               1,
								QuotaType:                   "rhosak",
								CapacityConsumed:            1,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Should fail because property MaxConnectionAttemptsPerSec was not specified",
			config: SupportedKafkaInstanceTypesConfig{
				SupportedKafkaInstanceTypes: []KafkaInstanceType{
					{
						Id:          "standard",
						DisplayName: "Standard",
						Sizes: []KafkaInstanceSize{
							{
								Id:                      "x1",
								IngressThroughputPerSec: "30Mi",
								EgressThroughputPerSec:  "30Mi",
								TotalMaxConnections:     1000,
								MaxDataRetentionSize:    "100Gi",
								MaxDataRetentionPeriod:  "P14D",
								MaxPartitions:           1000,
								QuotaConsumed:           1,
								QuotaType:               "rhosak",
								CapacityConsumed:        1,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Should return error when property IngressThroughputPerSec is undefined",
			config: SupportedKafkaInstanceTypesConfig{
				SupportedKafkaInstanceTypes: []KafkaInstanceType{
					{
						Id:          "standard",
						DisplayName: "Standard",
						Sizes: []KafkaInstanceSize{
							{
								Id:                          "x1",
								EgressThroughputPerSec:      "30Mi",
								TotalMaxConnections:         1000,
								MaxDataRetentionSize:        "100Gi",
								MaxPartitions:               1000,
								MaxDataRetentionPeriod:      "P14D",
								MaxConnectionAttemptsPerSec: 100,
								QuotaConsumed:               1,
								QuotaType:                   "rhosak",
								CapacityConsumed:            1,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Should return error when property EgressThroughputPerSec is undefined",
			config: SupportedKafkaInstanceTypesConfig{
				SupportedKafkaInstanceTypes: []KafkaInstanceType{
					{
						Id:          "standard",
						DisplayName: "Standard",
						Sizes: []KafkaInstanceSize{
							{
								Id:                          "x1",
								IngressThroughputPerSec:     "30Mi",
								TotalMaxConnections:         1000,
								MaxDataRetentionSize:        "100Gi",
								MaxPartitions:               1000,
								MaxDataRetentionPeriod:      "P14D",
								MaxConnectionAttemptsPerSec: 100,
								QuotaConsumed:               1,
								QuotaType:                   "rhosak",
								CapacityConsumed:            1,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Should return error when property MaxDataRetentionSize is undefined",
			config: SupportedKafkaInstanceTypesConfig{
				SupportedKafkaInstanceTypes: []KafkaInstanceType{
					{
						Id:          "standard",
						DisplayName: "Standard",
						Sizes: []KafkaInstanceSize{
							{
								Id:                          "x1",
								EgressThroughputPerSec:      "30Mi",
								IngressThroughputPerSec:     "30Mi",
								TotalMaxConnections:         1000,
								MaxPartitions:               1000,
								MaxDataRetentionPeriod:      "P14D",
								MaxConnectionAttemptsPerSec: 100,
								QuotaConsumed:               1,
								QuotaType:                   "rhosak",
								CapacityConsumed:            1,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Should return error when property MaxDataRetentionPeriod is undefined",
			config: SupportedKafkaInstanceTypesConfig{
				SupportedKafkaInstanceTypes: []KafkaInstanceType{
					{
						Id:          "standard",
						DisplayName: "Standard",
						Sizes: []KafkaInstanceSize{
							{
								Id:                          "x1",
								EgressThroughputPerSec:      "30Mi",
								IngressThroughputPerSec:     "30Mi",
								TotalMaxConnections:         1000,
								MaxPartitions:               1000,
								MaxConnectionAttemptsPerSec: 100,
								MaxDataRetentionSize:        "100Gi",
								QuotaConsumed:               1,
								QuotaType:                   "rhosak",
								CapacityConsumed:            1,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Should return error when property Id is undefined",
			config: SupportedKafkaInstanceTypesConfig{
				SupportedKafkaInstanceTypes: []KafkaInstanceType{
					{
						Id:          "standard",
						DisplayName: "Standard",
						Sizes: []KafkaInstanceSize{
							{
								EgressThroughputPerSec:      "30Mi",
								IngressThroughputPerSec:     "30Mi",
								TotalMaxConnections:         1000,
								MaxPartitions:               1000,
								MaxConnectionAttemptsPerSec: 100,
								MaxDataRetentionSize:        "100Gi",
								MaxDataRetentionPeriod:      "P14D",
								QuotaConsumed:               1,
								QuotaType:                   "rhosak",
								CapacityConsumed:            1,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Should return error when property with quantity format is invalid",
			config: SupportedKafkaInstanceTypesConfig{
				SupportedKafkaInstanceTypes: []KafkaInstanceType{
					{
						Id:          "standard",
						DisplayName: "Standard",
						Sizes: []KafkaInstanceSize{
							{
								Id:                          "x1",
								IngressThroughputPerSec:     "30Midk",
								EgressThroughputPerSec:      "30Mi",
								TotalMaxConnections:         1000,
								MaxDataRetentionSize:        "100Gi",
								MaxPartitions:               1000,
								MaxDataRetentionPeriod:      "P14D",
								MaxConnectionAttemptsPerSec: 100,
								QuotaConsumed:               1,
								QuotaType:                   "rhosak",
								CapacityConsumed:            1,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Should return error when property with quantity format is Zero",
			config: SupportedKafkaInstanceTypesConfig{
				SupportedKafkaInstanceTypes: []KafkaInstanceType{
					{
						Id:          "standard",
						DisplayName: "Standard",
						Sizes: []KafkaInstanceSize{
							{
								Id:                          "x1",
								IngressThroughputPerSec:     "30Mi",
								EgressThroughputPerSec:      "30Mi",
								TotalMaxConnections:         1000,
								MaxDataRetentionSize:        "0Gi",
								MaxPartitions:               1000,
								MaxDataRetentionPeriod:      "P14D",
								MaxConnectionAttemptsPerSec: 100,
								QuotaConsumed:               1,
								QuotaType:                   "rhosak",
								CapacityConsumed:            1,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Should return error when property with quantity format is less than Zero",
			config: SupportedKafkaInstanceTypesConfig{
				SupportedKafkaInstanceTypes: []KafkaInstanceType{
					{
						Id:          "standard",
						DisplayName: "Standard",
						Sizes: []KafkaInstanceSize{
							{
								Id:                          "x1",
								IngressThroughputPerSec:     "30Mi",
								EgressThroughputPerSec:      "-30Mi",
								TotalMaxConnections:         1000,
								MaxDataRetentionSize:        "100Gi",
								MaxPartitions:               1000,
								MaxDataRetentionPeriod:      "P14D",
								MaxConnectionAttemptsPerSec: 100,
								QuotaConsumed:               1,
								QuotaType:                   "rhosak",
								CapacityConsumed:            1,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Should return error when property MaxDataRetentionPeriod is invalid",
			config: SupportedKafkaInstanceTypesConfig{
				SupportedKafkaInstanceTypes: []KafkaInstanceType{
					{
						Id:          "standard",
						DisplayName: "Standard",
						Sizes: []KafkaInstanceSize{
							{
								Id:                          "x1",
								IngressThroughputPerSec:     "30Mi",
								EgressThroughputPerSec:      "30Mi",
								TotalMaxConnections:         1000,
								MaxDataRetentionSize:        "100Gi",
								MaxPartitions:               1000,
								MaxDataRetentionPeriod:      "P14Dygyuook",
								MaxConnectionAttemptsPerSec: 100,
								QuotaConsumed:               1,
								QuotaType:                   "rhosak",
								CapacityConsumed:            1,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Should return error when property MaxDataRetentionPeriod is zero",
			config: SupportedKafkaInstanceTypesConfig{
				SupportedKafkaInstanceTypes: []KafkaInstanceType{
					{
						Id:          "standard",
						DisplayName: "Standard",
						Sizes: []KafkaInstanceSize{
							{
								Id:                          "x1",
								IngressThroughputPerSec:     "30Mi",
								EgressThroughputPerSec:      "30Mi",
								TotalMaxConnections:         1000,
								MaxDataRetentionSize:        "100Gi",
								MaxPartitions:               1000,
								MaxDataRetentionPeriod:      "P0S",
								MaxConnectionAttemptsPerSec: 100,
								QuotaConsumed:               1,
								QuotaType:                   "rhosak",
								CapacityConsumed:            1,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Should return error when property KafkaProfile.id is undefined",
			config: SupportedKafkaInstanceTypesConfig{
				SupportedKafkaInstanceTypes: []KafkaInstanceType{
					{
						Sizes: []KafkaInstanceSize{
							{
								Id:                          "x1",
								IngressThroughputPerSec:     "30Mi",
								EgressThroughputPerSec:      "30Mi",
								TotalMaxConnections:         1000,
								MaxDataRetentionSize:        "100Gi",
								MaxPartitions:               1000,
								MaxDataRetentionPeriod:      "P14D",
								MaxConnectionAttemptsPerSec: 100,
								QuotaConsumed:               1,
								QuotaType:                   "rhosak",
								CapacityConsumed:            1,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Should return error when KafkaProfile.Sizes is empty",
			config: SupportedKafkaInstanceTypesConfig{
				SupportedKafkaInstanceTypes: []KafkaInstanceType{
					{
						Id:          "standard",
						DisplayName: "Standard",
						Sizes:       []KafkaInstanceSize{},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Should fail because profile was repeted",
			config: SupportedKafkaInstanceTypesConfig{
				SupportedKafkaInstanceTypes: []KafkaInstanceType{
					{
						Id:          "standard",
						DisplayName: "Standard",
						Sizes: []KafkaInstanceSize{
							{
								Id:                          "x1",
								IngressThroughputPerSec:     "30Mi",
								EgressThroughputPerSec:      "30Mi",
								TotalMaxConnections:         1000,
								MaxDataRetentionSize:        "100Gi",
								MaxPartitions:               1000,
								MaxDataRetentionPeriod:      "P14D",
								MaxConnectionAttemptsPerSec: 100,
								QuotaConsumed:               1,
								QuotaType:                   "rhosak",
								CapacityConsumed:            1,
							},
						},
					},
					{
						Id: "standard",
						Sizes: []KafkaInstanceSize{
							{
								Id:                          "x1",
								IngressThroughputPerSec:     "30Mi",
								EgressThroughputPerSec:      "30Mi",
								TotalMaxConnections:         1000,
								MaxDataRetentionSize:        "100Gi",
								MaxPartitions:               1000,
								MaxDataRetentionPeriod:      "P14D",
								MaxConnectionAttemptsPerSec: 100,
								QuotaConsumed:               1,
								QuotaType:                   "rhosak",
								CapacityConsumed:            1,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Should return an error when quota consumed is less than 1",
			config: SupportedKafkaInstanceTypesConfig{
				SupportedKafkaInstanceTypes: []KafkaInstanceType{
					{
						Id:          "standard",
						DisplayName: "Standard",
						Sizes: []KafkaInstanceSize{
							{
								Id:                          "x1",
								IngressThroughputPerSec:     "30Mi",
								EgressThroughputPerSec:      "30Mi",
								TotalMaxConnections:         1000,
								MaxDataRetentionSize:        "100Gi",
								MaxPartitions:               1000,
								MaxDataRetentionPeriod:      "P14D",
								MaxConnectionAttemptsPerSec: 100,
								QuotaConsumed:               -1,
								QuotaType:                   "rhosak",
								CapacityConsumed:            1,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Should return an error when quota consumed is undefined",
			config: SupportedKafkaInstanceTypesConfig{
				SupportedKafkaInstanceTypes: []KafkaInstanceType{
					{
						Id:          "standard",
						DisplayName: "Standard",
						Sizes: []KafkaInstanceSize{
							{
								Id:                          "x1",
								IngressThroughputPerSec:     "30Mi",
								EgressThroughputPerSec:      "30Mi",
								TotalMaxConnections:         1000,
								MaxDataRetentionSize:        "100Gi",
								MaxPartitions:               1000,
								MaxDataRetentionPeriod:      "P14D",
								MaxConnectionAttemptsPerSec: 100,
								QuotaType:                   "rhosak",
								CapacityConsumed:            1,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Should return an error when quota type is undefined",
			config: SupportedKafkaInstanceTypesConfig{
				SupportedKafkaInstanceTypes: []KafkaInstanceType{
					{
						Id:          "standard",
						DisplayName: "Standard",
						Sizes: []KafkaInstanceSize{
							{
								Id:                          "x1",
								IngressThroughputPerSec:     "30Mi",
								EgressThroughputPerSec:      "30Mi",
								TotalMaxConnections:         1000,
								MaxDataRetentionSize:        "100Gi",
								MaxPartitions:               1000,
								MaxDataRetentionPeriod:      "P14D",
								MaxConnectionAttemptsPerSec: 100,
								QuotaConsumed:               1,
								CapacityConsumed:            1,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Should return an error when capacity consumed is undefined",
			config: SupportedKafkaInstanceTypesConfig{
				SupportedKafkaInstanceTypes: []KafkaInstanceType{
					{
						Id:          "standard",
						DisplayName: "Standard",
						Sizes: []KafkaInstanceSize{
							{
								Id:                          "x1",
								IngressThroughputPerSec:     "30Mi",
								EgressThroughputPerSec:      "30Mi",
								TotalMaxConnections:         1000,
								MaxDataRetentionSize:        "100Gi",
								MaxPartitions:               1000,
								MaxDataRetentionPeriod:      "P14D",
								MaxConnectionAttemptsPerSec: 100,
								QuotaConsumed:               1,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Should return an error when capacity consumed is less than 1",
			config: SupportedKafkaInstanceTypesConfig{
				SupportedKafkaInstanceTypes: []KafkaInstanceType{
					{
						Id:          "standard",
						DisplayName: "Standard",
						Sizes: []KafkaInstanceSize{
							{
								Id:                          "x1",
								IngressThroughputPerSec:     "30Mi",
								EgressThroughputPerSec:      "30Mi",
								TotalMaxConnections:         1000,
								MaxDataRetentionSize:        "100Gi",
								MaxPartitions:               1000,
								MaxDataRetentionPeriod:      "P14D",
								MaxConnectionAttemptsPerSec: 100,
								QuotaConsumed:               1,
								CapacityConsumed:            0,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Should return an error when profile id is invalid",
			config: SupportedKafkaInstanceTypesConfig{
				SupportedKafkaInstanceTypes: []KafkaInstanceType{
					{
						Id:          "invalid",
						DisplayName: "Invalid",
						Sizes: []KafkaInstanceSize{
							{
								Id:                          "x1",
								IngressThroughputPerSec:     "30Mi",
								EgressThroughputPerSec:      "30Mi",
								TotalMaxConnections:         1000,
								MaxDataRetentionSize:        "100Gi",
								MaxPartitions:               1000,
								MaxDataRetentionPeriod:      "P14D",
								MaxConnectionAttemptsPerSec: 100,
								QuotaConsumed:               1,
								QuotaType:                   "rhosak",
								CapacityConsumed:            1,
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.validate(); (err != nil) != tt.wantErr {
				t.Errorf("SupportedKafkaSizesConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
