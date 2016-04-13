'use strict';

exports.__esModule = true;
exports['default'] = useRouterHistory;

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var _historyLibUseQueries = require('history/lib/useQueries');

var _historyLibUseQueries2 = _interopRequireDefault(_historyLibUseQueries);

var _historyLibUseBasename = require('history/lib/useBasename');

var _historyLibUseBasename2 = _interopRequireDefault(_historyLibUseBasename);

function useRouterHistory(createHistory) {
  return function (options) {
    var history = _historyLibUseQueries2['default'](_historyLibUseBasename2['default'](createHistory))(options);
    history.__v2_compatible__ = true;
    return history;
  };
}

module.exports = exports['default'];