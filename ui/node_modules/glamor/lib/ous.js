'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.clearfix = exports.pullLeft = exports.pullRight = exports.maxFullWidth = exports.fullWidth = exports.base = exports.labelBody = exports.primary = exports.button = exports.twoThirds = exports.oneThird = exports.half = exports.columns = exports.row = exports.container = undefined;

var _index = require('./index.js');

function _toConsumableArray(arr) { if (Array.isArray(arr)) { for (var i = 0, arr2 = Array(arr.length); i < arr.length; i++) { arr2[i] = arr[i]; } return arr2; } else { return Array.from(arr); } }

function log() {
  console.log(this); // eslint-disable-line
  return this;
}

(0, _index.insertRule)('html { font-size: 62.5% }');

var container = exports.container = (0, _index.merge)({
  position: 'relative',
  width: '100%',
  maxWidth: 960,
  margin: '0 auto',
  padding: '0 20px',
  boxSizing: 'border-box'
}, (0, _index.media)('(min-width: 400px)', {
  width: '85%',
  padding: 0
}), (0, _index.media)('(min-width: 550px)', {
  width: '80%'
}), (0, _index.after)({
  content: '""',
  display: 'table',
  clear: 'both'
}));

var widths = [null, // 0 - don't even 
'4.66666666667%', // 1
'13.3333333333%', // 2 
'22%', // 3
'30.6666666667%', // 4
'39.3333333333%', // 5
'48%', // 6
'56.6666666667%', // 7
'65.3333333333%', // 8
'74.0%', // 9
'82.6666666667%', // 10
'91.3333333333%', // 11
'100%' // 12 - don't forget the margin 
];

var fractionalWidths = {
  half: '48%',
  oneThird: '30.6666666667%',
  twoThirds: '65.3333333333%'
};

var offsets = [null, // 0 - don't even 
'8.66666666667%', // 1
'17.3333333333%', // 2
'26%', // 3
'34.6666666667%', // 4 
'43.3333333333%', // 5 
'52%', // 6 
'60.6666666667%', // 7
'69.3333333333%', // 8
'78.0%', // 9
'86.6666666667%', // 10
'95.3333333333%' // 11  
];

var fractionalOffsets = {
  half: '34.6666666667%',
  oneThird: '69.3333333333%',
  twoThirds: '52%'
};

var row = exports.row = (0, _index.after)({
  content: '""',
  display: 'table',
  clear: 'both'
});

var columns = exports.columns = function columns(n, offset) {
  return (0, _index.merge)({
    width: '100%',
    float: 'left',
    boxSizing: 'border-box'
  }, (0, _index.media)('(min-width: 550px)', {
    marginLeft: '4%'
  }, (0, _index.firstChild)({
    marginLeft: 0
  }), {
    width: typeof n === 'number' ? widths[n] : fractionalWidths[n]
  }, n === 12 ? {
    marginLeft: 0
  } : {}, offset ? {
    marginLeft: typeof n === 'number' ? offsets[offset] : fractionalOffsets[offset]
  } : {}));
};

var half = exports.half = function half(offset) {
  return columns('half', offset);
};
var oneThird = exports.oneThird = function oneThird(offset) {
  return columns('oneThird', offset);
};
var twoThirds = exports.twoThirds = function twoThirds(offset) {
  return columns('twoThirds', offset);
};

var button = exports.button = (0, _index.style)({ label: 'button' });
var primary = exports.primary = (0, _index.style)({ label: 'primary' });

var labelBody = exports.labelBody = (0, _index.style)({ label: 'labelBody' });

var buttonId = '[' + Object.keys(button)[0] + ']'; //`[data-css-${idFor(button)}]`
var primaryId = '[' + Object.keys(primary)[0] + ']'; //`[data-css-${idFor(primary)}]`
var labelBodyId = '[' + Object.keys(labelBody)[0] + ']'; //`[data-css-${idFor(labelBody)}]`

