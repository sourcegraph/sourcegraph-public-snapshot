'use strict';

Object.defineProperty(exports, "__esModule", {
    value: true
});

var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();

var _react = require('react');

var _react2 = _interopRequireDefault(_react);

var _reactDom = require('react-dom');

var _reactDom2 = _interopRequireDefault(_reactDom);

var _elementResizeDetector = require('element-resize-detector');

var _elementResizeDetector2 = _interopRequireDefault(_elementResizeDetector);

var _invariant = require('invariant');

var _invariant2 = _interopRequireDefault(_invariant);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _possibleConstructorReturn(self, call) { if (!self) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return call && (typeof call === "object" || typeof call === "function") ? call : self; }

function _inherits(subClass, superClass) { if (typeof superClass !== "function" && superClass !== null) { throw new TypeError("Super expression must either be null or a function, not " + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

var ContainerDimensions = function (_Component) {
    _inherits(ContainerDimensions, _Component);

    function ContainerDimensions() {
        _classCallCheck(this, ContainerDimensions);

        var _this = _possibleConstructorReturn(this, Object.getPrototypeOf(ContainerDimensions).call(this));

        _this.state = {
            initiated: false
        };
        _this.onResize = _this.onResize.bind(_this);
        return _this;
    }

    _createClass(ContainerDimensions, [{
        key: 'componentDidMount',
        value: function componentDidMount() {
            this.parentNode = _reactDom2.default.findDOMNode(this).parentNode;
            this.elementResizeDetector = (0, _elementResizeDetector2.default)({ strategy: 'scroll' });
            this.elementResizeDetector.listenTo(this.parentNode, this.onResize);
            this.onResize();
        }
    }, {
        key: 'componentWillUnmount',
        value: function componentWillUnmount() {
            this.elementResizeDetector.removeListener(this.parentNode, this.onResize);
        }
    }, {
        key: 'onResize',
        value: function onResize() {
            var clientRect = this.parentNode.getBoundingClientRect();
            this.setState({
                initiated: true,
                width: clientRect.width,
                height: clientRect.height
            });
        }
    }, {
        key: 'render',
        value: function render() {
            (0, _invariant2.default)(this.props.children, 'Expected children to be one of function or React.Element');

            if (!this.state.initiated) {
                return _react2.default.createElement('div', null);
            }
            if (typeof this.props.children === 'function') {
                var renderedChildren = this.props.children(this.state);
                return renderedChildren && _react.Children.only(renderedChildren);
            }
            return _react.Children.only(_react2.default.cloneElement(this.props.children, this.state));
        }
    }]);

    return ContainerDimensions;
}(_react.Component);

ContainerDimensions.propTypes = {
    children: _react.PropTypes.oneOfType([_react.PropTypes.element, _react.PropTypes.func]).isRequired
};
exports.default = ContainerDimensions;