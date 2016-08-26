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
var utilities = require('./utilities');
var constants = require('../lib/constants');


describe('eventTracker', function() {

  before(function() {
    browser.url('/test/event-tracker.html');
  });

  beforeEach(function() {
    browser
        .execute(ga.run, 'create', 'UA-XXXXX-Y', 'auto')
        .execute(ga.trackHitData);
  });

  afterEach(function () {
    browser
        .execute(utilities.unstopSubmitEvents)
        .execute(ga.clearHitData)
        .execute(ga.run, 'eventTracker:remove')
        .execute(ga.run, 'remove');
  });

  it('should support declarative event binding to DOM elements', function() {

    var hitData = browser
        .execute(ga.run, 'require', 'eventTracker')
        .click('#click-test')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].eventCategory, 'foo');
    assert.equal(hitData[0].eventAction, 'bar');
    assert.equal(hitData[0].eventLabel, 'qux');
    assert.equal(hitData[0].eventValue, '42');
    assert.equal(hitData[0].dimension1, 'baz');
    assert.equal(hitData[0].nonInteraction, true);
  });


  it('should support customizing the attribute prefix', function() {

    var hitData = browser
        .execute(ga.run, 'require', 'eventTracker', {
          attributePrefix: 'data-ga-'
        })
        .click('#custom-prefix')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].eventCategory, 'foo');
    assert.equal(hitData[0].eventAction, 'bar');
  });


  it('should support non-event hit types', function() {

    var hitData = browser
        .execute(ga.run, 'require', 'eventTracker')
        .click('#social-hit-type')
        .click('#pageview-hit-type')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 2);
    assert.equal(hitData[0].hitType, 'social');
    assert.equal(hitData[0].socialNetwork, 'Facebook');
    assert.equal(hitData[0].socialAction, 'like');
    assert.equal(hitData[0].socialTarget, 'me');
    assert.equal(hitData[1].hitType, 'pageview');
    assert.equal(hitData[1].page, '/foobar.html');
  });


  it('should support customizing what events to listen for', function() {

    var hitData = browser
        .execute(utilities.stopSubmitEvents)
        .execute(ga.run, 'require', 'eventTracker', {
          events: ['submit']
        })
        .click('#click-test')
        .click('#submit-test')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].eventCategory, 'Forms');
    assert.equal(hitData[0].eventAction, 'submit');
  });


  it('should support specifying a fields object for all hits', function() {

    var hitData = browser
        .execute(utilities.stopSubmitEvents)
        .execute(ga.run, 'require', 'eventTracker', {
          fieldsObj: {
            nonInteraction: true,
            dimension1: 'foo',
            dimension2: 'bar'
          }
        })
        .click('#social-hit-type')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].nonInteraction, true);
    assert.equal(hitData[0].dimension1, 'foo');
    assert.equal(hitData[0].dimension2, 'bar');
  });


  it('should support specifying a hit filter', function() {

    var hitData = browser
        .execute(utilities.stopSubmitEvents)
        .execute(requireEventTrackerWithHitFilter)
        .click('#click-test')
        .click('#pageview-hit-type')
        .click('#social-hit-type')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].nonInteraction, true);
    assert.equal(hitData[0].dimension1, 'foo');
    assert.equal(hitData[0].dimension2, 'bar');
  });


  it('includes usage params with all hits', function() {

    var hitData = browser
        .execute(ga.run, 'require', 'eventTracker')
        .execute(ga.run, 'send', 'pageview')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].devId, constants.DEV_ID);
    assert.equal(hitData[0][constants.VERSION_PARAM], constants.VERSION);

    // '2' = '000000010' in hex
    assert.equal(hitData[0][constants.USAGE_PARAM], '2');
  });

});


/**
 * Since function objects can't be passed via parameters from server to
 * client, this one-off function must be used to set the value for
 * `hitFilter`.
 */
function requireEventTrackerWithHitFilter() {
  ga('require', 'eventTracker', {
    hitFilter: function(model, element) {
      if (element.id != 'social-hit-type') {
        throw 'Aborting non-social hits';
      }
      else {
        model.set('nonInteraction', true);
        model.set('dimension1', 'foo');
        model.set('dimension2', 'bar');
      }
    }
  });
}
