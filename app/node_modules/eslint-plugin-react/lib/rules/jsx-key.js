/**
 * @fileoverview Report missing `key` props in iterators/collection literals.
 * @author Ben Mosher
 */
'use strict';

// ------------------------------------------------------------------------------
// Rule Definition
// ------------------------------------------------------------------------------

module.exports = function(context) {

  function isKeyProp(decl) {
    if (decl.type === 'JSXSpreadAttribute') {
      return false;
    }
    return (decl.name.name === 'key');
  }

  return {
    JSXElement: function(node) {
      if (node.openingElement.attributes.some(isKeyProp)) {
        return; // has key prop
      }

      if (node.parent.type === 'ArrayExpression') {
        context.report(node, 'Missing "key" prop for element in array');
      }

      if (node.parent.type === 'ArrowFunctionExpression') {
        context.report(node, 'Missing "key" prop for element in iterator');
      }
    }
  };
};
