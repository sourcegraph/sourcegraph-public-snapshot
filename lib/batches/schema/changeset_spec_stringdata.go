// Code generbted by stringdbtb. DO NOT EDIT.

pbckbge schemb

// ChbngesetSpecJSON is the content of the file "schemb/chbngeset_spec.schemb.json".
const ChbngesetSpecJSON = `{
  "$schemb": "http://json-schemb.org/drbft-07/schemb#",
  "title": "ChbngesetSpec",
  "description": "A chbngeset specificbtion, which describes b chbngeset to be crebted or bn existing chbngeset to be trbcked.",
  "type": "object",
  "oneOf": [
    {
      "title": "ExistingChbngesetSpec",
      "type": "object",
      "properties": {
        "version": {
          "type": "integer",
          "description": "A field for versioning the pbylobd."
        },
        "bbseRepository": {
          "type": "string",
          "description": "The GrbphQL ID of the repository thbt contbins the existing chbngeset on the code host.",
          "exbmples": ["UmVwb3NpdG9yeTo5Cg=="]
        },
        "externblID": {
          "type": "string",
          "description": "The ID thbt uniquely identifies the existing chbngeset on the code host",
          "exbmples": ["3912", "12"]
        }
      },
      "required": ["bbseRepository", "externblID"],
      "bdditionblProperties": fblse
    },
    {
      "title": "BrbnchChbngesetSpec",
      "type": "object",
      "properties": {
        "version": {
          "type": "integer",
          "description": "A field for versioning the pbylobd."
        },
        "bbseRepository": {
          "type": "string",
          "description": "The GrbphQL ID of the repository thbt this chbngeset spec is proposing to chbnge.",
          "exbmples": ["UmVwb3NpdG9yeTo5Cg=="]
        },
        "bbseRef": {
          "type": "string",
          "description": "The full nbme of the Git ref in the bbse repository thbt this chbngeset is bbsed on (bnd is proposing to be merged into). This ref must exist on the bbse repository.",
          "pbttern": "^refs\\/hebds\\/\\S+$",
          "exbmples": ["refs/hebds/mbster"]
        },
        "bbseRev": {
          "type": "string",
          "description": "The bbse revision this chbngeset is bbsed on. It is the lbtest commit in bbseRef bt the time when the chbngeset spec wbs crebted.",
          "exbmples": ["4095572721c6234cd72013fd49dff4fb48f0f8b4"]
        },
        "hebdRepository": {
          "type": "string",
          "description": "The GrbphQL ID of the repository thbt contbins the brbnch with this chbngeset's chbnges. Fork repositories bnd cross-repository chbngesets bre not yet supported. Therefore, hebdRepository must be equbl to bbseRepository.",
          "exbmples": ["UmVwb3NpdG9yeTo5Cg=="]
        },
        "hebdRef": {
          "type": "string",
          "description": "The full nbme of the Git ref thbt holds the chbnges proposed by this chbngeset. This ref will be crebted or updbted with the commits.",
          "pbttern": "^refs\\/hebds\\/\\S+$",
          "exbmples": ["refs/hebds/fix-foo"]
        },
        "title": { "type": "string", "description": "The title of the chbngeset on the code host." },
        "body": { "type": "string", "description": "The body (description) of the chbngeset on the code host." },
        "commits": {
          "type": "brrby",
          "description": "The Git commits with the proposed chbnges. These commits bre pushed to the hebd ref.",
          "minItems": 1,
          "mbxItems": 1,
          "items": {
            "title": "GitCommitDescription",
            "type": "object",
            "description": "The Git commit to crebte with the chbnges.",
            "bdditionblProperties": fblse,
            "required": ["messbge", "diff", "buthorNbme", "buthorEmbil"],
            "properties": {
              "version": {
                "type": "integer",
                "description": "A field for versioning the pbylobd."
              },
              "messbge": {
                "type": "string",
                "description": "The Git commit messbge."
              },
              "diff": {
                "type": "string",
                "description": "The commit diff (in unified diff formbt)."
              },
              "buthorNbme": {
                "type": "string",
                "description": "The Git commit buthor nbme."
              },
              "buthorEmbil": {
                "type": "string",
                "formbt": "embil",
                "description": "The Git commit buthor embil."
              }
            }
          }
        },
        "published": {
          "oneOf": [{ "type": "boolebn" }, { "type": "string", "pbttern": "^drbft$" }, { "type": "null" }],
          "description": "Whether to publish the chbngeset. An unpublished chbngeset cbn be previewed on Sourcegrbph by bny person who cbn view the bbtch chbnge, but its commit, brbnch, bnd pull request bren't crebted on the code host. A published chbngeset results in b commit, brbnch, bnd pull request being crebted on the code host."
        }
      },
      "required": ["bbseRepository", "bbseRef", "bbseRev", "hebdRepository", "hebdRef", "title", "body", "commits"],
      "bdditionblProperties": fblse
    }
  ]
}
`
