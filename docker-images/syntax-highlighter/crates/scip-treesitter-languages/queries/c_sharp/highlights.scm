(using_directive (identifier) @variable.module)
(qualified_name (identifier) @variable.module)

;; Preprocessors
(if_directive (identifier) @identifier.constant) @keyword
(endif_directive) @keyword
(elif_directive (identifier) @identifier.constant) @keyword
(else_directive) @keyword
(error_directive "error" @keyword)
(warning_directive "warning" @keyword)
(preproc_message) @string

;; Methods
(method_declaration name: (identifier) @method)

(invocation_expression
  (member_access_expression
    (generic_name
      (identifier) @method)))

(invocation_expression
  (member_access_expression

    name: (identifier) @method))

(invocation_expression
  function: (conditional_access_expression
             (member_binding_expression
               name: (identifier) @method)))

(invocation_expression
      (identifier) @method)

(invocation_expression
  function: (generic_name
              . (identifier) @method))

;; Types
(variable_declaration type: (identifier) @type)
(interface_declaration name: (identifier) @type)
(class_declaration name: (identifier) @type)
(enum_declaration name: (identifier) @type)
(struct_declaration (identifier) @type)
(record_declaration (identifier) @type)
(namespace_declaration name: (identifier) @type)
(object_creation_expression type: (identifier) @type)
(method_declaration type: (identifier) @type)
(property_declaration type: (identifier) @type)
(parameter type: (identifier) @type name: (identifier) @variable.parameter)

(constructor_declaration name: (identifier) @type)
(destructor_declaration name: (identifier) @type)

[
  (implicit_type)
  (nullable_type)
  (pointer_type)
  (function_pointer_type)
  (predefined_type)]
@type.builtin

;; Enum
(enum_member_declaration (identifier) @identifier.constant)

;; Literals
[
  (real_literal)
  (integer_literal)]
@number

[
  (character_literal)
  (string_literal)
  (verbatim_string_literal)
  (interpolated_string_text)
  (interpolated_verbatim_string_text)
  "\""
  "$\""
  "@$\""
  "$@\""] @string
(interpolation ["{" "}"] @string.escape)

[
  (boolean_literal)
  (null_literal)
  (void_keyword)] @constant.builtin

;; Comments
(comment) @comment

;; Operator
(operator_declaration ["+" "-" "true" "false" "==" "!="] @method)

;; Tokens
[
  ";"
  "."
  ","] @punctuation.delimiter

[
  "--"
  "-"
  "-="
  "&"
  "&&"
  "+"
  "++"
  "+="
  "<"
  "<<"
  "="
  "=="
  "!"
  "!="
  "=>"
  ">"
  ">>"
  "|"
  "||"
  "?"
  "??"
  "^"
  "~"
  "*"
  "/"
  "%"
  ":"] @operator

[
  "("
  ")"
  "["
  "]"
  "{"
  "}"] @punctuation.bracket

;; Keywords
(modifier) @keyword
(this_expression) @keyword
(escape_sequence) @keyword

[
  "as"
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
  "event"
  "explicit"
  "finally"
  "for"
  "foreach"
  "goto"
  "if"
  "implicit"
  "interface"
  "is"
  "lock"
  "namespace"
  "operator"
  "params"
  "return"
  "sizeof"
  "stackalloc"
  "struct"
  "switch"
  "throw"
  "try"
  "typeof"
  "unchecked"
  "using"
  "while"
  "new"
  "await"
  "in"
  "yield"
  "get"
  "set"
  "when"
  "out"
  "ref"
  "from"
  "join"
  "on"
  "equals"
  "var"
  "where"
  "select"
  "record"
  "init"
  "with"
  "this"
  "unmanaged"
  "notnull"
  "let"] @keyword


;; Linq
(from_clause (identifier) @variable)
(group_clause)
(order_by_clause)
(select_clause (identifier) @variable)
(query_continuation (identifier) @variable) @keyword

;; Record
(with_expression
  (with_initializer_expression
    (simple_assignment_expression
      (identifier) @variable)))

;; event
(event_declaration (accessor_list (accessor_declaration ["add" "remove"] @keyword)))

;; Class
(base_list (identifier) @type)

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
(initializer_expression (assignment_expression left: (identifier) @identifier.attribute))
(attribute_argument (name_equals . (identifier) @identifier.attribute))
(field_declaration (variable_declaration (variable_declarator . (identifier) @identifier.attribute)))

;; Lambda
(lambda_expression) @variable

;; Attribute
(attribute name: (identifier) @type)

;; Parameter
(parameter type: (identifier) @type name: (identifier) @variable.parameter)
(parameter (identifier) @variable.parameter)
(parameter_modifier) @keyword

;; Sunset restricted types
(make_ref_expression "__makeref" @keyword)
(ref_value_expression "__refvalue" @keyword)
(ref_type_expression "__reftype" @keyword)

;; Typeof
(type_of_expression (identifier) @type)

;; Return
(return_statement (identifier) @variable)
(yield_statement (identifier) @variable)

;; Type
(generic_name (identifier) @type)
(type_parameter (identifier) @property.definition)
(type_argument_list (identifier) @type)

;; Type constraints
(type_parameter_constraints_clause (identifier) @property.definition)
(type_constraint (identifier) @type)

;; Exception
(catch_declaration (identifier) @type (identifier) @variable)
(catch_declaration (identifier) @type)

;; Switch
(switch_statement (identifier) @variable)
(switch_expression (identifier) @variable)

;; Lock statement
(lock_statement (identifier) @variable)


(identifier) @variable
