var assert = require('assert');
var parseUrl = require('../lib/parse-url');

describe('parseUrl', function() {

  it('parses the a URL and returns a location-like object', function() {

    var url = parseUrl(
        'https://www.example.com:1234/path/to/file.html?a=b&c=d#hash');

    assert.deepEqual(url, {
      fragment: 'hash',
      hash: '#hash',
      host: 'www.example.com:1234',
      hostname: 'www.example.com',
      href: 'https://www.example.com:1234/path/to/file.html?a=b&c=d#hash',
      origin: 'https://www.example.com:1234',
      path: '/path/to/file.html?a=b&c=d',
      pathname: '/path/to/file.html',
      port: '1234',
      protocol: 'https:',
      query: 'a=b&c=d',
      search: '?a=b&c=d'
    });
  });


  it('parses a sparse URL', function() {

    var url = parseUrl('http://example.com');

    assert.deepEqual(url, {
      fragment: '',
      hash: '',
      host: 'example.com',
      hostname: 'example.com',
      href: 'http://example.com/', // Note the trailing slash.
      origin: 'http://example.com',
      path: '/',
      pathname: '/',
      port: '',
      protocol: 'http:',
      query: '',
      search: ''
    });
  });


  it('parses URLs relative to the root', function() {

    var url = parseUrl('/path/to/file.html?a=b&c=d#hash');

    // Specified portions of the URL.
    assert.equal(url.fragment, 'hash');
    assert.equal(url.hash, '#hash');
    assert.equal(url.path, '/path/to/file.html?a=b&c=d');
    assert.equal(url.pathname, '/path/to/file.html');
    assert.equal(url.query, 'a=b&c=d');
    assert.equal(url.search, '?a=b&c=d');

    // Non-specified portions of the URL should match `window.location`.
    assert.equal(url.host, location.host);
    assert.equal(url.hostname, location.hostname);
    assert.equal(url.port, location.port);
    assert.equal(url.protocol, location.protocol);

    // Not all browsers support the `origin` property, so we derive it.
    var origin = location.origin || location.protocol + '//' + location.host;
    assert.equal(url.origin, origin);
  });


  it('parses URLs relative to the file', function() {

    // Assumes the tests are hosted at `/test/`;
    var url = parseUrl('../path/to/file.html?a=b&c=d#hash');

    // Manually calculate the pathname since these tests run on servers as well
    // as using the file protocol.
    var pathname = location.pathname
        .replace(/test\/(index\.html)?/, '') + 'path/to/file.html';

    // Specified portions of the URL.
    assert.equal(url.fragment, 'hash');
    assert.equal(url.hash, '#hash');
    assert.equal(url.path, pathname + '?a=b&c=d');
    assert.equal(url.pathname, pathname);
    assert.equal(url.query, 'a=b&c=d');
    assert.equal(url.search, '?a=b&c=d');

    // Non-specified portions of the URL should match `window.location`.
    assert.equal(url.host, location.host);
    assert.equal(url.hostname, location.hostname);
    assert.equal(url.port, location.port);
    assert.equal(url.protocol, location.protocol);

    // Not all browsers support the `origin` property, so we derive it.
    var origin = location.origin || location.protocol + '//' + location.host;
    assert.equal(url.origin, origin);
  });


  it('should resolve various relative path types', function() {

    var url1 = parseUrl('.');
    assert.equal(url1.pathname, location.pathname);

    var url2 = parseUrl('..');
    assert.equal(url2.pathname,
        location.pathname.replace(/test\/(index.html)?$/, ''));

    var url3 = parseUrl('./foobar.html');
    assert.equal(url3.pathname,
        location.pathname.replace(/(index.html)?$/, 'foobar.html'));

    var url4 = parseUrl('../foobar.html');
    assert.equal(url4.pathname,
        location.pathname.replace(/test\/(index.html)?$/, 'foobar.html'));

    var url5 = parseUrl('.../foobar.html');
    assert.equal(url5.pathname,
        location.pathname.replace('index.html', '') + '.../foobar.html');
  });


  it('parses the current URL when given a falsy value', function() {

    var url = parseUrl();

    // Assumes the tests are hosted at `/test/`;
    assert.equal(url.fragment, '');
    assert.equal(url.hash, location.hash);
    assert.equal(url.path, location.pathname + location.search);
    assert.equal(url.pathname, location.pathname);
    assert.equal(url.query, '');
    assert.equal(url.search, location.search);

    // Non-specified portions of the URL should match `window.location`.
    assert.equal(url.host, location.host);
    assert.equal(url.hostname, location.hostname);
    assert.equal(url.port, location.port);
    assert.equal(url.protocol, location.protocol);

    // Not all browsers support the `origin` property, so we derive it.
    var origin = location.origin || location.protocol + '//' + location.host;
    assert.equal(url.origin, origin);

    assert.deepEqual(url, parseUrl(null));
    assert.deepEqual(url, parseUrl(''));
  });

});
