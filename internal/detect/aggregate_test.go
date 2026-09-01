package detect

import (
	"testing"

	"github.com/c3nk/omadev/internal/project"
)

func tech(value, component string) Finding {
	return Finding{Kind: KindTechnology, Value: value, Data: map[string]string{dataComponent: component}}
}

func TestAggregate_ComposeHigh(t *testing.T) {
	findings := []Finding{
		{Kind: KindExecution, Value: "Docker Compose"},
		{Kind: KindService, Value: "postgres", Data: map[string]string{dataImage: "postgres:17", dataHealth: "true", dataRole: "database"}},
		{Kind: KindService, Value: "backend"},
		{Kind: KindPort, Data: map[string]string{dataService: "backend", dataPublished: "8000", dataTarget: "8000"}},
	}

	p := Aggregate("/repo", "example", findings)

	if p.Confidence != project.ConfidenceHigh {
		t.Errorf("confidence = %v, want HIGH", p.Confidence)
	}
	if p.ExecutionStrategy != project.ExecutionCompose {
		t.Errorf("strategy = %v, want Compose", p.ExecutionStrategy)
	}
	if len(p.Services) != 2 {
		t.Fatalf("services = %d, want 2", len(p.Services))
	}
	if !p.Services[0].HasHealth || p.Services[0].Role != "database" {
		t.Errorf("postgres service metadata not carried: %+v", p.Services[0])
	}
	if len(p.Services[1].Ports) != 1 || p.Services[1].Ports[0].Published != 8000 {
		t.Errorf("backend port not attached: %+v", p.Services[1].Ports)
	}
	if len(p.URLs) != 1 || p.URLs[0].URL != "http://localhost:8000" {
		t.Errorf("url not derived from published port: %+v", p.URLs)
	}
}

func TestAggregate_AmbiguityBlocksHigh(t *testing.T) {
	findings := []Finding{
		{Kind: KindExecution, Value: "Docker Compose"},
		{Kind: KindService, Value: "web"},
		{Kind: KindWarning, Value: "multiple override files", Data: map[string]string{dataAmbiguous: "true"}},
	}

	p := Aggregate("/repo", "x", findings)

	if p.Confidence == project.ConfidenceHigh {
		t.Error("ambiguity must prevent HIGH confidence")
	}
	if p.Confidence != project.ConfidenceMedium {
		t.Errorf("confidence = %v, want MEDIUM", p.Confidence)
	}
	if len(p.Notes) != 1 {
		t.Errorf("warning should be recorded as a note: %+v", p.Notes)
	}
}

func TestAggregate_NonComposeMedium(t *testing.T) {
	findings := []Finding{
		tech("React", "frontend"),
		tech("TypeScript", "frontend"),
		tech("FastAPI", "backend"),
	}

	p := Aggregate("/repo", "mono", findings)

	if p.Confidence != project.ConfidenceMedium {
		t.Errorf("confidence = %v, want MEDIUM", p.Confidence)
	}
	if len(p.Components) != 2 {
		t.Fatalf("components = %d, want 2", len(p.Components))
	}
	if len(p.Components[0].Technology) != 2 {
		t.Errorf("frontend technologies = %v, want 2", p.Components[0].Technology)
	}
}

func TestAggregate_UnknownLow(t *testing.T) {
	p := Aggregate("/repo", "x", nil)
	if p.Confidence != project.ConfidenceLow {
		t.Errorf("confidence = %v, want LOW for no findings", p.Confidence)
	}
}
