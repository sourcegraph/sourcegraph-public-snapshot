package types

type ProviderID string

type APIVersionID string

type ModelID string

// `${ProviderID}::${APIVersionID}`
type ProviderApiVersionRef string

// `${ProviderID}::${APIVersionID}::${ModelID}`
type ModelRef string
