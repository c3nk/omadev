package project

import "testing"

func TestConfidenceString(t *testing.T) {
	cases := map[Confidence]string{
		ConfidenceLow:    "LOW",
		ConfidenceMedium: "MEDIUM",
		ConfidenceHigh:   "HIGH",
		Confidence(99):   "UNKNOWN",
	}
	for c, want := range cases {
		if got := c.String(); got != want {
			t.Errorf("Confidence(%d).String() = %q, want %q", c, got, want)
		}
	}
}

func TestConfidenceOrdering(t *testing.T) {
	if !(ConfidenceLow < ConfidenceMedium && ConfidenceMedium < ConfidenceHigh) {
		t.Error("confidence levels must order Low < Medium < High for gating")
	}
}

func TestExecutionStrategyString(t *testing.T) {
	cases := map[ExecutionStrategy]string{
		ExecutionNone:         "None",
		ExecutionCompose:      "Docker Compose",
		ExecutionStrategy(99): "Unknown",
	}
	for e, want := range cases {
		if got := e.String(); got != want {
			t.Errorf("ExecutionStrategy(%d).String() = %q, want %q", e, got, want)
		}
	}
}
