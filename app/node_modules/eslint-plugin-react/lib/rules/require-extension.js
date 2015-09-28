/**
 * @fileoverview Restrict file extensions that may be required
 * @author Scott Andrews
 */
'use strict';

var path = require('path');

// ------------------------------------------------------------------------------
// Constants
// ------------------------------------------------------------------------------

var DEFAULTS = {
  extentions: ['.jsx']
};

// ------------------------------------------------------------------------------
// Rule Definition
// ------------------------------------------------------------------------------

module.exports = function(context) {

  function isRequire(expression) {
    return expression.callee.name === 'require';
  }

  function getId(expression) {
    return expression.arguments[0] && expression.arguments[0].value;
  }

  function getExtension(id) {
    return path.extname(id || '');
  }

  function getExtentionsConfig() {
    return context.options[0] && context.options[0].extensions || DEFAULTS.extentions;
  }

  var forbiddenExtensions = getExtentionsConfig().reduce(function (extensions, extension) {
    extensions[extension] = true;
    return extensions;
  }, Object.create(null));

  function isForbiddenExtension(ext) {
    return ext in forbiddenExtensions;
  }

  // --------------------------------------------------------------------------
  // Public
  // --------------------------------------------------------------------------

  return {

    CallExpression: function(node) {
      if (isRequire(node)) {
        var ext = getExtension(getId(node));
        if (isForbiddenExtension(ext)) {
          context.report(node, 'Unable to require module with extension \'' + ext + '\'');
        }
      }
    }

  };

};

module.exports.schema = [{
  type: 'object',
  properties: {
    extensions: {
      type: 'array',
      items: {
        type: 'string'
      }
    }
  },
  additionalProperties: false
}];
