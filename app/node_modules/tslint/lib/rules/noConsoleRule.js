"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var BanRule = require("./banRule");
var Rule = (function (_super) {
    __extends(Rule, _super);
    function Rule() {
        _super.apply(this, arguments);
    }
    Rule.prototype.apply = function (sourceFile) {
        var options = this.getOptions();
        var consoleBanWalker = new BanRule.BanFunctionWalker(sourceFile, this.getOptions());
        for (var _i = 0, _a = options.ruleArguments; _i < _a.length; _i++) {
            var option = _a[_i];
            consoleBanWalker.addBannedFunction(["console", option]);
        }
        return this.applyWithWalker(consoleBanWalker);
    };
    Rule.metadata = {
        ruleName: "no-console",
        description: "Bans the use of specified `console` methods.",
        rationale: "In general, \`console\` methods aren't appropriate for production code.",
        optionsDescription: "A list of method names to ban.",
        options: {
            type: "array",
            items: { type: "string" },
        },
        optionExamples: ["[true, \"log\", \"error\"]"],
        type: "functionality",
    };
    return Rule;
}(BanRule.Rule));
exports.Rule = Rule;
