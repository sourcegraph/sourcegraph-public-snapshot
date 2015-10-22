/**
 * @fileoverview Prevent missing props validation in a React component definition
 * @author Yannick Croissant
 */
'use strict';

// As for exceptions for props.children or props.className (and alike) look at
// https://github.com/yannickcr/eslint-plugin-react/issues/7

var componentUtil = require('../util/component');
var ComponentList = componentUtil.List;

// ------------------------------------------------------------------------------
// Rule Definition
// ------------------------------------------------------------------------------

module.exports = function(context) {

  var configuration = context.options[0] || {};
  var ignored = configuration.ignore || [];
  var customValidators = configuration.customValidators || [];

  var componentList = new ComponentList();

  var MISSING_MESSAGE = '\'{{name}}\' is missing in props validation';
  var MISSING_MESSAGE_NAMED_COMP = '\'{{name}}\' is missing in props validation for {{component}}';

  /**
   * Checks if we are using a prop
   * @param {ASTNode} node The AST node being checked.
   * @returns {Boolean} True if we are using a prop, false if not.
   */
  function isPropTypesUsage(node) {
    var isClassUsage = (
      componentUtil.getNode(context, node) &&
      node.object.type === 'ThisExpression' && node.property.name === 'props'
    );
    var isStatelessFunctionUsage = node.object.name === 'props';
    return isClassUsage || isStatelessFunctionUsage;
  }

  /**
   * Checks if we are declaring a prop
   * @param {ASTNode} node The AST node being checked.
   * @returns {Boolean} True if we are declaring a prop, false if not.
   */
  function isPropTypesDeclaration(node) {

    // Special case for class properties
    // (babel-eslint does not expose property name so we have to rely on tokens)
    if (node && node.type === 'ClassProperty') {
      var tokens = context.getFirstTokens(node, 2);
      if (
        tokens[0].value === 'propTypes' ||
        (tokens[1] && tokens[1].value === 'propTypes')
      ) {
        return true;
      }
      return false;
    }

    return Boolean(
      node &&
      node.name === 'propTypes'
    );

  }

  /**
   * Checks if the prop is ignored
   * @param {String} name Name of the prop to check.
   * @returns {Boolean} True if the prop is ignored, false if not.
   */
  function isIgnored(name) {
    return ignored.indexOf(name) !== -1;
  }

  /**
   * Checks if prop should be validated by plugin-react-proptypes
   * @param {String} validator Name of validator to check.
   * @returns {Boolean} True if validator should be checked by custom validator.
   */
  function hasCustomValidator(validator) {
    return customValidators.indexOf(validator) !== -1;
  }

  /**
   * Checks if the component must be validated
   * @param {Object} component The component to process
   * @returns {Boolean} True if the component must be validated, false if not.
   */
  function mustBeValidated(component) {
    return Boolean(
      component &&
      component.usedPropTypes &&
      !component.ignorePropsValidation
    );
  }

  /**
   * Internal: Checks if the prop is declared
   * @param {Object} declaredPropTypes Description of propTypes declared in the current component
   * @param {String[]} keyList Dot separated name of the prop to check.
   * @returns {Boolean} True if the prop is declared, false if not.
   */
  function _isDeclaredInComponent(declaredPropTypes, keyList) {
    for (var i = 0, j = keyList.length; i < j; i++) {
      var key = keyList[i];
      var propType = (
        // Check if this key is declared
        declaredPropTypes[key] ||
        // If not, check if this type accepts any key
        declaredPropTypes.__ANY_KEY__
      );

      if (!propType) {
        // If it's a computed property, we can't make any further analysis, but is valid
        return key === '__COMPUTED_PROP__';
      }
      if (propType === true) {
        return true;
      }
      // Consider every children as declared
      if (propType.children === true) {
        return true;
      }
      if (propType.acceptedProperties) {
        return key in propType.acceptedProperties;
      }
      if (propType.type === 'union') {
        // If we fall in this case, we know there is at least one complex type in the union
        if (i + 1 >= j) {
          // this is the last key, accept everything
          return true;
        }
        // non trivial, check all of them
        var unionTypes = propType.children;
        var unionPropType = {};
        for (var k = 0, z = unionTypes.length; k < z; k++) {
          unionPropType[key] = unionTypes[k];
          var isValid = _isDeclaredInComponent(
            unionPropType,
            keyList.slice(i)
          );
          if (isValid) {
            return true;
          }
        }

        // every possible union were invalid
        return false;
      }
      declaredPropTypes = propType.children;
    }
    return true;
  }

  /**
   * Checks if the prop is declared
   * @param {Object} component The component to process
   * @param {String[]} names List of names of the prop to check.
   * @returns {Boolean} True if the prop is declared, false if not.
   */
  function isDeclaredInComponent(component, names) {
    return _isDeclaredInComponent(
      component.declaredPropTypes || {},
      names
    );
  }

  /**
   * Checks if the prop has spread operator.
   * @param {ASTNode} node The AST node being marked.
   * @returns {Boolean} True if the prop has spread operator, false if not.
   */
  function hasSpreadOperator(node) {
    var tokens = context.getTokens(node);
    return tokens.length && tokens[0].value === '...';
  }

  /**
   * Retrieve the name of a key node
   * @param {ASTNode} node The AST node with the key.
   * @return {string} the name of the key
   */
  function getKeyValue(node) {
    var key = node.key;
    return key.type === 'Identifier' ? key.name : key.value;
  }

  /**
   * Iterates through a properties node, like a customized forEach.
   * @param {Object[]} properties Array of properties to iterate.
   * @param {Function} fn Function to call on each property, receives property key
      and property value. (key, value) => void
   */
  function iterateProperties(properties, fn) {
    if (properties.length && typeof fn === 'function') {
      for (var i = 0, j = properties.length; i < j; i++) {
        var node = properties[i];
        var key = getKeyValue(node);

        var value = node.value;
        fn(key, value);
      }
    }
  }

  /**
   * Creates the representation of the React propTypes for the component.
   * The representation is used to verify nested used properties.
   * @param {ASTNode} value Node of the React.PropTypes for the desired propery
   * @return {Object|Boolean} The representation of the declaration, true means
   *    the property is declared without the need for further analysis.
   */
  function buildReactDeclarationTypes(value) {
    if (
      value &&
      value.callee &&
      value.callee.object &&
      hasCustomValidator(value.callee.object.name)
    ) {
      return true;
    }

    if (
      value.type === 'MemberExpression' &&
      value.property &&
      value.property.name &&
      value.property.name === 'isRequired'
    ) {
      value = value.object;
    }

    // Verify React.PropTypes that are functions
    if (
      value.type === 'CallExpression' &&
      value.callee &&
      value.callee.property &&
      value.callee.property.name &&
      value.arguments &&
      value.arguments.length > 0
    ) {
      var callName = value.callee.property.name;
      var argument = value.arguments[0];
      switch (callName) {
        case 'shape':
          if (argument.type !== 'ObjectExpression') {
            // Invalid proptype or cannot analyse statically
            return true;
          }
          var shapeTypeDefinition = {
            type: 'shape',
            children: {}
          };
          iterateProperties(argument.properties, function(childKey, childValue) {
            shapeTypeDefinition.children[childKey] = buildReactDeclarationTypes(childValue);
          });
          return shapeTypeDefinition;
        case 'arrayOf':
        case 'objectOf':
          return {
            type: 'object',
            children: {
              __ANY_KEY__: buildReactDeclarationTypes(argument)
            }
          };
        case 'oneOfType':
          if (
            !argument.elements ||
            !argument.elements.length
          ) {
            // Invalid proptype or cannot analyse statically
            return true;
          }
          var unionTypeDefinition = {
            type: 'union',
            children: []
          };
          for (var i = 0, j = argument.elements.length; i < j; i++) {
            var type = buildReactDeclarationTypes(argument.elements[i]);
            // keep only complex type
            if (type !== true) {
              if (type.children === true) {
                // every child is accepted for one type, abort type analysis
                unionTypeDefinition.children = true;
                return unionTypeDefinition;
              }
            }

            unionTypeDefinition.children.push(type);
          }
          if (unionTypeDefinition.length === 0) {
            // no complex type found, simply accept everything
            return true;
          }
          return unionTypeDefinition;
        case 'instanceOf':
          return {
            type: 'instance',
            // Accept all children because we can't know what type they are
            children: true
          };
        case 'oneOf':
        default:
          return true;
      }
    }
    // Unknown property or accepts everything (any, object, ...)
    return true;
  }

  /**
   * Check if we are in a class constructor
   * @return {boolean} true if we are in a class constructor, false if not
   */
  function inConstructor() {
    var scope = context.getScope();
    while (scope) {
      if (scope.block && scope.block.parent && scope.block.parent.kind === 'constructor') {
        return true;
      }
      scope = scope.upper;
    }
    return false;
  }

  /**
   * Retrieve the name of a property node
   * @param {ASTNode} node The AST node with the property.
   * @return {string} the name of the property or undefined if not found
   */
  function getPropertyName(node) {
    var directProp = /^props\./.test(context.getSource(node));
    if (directProp && componentUtil.getNode(context, node) && !inConstructor(node)) {
      return void 0;
    }
    if (!directProp) {
      node = node.parent;
    }
    var property = node.property;
    if (property) {
      switch (property.type) {
        case 'Identifier':
          if (node.computed) {
            return '__COMPUTED_PROP__';
          }
          return property.name;
        case 'MemberExpression':
          return void 0;
        case 'Literal':
          // Accept computed properties that are literal strings
          if (typeof property.value === 'string') {
            return property.value;
          }
          // falls through
        default:
          if (node.computed) {
            return '__COMPUTED_PROP__';
          }
          break;
      }
    }
  }

  /**
   * Mark a prop type as used
   * @param {ASTNode} node The AST node being marked.
   */
  function markPropTypesAsUsed(node, parentNames) {
    parentNames = parentNames || [];
    var type;
    var name;
    var allNames;
    var properties;
    switch (node.type) {
      case 'MemberExpression':
        name = getPropertyName(node);
        if (name) {
          allNames = parentNames.concat(name);
          if (node.parent.type === 'MemberExpression') {
            markPropTypesAsUsed(node.parent, allNames);
          }
          // Do not mark computed props as used.
          type = name !== '__COMPUTED_PROP__' ? 'direct' : null;
        } else if (
          node.parent.id &&
          node.parent.id.properties &&
          node.parent.id.properties.length &&
          getKeyValue(node.parent.id.properties[0])
        ) {
          type = 'destructuring';
          properties = node.parent.id.properties;
        }
        break;
      case 'VariableDeclarator':
        for (var i = 0, j = node.id.properties.length; i < j; i++) {
          if (
            (node.id.properties[i].key.name !== 'props' && node.id.properties[i].key.value !== 'props') ||
            node.id.properties[i].value.type !== 'ObjectPattern'
          ) {
            continue;
          }
          type = 'destructuring';
          properties = node.id.properties[i].value.properties;
          break;
        }
        break;
      default:
        throw new Error(node.type + ' ASTNodes are not handled by markPropTypesAsUsed');
    }

    var component = componentList.getByNode(context, node);
    var usedPropTypes = component && component.usedPropTypes || [];

    switch (type) {
      case 'direct':
        // Ignore Object methods
        if (Object.prototype[name]) {
          break;
        }
        usedPropTypes.push({
          name: name,
          allNames: allNames,
          node: node.object.name !== 'props' && !inConstructor() ? node.parent.property : node.property
        });
        break;
      case 'destructuring':
        for (var k = 0, l = properties.length; k < l; k++) {
          if (hasSpreadOperator(properties[k]) || properties[k].computed) {
            continue;
          }
          var propName = getKeyValue(properties[k]);

          var currentNode = node;
          allNames = [];
          while (currentNode.property && currentNode.property.name !== 'props') {
            allNames.unshift(currentNode.property.name);
            currentNode = currentNode.object;
          }
          allNames.push(propName);

          if (propName) {
            usedPropTypes.push({
              name: propName,
              allNames: allNames,
              node: properties[k]
            });
          }
        }
        break;
      default:
        break;
    }

    componentList.set(context, node, {
      usedPropTypes: usedPropTypes
    });
  }

  /**
   * Mark a prop type as declared
   * @param {ASTNode} node The AST node being checked.
   * @param {propTypes} node The AST node containing the proptypes
   */
  function markPropTypesAsDeclared(node, propTypes) {
    var component = componentList.getByNode(context, node);
    var declaredPropTypes = component && component.declaredPropTypes || {};
    var ignorePropsValidation = false;

    switch (propTypes && propTypes.type) {
      case 'ObjectExpression':
        iterateProperties(propTypes.properties, function(key, value) {
          declaredPropTypes[key] = buildReactDeclarationTypes(value);
        });
        break;
      case 'MemberExpression':
        var curDeclaredPropTypes = declaredPropTypes;
        // Walk the list of properties, until we reach the assignment
        // ie: ClassX.propTypes.a.b.c = ...
        while (
          propTypes &&
          propTypes.parent &&
          propTypes.parent.type !== 'AssignmentExpression' &&
          propTypes.property &&
          curDeclaredPropTypes
        ) {
          var propName = propTypes.property.name;
          if (propName in curDeclaredPropTypes) {
            curDeclaredPropTypes = curDeclaredPropTypes[propName].children;
            propTypes = propTypes.parent;
          } else {
            // This will crash at runtime because we haven't seen this key before
            // stop this and do not declare it
            propTypes = null;
          }
        }
        if (propTypes && propTypes.parent && propTypes.property) {
          curDeclaredPropTypes[propTypes.property.name] =
            buildReactDeclarationTypes(propTypes.parent.right);
        }
        break;
      case null:
        break;
      default:
        ignorePropsValidation = true;
        break;
    }

    componentList.set(context, node, {
      declaredPropTypes: declaredPropTypes,
      ignorePropsValidation: ignorePropsValidation
    });
  }

  /**
   * Reports undeclared proptypes for a given component
   * @param {Object} component The component to process
   */
  function reportUndeclaredPropTypes(component) {
    var allNames;
    for (var i = 0, j = component.usedPropTypes.length; i < j; i++) {
      allNames = component.usedPropTypes[i].allNames;
      if (
        isIgnored(allNames[0]) ||
        isDeclaredInComponent(component, allNames)
      ) {
        continue;
      }
      context.report(
        component.usedPropTypes[i].node,
        component.name === componentUtil.DEFAULT_COMPONENT_NAME ? MISSING_MESSAGE : MISSING_MESSAGE_NAMED_COMP, {
          name: allNames.join('.').replace(/\.__COMPUTED_PROP__/g, '[]'),
          component: component.name
        }
      );
    }
  }

  // --------------------------------------------------------------------------
  // Public
  // --------------------------------------------------------------------------

  return {

    ClassDeclaration: function(node) {
      componentList.set(context, node);
    },

    ClassProperty: function(node) {
      if (!isPropTypesDeclaration(node)) {
        return;
      }

      markPropTypesAsDeclared(node, node.value);
    },

    VariableDeclarator: function(node) {
      if (!node.init || node.init.type !== 'ThisExpression' || node.id.type !== 'ObjectPattern') {
        return;
      }
      markPropTypesAsUsed(node);
    },

    MemberExpression: function(node) {
      var type;
      if (isPropTypesUsage(node)) {
        type = 'usage';
      } else if (isPropTypesDeclaration(node.property)) {
        type = 'declaration';
      }

      switch (type) {
        case 'usage':
          markPropTypesAsUsed(node);
          break;
        case 'declaration':
          var component = componentList.getByName(node.object.name);
          if (!component) {
            return;
          }
          markPropTypesAsDeclared(component.node, node.parent.right || node.parent);
          break;
        default:
          break;
      }
    },

    MethodDefinition: function(node) {
      if (!isPropTypesDeclaration(node.key)) {
        return;
      }

      var i = node.value.body.body.length - 1;
      for (; i >= 0; i--) {
        if (node.value.body.body[i].type === 'ReturnStatement') {
          break;
        }
      }

      if (i >= 0) {
        markPropTypesAsDeclared(node, node.value.body.body[i].argument);
      }
    },

    ObjectExpression: function(node) {
      componentList.set(context, node);
      // Search for the proptypes declaration
      node.properties.forEach(function(property) {
        if (!isPropTypesDeclaration(property.key)) {
          return;
        }
        markPropTypesAsDeclared(node, property.value);
      });
    },

    ReturnStatement: function(node) {
      componentList.set(context, node);
    },

    'Program:exit': function() {
      var list = componentList.getList();
      // Report undeclared proptypes for all classes
      for (var component in list) {
        if (!list.hasOwnProperty(component) || !mustBeValidated(list[component])) {
          continue;
        }
        reportUndeclaredPropTypes(list[component]);
      }
    }
  };

};

module.exports.schema = [{
  type: 'object',
  properties: {
    ignore: {
      type: 'array',
      items: {
        type: 'string'
      }
    },
    customValidators: {
      type: 'array',
      items: {
        type: 'string'
      }
    }
  },
  additionalProperties: false
}];
