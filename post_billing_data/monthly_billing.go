package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

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
	Month   string
	Project string
	Total   float64
}

func main() {
	ctx := context.Background()
	projectID := os.Getenv("BILLING_PROJECT_ID")

	result, err := getBilling(ctx, projectID)
	if err != nil {
		// TODO: Handle error.
	}

	err = postBillingToGss(ctx, result)
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
			invoice.month as month,
			project.name as project,
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

func postBillingToGss(ctx context.Context, billingInfo []Billing) error {
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

	spreadsheetID := os.Getenv("GSS_ID")

	// ToDo : use valuable for sheet name
	exists, err := sheetExists(ctx, "monthly_billing")
	if err != nil {
		log.Fatalf("Unable to check Sheets existance: %v", err)
	}

	var sheetID int64 = -1
	// if there are no sheet titled "monthly_billing", create and insert header
	if !exists {

		// ToDo : define on out of the method / delete it
		gridProperties := &sheets.GridProperties{
			RowCount:    100, // for projects
			ColumnCount: 50,  // for month
		}

		tabColor := &sheets.Color{
			Red:   0.0,
			Green: 0.0,
			Blue:  0.0,
		}

		req := sheets.Request{
			AddSheet: &sheets.AddSheetRequest{
				Properties: &sheets.SheetProperties{
					Title:          "monthly_billing",
					GridProperties: gridProperties,
					TabColor:       tabColor,
				},
			},
		}

		rbb := &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{&req},
		}

		_, err = srv.Spreadsheets.BatchUpdate(spreadsheetID, rbb).Do()
		if err != nil {
			log.Fatalf("Unable to batch update from sheet. %v", err)
		}

		sheetID, err = getSheetID(ctx, "monthly_billing")
		if err != nil {
			log.Fatalf("Unable to get created sheet ID. %v", err)
		}

		// Insert Header
		rcreq := sheets.Request{
			RepeatCell: &sheets.RepeatCellRequest{
				Fields: "*",
				Range: &sheets.GridRange{
					SheetId:          sheetID, // set sheet ID
					StartRowIndex:    1,
					EndRowIndex:      2,
					StartColumnIndex: 1,
					EndColumnIndex:   4,
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

		// set header values
		hv := [][]interface{}{[]interface{}{"Month", "Project", "Total"}}
		hvr := &sheets.ValueRange{
			MajorDimension: "ROWS",
			Values:         hv,
		}

		// ToDo : use valuable
		_, err = srv.Spreadsheets.Values.Update(spreadsheetID, "monthly_billing!B2:D2", hvr).ValueInputOption("USER_ENTERED").Do()
		if err != nil {
			log.Fatalf("Unable to write value. %v", err)
		}

	}

	// ToDo : Copy value from Bigquery
	values := make([][]interface{}, len(billingInfo))

	for i, billing := range billingInfo {
		values[i] = []interface{}{billing.Month, billing.Project, billing.Total}
	}

	valueRange := &sheets.ValueRange{
		MajorDimension: "ROWS",
		Values:         values,
	}

	// ToDo : Change code adopt changing billing info columns
	_, err = srv.Spreadsheets.Values.Update(spreadsheetID, "monthly_billing!B3:d1000", valueRange).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Fatalf("Unable to write value. %v", err)
	}

	// ToDo : change SheetId
	if sheetID == -1 {
		sheetID, err = getSheetID(ctx, "monthly_billing")
		if err != nil {
			log.Fatalf("Unable to get created sheet ID. %v", err)
		}
	}
	gridRange := &sheets.GridRange{
		SheetId:          sheetID,
		StartRowIndex:    2,
		EndRowIndex:      1000,
		StartColumnIndex: 1,
		EndColumnIndex:   4,
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

func sheetExists(ctx context.Context, sheetName string) (bool, error) {
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
	spreadsheetID := os.Getenv("GSS_ID")

	// query parameter
	ranges := []string{}
	includeGridData := false

	var sps *sheets.Spreadsheet

	sps, err = srv.Spreadsheets.Get(spreadsheetID).Ranges(ranges...).IncludeGridData(includeGridData).Context(ctx).Do()

	var sheetArray []*sheets.Sheet
	sheetArray = sps.Sheets
	var sheet *sheets.Sheet
	var props *sheets.SheetProperties

	for i := 0; i < len(sheetArray); i++ {
		sheet = sheetArray[i]
		props = sheet.Properties
		if props.Title == sheetName {
			return true, nil
		}
	}
	return false, nil
}

// ToDo : Refactoring
func getSheetID(ctx context.Context, sheetName string) (int64, error) {
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
	spreadsheetID := os.Getenv("GSS_ID")

	// query parameter
	ranges := []string{}
	includeGridData := false

	var sps *sheets.Spreadsheet

	sps, err = srv.Spreadsheets.Get(spreadsheetID).Ranges(ranges...).IncludeGridData(includeGridData).Context(ctx).Do()

	var sheetArray []*sheets.Sheet
	sheetArray = sps.Sheets
	var sheet *sheets.Sheet
	var props *sheets.SheetProperties

	for i := 0; i < len(sheetArray); i++ {
		sheet = sheetArray[i]
		props = sheet.Properties
		if props.Title == sheetName {
			return props.SheetId, nil
		}
	}
	return -1, nil
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
