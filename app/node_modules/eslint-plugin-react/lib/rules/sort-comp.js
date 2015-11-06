/**
 * @fileoverview Enforce component methods order
 * @author Yannick Croissant
 */
'use strict';

var util = require('util');

var componentUtil = require('../util/component');
var ComponentList = componentUtil.List;

/**
 * Get the methods order from the default config and the user config
 * @param {Object} defaultConfig The default configuration.
 * @param {Object} userConfig The user configuration.
 * @returns {Array} Methods order
 */
function getMethodsOrder(defaultConfig, userConfig) {
  userConfig = userConfig || {};

  var groups = util._extend(defaultConfig.groups, userConfig.groups);
  var order = userConfig.order || defaultConfig.order;

  var config = [];
  var entry;
  for (var i = 0, j = order.length; i < j; i++) {
    entry = order[i];
    if (groups.hasOwnProperty(entry)) {
      config = config.concat(groups[entry]);
    } else {
      config.push(entry);
    }
  }

  return config;
}

// ------------------------------------------------------------------------------
// Rule Definition
// ------------------------------------------------------------------------------

module.exports = function(context) {

  var componentList = new ComponentList();
  var errors = {};

  var MISPOSITION_MESSAGE = '{{propA}} must be placed {{position}} {{propB}}';

  var methodsOrder = getMethodsOrder({
    order: [
      'lifecycle',
      'everything-else',
      'render'
    ],
    groups: {
      lifecycle: [
        'displayName',
        'propTypes',
        'contextTypes',
        'childContextTypes',
        'mixins',
        'statics',
        'defaultProps',
        'constructor',
        'getDefaultProps',
        'state',
        'getInitialState',
        'getChildContext',
        'componentWillMount',
        'componentDidMount',
        'componentWillReceiveProps',
        'shouldComponentUpdate',
        'componentWillUpdate',
        'componentDidUpdate',
        'componentWillUnmount'
      ]
    }
  }, context.options[0]);

  /**
   * Checks if the component must be validated
   * @param {Object} component The component to process
   * @returns {Boolean} True if the component must be validated, false if not.
   */
  function mustBeValidated(component) {
    return (
      component &&
      !component.hasDisplayName
    );
  }

  // --------------------------------------------------------------------------
  // Public
  // --------------------------------------------------------------------------

  var regExpRegExp = /\/(.*)\/([g|y|i|m]*)/;

  /**
   * Get indexes of the matching patterns in methods order configuration
   * @param {String} method - Method name.
   * @returns {Array} The matching patterns indexes. Return [Infinity] if there is no match.
   */
  function getRefPropIndexes(method) {
    var isRegExp;
    var matching;
    var i;
    var j;
    var indexes = [];
    for (i = 0, j = methodsOrder.length; i < j; i++) {
      isRegExp = methodsOrder[i].match(regExpRegExp);
      if (isRegExp) {
        matching = new RegExp(isRegExp[1], isRegExp[2]).test(method);
      } else {
        matching = methodsOrder[i] === method;
      }
      if (matching) {
        indexes.push(i);
      }
    }

    // No matching pattern, return 'everything-else' index
    if (indexes.length === 0) {
      for (i = 0, j = methodsOrder.length; i < j; i++) {
        if (methodsOrder[i] === 'everything-else') {
          indexes.push(i);
        }
      }
    }

    // No matching pattern and no 'everything-else' group
    if (indexes.length === 0) {
      indexes.push(Infinity);
    }

    return indexes;
  }

  /**
   * Get properties name
   * @param {Object} node - Property.
   * @returns {String} Property name.
   */
  function getPropertyName(node) {

    // Special case for class properties
    // (babel-eslint does not expose property name so we have to rely on tokens)
    if (node.type === 'ClassProperty') {
      var tokens = context.getFirstTokens(node, 2);
      return tokens[1].type === 'Identifier' ? tokens[1].value : tokens[0].value;
    }

    return node.key.name;
  }

  /**
   * Store a new error in the error list
   * @param {Object} propA - Mispositioned property.
   * @param {Object} propB - Reference property.
   */
  function storeError(propA, propB) {
    // Initialize the error object if needed
    if (!errors[propA.index]) {
      errors[propA.index] = {
        node: propA.node,
        score: 0,
        closest: {
          distance: Infinity,
          ref: {
            node: null,
            index: 0
          }
        }
      };
    }
    // Increment the prop score
    errors[propA.index].score++;
    // Stop here if we already have a closer reference
    if (Math.abs(propA.index - propB.index) > errors[propA.index].closest.distance) {
      return;
    }
    // Update the closest reference
    errors[propA.index].closest.distance = Math.abs(propA.index - propB.index);
    errors[propA.index].closest.ref.node = propB.node;
    errors[propA.index].closest.ref.index = propB.index;
  }

  /**
   * Dedupe errors, only keep the ones with the highest score and delete the others
   */
  function dedupeErrors() {
    for (var i in errors) {
      if (!errors.hasOwnProperty(i)) {
        continue;
      }
      var index = errors[i].closest.ref.index;
      if (!errors[index]) {
        continue;
      }
      if (errors[i].score > errors[index].score) {
        delete errors[index];
      } else {
        delete errors[i];
      }
    }
  }

  /**
   * Report errors
   */
  function reportErrors() {
    dedupeErrors();

    var nodeA;
    var nodeB;
    var indexA;
    var indexB;
    for (var i in errors) {
      if (!errors.hasOwnProperty(i)) {
        continue;
      }

      nodeA = errors[i].node;
      nodeB = errors[i].closest.ref.node;
      indexA = i;
      indexB = errors[i].closest.ref.index;

      context.report(nodeA, MISPOSITION_MESSAGE, {
        propA: getPropertyName(nodeA),
        propB: getPropertyName(nodeB),
        position: indexA < indexB ? 'before' : 'after'
      });
    }
  }

  /**
   * Get properties for a given AST node
   * @param {ASTNode} node The AST node being checked.
   * @returns {Array} Properties array.
   */
  function getComponentProperties(node) {
    if (node.type === 'ClassDeclaration') {
      return node.body.body;
    }
    return node.properties;
  }

  /**
   * Compare two properties and find out if they are in the right order
   * @param {Array} propertiesNames Array containing all the properties names.
   * @param {String} propA First property name.
   * @param {String} propB Second property name.
   * @returns {Object} Object containing a correct true/false flag and the correct indexes for the two properties.
   */
  function comparePropsOrder(propertiesNames, propA, propB) {
    var i;
    var j;
    var k;
    var l;
    var refIndexA;
    var refIndexB;

    // Get references indexes (the correct position) for given properties
    var refIndexesA = getRefPropIndexes(propA);
    var refIndexesB = getRefPropIndexes(propB);

    // Get current indexes for given properties
    var classIndexA = propertiesNames.indexOf(propA);
    var classIndexB = propertiesNames.indexOf(propB);

    // Loop around the references indexes for the 1st property
    for (i = 0, j = refIndexesA.length; i < j; i++) {
      refIndexA = refIndexesA[i];

      // Loop around the properties for the 2nd property (for comparison)
      for (k = 0, l = refIndexesB.length; k < l; k++) {
        refIndexB = refIndexesB[k];

        if (
          // Comparing the same properties
          refIndexA === refIndexB ||
          // 1st property is placed before the 2nd one in reference and in current component
          refIndexA < refIndexB && classIndexA < classIndexB ||
          // 1st property is placed after the 2nd one in reference and in current component
          refIndexA > refIndexB && classIndexA > classIndexB
        ) {
          return {
            correct: true,
            indexA: classIndexA,
            indexB: classIndexB
          };
        }

      }
    }

    // We did not find any correct match between reference and current component
    return {
      correct: false,
      indexA: refIndexA,
      indexB: refIndexB
    };
  }

  /**
   * Check properties order from a properties list and store the eventual errors
   * @param {Array} properties Array containing all the properties.
   */
  function checkPropsOrder(properties) {
    var propertiesNames = properties.map(getPropertyName);
    var i;
    var j;
    var k;
    var l;
    var propA;
    var propB;
    var order;

    // Loop around the properties
    for (i = 0, j = propertiesNames.length; i < j; i++) {
      propA = propertiesNames[i];

      // Loop around the properties a second time (for comparison)
      for (k = 0, l = propertiesNames.length; k < l; k++) {
        propB = propertiesNames[k];

        // Compare the properties order
        order = comparePropsOrder(propertiesNames, propA, propB);

        // Continue to next comparison is order is correct
        if (order.correct === true) {
          continue;
        }

        // Store an error if the order is incorrect
        storeError({
          node: properties[i],
          index: order.indexA
        }, {
          node: properties[k],
          index: order.indexB
        });
      }
    }

  }

  return {

    ClassDeclaration: function(node) {
      componentList.set(context, node);
    },

    ObjectExpression: function(node) {
      componentList.set(context, node);
    },

    'Program:exit': function() {
      var list = componentList.getList();
      for (var component in list) {
        if (!list.hasOwnProperty(component) || !mustBeValidated(list[component])) {
          continue;
        }
        var properties = getComponentProperties(list[component].node);
        checkPropsOrder(properties);
      }

      reportErrors();
    }
  };

};

module.exports.schema = [{
  type: 'object',
  properties: {
    order: {
      type: 'array',
      items: {
        type: 'string'
      }
    },
    groups: {
      type: 'object',
      patternProperties: {
        '^.*$': {
          type: 'array',
          items: {
            type: 'string'
          }
        }
      }
    }
  },
  additionalProperties: false
}];
