package ai

const (
	default_prompt = `
	Summarize the given Kubernetes error message, enclosed by triple dashes, in --- %s --- language; --- %s ---.  
	Provide the most likely solution in a concise, step-by-step manner within 100 characters.  
	Format the output as follows:  
	Error: {Brief error explanation}  
	Solution: {Step-by-step resolution}`

	promql_explaination_prompt = `Provide a clear, concise explanation for the given PromQL query in natural language.
	Query: %s
	User Question: %s
	Primary Metric: %s
	Aggregation Type: %s
	Labels: %v
	Explain the query's purpose and behavior in one well-formed sentence. Keep it succinct and informative.`

	promql_query_builder = `You are a PromQL query builder.
	Available metrics and their labels:
	%s

	Valid PromQL examples:
	- Simple sum: sum(prometheus_http_response_size_bytes_sum)
	- Rate with time: rate(prometheus_http_requests_total[5m])
	- Filtered sum: sum(prometheus_http_requests_total{code="200"})

	Return ONLY a JSON object with these fields:
	{
		"metric": "exact_metric_name_from_list",
		"labels": {"label": "value"},
		"timeRange": "5m",      // Omit for sum aggregation
		"aggregation": ""    // sum/rate/avg/count/increase - Leave empty if no aggregation needed
	}

	Rules:
	1. Exact metric names only
	2. Only use existing label values
	3. Omit timeRange for sum operations
	4. Use appropriate aggregation:
		- sum: for totals and sizes
		- rate: for per-second metrics
		- avg: for averages
		- count: for occurrences
		- increase: for total increases`

	promql_context_extractor = `
	Extract PromQL query components from: "%s"
	Return JSON with:
	- metric: main metric name
	- labels: key-value pairs
	- timeRange: duration string
	- aggregation: aggregation function
	`
)

var PromptMap = map[string]string{
	"default":                default_prompt,
	"PromQLExplanation":      promql_explaination_prompt,
	"PromQLBuilder":          promql_query_builder,
	"PromQLContextExtractor": promql_context_extractor,
}
