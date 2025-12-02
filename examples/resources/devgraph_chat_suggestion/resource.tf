resource "devgraph_chat_suggestion" "analyze_codebase" {
  title  = "Analyze codebase structure"
  label  = "Code Analysis"
  action = "Please analyze the main repository and provide insights on code quality, architecture, and dependencies."
  active = true
}

resource "devgraph_chat_suggestion" "list_services" {
  title  = "List all microservices"
  label  = "Services"
  action = "Show me all microservices in the production environment and their current deployment status."
  active = true
}

resource "devgraph_chat_suggestion" "security_audit" {
  title  = "Review security vulnerabilities"
  label  = "Security"
  action = "Show me any security vulnerabilities or outdated dependencies that need attention."
  active = true
}
