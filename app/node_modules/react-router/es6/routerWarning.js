'use strict';

export default routerWarning;
import warning from 'warning';
function routerWarning(falseToWarn, message) {
  message = '[react-router] ' + message;

  for (var _len = arguments.length, args = Array(_len > 2 ? _len - 2 : 0), _key = 2; _key < _len; _key++) {
    args[_key - 2] = arguments[_key];
  }

  process.env.NODE_ENV !== 'production' ? warning.apply(undefined, [falseToWarn, message].concat(args)) : undefined;
}