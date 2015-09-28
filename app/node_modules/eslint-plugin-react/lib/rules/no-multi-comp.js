/**
 * @fileoverview Prevent multiple component definition per file
 * @author Yannick Croissant
 */
'use strict';

var componentUtil = require('../util/component');
var ComponentList = componentUtil.List;

// ------------------------------------------------------------------------------
// Rule Definition
// ------------------------------------------------------------------------------

module.exports = function(context) {

  var componentList = new ComponentList();

  var MULTI_COMP_MESSAGE = 'Declare only one React component per file';

  // --------------------------------------------------------------------------
  // Public
  // --------------------------------------------------------------------------

  return {
    'Program:exit': function() {
      if (componentList.count() <= 1) {
        return;
      }

      var list = componentList.getList();
      var i = 0;

      for (var component in list) {
        if (!list.hasOwnProperty(component) || ++i === 1) {
          continue;
        }
        context.report(list[component].node, MULTI_COMP_MESSAGE);
      }
    },

    ReturnStatement: function(node) {
      if (!componentUtil.isReactComponent(context, node)) {
        return;
      }
      componentList.set(context, node);
    }
  };
};

module.exports.schema = [];
