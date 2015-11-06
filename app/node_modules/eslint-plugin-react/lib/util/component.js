/**
 * @fileoverview Utility functions for React components detection
 * @author Yannick Croissant
 */
'use strict';

var util = require('util');

var DEFAULT_COMPONENT_NAME = 'eslintReactComponent';

/**
 * Detect if the node is a component definition
 * @param {Object} context The current rule context.
 * @param {ASTNode} node The AST node being checked.
 * @returns {Boolean} True the node is a component definition, false if not.
 */
function isComponentDefinition(context, node) {
  switch (node.type) {
    case 'ObjectExpression':
      if (node.parent && node.parent.callee && context.getSource(node.parent.callee) === 'React.createClass') {
        return true;
      }
      break;
    case 'ClassDeclaration':
      var superClass = node.superClass && context.getSource(node.superClass);
      if (superClass === 'Component' || superClass === 'React.Component') {
        return true;
      }
      break;
    default:
      break;
  }
  return false;
}

/**
 * Check if we are in a stateless function component
 * @param {Object} context The current rule context.
 * @param {ASTNode} node The AST node being checked.
 * @returns {Boolean} True if we are in a stateless function component, false if not.
 */
function isStatelessFunctionComponent(context, node) {
  if (node.type !== 'ReturnStatement') {
    return false;
  }

  var scope = context.getScope();
  while (scope) {
    if (scope.type === 'class') {
      return false;
    }
    scope = scope.upper;
  }

  var returnsJSX =
    node.argument &&
    node.argument.type === 'JSXElement'
  ;
  var returnsReactCreateElement =
    node.argument &&
    node.argument.callee &&
    node.argument.callee.property &&
    node.argument.callee.property.name === 'createElement'
  ;

  return Boolean(returnsJSX || returnsReactCreateElement);
}

/**
 * Get the identifiers of a React component ASTNode
 * @param {ASTNode} node The React component ASTNode being checked.
 * @returns {Object} The component identifiers.
 */
function getIdentifiers(node) {
  var name = [];
  var loopNode = node;
  var namePart = [];
  while (loopNode) {
    namePart = (loopNode.id && loopNode.id.name) || (loopNode.key && loopNode.key.name);
    if (namePart) {
      name.unshift(namePart);
    }
    loopNode = loopNode.parent;
  }
  name = name.join('.') || DEFAULT_COMPONENT_NAME;
  var id = name + ':' + node.loc.start.line + ':' + node.loc.start.column;

  return {
    id: id,
    name: name
  };
}

/**
 * Get the React component ASTNode from any child ASTNode
 * @param {Object} context The current rule context.
 * @param {ASTNode} node The AST node being checked.
 * @returns {ASTNode} The ASTNode of the React component.
 */
function getNode(context, node, list) {
  var ancestors = context.getAncestors().reverse();

  ancestors.unshift(node);

  for (var i = 0, j = ancestors.length; i < j; i++) {
    if (isComponentDefinition(context, ancestors[i])) {
      return ancestors[i];
    }
    // Node is already in the component list
    var identifiers = getIdentifiers(ancestors[i]);
    if (list && list[identifiers.id] && list[identifiers.id].isComponent) {
      return ancestors[i];
    }
  }

  if (isStatelessFunctionComponent(context, node)) {
    var scope = context.getScope();
    while (scope.upper && scope.type !== 'function') {
      scope = scope.upper;
    }
    return scope.block;
  }

  return null;
}

/**
 * Store a React component list
 * @constructor
 */
function List() {
  this._list = {};
  this._length = 0;
}

/**
 * Find a component in the list by his node or one of his child node
 * @param {Object} context The current rule context.
 * @param {ASTNode} node The node to find.
 * @returns {Object|null} The component if it is found, null if not.
 */
List.prototype.getByNode = function(context, node) {
  var componentNode = getNode(context, node);
  if (!componentNode) {
    return null;
  }
  var identifiers = getIdentifiers(componentNode);

  return this._list[identifiers.id] && this._list[identifiers.id].isComponent ? this._list[identifiers.id] : null;
};

/**
 * Find a component in the list by his name
 * @param {String} name Name of the component to find.
 * @returns {Object|null} The component if it is found, null if not.
 */
List.prototype.getByName = function(name) {
  for (var component in this._list) {
    if (
      this._list.hasOwnProperty(component) &&
      this._list[component].name === name &&
      this._list[component].isComponent
    ) {
      return this._list[component];
    }
  }
  return null;
};

/**
 * Return the component list
 * @returns {Object} The component list.
 */
List.prototype.getList = function() {
  var components = {};
  for (var component in this._list) {
    if (!this._list.hasOwnProperty(component) || this._list[component].isComponent !== true) {
      continue;
    }
    components[component] = this._list[component];
  }
  return components;
};

/**
 * Add/update a component in the list
 * @param {Object} context The current rule context.
 * @param {ASTNode} node The node to add.
 * @param {Object} customProperties Additional properties to add to the component.
 * @returns {Object} The added component.
 */
List.prototype.set = function(context, node, customProperties) {
  var componentNode = getNode(context, node, this._list);
  var isComponent = false;
  if (componentNode) {
    node = componentNode;
    isComponent = true;
  }

  var identifiers = getIdentifiers(node);

  var component = util._extend({
    name: identifiers.name,
    node: node,
    isComponent: isComponent
  }, customProperties || {});

  if (!this._list[identifiers.id]) {
    this._length++;
  }

  this._list[identifiers.id] = util._extend(this._list[identifiers.id] || {}, component);

  return component;
};

/**
 * Remove a component from the list
 * @param {Object} context The current rule context.
 * @param {ASTNode} node The node to remove.
 */
List.prototype.remove = function(context, node) {
  var componentNode = getNode(context, node);
  if (!componentNode) {
    return null;
  }
  var identifiers = getIdentifiers(componentNode);

  if (!this._list[identifiers.id]) {
    return null;
  }

  delete this._list[identifiers.id];
  this._length--;

  return null;
};

/**
 * Return the component list length
 * @returns {Number} The component list length.
 */
List.prototype.count = function() {
  return this._length;
};

module.exports = {
  DEFAULT_COMPONENT_NAME: DEFAULT_COMPONENT_NAME,
  getNode: getNode,
  isComponentDefinition: isComponentDefinition,
  getIdentifiers: getIdentifiers,
  List: List
};
