/*!-----------------------------------------------------------
 * Copyright (c) Microsoft Corporation. All rights reserved.
 * Version: 0.5.3(793ede49d53dba79d39e52205f16321278f5183c)
 * Released under the MIT license
 * https://github.com/Microsoft/vscode/blob/master/LICENSE.txt
 *-----------------------------------------------------------*/

(function() {
var __m = ["exports","require","vs/languages/razor/common/razorTokenTypes","vs/languages/razor/common/vsxmlTokenTypes","vs/languages/razor/common/vsxml","vs/base/common/objects","vs/editor/common/modes/abstractMode","vs/editor/common/modes/abstractState","vs/languages/razor/common/csharpTokenization","vs/languages/html/common/html","vs/base/common/errors","vs/languages/razor/common/razor","vs/editor/common/modes","vs/platform/instantiation/common/instantiation","vs/editor/common/services/modeService","vs/editor/common/modes/languageConfigurationRegistry","vs/base/common/async","vs/editor/common/services/compatWorkerService"];
var __M = function(deps) {
  var result = [];
  for (var i = 0, len = deps.length; i < len; i++) {
    result[i] = __m[deps[i]];
  }
  return result;
};
define(__m[2], __M([1,0]), function (require, exports) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    exports.EMBED_CS = 'support.function.cshtml';
});

define(__m[3], __M([1,0]), function (require, exports) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    exports.TOKEN_VALUE = 'support.property-value.constant.other.json';
    exports.TOKEN_KEY = 'support.type.property-name.json';
});

/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Microsoft Corporation. All rights reserved.
 *  Licensed under the MIT License. See License.txt in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
