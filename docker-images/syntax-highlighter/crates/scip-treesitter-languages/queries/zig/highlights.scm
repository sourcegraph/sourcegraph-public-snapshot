[
  (container_doc_comment)
  (doc_comment)
  (line_comment)
] @comment

[
  variable: (IDENTIFIER)
  variable_type_function: (IDENTIFIER)
  field_member: (IDENTIFIER)
  field_access: (IDENTIFIER)
] @identifier

[
  (CompareOp)
  (BitwiseOp)
  (BitShiftOp)
  (AdditionOp)
  (AssignOp)
  (MultiplyOp)
  (PrefixOp)
  "*"
  "**"
  "->"
  ".?"
  ".*"
  "?"
] @operator

(INTEGER) @number
(FLOAT) @float

field_constant: (IDENTIFIER) @constant

; Special names (keywords and co.)

"return" @keyword.return
"fn" @keyword.function

[
  "addrspace"
  "align"
  "allowzero"
  "and"
  "anyframe"
  "anytype"
  "asm"
  "async"
  "await"
  "break"
  "catch"
  "comptime"
  "const"
  "continue"
  "defer"
  "while"
  "for"
  "enum"
  "errdefer"
  "error"
  "export"
  "extern"
  "inline"
  "linksection"
  "noalias"
  "nosuspend"
  "suspend"
  "or"
  "orelse"
  "packed"
  "pub"
  "resume"
  "struct"
  "test"
  "threadlocal"
  "try"
  "union"
  "unreachable"
  "usingnamespace"
  "var"
  "volatile"]
@keyword

[
  "true"
  "false"
] @boolean

[
  "null"
  "undefined"
] @constant.builtin

[
  "if"
  "else"
  "switch"
] @conditional

[
  "anytype"
  (BuildinTypeExpr)
] @type.builtin

; Strings

[
  (LINESTRING)
  (STRINGLITERALSINGLE)
] @string

(CHAR_LITERAL) @character
(EscapeSequence) @string.escape
(FormatSequence) @string.special

; Functions and builtins

(BUILTINIDENTIFIER) @function.builtin

((BUILTINIDENTIFIER) @include
  (#any-of? @include "@import" "@cImport"))

parameter: (IDENTIFIER) @variable.parameter

[
  function_call: (IDENTIFIER)
  function: (IDENTIFIER)
] @identifier.function
