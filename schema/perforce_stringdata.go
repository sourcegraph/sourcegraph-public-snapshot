// Code generated by stringdata. DO NOT EDIT.

package schema

// PerforceSchemaJSON is the content of the file "perforce.schema.json".
const PerforceSchemaJSON = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "perforce.schema.json#",
  "title": "PerforceConnection",
  "description": "Configuration for a connection to Perforce Server.",
  "allowComments": true,
  "type": "object",
  "additionalProperties": false,
  "required": ["p4.port", "p4.user", "p4.passwd"],
  "properties": {
    "p4.port": {
      "description": "The Perforce Server address to be used for p4 CLI (P4PORT).",
      "type": "string",
      "examples": ["ssl:111.222.333.444:1666"]
    },
    "p4.user": {
      "description": "The user to be authenticated for p4 CLI (P4USER).",
      "type": "string",
      "examples": ["admin"]
    },
    "p4.passwd": {
      "description": "The plain password of the user (P4PASSWD).",
      "type": "string"
    },
    "depots": {
      "description": "Depots can have arbitrary paths, e.g. a path to depot root or a subdirectory.",
      "type": "array",
      "items": { "type": "string", "pattern": "^\\/[\\/\\w]+/$" },
      "examples": [["//Sourcegraph/", "//Engineering/Cloud/"]]
    },
    "maxChanges": {
      "description": "Only import at most n changes when possible (git p4 clone --max-changes).",
      "type": "number",
      "default": 1000,
      "minimum": 1
    },
    "authorization": {
      "title": "PerforceAuthorization",
      "description": "If non-null, enforces Perforce depot permissions.",
      "type": "object",
      "properties": {}
    }
  }
}
`
