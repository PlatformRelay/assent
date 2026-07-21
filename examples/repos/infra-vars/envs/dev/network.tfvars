ingress_rules = {
  orders-api = {
    owner       = "orders-team"
    ports       = [443, 8080]
    allow_cidrs = ["10.40.0.0/16"]
  }
  checkout-session-svc = {
    owner       = "checkout-team"
    ports       = [8080]
    allow_cidrs = ["10.40.0.0/16"]
  }
}
