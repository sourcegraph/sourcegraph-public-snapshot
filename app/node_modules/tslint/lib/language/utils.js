"use strict";
var path = require("path");
var ts = require("typescript");
function getSourceFile(fileName, source) {
    var normalizedName = path.normalize(fileName).replace(/\\/g, "/");
    var compilerOptions = createCompilerOptions();
    var compilerHost = {
        fileExists: function () { return true; },
        getCanonicalFileName: function (filename) { return filename; },
        getCurrentDirectory: function () { return ""; },
        getDefaultLibFileName: function () { return "lib.d.ts"; },
        getNewLine: function () { return "\n"; },
        getSourceFile: function (filenameToGet) {
            if (filenameToGet === normalizedName) {
                return ts.createSourceFile(filenameToGet, source, compilerOptions.target, true);
            }
        },
        readFile: function () { return null; },
        useCaseSensitiveFileNames: function () { return true; },
        writeFile: function () { return null; },
    };
    var program = ts.createProgram([normalizedName], compilerOptions, compilerHost);
    return program.getSourceFile(normalizedName);
}
exports.getSourceFile = getSourceFile;
function createCompilerOptions() {
    return {
        noResolve: true,
        target: ts.ScriptTarget.ES5,
    };
}
exports.createCompilerOptions = createCompilerOptions;
function doesIntersect(failure, disabledIntervals) {
    return disabledIntervals.some(function (interval) {
        var maxStart = Math.max(interval.startPosition, failure.getStartPosition().getPosition());
        var minEnd = Math.min(interval.endPosition, failure.getEndPosition().getPosition());
        return maxStart <= minEnd;
    });
}
exports.doesIntersect = doesIntersect;
function scanAllTokens(scanner, callback) {
    var lastStartPos = -1;
    while (scanner.scan() !== ts.SyntaxKind.EndOfFileToken) {
        var startPos = scanner.getStartPos();
        if (startPos === lastStartPos) {
            break;
        }
        lastStartPos = startPos;
        callback(scanner);
    }
}
exports.scanAllTokens = scanAllTokens;
function hasModifier(modifiers) {
    var modifierKinds = [];
    for (var _i = 1; _i < arguments.length; _i++) {
        modifierKinds[_i - 1] = arguments[_i];
    }
    if (modifiers == null || modifierKinds == null) {
        return false;
    }
    return modifiers.some(function (m) {
        return modifierKinds.some(function (k) { return m.kind === k; });
    });
}
exports.hasModifier = hasModifier;
function isBlockScopedVariable(node) {
    var parentNode = (node.kind === ts.SyntaxKind.VariableDeclaration)
        ? node.parent
        : node.declarationList;
    return isNodeFlagSet(parentNode, ts.NodeFlags.Let)
        || isNodeFlagSet(parentNode, ts.NodeFlags.Const);
}
exports.isBlockScopedVariable = isBlockScopedVariable;
function isBlockScopedBindingElement(node) {
    var variableDeclaration = getBindingElementVariableDeclaration(node);
    return (variableDeclaration == null) || isBlockScopedVariable(variableDeclaration);
}
exports.isBlockScopedBindingElement = isBlockScopedBindingElement;
function getBindingElementVariableDeclaration(node) {
    var currentParent = node.parent;
    while (currentParent.kind !== ts.SyntaxKind.VariableDeclaration) {
        if (currentParent.parent == null) {
            return null;
        }
        else {
            currentParent = currentParent.parent;
        }
    }
    return currentParent;
}
exports.getBindingElementVariableDeclaration = getBindingElementVariableDeclaration;
function isNodeFlagSet(node, flagToCheck) {
    return (node.flags & flagToCheck) !== 0;
}
exports.isNodeFlagSet = isNodeFlagSet;
function isNestedModuleDeclaration(decl) {
    return decl.name.pos === decl.pos;
}
exports.isNestedModuleDeclaration = isNestedModuleDeclaration;
