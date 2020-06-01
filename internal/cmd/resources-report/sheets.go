package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	sheets "google.golang.org/api/sheets/v4"
)

func toSheetValues(resources Resources) [][]interface{} {
	values := make([][]interface{}, len(resources)+1)
	values[0] = []interface{}{"Platform", "Type", "ID", "Location", "Owner", "Created", "Meta"}
	for i, resource := range resources {
		var meta string
		metaBytes, err := json.Marshal(resource.Meta)
		if err == nil {
			meta = string(metaBytes)
		}
		values[i+1] = []interface{}{
			resource.Platform,
			resource.Type,
			resource.Identifier,
			resource.Location,
			resource.Owner,
			resource.Created.UTC(),
			meta,
		}
	}
	return values
}

const defaultPage = "GeneratedReport"

func updateSheet(ctx context.Context, sheetID string, resources Resources, highlighted int) error {
	client, err := sheets.NewService(ctx)
	if err != nil {
		return fmt.Errorf("failed to init client: %w", err)
	}

	_, err = client.Spreadsheets.Values.Clear(sheetID, defaultPage, &sheets.ClearValuesRequest{}).
		Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to reset report: %w", err)
	}

	_, err = client.Spreadsheets.Values.Append(sheetID, defaultPage, &sheets.ValueRange{
		Values: toSheetValues(resources),
	}).ValueInputOption("RAW").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to update report: %w", err)
	}

	log.Print(highlighted)
	_, err = client.Spreadsheets.BatchUpdate(sheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			// fix header: https://developers.google.com/sheets/api/samples/formatting#format_a_header_row
			{UpdateSheetProperties: &sheets.UpdateSheetPropertiesRequest{
				Properties: &sheets.SheetProperties{
					SheetId:        0,
					GridProperties: &sheets.GridProperties{FrozenRowCount: 1},
				},
				Fields: "gridProperties.frozenRowCount",
			}},

			// highlight most recent entries
			{RepeatCell: &sheets.RepeatCellRequest{
				Range: &sheets.GridRange{
					SheetId:       0,
					StartRowIndex: 1,
					EndRowIndex:   int64(highlighted),
				},
				Cell: &sheets.CellData{
					UserEnteredFormat: &sheets.CellFormat{
						TextFormat: &sheets.TextFormat{Bold: true},
					},
				},
				Fields: "userEnteredFormat.textFormat.bold",
			}},

			// set extra entries to not-bold
			{RepeatCell: &sheets.RepeatCellRequest{
				Range: &sheets.GridRange{
					SheetId:       0,
					StartRowIndex: int64(highlighted),
				},
				Cell: &sheets.CellData{
					UserEnteredFormat: &sheets.CellFormat{
						TextFormat: &sheets.TextFormat{Bold: false},
					},
				},
				Fields: "userEnteredFormat.textFormat.bold",
			}},
		},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to format report: %w", err)
	}

	return nil
}
