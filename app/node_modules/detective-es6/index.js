var Walker = require('node-source-walk');

/**
 * Extracts the dependencies of the supplied es6 module
 *
 * @param  {String|Object} src - File's content or AST
 * @return {String[]}
 */
module.exports = function(src) {
  var walker = new Walker();

  var dependencies = [];

  if (! src) throw new Error('src not given');

  walker.walk(src, function(node) {
    // If it's not an import, skip it
    if (node.type !== 'ImportDeclaration' || !node.source || !node.source.value) {
      return;
    }

    dependencies.push(node.source.value);
  });

  return dependencies;
};
