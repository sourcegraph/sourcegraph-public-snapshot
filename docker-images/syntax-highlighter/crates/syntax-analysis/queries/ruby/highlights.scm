; Identifiers

(global_variable) @identifier ;; Should SCIP SyntaxKind support global variables?
(identifier) @identifier
(constant) @identifier


; Keywords

[
  "alias"
  "and"
  "begin"
  "break"
  "case"
  "class"
  "def"
  "do"
  "else"
  "elsif"
  "end"
  "ensure"
  "for"
  "if"
  "in"
  "module"
  "next"
  "or"
  "rescue"
  "retry"
  "return"
  "then"
  "unless"
  "until"
  "when"
  "while"
  "yield"]
@keyword

((identifier) @keyword
 (#match? @keyword "^(private|protected|public)$"))

; Function calls

((identifier) @keyword
 (#eq? @keyword "require"))

((identifier) @keyword
 (#eq? @keyword "require_relative"))

"defined?" @identifier.builtin

(call
  method: [(identifier) (constant)] @identifier.function)

(call
  method: [(identifier) @type.builtin (block)]
  (#eq? @type.builtin "sig"))

; Function definitions

(alias (identifier) @identifier.function)
(setter (identifier) @identifier.function)
(method name: [(identifier) (constant)] @identifier.function)
(singleton_method name: [(identifier) (constant)] @identifier.function)

[
  (class_variable)
  (instance_variable)]
@identifier.attribute

((identifier) @constant.builtin
 (#match? @constant.builtin "^__(FILE|LINE|ENCODING)__$"))

(file) @constant.builtin
(line) @constant.builtin
(encoding) @constant.builtin

(hash_splat_nil
  "**" @operator)
@constant.builtin

(self) @identifier.builtin
(super) @identifier.builtin

(block_parameter (identifier) @identifier.parameter)
(block_parameters (identifier) @identifier.parameter)
(destructured_parameter (identifier) @identifier.parameter)
(hash_splat_parameter (identifier) @identifier.parameter)
(lambda_parameters (identifier) @identifier.parameter)
(method_parameters (identifier) @identifier.parameter)
(splat_parameter (identifier) @identifier.parameter)

(keyword_parameter name: (identifier) @identifier.parameter)
(optional_parameter name: (identifier) @identifier.parameter)

; Literals

[
  (string_content)
  (bare_string)
  (subshell)
  ; (heredoc_body)
  (heredoc_content)]
  ; (heredoc_beginning)
@string
(string "\"" @string)
; (string "'" @string)
; ((string (_) @string .))
; "''" @string

[
  (simple_symbol)
  (delimited_symbol)
  (hash_key_symbol)
  (bare_symbol)]
@character ; TODO: What else?

(escape_sequence) @string.escape
(regex) @string ; TODO: Missing regexp literal

[
  (integer)
  (float)]
@number

[
  (true)
  (false)]
@boolean

(nil) @constant.null

(interpolation ("#{") @string.escape)
(interpolation ("}") @string.escape)

(comment) @comment

; Operators

[
 "="
 "=>"
 "->"]
@operator
