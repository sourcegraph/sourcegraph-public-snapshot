/**
 * @fileoverview Prevent usage of setState in componentDidUpdate
 * @author Yannick Croissant
 */
'use strict';

// ------------------------------------------------------------------------------
// Rule Definition
// ------------------------------------------------------------------------------

module.exports = function(context) {

  // --------------------------------------------------------------------------
  // Public
  // --------------------------------------------------------------------------

  return {

    CallExpression: function(node) {
      var callee = node.callee;
      if (callee.type !== 'MemberExpression') {
        return;
      }
      if (callee.object.type !== 'ThisExpression' || callee.property.name !== 'setState') {
        return;
      }
      var ancestors = context.getAncestors(callee);
      for (var i = 0, j = ancestors.length; i < j; i++) {
        if (ancestors[i].type !== 'Property' || ancestors[i].key.name !== 'componentDidUpdate') {
          continue;
        }
        context.report(callee, 'Do not use setState in componentDidUpdate');
      }
    }
  };

};

module.exports.schema = [];
