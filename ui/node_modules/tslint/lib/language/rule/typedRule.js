"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var abstractRule_1 = require("./abstractRule");
var TypedRule = (function (_super) {
    __extends(TypedRule, _super);
    function TypedRule() {
        _super.apply(this, arguments);
    }
    TypedRule.prototype.apply = function (sourceFile) {
        throw new Error(this.getOptions().ruleName + " requires type checking");
    };
    return TypedRule;
}(abstractRule_1.AbstractRule));
exports.TypedRule = TypedRule;
