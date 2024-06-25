 ;; This file inherits from typescript/highlights.scm
;; This file should be kept in sync with javascript/highlights-jsx.scm
(jsx_opening_element
  "<" @tag.delimiter
  name: (identifier) @tag
  ">" @tag.delimiter)

; We need to match the tag characters individually since the version of the grammar we are using (v0.20.2) defines them that way
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
