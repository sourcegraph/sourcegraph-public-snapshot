'use strict';

Object.defineProperty(exports, "__esModule", {
    value: true
});

var _isObject2 = require('lodash/isObject');

var _isObject3 = _interopRequireDefault(_isObject2);

var _isArray2 = require('lodash/isArray');

var _isArray3 = _interopRequireDefault(_isArray2);

var _react = require('react');

var _react2 = _interopRequireDefault(_react);

var _makeConfiguration = require('./makeConfiguration');

var _makeConfiguration2 = _interopRequireDefault(_makeConfiguration);

var _isIterable = require('./isIterable');

var _isIterable2 = _interopRequireDefault(_isIterable);

var _parseStyleName = require('./parseStyleName');

var _parseStyleName2 = _interopRequireDefault(_parseStyleName);

var _generateAppendClassName = require('./generateAppendClassName');

var _generateAppendClassName2 = _interopRequireDefault(_generateAppendClassName);

var _objectUnfreeze = require('object-unfreeze');

var _objectUnfreeze2 = _interopRequireDefault(_objectUnfreeze);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

var linkElement = function linkElement(element, styles, configuration) {
    var appendClassName = void 0,
        elementIsFrozen = void 0,
        elementShallowCopy = void 0;

    elementShallowCopy = element;

    if (Object.isFrozen && Object.isFrozen(elementShallowCopy)) {
        elementIsFrozen = true;

        // https://github.com/facebook/react/blob/v0.13.3/src/classic/element/ReactElement.js#L131
        elementShallowCopy = (0, _objectUnfreeze2.default)(elementShallowCopy);
        elementShallowCopy.props = (0, _objectUnfreeze2.default)(elementShallowCopy.props);
    }

    var styleNames = (0, _parseStyleName2.default)(elementShallowCopy.props.styleName || '', configuration.allowMultiple);

    if (_react2.default.isValidElement(elementShallowCopy.props.children)) {
        elementShallowCopy.props.children = linkElement(_react2.default.Children.only(elementShallowCopy.props.children), styles, configuration);
    } else if ((0, _isArray3.default)(elementShallowCopy.props.children) || (0, _isIterable2.default)(elementShallowCopy.props.children)) {
        elementShallowCopy.props.children = _react2.default.Children.map(elementShallowCopy.props.children, function (node) {
            if (_react2.default.isValidElement(node)) {
                return linkElement(node, styles, configuration);
            } else {
                return node;
            }
        });
    }

    if (styleNames.length) {
        appendClassName = (0, _generateAppendClassName2.default)(styles, styleNames, configuration.errorWhenNotFound);

        if (appendClassName) {
            if (elementShallowCopy.props.className) {
                appendClassName = elementShallowCopy.props.className + ' ' + appendClassName;
            }

            elementShallowCopy.props.className = appendClassName;
            elementShallowCopy.props.styleName = null;
        }
    }

    if (elementIsFrozen) {
        Object.freeze(elementShallowCopy.props);
        Object.freeze(elementShallowCopy);
    }

    return elementShallowCopy;
};

/**
 * @param {ReactElement} element
 * @param {Object} styles CSS modules class map.
 * @param {CSSModules~Options} userConfiguration
 */

exports.default = function (element) {
    var styles = arguments.length <= 1 || arguments[1] === undefined ? {} : arguments[1];
    var userConfiguration = arguments[2];

    // @see https://github.com/gajus/react-css-modules/pull/30
    if (!(0, _isObject3.default)(element)) {
        return element;
    }

    var configuration = (0, _makeConfiguration2.default)(userConfiguration);

    return linkElement(element, styles, configuration);
};

module.exports = exports['default'];
//# sourceMappingURL=linkClass.js.map
