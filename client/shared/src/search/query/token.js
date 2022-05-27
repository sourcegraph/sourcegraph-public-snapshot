"use strict";
exports.__esModule = true;
exports.createLiteral = exports.KeywordKind = exports.PatternKind = void 0;
/**
 * A label associated with a pattern token. We don't use SearchPatternType because
 * that is used as a global quantifier for all patterns in a query. PatternKind
 * allows to qualify multiple pattern tokens differently within a single query.
 */
var PatternKind;
(function (PatternKind) {
    PatternKind[PatternKind["Literal"] = 1] = "Literal";
    PatternKind[PatternKind["Regexp"] = 2] = "Regexp";
    PatternKind[PatternKind["Structural"] = 3] = "Structural";
})(PatternKind = exports.PatternKind || (exports.PatternKind = {}));
var KeywordKind;
(function (KeywordKind) {
    KeywordKind["Or"] = "or";
    KeywordKind["And"] = "and";
    KeywordKind["Not"] = "not";
})(KeywordKind = exports.KeywordKind || (exports.KeywordKind = {}));
var createLiteral = function (value, range, quoted) {
    if (quoted === void 0) { quoted = false; }
    return ({
        type: 'literal',
        value: value,
        range: range,
        quoted: quoted
    });
};
exports.createLiteral = createLiteral;
