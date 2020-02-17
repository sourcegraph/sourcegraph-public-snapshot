// Code generated by stringdata. DO NOT EDIT.

package schema

// BitbucketServerSchemaJSON is the content of the file "bitbucket_server.schema.json".
const BitbucketServerSchemaJSON = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "bitbucket_server.schema.json#",
  "title": "BitbucketServerConnection",
  "description": "Configuration for a connection to Bitbucket Server.",
  "allowComments": true,
  "type": "object",
  "additionalProperties": false,
  "required": ["username", "url"],
  "oneOf": [
    {
      "required": ["token"],
      "properties": {
        "password": { "type": "null" }
      }
    },
    {
      "required": ["password"],
      "properties": {
        "token": { "type": "null" }
      }
    }
  ],
  "properties": {
    "url": {
      "description": "URL of a Bitbucket Server instance, such as https://bitbucket.example.com.",
      "type": "string",
      "pattern": "^https?://",
      "not": {
        "type": "string",
        "pattern": "example\\.com"
      },
      "format": "uri",
      "examples": ["https://bitbucket.example.com"]
    },
    "token": {
      "description": "A Bitbucket Server personal access token with Read scope. Create one at https://[your-bitbucket-hostname]/plugins/servlet/access-tokens/add. Also set the corresponding \"username\" field.\n\nFor Bitbucket Server instances that don't support personal access tokens (Bitbucket Server version 5.4 and older), specify user-password credentials in the \"username\" and \"password\" fields.",
      "type": "string",
      "minLength": 1
    },
    "username": {
      "description": "The username to use when authenticating to the Bitbucket Server instance. Also set the corresponding \"token\" or \"password\" field.",
      "type": "string"
    },
    "password": {
      "description": "The password to use when authenticating to the Bitbucket Server instance. Also set the corresponding \"username\" field.\n\nFor Bitbucket Server instances that support personal access tokens (Bitbucket Server version 5.5 and newer), it is recommended to provide a token instead (in the \"token\" field).",
      "type": "string"
    },
    "gitURLType": {
      "description": "The type of Git URLs to use for cloning and fetching Git repositories on this Bitbucket Server instance.\n\nIf \"http\", Sourcegraph will access Bitbucket Server repositories using Git URLs of the form http(s)://bitbucket.example.com/scm/myproject/myrepo.git (using https: if the Bitbucket Server instance uses HTTPS).\n\nIf \"ssh\", Sourcegraph will access Bitbucket Server repositories using Git URLs of the form ssh://git@example.bitbucket.com/myproject/myrepo.git. See the documentation for how to provide SSH private keys and known_hosts: https://docs.sourcegraph.com/admin/repo/auth#repositories-that-need-http-s-or-ssh-authentication.",
      "type": "string",
      "enum": ["http", "ssh"],
      "default": "http",
      "examples": ["ssh"]
    },
    "certificate": {
      "description": "TLS certificate of the Bitbucket Server instance. This is only necessary if the certificate is self-signed or signed by an internal CA. To get the certificate run ` + "`" + `openssl s_client -connect HOST:443 -showcerts < /dev/null 2> /dev/null | openssl x509 -outform PEM` + "`" + `. To escape the value into a JSON string, you may want to use a tool like https://json-escape-text.now.sh.",
      "type": "string",
      "pattern": "^-----BEGIN CERTIFICATE-----\n",
      "examples": ["-----BEGIN CERTIFICATE-----\n..."]
    },
    "webhooks": {
      "description": "DEPRECATED: Switch to \"plugin.webhooks\"",
      "type": "object",
      "properties": {
        "secret": {
          "description": "Secret for authenticating incoming webhook payloads",
          "type": "string",
          "minLength": 1
        }
      }
    },
    "plugin": {
      "title": "BitbucketServerPlugin",
      "description": "Configuration for Bitbucket Server Sourcegraph plugin",
      "type": "object",
      "properties": {
        "webhooks": {
          "title": "BitbucketServerPluginWebhooks",
          "type": "object",
          "required": ["secret"],
          "properties": {
            "secret": {
              "description": "Secret for authenticating incoming webhook payloads",
              "type": "string",
              "minLength": 1
            }
          }
        },
        "permissions": {
          "description": "Enables fetching Bitbucket Server permissions through the roaring bitmap endpoint. Warning: there may be performance degradation under significant load.",
          "type": "string",
          "enum": ["enabled", "disabled"],
          "default": "disabled"
        }
      }
    },
    "repositoryPathPattern": {
      "description": "The pattern used to generate the corresponding Sourcegraph repository name for a Bitbucket Server repository.\n\n - \"{host}\" is replaced with the Bitbucket Server URL's host (such as bitbucket.example.com)\n - \"{projectKey}\" is replaced with the Bitbucket repository's parent project key (such as \"PRJ\")\n - \"{repositorySlug}\" is replaced with the Bitbucket repository's slug key (such as \"my-repo\").\n\nFor example, if your Bitbucket Server is https://bitbucket.example.com and your Sourcegraph is https://src.example.com, then a repositoryPathPattern of \"{host}/{projectKey}/{repositorySlug}\" would mean that a Bitbucket Server repository at https://bitbucket.example.com/projects/PRJ/repos/my-repo is available on Sourcegraph at https://src.example.com/bitbucket.example.com/PRJ/my-repo.\n\nIt is important that the Sourcegraph repository name generated with this pattern be unique to this code host. If different code hosts generate repository names that collide, Sourcegraph's behavior is undefined.",
      "type": "string",
      "default": "{host}/{projectKey}/{repositorySlug}",
      "examples": ["{projectKey}/{repositorySlug}"]
    },
    "excludePersonalRepositories": {
      "description": "Whether or not personal repositories should be excluded or not. When true, Sourcegraph will ignore personal repositories it may have access to. See https://docs.sourcegraph.com/integration/bitbucket_server#excluding-personal-repositories for more information.",
      "type": "boolean",
      "default": false
    },
    "repositoryQuery": {
      "description": "An array of strings specifying which repositories to mirror on Sourcegraph. Each string is a URL query string with parameters that filter the list of returned repos. Examples: \"?name=my-repo&projectname=PROJECT&visibility=private\".\n\nThe special string \"none\" can be used as the only element to disable this feature. Repositories matched by multiple query strings are only imported once. Here's the official Bitbucket Server documentation about which query string parameters are valid: https://docs.atlassian.com/bitbucket-server/rest/6.1.2/bitbucket-rest.html#idp355",
      "type": "array",
      "items": {
        "type": "string",
        "minLength": 1
      },
      "default": ["none"],
      "minItems": 1,
      "examples": [["?name=my-repo&projectname=PROJECT&visibility=private"]]
    },
    "repos": {
      "description": "An array of repository \"projectKey/repositorySlug\" strings specifying repositories to mirror on Sourcegraph.",
      "type": "array",
      "minItems": 1,
      "items": {
        "type": "string",
        "pattern": "^[\\w-]+/[\\w.-]+$"
      },
      "examples": [["myproject/myrepo", "myproject/myotherrepo"]]
    },
    "exclude": {
      "description": "A list of repositories to never mirror from this Bitbucket Server instance. Takes precedence over \"repos\" and \"repositoryQuery\".\n\nSupports excluding by name ({\"name\": \"projectKey/repositorySlug\"}) or by ID ({\"id\": 42}).",
      "type": "array",
      "minItems": 1,
      "items": {
        "type": "object",
        "title": "ExcludedBitbucketServerRepo",
        "additionalProperties": false,
        "anyOf": [{ "required": ["name"] }, { "required": ["id"] }, { "required": ["pattern"] }],
        "properties": {
          "name": {
            "description": "The name of a Bitbucket Server repo (\"projectKey/repositorySlug\") to exclude from mirroring.",
            "type": "string",
            "pattern": "^[\\w-]+/[\\w.-]+$"
          },
          "id": {
            "description": "The ID of a Bitbucket Server repo (as returned by the Bitbucket Server instance's API) to exclude from mirroring.",
            "type": "integer"
          },
          "pattern": {
            "description": "Regular expression which matches against the name of a Bitbucket Server repo.",
            "type": "string",
            "format": "regex"
          }
        }
      },
      "examples": [
        [{ "name": "myproject/myrepo" }, { "id": 42 }],
        [{ "name": "myproject/myrepo" }, { "name": "myproject/myotherrepo" }, { "pattern": "^topsecretproject/.*" }]
      ]
    },
    "initialRepositoryEnablement": {
      "description": "Defines whether repositories from this Bitbucket Server instance should be enabled and cloned when they are first seen by Sourcegraph. If false, the site admin must explicitly enable Bitbucket Server repositories (in the site admin area) to clone them and make them searchable on Sourcegraph. If true, they will be enabled and cloned immediately (subject to rate limiting by Bitbucket Server); site admins can still disable them explicitly, and they'll remain disabled.",
      "type": "boolean",
      "default": false
    },
    "authorization": {
      "title": "BitbucketServerAuthorization",
      "description": "If non-null, enforces Bitbucket Server repository permissions.",
      "type": "object",
      "additionalProperties": false,
      "required": ["identityProvider", "oauth"],
      "properties": {
        "identityProvider": {
          "description": "The source of identity to use when computing permissions. This defines how to compute the Bitbucket Server identity to use for a given Sourcegraph user. When 'username' is used, Sourcegraph assumes usernames are identical in Sourcegraph and Bitbucket Server accounts and ` + "`" + `auth.enableUsernameChanges` + "`" + ` must be set to false for security reasons.",
          "title": "BitbucketServerIdentityProvider",
          "type": "object",
          "required": ["type"],
          "properties": {
            "type": {
              "type": "string",
              "enum": ["username"]
            }
          },
          "oneOf": [{ "$ref": "#/definitions/UsernameIdentity" }],
          "!go": {
            "taggedUnionType": true
          }
        },
        "oauth": {
          "title": "BitbucketServerOAuth",
          "type": "object",
          "additionalProperties": false,
          "required": ["consumerKey", "signingKey"],
          "description": "OAuth configuration specified when creating the Bitbucket Server Application Link with incoming authentication. Two Legged OAuth with 'ExecuteAs=admin' must be enabled as well as user impersonation.",
          "properties": {
            "consumerKey": {
              "description": "The OAuth consumer key specified when creating the Bitbucket Server Application Link with incoming authentication.",
              "type": "string",
              "minLength": 1
            },
            "signingKey": {
              "description": "Base64 encoding of the OAuth PEM encoded RSA private key used to generate the public key specified when creating the Bitbucket Server Application Link with incoming authentication.",
              "type": "string",
              "minLength": 1
            }
          }
        },
        "ttl": {
          "description": "Duration after which a user's cached permissions will be updated in the background (during which time the previously cached permissions will be used). This is 3 hours by default.\n\nDecreasing the TTL will increase the load on the code host API. If you have X repos on your instance, it will take ~X/1000 API requests to fetch the complete list for 1 user.  If you have Y users, you will incur X*Y/1000 API requests per cache refresh period.\n\nIf set to zero, Sourcegraph will sync a user's entire accessible repository list on every request (NOT recommended).",
          "type": "string",
          "default": "3h"
        },
        "hardTTL": {
          "description": "Duration after which a user's cached permissions must be updated before authorizing any user actions. This is 3 days by default.",
          "type": "string",
          "default": "72h"
        }
      }
    }
  },
  "definitions": {
    "UsernameIdentity": {
      "title": "BitbucketServerUsernameIdentity",
      "type": "object",
      "additionalProperties": false,
      "required": ["type"],
      "properties": {
        "type": {
          "type": "string",
          "const": "username"
        }
      }
    }
  }
}
`
