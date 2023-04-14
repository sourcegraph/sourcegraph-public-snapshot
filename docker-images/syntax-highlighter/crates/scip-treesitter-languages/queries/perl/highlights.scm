[
  ; Our keywords
  "return"
  ; Existing
  "use" "no"
  "package"
  "sub"
  "if" "elsif" "else" "unless"
  "while" "until"
  "for" "foreach"
  "do"
  "my" "our" "local"
  "require"
  "last" "next" "redo" "goto"
  "undef"]
@keyword

;; @keyword.phaser
[ "BEGIN" "INIT" "CHECK" "UNITCHECK" "END" ] @keyword

[
  "or" "and"
  "eq" "ne" "cmp" "lt" "le" "ge" "gt"
  "isa"]
@operator

(regex_literal
  (regex_start) @keyword
  (regex_modifier)? @keyword) @string

"=" @operator
"=~" @operator
(assignment_expression operator: (_) @operator)
(binary_expression operator: (_) @operator)
(equality_expression operator: (_) @operator)
(relational_expression operator: (_) @operator)
"=>" @operator

(comment) @comment

(number) @number
(version) @number

(string_literal) @string
(interpolated_string_literal) @string
(quoted_word_list) @string
(command_string) @string
[(heredoc_token) (command_heredoc_token)] @string.special
(heredoc_content) @string
(heredoc_end) @string.special
[(escape_sequence) (escaped_delimiter)] @string.special

(autoquoted_bareword) @string.special

(scalar) @variable
(scalar_deref_expression ["->" "$" "*"] @variable)
(array) @variable
(array_deref_expression ["->" "@" "*"] @variable)
(hash) @variable
(hash_deref_expression ["->" "%" "*"] @variable)
(array_element_expression [array:(_) "->" "[" "]"] @variable)
(hash_element_expression [hash:(_) "->" "{" "}"] @variable)

(hash_element_expression key: (bareword) @string.special)

(use_statement (package) @type)
(package_statement (package) @type)
(require_expression (bareword) @type)

(subroutine_declaration_statement name: (_) @function)

(function_call_expression (function) @function)
(method_call_expression (method) @function.method)
(method_call_expression invocant: (bareword) @type)

(func0op_call_expression function: _ @function.builtin)
(func1op_call_expression function: _ @function.builtin)

;; TODO: This may not be the best way to do it.
(bareword) @keyword

(data_section) @comment

; (attrlist (attribute) @decorator)
; (goto_expression (bareword) @label)
; (loopex_expression (bareword) @label)
; (statement_label label: (bareword) @label)