/* In order to use VSXML in your own modes, you need to have an IState
 * which implements IVSXMLWrapperState. Upon a START token such as '///',
 * the wrapper state can return a new VSXMLEmbeddedState as the nextState in
 * the tokenization result.
*/
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
define(__m[4], __M([1,0,5,10,7,3]), function (require, exports, objects, errors, abstractState_1, vsxmlTokenTypes) {
    'use strict';
    var separators = '<>"=/';
    var whitespace = '\t ';
    var isEntity = objects.createKeywordMatcher(['summary', 'reference', 'returns', 'param', 'loc']);
    var isAttribute = objects.createKeywordMatcher(['type', 'path', 'name', 'locid', 'filename', 'format', 'optional']);
    var isSeparator = objects.createKeywordMatcher(separators.split(''));
    var EmbeddedState = (function (_super) {
        __extends(EmbeddedState, _super);
        function EmbeddedState(mode, state, parentState) {
            _super.call(this, mode);
            this.state = state;
            this.parentState = parentState;
        }
        EmbeddedState.prototype.getParentState = function () {
            return this.parentState;
        };
        EmbeddedState.prototype.makeClone = function () {
            return new EmbeddedState(this.getMode(), abstractState_1.AbstractState.safeClone(this.state), abstractState_1.AbstractState.safeClone(this.parentState));
        };
        EmbeddedState.prototype.equals = function (other) {
            if (other instanceof EmbeddedState) {
                return (_super.prototype.equals.call(this, other) &&
                    abstractState_1.AbstractState.safeEquals(this.state, other.state) &&
                    abstractState_1.AbstractState.safeEquals(this.parentState, other.parentState));
            }
            return false;
        };
        EmbeddedState.prototype.setState = function (nextState) {
            this.state = nextState;
        };
        EmbeddedState.prototype.postTokenize = function (result, stream) {
            return result;
        };
        EmbeddedState.prototype.tokenize = function (stream) {
            var result = this.state.tokenize(stream);
            if (result.nextState !== undefined) {
                this.setState(result.nextState);
            }
            result.nextState = this;
            return this.postTokenize(result, stream);
        };
        return EmbeddedState;
    }(abstractState_1.AbstractState));
    exports.EmbeddedState = EmbeddedState;
    var VSXMLEmbeddedState = (function (_super) {
        __extends(VSXMLEmbeddedState, _super);
        function VSXMLEmbeddedState(mode, state, parentState) {
            _super.call(this, mode, state, parentState);
        }
        VSXMLEmbeddedState.prototype.equals = function (other) {
            if (other instanceof VSXMLEmbeddedState) {
                return (_super.prototype.equals.call(this, other));
            }
            return false;
        };
        VSXMLEmbeddedState.prototype.setState = function (nextState) {
            _super.prototype.setState.call(this, nextState);
            this.getParentState().setVSXMLState(nextState);
        };
        VSXMLEmbeddedState.prototype.postTokenize = function (result, stream) {
            if (stream.eos()) {
                result.nextState = this.getParentState();
            }
            return result;
        };
        return VSXMLEmbeddedState;
    }(EmbeddedState));
    exports.VSXMLEmbeddedState = VSXMLEmbeddedState;
    var VSXMLState = (function (_super) {
        __extends(VSXMLState, _super);
        function VSXMLState(mode, name, parent, whitespaceTokenType) {
            if (whitespaceTokenType === void 0) { whitespaceTokenType = ''; }
            _super.call(this, mode);
            this.name = name;
            this.parent = parent;
            this.whitespaceTokenType = whitespaceTokenType;
        }
        VSXMLState.prototype.equals = function (other) {
            if (other instanceof VSXMLState) {
                return (_super.prototype.equals.call(this, other) &&
                    this.whitespaceTokenType === other.whitespaceTokenType &&
                    this.name === other.name &&
                    abstractState_1.AbstractState.safeEquals(this.parent, other.parent));
            }
            return false;
        };
        VSXMLState.prototype.tokenize = function (stream) {
            stream.setTokenRules(separators, whitespace);
            if (stream.skipWhitespace().length > 0) {
                return { type: this.whitespaceTokenType };
            }
            return this.stateTokenize(stream);
        };
        VSXMLState.prototype.stateTokenize = function (stream) {
            throw errors.notImplemented();
        };
        return VSXMLState;
    }(abstractState_1.AbstractState));
    exports.VSXMLState = VSXMLState;
    var VSXMLString = (function (_super) {
        __extends(VSXMLString, _super);
        function VSXMLString(mode, parent) {
            _super.call(this, mode, 'string', parent, vsxmlTokenTypes.TOKEN_VALUE);
        }
        VSXMLString.prototype.makeClone = function () {
            return new VSXMLString(this.getMode(), this.parent ? this.parent.clone() : null);
        };
        VSXMLString.prototype.equals = function (other) {
            if (other instanceof VSXMLString) {
                return (_super.prototype.equals.call(this, other));
            }
            return false;
        };
        VSXMLString.prototype.stateTokenize = function (stream) {
            while (!stream.eos()) {
                var token = stream.nextToken();
                if (token === '"') {
                    return { type: vsxmlTokenTypes.TOKEN_VALUE, nextState: this.parent };
                }
            }
            return { type: vsxmlTokenTypes.TOKEN_VALUE, nextState: this.parent };
        };
        return VSXMLString;
    }(VSXMLState));
    exports.VSXMLString = VSXMLString;
    var VSXMLTag = (function (_super) {
        __extends(VSXMLTag, _super);
        function VSXMLTag(mode, parent) {
            _super.call(this, mode, 'expression', parent, 'vs');
        }
        VSXMLTag.prototype.makeClone = function () {
            return new VSXMLTag(this.getMode(), this.parent ? this.parent.clone() : null);
        };
        VSXMLTag.prototype.equals = function (other) {
            if (other instanceof VSXMLTag) {
                return (_super.prototype.equals.call(this, other));
            }
            return false;
        };
        VSXMLTag.prototype.stateTokenize = function (stream) {
            var token = stream.nextToken();
            var tokenType = this.whitespaceTokenType;
            if (token === '>') {
                return { type: 'punctuation.vs', nextState: this.parent };
            }
            else if (token === '"') {
                return { type: vsxmlTokenTypes.TOKEN_VALUE, nextState: new VSXMLString(this.getMode(), this) };
            }
            else if (isEntity(token)) {
                tokenType = 'tag.vs';
            }
            else if (isAttribute(token)) {
                tokenType = vsxmlTokenTypes.TOKEN_KEY;
            }
            else if (isSeparator(token)) {
                tokenType = 'punctuation.vs';
            }
            return { type: tokenType, nextState: this };
        };
        return VSXMLTag;
    }(VSXMLState));
    exports.VSXMLTag = VSXMLTag;
    var VSXMLExpression = (function (_super) {
        __extends(VSXMLExpression, _super);
        function VSXMLExpression(mode, parent) {
            _super.call(this, mode, 'expression', parent, 'vs');
        }
        VSXMLExpression.prototype.makeClone = function () {
            return new VSXMLExpression(this.getMode(), this.parent ? this.parent.clone() : null);
        };
        VSXMLExpression.prototype.equals = function (other) {
            if (other instanceof VSXMLExpression) {
                return (_super.prototype.equals.call(this, other));
            }
            return false;
        };
        VSXMLExpression.prototype.stateTokenize = function (stream) {
            var token = stream.nextToken();
            if (token === '<') {
                return { type: 'punctuation.vs', nextState: new VSXMLTag(this.getMode(), this) };
            }
            return { type: this.whitespaceTokenType, nextState: this };
        };
        return VSXMLExpression;
    }(VSXMLState));
    exports.VSXMLExpression = VSXMLExpression;
});






