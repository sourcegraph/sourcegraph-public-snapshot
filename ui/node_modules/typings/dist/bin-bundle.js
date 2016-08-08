"use strict";
var typings_core_1 = require('typings-core');
function help() {
    return "\ntypings bundle --out <filepath>\n\nOptions:\n  [--out|-o] <filepath>  The bundled output file path\n  [--global|-G]          Bundle as an global definition\n";
}
exports.help = help;
function exec(args, options) {
    return typings_core_1.bundle(options);
}
exports.exec = exec;
//# sourceMappingURL=bin-bundle.js.map