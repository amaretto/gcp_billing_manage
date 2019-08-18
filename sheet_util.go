package billing

import (
	"context"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

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

	// ToDo : Refactoring
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

// ToDo : implement
// func postBillingService
// func postBillingSku
