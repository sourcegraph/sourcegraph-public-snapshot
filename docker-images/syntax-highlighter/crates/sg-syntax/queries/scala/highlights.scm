(class_definition name: (identifier) @type)
(trait_definition name: (identifier) @type)
(
    (identifier) @constant.builtin
    (#eq? @constant.builtin "this")
)
(identifier) @variable
(case_class_pattern type: (type_identifier) @variable)
(operator_identifier) @variable
(type_identifier) @type
"this" @constant.builtin
(interpolated_string) @string
(string) @string
(character_literal) @character
(floating_point_literal) @float
(integer_literal) @number
(null_literal) @constant.null
(boolean_literal) @boolean
(comment) @comment
[
  "case"
  "match"
  "class"
  "object"
  "trait"
  "enum"
  "type"
  "package"
  "import"
  "def"
  "sealed"
  "final"
  "implicit"
  "override"
  "var"
  "val"
  "lazy"
  "extends"
  "derives"
  "new"
  "for"
  "while"
  "do"
  "if"
  "else"
  "then"
  "return"
  "throw"
  "try"
  "catch"
  "finally"
  "abstract"
  "private"
  "protected"
  "using"
  ;; "opaque"
  ;; "transparent"
  ;; "inline"
  "trait"
  "given"
  "end"
  "extension"
  "with"
] @keyword
