/**
 * @fileoverview Enforce boolean attributes notation in JSX
 * @author Yannick Croissant
 */
'use strict';

// ------------------------------------------------------------------------------
// Rule Definition
// ------------------------------------------------------------------------------

module.exports = function(context) {

  var configuration = context.options[0] || 'never';

  var NEVER_MESSAGE = 'Value must be omitted for boolean attributes';
  var ALWAYS_MESSAGE = 'Value must be set for boolean attributes';

  return {
    JSXAttribute: function(node) {
      switch (configuration) {
        case 'always':
          if (node.value === null) {
            context.report(node, ALWAYS_MESSAGE);
          }
          break;
        case 'never':
          if (node.value && node.value.type === 'JSXExpressionContainer' && node.value.expression.value === true) {
            context.report(node, NEVER_MESSAGE);
          }
          break;
        default:
          break;
      }
    }
  };
};

module.exports.schema = [{
  enum: ['always', 'never']
}];
