'use strict';

exports.__esModule = true;
exports.default = useScroll;

var _react = require('react');

var _react2 = _interopRequireDefault(_react);

var _ScrollBehaviorContext = require('./ScrollBehaviorContext');

var _ScrollBehaviorContext2 = _interopRequireDefault(_ScrollBehaviorContext);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function useScroll(shouldUpdateScroll) {
  return {
    renderRouterContext: function renderRouterContext(child, props) {
      return _react2.default.createElement(
        _ScrollBehaviorContext2.default,
        {
          shouldUpdateScroll: shouldUpdateScroll,
          routerProps: props
        },
        child
      );
    }
  };
}
module.exports = exports['default'];