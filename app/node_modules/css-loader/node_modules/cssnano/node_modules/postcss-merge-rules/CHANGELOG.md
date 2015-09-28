# 1.3.1

* Fixes a behaviour in which rule adjacent forward nodes were not being type
  checked before they were merged.
* Compatibility fixes for the PostCSS plugin guidelines.

# 1.3.0

* Better support for merging properties without the existance of a shorthand
  override.
* Can now 'merge forward' adjacent rules as well as the previous 'merge behind'
  behaviour, leading to better compression.

# 1.2.2

* Fixed an issue where the plugin crashed if node.parent was undefined.

# 1.2.1

* Fixed a bug where media queries were being merged when their parameters were
  different.

# 1.2.0

* Now uses the PostCSS `4.1` plugin API.

# 1.1.1

* Bugfix of last release, now difference is calculated in both directions.

# 1.1.0

* Less eager moving of properties, to avoid cases where moving a longhand
  property would allow a shorthand property to override it.

# 1.0.0

* Initial release.
