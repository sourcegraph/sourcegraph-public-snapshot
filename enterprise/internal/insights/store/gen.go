package store

//go:generate ../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store -i Interface -o mock_store_interface.go
//go:generate ../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store -i DataSeriesStore -o mock_store_dataseriesstore.go
//go:generate ../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store -i InsightMetadataStore -o mock_store_insightmetadatastore.go
