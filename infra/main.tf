terraform {
  required_version = ">= 1.5"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# Published catalog snapshots, one object per main-branch build.
resource "google_storage_bucket" "catalog_snapshots" {
  name                        = "${var.project_id}-tanuki-catalog-snapshots"
  location                    = var.region
  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }

  lifecycle_rule {
    action {
      type = "Delete"
    }
    condition {
      age = 90
    }
  }
}

resource "google_storage_bucket_iam_member" "catalog_publisher" {
  bucket = google_storage_bucket.catalog_snapshots.name
  role   = "roles/storage.objectCreator"
  member = "serviceAccount:${var.publisher_service_account}"
}
