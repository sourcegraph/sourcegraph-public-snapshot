/**
 * @fileoverview Prevent missing React when using JSX
 * @author Glen Mailer
 */
'use strict';

var variableUtil = require('../util/variable');

// -----------------------------------------------------------------------------
// Rule Definition
// -----------------------------------------------------------------------------

var JSX_ANNOTATION_REGEX = /^\*\s*@jsx\s+([^\s]+)/;

module.exports = function(context) {

  var id = 'React';
  var NOT_DEFINED_MESSAGE = '\'{{name}}\' must be in scope when using JSX';

  return {

    JSXOpeningElement: function(node) {
      var variables = variableUtil.variablesInScope(context);
      if (variableUtil.findVariable(variables, id)) {
        return;
      }
      context.report(node, NOT_DEFINED_MESSAGE, {
        name: id
      });
    },

    BlockComment: function(node) {
      var matches = JSX_ANNOTATION_REGEX.exec(node.value);
      if (!matches) {
        return;
      }
      id = matches[1].split('.')[0];
    }

  };

};

module.exports.schema = [];
