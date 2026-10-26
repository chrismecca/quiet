# quiet

A restrained color system for long-duration work.

quiet provides two themes, **quiet-day** and **quiet-night**. Both are
generated from quiet's canonical palettes and checked against its contrast
contract.

## Syntax

Ordinary source is Ink, comments are Muted, and language structure, such as
keywords, types, and operators, is Mineral. Moss, Ochre, and Rust are kept for
state: additions, warnings, and errors. Strings are not success, and numbers
are not warnings.

## Highlights that keep syntax colors

Search matches, symbol highlights, diff lines, and merge regions use quiet's
Wash roles: tints that keep every syntax color at 4.5:1 or better. Changed
words in the diff editor are outlined rather than filled.

## Integrated terminal

The terminal uses quiet's generic terminal palette, shared with quiet's ports
for kitty, Ghostty, iTerm2, WezTerm, and Terminal.app.

## More

quiet also has ports for Vim, Neovim, Emacs, JetBrains IDEs, Helix, Zed,
Xcode, browsers, and terminals. See the
[quiet repository](https://github.com/chrismecca/quiet).