var base = exports.base = _index.merge.apply(undefined, [(0, _index.style)({
  fontSize: '1.5em',
  lineHeight: '1.6',
  fontWeight: 400,
  fontFamily: '"Raleway", "HelveticaNeue", "Helvetica Neue", Helvetica, Arial, sans-serif',
  color: '#222'
}),

// typography
(0, _index.select)(' h1, h2, h3, h4, h5, h6', {
  marginTop: 0,
  marginBottom: '2rem',
  fontWeight: 300
})].concat(_toConsumableArray([[{ fontSize: '4.0rem', lineHeight: 1.2, letterSpacing: '-.1rem' }, 5.0], [{ fontSize: '3.6rem', lineHeight: 1.25, letterSpacing: '-.1rem' }, 4.2], [{ fontSize: '3.0rem', lineHeight: 1.3, letterSpacing: '-.1rem' }, 3.6], [{ fontSize: '2.4rem', lineHeight: 1.35, letterSpacing: '-.08rem' }, 3.0], [{ fontSize: '1.8rem', lineHeight: 1.5, letterSpacing: '-.05rem' }, 2.4], [{ fontSize: '1.5rem', lineHeight: 1.6, letterSpacing: 0 }, 1.5]].map(function (x, i) {
  return (0, _index.merge)((0, _index.select)(' h' + (i + 1), x[0]), (0, _index.media)('(min-width: 550px)', (0, _index.select)(' h' + (i + 1), { // larger than phablet 
    fontSize: x[1] + 'rem'
  })));
})), [(0, _index.select)(' p', {
  marginTop: 0
}),

// links 
(0, _index.select)(' a', {
  color: '#1EAEDB'
}), (0, _index.select)(' a:hover', {
  color: '#0FA0CE'
}),

// buttons 
(0, _index.select)(' button, input[type="submit"], input[type="reset"], input[type="button"], ' + buttonId, {
  display: 'inline-block',
  height: 38,
  padding: '0 30px',
  color: '#555',
  textAlign: 'center',
  fontSize: 11,
  fontWeight: 600,
  lineHeight: '38px',
  letterSpacing: '.1rem',
  textTransform: 'uppercase',
  textDecoration: 'none',
  whiteSpace: 'nowrap',
  backgroundColor: 'transparent',
  borderRadius: 4,
  border: '1px solid #bbb',
  cursor: 'pointer',
  boxSizing: 'border-box'
}), (0, _index.select)(' ' + buttonId + ':hover, button:hover, input[type="submit"]:hover, \n    input[type="reset"]:hover, input[type="button"]:hover, \n    ' + buttonId + ':focus, button:focus, input[type="submit"]:focus, \n    input[type="reset"]:focus, input[type="button"]:focus', {
  color: '#333',
  borderColor: '#888',
  outline: 0
}), (0, _index.select)(' ' + buttonId + primaryId + ', button' + primaryId + ', input[type="submit"]' + primaryId + ', input[type="reset"]' + primaryId + ', input[type="button"]' + primaryId, {
  color: '#FFF',
  backgroundColor: '#33C3F0',
  borderColor: '#33C3F0'
}), (0, _index.select)(' ' + buttonId + primaryId + ':hover, button' + primaryId + ':hover, input[type="submit"]' + primaryId + ':hover, \n    input[type="reset"]' + primaryId + ':hover, input[type="button"]' + primaryId + ':hover,\n     ' + buttonId + primaryId + ':focus, button' + primaryId + ':focus, input[type="submit"]' + primaryId + ':focus, \n    input[type="reset"]' + primaryId + ':focus, input[type="button"]' + primaryId + ':focus', {
  color: '#FFF',
  backgroundColor: '#1EAEDB',
  borderColor: '#1EAEDB'
}),

// forms 
(0, _index.select)(' input[type="email"], input[type="number"], input[type="search"], input[type="text"], \n    input[type="tel"], input[type="url"], input[type="password"], textarea, select', {
  height: '38px',
  padding: '6px 10px', /* The 6px vertically centers text on FF, ignored by Webkit */
  backgroundColor: '#fff',
  border: '1px solid #D1D1D1',
  borderRadius: '4px',
  boxShadow: 'none',
  boxSizing: 'border-box'
}), (0, _index.select)(' input[type="email"], input[type="number"], input[type="search"], input[type="text"],\n    input[type="tel"], input[type="url"], input[type="password"], textarea', {
  WebkitAppearance: 'none',
  MozAppearance: 'none',
  appearance: 'none'
}), (0, _index.select)(' textarea', {
  minHeight: 65,
  paddingTop: 6,
  paddingBottom: 6
}), (0, _index.select)(' input[type="email"]:focus, input[type="number"]:focus, input[type="search"]:focus,\n    input[type="text"]:focus, input[type="tel"]:focus, input[type="url"]:focus, input[type="password"]:focus,\n    textarea:focus, select:focus', {
  border: '1px solid #33C3F0',
  outline: 0
}), (0, _index.select)(' label, legend', {
  display: 'block',
  marginBottom: '.5rem',
  fontWeight: 600
}), (0, _index.select)(' fieldset', {
  padding: 0,
  borderWidth: 0
}), (0, _index.select)(' input[type="checkbox"], input[type="radio"]', {
  display: 'inline'
}), (0, _index.select)(' label > ' + labelBodyId, {
  display: 'inline-block',
  marginLeft: '.5rem',
  fontWeight: 'normal'
}),

// lists 
(0, _index.select)(' ul', {
  listStyle: 'circle inside'
}), (0, _index.select)(' ol', {
  listStyle: 'decimal inside'
}), (0, _index.select)(' ol, ul', {
  paddingLeft: 0,
  marginTop: 0
}), (0, _index.select)(' ul ul, ul ol, ol ul, ol ol', {
  margin: '1.5rem 0 1.5rem 3rem',
  fontSize: '90%'
}), (0, _index.select)(' li', {
  marginBottom: '1rem'
}),

// code 
(0, _index.select)(' code', {
  padding: '.2rem .5rem',
  margin: '0 .2rem',
  fontSize: '90%',
  whiteSpace: 'nowrap',
  background: '#F1F1F1',
  border: '1px solid #E1E1E1',
  borderRadius: '4px'
}), (0, _index.select)(' pre > code', {
  display: 'block',
  padding: '1rem 1.5 rem',
  whiteSpace: 'pre'
}),

// tables 
(0, _index.select)(' th, td', {
  padding: '12px 15px',
  textAlign: 'left',
  borderBottom: '1px solid #E1E1E1'
}), (0, _index.select)(' th:first-child, td:first-child', {
  paddingLeft: 0
}), (0, _index.select)(' th:last-child, td:last-child', {
  paddingRight: 0
}),

// spacing
(0, _index.select)(' button', {
  marginBottom: '1rem'
}), (0, _index.select)(' input, textarea, select, fieldset', {
  marginBottom: '1.5rem'
}), (0, _index.select)(' pre, blockquote, dl, figure, table, p, ul, ol, form', {
  marginBottom: '2.5rem'
}),

// misc 
(0, _index.select)(' hr', {
  marginTop: '3rem',
  marginBottom: '3.5rem',
  borderWidth: 0,
  borderTop: '1px solid #E1E1E1'
})]));

// utilities 
var fullWidth = exports.fullWidth = (0, _index.style)({
  width: '100%',
  boxSizing: 'border-box'
});

var maxFullWidth = exports.maxFullWidth = (0, _index.style)({
  maxWidth: '100%',
  boxSizing: 'border-box'
});

var pullRight = exports.pullRight = (0, _index.style)({
  float: 'right'
});

var pullLeft = exports.pullLeft = (0, _index.style)({
  float: 'left'
});

var clearfix = exports.clearfix = (0, _index.style)({
  content: '""',
  display: 'table',
  clear: 'both'
});