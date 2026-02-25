terraform {
  required_version = ">= 0.15.5"

  required_providers {
    wallarm = {
      source  = "wallarm/wallarm"
      version = "1.9.0"
    }
  }
}
