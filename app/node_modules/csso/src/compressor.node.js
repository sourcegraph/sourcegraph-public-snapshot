var translator = require('./translator.js').translator(),
    cleanInfo = require('./util.js').cleanInfo;

exports.compress = function(tree, ro) {
    return new CSSOCompressor().compress(tree, ro);
};
