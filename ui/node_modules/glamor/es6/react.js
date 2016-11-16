var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();

function _defineProperty(obj, key, value) { if (key in obj) { Object.defineProperty(obj, key, { value: value, enumerable: true, configurable: true, writable: true }); } else { obj[key] = value; } return obj; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _possibleConstructorReturn(self, call) { if (!self) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return call && (typeof call === "object" || typeof call === "function") ? call : self; }

function _inherits(subClass, superClass) { if (typeof superClass !== "function" && superClass !== null) { throw new TypeError("Super expression must either be null or a function, not " + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

function _toConsumableArray(arr) { if (Array.isArray(arr)) { for (var i = 0, arr2 = Array(arr.length); i < arr.length; i++) { arr2[i] = arr[i]; } return arr2; } else { return Array.from(arr); } }

function _objectWithoutProperties(obj, keys) { var target = {}; for (var i in obj) { if (keys.indexOf(i) >= 0) continue; if (!Object.prototype.hasOwnProperty.call(obj, i)) continue; target[i] = obj[i]; } return target; }

import React, { PropTypes } from 'react';
import { isLikeRule, style, merge } from './index.js';

export * from './index.js'; // convenience

// allows for elements to have a 'css' prop
export function createElement(tag, allProps) {
  for (var _len = arguments.length, children = Array(_len > 2 ? _len - 2 : 0), _key = 2; _key < _len; _key++) {
    children[_key - 2] = arguments[_key];
  }

  // todo - pull ids from className as well?
  if (allProps && allProps.css) {
    var css = allProps.css,
        className = allProps.className,
        props = _objectWithoutProperties(allProps, ['css', 'className']);

    var rule = Array.isArray(css) ? merge.apply(undefined, _toConsumableArray(css)) : isLikeRule(css) ? css : style(css);
    className = className ? className + ' ' + rule : rule;
    props.className = className;
    return React.createElement.apply(React, [tag, props].concat(children));
  }
  return React.createElement.apply(React, [tag, allProps].concat(children));
}

export var dom = createElement;

// css vars, done right
// see examples/vars for usage
export function vars() {
  var value = arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : {};

  return function (Target) {
    var _class, _temp;

    return _temp = _class = function (_React$Component) {
      _inherits(Vars, _React$Component);

      function Vars() {
        _classCallCheck(this, Vars);

        return _possibleConstructorReturn(this, (Vars.__proto__ || Object.getPrototypeOf(Vars)).apply(this, arguments));
      }

      _createClass(Vars, [{
        key: 'getChildContext',
        value: function getChildContext() {
          return {
            glamorCssVars: _extends({}, this.context.glamorCssVars, typeof value === 'function' ? value(this.props) : value)
          };
        }
      }, {
        key: 'render',
        value: function render() {
          return React.createElement(Target, _extends({}, this.props, { vars: this.context.glamorCssVars || (typeof value === 'function' ? value(this.props) : value) }), this.props.children);
        }
      }]);

      return Vars;
    }(React.Component), _class.childContextTypes = {
      glamorCssVars: PropTypes.object
    }, _class.contextTypes = {
      glamorCssVars: PropTypes.object
    }, _temp;
  };
}

var themeIndex = 0;

export function makeTheme() {

  var key = 'data-glamor-theme-' + themeIndex++;

  var fn = function fn(_default) {
    return function (Target) {
      var _class2, _temp2;

      return _temp2 = _class2 = function (_React$Component2) {
        _inherits(Theme, _React$Component2);

        function Theme() {
          _classCallCheck(this, Theme);

          return _possibleConstructorReturn(this, (Theme.__proto__ || Object.getPrototypeOf(Theme)).apply(this, arguments));
        }

        _createClass(Theme, [{
          key: 'render',
          value: function render() {
            return React.createElement(Target, _extends({}, this.props, _defineProperty({}, key, merge.apply(undefined, [typeof _default === 'function' ? _default(this.props) : _default].concat(_toConsumableArray(this.context[key] || []))))));
          }
        }]);

        return Theme;
      }(React.Component), _class2.contextTypes = _defineProperty({}, key, PropTypes.arrayOf(PropTypes.object)), _temp2;
    };
  };

  fn.key = key;
  fn.add = function (_style) {
    return function (Target) {
      var _class3, _temp3;

      return _temp3 = _class3 = function (_React$Component3) {
        _inherits(ThemeOverride, _React$Component3);

        function ThemeOverride() {
          _classCallCheck(this, ThemeOverride);

          return _possibleConstructorReturn(this, (ThemeOverride.__proto__ || Object.getPrototypeOf(ThemeOverride)).apply(this, arguments));
        }

        _createClass(ThemeOverride, [{
          key: 'getChildContext',
          value: function getChildContext() {
            return _defineProperty({}, key, [typeof _style === 'function' ? _style(this.props) : _style].concat(_toConsumableArray(this.context[key] || [])));
          }
        }, {
          key: 'render',
          value: function render() {
            return React.createElement(Target, this.props);
          }
        }]);

        return ThemeOverride;
      }(React.Component), _class3.contextTypes = _defineProperty({}, key, PropTypes.arrayOf(PropTypes.object)), _class3.childContextTypes = _defineProperty({}, key, PropTypes.arrayOf(PropTypes.object)), _temp3;
    };
  };
  return fn;
}

function toStyle(s) {
  return s != null && isLikeRule(s) ? s : style(s);
}

// propMerge will take an arbitrary object "props", filter out glamor data-css-* styles and merge it with "mergeStyle"
// use it for react components composing
export function propMerge(mergeStyle, props) {
  var glamorStyleKeys = Object.keys(props).filter(function (x) {
    return !!/data\-css\-([a-zA-Z0-9]+)/.exec(x);
  });

  // no glamor styles in obj
  if (glamorStyleKeys.length === 0) {
    return _extends({}, props, toStyle(mergeStyle));
  }

  if (glamorStyleKeys.length > 1) {
    console.warn('[glamor] detected multiple data attributes on an element. This may lead to unexpected style because of css insertion order.');

    // just append "mergeStyle" to props, because we dunno in which order glamor styles were added to props
    return _extends({}, props, toStyle(mergeStyle));
  }

  var dataCssKey = glamorStyleKeys[0];
  var cssData = props[dataCssKey];

  var mergedStyles = merge(mergeStyle, _defineProperty({}, dataCssKey, cssData));

  var restProps = Object.assign({}, props);
  delete restProps[dataCssKey];

  return _extends({}, restProps, mergedStyles);
}