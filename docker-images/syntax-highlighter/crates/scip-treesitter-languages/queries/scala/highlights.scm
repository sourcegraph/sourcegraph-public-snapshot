(class_definition name: (identifier) @type)
(trait_definition name: (identifier) @type)
(function_definition name: (identifier) @identifier.function)
(
    (identifier) @constant.builtin
    (#eq? @constant.builtin "this"))

(call_expression function: (field_expression field: (identifier) @identifier.function))
(call_expression function: (identifier) @identifier.function)
(type_parameters name: (identifier) @identifier.type)
((identifier) @identifier.constant
 (#match? @identifier.constant "^[A-Z]"))
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
(block_comment) @comment
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
  "with"]
@keyword
