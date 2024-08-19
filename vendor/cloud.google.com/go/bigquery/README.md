## BigQuery [![Go Reference](https://pkg.go.dev/badge/cloud.google.com/go/bigquery.svg)](https://pkg.go.dev/cloud.google.com/go/bigquery)

- [About BigQuery](https://cloud.google.com/bigquery/)
- [API documentation](https://cloud.google.com/bigquery/docs)
- [Go client documentation](https://pkg.go.dev/cloud.google.com/go/bigquery)
- [Complete sample programs](https://github.com/GoogleCloudPlatform/golang-samples/tree/main/bigquery)

### Example Usage

First create a `bigquery.Client` to use throughout your application:
[snip]:# (bq-1)
```go
c, err := bigquery.NewClient(ctx, "my-project-ID")
if err != nil {
	// TODO: Handle error.
}
```

Then use that client to interact with the API:
[snip]:# (bq-2)
```go
// Construct a query.
q := c.Query(`
    SELECT year, SUM(number)
    FROM [bigquery-public-data:usa_names.usa_1910_2013]
    WHERE name = "William"
    GROUP BY year
    ORDER BY year
`)
// Execute the query.
it, err := q.Read(ctx)
if err != nil {
	// TODO: Handle error.
}
// Iterate through the results.
for {
	var values []bigquery.Value
	err := it.Next(&values)
	if err == iterator.Done {  // from "google.golang.org/api/iterator"
		break
	}
	if err != nil {
		// TODO: Handle error.
	}
	fmt.Println(values)
}
```
