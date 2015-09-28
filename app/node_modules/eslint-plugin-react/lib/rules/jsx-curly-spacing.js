/**
 * @fileoverview Enforce or disallow spaces inside of curly braces in JSX attributes.
 * @author Jamund Ferguson, Brandyn Bennett, Michael Ficarra, Vignesh Anand, Jamund Ferguson, Yannick Croissant
 */
'use strict';

// ------------------------------------------------------------------------------
// Rule Definition
// ------------------------------------------------------------------------------

module.exports = function(context) {
  var spaced = context.options[0] === 'always';
  var multiline = context.options[1] ? context.options[1].allowMultiline : true;

  // --------------------------------------------------------------------------
  // Helpers
  // --------------------------------------------------------------------------

  /**
   * Determines whether two adjacent tokens have a newline between them.
   * @param {Object} left - The left token object.
   * @param {Object} right - The right token object.
   * @returns {boolean} Whether or not there is a newline between the tokens.
   */
  function isMultiline(left, right) {
    return left.loc.start.line !== right.loc.start.line;
  }

  /**
   * Determines whether two adjacent tokens have whitespace between them.
   * @param {Object} left - The left token object.
   * @param {Object} right - The right token object.
   * @returns {boolean} Whether or not there is space between the tokens.
   */
  function isSpaced(left, right) {
    return left.range[1] < right.range[0];
  }

  /**
  * Reports that there shouldn't be a newline after the first token
  * @param {ASTNode} node - The node to report in the event of an error.
  * @param {Token} token - The token to use for the report.
  * @returns {void}
  */
  function reportNoBeginningNewline(node, token) {
    context.report(node, token.loc.start,
      'There should be no newline after \'' + token.value + '\'');
  }

  /**
  * Reports that there shouldn't be a newline before the last token
  * @param {ASTNode} node - The node to report in the event of an error.
  * @param {Token} token - The token to use for the report.
  * @returns {void}
  */
  function reportNoEndingNewline(node, token) {
    context.report(node, token.loc.start,
      'There should be no newline before \'' + token.value + '\'');
  }

  /**
  * Reports that there shouldn't be a space after the first token
  * @param {ASTNode} node - The node to report in the event of an error.
  * @param {Token} token - The token to use for the report.
  * @returns {void}
  */
  function reportNoBeginningSpace(node, token) {
    context.report(node, token.loc.start,
      'There should be no space after \'' + token.value + '\'');
  }

  /**
  * Reports that there shouldn't be a space before the last token
  * @param {ASTNode} node - The node to report in the event of an error.
  * @param {Token} token - The token to use for the report.
  * @returns {void}
  */
  function reportNoEndingSpace(node, token) {
    context.report(node, token.loc.start,
      'There should be no space before \'' + token.value + '\'');
  }

  /**
  * Reports that there should be a space after the first token
  * @param {ASTNode} node - The node to report in the event of an error.
  * @param {Token} token - The token to use for the report.
  * @returns {void}
  */
  function reportRequiredBeginningSpace(node, token) {
    context.report(node, token.loc.start,
      'A space is required after \'' + token.value + '\'');
  }

  /**
  * Reports that there should be a space before the last token
  * @param {ASTNode} node - The node to report in the event of an error.
  * @param {Token} token - The token to use for the report.
  * @returns {void}
  */
  function reportRequiredEndingSpace(node, token) {
    context.report(node, token.loc.start,
          'A space is required before \'' + token.value + '\'');
  }

  /**
   * Determines if spacing in curly braces is valid.
   * @param {ASTNode} node The AST node to check.
   * @param {Token} first The first token to check (should be the opening brace)
   * @param {Token} second The second token to check (should be first after the opening brace)
   * @param {Token} penultimate The penultimate token to check (should be last before closing brace)
   * @param {Token} last The last token to check (should be closing brace)
   * @returns {void}
   */
  function validateBraceSpacing(node, first, second, penultimate, last) {
    if (spaced) {
      if (!isSpaced(first, second)) {
        reportRequiredBeginningSpace(node, first);
      } else if (!multiline && isMultiline(first, second)) {
        reportNoBeginningNewline(node, first);
      }

      if (!isSpaced(penultimate, last)) {
        reportRequiredEndingSpace(node, last);
      } else if (!multiline && isMultiline(penultimate, last)) {
        reportNoEndingNewline(node, last);
      }

      return;
    }

    // "never" setting if we get here.
    if (isSpaced(first, second) && !(multiline && isMultiline(first, second))) {
      reportNoBeginningSpace(node, first);
    }

    if (isSpaced(penultimate, last) && !(multiline && isMultiline(penultimate, last))) {
      reportNoEndingSpace(node, last);
    }
  }

  // --------------------------------------------------------------------------
  // Public
  // --------------------------------------------------------------------------

  return {
    JSXExpressionContainer: function(node) {
      var first = context.getFirstToken(node);
      var second = context.getFirstToken(node, 1);
      var penultimate = context.getLastToken(node, 1);
      var last = context.getLastToken(node);

      if (first === penultimate && second === last) {
        return;
      }

      validateBraceSpacing(node, first, second, penultimate, last);
    }
  };
};

module.exports.schema = [{
  enum: ['always', 'never']
}, {
  type: 'object',
  properties: {
    allowMultiline: {
      type: 'boolean'
    }
  },
  additionalProperties: false
}];
