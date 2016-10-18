'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.css = exports.StyleSheet = undefined;

var _index = require('./index.js');

// todo 
// - animations 
// - fonts 

var StyleSheet = exports.StyleSheet = {
  create: function create(spec) {
    return Object.keys(spec).reduce(function (o, name) {
      return o[name] = (0, _index.style)(spec[name]), o;
    }, {});
  }
};

var css = exports.css = _index.merge;