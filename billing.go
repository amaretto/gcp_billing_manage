package billing

// Billing has billing info from bigquery
type Billing struct {
	Month   string
	Project string
	Total   float64
}

// AmountWithService has billing info with service name from bigquery
type AmountWithService struct {
	Month   string
	Project string
	Service string
	Total   float64
}

// ToDo : implement AmountWithSKU
type AmountWithSKU struct {
	Month   string
	Project string
	Service string
	SKU     string
	Total   float64
}
