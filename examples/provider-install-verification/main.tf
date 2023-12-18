terraform {
  required_providers {
    flespi = {
      source = "hashicorp.com/edu/flespi"
    }
  }
}

provider "flespi" {
  token = "AlY..."
}


resource "flespi_limit" "edu" {
  name           = "name of the limit"
  webhooks_count = 23
  cdn_storage    = 12
}

resource "flespi_subaccount" "mitu" {
  name     = "mitu"
  limit_id = flespi_limit.edu.id
}
