/**
 * Copyright 2016 Google Inc. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */


var assert = require('assert');
var ga = require('./analytics');
var constants = require('../lib/constants');


describe('cleanUrlTracker', function() {

  before(function() {
    browser.url('/test/autotrack.html');
  });


  beforeEach(function() {
    browser
        .execute(ga.run, 'create', 'UA-XXXXX-Y', 'auto')
        .execute(ga.trackHitData);
  });


  afterEach(function () {
    browser
        .execute(ga.clearHitData)
        .execute(ga.run, 'cleanUrlTracker:remove')
        .execute(ga.run, 'remove');
  });


  it('does not modify the URL path by default',
      function() {

    var url = 'https://example.com/foo/bar?q=qux&b=baz#hash';
    var hitData = browser
        .execute(ga.run, 'require', 'cleanUrlTracker')
        .execute(ga.run, 'set', 'location', url)
        .execute(ga.run, 'send', 'pageview')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].location, url);
    assert.equal(hitData[0].page, '/foo/bar?q=qux&b=baz');
  });


  it('supports removing the query string from the URL path', function() {

    var url = 'https://example.com/foo/bar?q=qux&b=baz#hash';
    var hitData = browser
        .execute(ga.run, 'require', 'cleanUrlTracker', {
          stripQuery: true
        })
        .execute(ga.run, 'set', 'location', url)
        .execute(ga.run, 'send', 'pageview')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].location, url);
    assert.equal(hitData[0].page, '/foo/bar');
  });


  it('optionally adds the query string as a custom dimension', function() {

    var url = 'https://example.com/foo/bar?q=qux&b=baz#hash';
    var hitData = browser
        .execute(ga.run, 'require', 'cleanUrlTracker', {
          stripQuery: true,
          queryDimensionIndex: 1
        })
        .execute(ga.run, 'set', 'location', url)
        .execute(ga.run, 'send', 'pageview')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].location, url);
    assert.equal(hitData[0].page, '/foo/bar');
    assert.equal(hitData[0].dimension1, 'q=qux&b=baz');
  });


  it('adds the null dimensions when no query string is found', function() {

    var url = 'https://example.com/foo/bar';
    var hitData = browser
        .execute(ga.run, 'require', 'cleanUrlTracker', {
          stripQuery: true,
          queryDimensionIndex: 1
        })
        .execute(ga.run, 'set', 'location', url)
        .execute(ga.run, 'send', 'pageview')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].location, url);
    assert.equal(hitData[0].page, '/foo/bar');
    assert.equal(hitData[0].dimension1, constants.NULL_DIMENSION);
  });


  it('does not set a dimension if strip query is false', function() {

    var url = 'https://example.com/foo/bar?q=qux&b=baz#hash';
    var hitData = browser
        .execute(ga.run, 'require', 'cleanUrlTracker', {
          stripQuery: false,
          queryDimensionIndex: 1
        })
        .execute(ga.run, 'set', 'location', url)
        .execute(ga.run, 'send', 'pageview')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].location, url);
    assert.equal(hitData[0].page, '/foo/bar?q=qux&b=baz');
    assert.equal(hitData[0].dimension1, undefined);
  });


  it('cleans URLs in all hits, not just the initial pageview', function() {

    var url = 'https://example.com/foo/bar?q=qux&b=baz#hash';
    var hitData = browser
        .execute(ga.run, 'require', 'cleanUrlTracker', {
          stripQuery: true,
          queryDimensionIndex: 1
        })
        .execute(ga.run, 'set', 'location', url)
        .execute(ga.run, 'send', 'pageview')
        .execute(ga.run, 'set', 'page', '/updated?query=new' )
        .execute(ga.run, 'send', 'pageview')
        .execute(ga.run, 'set', 'page', '/more/updated?query=newest' )
        .execute(ga.run, 'send', 'event')
        .execute(ga.run, 'set', 'page', '/final#ly' )
        .execute(ga.run, 'send', 'event')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 4);
    assert.equal(hitData[0].page, '/foo/bar');
    assert.equal(hitData[0].dimension1, 'q=qux&b=baz');
    assert.equal(hitData[1].page, '/updated');
    assert.equal(hitData[1].dimension1, 'query=new');
    assert.equal(hitData[2].page, '/more/updated');
    assert.equal(hitData[2].dimension1, 'query=newest');
    assert.equal(hitData[3].page, '/final');
    assert.equal(hitData[3].dimension1, constants.NULL_DIMENSION);
  });


  it('supports removing index filenames', function() {

    var url = 'https://example.com/foo/bar/index.html?q=qux&b=baz#hash';
    var hitData = browser
        .execute(ga.run, 'require', 'cleanUrlTracker', {
          indexFilename: 'index.html'
        })
        .execute(ga.run, 'set', 'location', url)
        .execute(ga.run, 'send', 'pageview')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].page, '/foo/bar/?q=qux&b=baz');
  });


  it('only removes index filenames at the end of the URL after a slash',
    function() {

    var url = 'https://example.com/noindex.html';
    var hitData = browser
        .execute(ga.run, 'require', 'cleanUrlTracker', {
          indexFilename: 'index.html'
        })
        .execute(ga.run, 'set', 'location', url)
        .execute(ga.run, 'send', 'pageview')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].page, '/noindex.html');
  });


  it('supports stripping trailing slashes', function() {

    var url = 'https://example.com/foo/bar/';
    var hitData = browser
        .execute(ga.run, 'require', 'cleanUrlTracker', {
          trailingSlash: 'remove'
        })
        .execute(ga.run, 'set', 'location', url)
        .execute(ga.run, 'send', 'pageview')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].page, '/foo/bar');
  });


  it('supports adding trailing slashes to non-filename URLs', function() {

    var url = 'https://example.com/foo/bar?q=qux&b=baz#hash';
    var hitData = browser
        .execute(ga.run, 'require', 'cleanUrlTracker', {
          stripQuery: true,
          queryDimensionIndex: 1,
          trailingSlash: 'add'
        })
        .execute(ga.run, 'set', 'location', url)
        .execute(ga.run, 'send', 'pageview')
        .execute(ga.run, 'set', 'page', '/foo/bar.html')
        .execute(ga.run, 'send', 'pageview')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 2);
    assert.equal(hitData[0].page, '/foo/bar/');
    assert.equal(hitData[1].page, '/foo/bar.html');
  });


  it('works with many options in conjunction with each other', function() {

    var url = 'https://example.com/path/to/index.html?q=qux&b=baz#hash';
    var hitData = browser
        .execute(ga.run, 'require', 'cleanUrlTracker', {
          stripQuery: true,
          queryDimensionIndex: 1,
          indexFilename: 'index.html',
          trailingSlash: 'remove'
        })
        .execute(ga.run, 'set', 'location', url)
        .execute(ga.run, 'send', 'pageview')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].page, '/path/to');
    assert.equal(hitData[0].dimension1, 'q=qux&b=baz');
  });


  it('includes usage params with all hits', function() {

    var hitData = browser
        .execute(ga.run, 'require', 'cleanUrlTracker')
        .execute(ga.run, 'send', 'pageview')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].devId, constants.DEV_ID);
    assert.equal(hitData[0][constants.VERSION_PARAM], constants.VERSION);

    // '1' = '000000001' in hex
    assert.equal(hitData[0][constants.USAGE_PARAM], '1');
  });

});
