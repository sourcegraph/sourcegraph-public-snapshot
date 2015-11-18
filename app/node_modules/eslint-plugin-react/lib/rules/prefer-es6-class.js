/**
 * @fileoverview Prefer es6 class instead of createClass for React Component
 * @author Dan Hamilton
 */
'use strict';

var Components = require('../util/Components');

// ------------------------------------------------------------------------------
// Rule Definition
// ------------------------------------------------------------------------------

module.exports = Components.detect(function(context) {

  return {
    ObjectExpression: function(node) {
      if (context.react.isES5Component(node)) {
        context.report(node, 'Component should use es6 class instead of createClass');
      }
    }
  };
});

module.exports.schema = [];
