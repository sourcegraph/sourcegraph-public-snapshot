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


describe('impressionTracker', function() {

  before(function() {
    browserCaps = browser.session().value;

    browser
        .url('/test/impression-tracker.html')
        .setViewportSize({width: 500, height: 500}, true);
  });

  beforeEach(function() {
    browser
        .scroll(0, 0)
        .execute(ga.run, 'create', 'UA-XXXXX-Y', 'auto')
        .execute(ga.trackHitData);
  });

  afterEach(function () {
    browser
        .execute(ga.clearHitData)
        .execute(ga.run, 'impressionTracker:remove')
        .execute(ga.run, 'remove');
  });


  it('tracks when elements are visible in the viewport', function() {

    if (notSupportedInBrowser()) return;

    browser
        .execute(ga.run, 'require', 'impressionTracker', {
          elements: [
            'foo',
            'foo-1',
            'foo-1-1',
            'foo-1-2',
            'foo-2',
            'foo-2-1',
            'foo-2-2',
            'bar',
            'bar-1',
            'bar-1-1',
            'bar-1-2',
            'bar-2',
            'bar-2-1',
            'bar-2-2'
          ]
        })
        .scroll('#foo')
        .waitUntil(ga.hitDataMatches([
          ['length', 7],
          ['[0].eventCategory', 'Viewport'],
          ['[0].eventAction', 'impression'],
          ['[0].eventLabel', 'foo'],
          ['[1].eventCategory', 'Viewport'],
          ['[1].eventAction', 'impression'],
          ['[1].eventLabel', 'foo-1'],
          ['[2].eventCategory', 'Viewport'],
          ['[2].eventAction', 'impression'],
          ['[2].eventLabel', 'foo-1-1'],
          ['[3].eventCategory', 'Viewport'],
          ['[3].eventAction', 'impression'],
          ['[3].eventLabel', 'foo-1-2'],
          ['[4].eventCategory', 'Viewport'],
          ['[4].eventAction', 'impression'],
          ['[4].eventLabel', 'foo-2'],
          ['[5].eventCategory', 'Viewport'],
          ['[5].eventAction', 'impression'],
          ['[5].eventLabel', 'foo-2-1'],
          ['[6].eventCategory', 'Viewport'],
          ['[6].eventAction', 'impression'],
          ['[6].eventLabel', 'foo-2-2']
        ]));

    browser
        .scroll('#bar')
        .waitUntil(ga.hitDataMatches([
          ['length', 14],
          ['[7].eventCategory', 'Viewport'],
          ['[7].eventAction', 'impression'],
          ['[7].eventLabel', 'bar'],
          ['[8].eventCategory', 'Viewport'],
          ['[8].eventAction', 'impression'],
          ['[8].eventLabel', 'bar-1'],
          ['[9].eventCategory', 'Viewport'],
          ['[9].eventAction', 'impression'],
          ['[9].eventLabel', 'bar-1-1'],
          ['[10].eventCategory', 'Viewport'],
          ['[10].eventAction', 'impression'],
          ['[10].eventLabel', 'bar-1-2'],
          ['[11].eventCategory', 'Viewport'],
          ['[11].eventAction', 'impression'],
          ['[11].eventLabel', 'bar-2'],
          ['[12].eventCategory', 'Viewport'],
          ['[12].eventAction', 'impression'],
          ['[12].eventLabel', 'bar-2-1'],
          ['[13].eventCategory', 'Viewport'],
          ['[13].eventAction', 'impression'],
          ['[13].eventLabel', 'bar-2-2']
        ]));
  });


  it('handles elements being added and removed from the DOM', function() {

    if (notSupportedInBrowser()) return;

    browser
        .execute(ga.run, 'require', 'impressionTracker', {
          elements: [
            {id: 'fixture', trackFirstImpressionOnly: false},
            {id: 'fixture-1', trackFirstImpressionOnly: false},
            {id: 'fixture-2', trackFirstImpressionOnly: false}
          ]
        })
        .execute(addFixtures)
        .scroll('#fixture')
        .waitUntil(ga.hitDataMatches([
          ['length', 3],
          ['[0].eventCategory', 'Viewport'],
          ['[0].eventAction', 'impression'],
          ['[0].eventLabel', 'fixture'],
          ['[1].eventCategory', 'Viewport'],
          ['[1].eventAction', 'impression'],
          ['[1].eventLabel', 'fixture-1'],
          ['[2].eventCategory', 'Viewport'],
          ['[2].eventAction', 'impression'],
          ['[2].eventLabel', 'fixture-2']
        ]));

    browser
        .execute(removeFixtures)
        .scroll('#foo')
        .execute(addFixtures)
        .scroll('#fixture')
        .waitUntil(ga.hitDataMatches([
          ['length', 6],
          ['[3].eventCategory', 'Viewport'],
          ['[3].eventAction', 'impression'],
          ['[3].eventLabel', 'fixture'],
          ['[4].eventCategory', 'Viewport'],
          ['[4].eventAction', 'impression'],
          ['[4].eventLabel', 'fixture-1'],
          ['[5].eventCategory', 'Viewport'],
          ['[5].eventAction', 'impression'],
          ['[5].eventLabel', 'fixture-2']
        ]));

    browser.execute(removeFixtures);
  });


  it('uses a default threshold of 0', function() {

    if (notSupportedInBrowser()) return;

    browser
        .execute(ga.run, 'require', 'impressionTracker', {
          elements: ['foo']
        })
        // Scrolls so #foo is only 0% visible but on the viewport border.
        .scroll('#foo', 0, -500)
        .waitUntil(ga.hitDataMatches([
          ['length', 1],
          ['[0].eventCategory', 'Viewport'],
          ['[0].eventAction', 'impression'],
          ['[0].eventLabel', 'foo']
        ]));
  });


  it('supports tracking an element either once or every time', function() {

    if (notSupportedInBrowser()) return;

    browser
        .execute(ga.run, 'require', 'impressionTracker', {
          elements: [
            'foo-1',
            {id: 'foo-2', trackFirstImpressionOnly: false},
            'bar-1',
            {id: 'bar-2', trackFirstImpressionOnly: false}
          ]
        })
        .scroll('#foo')
        .waitUntil(ga.hitDataMatches([
          ['length', 2],
          ['[0].eventCategory', 'Viewport'],
          ['[0].eventAction', 'impression'],
          ['[0].eventLabel', 'foo-1'],
          ['[1].eventCategory', 'Viewport'],
          ['[1].eventAction', 'impression'],
          ['[1].eventLabel', 'foo-2']
        ]));

    browser
        .scroll('#bar')
        .waitUntil(ga.hitDataMatches([
          ['length', 4],
          ['[2].eventCategory', 'Viewport'],
          ['[2].eventAction', 'impression'],
          ['[2].eventLabel', 'bar-1'],
          ['[3].eventCategory', 'Viewport'],
          ['[3].eventAction', 'impression'],
          ['[3].eventLabel', 'bar-2']
        ]));

    browser
        .scroll('#foo')
        .waitUntil(ga.hitDataMatches([
          ['length', 5],
          ['[4].eventCategory', 'Viewport'],
          ['[4].eventAction', 'impression'],
          ['[4].eventLabel', 'foo-2']
        ]));

    browser
        .scroll('#bar')
        .waitUntil(ga.hitDataMatches([
          ['length', 6],
          ['[5].eventCategory', 'Viewport'],
          ['[5].eventAction', 'impression'],
          ['[5].eventLabel', 'bar-2']
        ]));
  });


  it('supports changing the default threshold per element', function() {

    if (notSupportedInBrowser()) return;

    browser
        .execute(ga.run, 'require', 'impressionTracker', {
          elements: [
            {id: 'foo-1-1', threshold: 1},
            {id: 'foo-1-2', threshold: .66},
            {id: 'foo-2-1', threshold: .33},
            {id: 'foo-2-2', threshold: 0}
          ]
        })
        // Scrolls so #foo is only 25% visible
        .scroll('#foo', 0, -475)
        .waitUntil(ga.hitDataMatches([
          ['length', 1],
          ['[0].eventLabel', 'foo-2-2']
        ]));

    browser
        // Scrolls so #foo is 50% visible
        .scroll('#foo', 0, -450)
        .waitUntil(ga.hitDataMatches([
          ['length', 2],
          ['[1].eventLabel', 'foo-2-1']
        ]));

    browser
        // Scrolls so #foo is 75% visible
        .scroll('#foo', 0, -425)
        .waitUntil(ga.hitDataMatches([
          ['length', 3],
          ['[2].eventLabel', 'foo-1-2']
        ]));

    browser
        // Scrolls so #foo is 100% visible
        .scroll('#foo', 0, -400)
        .waitUntil(ga.hitDataMatches([
          ['length', 4],
          ['[3].eventLabel', 'foo-1-1']
        ]));
  });


  it('supports setting a rootMargin', function() {

    if (notSupportedInBrowser()) return;

    browser
        .execute(ga.run, 'require', 'impressionTracker', {
          rootMargin: '-50px 0px',
          elements: [
            {id: 'foo-1-1', threshold: 1},
            {id: 'foo-1-2', threshold: .66},
            {id: 'foo-2-1', threshold: .33},
            {id: 'foo-2-2', threshold: 0}
          ]
        })
        // Scrolls so #foo is 100% visible but only 50% within rootMargin.
        .scroll('#foo', 0, -400)
        .waitUntil(ga.hitDataMatches([
          ['length', 2],
          ['[0].eventCategory', 'Viewport'],
          ['[0].eventAction', 'impression'],
          ['[0].eventLabel', 'foo-2-1'],
          ['[1].eventCategory', 'Viewport'],
          ['[1].eventAction', 'impression'],
          ['[1].eventLabel', 'foo-2-2']
        ]));
  });


  it('supports declarative event binding to DOM elements', function() {

    if (notSupportedInBrowser()) return;

    browser
        .execute(ga.run, 'require', 'impressionTracker', {
          elements: ['attrs-1']
        })
        .scroll('#attrs')
        .waitUntil(ga.hitDataMatches([
          ['length', 1],
          ['[0].eventCategory', 'Element'],
          ['[0].eventAction', 'visible'],
          ['[0].eventLabel', 'attrs-1']
        ]));
  });


  it('supports customizing the attribute prefix', function() {

    if (notSupportedInBrowser()) return;

    browser
        .execute(ga.run, 'require', 'impressionTracker', {
          attributePrefix: 'data-ga-',
          elements: ['attrs-1', 'attrs-2']
        })
        .scroll('#attrs')
        .waitUntil(ga.hitDataMatches([
          ['length', 2],
          ['[0].eventCategory', 'Viewport'],
          ['[0].eventAction', 'impression'],
          ['[0].eventLabel', 'attrs-1'],
          ['[1].eventCategory', 'Window'],
          ['[1].eventAction', 'impression'],
          ['[1].eventLabel', 'attrs-2'],
          ['[1].nonInteraction', true]
        ]));
  });


  it('supports specifying a fields object for all hits', function() {

    if (notSupportedInBrowser()) return;

    browser
        .execute(ga.run, 'require', 'impressionTracker', {
          elements: ['foo', 'bar'],
          fieldsObj: {
            eventCategory: 'Element',
            eventAction: 'visible',
            nonInteraction: true
          }
        })
        .scroll('#foo')
        .waitUntil(ga.hitDataMatches([
          ['length', 1],
          ['[0].eventCategory', 'Element'],
          ['[0].eventAction', 'visible'],
          ['[0].eventLabel', 'foo'],
          ['[0].nonInteraction', true]
        ]));

    browser
        .scroll('#bar')
        .waitUntil(ga.hitDataMatches([
          ['length', 2],
          ['[1].eventCategory', 'Element'],
          ['[1].eventAction', 'visible'],
          ['[1].eventLabel', 'bar'],
          ['[1].nonInteraction', true]
        ]));
  });


  it('supports specifying a hit filter', function() {

    if (notSupportedInBrowser()) return;

    browser
        .execute(requireImpressionTracker_hitFilter)
        .scroll('#foo')
        .waitUntil(ga.hitDataMatches([
          ['length', 1],
          ['[0].eventCategory', 'Viewport'],
          ['[0].eventAction', 'impression'],
          ['[0].eventLabel', 'foo-2'],
          ['[0].nonInteraction', true],
          ['[0].dimension1', 'one'],
          ['[0].dimension2', 'two']
        ]));
  });


  it('includes usage params with all hits', function() {

    var hitData = browser
        .execute(ga.run, 'require', 'impressionTracker')
        .execute(ga.run, 'send', 'pageview')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].devId, constants.DEV_ID);
    assert.equal(hitData[0][constants.VERSION_PARAM], constants.VERSION);

    // '4' = '000000100' in hex
    assert.equal(hitData[0][constants.USAGE_PARAM], '4');
  });

});


