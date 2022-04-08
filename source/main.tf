terraform {
  cloud {
    organization = "ryanwholey"

    workspaces {
      name = "source"
    }
  }
}

output "foo" {
  value = random_pet.foo.id
}

output "bar" {
  value = "bar"
}

resource "random_pet" "foo" {}
