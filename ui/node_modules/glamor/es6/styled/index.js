var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();

function _objectWithoutProperties(obj, keys) { var target = {}; for (var i in obj) { if (keys.indexOf(i) >= 0) continue; if (!Object.prototype.hasOwnProperty.call(obj, i)) continue; target[i] = obj[i]; } return target; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _possibleConstructorReturn(self, call) { if (!self) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return call && (typeof call === "object" || typeof call === "function") ? call : self; }

function _inherits(subClass, superClass) { if (typeof superClass !== "function" && superClass !== null) { throw new TypeError("Super expression must either be null or a function, not " + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

import React from 'react';
import { merge, style } from '../';

import { parser, convert } from '../css/raw';
var __val__ = function __val__(x, props) {
  return typeof x === 'function' ? x(props) : x;
};

export function styled(Component, obj) {
  // todo - read style off Component if it exists 
  if (obj) {
    // for when transpiled, and / or by hand 
    return function (_React$Component) {
      _inherits(StyledComponent, _React$Component);

      function StyledComponent() {
        _classCallCheck(this, StyledComponent);

        return _possibleConstructorReturn(this, (StyledComponent.__proto__ || Object.getPrototypeOf(StyledComponent)).apply(this, arguments));
      }

      _createClass(StyledComponent, [{
        key: 'render',
        value: function render() {
          var _props = this.props,
              css = _props.css,
              props = _objectWithoutProperties(_props, ['css']);

          var orig_css = typeof obj === 'function' ? obj(__val__, this.props) : obj;

          return React.createElement(Component, _extends({}, props, this.props.css ? merge(orig_css, css) : style(orig_css))); // todo - isLikeRule 
        }
      }]);

      return StyledComponent;
    }(React.Component);
  }
  return function (strings) {
    for (var _len = arguments.length, values = Array(_len > 1 ? _len - 1 : 0), _key = 1; _key < _len; _key++) {
      values[_key - 1] = arguments[_key];
    }

    // in-browser parsing
    var _parser = parser.apply(undefined, [strings].concat(values)),
        parsed = _parser.parsed,
        stubs = _parser.stubs;

    return function (_React$Component2) {
      _inherits(StyledComponent, _React$Component2);

      function StyledComponent() {
        _classCallCheck(this, StyledComponent);

        return _possibleConstructorReturn(this, (StyledComponent.__proto__ || Object.getPrototypeOf(StyledComponent)).apply(this, arguments));
      }

      _createClass(StyledComponent, [{
        key: 'render',
        value: function render() {
          var _props2 = this.props,
              _props2$className = _props2.className,
              className = _props2$className === undefined ? '' : _props2$className,
              css = _props2.css,
              props = _objectWithoutProperties(_props2, ['className', 'css']);

          return React.createElement(Component, _extends({ className: className + ' ' + merge(convert(parsed, { stubs: stubs, args: [this.props] }), css || {}) }, props));
        }
      }]);

      return StyledComponent;
    }(React.Component);
  };
}

// copied straight from styled-components 

styled.a = styled('a');
styled.abbr = styled('abbr');
styled.address = styled('address');
styled.area = styled('area');
styled.article = styled('article');
styled.aside = styled('aside');
styled.audio = styled('audio');
styled.b = styled('b');
styled.base = styled('base');
styled.bdi = styled('bdi');
styled.bdo = styled('bdo');
styled.big = styled('big');
styled.blockquote = styled('blockquote');
styled.body = styled('body');
styled.br = styled('br');
styled.button = styled('button');
styled.canvas = styled('canvas');
styled.caption = styled('caption');
styled.cite = styled('cite');
styled.code = styled('code');
styled.col = styled('col');
styled.colgroup = styled('colgroup');
styled.data = styled('data');
styled.datalist = styled('datalist');
styled.dd = styled('dd');
styled.del = styled('del');
styled.details = styled('details');
styled.dfn = styled('dfn');
styled.dialog = styled('dialog');
styled.div = styled('div');
styled.dl = styled('dl');
styled.dt = styled('dt');
styled.em = styled('em');
styled.embed = styled('embed');
styled.fieldset = styled('fieldset');
styled.figcaption = styled('figcaption');
styled.figure = styled('figure');
styled.footer = styled('footer');
styled.form = styled('form');
styled.h1 = styled('h1');
styled.h2 = styled('h2');
styled.h3 = styled('h3');
styled.h4 = styled('h4');
styled.h5 = styled('h5');
styled.h6 = styled('h6');
styled.head = styled('head');
styled.header = styled('header');
styled.hgroup = styled('hgroup');
styled.hr = styled('hr');
styled.html = styled('html');
styled.i = styled('i');
styled.iframe = styled('iframe');
styled.img = styled('img');
styled.input = styled('input');
styled.ins = styled('ins');
styled.kbd = styled('kbd');
styled.keygen = styled('keygen');
styled.label = styled('label');
styled.legend = styled('legend');
styled.li = styled('li');
styled.link = styled('link');
styled.main = styled('main');
styled.map = styled('map');
styled.mark = styled('mark');
styled.menu = styled('menu');
styled.menuitem = styled('menuitem');
styled.meta = styled('meta');
styled.meter = styled('meter');
styled.nav = styled('nav');
styled.noscript = styled('noscript');
styled.object = styled('object');
styled.ol = styled('ol');
styled.optgroup = styled('optgroup');
styled.option = styled('option');
styled.output = styled('output');
styled.p = styled('p');
styled.param = styled('param');
styled.picture = styled('picture');
styled.pre = styled('pre');
styled.progress = styled('progress');
styled.q = styled('q');
styled.rp = styled('rp');
styled.rt = styled('rt');
styled.ruby = styled('ruby');
styled.s = styled('s');
styled.samp = styled('samp');
styled.script = styled('script');
styled.section = styled('section');
styled.select = styled('select');
styled.small = styled('small');
styled.source = styled('source');
styled.span = styled('span');
styled.strong = styled('strong');
styled.style = styled('style');
styled.sub = styled('sub');
styled.summary = styled('summary');
styled.sup = styled('sup');
styled.table = styled('table');
styled.tbody = styled('tbody');
styled.td = styled('td');
styled.textarea = styled('textarea');
styled.tfoot = styled('tfoot');
styled.th = styled('th');
styled.thead = styled('thead');
styled.time = styled('time');
styled.title = styled('title');
styled.tr = styled('tr');
styled.track = styled('track');
styled.u = styled('u');
styled.ul = styled('ul');
styled.var = styled('var');
styled.video = styled('video');
styled.wbr = styled('wbr');

// SVG
styled.circle = styled('circle');
styled.clipPath = styled('clipPath');
styled.defs = styled('defs');
styled.ellipse = styled('ellipse');
styled.g = styled('g');
styled.image = styled('image');
styled.line = styled('line');
styled.linearGradient = styled('linearGradient');
styled.mask = styled('mask');
styled.path = styled('path');
styled.pattern = styled('pattern');
styled.polygon = styled('polygon');
styled.polyline = styled('polyline');
styled.radialGradient = styled('radialGradient');
styled.rect = styled('rect');
styled.stop = styled('stop');
styled.svg = styled('svg');
styled.text = styled('text');
styled.tspan = styled('tspan');