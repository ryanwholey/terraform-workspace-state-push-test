# terraform-copy-state

Pushes state between Terraform Cloud backends via API, so circumvents the need for the repo to exist locally and running local operations (`terraform init`, etc.)

Currently requires that the destination workspace has no current state.
