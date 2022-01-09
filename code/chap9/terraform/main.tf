terraform {
  required_providers {
    docker = {
      source = "kreuzwerker/docker"
      version = "2.11.0"
    }
  }
}

provider "docker" {
}

resource "docker_container" "nginx" {
  image = "nginx:latest"
  name  = "nginx_app"
  restart = "always"
  ports {
    internal = 80
    external = 8080
  }
}
