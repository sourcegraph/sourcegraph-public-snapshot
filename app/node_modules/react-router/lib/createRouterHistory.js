'use strict';

exports.__esModule = true;

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var _useRouterHistory = require('./useRouterHistory');

var _useRouterHistory2 = _interopRequireDefault(_useRouterHistory);

var canUseDOM = !!(typeof window !== 'undefined' && window.document && window.document.createElement);

exports['default'] = function (createHistory) {
  var history = undefined;
  if (canUseDOM) history = _useRouterHistory2['default'](createHistory)();
  return history;
};

module.exports = exports['default'];