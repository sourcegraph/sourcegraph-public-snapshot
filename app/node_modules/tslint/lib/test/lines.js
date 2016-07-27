"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var utils_1 = require("./utils");
var Line = (function () {
    function Line() {
    }
    return Line;
}());
exports.Line = Line;
var CodeLine = (function (_super) {
    __extends(CodeLine, _super);
    function CodeLine(contents) {
        _super.call(this);
        this.contents = contents;
    }
    return CodeLine;
}(Line));
exports.CodeLine = CodeLine;
var MessageSubstitutionLine = (function (_super) {
    __extends(MessageSubstitutionLine, _super);
    function MessageSubstitutionLine(key, message) {
        _super.call(this);
        this.key = key;
        this.message = message;
    }
    return MessageSubstitutionLine;
}(Line));
exports.MessageSubstitutionLine = MessageSubstitutionLine;
var ErrorLine = (function (_super) {
    __extends(ErrorLine, _super);
    function ErrorLine(startCol) {
        _super.call(this);
        this.startCol = startCol;
    }
    return ErrorLine;
}(Line));
exports.ErrorLine = ErrorLine;
var MultilineErrorLine = (function (_super) {
    __extends(MultilineErrorLine, _super);
    function MultilineErrorLine(startCol) {
        _super.call(this, startCol);
    }
    return MultilineErrorLine;
}(ErrorLine));
exports.MultilineErrorLine = MultilineErrorLine;
var EndErrorLine = (function (_super) {
    __extends(EndErrorLine, _super);
    function EndErrorLine(startCol, endCol, message) {
        _super.call(this, startCol);
        this.endCol = endCol;
        this.message = message;
    }
    return EndErrorLine;
}(ErrorLine));
exports.EndErrorLine = EndErrorLine;
var multilineErrorRegex = /^\s*(~+|~nil)$/;
var endErrorRegex = /^\s*(~+|~nil)\s*\[(.+)\]\s*$/;
var messageSubstitutionRegex = /^\[([\w\-\_]+?)]: \s*(.+?)\s*$/;
exports.ZERO_LENGTH_ERROR = "~nil";
function parseLine(text) {
    var multilineErrorMatch = text.match(multilineErrorRegex);
    if (multilineErrorMatch != null) {
        var startErrorCol = text.indexOf("~");
        return new MultilineErrorLine(startErrorCol);
    }
    var endErrorMatch = text.match(endErrorRegex);
    if (endErrorMatch != null) {
        var squiggles = endErrorMatch[1], message = endErrorMatch[2];
        var startErrorCol = text.indexOf("~");
        var zeroLengthError = (squiggles === exports.ZERO_LENGTH_ERROR);
        var endErrorCol = zeroLengthError ? startErrorCol : text.lastIndexOf("~") + 1;
        return new EndErrorLine(startErrorCol, endErrorCol, message);
    }
    var messageSubstitutionMatch = text.match(messageSubstitutionRegex);
    if (messageSubstitutionMatch != null) {
        var key = messageSubstitutionMatch[1], message = messageSubstitutionMatch[2];
        return new MessageSubstitutionLine(key, message);
    }
    return new CodeLine(text);
}
exports.parseLine = parseLine;
function printLine(line, code) {
    if (line instanceof ErrorLine) {
        if (code == null) {
            throw new Error("Must supply argument for code parameter when line is an ErrorLine");
        }
        var leadingSpaces = utils_1.replicateStr(" ", line.startCol);
        if (line instanceof MultilineErrorLine) {
            if (code.length === 0 && line.startCol === 0) {
                return exports.ZERO_LENGTH_ERROR;
            }
            var tildes = utils_1.replicateStr("~", code.length - leadingSpaces.length);
            return "" + leadingSpaces + tildes;
        }
        else if (line instanceof EndErrorLine) {
            var tildes = utils_1.replicateStr("~", line.endCol - line.startCol);
            var endSpaces = utils_1.replicateStr(" ", code.length - line.endCol);
            if (tildes.length === 0) {
                tildes = exports.ZERO_LENGTH_ERROR;
                endSpaces = endSpaces.substring(0, Math.max(endSpaces.length - 4, 1));
            }
            return "" + leadingSpaces + tildes + endSpaces + " [" + line.message + "]";
        }
    }
    else if (line instanceof MessageSubstitutionLine) {
        return "[" + line.key + "]: " + line.message;
    }
    else if (line instanceof CodeLine) {
        return line.contents;
    }
}
exports.printLine = printLine;
