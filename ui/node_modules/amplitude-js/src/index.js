/* jshint expr:true */

var Amplitude = require('./amplitude');

var old = window.amplitude || {};
var instance = new Amplitude();
instance._q = old._q || [];

// export the instance
module.exports = instance;
