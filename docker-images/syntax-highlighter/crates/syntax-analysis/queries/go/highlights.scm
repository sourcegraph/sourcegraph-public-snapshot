; Identifiers

(identifier) @variable
(field_identifier) @identifier.property
(package_identifier) @variable.module
(type_identifier) @type
((identifier) @constant (#match? @constant "^[A-Z][A-Z\\d_]+$"))
((identifier) @constant (#eq? @constant "_"))
(keyed_element . (literal_element (identifier) @identifier.attribute))

; Literals

(interpreted_string_literal) @string
(raw_string_literal) @string
(rune_literal) @string
(escape_sequence) @string.escape

(int_literal) @number
(float_literal) @float
(imaginary_literal) @number

(true) @boolean
(false) @boolean
(nil) @constant.null

(comment) @comment

; Builtin types

((type_identifier) @type.builtin
  (#match? @type.builtin
            "^(bool|byte|complex128|complex64|error|float32|float64|int|int16|int32|int64|int8|rune|string|uint|uint16|uint32|uint64|uint8|uintptr)$"))


; Function calls

(parameter_declaration (identifier) @variable.parameter)
(variadic_parameter_declaration (identifier) @variable.parameter)

(call_expression
  function: (identifier) @identifier.function)

(call_expression
  function: (selector_expression
             field: (field_identifier) @identifier.function))

; Builtin functions

((identifier) @function.builtin
  (#match? @function.builtin "^(append|cap|close|complex|copy|delete|imag|len|make|new|panic|print|println|real|recover)$"))

; Function definitions

(method_elem
 name: (field_identifier) @identifier.function)
(function_declaration
 name: (identifier) @identifier.function)
(method_declaration
 name: (field_identifier) @identifier.function)
(type_parameter_declaration
 name: (identifier) @identifier.parameter)

; Constants

(const_spec
 name: (identifier) @constant)

; Operators
[
 "--"
 "-"
 "-="
 ":="
 "!"
 "!="
 "..."
 "*"
 "*"
 "*="
 "/"
 "/="
 "&"
 "&&"
 "&="
 "%"
 "%="
 "^"
 "^="
 "+"
 "++"
 "+="
 "<-"
 "<"
 "<<"
 "<<="
 "<="
 "="
 "=="
 ">"
 ">="
 ">>"
 ">>="
 "|"
 "|="
 "||"] @operator

; Keywords

[
 "break"
 "chan"
 "const"
 "continue"
 "default"
 "defer"
 "go"
 "goto"
 "interface"
 "map"
 "range"
 "select"
 "struct"
 "type"
 "var"
 "fallthrough"] @keyword

"func" @keyword.function
"return" @keyword.return

"for" @keyword.repeat

[
 "import"
 "package"] @include

[
 "else"
 "case"
 "switch"
 "if"] @conditional
