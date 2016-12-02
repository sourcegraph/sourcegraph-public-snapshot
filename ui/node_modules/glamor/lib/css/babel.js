'use strict';

// a babel plugin that strips out the tagged literal syntax,
// and replaces with a json form. everybody wins!
// we can do this because we control the ast
// and there's a corresponding json representation for every kv pair / nesting form

// replace interpolations with stubs
// parse to get intermediate form
// print back, replacing stubs with interpolations
// tada?

// todo - custom function instead of merge
var _require = require('./spec.js'),
    parse = _require.parse;

function convert(node, ctx, interpolated) {
  if (interpolated && node.type === 'Stub') {
    return '${' + conversions[node.type](node, ctx, interpolated) + '}';
  }
  return conversions[node.type](node, ctx, interpolated);
}

function toCamel(x) {
  return x.replace(/(\-[a-z])/g, function ($1) {
    return $1.toUpperCase().replace('-', '');
  });
}

var conversions = {
  StyleSheet: function StyleSheet(node, ctx) {
    return '[ ' + node.rules.map(function (x) {
      return convert(x, ctx);
    }).join(', ') + ' ]';
  },
  MediaRule: function MediaRule(node, ctx) {
    var query = node.media.map(function (x) {
      return convert(x, ctx, true);
    }).join(',');
    var mq = query.indexOf('${') >= 0 ? '[`@media ' + query + '`]' : '\'@media ' + query + '\'';
    return '{ ' + mq + ': [ ' + node.rules.map(function (x) {
      return convert(x, ctx);
    }).join(', ') + ' ] }';
  },
  MediaQuery: function MediaQuery(node, ctx) {
    if (node.prefix) {
      return node.prefix + ' ' + node.type + ' ' + node.exprs.map(function (x) {
        return convert(x, ctx, true);
      }).join(' ');
    } else {
      return node.exprs.map(function (x) {
        return convert(x, ctx, true);
      }).join(' ');
    }
  },
  MediaExpr: function MediaExpr(node, ctx) {
    if (node.value) {
      return '(' + node.feature + ':' + node.value.map(function (x) {
        return convert(x, ctx, true);
      }) + ')';
    }
    return '(' + node.feature + ')';
  },
  RuleSet: function RuleSet(node, ctx) {
    var selector = node.selectors.map(function (x) {
      return convert(x, ctx, true);
    }).join('');
    var declarations = node.declarations.map(function (x) {
      return convert(x, ctx);
    });
    var declStr = declarations.length > 1 ? '[ ' + declarations.join(', ') + ' ]' : declarations;
    if (selector.indexOf('${') >= 0) {
      selector = '[`' + selector + '`]';
    } else {
      selector = '\'' + selector + '\'';
    }
    var x = '{ ' + selector + ': ' + declStr + ' }'; // todo - more nesting, accept rules, etc 

    return x;
  },
  Selector: function Selector(node, ctx) {
    return '' + convert(node.left, ctx, true) + node.combinator + convert(node.right, ctx, true);
  },
  SimpleSelector: function SimpleSelector(node, ctx) {
    var ret = '' + (node.all ? '*' : node.element !== '*' ? node.element : '') + node.qualifiers.map(function (x) {
      return convert(x, ctx, true);
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
  Function: function Function(node, ctx) {
    return node.name + '(' + convert(node.params, ctx) + ')';
  },
  Declaration: function Declaration(node, ctx) {
    // todo - fallbacks
    var val = convert(node.value, ctx, true);
    var icky = false;
    ['${', '\'', '"'].forEach(function (x) {
      icky = icky || (val + '').indexOf(x) >= 0;
    });
    val = icky ? '`' + val + '`' : '\'' + val + '\'';
    if (node.value.type === 'Stub') {
      val = convert(node.value, ctx);
    }
    return '{ \'' + (node.name.indexOf('--') === 0 ? node.name : toCamel(node.name)) + '\': ' + val + ' }'; // todo - numbers 
  },
  Quantity: function Quantity(node, ctx) {
    return (node.value.type === 'Stub' ? convert(node.value, ctx) : node.value) + node.unit;
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
    return convert(node.left, ctx, true) + (node.operator || ' ') + convert(node.right, ctx, true);
  },
  Stub: function Stub(node, ctx) {
    if (ctx.withProps) {
      return 'val(' + ctx.stubs[node.id] + ', props)';
    }
    return ctx.stubs[node.id];
  },
  Stubs: function Stubs(node, ctx) {
    return node.stubs.map(function (x) {
      return convert(x, ctx);
    });
  }
};

function parser(path) {
  var code = path.hub.file.code;
  var stubs = path.node.quasi.expressions.map(function (x) {
    return code.substring(x.start, x.end);
  });
  var stubCtx = stubs.reduce(function (o, stub, i) {
    return o['spur-' + i] = stub, o;
  }, {});
  var ctr = 0;
  var strs = path.node.quasi.quasis.map(function (x) {
    return x.value.cooked;
  });
  var src = strs.reduce(function (arr, str, i) {
    arr.push(str);
    if (i !== stubs.length) {
      arr.push('spur-' + ctr++);
    }
    return arr;
  }, []).join('');
  var parsed = parse(src.trim());
  return { parsed: parsed, stubs: stubCtx };
}

module.exports = {
  visitor: {
    TaggedTemplateExpression: function TaggedTemplateExpression(path) {
      var tag = path.node.tag;

      var code = path.hub.file.code;

      if (tag.name === 'css') {
        var _parser = parser(path),
            parsed = _parser.parsed,
            stubs = _parser.stubs;

        var newSrc = 'css(' + convert(parsed, { stubs: stubs }) + ')';
        path.replaceWithSourceString(newSrc);
      } else if (tag.type === 'CallExpression' && tag.callee.name === 'styled') {
        var _parser2 = parser(path),
            _parsed = _parser2.parsed,
            _stubs = _parser2.stubs;

        var _newSrc = 'styled(' + code.substring(tag.arguments[0].start, tag.arguments[0].end) + ', (val, props) => (' + convert(_parsed, { stubs: _stubs, withProps: true }) + '))';
        path.replaceWithSourceString(_newSrc);
      } else if (tag.type === 'MemberExpression' && tag.object.name === 'styled') {
        var _parser3 = parser(path),
            _parsed2 = _parser3.parsed,
            _stubs2 = _parser3.stubs;

        var _newSrc2 = 'styled(\'' + tag.property.name + '\', (val, props) => (' + convert(_parsed2, { stubs: _stubs2, withProps: true }) + '))';
        path.replaceWithSourceString(_newSrc2);
      }
    }
  }
};