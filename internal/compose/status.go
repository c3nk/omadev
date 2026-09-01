package compose

import (
	"encoding/json"
	"strings"
)

// ServiceState is a Compose-substantiated service state (D6). It is never an
// application-readiness claim.
type ServiceState string

const (
	StateRunning   ServiceState = "running"
	StateHealthy   ServiceState = "healthy"
	StateUnhealthy ServiceState = "unhealthy"
	StateStopped   ServiceState = "stopped"
	StateUnknown   ServiceState = "unknown"
)

// ServiceStatus is one service's state and published ports.
type ServiceStatus struct {
	Name  string
	State ServiceState
	Ports []int // published ports
}

// psEntry mirrors the fields of `docker compose ps --format json` we use.
type psEntry struct {
	Name       string `json:"Name"`
	Service    string `json:"Service"`
	State      string `json:"State"`
	Health     string `json:"Health"`
	Publishers []struct {
		PublishedPort int `json:"PublishedPort"`
	} `json:"Publishers"`
}

// ParsePS parses `docker compose ps --format json` output. Compose emits either a
// JSON array or one JSON object per line depending on version; both are handled.
func ParsePS(output string) []ServiceStatus {
	entries := parsePSEntries(output)
	out := make([]ServiceStatus, 0, len(entries))
	for _, e := range entries {
		name := e.Service
		if name == "" {
			name = e.Name
		}
		var ports []int
		for _, p := range e.Publishers {
			if p.PublishedPort > 0 {
				ports = append(ports, p.PublishedPort)
			}
		}
		out = append(out, ServiceStatus{Name: name, State: mapState(e.State, e.Health), Ports: dedupePorts(ports)})
	}
	return out
}

func parsePSEntries(output string) []psEntry {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return nil
	}
	if strings.HasPrefix(trimmed, "[") {
		var arr []psEntry
		if err := json.Unmarshal([]byte(trimmed), &arr); err == nil {
			return arr
		}
		return nil
	}
	var out []psEntry
	for _, line := range strings.Split(trimmed, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var e psEntry
		if err := json.Unmarshal([]byte(line), &e); err == nil {
			out = append(out, e)
		}
	}
	return out
}

func mapState(state, health string) ServiceState {
	switch strings.ToLower(health) {
	case "healthy":
		return StateHealthy
	case "unhealthy":
		return StateUnhealthy
	}
	switch strings.ToLower(state) {
	case "running":
		return StateRunning
	case "exited", "stopped", "dead", "created":
		return StateStopped
	case "":
		return StateUnknown
	default:
		return StateUnknown
	}
}

func dedupePorts(ports []int) []int {
	seen := map[int]bool{}
	var out []int
	for _, p := range ports {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	return out
}
