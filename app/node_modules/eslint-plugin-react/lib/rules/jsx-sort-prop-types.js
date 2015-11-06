/**
 * @fileoverview Enforce propTypes declarations alphabetical sorting
 */
'use strict';

// ------------------------------------------------------------------------------
// Rule Definition
// ------------------------------------------------------------------------------

module.exports = function(context) {

  var configuration = context.options[0] || {};
  var callbacksLast = configuration.callbacksLast || false;
  var ignoreCase = configuration.ignoreCase || false;

  /**
   * Checks if node is `propTypes` declaration
   * @param {ASTNode} node The AST node being checked.
   * @returns {Boolean} True if node is `propTypes` declaration, false if not.
   */
  function isPropTypesDeclaration(node) {

    // Special case for class properties
    // (babel-eslint does not expose property name so we have to rely on tokens)
    if (node.type === 'ClassProperty') {
      var tokens = context.getFirstTokens(node, 2);
      if (tokens[0].value === 'propTypes' || tokens[1].value === 'propTypes') {
        return true;
      }
      return false;
    }

    return Boolean(
      node &&
      node.name === 'propTypes'
    );
  }

  function getKey(node) {
    return node.key.type === 'Identifier' ? node.key.name : node.key.value;
  }

  function isCallbackPropName(propName) {
    return /^on[A-Z]/.test(propName);
  }

  /**
   * Checks if propTypes declarations are sorted
   * @param {Array} declarations The array of AST nodes being checked.
   * @returns {void}
   */
  function checkSorted(declarations) {
    declarations.reduce(function(prev, curr) {
      var prevPropName = getKey(prev);
      var currentPropName = getKey(curr);
      var previousIsCallback = isCallbackPropName(prevPropName);
      var currentIsCallback = isCallbackPropName(currentPropName);

      if (ignoreCase) {
        prevPropName = prevPropName.toLowerCase();
        currentPropName = currentPropName.toLowerCase();
      }

      if (callbacksLast) {
        if (!previousIsCallback && currentIsCallback) {
          // Entering the callback prop section
          return curr;
        }
        if (previousIsCallback && !currentIsCallback) {
          // Encountered a non-callback prop after a callback prop
          context.report(prev, 'Callback prop types must be listed after all other prop types');
          return prev;
        }
      }

      if (currentPropName < prevPropName) {
        context.report(curr, 'Prop types declarations should be sorted alphabetically');
        return prev;
      }

      return curr;
    }, declarations[0]);
  }

  return {
    ClassProperty: function(node) {
      if (isPropTypesDeclaration(node) && node.value && node.value.type === 'ObjectExpression') {
        checkSorted(node.value.properties);
      }
    },

    MemberExpression: function(node) {
      if (isPropTypesDeclaration(node.property)) {
        var right = node.parent.right;
        if (right && right.type === 'ObjectExpression') {
          checkSorted(right.properties);
        }
      }
    },

    ObjectExpression: function(node) {
      node.properties.forEach(function(property) {
        if (!property.key) {
          return;
        }

        if (!isPropTypesDeclaration(property.key)) {
          return;
        }
        if (property.value.type === 'ObjectExpression') {
          checkSorted(property.value.properties);
        }
      });
    }

  };
};

module.exports.schema = [{
  type: 'object',
  properties: {
    callbacksLast: {
      type: 'boolean'
    },
    ignoreCase: {
      type: 'boolean'
    }
  },
  additionalProperties: false
}];
