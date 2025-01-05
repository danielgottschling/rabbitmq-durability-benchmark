terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "6.14.1"
    }
  }
}

provider "google" {
  credentials = file("credentials.json")
  project     = "our-rampart-445315-d7"
  region      = "europe-west3"
  zone        = "europe-west3-c"
}

resource "google_compute_instance" "rabbit-mq-vm" {
  name         = "rabbit-mq"
  machine_type = "e2-standard-2"
  zone         = "europe-west3-c"
  boot_disk {
    auto_delete = true
    initialize_params {
      image = "projects/ubuntu-os-cloud/global/images/ubuntu-2404-noble-amd64-v20241219"
      size  = 10
    }
  }
  network_interface {
    subnetwork = google_compute_subnetwork.benchmark-subnetwork.id
    network_ip = "10.0.0.2"
    access_config {
    }
  }
  metadata_startup_script = file("./rabbit-mq-startup.sh")
  #tags = ["http-server"]
}

resource "google_compute_instance" "receiver-vm" {
  name         = "receiver-vm"
  machine_type = "e2-standard-2"
  zone         = "europe-west3-c"
  boot_disk {
    auto_delete = true
    initialize_params {
      image = "projects/ubuntu-os-cloud/global/images/ubuntu-2404-noble-amd64-v20241219"
      size  = 10
    }
  }
  network_interface {
    subnetwork = google_compute_subnetwork.benchmark-subnetwork.id
    network_ip = "10.0.0.3"
    access_config {
    }
  }
  #tags = ["http-server"]
}

resource "google_compute_instance" "publisher-vm" {
  name         = "publisher-vm"
  machine_type = "e2-standard-2"
  zone         = "europe-west3-c"
  boot_disk {
    auto_delete = true
    initialize_params {
      image = "projects/ubuntu-os-cloud/global/images/ubuntu-2404-noble-amd64-v20241219"
      size  = 10
    }
  }
  network_interface {
    subnetwork = google_compute_subnetwork.benchmark-subnetwork.id
    network_ip = "10.0.0.4"
    access_config {
    }
  }
  #tags = ["http-server"]
}

resource "google_compute_firewall" "benchmark-allow-http" {
  name    = "benchmark-allow-http"
  network = "default"
  allow {
    protocol = "tcp"
    ports    = ["80"]
  }
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["http-server"]
}

resource "google_compute_network" "benchmark-network" {
  name                    = "benchmark-network"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "benchmark-subnetwork" {
  name          = "benchmark-subnetwork"
  network       = google_compute_network.benchmark-network.id
  ip_cidr_range = "10.0.0.0/28"
}

resource "google_compute_firewall" "benchmark-firewall" {
  name = "benchmark-firewall"
  allow {
    protocol = "tcp"
    ports    = ["0-65535"]
  }
  allow {
    protocol = "icmp"
  }
  network       = google_compute_network.benchmark-network.id
  source_ranges = ["0.0.0.0/0"]
}
