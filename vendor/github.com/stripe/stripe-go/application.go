package stripe

import "encoding/json"

type Application struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// UnmarshalJSON handles deserialization of an Application.
// This custom unmarshaling is needed because the resulting
// property may be an id or the full struct if it was expanded.
func (a *Application) UnmarshalJSON(data []byte) error {
	type application Application
	var aa application
	err := json.Unmarshal(data, &aa)
	if err == nil {
		*a = Application(aa)
	} else {
		// the id is surrounded by "\" characters, so strip them
		a.ID = string(data[1 : len(data)-1])
	}

	return nil
}
