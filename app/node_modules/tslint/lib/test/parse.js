"use strict";
var lines_1 = require("./lines");
var lintError_1 = require("./lintError");
function removeErrorMarkup(text) {
    var textWithMarkup = text.split("\n");
    var lines = textWithMarkup.map(lines_1.parseLine);
    var codeText = lines.filter(function (line) { return (line instanceof lines_1.CodeLine); }).map(function (line) { return line.contents; });
    return codeText.join("\n");
}
exports.removeErrorMarkup = removeErrorMarkup;
function parseErrorsFromMarkup(text) {
    var textWithMarkup = text.split("\n");
    var lines = textWithMarkup.map(lines_1.parseLine);
    if (lines.length > 0 && !(lines[0] instanceof lines_1.CodeLine)) {
        throw lintError_1.lintSyntaxError("text cannot start with an error mark line.");
    }
    var messageSubstitutionLines = lines.filter(function (l) { return l instanceof lines_1.MessageSubstitutionLine; });
    var messageSubstitutions = {};
    for (var _i = 0, messageSubstitutionLines_1 = messageSubstitutionLines; _i < messageSubstitutionLines_1.length; _i++) {
        var line = messageSubstitutionLines_1[_i];
        messageSubstitutions[line.key] = line.message;
    }
    var errorLinesForCodeLines = createCodeLineNoToErrorsMap(lines);
    var lintErrors = [];
    errorLinesForCodeLines.forEach(function (errorLinesForLineOfCode, lineNo) {
        while (errorLinesForLineOfCode.length > 0) {
            var errorLine = errorLinesForLineOfCode.shift();
            var errorStartPos = { line: lineNo, col: errorLine.startCol };
            if (errorLine instanceof lines_1.EndErrorLine) {
                lintErrors.push({
                    startPos: errorStartPos,
                    endPos: { line: lineNo, col: errorLine.endCol },
                    message: messageSubstitutions[errorLine.message] || errorLine.message,
                });
            }
            else if (errorLine instanceof lines_1.MultilineErrorLine) {
                for (var nextLineNo = lineNo + 1;; ++nextLineNo) {
                    if (!isValidErrorMarkupContinuation(errorLinesForCodeLines, nextLineNo)) {
                        throw lintError_1.lintSyntaxError("Error mark starting at " + errorStartPos.line + ":" + errorStartPos.col + " does not end correctly.");
                    }
                    else {
                        var nextErrorLine = errorLinesForCodeLines[nextLineNo].shift();
                        if (nextErrorLine instanceof lines_1.EndErrorLine) {
                            lintErrors.push({
                                startPos: errorStartPos,
                                endPos: { line: nextLineNo, col: nextErrorLine.endCol },
                                message: messageSubstitutions[nextErrorLine.message] || nextErrorLine.message,
                            });
                            break;
                        }
                    }
                }
            }
        }
    });
    lintErrors.sort(lintError_1.errorComparator);
    return lintErrors;
}
exports.parseErrorsFromMarkup = parseErrorsFromMarkup;
function createMarkupFromErrors(code, lintErrors) {
    lintErrors.sort(lintError_1.errorComparator);
    var codeText = code.split("\n");
    var errorLinesForCodeText = codeText.map(function () { return []; });
    for (var _i = 0, lintErrors_1 = lintErrors; _i < lintErrors_1.length; _i++) {
        var error = lintErrors_1[_i];
        var startPos = error.startPos, endPos = error.endPos, message = error.message;
        if (startPos.line === endPos.line) {
            errorLinesForCodeText[startPos.line].push(new lines_1.EndErrorLine(startPos.col, endPos.col, message));
        }
        else {
            errorLinesForCodeText[startPos.line].push(new lines_1.MultilineErrorLine(startPos.col));
            for (var lineNo = startPos.line + 1; lineNo < endPos.line; ++lineNo) {
                errorLinesForCodeText[lineNo].push(new lines_1.MultilineErrorLine(0));
            }
            errorLinesForCodeText[endPos.line].push(new lines_1.EndErrorLine(0, endPos.col, message));
        }
    }
    var finalText = combineCodeTextAndErrorLines(codeText, errorLinesForCodeText);
    return finalText.join("\n");
}
exports.createMarkupFromErrors = createMarkupFromErrors;
function combineCodeTextAndErrorLines(codeText, errorLinesForCodeText) {
    return codeText.reduce(function (resultText, code, i) {
        resultText.push(code);
        resultText.push.apply(resultText, (errorLinesForCodeText[i].map(function (line) { return lines_1.printLine(line, code); })));
        return resultText;
    }, []);
}
function createCodeLineNoToErrorsMap(lines) {
    var errorLinesForCodeLine = [];
    for (var _i = 0, lines_2 = lines; _i < lines_2.length; _i++) {
        var line = lines_2[_i];
        if (line instanceof lines_1.CodeLine) {
            errorLinesForCodeLine.push([]);
        }
        else if (line instanceof lines_1.ErrorLine) {
            errorLinesForCodeLine[errorLinesForCodeLine.length - 1].push(line);
        }
    }
    return errorLinesForCodeLine;
}
function isValidErrorMarkupContinuation(errorLinesForCodeLines, lineNo) {
    return lineNo < errorLinesForCodeLines.length
        && errorLinesForCodeLines[lineNo].length !== 0
        && errorLinesForCodeLines[lineNo][0].startCol === 0;
}
