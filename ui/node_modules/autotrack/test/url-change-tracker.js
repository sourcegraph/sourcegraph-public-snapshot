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


var browserCaps;
var baseUrl = browser.options.baseUrl;


describe('urlTracker', function() {

  before(function() {
    browserCaps = browser.session().value;

    browser.url('/test/url-change-tracker.html');
  });


  beforeEach(function() {
    browser
        .execute(ga.run, 'create', 'UA-XXXXX-Y', 'auto')
        .execute(ga.trackHitData);
  });


  afterEach(function () {
    browser
        .execute(ga.clearHitData)
        .execute(ga.run, 'urlChangeTracker:remove')
        .execute(ga.run, 'remove');
  });


  it('should capture URL changes via pushState and popstate', function() {

    if (notSupportedInBrowser()) return;

    browser.execute(ga.run, 'require', 'urlChangeTracker');

    var fooUrl = browser
        .click('#foo')
        .url()
        .value;

    assert.equal(fooUrl, baseUrl + '/test/foo.html');

    var barUrl = browser
        .click('#bar')
        .url()
        .value;

    assert.equal(barUrl, baseUrl + '/test/bar.html');

    var quxUrl = browser
        .click('#qux')
        .url()
        .value;

    assert.equal(quxUrl, baseUrl + '/test/qux.html');

    // TODO(philipwalton): Safari currently doesn't allow chaining the `back()`
    // method, so we have to separate this into two expressions. This can
    // probably be reverted in a future version (here and elsewhere).
    browser.back();
    var back1Url = browser.url().value;

    assert.equal(back1Url, baseUrl + '/test/bar.html');

    browser.back();
    var back2Url = browser.url().value;

    assert.equal(back2Url, baseUrl + '/test/foo.html');

    browser.back();
    var back3Url = browser.url().value;

    assert.equal(back3Url, baseUrl + '/test/url-change-tracker.html');

    var hitData = browser
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 6);
    assert.equal(hitData[0].page, '/test/foo.html');
    assert.equal(hitData[0].title, 'Foo');
    assert.equal(hitData[1].page, '/test/bar.html');
    assert.equal(hitData[1].title, 'Bar');
    assert.equal(hitData[2].page, '/test/qux.html');
    assert.equal(hitData[2].title, 'Qux');
    assert.equal(hitData[3].page, '/test/bar.html');
    assert.equal(hitData[3].title, 'Bar');
    assert.equal(hitData[4].page, '/test/foo.html');
    assert.equal(hitData[4].title, 'Foo');
    assert.equal(hitData[5].page, '/test/url-change-tracker.html');
    assert.equal(hitData[5].title, 'Home');
  });

  it('should update the tracker but not send hits when using replaceState',
      function() {

    if (notSupportedInBrowser()) return;

    browser.execute(ga.run, 'require', 'urlChangeTracker');

    var url = browser
        .click('#replace')
        .url()
        .value;

    // Replace state was called to just use the pathname value.
    assert.equal(url, baseUrl + '/test/replaced.html');

    url = browser
        .click('#restore')
        .url()
        .value;

    // Replace state was called to just use the pathname value.
    assert.equal(url, baseUrl + '/test/url-change-tracker.html');

    var hitData = browser
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 0);
  });


  it('should not capture hash changes', function() {

    if (notSupportedInBrowser()) return;

    browser.execute(ga.run, 'require', 'urlChangeTracker');

    var url = browser
        .click('#hash')
        .url()
        .value;

    assert.equal(url, baseUrl + '/test/url-change-tracker.html#hash');

    browser.back();
    var backUrl = browser.url().value;

    assert.equal(backUrl, baseUrl + '/test/url-change-tracker.html');

    var hitData = browser
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 0);
  });


  it('should support customizing what is considered a change', function() {

    if (notSupportedInBrowser()) return;

    browser.execute(requireUrlChangeTrackerTracker_shouldTrackUrlChange);

    var fooUrl = browser
        .click('#foo')
        .url()
        .value;

    assert.equal(fooUrl, baseUrl + '/test/foo.html');

    browser.back();
    var backUrl = browser.url().value;

    assert.equal(backUrl, baseUrl + '/test/url-change-tracker.html');

    var hitData = browser
       .execute(ga.getHitData)
       .value;

    assert.equal(hitData.length, 0);
  });


  it('should support customizing any field via the fieldsObj', function() {

    if (notSupportedInBrowser()) return;

    browser.execute(ga.run, 'require', 'urlChangeTracker', {
      fieldsObj: {
        dimension1: 'urlChangeTracker'
      }
    });

    browser.click('#foo');
    browser.back();

    var hitData = browser
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 2);
    assert.equal(hitData[0].page, '/test/foo.html');
    assert.equal(hitData[0].title, 'Foo');
    assert.equal(hitData[0].dimension1, 'urlChangeTracker');
    assert.equal(hitData[1].page, '/test/url-change-tracker.html');
    assert.equal(hitData[1].title, 'Home');
    assert.equal(hitData[1].dimension1, 'urlChangeTracker');
  });


  it('should support specifying a hit filter', function() {

    if (notSupportedInBrowser()) return;

    browser.execute(requireUrlChangeTrackerTracker_hitFilter);

    browser.click('#foo');
    browser.back();

    var hitData = browser
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].page, '/test/url-change-tracker.html');
    assert.equal(hitData[0].title, 'Home');
    assert.equal(hitData[0].dimension1, 'urlChangeTracker');
  });


  it('includes usage params with all hits', function() {

    var hitData = browser
        .execute(ga.run, 'require', 'urlChangeTracker')
        .execute(ga.run, 'send', 'pageview')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].devId, constants.DEV_ID);
    assert.equal(hitData[0][constants.VERSION_PARAM], constants.VERSION);

    // '100' = '100000000' in hex
    assert.equal(hitData[0][constants.USAGE_PARAM], '100');
  });

});


/**
 * @return {boolean} True if the current browser doesn't support all features
 *    required for these tests.
 */
function notSupportedInBrowser() {
  // IE9 doesn't support the HTML5 History API.
  return browserCaps.browserName == 'internet explorer' &&
      browserCaps.version == '9';
}


/**
 * Since function objects can't be passed via parameters from server to
 * client, this one-off function must be used to set the value for
 * `shouldTrackOutboundForm`.
 */
function requireUrlChangeTrackerTracker_shouldTrackUrlChange() {
  ga('require', 'urlChangeTracker', {
    shouldTrackUrlChange: function() {
      return false;
    }
  });
}


/**
 * Since function objects can't be passed via parameters from server to
 * client, this one-off function must be used to set the value for
 * `hitFilter`.
 */
function requireUrlChangeTrackerTracker_hitFilter() {
  ga('require', 'urlChangeTracker', {
    hitFilter: function(model) {
      var title = model.get('title');
      if (title == 'Foo') {
        throw 'Exclude Foo pages';
      }
      else {
        model.set('dimension1', 'urlChangeTracker');
      }
    }
  });
}
