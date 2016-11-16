var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();

var _class, _temp;

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

function _objectWithoutProperties(obj, keys) { var target = {}; for (var i in obj) { if (keys.indexOf(i) >= 0) continue; if (!Object.prototype.hasOwnProperty.call(obj, i)) continue; target[i] = obj[i]; } return target; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _possibleConstructorReturn(self, call) { if (!self) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return call && (typeof call === "object" || typeof call === "function") ? call : self; }

function _inherits(subClass, superClass) { if (typeof superClass !== "function" && superClass !== null) { throw new TypeError("Super expression must either be null or a function, not " + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

function _toConsumableArray(arr) { if (Array.isArray(arr)) { for (var i = 0, arr2 = Array(arr.length); i < arr.length; i++) { arr2[i] = arr[i]; } return arr2; } else { return Array.from(arr); } }

import React from 'react';
import * as glamor from './index';

// import shallowCompare from 'react-addons-shallow-compare'
var withKeys = function withKeys() {
  for (var _len = arguments.length, keys = Array(_len), _key = 0; _key < _len; _key++) {
    keys[_key] = arguments[_key];
  }

  return keys.reduce(function (o, k) {
    return o[k] = true, o;
  }, {});
};

// need this list for SSR. 2k gzipped. worth the cost. 
var propNames = withKeys('alignContent', 'alignItems', 'alignSelf', 'alignmentBaseline', 'all', 'animation', 'animationDelay', 'animationDirection', 'animationDuration', 'animationFillMode', 'animationIterationCount', 'animationName', 'animationPlayState', 'animationTimingFunction', 'backfaceVisibility', 'background', 'backgroundAttachment', 'backgroundBlendMode', 'backgroundClip', 'backgroundColor', 'backgroundImage', 'backgroundOrigin', 'backgroundPosition', 'backgroundPositionX', 'backgroundPositionY', 'backgroundRepeat', 'backgroundRepeatX', 'backgroundRepeatY', 'backgroundSize', 'baselineShift', 'border', 'borderBottom', 'borderBottomColor', 'borderBottomLeftRadius', 'borderBottomRightRadius', 'borderBottomStyle', 'borderBottomWidth', 'borderCollapse', 'borderColor', 'borderImage', 'borderImageOutset', 'borderImageRepeat', 'borderImageSlice', 'borderImageSource', 'borderImageWidth', 'borderLeft', 'borderLeftColor', 'borderLeftStyle', 'borderLeftWidth', 'borderRadius', 'borderRight', 'borderRightColor', 'borderRightStyle', 'borderRightWidth', 'borderSpacing', 'borderStyle', 'borderTop', 'borderTopColor', 'borderTopLeftRadius', 'borderTopRightRadius', 'borderTopStyle', 'borderTopWidth', 'borderWidth', 'bottom', 'boxShadow', 'boxSizing', 'breakAfter', 'breakBefore', 'breakInside', 'bufferedRendering', 'captionSide', 'clear', 'clip', 'clipPath', 'clipRule', 'color', 'colorInterpolation', 'colorInterpolationFilters', 'colorRendering', 'columnCount', 'columnFill', 'columnGap', 'columnRule', 'columnRuleColor', 'columnRuleStyle', 'columnRuleWidth', 'columnSpan', 'columnWidth', 'columns', 'contain', 'content', 'counterIncrement', 'counterReset', 'cursor', 'cx', 'cy', 'd', 'direction', 'display', 'dominantBaseline', 'emptyCells', 'fill', 'fillOpacity', 'fillRule', 'filter', 'flex', 'flexBasis', 'flexDirection', 'flexFlow', 'flexGrow', 'flexShrink', 'flexWrap', 'float', 'floodColor', 'floodOpacity', 'font', 'fontFamily', 'fontFeatureSettings', 'fontKerning', 'fontSize', 'fontStretch', 'fontStyle', 'fontVariant', 'fontVariantCaps', 'fontVariantLigatures', 'fontVariantNumeric', 'fontWeight', 'height', 'imageRendering', 'isolation', 'justifyContent', 'left', 'letterSpacing', 'lightingColor', 'lineHeight', 'listStyle', 'listStyleImage', 'listStylePosition', 'listStyleType', 'margin', 'marginBottom', 'marginLeft', 'marginRight', 'marginTop', 'marker', 'markerEnd', 'markerMid', 'markerStart', 'mask', 'maskType', 'maxHeight', 'maxWidth', 'maxZoom', 'minHeight', 'minWidth', 'minZoom', 'mixBlendMode', 'motion', 'motionOffset', 'motionPath', 'motionRotation', 'objectFit', 'objectPosition', 'opacity', 'order', 'orientation', 'orphans', 'outline', 'outlineColor', 'outlineOffset', 'outlineStyle', 'outlineWidth', 'overflow', 'overflowWrap', 'overflowX', 'overflowY', 'padding', 'paddingBottom', 'paddingLeft', 'paddingRight', 'paddingTop', 'page', 'pageBreakAfter', 'pageBreakBefore', 'pageBreakInside', 'paintOrder', 'perspective', 'perspectiveOrigin', 'pointerEvents', 'position', 'quotes', 'r', 'resize', 'right', 'rx', 'ry', 'shapeImageThreshold', 'shapeMargin', 'shapeOutside', 'shapeRendering', 'size', 'speak', 'src', 'stopColor', 'stopOpacity', 'stroke', 'strokeDasharray', 'strokeDashoffset', 'strokeLinecap', 'strokeLinejoin', 'strokeMiterlimit', 'strokeOpacity', 'strokeWidth', 'tabSize', 'tableLayout', 'textAlign', 'textAlignLast', 'textAnchor', 'textCombineUpright', 'textDecoration', 'textIndent', 'textOrientation', 'textOverflow', 'textRendering', 'textShadow', 'textTransform', 'top', 'touchAction', 'transform', 'transformOrigin', 'transformStyle', 'transition', 'transitionDelay', 'transitionDuration', 'transitionProperty', 'transitionTimingFunction', 'unicodeBidi', 'unicodeRange', 'userZoom', 'vectorEffect', 'verticalAlign', 'visibility', 'appRegion', 'appearance', 'backgroundClip', 'backgroundOrigin', 'borderAfter', 'borderAfterColor', 'borderAfterStyle', 'borderAfterWidth', 'borderBefore', 'borderBeforeColor', 'borderBeforeStyle', 'borderBeforeWidth', 'borderEnd', 'borderEndColor', 'borderEndStyle', 'borderEndWidth', 'borderHorizontalSpacing', 'borderImage', 'borderStart', 'borderStartColor', 'borderStartStyle', 'borderStartWidth', 'borderVerticalSpacing', 'boxAlign', 'boxDecorationBreak', 'boxDirection', 'boxFlex', 'boxFlexGroup', 'boxLines', 'boxOrdinalGroup', 'boxOrient', 'boxPack', 'boxReflect', 'clipPath', 'columnBreakAfter', 'columnBreakBefore', 'columnBreakInside', 'filter', 'fontSizeDelta', 'fontSmoothing', 'highlight', 'hyphenateCharacter', 'lineBreak', 'lineClamp', 'locale', 'logicalHeight', 'logicalWidth', 'marginAfter', 'marginAfterCollapse', 'marginBefore', 'marginBeforeCollapse', 'marginBottomCollapse', 'marginCollapse', 'marginEnd', 'marginStart', 'marginTopCollapse', 'mask', 'maskBoxImage', 'maskBoxImageOutset', 'maskBoxImageRepeat', 'maskBoxImageSlice', 'maskBoxImageSource', 'maskBoxImageWidth', 'maskClip', 'maskComposite', 'maskImage', 'maskOrigin', 'maskPosition', 'maskPositionX', 'maskPositionY', 'maskRepeat', 'maskRepeatX', 'maskRepeatY', 'maskSize', 'maxLogicalHeight', 'maxLogicalWidth', 'minLogicalHeight', 'minLogicalWidth', 'paddingAfter', 'paddingBefore', 'paddingEnd', 'paddingStart', 'perspectiveOriginX', 'perspectiveOriginY', 'printColorAdjust', 'rtlOrdering', 'rubyPosition', 'tapHighlightColor', 'textCombine', 'textDecorationsInEffect', 'textEmphasis', 'textEmphasisColor', 'textEmphasisPosition', 'textEmphasisStyle', 'textFillColor', 'textOrientation', 'textSecurity', 'textStroke', 'textStrokeColor', 'textStrokeWidth', 'transformOriginX', 'transformOriginY', 'transformOriginZ', 'userDrag', 'userModify', 'userSelect', 'writingMode', 'whiteSpace', 'widows', 'width', 'willChange', 'wordBreak', 'wordSpacing', 'wordWrap', 'writingMode', 'x', 'y', 'zIndex', 'zoom');

var pseudos = withKeys(
// pseudoclasses
'active', 'any', 'checked', 'disabled', 'empty', 'enabled', '_default', 'first', 'firstChild', 'firstOfType', 'fullscreen', 'focus', 'hover', 'indeterminate', 'inRange', 'invalid', 'lastChild', 'lastOfType', 'left', 'link', 'onlyChild', 'onlyOfType', 'optional', 'outOfRange', 'readOnly', 'readWrite', 'required', 'right', 'root', 'scope', 'target', 'valid', 'visited',
// pseudoelements
'after', 'before', 'firstLetter', 'firstLine', 'selection', 'backdrop', 'placeholder');

var parameterizedPseudos = withKeys('dir', 'lang', 'not', 'nthChild', 'nthLastChild', 'nthLastOfType', 'nthOfType');

// const STYLE_PROP_NAMES = propNames.reduce((styles, key) => {
//   styles[key] = true
//   return styles
// }, { label: true })

// /^(webkit|moz|ms)([A-Za-z]+)/
var prefixCache = {};
function prefixed(key) {
  if (prefixCache.hasOwnProperty(key)) {
    return prefixCache[key];
  }
  var m = /^(webkit|moz|ms|o){1}([A-Z][A-Za-z]+)/.exec(key),
      subKey = void 0;
  if (m) {
    subKey = m[2];
    subKey = subKey.charAt(0).toLowerCase() + subKey.slice(1);
  }
  prefixCache[key] = subKey;
  return subKey;
}

function isHandler(key) {
  return !!/^on[A-Z]/.exec(key);
}

var splitStyles = function splitStyles(combinedProps) {
  var props = {},
      gStyle = [],
      style = {};
  Object.keys(combinedProps).forEach(function (key) {

    if (propNames[key]) {
      style[key] = combinedProps[key];
    } else if (prefixed(key) && propNames[prefixed(key)]) {
      style[key] = combinedProps[key];
    } else if (key === 'css') {
      Object.assign(style, combinedProps[key]);
    } else if (pseudos[key] >= 0) {
      gStyle.push(glamor[key](combinedProps[key]));
    } else if (parameterizedPseudos[key] || key === 'media' || key === 'select') {
      gStyle.push(glamor[key].apply(glamor, _toConsumableArray(combinedProps[key])));
    } else if (key === 'merge' || key === 'compose') {
      if (Array.isArray(combinedProps[key])) {
        gStyle.splice.apply(gStyle, [gStyle.length, 0].concat(_toConsumableArray(combinedProps[key])));
      } else {
        gStyle.splice(gStyle.length, 0, combinedProps[key]);
      }
    } else if (isHandler(key)) {
      props[key] = combinedProps[key];
    } else if (key === 'props') {
      Object.assign(props, combinedProps[key]);
    } else if (key === 'style' || key === 'className' || key === 'children') {
      props[key] = combinedProps[key];
    } else {
      // console.warn('irregular key ' + key)    //eslint-disable-line no-console
      props[key] = combinedProps[key];
    }
  });
  return _extends({}, gStyle.length > 0 ? glamor.merge.apply(glamor, [style].concat(gStyle)) : Object.keys(style).length > 0 ? glamor.style(style) : {}, props);
};

export var View = (_temp = _class = function (_React$Component) {
  _inherits(View, _React$Component);

  function View() {
    _classCallCheck(this, View);

    return _possibleConstructorReturn(this, (View.__proto__ || Object.getPrototypeOf(View)).apply(this, arguments));
  }

  _createClass(View, [{
    key: 'render',
    value: function render() {
      var _props = this.props,
          Component = _props.component,
          props = _objectWithoutProperties(_props, ['component']);

      return React.createElement(Component, splitStyles(props));
    }
  }]);

  return View;
}(React.Component), _class.defaultProps = {
  component: 'div'
}, _temp);

export var Block = function (_React$Component2) {
  _inherits(Block, _React$Component2);

  function Block() {
    _classCallCheck(this, Block);

    return _possibleConstructorReturn(this, (Block.__proto__ || Object.getPrototypeOf(Block)).apply(this, arguments));
  }

  _createClass(Block, [{
    key: 'render',
    value: function render() {
      return React.createElement(View, _extends({ display: 'block' }, this.props));
    }
  }]);

  return Block;
}(React.Component);

export var InlineBlock = function (_React$Component3) {
  _inherits(InlineBlock, _React$Component3);

  function InlineBlock() {
    _classCallCheck(this, InlineBlock);

    return _possibleConstructorReturn(this, (InlineBlock.__proto__ || Object.getPrototypeOf(InlineBlock)).apply(this, arguments));
  }

  _createClass(InlineBlock, [{
    key: 'render',
    value: function render() {
      return React.createElement(View, _extends({ display: 'inline-block' }, this.props));
    }
  }]);

  return InlineBlock;
}(React.Component);

export var Flex = function (_React$Component4) {
  _inherits(Flex, _React$Component4);

  function Flex() {
    _classCallCheck(this, Flex);

    return _possibleConstructorReturn(this, (Flex.__proto__ || Object.getPrototypeOf(Flex)).apply(this, arguments));
  }

  _createClass(Flex, [{
    key: 'render',
    value: function render() {
      return React.createElement(View, _extends({ display: 'flex' }, this.props));
    }
  }]);

  return Flex;
}(React.Component);

export var Row = function (_React$Component5) {
  _inherits(Row, _React$Component5);

  function Row() {
    _classCallCheck(this, Row);

    return _possibleConstructorReturn(this, (Row.__proto__ || Object.getPrototypeOf(Row)).apply(this, arguments));
  }

  _createClass(Row, [{
    key: 'render',
    value: function render() {
      return React.createElement(Flex, _extends({ flexDirection: 'row' }, this.props));
    }
  }]);

  return Row;
}(React.Component);

export var Column = function (_React$Component6) {
  _inherits(Column, _React$Component6);

  function Column() {
    _classCallCheck(this, Column);

    return _possibleConstructorReturn(this, (Column.__proto__ || Object.getPrototypeOf(Column)).apply(this, arguments));
  }

  _createClass(Column, [{
    key: 'render',
    value: function render() {
      return React.createElement(Flex, _extends({ flexDirection: 'column' }, this.props));
    }
  }]);

  return Column;
}(React.Component);