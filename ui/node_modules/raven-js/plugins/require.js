/*global define*/
/**
 * require.js plugin
 *
 * Automatically wrap define/require callbacks. (Experimental)
 */
'use strict';

function requirePlugin(Raven) {
    if (typeof define === 'function' && define.amd) {
        window.define = Raven.wrap({deep: false}, define);
        window.require = Raven.wrap({deep: false}, require);
    }
}

module.exports = requirePlugin;
