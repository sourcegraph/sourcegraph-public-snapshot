/* jshint loopfunc:true */

'use strict';

module.exports = function clone (obj, parent) {
    if (typeof obj !== 'object') {
        return obj;
    }
    var cloned = new obj.constructor();
    for (var i in obj) {
        if (!({}.hasOwnProperty.call(obj, i))) {
            continue;
        }
        var value = obj[i];
        if (i === 'parent' && typeof value === 'object') {
            if (parent) {
                cloned[i] = parent;
            }
        } else if (i === 'source') {
            cloned[i] = value;
        } else if (value instanceof Array) {
            cloned[i] = value.map(function (i) {
                return clone(i, cloned);
            });
        } else {
            cloned[i] = clone(value, cloned);
        }
    }
    return cloned;
};
