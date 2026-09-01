package detect

// DefaultRegistry returns the standard set of detectors in a stable order.
func DefaultRegistry() *Registry {
	return NewRegistry(
		ComposeDetector{},
		DockerfileDetector{},
		NodeDetector{},
		PythonDetector{},
		PostgresDetector{},
		MiseDetector{},
	)
}
