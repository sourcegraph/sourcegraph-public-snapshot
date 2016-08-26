/**
 * Enforces a single instance of the Raven client, and the
 * main entry point for Raven. If you are a consumer of the
 * Raven library, you SHOULD load this file (vs raven.js).
 **/

'use strict';

var RavenConstructor = require('./raven');

var _Raven = window.Raven;

var Raven = new RavenConstructor();

/*
 * Allow multiple versions of Raven to be installed.
 * Strip Raven from the global context and returns the instance.
 *
 * @return {Raven}
 */
Raven.noConflict = function () {
	window.Raven = _Raven;
	return Raven;
};

Raven.afterLoad();

module.exports = Raven;
