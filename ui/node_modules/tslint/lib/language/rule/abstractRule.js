"use strict";
var AbstractRule = (function () {
    function AbstractRule(ruleName, value, disabledIntervals) {
        this.value = value;
        var ruleArguments = [];
        if (Array.isArray(value) && value.length > 1) {
            ruleArguments = value.slice(1);
        }
        this.options = {
            disabledIntervals: disabledIntervals,
            ruleArguments: ruleArguments,
            ruleName: ruleName,
        };
    }
    AbstractRule.prototype.getOptions = function () {
        return this.options;
    };
    AbstractRule.prototype.applyWithWalker = function (walker) {
        walker.walk(walker.getSourceFile());
        return walker.getFailures();
    };
    AbstractRule.prototype.isEnabled = function () {
        var value = this.value;
        if (typeof value === "boolean") {
            return value;
        }
        if (Array.isArray(value) && value.length > 0) {
            return value[0];
        }
        return false;
    };
    return AbstractRule;
}());
exports.AbstractRule = AbstractRule;
