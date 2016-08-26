"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var Lint = require("../lint");
var Rule = (function (_super) {
    __extends(Rule, _super);
    function Rule() {
        _super.apply(this, arguments);
    }
    Rule.prototype.apply = function (sourceFile) {
        var walker = new FileHeaderWalker(sourceFile, this.getOptions());
        var options = this.getOptions().ruleArguments;
        walker.setRegexp(new RegExp(options[0].toString()));
        return this.applyWithWalker(walker);
    };
    Rule.metadata = {
        ruleName: "file-header",
        description: "Enforces a certain header comment for all files, matched by a regular expression.",
        optionsDescription: "Regular expression to match the header.",
        options: {
            type: "string",
        },
        optionExamples: ['"true", "Copyright \\d{4}"'],
        type: "style",
    };
    Rule.FAILURE_STRING = "missing file header";
    return Rule;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var FileHeaderWalker = (function (_super) {
    __extends(FileHeaderWalker, _super);
    function FileHeaderWalker() {
        _super.apply(this, arguments);
        this.commentRegexp = /^\s*(\/\/(.*?)|\/\*([\s\S]*?)\*\/)/;
    }
    FileHeaderWalker.prototype.setRegexp = function (headerRegexp) {
        this.headerRegexp = headerRegexp;
    };
    FileHeaderWalker.prototype.visitSourceFile = function (node) {
        if (this.headerRegexp) {
            var text = node.getFullText();
            var offset = 0;
            if (text.indexOf("#!") === 0) {
                offset = text.indexOf("\n") + 1;
                text = text.substring(offset);
            }
            var match = text.match(this.commentRegexp);
            if (!match) {
                this.addFailure(this.createFailure(offset, 0, Rule.FAILURE_STRING));
            }
            else {
                var comment = match[2] ? match[2] : match[3];
                if (comment.search(this.headerRegexp) < 0) {
                    this.addFailure(this.createFailure(offset, 0, Rule.FAILURE_STRING));
                }
            }
        }
    };
    return FileHeaderWalker;
}(Lint.RuleWalker));
