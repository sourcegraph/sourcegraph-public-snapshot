(identifier) @variable

;; Methods

(method_declaration name: (identifier) @method)
(local_function_statement name: (identifier) @function)

(using_directive (identifier) @variable.module)
(qualified_name (identifier) @variable.module)

;; Types

(interface_declaration name: (identifier) @type)
(class_declaration name: (identifier) @type)
(enum_declaration name: (identifier) @type)
(struct_declaration (identifier) @type)
(record_declaration (identifier) @type)
(namespace_declaration name: (identifier) @type)
(object_creation_expression type: (identifier) @type)
(method_declaration returns: (identifier) @type)
(variable_declaration type: (identifier) @type)
(property_declaration type: (identifier) @type)

(constructor_declaration name: (identifier) @type)
(destructor_declaration name: (identifier) @type)


;; Parameter

(parameter (identifier) @variable.parameter)
(parameter
 type: (identifier) @type
 name: (identifier) @variable.parameter)


[
  (implicit_type)
  (nullable_type)
  (pointer_type)
  (function_pointer_type)
  (predefined_type)
] @type.builtin

;; Class
(base_list (identifier) @type)


;; Preprocessors
(preproc_if condition: (identifier) @identifier.constant)
(preproc_elif condition: (identifier) @identifier.constant)
(preproc_arg) @string

;; Members

; The constructor_declaration queries below assume that the left-hand side of
; assignments assign to class fields.
(constructor_declaration
  body: (block (expression_statement (assignment_expression .
    left: (identifier) @identifier.attribute))))
(constructor_declaration
  body: (arrow_expression_clause (assignment_expression .
    left: (identifier) @identifier.attribute)))

(member_access_expression name: (identifier) @identifier.attribute)
(property_declaration name: (identifier) @identifier.attribute)


(invocation_expression
  (member_access_expression
    (generic_name
      (identifier) @identifier.function)))

(invocation_expression
  (member_access_expression
    name: (identifier) @identifier.function))

(invocation_expression
  function: (conditional_access_expression
             (member_binding_expression
               name: (identifier) @identifier.function)))

(invocation_expression
      (identifier) @identifier.function)

(invocation_expression
  function: (generic_name
              . (identifier) @identifier.function))

;; Enum
(enum_member_declaration (identifier) @identifier.constant)

;; Literals
[
  (real_literal)
  (integer_literal)
] @number

[
  (character_literal)
  (string_literal)
  (raw_string_literal)
  (verbatim_string_literal)
  (interpolated_string_expression)
  (interpolation_start)
  (interpolation_quote)
  (escape_sequence)
]@string

[
  (boolean_literal)
  (null_literal)
] @constant.builtin

;; Comments
(comment) @comment

;; Operator
[
  "--"
  "-"
  "-="
  "&"
  "&="
  "&&"
  "+"
  "++"
  "+="
  "<"
  "<="
  "<<"
  "<<="
  "="
  "=="
  "!"
  "!="
  "=>"
  ">"
  ">="
  ">>"
  ">>="
  ">>>"
  ">>>="
  "|"
  "|="
  "||"
  "?"
  "??"
  "??="
  "^"
  "^="
  "~"
  "*"
  "*="
  "/"
  "/="
  "%"
  "%="
  ":"
] @operator
(operator_declaration ["+" "-" "true" "false" "==" "!="] @method)

;; Tokens
[
  ";"
  "."
  ","
] @punctuation.delimiter

[
  "("
  ")"
  "["
  "]"
  "{"
  "}"
  (interpolation_brace)
] @punctuation.bracket

;; Keywords
[
  (modifier)
  "this"
  (implicit_type)
] @keyword

[
  "as"
  "await"
  "base"
  "break"
  "case"
  "catch"
  "checked"
  "class"
  "continue"
  "default"
  "delegate"
  "do"
  "else"
  "enum"
  "equals"
  "event"
  "explicit"
  "finally"
  "for"
  "foreach"
  "from"
  "get"
  "goto"
  "if"
  "implicit"
  "in"
  "init"
  "interface"
  "is"
  "join"
  "let"
  "lock"
  "namespace"
  "new"
  "notnull"
  "on"
  "operator"
  "out"
  "params"
  "record"
  "ref"
  "return"
  "select"
  "set"
  "sizeof"
  "stackalloc"
  "static"
  "struct"
  "switch"
  "this"
  "throw"
  "try"
  "typeof"
  "unchecked"
  "unmanaged"
  "using"
  "var"
  "when"
  "where"
  "while"
  "with"
  "yield"
  ] @keyword

;; event
(event_declaration (accessor_list (accessor_declaration ["add" "remove"] @keyword)))

; (initializer_expression (assignment_expression left: (identifier) @identifier.attribute))
; (attribute_argument (name_equals . (identifier) @identifier.attribute))
(field_declaration (variable_declaration (variable_declarator . (identifier) @identifier.attribute)))

;; Lambda
(lambda_expression) @variable

;; Attribute
(attribute name: (identifier) @type)

;; Sunset restricted types
"__makeref" @keyword
"__refvalue" @keyword
"__reftype" @keyword

;; Typeof
; (type_of_expression (identifier) @type)

;; Type
(generic_name (identifier) @type)
(type_parameter (identifier) @property.definition)
(type_argument_list (identifier) @type)

;; Type constraints
(type_parameter_constraints_clause (identifier) @property.definition)
(type_parameter_constraint (identifier) @type)

;; Exception
(catch_declaration (identifier) @type (identifier) @variable)
(catch_declaration (identifier) @type)
