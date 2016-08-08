/*
 * Cookie data
 */

var Base64 = require('./base64');
var JSON = require('json'); // jshint ignore:line
var topDomain = require('top-domain');
var utils = require('./utils');


var _options = {
  expirationDays: undefined,
  domain: undefined
};


var reset = function() {
  _options = {
    expirationDays: undefined,
    domain: undefined
  };
};


var options = function(opts) {
  if (arguments.length === 0) {
    return _options;
  }

  opts = opts || {};

  _options.expirationDays = opts.expirationDays;

  var domain = (!utils.isEmptyString(opts.domain)) ? opts.domain : '.' + topDomain(window.location.href);
  var token = Math.random();
  _options.domain = domain;
  set('amplitude_test', token);
  var stored = get('amplitude_test');
  if (!stored || stored !== token) {
    domain = null;
  }
  remove('amplitude_test');
  _options.domain = domain;
};

var _domainSpecific = function(name) {
  // differentiate between cookies on different domains
  var suffix = '';
  if (_options.domain) {
    suffix = _options.domain.charAt(0) === '.' ? _options.domain.substring(1) : _options.domain;
  }
  return name + suffix;
};


var get = function(name) {
  try {
    var nameEq = _domainSpecific(name) + '=';
    var ca = document.cookie.split(';');
    var value = null;
    for (var i = 0; i < ca.length; i++) {
      var c = ca[i];
      while (c.charAt(0) === ' ') {
        c = c.substring(1, c.length);
      }
      if (c.indexOf(nameEq) === 0) {
        value = c.substring(nameEq.length, c.length);
        break;
      }
    }

    if (value) {
      return JSON.parse(Base64.decode(value));
    }
    return null;
  } catch (e) {
    return null;
  }
};


var set = function(name, value) {
  try {
    _set(_domainSpecific(name), Base64.encode(JSON.stringify(value)), _options);
    return true;
  } catch (e) {
    return false;
  }
};


var _set = function(name, value, opts) {
  var expires = value !== null ? opts.expirationDays : -1 ;
  if (expires) {
    var date = new Date();
    date.setTime(date.getTime() + (expires * 24 * 60 * 60 * 1000));
    expires = date;
  }
  var str = name + '=' + value;
  if (expires) {
    str += '; expires=' + expires.toUTCString();
  }
  str += '; path=/';
  if (opts.domain) {
    str += '; domain=' + opts.domain;
  }
  document.cookie = str;
};


var remove = function(name) {
  try {
    _set(_domainSpecific(name), null, _options);
    return true;
  } catch (e) {
    return false;
  }
};


module.exports = {
  reset: reset,
  options: options,
  get: get,
  set: set,
  remove: remove

};
