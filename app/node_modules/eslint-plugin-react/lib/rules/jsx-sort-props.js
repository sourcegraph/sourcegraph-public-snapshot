/**
 * @fileoverview Enforce props alphabetical sorting
 * @author Ilya Volodin, Yannick Croissant
 */
'use strict';

// ------------------------------------------------------------------------------
// Rule Definition
// ------------------------------------------------------------------------------

function isCallbackPropName(propName) {
  return /^on[A-Z]/.test(propName);
}

module.exports = function(context) {

  var configuration = context.options[0] || {};
  var ignoreCase = configuration.ignoreCase || false;
  var callbacksLast = configuration.callbacksLast || false;

  return {
    JSXOpeningElement: function(node) {
      node.attributes.reduce(function(memo, decl, idx, attrs) {
        if (decl.type === 'JSXSpreadAttribute') {
          return attrs[idx + 1];
        }

        var previousPropName = memo.name.name;
        var currentPropName = decl.name.name;
        var previousIsCallback = isCallbackPropName(previousPropName);
        var currentIsCallback = isCallbackPropName(currentPropName);

        if (ignoreCase) {
          previousPropName = previousPropName.toLowerCase();
          currentPropName = currentPropName.toLowerCase();
        }

        if (callbacksLast) {
          if (!previousIsCallback && currentIsCallback) {
            // Entering the callback prop section
            return decl;
          }
          if (previousIsCallback && !currentIsCallback) {
            // Encountered a non-callback prop after a callback prop
            context.report(memo, 'Callbacks must be listed after all other props');
            return memo;
          }
        }

        if (currentPropName < previousPropName) {
          context.report(decl, 'Props should be sorted alphabetically');
          return memo;
        }

        return decl;
      }, node.attributes[0]);
    }
  };
};

module.exports.schema = [{
  type: 'object',
  properties: {
    // Whether callbacks (prefixed with "on") should be listed at the very end,
    // after all other props.
    callbacksLast: {
      type: 'boolean'
    },
    ignoreCase: {
      type: 'boolean'
    }
  },
  additionalProperties: false
}];
