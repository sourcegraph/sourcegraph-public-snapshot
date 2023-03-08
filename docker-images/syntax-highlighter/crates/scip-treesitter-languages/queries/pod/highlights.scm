[(pod_directive)
 (head_directive)
 (over_directive)
 (item_directive)
 (back_directive)
 (encoding_directive)
 (cut_directive)] @keyword

(head_paragraph (content) @string)

(over_paragraph (content) @string)
(item_paragraph (content) @string)
(encoding_paragraph (content) @string)

(verbatim_paragraph (content) @comment)

(interior_sequence) @variable

(interior_sequence
  (sequence_letter) @letter
  (#match? @letter "B")
  (content) @text.strong)
(interior_sequence
  (sequence_letter) @letter
  (#match? @letter "C")
  (content) @text.literal)
(interior_sequence
  (sequence_letter) @letter
  (#match? @letter "F")
  (content) @text.quote) ; not really "quote" but there isn't a better one
(interior_sequence
  (sequence_letter) @letter
  (#match? @letter "I")
  (content) @text.emphasis)
(interior_sequence
  (sequence_letter) @letter
  (#match? @letter "L")
  (content) @text.uri)
