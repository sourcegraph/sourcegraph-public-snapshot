/**
 * @fileoverview Utility functions for React components detection
 * @author Yannick Croissant
 */
'use strict';

var util = require('util');

var DEFAULT_COMPONENT_NAME = 'eslintReactComponent';

/**
 * Detect if we are in a React Component
 * A React component is defined has an object/class with a property "render"
 * that return a JSXElement or null/false
 * @param {Object} context The current rule context.
 * @param {ASTNode} node The AST node being checked.
 * @returns {Boolean} True if we are in a React Component, false if not.
 */
function isReactComponent(context, node) {
  if (node.type !== 'ReturnStatement') {
    throw new Error('React Component detection must be done from a ReturnStatement ASTNode');
  }

  var scope = context.getScope();
  while (scope.upper && scope.type !== 'function') {
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
  var isComponentRender =
    (returnsJSX || returnsReactCreateElement) &&
    scope.block.parent.key && scope.block.parent.key.name === 'render'
  ;
  var isEmptyComponentRender =
    node.argument &&
    node.argument.type === 'Literal' && (node.argument.value === null || node.argument.value === false) &&
    scope.block.parent.key && scope.block.parent.key.name === 'render'
  ;

  return Boolean(isEmptyComponentRender || isComponentRender);
}

/**
 * Detect if the node is a component definition
 * @param {ASTNode} node The AST node being checked.
 * @returns {Boolean} True the node is a component definition, false if not.
 */
function isComponentDefinition(node) {
  var isES6Component = node.type === 'ClassDeclaration';
  var isES5Component = Boolean(
    node.type === 'ObjectExpression' &&
    node.parent &&
    node.parent.callee &&
    node.parent.callee.object &&
    node.parent.callee.property &&
    node.parent.callee.object.name === 'React' &&
    node.parent.callee.property.name === 'createClass'
  );
  return isES5Component || isES6Component;
}

/**
 * Get the React component ASTNode from any child ASTNode
 * @param {Object} context The current rule context.
 * @param {ASTNode} node The AST node being checked.
 * @returns {ASTNode} The ASTNode of the React component.
 */
function getNode(context, node) {
  var componentNode = null;
  var ancestors = context.getAncestors().reverse();

  ancestors.unshift(node);

  for (var i = 0, j = ancestors.length; i < j; i++) {
    if (isComponentDefinition(ancestors[i])) {
      componentNode = ancestors[i];
      break;
    }
  }

  return componentNode;
}

/**
 * Get the identifiers of a React component ASTNode
 * @param {ASTNode} node The React component ASTNode being checked.
 * @returns {Object} The component identifiers.
 */
function getIdentifiers(node) {
  var name = node.id && node.id.name || DEFAULT_COMPONENT_NAME;
  var id = name + ':' + node.loc.start.line + ':' + node.loc.start.column;

  return {
    id: id,
    name: name
  };
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

  return this._list[identifiers.id] || null;
};

/**
 * Find a component in the list by his name
 * @param {String} name Name of the component to find.
 * @returns {Object|null} The component if it is found, null if not.
 */
List.prototype.getByName = function(name) {
  for (var component in this._list) {
    if (this._list.hasOwnProperty(component) && this._list[component].name === name) {
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
  return this._list;
};

/**
 * Add/update a component in the list
 * @param {Object} context The current rule context.
 * @param {ASTNode} node The node to add.
 * @param {Object} customProperties Additional properties to add to the component.
 * @returns {Object} The added component.
 */
List.prototype.set = function(context, node, customProperties) {
  var componentNode = getNode(context, node);
  if (!componentNode) {
    return null;
  }
  var identifiers = getIdentifiers(componentNode);

  var component = util._extend({
    name: identifiers.name,
    node: componentNode
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
  isReactComponent: isReactComponent,
  getNode: getNode,
  isComponentDefinition: isComponentDefinition,
  getIdentifiers: getIdentifiers,
  List: List
};
