'use strict';

module.exports = function trimLeadingZero (str) {
    return str.replace(/([^\d])0(\.\d*)/g, '$1$2');
};
