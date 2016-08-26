var closest = require('./closest');
var matches = require('./matches');

/**
 * Delegates the handling of events for an element matching a selector to an
 * ancestor of the matching element.
 * @param {Element} ancestor The ancestor element to add the listener to.
 * @param {string} eventType The event type to listen to.
 * @param {string} selector A CSS selector to match against child elements.
 * @param {Function} callback A function to run any time the event happens.
 * @param {Object} opts A configuration options object. The available options:
 *     - useCapture<boolean>: If true, bind to the event capture phase.
 *     - deep<boolean>: If true, delegate into shadow trees.
 * @return {Object} The delegate object. It contains a destroy method.
 */
 module.exports = function delegate(
    ancestor, eventType, selector, callback, opts) {

  opts = opts || {};

  // Defines the event listener.
  var listener = function(event) {

    // If opts.deep is true and the event originated from inside a Shadow DOM,
    // check the deep nodes.
    if (opts.deep && typeof event.deepPath == 'function') {
      var path = event.deepPath();
      for (var i = 0, node; node = path[i]; i++) {
        if (node.nodeType == 1 && matches(node, selector)) {
          delegateTarget = node;
        }
      }
    }
    // Otherwise check the parents.
    else {
      var delegateTarget = closest(event.target, selector, true);
    }

    if (delegateTarget) {
      callback.call(delegateTarget, event, delegateTarget);
    }
  };

  ancestor.addEventListener(eventType, listener, opts.useCapture);

  return {
    destroy: function() {
      ancestor.removeEventListener(eventType, listener, opts.useCapture);
    }
  };
};
