(pod_directive) @keyword
(head_directive) @keyword
(over_directive) @keyword
(item_directive) @keyword
(back_directive) @keyword
(encoding_directive) @keyword
(cut_directive) @keyword

;; TODO: I'm not sure why this messes up the injected highlights...
; (head_paragraph (head_directive) (content) @keyword)

(over_paragraph (content) @string)
(item_paragraph (content) @string)
(encoding_paragraph (content) @string)

(plain_paragraph) @comment

;; Do not add verbatim_paragraph, since that is added in the injections
; (verbatim_paragraph (content) @comment)

(interior_sequence) @variable

;; TODO: We could probably do a few nicer things here...
;; we haven't really added anything for styling items
; (interior_sequence
;   (sequence_letter) @letter
;   (#match? @letter "B")
;   (content) @text.strong)
; (interior_sequence
;   (sequence_letter) @letter
;   (#match? @letter "C")
;   (content) @text.literal)
; (interior_sequence
;   (sequence_letter) @letter
;   (#match? @letter "F")
;   (content) @text.quote) ; not really "quote" but there isn't a better one
; (interior_sequence
;   (sequence_letter) @letter
;   (#match? @letter "I")
;   (content) @text.emphasis)
; (interior_sequence
;   (sequence_letter) @letter
;   (#match? @letter "L")
;   (content) @text.uri)
