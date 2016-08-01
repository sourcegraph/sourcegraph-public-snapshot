"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var abstractFormatter_1 = require("../language/formatter/abstractFormatter");
var Formatter = (function (_super) {
    __extends(Formatter, _super);
    function Formatter() {
        _super.apply(this, arguments);
    }
    Formatter.prototype.format = function (failures) {
        var output = "<pmd version=\"tslint\">";
        for (var _i = 0, failures_1 = failures; _i < failures_1.length; _i++) {
            var failure = failures_1[_i];
            var failureString = failure.getFailure()
                .replace(/&/g, "&amp;")
                .replace(/</g, "&lt;")
                .replace(/>/g, "&gt;")
                .replace(/'/g, "&#39;")
                .replace(/"/g, "&quot;");
            var lineAndCharacter = failure.getStartPosition().getLineAndCharacter();
            output += "<file name=\"" + failure.getFileName();
            output += "\"><violation begincolumn=\"" + (lineAndCharacter.character + 1);
            output += "\" beginline=\"" + (lineAndCharacter.line + 1);
            output += "\" priority=\"1\"";
            output += " rule=\"" + failureString + "\"> </violation></file>";
        }
        output += "</pmd>";
        return output;
    };
    return Formatter;
}(abstractFormatter_1.AbstractFormatter));
exports.Formatter = Formatter;
