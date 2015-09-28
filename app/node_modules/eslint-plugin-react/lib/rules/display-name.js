/**
 * @fileoverview Prevent missing displayName in a React component definition
 * @author Yannick Croissant
 */
'use strict';

var componentUtil = require('../util/component');
var ComponentList = componentUtil.List;

// ------------------------------------------------------------------------------
// Rule Definition
// ------------------------------------------------------------------------------

module.exports = function(context) {

  var config = context.options[0] || {};
  var acceptTranspilerName = config.acceptTranspilerName || false;

  var componentList = new ComponentList();

  var MISSING_MESSAGE = 'Component definition is missing display name';
  var MISSING_MESSAGE_NAMED_COMP = '{{component}} component definition is missing display name';

  /**
   * Checks if the component must be validated
   * @param {Object} component The component to process
   * @returns {Boolean} True if the component must be validated, false if not.
   */
  function mustBeValidated(component) {
    return (
      component &&
      component.isReactComponent &&
      !component.hasDisplayName
    );
  }

  /**
   * Checks if we are declaring a display name
   * @param {ASTNode} node The AST node being checked.
   * @returns {Boolean} True if we are declaring a display name, false if not.
   */
  function isDisplayNameDeclaration(node) {

    // Special case for class properties
    // (babel-eslint does not expose property name so we have to rely on tokens)
    if (node.type === 'ClassProperty') {
      var tokens = context.getFirstTokens(node, 2);
      if (
        tokens[0].value === 'displayName' ||
        (tokens[1] && tokens[1].value === 'displayName')
      ) {
        return true;
      }
      return false;
    }

    return Boolean(
      node &&
      node.name === 'displayName'
    );
  }

  /**
   * Mark a prop type as declared
   * @param {ASTNode} node The AST node being checked.
   */
  function markDisplayNameAsDeclared(node) {
    componentList.set(context, node, {
      hasDisplayName: true
    });
  }

  /**
   * Reports missing display name for a given component
   * @param {Object} component The component to process
   */
  function reportMissingDisplayName(component) {
    context.report(
      component.node,
      component.name === componentUtil.DEFAULT_COMPONENT_NAME ? MISSING_MESSAGE : MISSING_MESSAGE_NAMED_COMP, {
        component: component.name
      }
    );
  }

  /**
   * Checks if the component have a name set by the transpiler
   * @param {ASTNode} node The AST node being checked.
   * @returns {Boolean} True ifcomponent have a name, false if not.
   */
  function hasTranspilerName(node) {
    var namedAssignment = (
      node.type === 'ObjectExpression' &&
      node.parent &&
      node.parent.parent &&
      node.parent.parent.type === 'AssignmentExpression' && (
        !node.parent.parent.left.object ||
        node.parent.parent.left.object.name !== 'module' ||
        node.parent.parent.left.property.name !== 'exports'
      )
    );
    var namedDeclaration = (
        node.type === 'ObjectExpression' &&
        node.parent &&
        node.parent.parent &&
        node.parent.parent.type === 'VariableDeclarator'
    );
    var namedClass = (
      node.type === 'ClassDeclaration' &&
      node.id && node.id.name
    );

    if (namedAssignment || namedDeclaration || namedClass) {
      return true;
    }
    return false;
  }

  // --------------------------------------------------------------------------
  // Public
  // --------------------------------------------------------------------------

  return {

    ClassProperty: function(node) {
      if (!isDisplayNameDeclaration(node)) {
        return;
      }

      markDisplayNameAsDeclared(node);
    },

    MemberExpression: function(node) {
      if (!isDisplayNameDeclaration(node.property)) {
        return;
      }
      var component = componentList.getByName(node.object.name);
      if (!component) {
        return;
      }
      markDisplayNameAsDeclared(component.node);
    },

    MethodDefinition: function(node) {
      if (!isDisplayNameDeclaration(node.key)) {
        return;
      }
      markDisplayNameAsDeclared(node);
    },

    ClassDeclaration: function(node) {
      if (!acceptTranspilerName || !hasTranspilerName(node)) {
        return;
      }
      markDisplayNameAsDeclared(node);
    },

    ObjectExpression: function(node) {
      // Search for the displayName declaration
      node.properties.forEach(function(property) {
        if (!property.key) {
          return;
        }

        if (!isDisplayNameDeclaration(property.key)) {
          return;
        }
        markDisplayNameAsDeclared(node);
      });
      // Has transpiler name
      if (acceptTranspilerName && hasTranspilerName(node)) {
        markDisplayNameAsDeclared(node);
      }

      if (componentUtil.isComponentDefinition(node)) {
        componentList.set(context, node, {
          isReactComponent: true
        });
      }
    },

    'Program:exit': function() {
      var list = componentList.getList();
      // Report missing display name for all classes
      for (var component in list) {
        if (!list.hasOwnProperty(component) || !mustBeValidated(list[component])) {
          continue;
        }
        reportMissingDisplayName(list[component]);
      }
    },

    ReturnStatement: function(node) {
      if (!componentUtil.isReactComponent(context, node)) {
        return;
      }
      componentList.set(context, node, {
        isReactComponent: true
      });
    }
  };

};

module.exports.schema = [{
  type: 'object',
  properties: {
    acceptTranspilerName: {
      type: 'boolean'
    }
  },
  additionalProperties: false
}];
