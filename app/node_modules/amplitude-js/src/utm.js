var getUtmParam = function(name, query) {
  name = name.replace(/[\[]/, "\\[").replace(/[\]]/, "\\]");
  var regex = new RegExp("[\\?&]" + name + "=([^&#]*)");
  var results = regex.exec(query);
  return results === null ? undefined : decodeURIComponent(results[1].replace(/\+/g, " "));
};

var getUtmData = function(rawCookie, query) {
  // Translate the utmz cookie format into url query string format.
  var cookie = rawCookie ? '?' + rawCookie.split('.').slice(-1)[0].replace(/\|/g, '&') : '';

  var fetchParam = function (queryName, query, cookieName, cookie) {
    return getUtmParam(queryName, query) ||
           getUtmParam(cookieName, cookie);
  };

  return {
    utm_source: fetchParam('utm_source', query, 'utmcsr', cookie),
    utm_medium: fetchParam('utm_medium', query, 'utmcmd', cookie),
    utm_campaign: fetchParam('utm_campaign', query, 'utmccn', cookie),
    utm_term: fetchParam('utm_term', query, 'utmctr', cookie),
    utm_content: fetchParam('utm_content', query, 'utmcct', cookie),
  };
};

module.exports = getUtmData;
