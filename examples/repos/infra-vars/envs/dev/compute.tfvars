# Dev twin — same shape as prod, smaller sizes, looser bands.
workloads = {
  orders-api = {
    owner        = "orders-team"
    instance_set = "standard-2"
    min_replicas = 1
    max_replicas = 4
    memory_mb    = 1024
  }
  checkout-session-svc = {
    owner        = "checkout-team"
    instance_set = "standard-2"
    min_replicas = 1
    max_replicas = 2
    memory_mb    = 512
  }
}
