package billing

import (
	// bigquery
	"log"
	"os"

	"cloud.google.com/go/bigquery"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

// GetAllBilling get billing info(project,service,amount for each month) from bigquery
func GetAllBilling(ctx context.Context, projectID string) ([]Billing, error) {
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

// GetAllBillingService get billing info(project,amount for each month) from bigquery
func GetAllBillingService(ctx context.Context, projectID string) ([]AmountWithService, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	q := client.Query(`
		SELECT
			invoice.month AS month,
			IFNULL(project.name, "TAX") AS project,
			service.description AS service,
			(SUM(CAST(cost * 1000000 AS int64))
			+ SUM(IFNULL((SELECT SUM(CAST(c.amount * 1000000 as int64))
		FROM UNNEST(credits) c), 0))) / 1000000
		AS total
		FROM ` + os.Getenv("BILLING_BQ_TBL") + `
		GROUP BY month, project, service
		ORDER BY month, project, service  ASC
	`)

	it, err := q.Read(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	var billingInfo []AmountWithService

	for {
		var b AmountWithService
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

// GetAllBillingSKU get billing info(project,service,sku,amount for each month) from bigquery
func GetAllBillingSKU(ctx context.Context, projectID string) ([]AmountWithSKU, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	q := client.Query(`
		SELECT
			invoice.month,
			IFNULL(project.name, "TAX") AS project,
			service.description as service,
			sku.description as sku,
			(SUM(CAST(cost * 1000000 AS int64))
			+ SUM(IFNULL((SELECT SUM(CAST(c.amount * 1000000 as int64))
		FROM UNNEST(credits) c), 0))) / 1000000
		AS total
		FROM ` + os.Getenv("BILLING_BQ_TBL") + `
		GROUP BY month,project,service,sku
		ORDER BY month,project,service,sku ASC
		;
	`)

	it, err := q.Read(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	var billingInfo []AmountWithSKU

	for {
		var b AmountWithSKU
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
// func GetBilling
// func GetBillingService
// func GetBillingSku
