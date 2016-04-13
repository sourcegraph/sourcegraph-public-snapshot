/*eslint no-empty: 0*/
'use strict';

export default deprecateObjectProperties;
import warning from './routerWarning';

var useMembrane = false;

if (process.env.NODE_ENV !== 'production') {
  try {
    if (Object.defineProperty({}, 'x', { get: function get() {
        return true;
      } }).x) {
      useMembrane = true;
    }
  } catch (e) {}
}

// wraps an object in a membrane to warn about deprecated property access

function deprecateObjectProperties(object, message) {
  if (!useMembrane) return object;

  var membrane = {};

  var _loop = function (prop) {
    if (typeof object[prop] === 'function') {
      membrane[prop] = function () {
        process.env.NODE_ENV !== 'production' ? warning(false, message) : undefined;
        return object[prop].apply(object, arguments);
      };
    } else {
      Object.defineProperty(membrane, prop, {
        configurable: false,
        enumerable: false,
        get: function get() {
          process.env.NODE_ENV !== 'production' ? warning(false, message) : undefined;
          return object[prop];
        }
      });
    }
  };

  for (var prop in object) {
    _loop(prop);
  }

  return membrane;
}