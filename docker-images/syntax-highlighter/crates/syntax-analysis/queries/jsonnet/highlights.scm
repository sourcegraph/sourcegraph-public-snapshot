(id) @variable
(param identifier: (id) @variable.parameter)
(bind function: (id) @function)
(fieldname) @string.special
[
  "["
  "]"
  "{"
  "}"]
@punctuation.bracket

[ "error" "assert" ] @identifier.function

; keyword style
[
  "if"
  "then"
  "else"]
@conditional

[
  (local)
  (super)
  "function"
  "for"
  "in"
  "import"
  "importstr"]
@keyword

; Language basics
(comment) @comment
(number) @number
[ (true) (false) ] @boolean
[ (self) (dollar) ] @constant.builtin
(binaryop) @operator
(unaryop) @operator


; It's possible for us to give a "special" highlight to
; the triple ||| to start a string if we wanted with this query
;; ((string_start) @comment
;;  (#eq? @comment "|||"))

[
  (string_start)
  (string_end)]
@string.special

(string_content) @string

; Imports
(import) @variable.module
