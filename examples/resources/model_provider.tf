resource "devgraph_model_provider" "anthropic" {
  type    = "anthropic"
  name    = "my-anthropic-provider"
  api_key = var.anthropic_api_key
  default = true
}

resource "devgraph_model_provider" "openai" {
  type    = "openai"
  name    = "my-openai-provider"
  api_key = var.openai_api_key
  default = false
}
