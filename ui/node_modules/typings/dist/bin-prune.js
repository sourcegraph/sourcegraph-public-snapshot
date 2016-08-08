"use strict";
var typings_core_1 = require('typings-core');
function help() {
    return "\ntypings prune\n\nOptions:\n  [--production] Also prune non-production dependencies\n";
}
exports.help = help;
function exec(args, options) {
    return typings_core_1.prune(options);
}
exports.exec = exec;
//# sourceMappingURL=bin-prune.js.map