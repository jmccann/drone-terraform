workflow "New workflow" {
  on = "push"
  resolves = ["lint"]
}

action "lint" {
  uses = "docker://golang:1.10"
  runs = "go fmt"
}
