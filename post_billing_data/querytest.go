// Sample bigquery-quickstart creates a Google BigQuery dataset.
package main

import (
	"fmt"
	"log"
	"time"

	// Imports the Google Cloud BigQuery client package.
	"cloud.google.com/go/bigquery"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

func main() {
	ctx := context.Background()

	// Sets your Google Cloud Platform project ID.
	// ToDo:fix before commit!
	projectID := "gobot-164815"

	// Creates a client.
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	q := client.Query(`
	    SELECT
			service.description as service,
			sku,
			usage_start_time as ustart,
			usage_end_time as uend,
			project.id as prjid,
			project.name as prjname,
			location.country as country,
			location.region as region,
			export_time as exptime,
			cost,
			currency_conversion_rate as currency,
			usage.amount as uamount,
			usage.unit as uunit,
			usage.amount_in_pricing_units as uapriceunit,
			usage.pricing_unit as upriceunit
		FROM ` + "`billing.gcp_billing_export_v1_000D1F_7C9B2E_312DA4`" + `
	    LIMIT 3
		`)

	it, err := q.Read(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	//Billing is billing info from Bigquery
	type SKU struct {
		ID          string
		Description string
	}
	type Billing struct {
		Service     string
		Sku         SKU
		Ustart      time.Time
		Uend        time.Time
		Prjid       string
		Prjname     string
		Country     string
		Region      string
		Exptime     time.Time
		Cost        float64
		Currency    float64
		Uamount     float64
		Uunit       string
		Uapriceunit float64
		Upriceunit  string
	}

	for {
		//var values []bigquery.Value
		var b Billing
		//err := it.Next(&values)
		err := it.Next(&b)
		if err == iterator.Done {
			break
		}
		if err != nil {
			// TODO: Handle error.
			fmt.Println(err)
		}
		fmt.Println(
			"-------------------------\n",
			"Service", b.Service, "\n",
			"Sku", b.Sku, "\n",
			"Ustart", b.Ustart, "\n",
			"Uend", b.Uend, "\n",
			"Prjid", b.Prjid, "\n",
			"Prjname", b.Prjname, "\n",
			"Country", b.Country, "\n",
			"Region", b.Region, "\n",
			"Exptime", b.Exptime, "\n",
			"Cost", b.Cost, "\n",
			"Currency", b.Currency, "\n",
			"Uamount", b.Uamount, "\n",
			"Uunit", b.Uunit, "\n",
			"Uapriceunit", b.Uapriceunit, "\n",
			"Upriceunit", b.Upriceunit, "\n",
		)

	}

}
