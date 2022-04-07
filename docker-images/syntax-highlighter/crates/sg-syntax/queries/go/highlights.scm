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

; Function calls

(parameter_declaration (identifier) @variable.parameter)
(variadic_parameter_declaration (identifier) @variable.parameter)

; (call_expression
;  function: (selector_expression
;    operand: (identifier)
;    field: (field_identifier) @identifier.function)
;  arguments: (_))
;
; (call_expression
;  function: (identifier) @identifier)

(call_expression
  function: (identifier) @identifer.function)

(call_expression
  function: (selector_expression
    field: (field_identifier) @identifier.function))

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


;; Builtin types

((type_identifier) @type.builtin
  (#any-of? @type.builtin
            "bool"
            "byte"
            "complex128"
            "complex64"
            "error"
            "float32"
            "float64"
            "int"
            "int16"
            "int32"
            "int64"
            "int8"
            "rune"
            "string"
            "uint"
            "uint16"
            "uint32"
            "uint64"
            "uint8"
            "uintptr"))


;; Builtin functions

((identifier) @function.builtin
  (#any-of? @function.builtin
            "append"
            "cap"
            "close"
            "complex"
            "copy"
            "delete"
            "imag"
            "len"
            "make"
            "new"
            "panic"
            "print"
            "println"
            "real"
            "recover"))


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


(
 (const_spec
  name: (identifier) @_id
  value: (expression_list (raw_string_literal) @keyword))

 (#match? @_id ".*Query$")
)

;;
; Identifiers

(package_identifier) @variable.module
(type_identifier) @type
(identifier) @variable

; (field_identifier) @property

(
 (identifier) @constant
  (#eq? @constant "_"))

; ((identifier) @constant
;  (#vim-match? @constant "^[A-Z][A-Z\\d_]+$"))

