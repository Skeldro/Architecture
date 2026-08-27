# Decision 4: everything needed to reach a running system is declared here.
# There is no cluster to hand-build — that is what closed Phase 1's
# reproducibility hole, and it is why this file is roughly fifty lines instead
# of a directory of manifests.
#
# NOT APPLIED. This configuration has never been run: doing so creates billable
# cloud resources and requires credentials this project does not have. It is the
# documented deployment path, not a verified one.

terraform {
  required_version = ">= 1.5"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0"
    }
  }
}

variable "project_id" { type = string }
variable "region" {
  type    = string
  default = "europe-west1"
}
variable "image" { type = string } # e.g. europe-west1-docker.pkg.dev/PROJECT/collabdocs/app:TAG
variable "db_password" {
  type      = string
  sensitive = true
  # At one developer this is a tfvars entry. Decision 4 records that at ~15
  # developers it moves into a managed secret store instead.
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# ---------- state: its own lifecycle, its own failure domain (Decision 3) ----------

resource "google_sql_database_instance" "pg" {
  name             = "collabdocs-pg"
  database_version = "POSTGRES_17"
  region           = var.region

  settings {
    tier = "db-custom-2-7680" # 2 vCPU: sized for MVP, resized for 10x (NFR4)

    # FR2: state outlives any runtime. Automated backups plus point-in-time
    # recovery are the only path in this architecture whose recovery time
    # exceeds NFR3's 60 seconds, which Decision 3 states explicitly.
    backup_configuration {
      enabled                        = true
      point_in_time_recovery_enabled = true
      start_time                     = "03:00"
    }

    database_flags {
      # Decision 2's ceiling: 8 instances x 10 connections = 80, leaving room
      # for migrations, monitoring and administrative sessions.
      name  = "max_connections"
      value = "100"
    }
  }

  deletion_protection = true
}

resource "google_sql_database" "app" {
  name     = "collabdocs"
  instance = google_sql_database_instance.pg.name
}

resource "google_sql_user" "app" {
  name     = "collabdocs"
  instance = google_sql_database_instance.pg.name
  password = var.db_password
}

# ---------- compute: stateless, two warm instances (Decision 1) ----------

resource "google_cloud_run_v2_service" "app" {
  name     = "collabdocs"
  location = var.region

  template {
    scaling {
      # Derived in Decision 1: one instance carries MVP load at ~30%
      # utilisation, so the second exists to survive losing one and to allow
      # rolling deploys with no downtime. The ceiling is set by database
      # connections, not by compute.
      min_instance_count = 2
      max_instance_count = 8
    }

    containers {
      image = var.image

      env {
        name = "DATABASE_URL"
        # Unix socket through the Cloud SQL connector, so no public IP exists.
        value = "postgres://collabdocs:${var.db_password}@/collabdocs?host=/cloudsql/${google_sql_database_instance.pg.connection_name}"
      }
      env {
        name  = "DB_MAX_CONNS"
        value = "10"
      }

      resources {
        limits = { cpu = "1", memory = "512Mi" }
        # Minimum instances stay warm: no request pays a cold start, which is
        # what keeps NFR1's latency budget honest.
        cpu_idle = false
      }

      # FR4 in the platform's hands: liveness restarts a wedged instance,
      # readiness removes a database-blind one from rotation without killing it.
      startup_probe {
        http_get { path = "/healthz" }
        initial_delay_seconds = 2
        failure_threshold     = 10
        period_seconds        = 2
      }
      liveness_probe {
        http_get { path = "/livez" }
        period_seconds = 10
      }

      volume_mounts {
        name       = "cloudsql"
        mount_path = "/cloudsql"
      }
    }

    volumes {
      name = "cloudsql"
      cloud_sql_instance { instances = [google_sql_database_instance.pg.connection_name] }
    }
  }

  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
    percent = 100
  }
}

# Constraint 2 forbids authentication this phase, so the service is public.
# Phase 2 replaces this with real identity.
resource "google_cloud_run_v2_service_iam_member" "public" {
  name     = google_cloud_run_v2_service.app.name
  location = google_cloud_run_v2_service.app.location
  role     = "roles/run.invoker"
  member   = "allUsers"
}

output "url" { value = google_cloud_run_v2_service.app.uri }
