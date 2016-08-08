'use strict';

var _babelTemplate = require('babel-template');

var _babelTemplate2 = _interopRequireDefault(_babelTemplate);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

var buildRegistration = (0, _babelTemplate2.default)('__REACT_HOT_LOADER__.register(ID, NAME, FILENAME);');
var buildSemi = (0, _babelTemplate2.default)(';');
var buildTagger = (0, _babelTemplate2.default)('\n(function () {\n  if (typeof __REACT_HOT_LOADER__ === \'undefined\') {\n    return;\n  }\n\n  REGISTRATIONS\n})();\n');

module.exports = function plugin(args) {
  // This is a Babel plugin, but the user put it in the Webpack config.
  if (this && this.callback) {
    throw new Error('React Hot Loader: You are erroneously trying to use a Babel plugin ' + 'as a Webpack loader. We recommend that you use Babel, ' + 'remove "react-hot-loader/babel" from the "loaders" section ' + 'of your Webpack configuration, and instead add ' + '"react-hot-loader/babel" to the "plugins" section of your .babelrc file. ' + 'If you prefer not to use Babel, replace "react-hot-loader/babel" with ' + '"react-hot-loader/webpack" in the "loaders" section of your Webpack configuration. ');
  }
  var t = args.types;

  // No-op in production.

  if (process.env.NODE_ENV === 'production') {
    return { visitor: {} };
  }

  // Gather top-level variables, functions, and classes.
  // Try our best to avoid variables from require().
  // Ideally we only want to find components defined by the user.
  function shouldRegisterBinding(binding) {
    var _binding$path = binding.path;
    var type = _binding$path.type;
    var node = _binding$path.node;

    switch (type) {
      case 'FunctionDeclaration':
      case 'ClassDeclaration':
      case 'VariableDeclaration':
        return true;
      case 'VariableDeclarator':
        {
          var init = node.init;

          if (t.isCallExpression(init) && init.callee.name === 'require') {
            return false;
          }
          return true;
        }
      default:
        return false;
    }
  }

  var REGISTRATIONS = Symbol();
  return {
    visitor: {
      ExportDefaultDeclaration: function ExportDefaultDeclaration(path, _ref) {
        var file = _ref.file;

        // Default exports with names are going
        // to be in scope anyway so no need to bother.
        if (path.node.declaration.id) {
          return;
        }

        // Move export default right hand side to a variable
        // so we can later refer to it and tag it with __source.
        var id = path.scope.generateUidIdentifier('default');
        var expression = t.isExpression(path.node.declaration) ? path.node.declaration : t.toExpression(path.node.declaration);
        path.insertBefore(t.variableDeclaration('const', [t.variableDeclarator(id, expression)]));
        path.node.declaration = id; // eslint-disable-line no-param-reassign

        // It won't appear in scope.bindings
        // so we'll manually remember it exists.
        path.parent[REGISTRATIONS].push(buildRegistration({
          ID: id,
          NAME: t.stringLiteral('default'),
          FILENAME: t.stringLiteral(file.opts.filename)
        }));
      },


      Program: {
        enter: function enter(_ref2, _ref3) {
          var node = _ref2.node;
          var scope = _ref2.scope;
          var file = _ref3.file;

          node[REGISTRATIONS] = []; // eslint-disable-line no-param-reassign

          // Everything in the top level scope, when reasonable,
          // is going to get tagged with __source.
          /* eslint-disable guard-for-in,no-restricted-syntax */
          for (var id in scope.bindings) {
            var binding = scope.bindings[id];
            if (shouldRegisterBinding(binding)) {
              node[REGISTRATIONS].push(buildRegistration({
                ID: binding.identifier,
                NAME: t.stringLiteral(id),
                FILENAME: t.stringLiteral(file.opts.filename)
              }));
            }
          }
          /* eslint-enable */
        },
        exit: function exit(_ref4) {
          var node = _ref4.node;

          var registrations = node[REGISTRATIONS];
          node[REGISTRATIONS] = null; // eslint-disable-line no-param-reassign

          // Inject the generated tagging code at the very end
          // so that it is as minimally intrusive as possible.
          node.body.push(buildSemi());
          node.body.push(buildTagger({
            REGISTRATIONS: registrations
          }));
          node.body.push(buildSemi());
        }
      }
    }
  };
};