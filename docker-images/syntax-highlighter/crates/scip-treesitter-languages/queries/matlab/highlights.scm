(comment) @comment
(identifier) @identifier

"function" @keyword.function
(return_statement) @keyword.return

(string) @string
(escape_sequence) @string.escape

[
    (global_operator)
    (persistent_operator)
    "+"
    ".+"
    "-"
    ".-"
    "*"
    ".*"
    "/"
    "./"
    "\\"
    ".\\"
    "^"
    ".^"
    "|"
    "&"
    "&&"
    "||"
    "<"
    "<="
    "=="
    "~="
    ">="
    ">"
    "@"
    "?"
    "~"
] @operator

[
  "true"
  "false"
] @boolean

(number) @number

[
    "end"
    "if"
    "else"
    "for"
    "case"
    "switch"
    "otherwise"
    (continue_statement)
    (break_statement)
] @keyword
