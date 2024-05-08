;; This file inherits from typescript/highlights.scm
;; This file should be kept in sync with javascript/highlights-jsx.scm
(jsx_opening_element
  name: (identifier) @tag)
(jsx_closing_element
  name: (identifier) @tag)
(jsx_attribute (property_identifier) @tag.attribute)
