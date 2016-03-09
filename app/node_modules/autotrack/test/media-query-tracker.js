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
var get = require('lodash/object/get');


var browserCaps;
var TIMEOUT = 1000;


describe('mediaQueryTracker', function() {

  before(function *() {
    browserCaps = (yield browser.session()).value;
  });


  beforeEach(function() {

    if (notSupportedInBrowser()) return;

    return browser
        // Loads a blank page to speed up testing.
        .url('/test/blank.html')
        .setViewportSize({width:800, height:600}, false);
  });


  it('should set initial data via custom dimensions', function() {

    if (notSupportedInBrowser()) return;

    return browser
        .url('/test/media-query-tracker.html')
        .waitUntil(dimensionDataMatches([
          ['dimension1', 'lg'],
          ['dimension2', 'md']
        ]));
  });


  it('should send events when the matched media changes', function() {

    if (notSupportedInBrowser()) return;

    return browser
        .url('/test/media-query-tracker.html')
        .setViewportSize({width:400, height:400}, false)
        .waitUntil(dimensionDataMatches([
          ['dimension1', 'sm'],
          ['dimension2', 'sm']
        ]))
        .waitUntil(hitDataMatches([
          ['[0].eventCategory', 'Width'],
          ['[0].eventAction', 'change'],
          ['[0].eventLabel', 'lg => sm'],
          ['[1].eventCategory', 'Height'],
          ['[1].eventAction', 'change'],
          ['[1].eventLabel', 'md => sm']
        ]));
  });


  it('should wait for the timeout to set or send changes', function *() {

    if (notSupportedInBrowser()) return;

    yield browser
        .url('/test/media-query-tracker.html')
        .setViewportSize({width:400, height:400}, false)

    var timeoutStart = Date.now();
    yield browser.waitUntil(dimensionDataMatches([
      ['dimension1', 'sm'],
      ['dimension2', 'sm']
    ]))
    .waitUntil(hitDataMatches([
      ['count', 2]
    ]));
    var timeoutDuration = Date.now() - timeoutStart;

    assert(timeoutDuration >= TIMEOUT);
  });


  it('should support customizing the timeout period', function *() {

    if (notSupportedInBrowser()) return;

    yield browser
        .url('/test/media-query-tracker-change-timeout.html')
        .setViewportSize({width:400, height:400}, false)

    var shortTimeoutStart = Date.now();
    yield browser.waitUntil(dimensionDataMatches([
      ['dimension1', 'sm'],
      ['dimension2', 'sm']
    ]))
    .waitUntil(hitDataMatches([
      ['count', 2]
    ]));
    var shortTimeoutDuration = Date.now() - shortTimeoutStart;

    yield browser
        .setViewportSize({width:800, height:600}, false)
        .url('/test/media-query-tracker.html')
        .setViewportSize({width:400, height:400}, false);

    var longTimeoutStart = Date.now();
    yield browser.waitUntil(dimensionDataMatches([
      ['dimension1', 'sm'],
      ['dimension2', 'sm']
    ]))
    .waitUntil(hitDataMatches([
      ['count', 2]
    ]));
    var longTimeoutDuration = Date.now() - longTimeoutStart;

    // The long timeout should, in theory, be 1000ms longer, but we compare
    // to 500 just to be safe and avoid flakiness.
    assert(longTimeoutDuration - shortTimeoutDuration > (TIMEOUT/2));
  });


  it('should support customizing the change template', function() {

    if (notSupportedInBrowser()) return;

    return browser
        .url('/test/media-query-tracker-change-template.html')
        .setViewportSize({width:400, height:400}, false)
        .waitUntil(hitDataMatches([
          ['[0].eventLabel', 'lg:sm'],
          ['[1].eventLabel', 'md:sm']
        ]));
  });


  it('should include the &did param with all hits', function() {

    return browser
        .url('/test/media-query-tracker.html')
        .execute(sendPageview)
        .waitUntil(hitDataMatches([['[0].devId', 'i5iSjo']]));
  });

});


function sendPageview() {
  ga('send', 'pageview');
}


function getHitData() {
  return hitData;
}


function hitDataMatches(expected) {
  return function() {
    return browser.execute(getHitData).then(function(hitData) {
      return expected.every(function(item) {
        return get(hitData.value, item[0]) === item[1];
      });
    });
  };
}


function getDimensionData() {
  var tracker = ga.getAll()[0];
  return {
    dimension1: tracker.get('dimension1'),
    dimension2: tracker.get('dimension2')

  };
}


function dimensionDataMatches(expected) {
  return function() {
    return browser.execute(getDimensionData).then(function(dimensionData) {
      return expected.every(function(item) {
        return get(dimensionData.value, item[0]) === item[1];
      });
    });
  };
}


function isEdge() {
  return browserCaps.browserName == 'MicrosoftEdge';
}


function isIE9() {
  return browserCaps.browserName == 'internet explorer' &&
         browserCaps.version == '9';
}


function notSupportedInBrowser() {
  // TODO(philipwalton): Some capabilities aren't implemented, so we can't test
  // against Edge right now. Wait for build 10532 to support setViewportSize
  // https://dev.windows.com/en-us/microsoft-edge/platform/status/webdriver/details/

  // IE9 doesn't support matchMedia, so it's not tested.
  return isEdge() || isIE9();
}
