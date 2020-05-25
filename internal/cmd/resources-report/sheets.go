package main

import (
	"context"
	"encoding/json"

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

func updateSheet(ctx context.Context, sheetID string, resources Resources) error {
	client, err := sheets.NewService(ctx)
	if err != nil {
		return err
	}

	_, err = client.Spreadsheets.Values.Clear(sheetID, "GeneratedReport", &sheets.ClearValuesRequest{}).
		Context(ctx).Do()
	if err != nil {
		return err
	}

	_, err = client.Spreadsheets.Values.Append(sheetID, "GeneratedReport", &sheets.ValueRange{
		Values: toSheetValues(resources),
	}).ValueInputOption("RAW").Context(ctx).Do()
	if err != nil {
		return err
	}

	return nil
}
