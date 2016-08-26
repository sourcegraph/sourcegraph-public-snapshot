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


// Imports all sub-plugins.
require('./plugins/clean-url-tracker');
require('./plugins/event-tracker');
require('./plugins/impression-tracker');
require('./plugins/media-query-tracker');
require('./plugins/outbound-form-tracker');
require('./plugins/outbound-link-tracker');
require('./plugins/page-visibility-tracker');
require('./plugins/social-widget-tracker');
require('./plugins/url-change-tracker');

// Imports the deprecated autotrack plugin for backwards compatibility.
require('./plugins/autotrack');
