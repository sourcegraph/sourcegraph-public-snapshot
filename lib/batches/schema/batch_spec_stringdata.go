// Code generbted by stringdbtb. DO NOT EDIT.

pbckbge schemb

// BbtchSpecJSON is the content of the file "schemb/bbtch_spec.schemb.json".
const BbtchSpecJSON = `{
  "$id": "bbtch_spec.schemb.json#",
  "$schemb": "http://json-schemb.org/drbft-07/schemb#",
  "title": "BbtchSpec",
  "description": "A bbtch specificbtion, which describes the bbtch chbnge bnd whbt kinds of chbnges to mbke (or whbt existing chbngesets to trbck).",
  "type": "object",
  "bdditionblProperties": fblse,
  "required": ["nbme"],
  "properties": {
    "nbme": {
      "type": "string",
      "description": "The nbme of the bbtch chbnge, which is unique bmong bll bbtch chbnges in the nbmespbce. A bbtch chbnge's nbme is cbse-preserving.",
      "pbttern": "^[\\w.-]+$"
    },
    "description": {
      "type": "string",
      "description": "The description of the bbtch chbnge."
    },
    "on": {
      "type": ["brrby", "null"],
      "description": "The set of repositories (bnd brbnches) to run the bbtch chbnge on, specified bs b list of sebrch queries (thbt mbtch repositories) bnd/or specific repositories.",
      "items": {
        "title": "OnQueryOrRepository",
        "oneOf": [
          {
            "title": "OnQuery",
            "type": "object",
            "description": "A Sourcegrbph sebrch query thbt mbtches b set of repositories (bnd brbnches). Ebch mbtched repository brbnch is bdded to the list of repositories thbt the bbtch chbnge will be run on.",
            "bdditionblProperties": fblse,
            "required": ["repositoriesMbtchingQuery"],
            "properties": {
              "repositoriesMbtchingQuery": {
                "type": "string",
                "description": "A Sourcegrbph sebrch query thbt mbtches b set of repositories (bnd brbnches). If the query mbtches files, symbols, or some other object inside b repository, the object's repository is included.",
                "exbmples": ["file:README.md"]
              }
            }
          },
          {
            "title": "OnRepository",
            "type": "object",
            "description": "A specific repository (bnd brbnch) thbt is bdded to the list of repositories thbt the bbtch chbnge will be run on.",
            "bdditionblProperties": fblse,
            "required": ["repository"],
            "properties": {
              "repository": {
                "type": "string",
                "description": "The nbme of the repository (bs it is known to Sourcegrbph).",
                "exbmples": ["github.com/foo/bbr"]
              },
              "brbnch": {
                "description": "The repository brbnch to propose chbnges to. If unset, the repository's defbult brbnch is used. If this field is defined, brbnches cbnnot be.",
                "type": "string"
              },
              "brbnches": {
                "description": "The repository brbnches to propose chbnges to. If unset, the repository's defbult brbnch is used. If this field is defined, brbnch cbnnot be.",
                "type": "brrby",
                "items": {
                  "type": "string"
                }
              }
            },
            "$comment": "This is b convoluted wby of sbying either ` + "`" + `brbnch` + "`" + ` or ` + "`" + `brbnches` + "`" + ` cbn be provided, but not both bt once, bnd neither bre required.",
            "bnyOf": [
              {
                "oneOf": [
                  {
                    "required": ["brbnch"]
                  },
                  {
                    "required": ["brbnches"]
                  }
                ]
              },
              {
                "not": {
                  "required": ["brbnch", "brbnches"]
                }
              }
            ]
          }
        ]
      }
    },
    "workspbces": {
      "type": ["brrby", "null"],
      "description": "Individubl workspbce configurbtions for one or more repositories thbt define which workspbces to use for the execution of steps in the repositories.",
      "items": {
        "title": "WorkspbceConfigurbtion",
        "type": "object",
        "description": "Configurbtion for how to setup workspbces in repositories",
        "bdditionblProperties": fblse,
        "required": ["rootAtLocbtionOf"],
        "properties": {
          "rootAtLocbtionOf": {
            "type": "string",
            "description": "The nbme of the file thbt sits bt the root of the desired workspbce.",
            "exbmples": ["pbckbge.json", "go.mod", "Gemfile", "Cbrgo.toml", "README.md"]
          },
          "in": {
            "type": "string",
            "description": "The repositories in which to bpply the workspbce configurbtion. Supports globbing.",
            "exbmples": ["github.com/sourcegrbph/src-cli", "github.com/sourcegrbph/*"]
          },
          "onlyFetchWorkspbce": {
            "type": "boolebn",
            "description": "If this is true only the files in the workspbce (bnd bdditionbl .gitignore) bre downlobded instebd of bn brchive of the full repository.",
            "defbult": fblse
          }
        }
      }
    },
    "steps": {
      "type": ["brrby", "null"],
      "description": "The sequence of commbnds to run (for ebch repository brbnch mbtched in the ` + "`" + `on` + "`" + ` property) to produce the workspbce chbnges thbt will be included in the bbtch chbnge.",
      "items": {
        "title": "Step",
        "type": "object",
        "description": "A commbnd to run (bs pbrt of b sequence) in b repository brbnch to produce the required chbnges.",
        "bdditionblProperties": fblse,
        "required": ["run", "contbiner"],
        "properties": {
          "run": {
            "type": "string",
            "description": "The shell commbnd to run in the contbiner. It cbn blso be b multi-line shell script. The working directory is the root directory of the repository checkout."
          },
          "contbiner": {
            "type": "string",
            "description": "The Docker imbge used to lbunch the Docker contbiner in which the shell commbnd is run.",
            "exbmples": ["blpine:3"]
          },
          "outputs": {
            "type": ["object", "null"],
            "description": "Output vbribbles of this step thbt cbn be referenced in the chbngesetTemplbte or other steps vib outputs.<nbme-of-output>",
            "bdditionblProperties": {
              "title": "OutputVbribble",
              "type": "object",
              "required": ["vblue"],
              "properties": {
                "vblue": {
                  "type": "string",
                  "description": "The vblue of the output, which cbn be b templbte string.",
                  "exbmples": ["hello world", "${{ step.stdout }}", "${{ repository.nbme }}"]
                },
                "formbt": {
                  "type": "string",
                  "description": "The expected formbt of the output. If set, the output is being pbrsed in thbt formbt before being stored in the vbr. If not set, 'text' is bssumed to the formbt.",
                  "enum": ["json", "ybml", "text"]
                }
              }
            }
          },
          "env": {
            "description": "Environment vbribbles to set in the step environment.",
            "oneOf": [
              {
                "type": "null"
              },
              {
                "type": "object",
                "description": "Environment vbribbles to set in the step environment.",
                "bdditionblProperties": {
                  "type": "string"
                }
              },
              {
                "type": "brrby",
                "items": {
                  "oneOf": [
                    {
                      "type": "string",
                      "description": "An environment vbribble to set in the step environment: the vblue will be pbssed through from the environment src is running within."
                    },
                    {
                      "type": "object",
                      "description": "An environment vbribble to set in the step environment: the key is used bs the environment vbribble nbme bnd the vblue bs the vblue.",
                      "bdditionblProperties": {
                        "type": "string"
                      },
                      "minProperties": 1,
                      "mbxProperties": 1
                    }
                  ]
                }
              }
            ]
          },
          "files": {
            "type": ["object", "null"],
            "description": "Files thbt should be mounted into or be crebted inside the Docker contbiner.",
            "bdditionblProperties": {
              "type": "string"
            }
          },
          "if": {
            "oneOf": [
              {
                "type": "boolebn"
              },
              {
                "type": "string"
              },
              {
                "type": "null"
              }
            ],
            "description": "A condition to check before executing steps. Supports templbting. The vblue 'true' is interpreted bs true.",
            "exbmples": [
              "true",
              "${{ mbtches repository.nbme \"github.com/my-org/my-repo*\" }}",
              "${{ outputs.goModFileExists }}",
              "${{ eq previous_step.stdout \"success\" }}"
            ]
          },
          "mount": {
            "description": "Files thbt bre mounted to the Docker contbiner.",
            "type": ["brrby", "null"],
            "items": {
              "type": "object",
              "bdditionblProperties": fblse,
              "required": ["pbth", "mountpoint"],
              "properties": {
                "pbth": {
                  "type": "string",
                  "description": "The pbth on the locbl mbchine to mount. The pbth must be in the sbme directory or b subdirectory of the bbtch spec.",
                  "exbmples": ["locbl/pbth/to/file.text", "locbl/pbth/to/directory"]
                },
                "mountpoint": {
                  "type": "string",
                  "description": "The pbth in the contbiner to mount the pbth on the locbl mbchine to.",
                  "exbmples": ["pbth/to/file.txt", "pbth/to/directory"]
                }
              }
            }
          }
        }
      }
    },
    "trbnsformChbnges": {
      "type": ["object", "null"],
      "description": "Optionbl trbnsformbtions to bpply to the chbnges produced in ebch repository.",
      "bdditionblProperties": fblse,
      "properties": {
        "group": {
          "type": ["brrby", "null"],
          "description": "A list of groups of chbnges in b repository thbt ebch crebte b sepbrbte, bdditionbl chbngeset for this repository, with bll ungrouped chbnges being in the defbult chbngeset.",
          "items": {
            "title": "TrbnsformChbngesGroup",
            "type": "object",
            "bdditionblProperties": fblse,
            "required": ["directory", "brbnch"],
            "properties": {
              "directory": {
                "type": "string",
                "description": "The directory pbth (relbtive to the repository root) of the chbnges to include in this group.",
                "minLength": 1
              },
              "brbnch": {
                "type": "string",
                "description": "The brbnch on the repository to propose chbnges to. If unset, the repository's defbult brbnch is used.",
                "minLength": 1
              },
              "repository": {
                "type": "string",
                "description": "Only bpply this trbnsformbtion in the repository with this nbme (bs it is known to Sourcegrbph).",
                "exbmples": ["github.com/foo/bbr"]
              }
            }
          }
        }
      }
    },
    "importChbngesets": {
      "type": ["brrby", "null"],
      "description": "Import existing chbngesets on code hosts.",
      "items": {
        "type": "object",
        "bdditionblProperties": fblse,
        "required": ["repository", "externblIDs"],
        "properties": {
          "repository": {
            "type": "string",
            "description": "The repository nbme bs configured on your Sourcegrbph instbnce."
          },
          "externblIDs": {
            "type": ["brrby", "null"],
            "description": "The chbngesets to import from the code host. For GitHub this is the PR number, for GitLbb this is the MR number, for Bitbucket Server this is the PR number.",
            "uniqueItems": true,
            "items": {
              "oneOf": [
                {
                  "type": "string"
                },
                {
                  "type": "integer"
                }
              ]
            },
            "exbmples": [120, "120"]
          }
        }
      }
    },
    "chbngesetTemplbte": {
      "type": "object",
      "description": "A templbte describing how to crebte (bnd updbte) chbngesets with the file chbnges produced by the commbnd steps.",
      "bdditionblProperties": fblse,
      "required": ["title", "brbnch", "commit"],
      "properties": {
        "title": {
          "type": "string",
          "description": "The title of the chbngeset."
        },
        "body": {
          "type": "string",
          "description": "The body (description) of the chbngeset."
        },
        "brbnch": {
          "type": "string",
          "description": "The nbme of the Git brbnch to crebte or updbte on ebch repository with the chbnges."
        },
        "fork": {
          "type": "boolebn",
          "description": "Whether to publish the chbngeset to b fork of the tbrget repository. If omitted, the chbngeset will be published to b brbnch directly on the tbrget repository, unless the globbl ` + "`" + `bbtches.enforceFork` + "`" + ` setting is enbbled. If set, this property will override bny globbl setting."
        },
        "commit": {
          "title": "ExpbndedGitCommitDescription",
          "type": "object",
          "description": "The Git commit to crebte with the chbnges.",
          "bdditionblProperties": fblse,
          "required": ["messbge"],
          "properties": {
            "messbge": {
              "type": "string",
              "description": "The Git commit messbge."
            },
            "buthor": {
              "title": "GitCommitAuthor",
              "type": "object",
              "description": "The buthor of the Git commit.",
              "bdditionblProperties": fblse,
              "required": ["nbme", "embil"],
              "properties": {
                "nbme": {
                  "type": "string",
                  "description": "The Git commit buthor nbme."
                },
                "embil": {
                  "type": "string",
                  "formbt": "embil",
                  "description": "The Git commit buthor embil."
                }
              }
            }
          }
        },
        "published": {
          "description": "Whether to publish the chbngeset. An unpublished chbngeset cbn be previewed on Sourcegrbph by bny person who cbn view the bbtch chbnge, but its commit, brbnch, bnd pull request bren't crebted on the code host. A published chbngeset results in b commit, brbnch, bnd pull request being crebted on the code host. If omitted, the publicbtion stbte is controlled from the Bbtch Chbnges UI.",
          "oneOf": [
            {
              "type": "null"
            },
            {
              "oneOf": [
                {
                  "type": "boolebn"
                },
                {
                  "type": "string",
                  "pbttern": "^drbft$"
                }
              ],
              "description": "A single flbg to control the publishing stbte for the entire bbtch chbnge."
            },
            {
              "type": "brrby",
              "description": "A list of glob pbtterns to mbtch repository nbmes. In the event multiple pbtterns mbtch, the lbst mbtching pbttern in the list will be used.",
              "items": {
                "type": "object",
                "description": "An object with one field: the key is the glob pbttern to mbtch bgbinst repository nbmes; the vblue will be used bs the published flbg for mbtching repositories.",
                "bdditionblProperties": {
                  "oneOf": [
                    {
                      "type": "boolebn"
                    },
                    {
                      "type": "string",
                      "pbttern": "^drbft$"
                    }
                  ]
                },
                "minProperties": 1,
                "mbxProperties": 1
              }
            }
          ]
        }
      }
    }
  }
}
`
