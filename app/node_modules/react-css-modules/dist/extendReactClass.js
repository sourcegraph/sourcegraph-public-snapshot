'use strict';

Object.defineProperty(exports, "__esModule", {
    value: true
});

var _assign2 = require('lodash/assign');

var _assign3 = _interopRequireDefault(_assign2);

var _isObject2 = require('lodash/isObject');

var _isObject3 = _interopRequireDefault(_isObject2);

var _linkClass = require('./linkClass');

var _linkClass2 = _interopRequireDefault(_linkClass);

var _react = require('react');

var _react2 = _interopRequireDefault(_react);

var _hoistNonReactStatics = require('hoist-non-react-statics');

var _hoistNonReactStatics2 = _interopRequireDefault(_hoistNonReactStatics);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _defaults(obj, defaults) { var keys = Object.getOwnPropertyNames(defaults); for (var i = 0; i < keys.length; i++) { var key = keys[i]; var value = Object.getOwnPropertyDescriptor(defaults, key); if (value && value.configurable && obj[key] === undefined) { Object.defineProperty(obj, key, value); } } return obj; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _possibleConstructorReturn(self, call) { if (!self) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return call && (typeof call === "object" || typeof call === "function") ? call : self; }

function _inherits(subClass, superClass) { if (typeof superClass !== "function" && superClass !== null) { throw new TypeError("Super expression must either be null or a function, not " + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? _defaults(subClass, superClass) : _defaults(subClass, superClass); } /* eslint-disable react/prop-types */

/**
 * @param {ReactClass} Component
 * @param {Object} defaultStyles
 * @param {Object} options
 * @returns {ReactClass}
 */

exports.default = function (Component, defaultStyles, options) {
    var WrappedComponent = function (_Component) {
        _inherits(WrappedComponent, _Component);

        function WrappedComponent() {
            _classCallCheck(this, WrappedComponent);

            return _possibleConstructorReturn(this, _Component.apply(this, arguments));
        }

        WrappedComponent.prototype.render = function render() {
            var propsChanged = void 0,
                styles = void 0;

            propsChanged = false;

            if (this.props.styles) {
                styles = this.props.styles;
            } else if ((0, _isObject3.default)(defaultStyles)) {
                this.props = (0, _assign3.default)({}, this.props, {
                    styles: defaultStyles
                });

                propsChanged = true;
                styles = defaultStyles;
            } else {
                styles = {};
            }

            var renderResult = _Component.prototype.render.call(this);

            if (propsChanged) {
                delete this.props.styles;
            }

            if (renderResult) {
                return (0, _linkClass2.default)(renderResult, styles, options);
            }

            return _react2.default.createElement('noscript');
        };

        return WrappedComponent;
    }(Component);

    return (0, _hoistNonReactStatics2.default)(WrappedComponent, Component);
};

module.exports = exports['default'];
//# sourceMappingURL=extendReactClass.js.map
