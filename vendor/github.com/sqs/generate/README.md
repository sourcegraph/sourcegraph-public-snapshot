# generate

Generates Go (golang) Structs from JSON schema.

# Requirements

* Go 1.8+

# Examples

```javascript
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "title": "Example",
  "id": "Example",
  "type": "object",
  "description": "example",
  "definitions": {
    "address": {
      "id": "address",
      "type": "object",
      "description": "Address",
      "properties": {
        "houseName": {
          "type": "string",
          "description": "House Name",
          "maxLength": 30
        },
        "houseNumber": {
          "type": "string",
          "description": "House Number",
          "maxLength": 4
        },
        "flatNumber": {
          "type": "string",
          "description": "Flat",
          "maxLength": 15
        },
        "street": {
          "type": "string",
          "description": "Address 1",
          "maxLength": 40
        },
        "district": {
          "type": "string",
          "description": "Address 2",
          "maxLength": 30
        },
        "town": {
          "type": "string",
          "description": "City",
          "maxLength": 20
        },
        "county": {
          "type": "string",
          "description": "County",
          "maxLength": 20
        },
        "postcode": {
          "type": "string",
          "description": "Postcode",
          "maxLength": 8
        }
      }
    },
    "status": {
      "type": "object",
      "properties": {
        "favouritecat": {
          "enum": [
            "A",
            "B",
            "C",
            "D",
            "E",
            "F"
          ],
          "type": "string",
          "description": "The favourite cat.",
          "maxLength": 1
        }
      }
    }
  },
  "properties": {
    "name": {
      "type": "string"
    },
    "address": {
      "$ref": "#/definitions/address"
    },
    "status": {
      "$ref": "#/definitions/status"
    }
  }
}
```

Run this:

```bash
go run main.go exampleschema.json
```

Get this:

```go
package main

type Address struct {
  County string `json:"county,omitempty"`
  District string `json:"district,omitempty"`
  FlatNumber string `json:"flatNumber,omitempty"`
  HouseName string `json:"houseName,omitempty"`
  HouseNumber string `json:"houseNumber,omitempty"`
  Postcode string `json:"postcode,omitempty"`
  Street string `json:"street,omitempty"`
  Town string `json:"town,omitempty"`
}

type Example struct {
  Address *Address `json:"address,omitempty"`
  Name string `json:"name,omitempty"`
  Status *Status `json:"status,omitempty"`
}

type Status struct {
  Favouritecat string `json:"favouritecat,omitempty"`
}
```
