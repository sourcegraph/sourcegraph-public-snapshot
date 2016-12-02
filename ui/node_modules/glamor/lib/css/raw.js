'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.conversions = undefined;
exports.css = css;
exports._css = _css;

var _spec = require('./spec.js');

var _index = require('../index.js');

function _toConsumableArray(arr) { if (Array.isArray(arr)) { for (var i = 0, arr2 = Array(arr.length); i < arr.length; i++) { arr2[i] = arr[i]; } return arr2; } else { return Array.from(arr); } }

function _defineProperty(obj, key, value) { if (key in obj) { Object.defineProperty(obj, key, { value: value, enumerable: true, configurable: true, writable: true }); } else { obj[key] = value; } return obj; }

function log(x) {

  console.log(x || '', this); // eslint-disable-line no-console
  return this;
}

function stringify() {
  return JSON.stringify(this, null, ' ');
}

function convert(node, ctx) {
  return conversions[node.type](node, ctx);
}

function toCamel(x) {
  return x.replace(/(\-[a-z])/g, function ($1) {
    return $1.toUpperCase().replace('-', '');
  });
}

var conversions = exports.conversions = {
  StyleSheet: function StyleSheet(node, ctx) {
    return node.rules.map(function (x) {
      return convert(x, ctx);
    });
  },
  MediaRule: function MediaRule(node, ctx) {
    var query = node.media.map(function (x) {
      return convert(x, ctx);
    }).join(',');
    return _defineProperty({}, '@media ' + query, node.rules.map(function (x) {
      return convert(x, ctx);
    }));
  },
  MediaQuery: function MediaQuery(node, ctx) {
    if (node.prefix) {
      return node.prefix + ' ' + node.type + ' ' + node.exprs.map(function (x) {
        return convert(x, ctx);
      }).join(' '); // todo - bug - "and" 
    } else {
      return node.exprs.map(function (x) {
        return convert(x, ctx);
      }).join(' ');
    }
  },
  MediaExpr: function MediaExpr(node, ctx) {
    if (node.value) {
      return '(' + node.feature + ':' + node.value.map(function (x) {
        return convert(x, ctx);
      }) + ')';
    }
    return '(' + node.feature + ')';
  },
  RuleSet: function RuleSet(node, ctx) {
    var selector = node.selectors.map(function (x) {
      return convert(x, ctx);
    }).join('');
    var x = _defineProperty({}, selector, Object.assign.apply(Object, [{}].concat(_toConsumableArray(node.declarations.map(function (x) {
      return convert(x, ctx);
    }))))); // todo - more nesting, accept rules, etc 

    return x;
  },
  Selector: function Selector(node, ctx) {
    return '' + convert(node.left, ctx) + node.combinator + convert(node.right, ctx);
  },
  SimpleSelector: function SimpleSelector(node, ctx) {
    var ret = '' + (node.all ? '*' : node.element !== '*' ? node.element : '') + node.qualifiers.map(function (x) {
      return convert(x, ctx);
    }).join('');
    return ret;
  },
  Contextual: function Contextual() {
    return '&';
  },
  IDSelector: function IDSelector(node, ctx) {
    return node.id;
  },
  ClassSelector: function ClassSelector(node, ctx) {
    return '.' + node['class'];
  },
  PseudoSelector: function PseudoSelector(node, ctx) {
    return ':' + node.value;
  },
  AttributeSelector: function AttributeSelector(node, ctx) {
    return '[' + node.attribute + (node.operator ? node.operator + node.value : '') + ']';
  },
  Function: function Function() {},
  Declaration: function Declaration(node, ctx) {
    // todo - fallbacks
    return _defineProperty({}, toCamel(node.name), convert(node.value, ctx));
  },
  Quantity: function Quantity(node) {
    return node.value + node.unit;
  },
  String: function String(node) {
    return node.value;
  },
  URI: function URI(node) {
    return 'url(' + node.value + ')';
  },
  Ident: function Ident(node) {
    return node.value;
  },
  Hexcolor: function Hexcolor(node) {
    return node.value;
  },
  Expression: function Expression(node, ctx) {
    return convert(node.left, ctx) + (node.operator || ' ') + convert(node.right, ctx);
  },
  Stub: function Stub(node, ctx) {
    return ctx.stubs[node.id];
  },
  Stubs: function Stubs(node, ctx) {
    return node.stubs.map(function (x) {
      return convert(x, ctx);
    });
  }
};

function css(strings) {
  for (var _len = arguments.length, values = Array(_len > 1 ? _len - 1 : 0), _key = 1; _key < _len; _key++) {
    values[_key - 1] = arguments[_key];
  }

  return (0, _index.merge)(_css.apply(undefined, [strings].concat(values)));
}

function _css(strings) {
  for (var _len2 = arguments.length, values = Array(_len2 > 1 ? _len2 - 1 : 0), _key2 = 1; _key2 < _len2; _key2++) {
    values[_key2 - 1] = arguments[_key2];
  }

  var stubs = {},
      ctr = 0;
  strings = strings.reduce(function (arr, x, i) {
    arr.push(x);
    if (values[i] === undefined || values[i] === null) {
      return arr;
    }
    var j = ctr++;
    stubs['spur-' + j] = values[i];
    arr.push('spur-' + j);

    return arr;
  }, []).join('').trim();

  var parsed = (0, _spec.parse)(strings);
  return convert(parsed, { stubs: stubs });
}