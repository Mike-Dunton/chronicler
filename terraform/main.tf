terraform {
  backend "remote" {
    organization = "blast-hardcheese"

    workspaces {
      name = "chronicler-dev"
    }
  }
  required_providers {
    google = {
      source = "hashicorp/google"
    }
  }
}

provider "google" {
    credentials = file(".creds/chronicler-terraform.json")
    project = "chronicler-294805" 
}

resource "google_project_service" "cloud-build-service" {
  service = "cloudbuild.googleapis.com"
}
resource "google_project_service" "cloud-run-service" {
  service = "cloudrun.googleapis.com"
}
resource "google_project_service" "pubsublite-service" {
  service = " pubsublite.googleapis.com"
}
resource "google_project_service" "firestore-service" {
  service = "firestore.googleapis.com"
}