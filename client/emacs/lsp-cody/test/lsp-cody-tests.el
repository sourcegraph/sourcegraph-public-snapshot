;;; lsp-cody-tests.el -*- lexical-binding: t; -*-

;; This file is NOT part of GNU Emacs.

;;; Commentary:

;; This file is part of LSP-CODY

;;; Code:

(require 'buttercup)
(require 'lsp-cody)

(describe "buttercup tests"
    (it "can run"
        (expect t :to-equal t)))
