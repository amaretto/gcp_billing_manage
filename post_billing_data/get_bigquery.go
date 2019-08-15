package billing

import (
	// bigquery
	"log"
	"os"

	"cloud.google.com/go/bigquery"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

// GetBilling get billing info from bigquery
func GetBilling(ctx context.Context, projectID string) ([]Billing, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	q := client.Query(`
		SELECT
			invoice.month as month,
			IFNULL(project.name, "TAX") as project,
			(SUM(CAST(cost * 1000000 AS int64))
			+ SUM(IFNULL((SELECT SUM(CAST(c.amount * 1000000 as int64))
		FROM UNNEST(credits) c), 0))) / 1000000
		AS total
		FROM ` + os.Getenv("BILLING_BQ_TBL") + `
		GROUP BY month, project
		ORDER BY month, project  ASC
	`)

	it, err := q.Read(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	var billingInfo []Billing

	for {
		var b Billing
		err := it.Next(&b)
		if err == iterator.Done {
			break
		}
		if err != nil {
			// TODO: Handle error.
		}
		billingInfo = append(billingInfo, b)
	}
	return billingInfo, nil
}

// ToDo : implement
// func GetBillingService
// func GetBillingSku
