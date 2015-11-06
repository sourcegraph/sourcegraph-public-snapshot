var acorn = require('acorn');
var walk = require('acorn/dist/walk');
var escodegen = require('escodegen');
var defined = require('defined');

var requireRe = /\brequire\b/;

function parse (src, opts) {
    if (!opts) opts = {};
    return acorn.parse(src, {
        ecmaVersion: defined(opts.ecmaVersion, 6),
        sourceType: opts.sourceType,
        ranges: defined(opts.ranges, opts.range),
        locations: defined(opts.locations, opts.loc),
        allowReserved: defined(opts.allowReserved, true),
        allowReturnOutsideFunction: defined(
            opts.allowReturnOutsideFunction, true
        ),
        allowHashBang: defined(opts.allowHashBang, true)
    });
}

var exports = module.exports = function (src, opts) {
    return exports.find(src, opts).strings;
};

exports.find = function (src, opts) {
    if (!opts) opts = {};
    
    var word = opts.word === undefined ? 'require' : opts.word;
    if (typeof src !== 'string') src = String(src);
    
    var isRequire = opts.isRequire || function (node) {
        return node.callee.type === 'Identifier'
            && node.callee.name === word
        ;
    };
    
    var modules = { strings : [], expressions : [] };
    if (opts.nodes) modules.nodes = [];
    
    var wordRe = word === 'require' ? requireRe : RegExp('\\b' + word + '\\b');
    if (!wordRe.test(src)) return modules;
    
    var ast = parse(src, opts.parse);
    
    walk.simple(ast, {
        CallExpression: function (node) {
            if (!isRequire(node)) return;
            if (node.arguments.length) {
                if (node.arguments[0].type === 'Literal') {
                    modules.strings.push(node.arguments[0].value);
                }
                else {
                    modules.expressions.push(escodegen.generate(node.arguments[0]));
                }
            }
            if (opts.nodes) modules.nodes.push(node);
        }
    });
    
    return modules;
};