define(__m[8], __M([1,0,5,9,4,7,6,2]), function (require, exports, objects, htmlMode, VSXML, abstractState_1, abstractMode_1, razorTokenTypes) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var htmlTokenTypes = htmlMode.htmlTokenTypes;
    var punctuations = '+-*%&|^~!=<>/?;:.,';
    var separators = '+-*/%&|^~!=<>(){}[]\"\'\\/?;:.,';
    var whitespace = '\t ';
    var brackets = (function () {
        var bracketsSource = [
            { tokenType: 'punctuation.bracket.cs', open: '{', close: '}' },
            { tokenType: 'punctuation.array.cs', open: '[', close: ']' },
            { tokenType: 'punctuation.parenthesis.cs', open: '(', close: ')' }
        ];
        var MAP = Object.create(null);
        for (var i = 0; i < bracketsSource.length; i++) {
            var bracket = bracketsSource[i];
            MAP[bracket.open] = {
                tokenType: bracket.tokenType,
            };
            MAP[bracket.close] = {
                tokenType: bracket.tokenType,
            };
        }
        return {
            stringIsBracket: function (text) {
                return !!MAP[text];
            },
            tokenTypeFromString: function (text) {
                return MAP[text].tokenType;
            }
        };
    })();
    var isKeyword = objects.createKeywordMatcher([
        'abstract', 'as', 'async', 'await', 'base', 'bool',
        'break', 'by', 'byte', 'case',
        'catch', 'char', 'checked', 'class',
        'const', 'continue', 'decimal', 'default',
        'delegate', 'do', 'double', 'descending',
        'explicit', 'event', 'extern', 'else',
        'enum', 'false', 'finally', 'fixed',
        'float', 'for', 'foreach', 'from',
        'goto', 'group', 'if', 'implicit',
        'in', 'int', 'interface', 'internal',
        'into', 'is', 'lock', 'long', 'nameof',
        'new', 'null', 'namespace', 'object',
        'operator', 'out', 'override', 'orderby',
        'params', 'private', 'protected', 'public',
        'readonly', 'ref', 'return', 'switch',
        'struct', 'sbyte', 'sealed', 'short',
        'sizeof', 'stackalloc', 'static', 'string',
        'select', 'this', 'throw', 'true',
        'try', 'typeof', 'uint', 'ulong',
        'unchecked', 'unsafe', 'ushort', 'using',
        'var', 'virtual', 'volatile', 'void', 'when',
        'while', 'where', 'yield',
        'model', 'inject' // Razor specific
    ]);
    var ispunctuation = function (character) {
        return punctuations.indexOf(character) > -1;
    };
    var CSState = (function (_super) {
        __extends(CSState, _super);
        function CSState(mode, name, parent) {
            _super.call(this, mode);
            this.name = name;
            this.parent = parent;
        }
        CSState.prototype.equals = function (other) {
            if (!_super.prototype.equals.call(this, other)) {
                return false;
            }
            var otherCSState = other;
            return (other instanceof CSState) && (this.getMode() === otherCSState.getMode()) && (this.name === otherCSState.name) && ((this.parent === null && otherCSState.parent === null) || (this.parent !== null && this.parent.equals(otherCSState.parent)));
        };
        CSState.prototype.tokenize = function (stream) {
            stream.setTokenRules(separators, whitespace);
            if (stream.skipWhitespace().length > 0) {
                return { type: '' };
            }
            return this.stateTokenize(stream);
        };
        CSState.prototype.stateTokenize = function (stream) {
            throw new Error('To be implemented');
        };
        return CSState;
    }(abstractState_1.AbstractState));
    exports.CSState = CSState;
    var CSString = (function (_super) {
        __extends(CSString, _super);
        function CSString(mode, parent, punctuation) {
            _super.call(this, mode, 'string', parent);
            this.isAtBeginning = true;
            this.punctuation = punctuation;
        }
        CSString.prototype.makeClone = function () {
            return new CSString(this.getMode(), this.parent ? this.parent.clone() : null, this.punctuation);
        };
        CSString.prototype.equals = function (other) {
            return _super.prototype.equals.call(this, other) && this.punctuation === other.punctuation;
        };
        CSString.prototype.tokenize = function (stream) {
            var readChars = this.isAtBeginning ? 1 : 0;
            this.isAtBeginning = false;
            while (!stream.eos()) {
                var c = stream.next();
                if (c === '\\') {
                    if (readChars === 0) {
                        if (stream.eos()) {
                            return { type: 'string.escape.cs' };
                        }
                        else {
                            stream.next();
                            if (stream.eos()) {
                                return { type: 'string.escape.cs', nextState: this.parent };
                            }
                            else {
                                return { type: 'string.escape.cs' };
                            }
                        }
                    }
                    else {
                        stream.goBack(1);
                        return { type: 'string.cs' };
                    }
                }
                else if (c === this.punctuation) {
                    break;
                }
                readChars += 1;
            }
            return { type: 'string.cs', nextState: this.parent };
        };
        return CSString;
    }(CSState));
    var CSVerbatimString = (function (_super) {
        __extends(CSVerbatimString, _super);
        function CSVerbatimString(mode, parent) {
            _super.call(this, mode, 'verbatimstring', parent);
        }
        CSVerbatimString.prototype.makeClone = function () {
            return new CSVerbatimString(this.getMode(), this.parent ? this.parent.clone() : null);
        };
        CSVerbatimString.prototype.tokenize = function (stream) {
            while (!stream.eos()) {
                var token = stream.next();
                if (token === '"') {
                    if (!stream.eos() && stream.peek() === '"') {
                        stream.next();
                    }
                    else {
                        return { type: 'string.cs', nextState: this.parent };
                    }
                }
            }
            return { type: 'string.cs' };
        };
        return CSVerbatimString;
    }(CSState));
    var CSNumber = (function (_super) {
        __extends(CSNumber, _super);
        function CSNumber(mode, parent, firstDigit) {
            _super.call(this, mode, 'number', parent);
            this.firstDigit = firstDigit;
        }
        CSNumber.prototype.makeClone = function () {
            return new CSNumber(this.getMode(), this.parent ? this.parent.clone() : null, this.firstDigit);
        };
        CSNumber.prototype.tokenize = function (stream) {
            var character = this.firstDigit;
            var base = 10, isDecimal = false, isExponent = false;
            if (character === '0' && !stream.eos()) {
                character = stream.peek();
                if (character === 'x') {
                    base = 16;
                }
                else if (character === '.') {
                    base = 10;
                }
                else {
                    return { type: 'number.cs', nextState: this.parent };
                }
                stream.next();
            }
            while (!stream.eos()) {
                character = stream.peek();
                if (abstractMode_1.isDigit(character, base)) {
                    stream.next();
                }
                else if (base === 10) {
                    if (character === '.' && !isExponent && !isDecimal) {
                        isDecimal = true;
                        stream.next();
                    }
                    else if (character.toLowerCase() === 'e' && !isExponent) {
                        isExponent = true;
                        stream.next();
                        if (!stream.eos() && stream.peek() === '-') {
                            stream.next();
                        }
                    }
                    else if (character.toLowerCase() === 'f' || character.toLowerCase() === 'd') {
                        stream.next();
                        break;
                    }
                    else {
                        break;
                    }
                }
                else {
                    break;
                }
            }
            var tokenType = 'number';
            if (base === 16) {
                tokenType += '.hex';
            }
            return { type: tokenType + '.cs', nextState: this.parent };
        };
        return CSNumber;
    }(CSState));
    // the multi line comment
    var CSComment = (function (_super) {
        __extends(CSComment, _super);
        function CSComment(mode, parent, commentChar) {
            _super.call(this, mode, 'comment', parent);
            this.commentChar = commentChar;
        }
        CSComment.prototype.makeClone = function () {
            return new CSComment(this.getMode(), this.parent ? this.parent.clone() : null, this.commentChar);
        };
        CSComment.prototype.tokenize = function (stream) {
            while (!stream.eos()) {
                var token = stream.next();
                if (token === '*' && !stream.eos() && !stream.peekWhitespace() && stream.peek() === this.commentChar) {
                    stream.next();
                    return { type: 'comment.cs', nextState: this.parent };
                }
            }
            return { type: 'comment.cs' };
        };
        return CSComment;
    }(CSState));
    exports.CSComment = CSComment;
    var CSStatement = (function (_super) {
        __extends(CSStatement, _super);
        function CSStatement(mode, parent, level, plevel, razorMode, expression, firstToken, firstTokenWasKeyword) {
            _super.call(this, mode, 'expression', parent);
            this.level = level;
            this.plevel = plevel;
            this.razorMode = razorMode;
            this.expression = expression;
            this.vsState = new VSXML.VSXMLExpression(mode, null);
            this.firstToken = firstToken;
            this.firstTokenWasKeyword = firstTokenWasKeyword;
        }
        CSStatement.prototype.setVSXMLState = function (newVSState) {
            this.vsState = newVSState;
        };
        CSStatement.prototype.makeClone = function () {
            var st = new CSStatement(this.getMode(), this.parent ? this.parent.clone() : null, this.level, this.plevel, this.razorMode, this.expression, this.firstToken, this.firstTokenWasKeyword);
            if (this.vsState !== null) {
                st.setVSXMLState(this.vsState.clone());
            }
            return st;
        };
        CSStatement.prototype.equals = function (other) {
            return _super.prototype.equals.call(this, other) &&
                (other instanceof CSStatement) &&
                ((this.vsState === null && other.vsState === null) ||
                    (this.vsState !== null && this.vsState.equals(other.vsState)));
        };
        CSStatement.prototype.stateTokenize = function (stream) {
            if (abstractMode_1.isDigit(stream.peek(), 10)) {
                this.firstToken = false;
                return { nextState: new CSNumber(this.getMode(), this, stream.next()) };
            }
            var token = stream.nextToken();
            var acceptNestedModes = !this.firstTokenWasKeyword;
            var nextStateAtEnd = (this.level <= 0 && this.plevel <= 0 && stream.eos() ? this.parent : undefined);
            if (stream.eos()) {
                this.firstTokenWasKeyword = false; // Set this for the state starting on the next line.
            }
            if (isKeyword(token)) {
                if (this.level <= 0) {
                    this.expression = false;
                }
                if (this.firstToken) {
                    this.firstTokenWasKeyword = true;
                }
                return { type: 'keyword.cs' };
            }
            this.firstToken = false;
            if (this.razorMode && token === '<' && acceptNestedModes) {
                if (!stream.eos() && /[_:!\/\w]/.test(stream.peek())) {
                    return { nextState: new CSSimpleHTML(this.getMode(), this, htmlMode.States.Content) };
                }
            }
            // exit expressions on anything that doesn't look like part of the same expression
            if (this.razorMode && this.expression && this.level <= 0 && this.plevel <= 0 && !stream.eos()) {
                if (!/^(\.|\[|\(|\{\w+)$/.test(stream.peekToken())) {
                    nextStateAtEnd = this.parent;
                }
            }
            if (token === '/') {
                if (!stream.eos() && !stream.peekWhitespace()) {
                    switch (stream.peekToken()) {
                        case '/':
                            stream.nextToken();
                            if (!stream.eos() && stream.peekToken() === '/') {
                                stream.nextToken();
                                if (stream.eos()) {
                                    return {
                                        type: 'comment.vs'
                                    };
                                }
                                if (stream.peekToken() !== '/') {
                                    return {
                                        type: 'comment.vs',
                                        nextState: new VSXML.VSXMLEmbeddedState(this.getMode(), this.vsState, this)
                                    };
                                }
                            }
                            stream.advanceToEOS();
                            return { type: 'comment.cs' };
                        case '*':
                            stream.nextToken();
                            return { nextState: new CSComment(this.getMode(), this, '/') };
                    }
                }
                return { type: 'punctuation.cs', nextState: nextStateAtEnd };
            }
            if (token === '@') {
                if (!stream.eos()) {
                    switch (stream.peekToken()) {
                        case '"':
                            stream.nextToken();
                            return { nextState: new CSVerbatimString(this.getMode(), this) };
                        case '*':
                            stream.nextToken();
                            return { nextState: new CSComment(this.getMode(), this, '@') };
                    }
                }
            }
            if (/@?\w+/.test(token)) {
                return { type: 'ident.cs', nextState: nextStateAtEnd };
            }
            if (token === '"' || token === '\'') {
                return { nextState: new CSString(this.getMode(), this, token) };
            }
            if (brackets.stringIsBracket(token)) {
                var tr = {
                    type: brackets.tokenTypeFromString(token),
                    nextState: nextStateAtEnd
                };
                if (this.razorMode) {
                    if (token === '{') {
                        this.expression = false; // whenever we enter a block, we exit expression mode
                        this.level++;
                        if (this.level === 1) {
                            tr.type = razorTokenTypes.EMBED_CS;
                            tr.nextState = undefined;
                        }
                    }
                    if (token === '}') {
                        this.level--;
                        if (this.level <= 0) {
                            tr.type = razorTokenTypes.EMBED_CS;
                            tr.nextState = this.parent;
                        }
                    }
                    if (this.expression) {
                        if (token === '(') {
                            this.plevel++;
                            if (this.plevel === 1) {
                                tr.type = razorTokenTypes.EMBED_CS;
                                tr.nextState = undefined;
                            }
                        }
                        if (token === ')') {
                            this.plevel--;
                            if (this.expression && this.plevel <= 0) {
                                tr.type = razorTokenTypes.EMBED_CS;
                                tr.nextState = this.parent;
                            }
                        }
                        if (token === '[') {
                            this.plevel++;
                            tr.nextState = undefined;
                        }
                        if (token === ']') {
                            this.plevel--;
                        }
                    }
                }
                return tr;
            }
            if (ispunctuation(token)) {
                return { type: 'punctuation.cs', nextState: nextStateAtEnd };
            }
            if (this.razorMode && this.expression && this.plevel <= 0) {
                return { type: '', nextState: this.parent };
            }
            return { type: '', nextState: nextStateAtEnd };
        };
        return CSStatement;
    }(CSState));
    exports.CSStatement = CSStatement;
    // this state always returns to parent state if it leaves a html tag
    var CSSimpleHTML = (function (_super) {
        __extends(CSSimpleHTML, _super);
        function CSSimpleHTML(mode, parent, state) {
            _super.call(this, mode, 'number', parent);
            this.state = state;
        }
        CSSimpleHTML.prototype.makeClone = function () {
            return new CSSimpleHTML(this.getMode(), this.parent ? this.parent.clone() : null, this.state);
        };
        CSSimpleHTML.prototype.nextName = function (stream) {
            return stream.advanceIfRegExp(/^[_:\w][_:\w-.\d]*/);
        };
        CSSimpleHTML.prototype.nextAttrValue = function (stream) {
            return stream.advanceIfRegExp(/^('|').*?\1/);
        };
        CSSimpleHTML.prototype.tokenize = function (stream) {
            switch (this.state) {
                case htmlMode.States.WithinComment:
                    if (stream.advanceUntil('-->', false).length > 0) {
                        return { type: htmlTokenTypes.COMMENT };
                    }
                    if (stream.advanceIfString('-->').length > 0) {
                        this.state = htmlMode.States.Content;
                        return { type: htmlTokenTypes.DELIM_COMMENT, nextState: this.parent };
                    }
                    break;
                case htmlMode.States.WithinDoctype:
                    if (stream.advanceUntil('>', false).length > 0) {
                        return { type: htmlTokenTypes.DOCTYPE };
                    }
                    if (stream.advanceIfString('>').length > 0) {
                        this.state = htmlMode.States.Content;
                        return { type: htmlTokenTypes.DELIM_DOCTYPE, nextState: this.parent };
                    }
                    break;
                case htmlMode.States.Content:
                    if (stream.advanceIfString('!--').length > 0) {
                        this.state = htmlMode.States.WithinComment;
                        return { type: htmlTokenTypes.DELIM_COMMENT };
                    }
                    if (stream.advanceIfRegExp(/!DOCTYPE/i).length > 0) {
                        this.state = htmlMode.States.WithinDoctype;
                        return { type: htmlTokenTypes.DELIM_DOCTYPE };
                    }
                    if (stream.advanceIfString('/').length > 0) {
                        this.state = htmlMode.States.OpeningEndTag;
                        return { type: htmlTokenTypes.DELIM_END };
                    }
                    this.state = htmlMode.States.OpeningStartTag;
                    return { type: htmlTokenTypes.DELIM_START };
                case htmlMode.States.OpeningEndTag: {
                    var tagName = this.nextName(stream);
                    if (tagName.length > 0) {
                        return {
                            type: htmlTokenTypes.getTag(tagName)
                        };
                    }
                    if (stream.advanceIfString('>').length > 0) {
                        this.state = htmlMode.States.Content;
                        return { type: htmlTokenTypes.DELIM_END, nextState: this.parent };
                    }
                    stream.advanceUntil('>', false);
                    return { type: '' };
                }
                case htmlMode.States.OpeningStartTag: {
                    var tagName = this.nextName(stream);
                    if (tagName.length > 0) {
                        this.state = htmlMode.States.WithinTag;
                        return {
                            type: htmlTokenTypes.getTag(tagName)
                        };
                    }
                    break;
                }
                case htmlMode.States.WithinTag:
                    if (stream.skipWhitespace().length > 0) {
                        return { type: '' };
                    }
                    var name = this.nextName(stream);
                    if (name.length > 0) {
                        this.state = htmlMode.States.AttributeName;
                        return { type: htmlTokenTypes.ATTRIB_NAME };
                    }
                    if (stream.advanceIfRegExp(/^\/?>/).length > 0) {
                        this.state = htmlMode.States.Content;
                        return { type: htmlTokenTypes.DELIM_START, nextState: this.parent };
                    }
                    stream.next();
                    return { type: '' };
                case htmlMode.States.AttributeName:
                    if (stream.skipWhitespace().length > 0 || stream.eos()) {
                        return { type: '' };
                    }
                    if (stream.peek() === '=') {
                        stream.next();
                        this.state = htmlMode.States.AttributeValue;
                        return { type: '' };
                    }
                    this.state = htmlMode.States.WithinTag;
                    return this.tokenize(stream); // no advance yet - jump to WithinTag
                case htmlMode.States.AttributeValue:
                    if (stream.skipWhitespace().length > 0 || stream.eos()) {
                        return { type: '' };
                    }
                    var value = this.nextAttrValue(stream);
                    if (value.length > 0) {
                        this.state = htmlMode.States.WithinTag;
                        return { type: htmlTokenTypes.ATTRIB_VALUE };
                    }
                    this.state = htmlMode.States.WithinTag;
                    return this.tokenize(stream); // no advance yet - jump to WithinTag
            }
            stream.next();
            this.state = htmlMode.States.Content;
            return { type: '', nextState: this.parent };
        };
        return CSSimpleHTML;
    }(CSState));
});






var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __param = (this && this.__param) || function (paramIndex, decorator) {
    return function (target, key) { decorator(target, key, paramIndex); }
};
define(__m[11], __M([1,0,12,9,8,6,2,13,14,15,16,17]), function (require, exports, modes, htmlMode, csharpTokenization, abstractMode_1, razorTokenTypes, instantiation_1, modeService_1, languageConfigurationRegistry_1, async_1, compatWorkerService_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    // for a brief description of the razor syntax see http://www.mikesdotnetting.com/Article/153/Inline-Razor-Syntax-Overview
    var RAZORState = (function (_super) {
        __extends(RAZORState, _super);
        function RAZORState(mode, kind, lastTagName, lastAttributeName, embeddedContentType, attributeValueQuote, attributeValue) {
            _super.call(this, mode, kind, lastTagName, lastAttributeName, embeddedContentType, attributeValueQuote, attributeValue);
        }
        RAZORState.prototype.makeClone = function () {
            return new RAZORState(this.getMode(), this.kind, this.lastTagName, this.lastAttributeName, this.embeddedContentType, this.attributeValueQuote, this.attributeValue);
        };
        RAZORState.prototype.equals = function (other) {
            if (other instanceof RAZORState) {
                return (_super.prototype.equals.call(this, other));
            }
            return false;
        };
        RAZORState.prototype.tokenize = function (stream) {
            if (!stream.eos() && stream.peek() === '@') {
                stream.next();
                if (!stream.eos() && stream.peek() === '*') {
                    return { nextState: new csharpTokenization.CSComment(this.getMode(), this, '@') };
                }
                if (stream.eos() || stream.peek() !== '@') {
                    return { type: razorTokenTypes.EMBED_CS, nextState: new csharpTokenization.CSStatement(this.getMode(), this, 0, 0, true, true, true, false) };
                }
            }
            return _super.prototype.tokenize.call(this, stream);
        };
        return RAZORState;
    }(htmlMode.State));
    var RAZORMode = (function (_super) {
        __extends(RAZORMode, _super);
        function RAZORMode(descriptor, instantiationService, modeService, compatWorkerService) {
            _super.call(this, descriptor, instantiationService, modeService, compatWorkerService);
        }
        RAZORMode.prototype._registerSupports = function () {
            var _this = this;
            modes.SuggestRegistry.register(this.getId(), {
                triggerCharacters: ['.', ':', '<', '"', '=', '/'],
                shouldAutotriggerSuggest: true,
                provideCompletionItems: function (model, position, token) {
                    return async_1.wireCancellationToken(token, _this._provideCompletionItems(model.uri, position));
                }
            }, true);
            modes.DocumentHighlightProviderRegistry.register(this.getId(), {
                provideDocumentHighlights: function (model, position, token) {
                    return async_1.wireCancellationToken(token, _this._provideDocumentHighlights(model.uri, position));
                }
            }, true);
            modes.LinkProviderRegistry.register(this.getId(), {
                provideLinks: function (model, token) {
                    return async_1.wireCancellationToken(token, _this._provideLinks(model.uri));
                }
            }, true);
            languageConfigurationRegistry_1.LanguageConfigurationRegistry.register(this.getId(), RAZORMode.LANG_CONFIG);
        };
        RAZORMode.prototype._createModeWorkerManager = function (descriptor, instantiationService) {
            return new abstractMode_1.ModeWorkerManager(descriptor, 'vs/languages/razor/common/razorWorker', 'RAZORWorker', 'vs/languages/html/common/htmlWorker', instantiationService);
        };
        RAZORMode.prototype.getInitialState = function () {
            return new RAZORState(this, htmlMode.States.Content, '', '', '', '', '');
        };
        RAZORMode.prototype.getLeavingNestedModeData = function (line, state) {
            var leavingNestedModeData = _super.prototype.getLeavingNestedModeData.call(this, line, state);
            if (leavingNestedModeData) {
                leavingNestedModeData.stateAfterNestedMode = new RAZORState(this, htmlMode.States.Content, '', '', '', '', '');
            }
            return leavingNestedModeData;
        };
        RAZORMode.LANG_CONFIG = {
            wordPattern: abstractMode_1.createWordRegExp('#?%'),
            comments: {
                blockComment: ['<!--', '-->']
            },
            brackets: [
                ['<!--', '-->'],
                ['{', '}'],
                ['(', ')']
            ],
            __electricCharacterSupport: {
                embeddedElectricCharacters: ['*', '}', ']', ')']
            },
            autoClosingPairs: [
                { open: '{', close: '}' },
                { open: '[', close: ']' },
                { open: '(', close: ')' },
                { open: '"', close: '"' },
                { open: '\'', close: '\'' }
            ],
            surroundingPairs: [
                { open: '"', close: '"' },
                { open: '\'', close: '\'' }
            ],
            onEnterRules: [
                {
                    beforeText: new RegExp("<(?!(?:" + htmlMode.EMPTY_ELEMENTS.join('|') + "))(\\w[\\w\\d]*)([^/>]*(?!/)>)[^<]*$", 'i'),
                    afterText: /^<\/(\w[\w\d]*)\s*>$/i,
                    action: { indentAction: modes.IndentAction.IndentOutdent }
                },
                {
                    beforeText: new RegExp("<(?!(?:" + htmlMode.EMPTY_ELEMENTS.join('|') + "))(\\w[\\w\\d]*)([^/>]*(?!/)>)[^<]*$", 'i'),
                    action: { indentAction: modes.IndentAction.Indent }
                }
            ],
        };
        RAZORMode = __decorate([
            __param(1, instantiation_1.IInstantiationService),
            __param(2, modeService_1.IModeService),
            __param(3, compatWorkerService_1.ICompatWorkerService)
        ], RAZORMode);
        return RAZORMode;
    }(htmlMode.HTMLMode));
    exports.RAZORMode = RAZORMode;
});

}).call(this);
//# sourceMappingURL=razor.js.map
