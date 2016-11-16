import { merge, style } from './index.js';

// todo 
// - animations 
// - fonts 

export var StyleSheet = {
  create: function create(spec) {
    return Object.keys(spec).reduce(function (o, name) {
      return o[name] = style(spec[name]), o;
    }, {});
  }
};

export var css = merge;