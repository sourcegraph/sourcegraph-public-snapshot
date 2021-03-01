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
      "items": { "type": "string", "pattern": "^\\/[\\/\\S]+\\/$" },
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
      "properties": {},
      "hide": true
    },
    "repositoryPathPattern": {
      "description": "The pattern used to generate the corresponding Sourcegraph repository name for a Perforce depot. In the pattern, the variable \"{depot}\" is replaced with the Perforce depot's path.\n\nFor example, if your Perforce depot path is \"//Sourcegraph/\" and your Sourcegraph URL is https://src.example.com, then a repositoryPathPattern of \"perforce/{depot}\" would mean that the Perforce depot is available on Sourcegraph at https://src.example.com/perforce/Sourcegraph.\n\nIt is important that the Sourcegraph repository name generated with this pattern be unique to this Perforce Server. If different Perforce Servers generate repository names that collide, Sourcegraph's behavior is undefined.",
      "type": "string",
      "default": "{depot}"
    }
  }
}
`
