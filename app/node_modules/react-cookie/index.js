var cookie = require('cookie');

var _rawCookies = {};
var _cookies = {};

if (typeof document !== 'undefined') {
  setRawCookie(document.cookie);
}

function load(name, doNotParse) {
  if (doNotParse) {
    return _rawCookies[name];
  }

  return _cookies[name];
}

function save(name, val, opt) {
  _cookies[name] = val;
  _rawCookies[name] = val;

  // allow you to work with cookies as objects.
  if (typeof val === 'object') {
    _rawCookies[name] = JSON.stringify(val);
  }

  // Cookies only work in the browser
  if (typeof document !== 'undefined') {
    document.cookie = cookie.serialize(name, _rawCookies[name], opt);
  }
}

function remove(name) {
  delete _rawCookies[name];
  delete _cookies[name];

  if (typeof document !== 'undefined') {
    document.cookie = name + '=; expires=Thu, 01 Jan 1970 00:00:01 GMT;';
  }
}

function setRawCookie(rawCookie) {
  if (!rawCookie) {
    return;
  }

  var rawCookies = cookie.parse(rawCookie);

  for (var key in rawCookies) {
    _rawCookies[key] = rawCookies[key];

    try {
      _cookies[key] = JSON.parse(rawCookies[key]);
    } catch(e) {
      // Not serialized object
      _cookies[key] = rawCookies[key];
    }
  }
}

var reactCookie = {
  load: load,
  save: save,
  remove: remove,
  setRawCookie: setRawCookie
};

if (typeof window !== 'undefined') {
  window['reactCookie'] = reactCookie;
}

module.exports = reactCookie;
