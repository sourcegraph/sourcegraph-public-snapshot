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


/* eslint no-console: ["error", {allow: ["error"]}] */


// Imports dependencies.
var provide = require('../provide');


/**
 * Warns users that this functionality is deprecated as of version 1.0
 * @deprecated
 * @constructor
 */
function Autotrack() {
  console.error('Requiring the `autotrack` plugin no longer requires all ' +
      'sub-plugins be default. See https://goo.gl/sZ2WrW for details.');
}


provide('autotrack', Autotrack);
