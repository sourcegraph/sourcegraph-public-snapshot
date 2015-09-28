# 1.2.3

* Fixed an issue where `-webkit-tap-highlight-color` was being incorrectly
  transformed to `transparent`. This is not supported in Safari.

# 1.2.2

* Fixed a bug where the module crashed on parsing comma separated values for
  properties such as `box-shadow`.

# 1.2.1

* Extracted each color logic into a function for better readability.

# 1.2.0

* Now uses the PostCSS `4.1` plugin API.

# 1.1.0

* Now supports optimisation of colors in gradient values.

# 1.0.0

* Initial release.
