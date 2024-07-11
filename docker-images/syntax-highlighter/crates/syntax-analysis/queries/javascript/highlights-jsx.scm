;; This file inherits from javascript/highlights.scm
;; This file should be kept in sync with tsx/highlights.scm
(jsx_opening_element
  "<" @tag.delimiter
  name: (identifier) @tag
  ">" @tag.delimiter)

; We need to match the tag characters individually since the version of the grammar we are using (v0.20.0) defines them that way
; when we update to the newest version we can remove this
(jsx_closing_element
  ["<" "/"] @tag.delimiter
  name: (identifier) @tag
  ">" @tag.delimiter)

(jsx_self_closing_element
  "<" @tag.delimiter
  name: (identifier) @tag
  ["/" ">"] @tag.delimiter)

(jsx_attribute (property_identifier) @tag.attribute)
