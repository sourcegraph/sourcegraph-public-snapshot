; ;; Forked from tree-sitter-go
; ;; Copyright (c) 2014 Max Brunsfeld (The MIT License)

;; TODO: We can re-enable when we update SQL a bit more.
; (
;  (const_spec
;   name: (identifier) @_id
;   value: (expression_list (raw_string_literal) @none))
; 
;  (#match? @_id ".*Query$")
; )

;; TODO: It would be cool to explore trying to guess whether
;; something is a local or not to use as modules, but let's
;; leave that for smarter parsers than us.
;; I will leave this as an example for later though.
; (
;  (call_expression
;   function: (selector_expression
;    operand: (identifier) @variable.module (#is-not? @variable.module local)
;    field: (field_identifier))))


;; Builtin types

((type_identifier) @type.builtin
 (#match? @type.builtin 
  "(bool|byte|complex128|complex64|error|float32|float64|int|int16|int32|int64|int8|rune|string|uint|uint16|uint32|uint64|uint8|uintptr)"
    ))

;; Builtin functions

((identifier) @function.builtin
  (#match? @function.builtin
   "(append|cap|close|complex|copy|delete|imag|len|make|new|panic|print|println|real|recover)"))

; Function calls

(parameter_declaration (identifier) @variable.parameter)
(variadic_parameter_declaration (identifier) @variable.parameter)

(call_expression
  function: (identifier) @identifer.function)

(call_expression
  function: (selector_expression
    field: (field_identifier) @function))

; Function definitions

(function_declaration
 name: (identifier) @function)

(method_declaration
 name: (field_identifier) @method)

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
 "||"
] @operator

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
 "fallthrough"
] @keyword

"func" @keyword.function
"return" @keyword.return

"for" @keyword.repeat

[
 "import"
 "package"
] @include

[
 "else"
 "case"
 "switch"
 "if"
] @conditional




; Delimiters

; TODO: Olaf, you can decide if you want this one or not :)
; "." @punctuation.delimiter

"," @punctuation.delimiter
":" @punctuation.delimiter
";" @punctuation.delimiter

"(" @punctuation.bracket
")" @punctuation.bracket
"{" @punctuation.bracket
"}" @punctuation.bracket
"[" @punctuation.bracket
"]" @punctuation.bracket


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

(ERROR) @error

;;
; Identifiers

((identifier) @constant (#eq? @constant "_"))
((identifier) @constant (#match? @constant "^[A-Z][A-Z\\d_]+$"))

(package_identifier) @variable.module
(type_identifier) @type
(field_identifier) @property
(identifier) @variable
