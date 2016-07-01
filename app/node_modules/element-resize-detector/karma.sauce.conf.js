"use strict";

var sharedConfig = require("./karma.conf.js");

module.exports = function(config) {
    sharedConfig(config);

    // define SL browsers
    var customLaunchers = {

        //Chrome
        "SL_CHROME_LATEST_OSX": {
            base: "SauceLabs",
            platform: "Mac 10.9",
            browserName: "chrome"
        },
        "SL_CHROME_LATEST_WINDOWS": {
            base: "SauceLabs",
            platform: "Windows 8.1",
            browserName: "chrome"
        },
        "SL_CHROME_LATEST_LINUX": {
            base: "SauceLabs",
            platform: "Linux",
            browserName: "chrome"
        },

        //Firefox
        "SL_FIREFOX_LATEST_OSX": {
            base: "SauceLabs",
            platform: "Mac 10.9",
            browserName: "firefox"
        },
        "SL_FIREFOX_LATEST_WINDOWS": {
            base: "SauceLabs",
            platform: "Windows 8.1",
            browserName: "firefox"
        },
        "SL_FIREFOX_LATEST_LINUX": {
            base: "SauceLabs",
            platform: "Linux",
            browserName: "firefox"
        },

        //Safari
        "SL_SAFARI_LATEST_OSX": {
            base: "SauceLabs",
            platform: "Mac 10.9",
            browserName: "safari"
        },
        "SL_SAFARI_LATEST_WINDOWS": {
            base: "SauceLabs",
            platform: "Windows 7",
            browserName: "safari"
        },

        //IE
        "SL_IE_LATEST_WINDOWS": {
            base: "SauceLabs",
            platform: "Windows 8.1",
            browserName: "internet explorer"
        },
        "SL_IE_10_WINDOWS": {
            base: "SauceLabs",
            platform: "Windows 7",
            browserName: "internet explorer",
            version: "10"
        },
        "SL_IE_9_WINDOWS": {
            base: "SauceLabs",
            platform: "Windows 7",
            browserName: "internet explorer",
            version: "9"
        },
        "SL_IE_8_WINDOWS": {
            base: "SauceLabs",
            platform: "Windows xp",
            browserName: "internet explorer",
            version: "8"
        },

        //Opera
        "SL_OPERA_LATEST_WINDOWS": {
            base: "SauceLabs",
            platform: "Windows 7",
            browserName: "opera"
        },
        "SL_OPERA_LATEST_LINUX": {
            base: "SauceLabs",
            platform: "Linux",
            browserName: "opera"
        },

        //iPhone,
        "SL_IOS_LATEST_IPHONE": {
            base: "SauceLabs",
            platform: "OS X 10.9",
            browserName: "iphone",
            version: "8" //Sauce defaults to 5 if this is omitted.
        },
        "SL_IOS_7_IPHONE": {
            base: "SauceLabs",
            platform: "OS X 10.9",
            browserName: "iphone",
            version: "7"
        },

        //iPad,
        "SL_IOS_LATEST_IPAD": {
            base: "SauceLabs",
            platform: "OS X 10.9",
            browserName: "ipad",
            version: "8" //Sauce defaults to 5 if this is omitted.
        },
        "SL_IOS_7_IPAD": {
            base: "SauceLabs",
            platform: "OS X 10.9",
            browserName: "ipad",
            version: "7"
        }
    };

    config.set({
        autoWatch: false,

        reporters: ["dots", "saucelabs"],

        // If browser does not capture in given timeout [ms], kill it
        captureTimeout: 5*60*1000,
        browserNoActivityTimeout: 60*1000,

        sauceLabs: {
            testName: "element-resize-detector",
            recordScreenshots: false,
            startConnect: false
        },

        customLaunchers: customLaunchers,
        singleRun: true
    });
};
