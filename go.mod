module github.com/jcowgar/acme-utils

go 1.23.4

require 9fans.net/go v0.0.7

require (
	github.com/ollama/ollama v0.5.7
	gopkg.in/yaml.v2 v2.4.0
)

require github.com/sashabaranov/go-openai v1.36.1 // indirect

replace github.com/jcowgar/acme-utils => ./