/**
 * @return {boolean} True if the current browser doesn't support all features
 *    required for these tests.
 */
function notSupportedInBrowser() {
  // IE9 doesn't support the HTML5 History API.
  return browserCaps.browserName == 'internet explorer' &&
      (browserCaps.version == '9' || browserCaps.version == '10');
}


/**
 * Since function objects can't be passed via parameters from server to
 * client, this one-off function must be used to set the value for
 * `hitFilter`.
 */
function requireImpressionTracker_hitFilter() {
  ga('require', 'impressionTracker', {
    elements: ['foo-1', 'foo-2'],
    hitFilter: function(model, element) {
      if (element.id == 'foo-1') {
        throw 'Aborting hits with ID "foo-1"';
      }
      else {
        model.set('nonInteraction', true);
        model.set('dimension1', 'one');
        model.set('dimension2', 'two');
      }
    }
  });
}


/**
 * Adds a div#fixture.box element to the page.
 */
function addFixtures() {
  var fixture = document.createElement('div');
  fixture.id = 'fixture';
  fixture.className = 'container';
  fixture.innerHTML =
      '<div class="box" id="fixture-1"></div>' +
      '<div class="box" id="fixture-2"></div>';
  document.body.appendChild(fixture);
}


/**
 * Removes the div#fixture.box element from the page.
 */
function removeFixtures() {
  var fixture = document.getElementById('fixture');
  document.body.removeChild(fixture);
}
