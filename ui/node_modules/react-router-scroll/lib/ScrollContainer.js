'use strict';

exports.__esModule = true;

var _react = require('react');

var _react2 = _interopRequireDefault(_react);

var _reactDom = require('react-dom');

var _reactDom2 = _interopRequireDefault(_reactDom);

var _warning = require('warning');

var _warning2 = _interopRequireDefault(_warning);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _possibleConstructorReturn(self, call) { if (!self) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return call && (typeof call === "object" || typeof call === "function") ? call : self; }

function _inherits(subClass, superClass) { if (typeof superClass !== "function" && superClass !== null) { throw new TypeError("Super expression must either be null or a function, not " + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

var propTypes = {
  scrollKey: _react2.default.PropTypes.string.isRequired,
  shouldUpdateScroll: _react2.default.PropTypes.func,
  children: _react2.default.PropTypes.element.isRequired
};

var contextTypes = {
  // This is necessary when rendering on the client. However, when rendering on
  // the server, this container will do nothing, and thus does not require the
  // scroll behavior context.
  scrollBehavior: _react2.default.PropTypes.object
};

var ScrollContainer = function (_React$Component) {
  _inherits(ScrollContainer, _React$Component);

  function ScrollContainer(props, context) {
    _classCallCheck(this, ScrollContainer);

    // We don't re-register if the scroll key changes, so make sure we
    // unregister with the initial scroll key just in case the user changes it.
    var _this = _possibleConstructorReturn(this, _React$Component.call(this, props, context));

    _this.shouldUpdateScroll = function (prevRouterProps, routerProps) {
      var shouldUpdateScroll = _this.props.shouldUpdateScroll;

      if (!shouldUpdateScroll) {
        return true;
      }

      // Hack to allow accessing scrollBehavior._stateStorage.
      return shouldUpdateScroll.call(_this.context.scrollBehavior.scrollBehavior, prevRouterProps, routerProps);
    };

    _this.scrollKey = props.scrollKey;
    return _this;
  }

  ScrollContainer.prototype.componentDidMount = function componentDidMount() {
    this.context.scrollBehavior.registerElement(this.props.scrollKey, _reactDom2.default.findDOMNode(this), // eslint-disable-line react/no-find-dom-node
    this.shouldUpdateScroll);

    // Only keep around the current DOM node in development, as this is only
    // for emitting the appropriate warning.
    if (process.env.NODE_ENV !== 'production') {
      this.domNode = _reactDom2.default.findDOMNode(this); // eslint-disable-line react/no-find-dom-node
    }
  };

  ScrollContainer.prototype.componentWillReceiveProps = function componentWillReceiveProps(nextProps) {
    process.env.NODE_ENV !== 'production' ? (0, _warning2.default)(nextProps.scrollKey === this.props.scrollKey, '<ScrollContainer> does not support changing scrollKey.') : void 0;
  };

  ScrollContainer.prototype.componentDidUpdate = function componentDidUpdate() {
    if (process.env.NODE_ENV !== 'production') {
      var prevDomNode = this.domNode;
      this.domNode = _reactDom2.default.findDOMNode(this); // eslint-disable-line react/no-find-dom-node

      process.env.NODE_ENV !== 'production' ? (0, _warning2.default)(this.domNode === prevDomNode, '<ScrollContainer> does not support changing DOM node.') : void 0;
    }
  };

  ScrollContainer.prototype.componentWillUnmount = function componentWillUnmount() {
    this.context.scrollBehavior.unregisterElement(this.scrollKey);
  };

  ScrollContainer.prototype.render = function render() {
    return this.props.children;
  };

  return ScrollContainer;
}(_react2.default.Component);

ScrollContainer.propTypes = propTypes;
ScrollContainer.contextTypes = contextTypes;

exports.default = ScrollContainer;
module.exports = exports['default'];