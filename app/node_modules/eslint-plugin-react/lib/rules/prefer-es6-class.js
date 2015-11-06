/**
 * @fileoverview Prefer es6 class instead of createClass for React Component
 * @author Dan Hamilton
 */
'use strict';

var componentUtil = require('../util/component');

// ------------------------------------------------------------------------------
// Rule Definition
// ------------------------------------------------------------------------------

module.exports = function(context) {

  return {
    ObjectExpression: function(node) {
      if (componentUtil.isComponentDefinition(context, node)) {
        context.report(node, 'Component should use es6 class instead of createClass');
      }
    }
  };
};

module.exports.schema = [];
