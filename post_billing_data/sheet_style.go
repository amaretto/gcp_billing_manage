package billing

import sheets "google.golang.org/api/sheets/v4"

// HEDDER_REQUEST() return request object that change reffered cell format
func HEADER_REQUEST(shtId string, strtRwIdx, endRowIdx, strtClmnIdx, endClmnIdx int) sheets.Request {
	return sheets.Request{
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
}
