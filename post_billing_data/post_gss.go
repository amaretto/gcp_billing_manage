package billing

import (
	// google spread sheet
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

func PostBilling(ctx context.Context, billingInfo []Billing) error {
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
	targetSheetName := "monthly_billing"

	// ToDo : use valuable for sheet name
	exists, err := sheetExists(ctx, targetSheetName)
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
					Title:          targetSheetName,
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

		sheetID, err = getSheetID(ctx, targetSheetName)
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
					EndColumnIndex:   10,
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
	}

	//values := make([][]interface{}, len(billingInfo))
	// create project map,index and month map, index
	prjMap := make(map[string]int)
	prjIdx := 1
	mnthMap := make(map[string]int)
	mnthIdx := 1

	for _, billing := range billingInfo {
		//values[i] = []interface{}{billing.Month, billing.Project, billing.Total}
		if prjMap[billing.Project] == 0 {
			prjMap[billing.Project] = prjIdx
			prjIdx++
		}
		if mnthMap[billing.Month] == 0 {
			mnthMap[billing.Month] = mnthIdx
			mnthIdx++
		}

	}

	// make 2 dimention slice
	nums := make([][]interface{}, len(prjMap)+1)
	for i := 0; i < len(prjMap)+1; i++ {
		nums[i] = make([]interface{}, len(mnthMap)+1)
	}
	// x header
	nums[0][0] = "Project"
	for prj, idx := range prjMap {
		nums[idx][0] = prj
	}
	// y header
	for mnth, idx := range mnthMap {
		nums[0][idx] = mnth
	}

	for _, billing := range billingInfo {
		nums[prjMap[billing.Project]][mnthMap[billing.Month]] = billing.Total
	}

	valueRange := &sheets.ValueRange{
		MajorDimension: "ROWS",
		Values:         nums,
	}

	// ToDo : Change code adopt changing billing info columns
	_, err = srv.Spreadsheets.Values.Update(spreadsheetID, targetSheetName+"!B2:J10", valueRange).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Fatalf("Unable to write value. %v", err)
	}

	if sheetID == -1 {
		sheetID, err = getSheetID(ctx, targetSheetName)
		if err != nil {
			log.Fatalf("Unable to get created sheet ID. %v", err)
		}
	}

	// ToDo : gathering parameters to one struct
	gridRange := &sheets.GridRange{
		SheetId:          sheetID,
		StartRowIndex:    2,
		EndRowIndex:      10,
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

//---------------------------------------------------------
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

// ToDo : implement
// func postBillingService
// func postBillingSku