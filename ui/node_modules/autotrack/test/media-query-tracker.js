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
var TIMEOUT = 1000;


var opts = {
  definitions: [
    {
      name: 'Width',
      dimensionIndex: 1,
      items: [
        {name: 'sm', media: 'all'},
        {name: 'md', media: '(min-width: 480px)'},
        {name: 'lg', media: '(min-width: 640px)'}
      ]
    },
    {
      name: 'Height',
      dimensionIndex: 2,
      items: [
        {name: 'sm', media: 'all'},
        {name: 'md', media: '(min-height: 480px)'},
        {name: 'lg', media: '(min-height: 640px)'}
      ]
    }
  ]
};


describe('mediaQueryTracker', function() {

  before(function() {
    browserCaps = browser.session().value;

    // Loads the autotrack file since no custom HTML is needed.
    browser.url('/test/autotrack.html');
  });


  beforeEach(function() {
    browser
        .setViewportSize({width: 800, height: 600}, false)
        .execute(ga.run, 'create', 'UA-XXXXX-Y', 'auto')
        .execute(ga.trackHitData);
  });


  afterEach(function () {
    browser
        .execute(ga.clearHitData)
        .execute(ga.run, 'mediaQueryTracker:remove')
        .execute(ga.run, 'remove');
  });


  it('should set initial data via custom dimensions', function() {

    if (notSupportedInBrowser()) return;

    browser
        .execute(ga.run, 'require', 'mediaQueryTracker', opts)
        .waitUntil(ga.trackerDataMatches([
          ['dimension1', 'lg'],
          ['dimension2', 'md']
        ]));
  });


  it('should send events when the matched media changes', function() {

    if (notSupportedInBrowser()) return;

    browser
        .execute(ga.run, 'require', 'mediaQueryTracker', opts)
        .setViewportSize({width: 400, height: 400}, false)
        .waitUntil(ga.trackerDataMatches([
          ['dimension1', 'sm'],
          ['dimension2', 'sm']
        ]));

    browser
        .waitUntil(ga.hitDataMatches([
          ['[0].eventCategory', 'Width'],
          ['[0].eventAction', 'change'],
          ['[0].eventLabel', 'lg => sm'],
          ['[1].eventCategory', 'Height'],
          ['[1].eventAction', 'change'],
          ['[1].eventLabel', 'md => sm']
        ]));
  });


  it('should wait for the timeout to set or send changes', function() {

    if (notSupportedInBrowser()) return;

   browser
        .execute(ga.run, 'require', 'mediaQueryTracker', opts)
        .setViewportSize({width: 400, height: 400}, false);

    var timeoutStart = Date.now();
    browser.waitUntil(ga.trackerDataMatches([
      ['dimension1', 'sm'],
      ['dimension2', 'sm']
    ]));
    browser.waitUntil(ga.hitDataMatches([
      ['length', 2]
    ]));
    var timeoutDuration = Date.now() - timeoutStart;

    assert(timeoutDuration >= TIMEOUT);
  });


  it('should support customizing the timeout period', function() {

    if (notSupportedInBrowser()) return;

    browser
        .execute(ga.run, 'require', 'mediaQueryTracker',
            Object.assign({}, opts, {changeTimeout: 0}))
        .setViewportSize({width: 400, height: 400}, false);

    var shortTimeoutStart = Date.now();
    browser.waitUntil(ga.trackerDataMatches([
      ['dimension1', 'sm'],
      ['dimension2', 'sm']
    ]));
    browser.waitUntil(ga.hitDataMatches([
      ['length', 2]
    ]));
    var shortTimeoutDuration = Date.now() - shortTimeoutStart;

    browser
        .execute(ga.clearHitData)
        .execute(ga.run, 'mediaQueryTracker:remove')
        .execute(ga.run, 'remove')
        .execute(ga.run, 'create', 'UA-XXXXX-Y', 'auto')
        .execute(ga.trackHitData)
        .setViewportSize({width: 800, height: 600}, false)
        .execute(ga.run, 'require', 'mediaQueryTracker', opts)
        .setViewportSize({width: 400, height: 400}, false);

    var longTimeoutStart = Date.now();
    browser.waitUntil(ga.trackerDataMatches([
      ['dimension1', 'sm'],
      ['dimension2', 'sm']
    ]));
    browser.waitUntil(ga.hitDataMatches([
      ['length', 2]
    ]));
    var longTimeoutDuration = Date.now() - longTimeoutStart;

    // The long timeout should, in theory, be 1000ms longer, but we compare
    // to 500 just to be safe and avoid flakiness.
    assert(longTimeoutDuration - shortTimeoutDuration > (TIMEOUT/2));
  });


  it('should support customizing the change template', function() {

    if (notSupportedInBrowser()) return;

    browser
        .execute(requireMediaQueryTracker_changeTemplate)
        .setViewportSize({width: 400, height: 400}, false)
        .waitUntil(ga.hitDataMatches([
          ['[0].eventLabel', 'lg:sm'],
          ['[1].eventLabel', 'md:sm']
        ]));
  });

  it('should support customizing any field via the fieldsObj', function() {

    if (notSupportedInBrowser()) return;

    browser
        .execute(ga.run, 'require', 'mediaQueryTracker',
            Object.assign({}, opts, {
              changeTimeout: 0,
              fieldsObj: {
                nonInteraction: true
              }
            }))
        .setViewportSize({width: 400, height: 400}, false)
        .waitUntil(ga.hitDataMatches([
          ['[0].eventCategory', 'Width'],
          ['[0].eventAction', 'change'],
          ['[0].eventLabel', 'lg => sm'],
          ['[0].nonInteraction', true],
          ['[1].eventCategory', 'Height'],
          ['[1].eventAction', 'change'],
          ['[1].eventLabel', 'md => sm'],
          ['[1].nonInteraction', true]
        ]));
  });


  it('should support specifying a hit filter', function() {

    if (notSupportedInBrowser()) return;

    browser
        .execute(requireMediaQueryTracker_hitFilter)
        .setViewportSize({width: 400, height: 400}, false)
        .waitUntil(ga.hitDataMatches([
          ['[0].eventCategory', 'Height'],
          ['[0].eventAction', 'change'],
          ['[0].eventLabel', 'md => sm'],
          ['[0].nonInteraction', true]
        ]));
  });


  it('includes usage params with all hits', function() {

    var hitData = browser
        .execute(ga.run, 'require', 'mediaQueryTracker')
        .execute(ga.run, 'send', 'pageview')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].devId, constants.DEV_ID);
    assert.equal(hitData[0][constants.VERSION_PARAM], constants.VERSION);

    // '8' = '000001000' in hex
    assert.equal(hitData[0][constants.USAGE_PARAM], '8');
  });

});


