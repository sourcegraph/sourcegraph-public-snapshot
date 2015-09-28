/**
 * Backbone.js plugin
 *
 * Patches Backbone.Events callbacks.
 */
;(function(window, Raven, Backbone) {
'use strict';

// quit if Backbone isn't on the page
if (!Backbone) {
    return;
}

function makeBackboneEventsOn(oldOn) {
  return function BackboneEventsOn(name, callback, context) {
    var wrapCallback = function (cb) {
      if (Object.prototype.toString.call(cb) === '[object Function]') {
        var _callback = cb._callback || cb;
        cb = Raven.wrap(cb);
        cb._callback = _callback;
      }
      return cb;
    };
    if (Object.prototype.toString.call(name) === '[object Object]') {
      // Handle event maps.
      for (var key in name) {
        if (name.hasOwnProperty(key)) {
          name[key] = wrapCallback(name[key]);
        }
      }
    } else {
      callback = wrapCallback(callback);
    }
    return oldOn.call(this, name, callback, context);
  };
}

// We're too late to catch all of these by simply patching Backbone.Events.on
var affectedObjects = [
  Backbone.Events,
  Backbone,
  Backbone.Model.prototype,
  Backbone.Collection.prototype,
  Backbone.View.prototype,
  Backbone.Router.prototype,
  Backbone.History.prototype
], i = 0, l = affectedObjects.length;

for (; i < l; i++) {
  var affected = affectedObjects[i];
  affected.on = makeBackboneEventsOn(affected.on);
  affected.bind = affected.on;
}

}(window, window.Raven, window.Backbone));
