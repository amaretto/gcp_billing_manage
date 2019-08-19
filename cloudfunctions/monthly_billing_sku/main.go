package main

import (
	"os"

	"golang.org/x/net/context"

	// billing
	"github.com/amaretto/gcp_billing_manage"
)

func main() {
	ctx := context.Background()
	projectID := os.Getenv("BILLING_PROJECT_ID")

	result, err := billing.GetAllBillingSKU(ctx, projectID)
	if err != nil {
		// TODO: Handle error.
	}

	err = billing.PostBillingWithSKU(ctx, result)
	if err != nil {
		// TODO: Handle error.
	}
}
