function _toConsumableArray(arr) { if (Array.isArray(arr)) { for (var i = 0, arr2 = Array(arr.length); i < arr.length; i++) { arr2[i] = arr[i]; } return arr2; } else { return Array.from(arr); } }

import { style, media, merge, firstChild, after, select, hover, idFor, insertRule } from './index.js';

function log() {
  console.log(this); // eslint-disable-line
  return this;
}

insertRule('html { font-size: 62.5% }');

export var container = merge({
  position: 'relative',
  width: '100%',
  maxWidth: 960,
  margin: '0 auto',
  padding: '0 20px',
  boxSizing: 'border-box'
}, media('(min-width: 400px)', {
  width: '85%',
  padding: 0
}), media('(min-width: 550px)', {
  width: '80%'
}), after({
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

export var row = after({
  content: '""',
  display: 'table',
  clear: 'both'
});

export var columns = function columns(n, offset) {
  return merge({
    width: '100%',
    float: 'left',
    boxSizing: 'border-box'
  }, media('(min-width: 550px)', {
    marginLeft: '4%'
  }, firstChild({
    marginLeft: 0
  }), {
    width: typeof n === 'number' ? widths[n] : fractionalWidths[n]
  }, n === 12 ? {
    marginLeft: 0
  } : {}, offset ? {
    marginLeft: typeof n === 'number' ? offsets[offset] : fractionalOffsets[offset]
  } : {}));
};

export var half = function half(offset) {
  return columns('half', offset);
};
export var oneThird = function oneThird(offset) {
  return columns('oneThird', offset);
};
export var twoThirds = function twoThirds(offset) {
  return columns('twoThirds', offset);
};

export var button = style({ label: 'button' });
export var primary = style({ label: 'primary' });

export var labelBody = style({ label: 'labelBody' });

var buttonId = '[' + Object.keys(button)[0] + ']'; //`[data-css-${idFor(button)}]`
var primaryId = '[' + Object.keys(primary)[0] + ']'; //`[data-css-${idFor(primary)}]`
var labelBodyId = '[' + Object.keys(labelBody)[0] + ']'; //`[data-css-${idFor(labelBody)}]`

export var base = merge.apply(undefined, [style({
  fontSize: '1.5em',
  lineHeight: '1.6',
  fontWeight: 400,
  fontFamily: '"Raleway", "HelveticaNeue", "Helvetica Neue", Helvetica, Arial, sans-serif',
  color: '#222'
}),

// typography
select(' h1, h2, h3, h4, h5, h6', {
  marginTop: 0,
  marginBottom: '2rem',
  fontWeight: 300
})].concat(_toConsumableArray([[{ fontSize: '4.0rem', lineHeight: 1.2, letterSpacing: '-.1rem' }, 5.0], [{ fontSize: '3.6rem', lineHeight: 1.25, letterSpacing: '-.1rem' }, 4.2], [{ fontSize: '3.0rem', lineHeight: 1.3, letterSpacing: '-.1rem' }, 3.6], [{ fontSize: '2.4rem', lineHeight: 1.35, letterSpacing: '-.08rem' }, 3.0], [{ fontSize: '1.8rem', lineHeight: 1.5, letterSpacing: '-.05rem' }, 2.4], [{ fontSize: '1.5rem', lineHeight: 1.6, letterSpacing: 0 }, 1.5]].map(function (x, i) {
  return merge(select(' h' + (i + 1), x[0]), media('(min-width: 550px)', select(' h' + (i + 1), { // larger than phablet 
    fontSize: x[1] + 'rem'
  })));
})), [select(' p', {
  marginTop: 0
}),

// links 
select(' a', {
  color: '#1EAEDB'
}), select(' a:hover', {
  color: '#0FA0CE'
}),

// buttons 
select(' button, input[type="submit"], input[type="reset"], input[type="button"], ' + buttonId, {
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
}), select(' ' + buttonId + ':hover, button:hover, input[type="submit"]:hover, \n    input[type="reset"]:hover, input[type="button"]:hover, \n    ' + buttonId + ':focus, button:focus, input[type="submit"]:focus, \n    input[type="reset"]:focus, input[type="button"]:focus', {
  color: '#333',
  borderColor: '#888',
  outline: 0
}), select(' ' + buttonId + primaryId + ', button' + primaryId + ', input[type="submit"]' + primaryId + ', input[type="reset"]' + primaryId + ', input[type="button"]' + primaryId, {
  color: '#FFF',
  backgroundColor: '#33C3F0',
  borderColor: '#33C3F0'
}), select(' ' + buttonId + primaryId + ':hover, button' + primaryId + ':hover, input[type="submit"]' + primaryId + ':hover, \n    input[type="reset"]' + primaryId + ':hover, input[type="button"]' + primaryId + ':hover,\n     ' + buttonId + primaryId + ':focus, button' + primaryId + ':focus, input[type="submit"]' + primaryId + ':focus, \n    input[type="reset"]' + primaryId + ':focus, input[type="button"]' + primaryId + ':focus', {
  color: '#FFF',
  backgroundColor: '#1EAEDB',
  borderColor: '#1EAEDB'
}),

// forms 
select(' input[type="email"], input[type="number"], input[type="search"], input[type="text"], \n    input[type="tel"], input[type="url"], input[type="password"], textarea, select', {
  height: '38px',
  padding: '6px 10px', /* The 6px vertically centers text on FF, ignored by Webkit */
  backgroundColor: '#fff',
  border: '1px solid #D1D1D1',
  borderRadius: '4px',
  boxShadow: 'none',
  boxSizing: 'border-box'
}), select(' input[type="email"], input[type="number"], input[type="search"], input[type="text"],\n    input[type="tel"], input[type="url"], input[type="password"], textarea', {
  WebkitAppearance: 'none',
  MozAppearance: 'none',
  appearance: 'none'
}), select(' textarea', {
  minHeight: 65,
  paddingTop: 6,
  paddingBottom: 6
}), select(' input[type="email"]:focus, input[type="number"]:focus, input[type="search"]:focus,\n    input[type="text"]:focus, input[type="tel"]:focus, input[type="url"]:focus, input[type="password"]:focus,\n    textarea:focus, select:focus', {
  border: '1px solid #33C3F0',
  outline: 0
}), select(' label, legend', {
  display: 'block',
  marginBottom: '.5rem',
  fontWeight: 600
}), select(' fieldset', {
  padding: 0,
  borderWidth: 0
}), select(' input[type="checkbox"], input[type="radio"]', {
  display: 'inline'
}), select(' label > ' + labelBodyId, {
  display: 'inline-block',
  marginLeft: '.5rem',
  fontWeight: 'normal'
}),

// lists 
select(' ul', {
  listStyle: 'circle inside'
}), select(' ol', {
  listStyle: 'decimal inside'
}), select(' ol, ul', {
  paddingLeft: 0,
  marginTop: 0
}), select(' ul ul, ul ol, ol ul, ol ol', {
  margin: '1.5rem 0 1.5rem 3rem',
  fontSize: '90%'
}), select(' li', {
  marginBottom: '1rem'
}),

// code 
select(' code', {
  padding: '.2rem .5rem',
  margin: '0 .2rem',
  fontSize: '90%',
  whiteSpace: 'nowrap',
  background: '#F1F1F1',
  border: '1px solid #E1E1E1',
  borderRadius: '4px'
}), select(' pre > code', {
  display: 'block',
  padding: '1rem 1.5 rem',
  whiteSpace: 'pre'
}),

// tables 
select(' th, td', {
  padding: '12px 15px',
  textAlign: 'left',
  borderBottom: '1px solid #E1E1E1'
}), select(' th:first-child, td:first-child', {
  paddingLeft: 0
}), select(' th:last-child, td:last-child', {
  paddingRight: 0
}),

// spacing
select(' button', {
  marginBottom: '1rem'
}), select(' input, textarea, select, fieldset', {
  marginBottom: '1.5rem'
}), select(' pre, blockquote, dl, figure, table, p, ul, ol, form', {
  marginBottom: '2.5rem'
}),

// misc 
select(' hr', {
  marginTop: '3rem',
  marginBottom: '3.5rem',
  borderWidth: 0,
  borderTop: '1px solid #E1E1E1'
})]));

// utilities 
export var fullWidth = style({
  width: '100%',
  boxSizing: 'border-box'
});

export var maxFullWidth = style({
  maxWidth: '100%',
  boxSizing: 'border-box'
});

export var pullRight = style({
  float: 'right'
});

export var pullLeft = style({
  float: 'left'
});

export var clearfix = style({
  content: '""',
  display: 'table',
  clear: 'both'
});