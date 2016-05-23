;;; sourcegraph.el --- Minor mode for Sourcegraph -*- lexical-binding: t; -*-

;; Copyright (c) 2014, 2015 sourcegraph-mode Authors.
;; Copyright (c) 2012 The go-mode Authors. All rights reserved.

;; Author: The sourcegraph-mode Authors
;; Version: 0.1
;; Keywords: sourcegraph
;; URL: https://sourcegraph.com/sourcegraph/sourcegraph/-/tree/contrib/emacs
;; Package-Requires: ((emacs "24.3"))
;;
;; This file is not part of GNU Emacs.

;; Use of this source code is governed by a BSD-style license that can
;; be found in the LICENSE file.

;;; Commentary:

;; Emacs mode for Sourcegraph.

;; Inside a file, call `sourcegraph-describe-in-browser' with the
;; point on an identifier to open your web browser to the
;; corresponding Sourcegraph page.

;; `sourcegraph-search-site' will open Sourcegraph.com in your browser
;; for your search terms.

;; `sourcegraph-mode' sets the keybinding for
;; `sourcegraph-describe-in-browser' to "C-M-.".  To activate
;; `sourcegraph-mode' when editing code, add the following hook to
;; your Emacs config:

;; (add-hook 'prog-mode-hook 'sourcegraph-mode)

;;; Code:

(require 'url)

(defcustom sourcegraph-mode-line " src"
  "Mode line lighter for Sourcegraph."
  :group 'sourcegraph
  :type 'sexp
  :risky t)

(defvar sourcegraph-base-url "https://staging.sourcegraph.com"
  "Base URL for Sourcegraph.com.")

(defvar sourcegraph-process-buffer "*sourcegraph-process*"
  "Process buffer for sourcegraph-mode.")

(defcustom godefinfo-command "godefinfo"
  "The 'godefinfo' command."
  :type 'string
  :group 'go)


;;;###autoload
(define-minor-mode sourcegraph-mode
  "Minor mode for using Sourcegraph with Emacs.
This mode sets the keybinding for `sourcegraph-describe-in-browser' to
\"C-M-.\".
To activate `sourcegraph-mode', add the following hook to your Emacs config:
\(add-hook 'prog-mode-hook 'sourcegraph-mode)
"

  :group 'sourcegraph
  :require 'sourcegraph
  :lighter sourcegraph-mode-line
  :keymap (let ((map (make-sparse-keymap)))
            (define-key map (kbd "C-M-.") 'sourcegraph-describe-in-browser)
            map))

;;;;;;;;;; HACK: TODO(sqs): move channel name generation to server side, we can't risk making these guessable or leaked
(random t)
(defvar sourcegraph-channel)
(defun sourcegraph--get-channel ()
  "Get existing or make a new random-ish channel name for Sourcegraph."
  (prog2
	  (if (not (boundp 'sourcegraph-channel))
		  (setq sourcegraph-channel (format "%s-%06x%06x%06x%06x%06x%06x"
											(getenv "USER")
											(random (expt 16 6))
											(random (expt 16 6))
											(random (expt 16 6))
											(random (expt 16 6))
											(random (expt 16 6))
											(random (expt 16 6)))))
	  sourcegraph-channel
	))

(defun godefinfo--call (point)
  "Call godefinfo, acquiring definition position and expression description at POINT."
  (if (and (not (buffer-file-name (go--coverage-origin-buffer))))
	  (message "Cannot use godefinfo on a buffer without a file name")
    (let ((outbuf (get-buffer-create "*godefinfo*"))
          (coding-system-for-read 'utf-8)
          (coding-system-for-write 'utf-8))
      (with-current-buffer outbuf
        (erase-buffer))
      (call-process-region (point-min)
                           (point-max)
                           godefinfo-command
                           nil
                           outbuf
                           nil
                           "-i"
                           "-f"
                           (file-truename (buffer-file-name (go--coverage-origin-buffer)))
                           "-o"
                           (number-to-string (go--position-bytes point)))
      (with-current-buffer outbuf
        (buffer-substring-no-properties (point-min) (point-max))))))


(defun sourcegraph--invoke-godefinfo (point)
  (condition-case nil
	  (let ((description (cdr (butlast (godefinfo--call point) 1))))
		(if (not description)
			(file-error (message "godefinfo: no description found for expression at point"))
		  (message "%s" (mapconcat #'identity description "\n"))))
	(file-error (message "Could not run godefinfo binary"))))

(defun sourcegraph--send-channel-goto-url-action (def-action-json try-open-channel)
  "Call Channel.Send with an action (DEF-ACTION-JSON) to go to the given URL."
  (let ((url-request-method "POST")
		(url-request-extra-headers
		 '(("Content-Type" . "application/json")))
		(url-request-data (format "{\"Action\": %s, \"CheckForListeners\": true}" def-action-json))
		(api-url (format "%s/.api/channel/%s" "https://staging.sourcegraph.com" (sourcegraph--get-channel))))
	(url-retrieve api-url (lambda (status) (sourcegraph--receive-channel-send-result def-action-json status try-open-channel)) nil t)))

(defun sourcegraph--receive-channel-send-result (def-action-json status try-open-channel)
  "Receive the HTTP response STATUS from sending the action to URL. If TRY-OPEN-CHANNEL is non-nil, open a new browser window and try again."
  (if (and (equal (car status) :error) (equal (car (cdr status)) '(error http 408)))
	  (if try-open-channel
		  (progn
			(browse-url (format "%s/-/channel/%s" sourcegraph-base-url (sourcegraph--get-channel)))
			(sit-for 3)
			(sourcegraph--send-channel-goto-url-action def-action-json nil))
		(message "Could not open channel"))))

;;;###autoload
(defun sourcegraph-describe-in-browser (point)
  "Describe the expression at POINT in the web browser."
  (interactive "d")
  (condition-case nil
	  (let* ((output (godefinfo--call point))
			 (name-parts (split-string output))
			 (def-action-json (format "{\"Repo\": \"%s\", \"Package\": \"%s\", \"Def\": \"%s\"}" (car name-parts) (car name-parts) (mapconcat 'identity (cdr name-parts) "/"))))
		(sourcegraph--send-channel-goto-url-action def-action-json t))
	(file-error (message "Could not run godefinfo binary"))))

(defvar last-post-command-position 0
  "Holds the cursor position from the last run of post-command-hooks.")
(make-variable-buffer-local 'last-post-command-position)

(defun sourcegraph-describe-when-idle ()
  "Set up timers so that sourcegraph-describe-in-browser is called when the cursor moves."
  (cancel-function-timers 'sourcegraph-describe-if-moved-post-command)
  (run-at-time 0.3 nil 'sourcegraph-describe-if-moved-post-command))

(defun sourcegraph-describe-if-moved-post-command ()
  "Call sourcegraph-describe-in-browser on the cursor position if the cursor moved since the last such call."
  (if (and (not (equal (point) last-post-command-position)) (string-match "\\.go\\'" (buffer-name)))
	  (sourcegraph-describe-in-browser (point))
	(setq last-post-command-position (point))))

(add-hook 'go-mode-hook
          (lambda ()
			(add-hook 'post-command-hook 'sourcegraph-describe-when-idle)))

(provide 'sourcegraph)

;;; sourcegraph.el ends here
