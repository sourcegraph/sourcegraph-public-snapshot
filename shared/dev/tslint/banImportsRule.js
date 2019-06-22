"use strict";
var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
var __makeTemplateObject = (this && this.__makeTemplateObject) || function (cooked, raw) {
    if (Object.defineProperty) { Object.defineProperty(cooked, "raw", { value: raw }); } else { cooked.raw = raw; }
    return cooked;
};
exports.__esModule = true;
/**
 * @license
 * Copyright 2018 Palantir Technologies, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
var Lint = require("tslint");
var tsutils_1 = require("tsutils");
var Rule = /** @class */ (function (_super) {
    __extends(Rule, _super);
    function Rule() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    /* tslint:enable:object-literal-sort-keys */
    Rule.FAILURE_STRING_FACTORY = function (pattern, messageAddition) {
        return "Import of module matching pattern '" + pattern + "' is banned." + (messageAddition !== undefined ? " " + messageAddition : '');
    };
    Rule.prototype.apply = function (sourceFile) {
        return this.applyWithFunction(sourceFile, walk, [parseOption(this.ruleArguments[0], this.ruleArguments[1])]);
    };
    Rule.metadata = {
        ruleName: 'ban-imports',
        description: Lint.Utils.dedent(templateObject_1 || (templateObject_1 = __makeTemplateObject(["\n            Bans specific modules from being imported."], ["\n            Bans specific modules from being imported."]))),
        options: {
            type: 'list',
            listType: {
                type: 'array',
                items: { type: 'string' },
                minLength: 1,
                maxLength: 2
            }
        },
        optionsDescription: Lint.Utils.dedent(templateObject_2 || (templateObject_2 = __makeTemplateObject(["\n            A list of `[\"regex\", \"optional explanation here\"]`, which bans\n            imports that match `regex`"], ["\n            A list of \\`[\"regex\", \"optional explanation here\"]\\`, which bans\n            imports that match \\`regex\\`"]))),
        optionExamples: [[true, ['react-router-dom', 'Use {} instead.'], ['String']]],
        type: 'typescript',
        typescriptOnly: false
    };
    return Rule;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
function parseOption(pattern, message) {
    return { message: message, pattern: new RegExp("" + pattern) };
}
function walk(ctx) {
    var _loop_1 = function (name_1) {
        var ban = ctx.options.find(function (_a) {
            var pattern = _a.pattern;
            return pattern.test(name_1.text);
        });
        if (ban) {
            ctx.addFailure(name_1.getStart(ctx.sourceFile) + 1, name_1.end - 1, Rule.FAILURE_STRING_FACTORY(ban.pattern.toString(), ban.message));
        }
    };
    for (var _i = 0, _a = tsutils_1.findImports(ctx.sourceFile, tsutils_1.ImportKind.All); _i < _a.length; _i++) {
        var name_1 = _a[_i];
        _loop_1(name_1);
    }
}
var templateObject_1, templateObject_2;
