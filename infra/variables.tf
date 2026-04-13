variable "project_id" {
  description = "GCP project that hosts the catalog snapshot bucket"
  type        = string
}

variable "region" {
  description = "Region for the snapshot bucket"
  type        = string
  default     = "us-central1"
}

variable "publisher_service_account" {
  description = "Service account email that CI uses to publish catalog snapshots"
  type        = string
}
