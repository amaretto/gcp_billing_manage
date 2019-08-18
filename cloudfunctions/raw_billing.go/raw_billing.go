package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	// bigquery
	"cloud.google.com/go/bigquery"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"

	// google spread sheet
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

// Billing has billing info from bigquery
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

// SKU has category id and name of billing
type SKU struct {
	ID          string
	Description string
}

func main() {
	ctx := context.Background()
	projectID := os.Getenv("BILLING_PROJECT_ID")

	result, err := getBilling(ctx, projectID)
	if err != nil {
		// TODO: Handle error.
	}

	err = postBillingToGss(result)

	if err != nil {
		// TODO: Handle error.
	}
}

// getBilling get billing info from bigquery
func getBilling(ctx context.Context, projectID string) ([]Billing, error) {
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
		FROM ` + os.Getenv("BILLING_BQ_TBL") + //"`billing.gcp_billing_export_v1_000D1F_7C9B2E_312DA4`" + `
		`
		LIMIT 10000
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

func postBillingToGss(billingInfo []Billing) error {
	// ToDo : use environment value
	credentialPath := os.Getenv("CREDENTIAL_PATH")
	b, err := ioutil.ReadFile(credentialPath)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	// Prints the names and majors of students in a sample spreadsheet:
	// ToDo : use environment value
	spreadsheetID := os.Getenv("GSS_ID")

	// ToDo : Create Sheet
	//	gridProperties := &sheets.GridProperties{
	//		RowCount:    20,
	//		ColumnCount: 12,
	//	}
	//
	//	tabColor := &sheets.Color{
	//		Red:   1.0,
	//		Green: 0.3,
	//		Blue:  0.4,
	//	}
	//
	//	req := sheets.Request{
	//		AddSheet: &sheets.AddSheetRequest{
	//			Properties: &sheets.SheetProperties{
	//				Title:          "fuga",
	//				GridProperties: gridProperties,
	//				TabColor:       tabColor,
	//			},
	//		},
	//	}
	//
	//	rbb := &sheets.BatchUpdateSpreadsheetRequest{
	//		Requests: []*sheets.Request{&req},
	//	}
	//
	//	_, err = srv.Spreadsheets.BatchUpdate(spreadsheetID, rbb).Do()
	//	if err != nil {
	//		log.Fatalf("Unable to batch update from sheet. %v", err)
	//	}

	// ToDo : Insert Header
	rcreq := sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Fields: "*",
			Range: &sheets.GridRange{
				SheetId:          0,
				StartRowIndex:    0,
				EndRowIndex:      1,
				StartColumnIndex: 1,
				EndColumnIndex:   17,
			},
			Cell: &sheets.CellData{
				UserEnteredFormat: &sheets.CellFormat{
					BackgroundColor: &sheets.Color{
						Red:   0.5,
						Green: 0.5,
						Blue:  1.0,
					},
					TextFormat: &sheets.TextFormat{
						ForegroundColor: &sheets.Color{
							Red:   1.0,
							Green: 1.0,
							Blue:  1.0,
						},
						Bold:     true,
						FontSize: 12,
					},
				},
			},
		},
	}

	busr := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{&rcreq},
	}

	_, err = srv.Spreadsheets.BatchUpdate(spreadsheetID, busr).Do()
	if err != nil {
		log.Fatalf("Unable to batch update from sheet. %v", err)
	}

	hv := [][]interface{}{[]interface{}{"Service", "Sku.ID", "Sku.Description",
		"Ustart", "Uend", "Prjid", "Prjname", "Country",
		"Region", "Exptime", "Cost", "Currency",
		"Uamount", "Uunit", "Uapriceunit", "Upriceunit"}}
	hvr := &sheets.ValueRange{
		MajorDimension: "ROWS",
		Values:         hv,
	}

	_, err = srv.Spreadsheets.Values.Update(spreadsheetID, "シート1!B1:Q1", hvr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Fatalf("Unable to write value. %v", err)
	}

	// ToDo : Copy value from Bigquery
	values := make([][]interface{}, len(billingInfo))

	for i, billing := range billingInfo {
		values[i] = []interface{}{billing.Service, billing.Sku.ID, billing.Sku.Description,
			billing.Ustart, billing.Uend, billing.Prjid, billing.Prjname, billing.Country,
			billing.Region, billing.Exptime, billing.Cost, billing.Currency,
			billing.Uamount, billing.Uunit, billing.Uapriceunit, billing.Upriceunit}
	}

	valueRange := &sheets.ValueRange{
		MajorDimension: "ROWS",
		Values:         values,
	}

	// ToDo : Change code adopt changing billing info columns
	_, err = srv.Spreadsheets.Values.Update(spreadsheetID, "シート1!B2:Q15000", valueRange).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Fatalf("Unable to write value. %v", err)
	}

	// ToDo : Set Boarder
	// ToDo : change SheetId
	gridRange := &sheets.GridRange{
		SheetId:          0,
		StartRowIndex:    3,
		EndRowIndex:      100,
		StartColumnIndex: 1,
		EndColumnIndex:   17,
	}

	borderColor := &sheets.Color{
		Red:   0,
		Green: 0,
		Blue:  0,
		Alpha: 0,
	}

	border := &sheets.Border{
		Style: "SOLID",
		Width: 1,
		Color: borderColor,
	}

	updateBordersRequest := &sheets.UpdateBordersRequest{
		Range:           gridRange,
		Top:             border,
		Bottom:          border,
		Left:            border,
		Right:           border,
		InnerHorizontal: border,
		InnerVertical:   border,
	}

	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			&sheets.Request{
				UpdateBorders: updateBordersRequest,
			},
		},
	}

	_, err = srv.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateRequest).Do()
	if err != nil {
		log.Fatalf("Unable to batch update from sheet. %v", err)
	}

	return nil

}

// --------------auth modules-------------------
// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
