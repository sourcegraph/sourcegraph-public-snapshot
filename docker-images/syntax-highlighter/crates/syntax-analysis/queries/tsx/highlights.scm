;; This file inherits from typescript/highlights.scm
;; This file should be kept in sync with javascript/highlights-jsx.scm
(jsx_attribute (property_identifier) @tag.attribute)
(jsx_opening_element (identifier) @tag (#match? @tag "^[a-z][^.]*$"))
(jsx_closing_element (identifier) @tag (#match? @tag "^[a-z][^.]*$"))
(jsx_self_closing_element (identifier) @tag (#match? @tag "^[a-z][^.]*$"))

(jsx_attribute (property_identifier) @tag.attribute)
(jsx_opening_element (["<" ">"]) @tag.delimiter)
(jsx_closing_element (["</" ">"]) @tag.delimiter)
(jsx_self_closing_element (["<" "/>"])  @tag.delimiter)
