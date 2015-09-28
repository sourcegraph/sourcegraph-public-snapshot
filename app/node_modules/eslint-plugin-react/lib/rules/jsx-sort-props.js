/**
 * @fileoverview Enforce props alphabetical sorting
 * @author Ilya Volodin, Yannick Croissant
 */
'use strict';

// ------------------------------------------------------------------------------
// Rule Definition
// ------------------------------------------------------------------------------

module.exports = function(context) {

  var configuration = context.options[0] || {};
  var ignoreCase = configuration.ignoreCase || false;

  return {
    JSXOpeningElement: function(node) {
      node.attributes.reduce(function(memo, decl, idx, attrs) {
        if (decl.type === 'JSXSpreadAttribute') {
          return attrs[idx + 1];
        }

        var lastPropName = memo.name.name;
        var currenPropName = decl.name.name;

        if (ignoreCase) {
          lastPropName = lastPropName.toLowerCase();
          currenPropName = currenPropName.toLowerCase();
        }

        if (currenPropName < lastPropName) {
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
    ignoreCase: {
      type: 'boolean'
    }
  },
  additionalProperties: false
}];
