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


var SauceLabs = require('saucelabs');


// When running on CI, this will be true
var isSauceLabs = process.env.SAUCE_USERNAME && process.env.SAUCE_ACCESS_KEY;


// https://wiki.saucelabs.com/display/DOCS/Platform+Configurator#/
var capabilities = [
  {browserName: 'chrome'},
  {browserName: 'firefox'}
];

if (isSauceLabs) {
  capabilities = [
    {
      browserName: 'chrome',
      platform: 'Windows 10'
    },
    {
      browserName: 'firefox',
      platform: 'OS X 10.11'
    },
    {
      browserName: 'safari',
      platform: 'OS X 10.11',
      version: '9.0',
    },
    {
      browserName: 'safari',
      platform: 'OS X 10.8',
      version: '6'
    },
    {
      browserName: 'MicrosoftEdge',
      platform: 'Windows 10'
    },
    {
      browserName: 'internet explorer',
      platform: 'Windows 8.1',
      version: '11'
    },
    {
      browserName: 'internet explorer',
      platform: 'Windows 8',
      version: '10'
    },
    {
      browserName: 'internet explorer',
      platform: 'Windows 7',
      version: '9'
    }
  ];

  capabilities.forEach(function(cap) {
    cap['name'] = 'analytics.js autotrack tests - ' + cap.browserName +
                  ' - ' + (cap.version || 'latest');

    cap['build'] = process.env.TRAVIS_BUILD_NUMBER;
    cap['tunnel-identifier'] = process.env.TRAVIS_JOB_NUMBER;
  });
}

exports.config = {

  user: process.env.SAUCE_USERNAME,
  key:  process.env.SAUCE_ACCESS_KEY,
  // updateJob: true,

  //
  // ==================
  // Specify Test Files
  // ==================
  // Define which test specs should run. The pattern is relative to the directory
  // from which `wdio` was called. Notice that, if you are calling `wdio` from an
  // NPM script (see https://docs.npmjs.com/cli/run-script) then the current working
  // directory is where your package.json resides, so `wdio` will be called from there.
  //
  specs: [
    './test/*.js'
  ],
  // Patterns to exclude.
  exclude: [
    // 'path/to/excluded/files'
  ],
  //
  // ============
  // Capabilities
  // ============
  // Define your capabilities here. WebdriverIO can run multiple capabilties at the same
  // time. Depending on the number of capabilities, WebdriverIO launches several test
  // sessions. Within your capabilities you can overwrite the spec and exclude option in
  // order to group specific specs to a specific capability.
  //
  // If you have trouble getting all important capabilities together, check out the
  // Sauce Labs platform configurator - a great tool to configure your capabilities:
  // https://docs.saucelabs.com/reference/platforms-configurator
  //
  capabilities: capabilities,
  //
  // ===================
  // Test Configurations
  // ===================
  // Define all options that are relevant for the WebdriverIO instance here
  //
  // Level of logging verbosity: silent | verbose | command | data | result | error
  logLevel: 'silent',
  //
  // Enables colors for log output.
  coloredLogs: true,
  //
  // Set a base URL in order to shorten url command calls. If your url parameter starts
  // with "/", the base url gets prepended.
  baseUrl: 'http://localhost:8080',
  //
  // Default timeout for all waitForXXX commands.
  waitforTimeout: process.env.CI ? 60000 : 5000,
  //
  // Initialize the browser instance with a WebdriverIO plugin. The object should have the
  // plugin name as key and the desired plugin options as property. Make sure you have
  // the plugin installed before running any tests. The following plugins are currently
  // available:
  // WebdriverCSS: https://github.com/webdriverio/webdrivercss
  // WebdriverRTC: https://github.com/webdriverio/webdriverrtc
  // Browserevent: https://github.com/webdriverio/browserevent
  // plugins: {
  //     webdrivercss: {
  //         screenshotRoot: 'my-shots',
  //         failedComparisonsRoot: 'diffs',
  //         misMatchTolerance: 0.05,
  //         screenWidth: [320,480,640,1024]
  //     },
  //     webdriverrtc: {},
  //     browserevent: {}
  // },
  //
  // Framework you want to run your specs with.
  // The following are supported: mocha, jasmine and cucumber
  // see also: http://webdriver.io/guide/testrunner/frameworks.html
  //
  // Make sure you have the node package for the specific framework installed before running
  // any tests. If not please install the following package:
  // Mocha: `$ npm install mocha`
  // Jasmine: `$ npm install jasmine`
  // Cucumber: `$ npm install cucumber`
  framework: 'mocha',
  //
  // Test reporter for stdout.
  // The following are supported: dot (default), spec and xunit
  // see also: http://webdriver.io/guide/testrunner/reporters.html
  reporter: 'spec',

  //
  // Options to be passed to Mocha.
  // See the full list at http://mochajs.org/
  mochaOpts: {
    ui: 'bdd',
    timeout: 60000
  },

  //
  // =====
  // Hooks
  // =====
  // Run functions before or after the test. If one of them returns with a promise, WebdriverIO
  // will wait until that promise got resolved to continue.
  //
  // Gets executed before all workers get launched.
  onPrepare: function() {
    // do something
  },
  //
  // Gets executed before test execution begins. At this point you will have access to all global
  // variables like `browser`. It is the perfect place to define custom commands.
  before: function() {
    // do something
  },
  //
  // Gets executed after all tests are done. You still have access to all global variables from
  // the test.
  after: function(failures, sessionId) {
    // When runnign in SauceLabs, update the job with the test status.
    if (process.env.SAUCE_USERNAME && process.env.SAUCE_ACCESS_KEY) {
      var sauceAccount = new SauceLabs({
        username: process.env.SAUCE_USERNAME,
        password: process.env.SAUCE_ACCESS_KEY
      });
      return new Promise(function(resolve, reject) {
        var data = {
          passed: !failures,
          build: process.env.TRAVIS_BUILD_NUMBER
        };
        sauceAccount.updateJob(sessionId, data, function(err, res) {
          if (err) {
            reject(err);
          }
          else {
            resolve();
          }
        });
      });
    }
  },
  //
  // Gets executed after all workers got shut down and the process is about to exit. It is not
  // possible to defer the end of the process using a promise.
  onComplete: function() {
    // do something
  }
};
