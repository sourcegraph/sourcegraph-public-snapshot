// Standard setup for next.js apps
const base = require("eslint-config-next/core-web-vitals");
// Add a simple rule that's easy to violate so we can demonstrate that linting is working
base["rules"] = {"no-debugger": 2};
module.exports = base;