/**
 * @return {boolean} True if the current browser doesn't support all features
 *    required for these tests.
 */
function notSupportedInBrowser() {
  // TODO(philipwalton): Some capabilities aren't implemented, so we can't test
  // against Edge right now. Wait for build 10532 to support setViewportSize
  // https://dev.windows.com/en-us/microsoft-edge/platform/status/webdriver/details/

  // IE9 doesn't support matchMedia, so it's not tested.
  return browserCaps.browserName == 'MicrosoftEdge' ||
      (browserCaps.browserName == 'internet explorer' &&
          browserCaps.version == '9');
}


/**
 * Since function objects can't be passed via parameters from server to
 * client, this one-off function must be used to set the value for
 * `changeTemplate`.
 */
function requireMediaQueryTracker_changeTemplate() {
  ga('require', 'mediaQueryTracker', {
    definitions: [
      {
        name: 'Width',
        dimensionIndex: 1,
        items: [
          {name: 'sm', media: 'all'},
          {name: 'md', media: '(min-width: 480px)'},
          {name: 'lg', media: '(min-width: 640px)'}
        ]
      },
      {
        name: 'Height',
        dimensionIndex: 2,
        items: [
          {name: 'sm', media: 'all'},
          {name: 'md', media: '(min-height: 480px)'},
          {name: 'lg', media: '(min-height: 640px)'}
        ]
      }
    ],
    changeTemplate: function(oldValue, newValue) {
      return oldValue + ':' + newValue;
    }
  });
}


/**
 * Since function objects can't be passed via parameters from server to
 * client, this one-off function must be used to set the value for
 * `hitFilter`.
 */
function requireMediaQueryTracker_hitFilter() {
  ga('require', 'mediaQueryTracker', {
    definitions: [
      {
        name: 'Width',
        dimensionIndex: 1,
        items: [
          {name: 'sm', media: 'all'},
          {name: 'md', media: '(min-width: 480px)'},
          {name: 'lg', media: '(min-width: 640px)'}
        ]
      },
      {
        name: 'Height',
        dimensionIndex: 2,
        items: [
          {name: 'sm', media: 'all'},
          {name: 'md', media: '(min-height: 480px)'},
          {name: 'lg', media: '(min-height: 640px)'}
        ]
      }
    ],
    hitFilter: function(model) {
      var category = model.get('eventCategory');
      if (category == 'Width') {
        throw 'Exclude width changes';
      }
      else {
        model.set('nonInteraction', true);
      }
    }
  });
}

