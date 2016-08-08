"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var abstractFormatter_1 = require("../language/formatter/abstractFormatter");
var colors = require("colors");
var Formatter = (function (_super) {
    __extends(Formatter, _super);
    function Formatter() {
        _super.apply(this, arguments);
    }
    Formatter.prototype.format = function (failures) {
        if (typeof failures[0] === "undefined") {
            return "\n";
        }
        var fileName = failures[0].getFileName();
        var positionMaxSize = this.getPositionMaxSize(failures);
        var ruleMaxSize = this.getRuleMaxSize(failures);
        var outputLines = [
            fileName,
        ];
        for (var _i = 0, failures_1 = failures; _i < failures_1.length; _i++) {
            var failure = failures_1[_i];
            var failureString = failure.getFailure();
            var ruleName = failure.getRuleName();
            ruleName = this.pad(ruleName, ruleMaxSize);
            ruleName = colors.yellow(ruleName);
            var lineAndCharacter = failure.getStartPosition().getLineAndCharacter();
            var positionTuple = (lineAndCharacter.line + 1) + ":" + (lineAndCharacter.character + 1);
            positionTuple = this.pad(positionTuple, positionMaxSize);
            positionTuple = colors.red(positionTuple);
            var output = positionTuple + "  " + ruleName + "  " + failureString;
            outputLines.push(output);
        }
        return outputLines.join("\n") + "\n\n";
    };
    Formatter.prototype.pad = function (str, len) {
        var padder = Array(len + 1).join(" ");
        return (str + padder).substring(0, padder.length);
    };
    Formatter.prototype.getPositionMaxSize = function (failures) {
        var positionMaxSize = 0;
        for (var _i = 0, failures_2 = failures; _i < failures_2.length; _i++) {
            var failure = failures_2[_i];
            var lineAndCharacter = failure.getStartPosition().getLineAndCharacter();
            var positionSize = ((lineAndCharacter.line + 1) + ":" + (lineAndCharacter.character + 1)).length;
            if (positionSize > positionMaxSize) {
                positionMaxSize = positionSize;
            }
        }
        return positionMaxSize;
    };
    Formatter.prototype.getRuleMaxSize = function (failures) {
        var ruleMaxSize = 0;
        for (var _i = 0, failures_3 = failures; _i < failures_3.length; _i++) {
            var failure = failures_3[_i];
            var ruleSize = failure.getRuleName().length;
            if (ruleSize > ruleMaxSize) {
                ruleMaxSize = ruleSize;
            }
        }
        return ruleMaxSize;
    };
    return Formatter;
}(abstractFormatter_1.AbstractFormatter));
exports.Formatter = Formatter;
