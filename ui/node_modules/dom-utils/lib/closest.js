var matches = require('./matches');
var parents = require('./parents');

/**
 * Gets the closest parent element that matches the passed selector.
 * @param {Element} element The element whose parents to check.
 * @param {string} selector The CSS selector to match against.
 * @param {boolean} shouldCheckSelf True if the selector should test against
 *     the passed element itself.
 * @return {?Element} The matching element or undefined.
 */
module.exports = function closest(element, selector, shouldCheckSelf) {
  if (!(element && element.nodeType == 1 && selector)) return;

  var parentElements =
      (shouldCheckSelf ? [element] : []).concat(parents(element));

  for (var i = 0, parent; parent = parentElements[i]; i++) {
    if (matches(parent, selector)) return parent;
  }
};
