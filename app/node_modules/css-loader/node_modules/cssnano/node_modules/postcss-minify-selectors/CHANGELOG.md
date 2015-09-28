# 1.4.2

* Bump dependencies.
* Fixes for PostCSS plugin guidelines.

# 1.4.1

* Fixes incorrect deduplication of pseudo selector rules.

# 1.4.0

* Update to postcss-selector-parser to greatly improve parsing logic.

# 1.3.1

* Fixes a crash when nothing was passed to `node-balanced`.

# 1.3.0

* Now uses the PostCSS `4.1` plugin API.

# 1.2.1

* Passes original test case in issue 1.

# 1.2.0

* Does not touch quoted values in attribute selectors.
* No longer will mangle values such as `2100%` in keyframes.

# 1.1.0

* Now minifies `from` to `0%` and `100%` to `to` in keyframe declarations.

# 1.0.0

* Initial release.
