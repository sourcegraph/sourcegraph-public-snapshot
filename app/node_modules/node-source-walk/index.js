var acorn = require('acorn');

/**
 * @param  {Object} options - Options to configure parser
 * @param  {Boolean} [options.ecmaVersion=5]
 */
module.exports = function (options) {
  this.options = options || {};

  this.options.ecmaVersion = this.options.ecmaVersion || 6;
  this.options.sourceType = this.options.sourceType || 'module';

  // We use global state to stop the recursive traversal of the AST
  this.shouldStop = false;
};

/**
 * @param  {String} src
 * @param  {Object} [options] - Parser options
 * @return {Object} The AST of the given src
 */
module.exports.prototype.parse = function(src, options) {
  options = options || {};

  if (typeof options.allowHashBang === 'undefined') {
    options.allowHashBang = true;
  }

  return acorn.parse(src, options);
};

/**
 * Adapted from substack/node-detective
 * Executes cb on a non-array AST node
 */
module.exports.prototype.traverse = function (node, cb) {
  var that = this;

  if (this.shouldStop) return;

  if (Array.isArray(node)) {
    node.forEach(function (x) {
      if(x !== null) {
        // Mark that the node has been visited
        x.parent = node;
        that.traverse(x, cb);
      }
    });

  } else if (node && typeof node === 'object') {
    cb(node);

    Object.keys(node).forEach(function (key) {
      // Avoid visited nodes
      if (key === 'parent' || ! node[key]) return;

      node[key].parent = node;
      that.traverse(node[key], cb);
    });
  }
};

/**
 * Executes the passed callback for every traversed node of
 * the passed in src's ast
 *
 * @param {String|Object} src - The source code or AST to traverse
 * @param {Function} cb - Called for every node
 */
module.exports.prototype.walk = function (src, cb) {
  this.shouldStop = false;

  var ast = typeof src === 'object' ?
            src :
            this.parse(src, this.options);

  this.traverse(ast, cb);
};

/**
 * Halts further traversal of the AST
 */
module.exports.prototype.stopWalking = function () {
  this.shouldStop = true;
};
