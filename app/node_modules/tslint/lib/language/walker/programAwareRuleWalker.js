"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var ruleWalker_1 = require("./ruleWalker");
var ProgramAwareRuleWalker = (function (_super) {
    __extends(ProgramAwareRuleWalker, _super);
    function ProgramAwareRuleWalker(sourceFile, options, program) {
        _super.call(this, sourceFile, options);
        this.program = program;
        this.typeChecker = program.getTypeChecker();
    }
    ProgramAwareRuleWalker.prototype.getProgram = function () {
        return this.program;
    };
    ProgramAwareRuleWalker.prototype.getTypeChecker = function () {
        return this.typeChecker;
    };
    return ProgramAwareRuleWalker;
}(ruleWalker_1.RuleWalker));
exports.ProgramAwareRuleWalker = ProgramAwareRuleWalker;
