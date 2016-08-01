"use strict";
function errorComparator(err1, err2) {
    if (err1.startPos.line !== err2.startPos.line) {
        return err1.startPos.line - err2.startPos.line;
    }
    else if (err1.startPos.col !== err2.startPos.col) {
        return err1.startPos.col - err2.startPos.col;
    }
    else if (err1.endPos.line !== err2.endPos.line) {
        return err1.endPos.line - err2.endPos.line;
    }
    else if (err1.endPos.col !== err2.endPos.col) {
        return err1.endPos.col - err2.endPos.col;
    }
    else {
        return err1.message.localeCompare(err2.message);
    }
}
exports.errorComparator = errorComparator;
function lintSyntaxError(message) {
    return new Error("Lint File Syntax Error: " + message);
}
exports.lintSyntaxError = lintSyntaxError;
