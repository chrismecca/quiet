;;; quiet-themes.el --- Restrained color themes for long-duration work -*- lexical-binding: t; -*-
;; @(#)quiet / elisp / ports/emacs/quiet-themes.el / 2026-10-26

;; Copyright (C) 2026 Chris Mecca
;; SPDX-License-Identifier: MIT

;; Author: Chris Mecca <cm@chrismecca.com>
;; Maintainer: Chris Mecca <cm@chrismecca.com>
;; URL: https://github.com/chrismecca/quiet
;; Version: 0.5.0
;; Package-Requires: ((emacs "29.1"))
;; Keywords: faces, theme, accessibility

;;; Commentary:

;; quiet is a restrained color system for long-duration work.
;;
;; This package provides two adaptations of the quiet semantic system:
;;
;;   quiet-day
;;   quiet-night
;;
;; The theme files are generated from quiet's canonical palette data.
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
;; When one quiet theme is active, switch between the Day and Night
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

