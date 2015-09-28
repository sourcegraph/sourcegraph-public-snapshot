/**
 * @fileoverview Disallow undeclared variables in JSX
 * @author Yannick Croissant
 */

'use strict';

/**
 * Checks if a node name match the JSX tag convention.
 * @param {String} name - Name of the node to check.
 * @returns {boolean} Whether or not the node name match the JSX tag convention.
 */
var tagConvention = /^[a-z]|\-/;
function isTagName(name) {
  return tagConvention.test(name);
}

// ------------------------------------------------------------------------------
// Rule Definition
// ------------------------------------------------------------------------------

module.exports = function(context) {

  /**
   * Compare an identifier with the variables declared in the scope
   * @param {ASTNode} node - Identifier or JSXIdentifier node
   * @returns {void}
   */
  function checkIdentifierInJSX(node) {
    var scope = context.getScope();
    var variables = scope.variables;
    var i;
    var len;

    while (scope.type !== 'global') {
      scope = scope.upper;
      variables = scope.variables.concat(variables);
    }
    if (scope.childScopes.length) {
      variables = scope.childScopes[0].variables.concat(variables);
      // Temporary fix for babel-eslint
      if (scope.childScopes[0].childScopes.length) {
        variables = scope.childScopes[0].childScopes[0].variables.concat(variables);
      }
    }

    for (i = 0, len = variables.length; i < len; i++) {
      if (variables[i].name === node.name) {
        return;
      }
    }

    context.report(node, '\'' + node.name + '\' is not defined.');
  }

  return {
    JSXOpeningElement: function(node) {
      switch (node.name.type) {
        case 'JSXIdentifier':
          node = node.name;
          break;
        case 'JSXMemberExpression':
          node = node.name.object;
          break;
        case 'JSXNamespacedName':
          node = node.name.namespace;
          break;
        default:
          break;
      }
      if (isTagName(node.name)) {
        return;
      }
      checkIdentifierInJSX(node);
    }
  };

};

module.exports.schema = [];
