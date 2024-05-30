; Based on https://github.com/krn-robin/tree-sitter-magik/blob/main/queries/highlights.scm

; Methods
(method
  exemplarname: (identifier) @type)
(method
  name: (identifier) @function.method)

(procedure (label) @function.method)


; Literals
[
  (number)
] @number

[
  (string_literal)
] @string


[
  (true)
  (false)
] @boolean

[
  (maybe)
  (unset)
] @constant.builtin

[
 (self)
 (super)
 (clone)
] @variable.builtin

[
 (symbol)
 (character_literal)
] @constant

(documentation) @comment
(comment) @comment

(package (identifier) @variable.module)


; Expression

[
    "<<"
    ">>"
    ">"
    ">="
    "<"
    "<="
    "="
    "~="
    "<>"
    "+"
    "-"
    "/"
    "*"
    "**"
    "<<"
    "^<<"
    "_and<<"
    "_andif<<"
    "_or<<"
    "_orif<<"
    "_xor<<"
    "_xorif"
    "**<<"
    "**^<<"
    "*<<"
    "*^<<"
    "/<<"
    "/^<<"
    "_mod<<"
    "_div<<"
    "-<<"
    "-^<<"
    "+<<"
    "+^<<"
] @operator


(relational_operator
    operator: _ @operator)

(logical_operator
    operator: _ @operator)

(arithmetic_operator
    operator: _ @operator)

(unary_operator
    operator: _ @operator)

(invoke
  receiver: (variable) @function.call)

(call
  receiver: (variable) @variable )
(call
  operator: "." @operator)
(call
  message: (identifier) @function.call)

; Keywords

[
  "_iter"
  "_while"
  "_over"
  "_for"
  "_loop"
  "_endloop"
  "_over"
  "_try"
  "_endtry"
  "_throw"
  "_catch"
  "_primitive"
  "_finally"
  "_default"
  "_with"
  "_when"
  "_method"
  "_endmethod"
  "_class"
  "_loopbody"
  "_gather"
  "_continue"

  "_allresults"
  "_dynamic"

  "_handling"
  "_leave"
  "_primitive"

  "_block"
  "_endblock"

  "_protect"
  "_protection"
  "_endprotect"

  "_if"
  "_then"
  "_elif"
  "_else"
  "_endif"

  "_thisthread"

  "_return"

  "_lock"

  "_abstract"
  "_private"
  "_constant"
  "_local"
  "_global"
  "_proc"
  "_endproc"
] @keyword

[
 "_package"
] @include


[
 "_pragma"
] @function.builtin


(argument) @identifier.parameter

[
 (variable)
 (dynamic_variable)
 (global_variable)
 (identifier)
 (slot_accessor)
 (label)
] @variable
