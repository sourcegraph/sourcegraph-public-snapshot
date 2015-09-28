/**
 * @fileoverview Prevent React to be marked as unused
 * @author Glen Mailer
 */
'use strict';

var variableUtil = require('../util/variable');

// ------------------------------------------------------------------------------
// Rule Definition
// ------------------------------------------------------------------------------

var JSX_ANNOTATION_REGEX = /^\*\s*@jsx\s+([^\s]+)/;

module.exports = function(context) {

  var config = context.options[0] || {};
  var id = config.pragma || 'React';

  // --------------------------------------------------------------------------
  // Public
  // --------------------------------------------------------------------------

  return {

    JSXOpeningElement: function() {
      variableUtil.markVariableAsUsed(context, id);
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

module.exports.schema = [{
  type: 'object',
  properties: {
    pragma: {
      type: 'string'
    }
  },
  additionalProperties: false
}];
