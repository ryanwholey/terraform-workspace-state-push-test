terraform {
  cloud {
    organization = "ryanwholey"

    workspaces {
      name = "destination"
    }
  }
}
