/**
 * Vue.js 2.0 plugin
 *
 */
'use strict';

function vuePlugin(Raven, Vue) {
    Vue = Vue || window.Vue;

    // quit if Vue isn't on the page
    if (!Vue || !Vue.config) return;

    var _oldOnError = Vue.config.errorHandler;
    Vue.config.errorHandler = function VueErrorHandler(error, vm) {
        Raven.captureException(error, {
          extra: {
            componentName: Vue.util.formatComponentName(vm),
            propsData: vm.$options.propsData
          }
        });

        if (typeof _oldOnError === 'function') {
            _oldOnError.call(this, error, vm);
        }
    };
}

module.exports = vuePlugin;

