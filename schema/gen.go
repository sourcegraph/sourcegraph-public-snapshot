package schema

//go:generate env GO111MODULE=on go run stringdata.go -i actions.schema.json -name ActionSchemaJSON -pkg schema -o action_stringdata.go
//go:generate gofmt -s -w action_stringdata.go

//go:generate env GO111MODULE=on go run stringdata.go -i campaign_spec.schema.json -name CampaignSpecJSON -pkg schema -o campaign_spec_stringdata.go
//go:generate gofmt -s -w campaign_spec_stringdata.go
