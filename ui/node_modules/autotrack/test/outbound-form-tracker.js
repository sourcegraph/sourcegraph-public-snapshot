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


describe('outboundFormTracker', function() {

  before(setupPage);
  beforeEach(startTracking);
  afterEach(stopTracking);


  it('should send events on outbound form submits', function() {

    var hitData = browser
        .execute(utilities.stopSubmitEvents)
        .execute(utilities.stubBeacon)
        .execute(ga.run, 'require', 'outboundFormTracker')
        .click('#outbound-submit')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].eventCategory, 'Outbound Form');
    assert.equal(hitData[0].eventAction, 'submit');
    assert.equal(hitData[0].eventLabel,
        'https://www.google-analytics.com/collect');
  });


  it('should not send events on local form submits', function() {

    var hitData = browser
        .execute(utilities.stopSubmitEvents)
        .execute(utilities.stubBeacon)
        .execute(ga.run, 'require', 'outboundFormTracker')
        .click('#local-submit')
        .execute(ga.getHitData)
        .value;

    assert(!hitData.length);
  });


  it('should work with forms missing the action attribute', function() {

    var hitData = browser
        .execute(utilities.stopSubmitEvents)
        .execute(utilities.stubBeacon)
        .execute(ga.run, 'require', 'outboundFormTracker')
        .click('#action-less-submit')
        .execute(ga.getHitData)
        .value;

    assert(!hitData.length);
  });


  it('should navigate to the proper outbound location on submit', function() {

    browser
        .execute(utilities.stubBeacon)
        .execute(ga.run, 'require', 'outboundFormTracker')
        .click('#outbound-submit')
        .waitUntil(utilities.urlMatches(
            'https://www.google-analytics.com/collect'));

    // Restores the page state.
    setupPage();
  });


  it('should navigate to the proper local location on submit', function() {

    browser
        .execute(utilities.stubBeacon)
        .execute(ga.run, 'require', 'outboundFormTracker')
        .click('#local-submit')
        .waitUntil(utilities.urlMatches('/test/blank.html'));

    // Restores the page state.
    setupPage();
  });


  it('should stop the event when beacon is not supported and re-emit ' +
      'after the hit succeeds or times out', function() {

    var hitData = browser
        .execute(utilities.disableProgramaticFormSubmits)
        .execute(utilities.stubNoBeacon)
        .execute(ga.run, 'require', 'outboundFormTracker')
        .click('#outbound-submit')
        .execute(ga.getHitData)
        .value;

    // Tests that the hit is sent.
    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].eventCategory, 'Outbound Form');
    assert.equal(hitData[0].eventAction, 'submit');
    assert.equal(hitData[0].eventLabel,
        'https://www.google-analytics.com/collect');

    // Tests that navigation actually happens
    setupPage();
    startTracking();
    browser
        .execute(utilities.stubNoBeacon)
        .execute(ga.run, 'require', 'outboundFormTracker')
        .click('#outbound-submit')
        .waitUntil(utilities.urlMatches(
            'https://www.google-analytics.com/collect'));

    // Restores the page state.
    setupPage();

    // TODO(philipwalton): figure out a way to test the hitCallback timing out.
  });


  it('should support customizing the selector used to detect form submits',
      function() {

    var hitData = browser
        .execute(utilities.stopSubmitEvents)
        .execute(utilities.stubBeacon)
        .execute(ga.run, 'require', 'outboundFormTracker', {
          formSelector: '.form'
        })
        .click('#outbound-submit-with-class')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].eventCategory, 'Outbound Form');
    assert.equal(hitData[0].eventAction, 'submit');
    assert.equal(hitData[0].eventLabel, 'https://example.com/');
  });


  it('should support customizing what is considered an outbound form',
      function() {

    var hitData = browser
        .execute(utilities.stopSubmitEvents)
        .execute(utilities.stubBeacon)
        .execute(requireOutboundFormTracker_shouldTrackOutboundForm)
        .click('#outbound-submit')
        .click('#outbound-submit-with-class')
        .click('#local-submit')
        .click('#action-less-submit')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].eventCategory, 'Outbound Form');
    assert.equal(hitData[0].eventAction, 'submit');
    assert.equal(hitData[0].eventLabel, 'https://example.com/');
  });


  it('should support customizing any field via the fieldsObj', function() {

    var hitData = browser
        .execute(utilities.stopSubmitEvents)
        .execute(utilities.stubBeacon)
        .execute(ga.run, 'require', 'outboundFormTracker', {
          fieldsObj: {
            eventCategory: 'External Form',
            eventAction: 'send',
            nonInteraction: true
          }
        })
        .click('#outbound-submit')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].eventCategory, 'External Form');
    assert.equal(hitData[0].eventAction, 'send');
    assert.equal(hitData[0].eventLabel,
        'https://www.google-analytics.com/collect');
    assert.equal(hitData[0].nonInteraction, true);
  });


  it('supports setting attributes declaratively', function() {

    var hitData = browser
        .execute(utilities.stopSubmitEvents)
        .execute(utilities.stubBeacon)
        .execute(ga.run, 'require', 'outboundFormTracker')
        .click('#declarative-attributes-submit')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].eventCategory, 'External Form');
    assert.equal(hitData[0].eventAction, 'submit');
    assert.equal(hitData[0].dimension1, true);
  });


  it('supports customizing the attribute prefix', function() {

    var hitData = browser
        .execute(utilities.stopSubmitEvents)
        .execute(utilities.stubBeacon)
        .execute(ga.run, 'require', 'outboundFormTracker', {
          attributePrefix: 'data-ga-'
        })
        .click('#declarative-attributes-prefix-submit')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].eventAction, 'submit');
    assert.equal(hitData[0].eventLabel, 'www.google-analytics.com');
    assert.equal(hitData[0].nonInteraction, true);
  });


  it('should support specifying a hit filter', function() {

    var hitData = browser
        .execute(utilities.stopSubmitEvents)
        .execute(utilities.stubBeacon)
        .execute(requireOutboundFormTracker_hitFilter)
        .click('#outbound-submit')
        .click('#outbound-submit-with-class')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].eventCategory, 'Outbound Form');
    assert.equal(hitData[0].eventAction, 'submit');
    assert.equal(hitData[0].eventLabel, 'https://example.com/');
    assert.equal(hitData[0].nonInteraction, true);
  });


  it('should support forms in shadow DOM and event retargetting', function() {

    if (notSupportedInBrowser()) return;

    var hitData = browser
        .execute(utilities.stopSubmitEvents)
        .execute(utilities.stubBeacon)
        .execute(ga.run, 'require', 'outboundFormTracker')
        .execute(simulateSubmitFromInsideShadowDom)
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].eventCategory, 'Outbound Form');
    assert.equal(hitData[0].eventAction, 'submit');
    assert.equal(hitData[0].eventLabel, 'https://example.com/');
  });


  it('includes usage params with all hits', function() {

    var hitData = browser
        .execute(ga.run, 'require', 'outboundFormTracker')
        .execute(ga.run, 'send', 'pageview')
        .execute(ga.getHitData)
        .value;

    assert.equal(hitData.length, 1);
    assert.equal(hitData[0].devId, constants.DEV_ID);
    assert.equal(hitData[0][constants.VERSION_PARAM], constants.VERSION);

    // '10' = '000010000' in hex
    assert.equal(hitData[0][constants.USAGE_PARAM], '10');
  });

});


