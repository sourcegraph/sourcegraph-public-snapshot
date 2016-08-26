"use strict";
var Replacement = (function () {
    function Replacement(innerStart, innerLength, innerText) {
        this.innerStart = innerStart;
        this.innerLength = innerLength;
        this.innerText = innerText;
    }
    Replacement.applyAll = function (content, replacements) {
        replacements.sort(function (a, b) { return b.end - a.end; });
        return replacements.reduce(function (text, r) { return r.apply(text); }, content);
    };
    Object.defineProperty(Replacement.prototype, "start", {
        get: function () {
            return this.innerStart;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(Replacement.prototype, "length", {
        get: function () {
            return this.innerLength;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(Replacement.prototype, "end", {
        get: function () {
            return this.innerStart + this.innerLength;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(Replacement.prototype, "text", {
        get: function () {
            return this.innerText;
        },
        enumerable: true,
        configurable: true
    });
    Replacement.prototype.apply = function (content) {
        return content.substring(0, this.start) + this.text + content.substring(this.start + this.length);
    };
    return Replacement;
}());
exports.Replacement = Replacement;
var Fix = (function () {
    function Fix(innerRuleName, innerReplacements) {
        this.innerRuleName = innerRuleName;
        this.innerReplacements = innerReplacements;
    }
    Fix.applyAll = function (content, fixes) {
        var replacements = [];
        for (var _i = 0, fixes_1 = fixes; _i < fixes_1.length; _i++) {
            var fix = fixes_1[_i];
            replacements = replacements.concat(fix.replacements);
        }
        return Replacement.applyAll(content, replacements);
    };
    Object.defineProperty(Fix.prototype, "ruleName", {
        get: function () {
            return this.innerRuleName;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(Fix.prototype, "replacements", {
        get: function () {
            return this.innerReplacements;
        },
        enumerable: true,
        configurable: true
    });
    Fix.prototype.apply = function (content) {
        return Replacement.applyAll(content, this.innerReplacements);
    };
    return Fix;
}());
exports.Fix = Fix;
var RuleFailurePosition = (function () {
    function RuleFailurePosition(position, lineAndCharacter) {
        this.position = position;
        this.lineAndCharacter = lineAndCharacter;
    }
    RuleFailurePosition.prototype.getPosition = function () {
        return this.position;
    };
    RuleFailurePosition.prototype.getLineAndCharacter = function () {
        return this.lineAndCharacter;
    };
    RuleFailurePosition.prototype.toJson = function () {
        return {
            character: this.lineAndCharacter.character,
            line: this.lineAndCharacter.line,
            position: this.position,
        };
    };
    RuleFailurePosition.prototype.equals = function (ruleFailurePosition) {
        var ll = this.lineAndCharacter;
        var rr = ruleFailurePosition.lineAndCharacter;
        return this.position === ruleFailurePosition.position
            && ll.line === rr.line
            && ll.character === rr.character;
    };
    return RuleFailurePosition;
}());
exports.RuleFailurePosition = RuleFailurePosition;
var RuleFailure = (function () {
    function RuleFailure(sourceFile, start, end, failure, ruleName, fix) {
        this.sourceFile = sourceFile;
        this.failure = failure;
        this.ruleName = ruleName;
        this.fix = fix;
        this.fileName = sourceFile.fileName;
        this.startPosition = this.createFailurePosition(start);
        this.endPosition = this.createFailurePosition(end);
    }
    RuleFailure.prototype.getFileName = function () {
        return this.fileName;
    };
    RuleFailure.prototype.getRuleName = function () {
        return this.ruleName;
    };
    RuleFailure.prototype.getStartPosition = function () {
        return this.startPosition;
    };
    RuleFailure.prototype.getEndPosition = function () {
        return this.endPosition;
    };
    RuleFailure.prototype.getFailure = function () {
        return this.failure;
    };
    RuleFailure.prototype.hasFix = function () {
        return this.fix !== undefined;
    };
    RuleFailure.prototype.getFix = function () {
        return this.fix;
    };
    RuleFailure.prototype.toJson = function () {
        return {
            endPosition: this.endPosition.toJson(),
            failure: this.failure,
            fix: this.fix,
            name: this.fileName,
            ruleName: this.ruleName,
            startPosition: this.startPosition.toJson(),
        };
    };
    RuleFailure.prototype.equals = function (ruleFailure) {
        return this.failure === ruleFailure.getFailure()
            && this.fileName === ruleFailure.getFileName()
            && this.startPosition.equals(ruleFailure.getStartPosition())
            && this.endPosition.equals(ruleFailure.getEndPosition());
    };
    RuleFailure.prototype.createFailurePosition = function (position) {
        var lineAndCharacter = this.sourceFile.getLineAndCharacterOfPosition(position);
        return new RuleFailurePosition(position, lineAndCharacter);
    };
    return RuleFailure;
}());
exports.RuleFailure = RuleFailure;
