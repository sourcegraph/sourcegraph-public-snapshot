"use strict";
exports.rules = {
    "adjacent-overload-signatures": true,
    "no-unsafe-finally": true,
    "object-literal-key-quotes": [true, "as-needed"],
    "object-literal-shorthand": true,
    "only-arrow-functions": [true, "allow-declarations"],
    "ordered-imports": [true, {
            "import-sources-order": "case-insensitive",
            "named-imports-order": "lowercase-last",
        }],
};
var xtends = "tslint:recommended";
exports.extends = xtends;
