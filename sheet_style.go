package billing

import sheets "google.golang.org/api/sheets/v4"

// HeaderRequest return request object that change reffered cell format
func HeaderRequest(sheetID, strtRwIdx, endRwIdx, strtClmnIdx, endClmnIdx int64) sheets.Request {
	return sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Fields: "*",
			Range: &sheets.GridRange{
				SheetId:          sheetID,
				StartRowIndex:    strtRwIdx,
				EndRowIndex:      endRwIdx,
				StartColumnIndex: strtClmnIdx,
				EndColumnIndex:   endClmnIdx,
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

// BorderRequest return request object that set borders for reffered cell
func BorderRequest(sheetID, strtRwIdx, endRwIdx, strtClmnIdx, endClmnIdx int64) sheets.Request {
	border := &sheets.Border{
		Style: "SOLID",
		Width: 1,
		Color: &sheets.Color{
			Red:   0,
			Green: 0,
			Blue:  0,
			Alpha: 0,
		},
	}

	return sheets.Request{
		UpdateBorders: &sheets.UpdateBordersRequest{
			Range: &sheets.GridRange{
				SheetId:          sheetID,
				StartRowIndex:    strtRwIdx,
				EndRowIndex:      endRwIdx,
				StartColumnIndex: strtClmnIdx,
				EndColumnIndex:   endClmnIdx,
			},
			Top:             border,
			Bottom:          border,
			Left:            border,
			Right:           border,
			InnerHorizontal: border,
			InnerVertical:   border,
		},
	}
}

// CreateSheetRequest return request that create new sheet
func CreateSheetRequest(targetSheetName string) sheets.Request {
	return sheets.Request{
		AddSheet: &sheets.AddSheetRequest{
			Properties: &sheets.SheetProperties{
				Title: targetSheetName,
				GridProperties: &sheets.GridProperties{
					RowCount:    100, // for projects
					ColumnCount: 50,  // for month
				},
				TabColor: &sheets.Color{
					Red:   0.0,
					Green: 0.0,
					Blue:  0.0,
				},
			},
		},
	}
}
