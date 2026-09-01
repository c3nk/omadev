package detect

// PostgresDetector finds PostgreSQL from client-library evidence in dependency files
// (the Compose-image case is handled by the Compose detector, so the file is not
// parsed twice). It never reads database credentials.
type PostgresDetector struct{}

func (PostgresDetector) Name() string { return "postgres" }

// pgIndicators are dependency substrings that imply a PostgreSQL client.
var pgIndicators = []string{"psycopg", "asyncpg", "pg8000", "postgres", "pg-promise"}

func (d PostgresDetector) Detect(ctx Context) ([]Finding, error) {
	seen := map[string]bool{}
	var out []Finding

	depFiles := []string{"requirements.txt", "pyproject.toml", "Pipfile", "package.json"}
	for _, name := range depFiles {
		for _, p := range findFilesNamed(ctx.FS, name) {
			text, err := readText(ctx.FS, p)
			if err != nil {
				continue
			}
			component := dirOf(p)
			if seen[component] {
				continue
			}
			for _, needle := range pgIndicators {
				if containsFold(text, needle) {
					out = append(out, techFinding(d.Name(), "PostgreSQL", component))
					seen[component] = true
					break
				}
			}
		}
	}
	return out, nil
}
