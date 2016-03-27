#!/usr/bin/env node
/**
 * Copyright (c) 2015-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 */
'use strict';
var spawn = require('child_process').spawn;
var input = process.argv.slice(2);
var bin = require('./');

spawn(bin, input, {stdio: 'inherit'})
	.on('exit', process.exit);
