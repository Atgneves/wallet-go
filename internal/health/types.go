package health

type HealthStatus string

const (
	HealthStatusUp   HealthStatus = "UP"
	HealthStatusDown HealthStatus = "DOWN"
)

type Health struct {
	Status  HealthStatus           `json:"status"`
	Details map[string]interface{} `json:"details,omitempty"`
}

type ComponentHealth struct {
	Status HealthStatus `json:"status"`
	Error  string       `json:"error,omitempty"`
}
