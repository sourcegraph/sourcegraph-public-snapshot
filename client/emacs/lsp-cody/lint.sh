#!/usr/bin/env sh

FILES="lsp-cody.el"

echo "--- Linting Lisp ---"
# TODO: Add elsa, blocked by https://github.com/emacs-elsa/Elsa/issues/219
eask lint elint "${FILES}" \
  && eask lint elisp-lint "${FILES}" \
  && eask lint indent "${FILES}" \
  && eask lint declare \
  && eask lint regexps "${FILES}"
echo

echo "--- Linting Package ---"
eask lint package "${FILES}" \
  && eask lint checkdoc --strict "${FILES}" \
  && eask lint keywords