/**
 * Navigates to the outbound form tracker test page.
 */
function setupPage() {
  browser.url('/test/outbound-form-tracker.html');
}


/**
 * Initiates the tracker and capturing hit data.
 */
function startTracking() {
  browser
      .execute(ga.run, 'create', 'UA-XXXXX-Y', 'auto')
      .execute(ga.trackHitData);
}


/**
 * Stops capturing hit data and remove the plugin and tracker.
 */
function stopTracking() {
  browser
      .execute(utilities.unstopSubmitEvents)
      .execute(ga.clearHitData)
      .execute(ga.run, 'outboundFormTracker:remove')
      .execute(ga.run, 'remove');
}


/**
 * @return {boolean} True if the current browser doesn't support all features
 *    required for these tests.
 */
function notSupportedInBrowser() {
  return browser.execute(function() {
    return !Element.prototype.attachShadow;
  }).value;
}


/**
 * Since function objects can't be passed via parameters from server to
 * client, this one-off function must be used to set the value for
 * `shouldTrackOutboundForm`.
 */
function requireOutboundFormTracker_shouldTrackOutboundForm() {
  ga('require', 'outboundFormTracker', {
    shouldTrackOutboundForm: function(form, parseUrl) {
      return parseUrl(form.action).hostname == 'example.com';
    }
  });
}


/**
 * Since function objects can't be passed via parameters from server to
 * client, this one-off function must be used to set the value for
 * `hitFilter`.
 */
function requireOutboundFormTracker_hitFilter() {
  ga('require', 'outboundFormTracker', {
    hitFilter: function(model, form) {
      if (form.action.indexOf('www.google-analytics.com') > -1) {
        throw 'Exclude hits to www.google-analytics.com';
      }
      else {
        model.set('nonInteraction', true);
      }
    }
  });
}


/**
 * Webdriver does not currently support selecting elements inside a shadow
 * tree, so we have to fake it.
 */
function simulateSubmitFromInsideShadowDom() {
  var shadowHost = document.getElementById('shadow-host');
  var form = shadowHost.shadowRoot.querySelector('form');

  var event = document.createEvent('Event');
  event.initEvent('submit', true, true);
  form.dispatchEvent(event);
}
