package billing

import (
	// google spread sheet
	"context"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

// PostBilling posts billing info to GSS
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

	// if there are no sheet titled "monthly_billing", create and insert header
	if !exists {
		req := CreateSheetRequest(targetSheetName)

		rbb := &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{&req},
		}

		_, err = srv.Spreadsheets.BatchUpdate(spreadsheetID, rbb).Do()
		if err != nil {
			log.Fatalf("Unable to batch update from sheet. %v", err)
		}
	}
	sheetID, err := getSheetID(ctx, targetSheetName)
	if err != nil {
		log.Fatalf("Unable to get created sheet ID. %v", err)
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

	// set header
	rcreq := HeaderRequest(sheetID, 1, 2, 1, int64(len(mnthMap)+2))

	busr := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{&rcreq},
	}

	_, err = srv.Spreadsheets.BatchUpdate(spreadsheetID, busr).Do()
	if err != nil {
		log.Fatalf("Unable to batch update from sheet. %v", err)
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

	// set border
	br := BorderRequest(sheetID, 1, int64(len(prjMap))+2, 1, 4)

	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{&br},
	}

	_, err = srv.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateRequest).Do()
	if err != nil {
		log.Fatalf("Unable to batch update from sheet. %v", err)
	}

	return nil
}
