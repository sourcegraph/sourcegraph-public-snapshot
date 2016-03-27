'use strict';

import useRouterHistory from './useRouterHistory';

var canUseDOM = !!(typeof window !== 'undefined' && window.document && window.document.createElement);

export default function (createHistory) {
  var history = undefined;
  if (canUseDOM) history = useRouterHistory(createHistory)();
  return history;
}