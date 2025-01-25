package query

type Builder struct {
	templates map[string]string
}

func NewBuilder() *Builder {
	return &Builder{
		templates: map[string]string{
			"cpu_usage": `100 - (avg by(instance)(rate(node_cpu_seconds_total{mode="idle"}[$DURATION])) * 100)`,
		},
	}
}

// @params: intent Intent, params map[string]string
func (b *Builder) Build() string {
	// Template replacement logic
	return ""
}
