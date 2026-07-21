# Allow-listed ingress entries per workload; CIDRs must stay inside the org ranges.
ingress_rules = {
  orders-api = {
    owner       = "orders-team"
    ports       = [443]
    allow_cidrs = ["10.20.0.0/16"]
  }
  payments-gateway = {
    owner       = "payments-team"
    ports       = [443]
    allow_cidrs = ["10.20.0.0/16", "10.30.4.0/24"]
  }
}
