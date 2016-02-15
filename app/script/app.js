// webpack entry point

require("babel-polyfill");

window.jQuery = window.$ = require("jquery");

require("jquery.hotkeys/jquery.hotkeys");
require("bootstrap-sass/assets/javascripts/bootstrap/tab.js");
require("bootstrap-sass/assets/javascripts/bootstrap/tooltip.js");
require("bootstrap-sass/assets/javascripts/bootstrap/transition.js");
require("bootstrap-sass/assets/javascripts/bootstrap/collapse.js");
require("bootstrap-sass/assets/javascripts/bootstrap/button.js");
require("bootstrap-sass/assets/javascripts/bootstrap/dropdown.js");
require("bootstrap-sass/assets/javascripts/bootstrap/modal.js");
require("google-code-prettify/prettify");

require("./auth");
require("./appdash");
require("./globals");
require("./invite");
require("./mirror-repos");

// Dispatchers
require("./dispatchers/AppDispatcher");

// Stores
require("./stores/models/CodeLineModel");
require("./stores/collections/CodeLineCollection");

// Views
require("./componentInjection");

require("./activateDefnPopovers");
require("./buttons");
require("./debounce");
require("./defn-popover");
require("./feedback-form");
require("./history");
require("./keyboard_shortcuts");
require("./links");
require("./tooltip");
require("./syntax-highlight");

require("../style/web.scss");

require("sourcegraph/util/actionLogger");

window.Sourcegraph = require("sourcegraph/platform/api").default;
