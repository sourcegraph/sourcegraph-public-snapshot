'use strict';

export default useRouterHistory;
import useQueries from 'history/lib/useQueries';
import useBasename from 'history/lib/useBasename';
function useRouterHistory(createHistory) {
  return function (options) {
    var history = useQueries(useBasename(createHistory))(options);
    history.__v2_compatible__ = true;
    return history;
  };
}