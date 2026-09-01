package detect

import (
	"testing"
)

// runDetector opens a fixture and runs a single detector against it.
func runDetector(t *testing.T, fixture string, d Detector) []Finding {
	t.Helper()
	ctx, root, err := OpenContext("../../testdata/" + fixture)
	if err != nil {
		t.Fatalf("open fixture %s: %v", fixture, err)
	}
	defer root.Close()
	findings, err := d.Detect(ctx)
	if err != nil {
		t.Fatalf("detect %s: %v", fixture, err)
	}
	return findings
}

func has(findings []Finding, kind FindingKind, value string) bool {
	for _, f := range findings {
		if f.Kind == kind && f.Value == value {
			return true
		}
	}
	return false
}

func serviceFinding(findings []Finding, name string) (Finding, bool) {
	for _, f := range findings {
		if f.Kind == KindService && f.Value == name {
			return f, true
		}
	}
	return Finding{}, false
}

func TestComposeDetector_FastapiReact(t *testing.T) {
	f := runDetector(t, "docker-compose-fastapi-react", ComposeDetector{})

	if !has(f, KindTechnology, "Docker Compose") {
		t.Error("expected Docker Compose technology finding")
	}
	if !has(f, KindExecution, "Docker Compose") {
		t.Error("expected Docker Compose execution strategy")
	}
	if !has(f, KindTechnology, "PostgreSQL") {
		t.Error("expected PostgreSQL technology from postgres image")
	}
	for _, svc := range []string{"frontend", "backend", "postgres"} {
		if !has(f, KindService, svc) {
			t.Errorf("expected service %q", svc)
		}
	}

	pg, ok := serviceFinding(f, "postgres")
	if !ok {
		t.Fatal("missing postgres service")
	}
	if pg.Data[dataRole] != "database" {
		t.Errorf("postgres role = %q, want database", pg.Data[dataRole])
	}
	if pg.Data[dataHealth] != "true" {
		t.Errorf("postgres should report a healthcheck")
	}

	// Ports parsed from short form.
	wantPorts := map[string]string{"frontend": "5173", "backend": "8000", "postgres": "5432"}
	for _, f2 := range f {
		if f2.Kind == KindPort {
			svc := f2.Data[dataService]
			if wantPorts[svc] != "" && f2.Data[dataPublished] != wantPorts[svc] {
				t.Errorf("service %q published port = %q, want %q", svc, f2.Data[dataPublished], wantPorts[svc])
			}
			delete(wantPorts, svc)
		}
	}
	if len(wantPorts) != 0 {
		t.Errorf("missing published ports for: %v", wantPorts)
	}
}

func TestComposeDetector_Invalid(t *testing.T) {
	f := runDetector(t, "invalid-compose", ComposeDetector{})

	if has(f, KindExecution, "Docker Compose") {
		t.Error("an unparseable compose file must not yield an execution strategy")
	}
	foundInvalid := false
	for _, x := range f {
		if x.Kind == KindWarning && x.Data["invalid"] == "true" && x.Data[dataAmbiguous] == "true" {
			foundInvalid = true
		}
	}
	if !foundInvalid {
		t.Errorf("expected an invalid-compose warning, got: %+v", f)
	}
}

func TestComposeDetector_NoCompose(t *testing.T) {
	f := runDetector(t, "node-vite", ComposeDetector{})
	if len(f) != 0 {
		t.Errorf("expected no findings without a compose file, got: %+v", f)
	}
}
