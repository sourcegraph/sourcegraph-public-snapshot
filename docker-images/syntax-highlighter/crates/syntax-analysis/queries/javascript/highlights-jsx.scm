;; This file inherits from javascript/highlights.scm
;; This file should be kept in sync with tsx/highlights.scm
(jsx_opening_element
  "<" @tag.delimiter
  name: (identifier) @tag
  ">" @tag.delimiter)

(jsx_closing_element
  ["<" "/"] @tag.delimiter
  name: (identifier) @tag
  ">" @tag.delimiter)

(jsx_self_closing_element
  "<" @tag.delimiter
  name: (identifier) @tag
  ["/" ">"] @tag.delimiter)

(jsx_attribute (property_identifier) @tag.attribute)
