// webpack entry point

require("babel-polyfill");

window.jQuery = window.$ = require("jquery");

require("bootstrap-sass/assets/javascripts/bootstrap/transition.js");
require("bootstrap-sass/assets/javascripts/bootstrap/collapse.js");
require("bootstrap-sass/assets/javascripts/bootstrap/button.js");
require("bootstrap-sass/assets/javascripts/bootstrap/dropdown.js");
require("bootstrap-sass/assets/javascripts/bootstrap/modal.js");

require("./auth");
require("./appdash");
require("./invite");

// Views
require("./componentInjection");

require("./feedback-form");
require("./history");
require("./links");

require("../style/web.scss");

require("sourcegraph/util/actionLogger");
