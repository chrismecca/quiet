;;; quiet-themes.el --- Quiet themes for long-duration work -*- lexical-binding: t; -*-
;; @(#)quiet / elisp / ports/emacs/quiet-themes.el / 2026-10-19

;; Copyright (C) 2026 Christopher Mecca
;; SPDX-License-Identifier: MIT

;; Author: Christopher Mecca
;; Maintainer: Christopher Mecca
;; URL: https://github.com/chrismecca/quiet
;; Version: 0
;; Package-Requires: ((emacs "29.1"))
;; Keywords: faces, theme, accessibility

;;; Commentary:

;; Quiet is a restrained color system for long-duration work.
;;
;; This package provides two adaptations of the Quiet semantic system:
;;
;;   quiet-day
;;   quiet-night
;;
;; The theme files are generated from Quiet's canonical palette data.
;; This library exists as the package entry point and makes the installed
;; theme files visible through `custom-theme-load-path'.
;;
;; Load a theme with:
;;
;;   M-x load-theme RET quiet-day RET
;;
;; or:
;;
;;   M-x load-theme RET quiet-night RET
;;
;; When one Quiet theme is active, switch between the Day and Night
;; variants with:
;;
;;   M-x theme-choose-variant

;;; Code:

;;;###autoload
(when load-file-name
  (add-to-list
   'custom-theme-load-path
   (file-name-as-directory
    (file-name-directory load-file-name))))

(provide 'quiet-themes)

;;; quiet-themes.el ends here

