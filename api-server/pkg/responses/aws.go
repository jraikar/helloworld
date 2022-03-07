package responses

// swagger:model
type AwsRegion struct {
	Name      string `json:"name,omitempty"`
	Region    string `json:"region,omitempty"`
	Continent string `json:"continent,omitempty"`
}

// swagger:model
type AwsInstanceType struct {
	Name    string   `json:"name,omitempty"`
	Cost    float32  `json:"cost,omitempty"`
	Flavors []string `json:"flavors,omitempty"`
}
