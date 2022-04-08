terraform {
  cloud {
    organization = "ryanwholey"

    workspaces {
      name = "source"
    }
  }
}

output "foo" {
  value = random_pet.foo.value
}

output "bar" {
  value = "bar"
}

resource "random_pet" "foo" {}
