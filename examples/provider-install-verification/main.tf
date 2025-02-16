terraform {
  required_providers {
    flespi = {
      source = "mixser/flespi"
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

  tokens_count = 5
}

resource "flespi_subaccount" "mitu" {
  name     = "mitu"
  limit_id = flespi_limit.edu.id
}

resource "flespi_webhook" "nimbus" {
  name = "nimbus 🚎"
  type = "single-webhook"

  triggers = [
    {
      topic = "flespi/state/gw/streams/+/blocked"
    },
    {
      topic = "flespi/state/gw/channels/+/blocked"
    }
  ]

  configurations = [
    {
      type   = "custom-server"
      uri    = "https://api.telegram.org/bot<YOUR_BOT_TOKEN>/sendMessage"
      method = "POST"
      body   = "{\"chat_id\": \"<YOUR_CHAT_ID>\", \"text\": \"%topics[4]% blocked: %if(payload == 1, \"Exceeded connections limit\", if(payload == 2, \"Exceeded messages limit per minute\", if(payload == 3, \"Exceeded traffic limit per minute\", if(payload == 4, \"Exceeded storage limit\", if(payload == 5, \"Exceeded items limit\", if(payload == 6, \"Item configuration is invalid\", if(payload == 7, \"Customer was moved to another region\", if(payload==null, \"unblocked\", payload))))))))% https://flespi.io/#/panel/open/%topics[3]%/%topics[4]%\", \"disable_notification\": false}"
      headers = [
        {
          name  = "Content-Type"
          value = "application/json"
        }
      ]
    },
  ]
}


resource "flespi_device" "iphone" {
  name = "iPhone 📱"
  enabled = false

  device_type_id = 1

  media_ttl = 86401
  messages_ttl = 86400

  configuration = {
    ident = 123456789123456
    settings_polling = "once"
  }
}

resource "flespi_channel" "test_channel" {
  name = "Test Channel"
  enabled = true

  protocol_name = "test"

  configuration = "{\"timeout\":120}"
}
