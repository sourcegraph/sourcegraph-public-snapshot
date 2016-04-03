'use strict';

Object.defineProperty(exports, "__esModule", {
    value: true
});

var _isFunction2 = require('lodash/isFunction');

var _isFunction3 = _interopRequireDefault(_isFunction2);

var _extendReactClass = require('./extendReactClass');

var _extendReactClass2 = _interopRequireDefault(_extendReactClass);

var _wrapStatelessFunction = require('./wrapStatelessFunction');

var _wrapStatelessFunction2 = _interopRequireDefault(_wrapStatelessFunction);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

/**
 * Determines if the given object has the signature of a class that inherits React.Component.
 */


/**
 * @see https://github.com/gajus/react-css-modules#options
 */
var isReactComponent = function isReactComponent(maybeReactComponent) {
    return 'prototype' in maybeReactComponent && (0, _isFunction3.default)(maybeReactComponent.prototype.render);
};

/**
 * When used as a function.
 */
var functionConstructor = function functionConstructor(Component, defaultStyles, options) {
    var decoratedClass = void 0;

    if (isReactComponent(Component)) {
        decoratedClass = (0, _extendReactClass2.default)(Component, defaultStyles, options);
    } else {
        decoratedClass = (0, _wrapStatelessFunction2.default)(Component, defaultStyles, options);
    }

    if (Component.displayName) {
        decoratedClass.displayName = Component.displayName;
    } else {
        decoratedClass.displayName = Component.name;
    }

    return decoratedClass;
};

/**
 * When used as a ES7 decorator.
 */
var decoratorConstructor = function decoratorConstructor(defaultStyles, options) {
    return function (Component) {
        return functionConstructor(Component, defaultStyles, options);
    };
};

exports.default = function () {
    if ((0, _isFunction3.default)(arguments.length <= 0 ? undefined : arguments[0])) {
        return functionConstructor(arguments.length <= 0 ? undefined : arguments[0], arguments.length <= 1 ? undefined : arguments[1], arguments.length <= 2 ? undefined : arguments[2]);
    } else {
        return decoratorConstructor(arguments.length <= 0 ? undefined : arguments[0], arguments.length <= 1 ? undefined : arguments[1]);
    }
};

module.exports = exports['default'];
//# sourceMappingURL=index.js.map
