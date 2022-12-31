variable "personal_phone_number" {
  description = "Phone number to send SMS to"
  type        = string
}

variable "sms_endpoint" {
  default     = ""
  description = "Custom endpoint to send a request to when creating SMS. Used for testing"
  type        = string
}

variable "twilio_account_sid" {
  description = "SID for Twilio account"
  type        = string
}

variable "twilio_auth_token" {
  description = "Auth token for Twilio"
  type        = string
}

variable "twilio_phone_number" {
  description = "Phone number used by Twilio"
  type        = string
}
