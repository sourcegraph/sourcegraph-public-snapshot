; inherits: ecma,jsx

; Types

; Javascript

; Variables
;-----------
(identifier) @variable

; Properties
;-----------

(property_identifier) @property
(shorthand_property_identifier) @property
(private_property_identifier) @property

(variable_declarator
  name: (object_pattern
    (shorthand_property_identifier_pattern))) @variable

; Special identifiers
;--------------------

((identifier) @constructor
 (#lua-match? @constructor "^[A-Z]"))

((identifier) @constant
 (#lua-match? @constant "^[A-Z_][A-Z%d_]+$"))

((shorthand_property_identifier) @constant
 (#lua-match? @constant "^[A-Z_][A-Z%d_]+$"))

((identifier) @variable.builtin
 (#vim-match? @variable.builtin "^(arguments|module|console|window|document)$"))

((identifier) @function.builtin
 (#eq? @function.builtin "require"))

; Function and method definitions
;--------------------------------

(function
  name: (identifier) @function)
(function_declaration
  name: (identifier) @function)
(generator_function
  name: (identifier) @function)
(generator_function_declaration
  name: (identifier) @function)
(method_definition
  name: [(property_identifier) (private_property_identifier)] @method)

(pair
  key: (property_identifier) @method
  value: (function))
(pair
  key: (property_identifier) @method
  value: (arrow_function))

(assignment_expression
  left: (member_expression
    property: (property_identifier) @method)
  right: (arrow_function))
(assignment_expression
  left: (member_expression
    property: (property_identifier) @method)
  right: (function))

(variable_declarator
  name: (identifier) @function
  value: (arrow_function))
(variable_declarator
  name: (identifier) @function
  value: (function))

(assignment_expression
  left: (identifier) @function
  right: (arrow_function))
(assignment_expression
  left: (identifier) @function
  right: (function))

; Function and method calls
;--------------------------

(call_expression
  function: (identifier) @function)

(call_expression
  function: (member_expression
    property: [(property_identifier) (private_property_identifier)] @method))

; Variables
;----------
(namespace_import
  (identifier) @namespace)

; Literals
;---------

(this) @variable.builtin
(super) @variable.builtin

(true) @boolean
(false) @boolean
(null) @constant.builtin
[
(comment)
(hash_bang_line)
] @comment
(string) @string
(regex) @punctuation.delimiter
(regex_pattern) @string.regex
(template_string) @string
(escape_sequence) @string.escape
(number) @number

; Punctuation
;------------

"..." @punctuation.special

";" @punctuation.delimiter
"." @punctuation.delimiter
"," @punctuation.delimiter
"?." @punctuation.delimiter

(pair ":" @punctuation.delimiter)

[
  "--"
  "-"
  "-="
  "&&"
  "+"
  "++"
  "+="
  "&="
  "/="
  "**="
  "<<="
  "<"
  "<="
  "<<"
  "="
  "=="
  "==="
  "!="
  "!=="
  "=>"
  ">"
  ">="
  ">>"
  "||"
  "%"
  "%="
  "*"
  "**"
  ">>>"
  "&"
  "|"
  "^"
  "??"
  "*="
  ">>="
  ">>>="
  "^="
  "|="
  "&&="
  "||="
  "??="
] @operator

(binary_expression "/" @operator)
(ternary_expression ["?" ":"] @conditional)
(unary_expression ["!" "~" "-" "+" "delete" "void" "typeof"]  @operator)

[
  "("
  ")"
  "["
  "]"
  "{"
  "}"
] @punctuation.bracket

((template_substitution ["${" "}"] @punctuation.special) @none)

; Keywords
;----------

[
"if"
"else"
"switch"
"case"
"default"
] @conditional

[
"import"
"from"
"as"
] @include

[
"for"
"of"
"do"
"while"
"continue"
] @repeat

[
"async"
"await"
"break"
"class"
"const"
"debugger"
"export"
"extends"
"get"
"in"
"instanceof"
"let"
"set"
"static"
"switch"
"target"
"typeof"
"var"
"void"
"with"
] @keyword

[
"return"
"yield"
] @keyword.return

[
 "function"
] @keyword.function

[
 "new"
 "delete"
] @keyword.operator

[
 "throw"
 "try"
 "catch"
 "finally"
] @exception

(jsx_element
  open_tag: (jsx_opening_element ["<" ">"] @tag.delimiter))
(jsx_element
  close_tag: (jsx_closing_element ["<" "/" ">"] @tag.delimiter))
(jsx_self_closing_element ["/" ">" "<"] @tag.delimiter)
(jsx_fragment [">" "<" "/"] @tag.delimiter)
(jsx_attribute (property_identifier) @tag.attribute)

(jsx_opening_element
  name: (identifier) @tag)

(jsx_closing_element
  name: (identifier) @tag)

(jsx_self_closing_element
  name: (identifier) @tag)

(jsx_opening_element ((identifier) @constructor
 (#lua-match? @constructor "^[A-Z]")))

; Handle the dot operator effectively - <My.Component>
(jsx_opening_element ((nested_identifier (identifier) @tag (identifier) @constructor)))

(jsx_closing_element ((identifier) @constructor
 (#lua-match? @constructor "^[A-Z]")))

; Handle the dot operator effectively - </My.Component>
(jsx_closing_element ((nested_identifier (identifier) @tag (identifier) @constructor)))

(jsx_self_closing_element ((identifier) @constructor
 (#lua-match? @constructor "^[A-Z]")))

; Handle the dot operator effectively - <My.Component />
(jsx_self_closing_element ((nested_identifier (identifier) @tag (identifier) @constructor)))

(jsx_text) @none

;;; Parameters
(formal_parameters (identifier) @parameter)

(formal_parameters
  (rest_pattern
    (identifier) @parameter))

;; ({ a }) => null
(formal_parameters
  (object_pattern
    (shorthand_property_identifier_pattern) @parameter))

;; ({ a: b }) => null
(formal_parameters
  (object_pattern
    (pair_pattern
      value: (identifier) @parameter)))

;; ([ a ]) => null
(formal_parameters
  (array_pattern
    (identifier) @parameter))

;; a => null
(arrow_function
  parameter: (identifier) @parameter)

;; optional parameters
(formal_parameters
  (assignment_pattern
    left: (identifier) @parameter))
