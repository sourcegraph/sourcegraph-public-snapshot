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


var browserCaps;


describe('autotrack', function() {

  before(function() {
    browserCaps = browser.session().value;
    browser.url('/test/autotrack.html');
  });


  afterEach(function() {
    browser
        .execute(utilities.untrackConsoleErrors)
        .execute(ga.run, 'remove');
  });


  it('should log a deprecation error when requiring autotrack directly',
      function() {

    if (notSupportedInBrowser()) return;

    var consoleErrors = browser
        .execute(utilities.trackConsoleErrors)
        .execute(ga.run, 'create', 'UA-XXXXX-Y', 'auto')
        .execute(ga.run, 'require', 'autotrack')
        .execute(utilities.getConsoleErrors)
        .value;

    assert(consoleErrors.length, 1);
    assert(consoleErrors[0][0].indexOf('https://goo.gl/sZ2WrW') > -1);
  });

});


/**
 * @return {boolean} True if the current browser doesn't support all features
 *    required for these tests.
 */
function notSupportedInBrowser() {
  // IE9 doesn't support `console.error`, so it's not tested.
  return browserCaps.browserName == 'MicrosoftEdge' ||
      (browserCaps.browserName == 'internet explorer' &&
          browserCaps.version == '9');
}
