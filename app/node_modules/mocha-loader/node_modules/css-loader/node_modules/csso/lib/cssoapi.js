var util = require('./util.js'),
    gonzales = require('./gonzales.cssp.node.js'),
    translator = require('./translator.js'),
    compressor = require('./compressor.js');

var parse = exports.parse = function(s, rule, needInfo) {
    return gonzales.srcToCSSP(s, rule, needInfo);
};

var cleanInfo = exports.cleanInfo = util.cleanInfo;

exports.treeToString = util.treeToString;

exports.printTree = util.printTree;

var translate = exports.translate = translator.translate;

var compress = exports.compress = compressor.compress;

exports.justDoIt = function(src, ro, needInfo) {
    return translate(cleanInfo(compress(parse(src, 'stylesheet', needInfo), ro)));
};
