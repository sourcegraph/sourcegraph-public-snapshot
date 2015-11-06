/**
 * @fileoverview Prevent usage of setState in componentDidMount
 * @author David Petersen
 */
'use strict';

var componentUtil = require('../util/component');
var ComponentList = componentUtil.List;

// ------------------------------------------------------------------------------
// Rule Definition
// ------------------------------------------------------------------------------

module.exports = function(context) {

  var componentList = new ComponentList();

  /**
   * Checks if the component is valid
   * @param {Object} component The component to process
   * @returns {Boolean} True if the component is valid, false if not.
   */
  function isValid(component) {
    return Boolean(component && !component.mutateSetState);
  }

  /**
   * Reports undeclared proptypes for a given component
   * @param {Object} component The component to process
   */
  function reportMutations(component) {
    var mutation;
    for (var i = 0, j = component.mutations.length; i < j; i++) {
      mutation = component.mutations[i];
      context.report(mutation, 'Do not mutate state directly. Use setState().');
    }
  }

  // --------------------------------------------------------------------------
  // Public
  // --------------------------------------------------------------------------

  return {

    ObjectExpression: function(node) {
      componentList.set(context, node);
    },

    ClassDeclaration: function(node) {
      componentList.set(context, node);
    },

    AssignmentExpression: function(node) {
      var item;
      if (!node.left || !node.left.object || !node.left.object.object) {
        return;
      }
      item = node.left.object;
      while (item.object.property) {
        item = item.object;
      }
      if (
        item.object.type === 'ThisExpression' &&
        item.property.name === 'state'
      ) {
        var component = componentList.getByNode(context, node);
        var mutations = component && component.mutations || [];
        mutations.push(node.left.object);
        componentList.set(context, node, {
          mutateSetState: true,
          mutations: mutations
        });
      }
    },

    'Program:exit': function() {
      var list = componentList.getList();
      for (var component in list) {
        if (!list.hasOwnProperty(component) || isValid(list[component])) {
          continue;
        }
        reportMutations(list[component]);
      }
    }
  };

};
