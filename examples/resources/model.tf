resource "devgraph_model" "claude_opus" {
  name        = "claude-3-opus-20240229"
  description = "Claude 3 Opus model"
  provider_id = devgraph_model_provider.anthropic.id
  default     = true
}

resource "devgraph_model" "gpt4" {
  name        = "gpt-4-turbo-preview"
  description = "GPT-4 Turbo model"
  provider_id = devgraph_model_provider.openai.id
  default     = false
}
