package detect

// DockerfileDetector notes the presence of Dockerfiles. It never parses or executes
// their contents.
type DockerfileDetector struct{}

func (DockerfileDetector) Name() string { return "dockerfile" }

func (d DockerfileDetector) Detect(ctx Context) ([]Finding, error) {
	files := findFilesNamed(ctx.FS, "Dockerfile")
	if len(files) == 0 {
		return nil, nil
	}
	// A single technology finding is enough; the build contexts are already implied
	// by the components and any Compose services.
	return []Finding{{Kind: KindTechnology, Detector: d.Name(), Value: "Dockerfile"}}, nil
}
