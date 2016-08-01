"use strict";
function isHostObject(object) {
    return typeof object.pipe === 'function' || Buffer.isBuffer(object);
}
Object.defineProperty(exports, "__esModule", { value: true });
exports.default = isHostObject;
//# sourceMappingURL=index.js.map