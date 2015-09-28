# Changelog

### 0.1.0 (Jul 25, 2014)

- Initial release

### 0.1.1 (Jul 26, 2014)

- Fixing dragging not stopping on mouseup in some cases

### 0.2.0 (Sep 10, 2014)

- Adding support for snapping to a grid
- Adding support for specifying start position
- Ensure event handlers are destroyed on unmount
- Adding browserify support
- Adding bower support

### 0.2.1 (Sep 10, 2014)

- Exporting as ReactDraggable

### 0.3.0 (Oct 21, 2014)

- Adding support for touch devices

### 0.4.0 (Jan 03, 2015)

- Improving accuracy of snap to grid
- Updating to React 0.12
- Adding dragging className
- Adding reactify support for browserify
- Fixing issue with server side rendering

### 0.4.1 (Apr 30, 2015)

- Remove react/addons dependency (now depending on `react` directly).
- Add MIT License file.
- Fix an issue where browser may be detected as touch-enabled but touch event isn't thrown.

### 0.4.2 (Apr 30, 2015)

- Add `"browser"` config to package.json for browserify imports (fix #45).
- Remove unnecessary `emptyFunction` and `React.addons.classSet` imports.

### 0.4.3 (Apr 30, 2015)

- Fix React.addons error caused by faulty test.

### 0.5.0 (May 2, 2015)

- Remove browserify browser config, reactify, and jsx pragma. Fixes #38
- Use React.cloneElement instead of addons cloneWithProps (requires React 0.13)
- Move to CSS transforms. Simplifies implementation and fixes #48, #34, #31.
- Fixup linting and space/tab errors. Fixes #46.

### 0.6.0 (May 2, 2015)

- Breaking change: Cancel dragging when onDrag or onStart handlers return an explicit `false`.
- Fix sluggish movement when `grid` option was active.
- Example updates.
- Move `user-select:none` hack to document.body for better highlight prevention.
- Add `bounds` option to restrict dragging within parent or within coordinates.

### 0.7.0 (May 7, 2015)

- Breaking change: `bounds` with coordinates was confusing because it was using the item's width/height,
  which was not intuitive. When providing coordinates, `bounds` now simply restricts movement in each
  direction by that many pixels.

### 0.7.1 (May 7, 2015)

- The `start` param is back. Pass `{x: Number, y: Number}` to kickoff the CSS transform. Useful in certain
  cases for simpler callback math (so you don't have to know its existing relative position and add it to
  the dragged position). Fixes #52.

### 0.7.2 (May 8, 2015)

- Added `moveOnStartChange` property. See README.

### 0.7.3 (May 13, 2015)

- Removed a `moveOnStartChange` optimization that was causing problems when attempting to move a `<Draggable>` back
  to its initial position. See https://github.com/STRML/react-grid-layout/issues/56

### 0.7.4 (May 18, 2015)

- Fix a bug where a quick drag out of bounds to `0,0` would cause the element to remain in an inaccurate position,
  because the translation was removed from the CSS. See #55.

### 0.8.0 (May 19, 2015)

- Touch/mouse events rework. Fixes #51, #37, and #43, as well as IE11 support.
- Moved mousemove/mouseup and touch event handlers to document from window. Fixes IE9/10 support.
  IE8 is still not supported, as it is not supported by React.

### 0.8.1 (June 3, 2015)

- Add `resetState()` instance method for use by parents. See README ("State Problems?").
