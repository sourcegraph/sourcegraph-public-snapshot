"use strict";
function arrayify(arg) {
    if (Array.isArray(arg)) {
        return arg;
    }
    else if (arg != null) {
        return [arg];
    }
    else {
        return [];
    }
}
exports.arrayify = arrayify;
function objectify(arg) {
    if (typeof arg === "object" && arg != null) {
        return arg;
    }
    else {
        return {};
    }
}
exports.objectify = objectify;
function dedent(strings) {
    var values = [];
    for (var _i = 1; _i < arguments.length; _i++) {
        values[_i - 1] = arguments[_i];
    }
    var fullString = strings.reduce(function (accumulator, str, i) {
        return accumulator + values[i - 1] + str;
    });
    var match = fullString.match(/^[ \t]*(?=\S)/gm);
    if (!match) {
        return fullString;
    }
    var indent = Math.min.apply(Math, match.map(function (el) { return el.length; }));
    var regexp = new RegExp("^[ \\t]{" + indent + "}", "gm");
    fullString = indent > 0 ? fullString.replace(regexp, "") : fullString;
    return fullString;
}
exports.dedent = dedent;
function stripComments(content) {
    var regexp = /("(?:[^\\\"]*(?:\\.)?)*")|('(?:[^\\\']*(?:\\.)?)*')|(\/\*(?:\r?\n|.)*?\*\/)|(\/{2,}.*?(?:(?:\r?\n)|$))/g;
    var result = content.replace(regexp, function (match, m1, m2, m3, m4) {
        if (m3) {
            return "";
        }
        else if (m4) {
            var length_1 = m4.length;
            if (length_1 > 2 && m4[length_1 - 1] === "\n") {
                return m4[length_1 - 2] === "\r" ? "\r\n" : "\n";
            }
            else {
                return "";
            }
        }
        else {
            return match;
        }
    });
    return result;
}
exports.stripComments = stripComments;
;
