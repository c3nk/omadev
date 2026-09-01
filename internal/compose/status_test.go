package compose

import "testing"

func TestParsePS_JSONL(t *testing.T) {
	out := `{"Name":"app-frontend-1","Service":"frontend","State":"running","Health":"","Publishers":[{"PublishedPort":5173}]}
{"Name":"app-postgres-1","Service":"postgres","State":"running","Health":"healthy","Publishers":[{"PublishedPort":5432}]}
{"Name":"app-worker-1","Service":"worker","State":"exited","Health":""}`

	got := ParsePS(out)
	if len(got) != 3 {
		t.Fatalf("parsed %d services, want 3", len(got))
	}

	byName := map[string]ServiceStatus{}
	for _, s := range got {
		byName[s.Name] = s
	}
	if byName["frontend"].State != StateRunning || len(byName["frontend"].Ports) != 1 || byName["frontend"].Ports[0] != 5173 {
		t.Errorf("frontend = %+v", byName["frontend"])
	}
	if byName["postgres"].State != StateHealthy {
		t.Errorf("postgres state = %q, want healthy", byName["postgres"].State)
	}
	if byName["worker"].State != StateStopped {
		t.Errorf("worker state = %q, want stopped", byName["worker"].State)
	}
}

func TestParsePS_Array(t *testing.T) {
	out := `[{"Service":"web","State":"running","Health":"unhealthy","Publishers":[{"PublishedPort":8080}]}]`
	got := ParsePS(out)
	if len(got) != 1 || got[0].State != StateUnhealthy {
		t.Fatalf("got %+v, want one unhealthy service", got)
	}
}

func TestParsePS_Empty(t *testing.T) {
	if got := ParsePS("  \n"); len(got) != 0 {
		t.Errorf("empty output should yield no services, got %+v", got)
	}
}
