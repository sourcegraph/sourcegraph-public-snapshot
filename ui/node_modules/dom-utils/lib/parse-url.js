var HTTP_PORT = '80';
var HTTPS_PORT = '443';
var DEFAULT_PORT = RegExp(':(' + HTTP_PORT + '|' + HTTPS_PORT + ')$');


var a = document.createElement('a');
var cache = {};


/**
 * Parses the given url and returns an object mimicing a `Location` object.
 * @param {string} url The url to parse.
 * @return {Object} An object with the same properties as a `Location`
 *    plus the convience properties `path` and `query`.
 */
module.exports = function parseUrl(url) {

  // All falsy values (as well as ".") should map to the current URL.
  url = (!url || url == '.') ? location.href : url;

  if (cache[url]) return cache[url];

  a.href = url;

  // When parsing file relative paths (e.g. `../index.html`), IE will correctly
  // resolve the `href` property but will keep the `..` in the `path` property.
  // To workaround this, we reparse with the full URL from the `href` property.
  if (url.charAt(0) == '.') return parseUrl(a.href);

  // Sometimes IE will return no port or just a colon, especially for things
  // like relative port URLs (e.g. "//google.com").
  var protocol = !a.protocol || ':' == a.protocol ?
      location.protocol : a.protocol;

  // Don't include default ports.
  var port = (a.port == HTTP_PORT || a.port == HTTPS_PORT) ? '' : a.port;

  // PhantomJS sets the port to "0" when using the file: protocol.
  port = port == '0' ? '' : port;

  // IE will return an empty string for host and hostname with a relative URL.
  var host = a.host == '' ? location.host : a.host;
  var hostname = a.hostname == '' ? location.hostname : a.hostname;

  // Sometimes IE incorrectly includes a port for default ports
  // (e.g. `:80` or `:443`) even when no port is specified in the URL.
  // http://bit.ly/1rQNoMg
  host = host.replace(DEFAULT_PORT, '');

  // Not all browser support `origin` so we have to build it.
  var origin = a.origin ? a.origin : protocol + '//' + host;

  // Sometimes IE doesn't include the leading slash for pathname.
  // http://bit.ly/1rQNoMg
  var pathname = a.pathname.charAt(0) == '/' ? a.pathname : '/' + a.pathname;

  return cache[url] = {
    hash: a.hash,
    host: host,
    hostname: hostname,
    href: a.href,
    origin: origin,

    pathname: pathname,
    port: port,
    protocol: protocol,
    search: a.search,

    // Expose additional helpful properties not part of the Location object.
    fragment: a.hash.slice(1), // The hash without the starting "#".
    path: pathname + a.search, // The pathname and the search query (w/o hash).
    query: a.search.slice(1) // The search without the starting "?".
  };
};
