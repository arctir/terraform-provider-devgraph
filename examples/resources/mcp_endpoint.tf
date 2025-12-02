resource "devgraph_mcp_endpoint" "example" {
  name        = "my-mcp-server"
  url         = "https://mcp.example.com/v1"
  description = "Example MCP endpoint"

  headers = {
    "X-Custom-Header" = "value"
  }

  devgraph_auth      = false
  supports_resources = true
  active             = true

  allowed_tools = ["tool1", "tool2"]
}
