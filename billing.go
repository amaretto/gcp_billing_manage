package billing

// Billing has billing info from bigquery
type Billing struct {
	Month   string
	Project string
	Total   float64
}

// Billing has billing info from bigquery
type BillingService struct {
	Month   string
	Project string
	Service string
	Total   float64
}
