;; This file inherits from javascript/highlights.scm
;; This file should be kept in sync with tsx/highlights.scm
(jsx_opening_element
  name: (identifier) @tag)
(jsx_closing_element
  name: (identifier) @tag)
(jsx_attribute (property_identifier) @tag.attribute)
