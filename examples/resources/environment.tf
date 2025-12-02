resource "devgraph_environment" "example" {
  name                   = "my-environment"
  stripe_subscription_id = "sub_1234567890"
  instance_url           = "https://my-env.devgraph.ai"

  invited_users = [
    "user1@example.com",
    "user2@example.com"
  ]
}
