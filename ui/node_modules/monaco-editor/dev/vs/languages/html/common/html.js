/*!-----------------------------------------------------------
 * Copyright (c) Microsoft Corporation. All rights reserved.
 * Version: 0.5.3(793ede49d53dba79d39e52205f16321278f5183c)
 * Released under the MIT license
 * https://github.com/Microsoft/vscode/blob/master/LICENSE.txt
 *-----------------------------------------------------------*/

(function() {
var __m = ["exports","require","vs/languages/html/common/htmlEmptyTagsShared","vs/languages/html/common/htmlTokenTypes","vs/editor/common/services/compatWorkerService","vs/base/common/strings","vs/languages/html/common/html","vs/editor/common/modes","vs/base/common/arrays","vs/editor/common/modes/abstractState","vs/editor/common/services/modeService","vs/platform/instantiation/common/instantiation","vs/editor/common/modes/languageConfigurationRegistry","vs/editor/common/modes/supports/tokenizationSupport","vs/base/common/async","vs/editor/common/modes/abstractMode"];
var __M = function(deps) {
  var result = [];
  for (var i = 0, len = deps.length; i < len; i++) {
    result[i] = __m[deps[i]];
  }
  return result;
};
/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Microsoft Corporation. All rights reserved.
 *  Licensed under the MIT License. See License.txt in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
define(__m[2], __M([1,0,8]), function (require, exports, arrays) {
    "use strict";
    exports.EMPTY_ELEMENTS = ['area', 'base', 'br', 'col', 'embed', 'hr', 'img', 'input', 'keygen', 'link', 'menuitem', 'meta', 'param', 'source', 'track', 'wbr'];
    function isEmptyElement(e) {
        return arrays.binarySearch(exports.EMPTY_ELEMENTS, e, function (s1, s2) { return s1.localeCompare(s2); }) >= 0;
    }
    exports.isEmptyElement = isEmptyElement;
});

define(__m[3], __M([1,0,5]), function (require, exports, strings) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    exports.DELIM_END = 'punctuation.definition.meta.tag.end.html';
    exports.DELIM_START = 'punctuation.definition.meta.tag.begin.html';
    exports.DELIM_ASSIGN = 'meta.tag.assign.html';
    exports.ATTRIB_NAME = 'entity.other.attribute-name.html';
    exports.ATTRIB_VALUE = 'string.html';
    exports.COMMENT = 'comment.html.content';
    exports.DELIM_COMMENT = 'comment.html';
    exports.DOCTYPE = 'entity.other.attribute-name.html';
    exports.DELIM_DOCTYPE = 'entity.name.tag.html';
    var TAG_PREFIX = 'entity.name.tag.tag-';
    function isTag(tokenType) {
        return strings.startsWith(tokenType, TAG_PREFIX);
    }
    exports.isTag = isTag;
    function getTag(name) {
        return TAG_PREFIX + name;
    }
    exports.getTag = getTag;
});

var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __param = (this && this.__param) || function (paramIndex, decorator) {
    return function (target, key) { decorator(target, key, paramIndex); }
};
define(__m[6], __M([1,0,7,15,9,10,11,3,2,12,13,14,4]), function (require, exports, modes, abstractMode_1, abstractState_1, modeService_1, instantiation_1, htmlTokenTypes, htmlEmptyTagsShared_1, languageConfigurationRegistry_1, tokenizationSupport_1, async_1, compatWorkerService_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    exports.htmlTokenTypes = htmlTokenTypes;
    exports.EMPTY_ELEMENTS = htmlEmptyTagsShared_1.EMPTY_ELEMENTS;
     // export to be used by Razor. We are the main module, so Razor should get it from us.
     // export to be used by Razor. We are the main module, so Razor should get it from us.
    (function (States) {
        States[States["Content"] = 0] = "Content";
        States[States["OpeningStartTag"] = 1] = "OpeningStartTag";
        States[States["OpeningEndTag"] = 2] = "OpeningEndTag";
        States[States["WithinDoctype"] = 3] = "WithinDoctype";
        States[States["WithinTag"] = 4] = "WithinTag";
        States[States["WithinComment"] = 5] = "WithinComment";
        States[States["WithinEmbeddedContent"] = 6] = "WithinEmbeddedContent";
        States[States["AttributeName"] = 7] = "AttributeName";
        States[States["AttributeValue"] = 8] = "AttributeValue";
    })(exports.States || (exports.States = {}));
    var States = exports.States;
    // list of elements that embed other content
    var tagsEmbeddingContent = ['script', 'style'];
    var State = (function (_super) {
        __extends(State, _super);
        function State(mode, kind, lastTagName, lastAttributeName, embeddedContentType, attributeValueQuote, attributeValue) {
            _super.call(this, mode);
            this.kind = kind;
            this.lastTagName = lastTagName;
            this.lastAttributeName = lastAttributeName;
            this.embeddedContentType = embeddedContentType;
            this.attributeValueQuote = attributeValueQuote;
            this.attributeValue = attributeValue;
        }
        State.escapeTagName = function (s) {
            return htmlTokenTypes.getTag(s.replace(/[:_.]/g, '-'));
        };
        State.prototype.makeClone = function () {
            return new State(this.getMode(), this.kind, this.lastTagName, this.lastAttributeName, this.embeddedContentType, this.attributeValueQuote, this.attributeValue);
        };
        State.prototype.equals = function (other) {
            if (other instanceof State) {
                return (_super.prototype.equals.call(this, other) &&
                    this.kind === other.kind &&
                    this.lastTagName === other.lastTagName &&
                    this.lastAttributeName === other.lastAttributeName &&
                    this.embeddedContentType === other.embeddedContentType &&
                    this.attributeValueQuote === other.attributeValueQuote &&
                    this.attributeValue === other.attributeValue);
            }
            return false;
        };
        State.prototype.nextElementName = function (stream) {
            return stream.advanceIfRegExp(/^[_:\w][_:\w-.\d]*/).toLowerCase();
        };
        State.prototype.nextAttributeName = function (stream) {
            return stream.advanceIfRegExp(/^[^\s"'>/=\x00-\x0F\x7F\x80-\x9F]*/).toLowerCase();
        };
        State.prototype.tokenize = function (stream) {
            switch (this.kind) {
                case States.WithinComment:
                    if (stream.advanceUntilString2('-->', false)) {
                        return { type: htmlTokenTypes.COMMENT };
                    }
                    else if (stream.advanceIfString2('-->')) {
                        this.kind = States.Content;
                        return { type: htmlTokenTypes.DELIM_COMMENT, dontMergeWithPrev: true };
                    }
                    break;
                case States.WithinDoctype:
                    if (stream.advanceUntilString2('>', false)) {
                        return { type: htmlTokenTypes.DOCTYPE };
                    }
                    else if (stream.advanceIfString2('>')) {
                        this.kind = States.Content;
                        return { type: htmlTokenTypes.DELIM_DOCTYPE, dontMergeWithPrev: true };
                    }
                    break;
                case States.Content:
                    if (stream.advanceIfCharCode2('<'.charCodeAt(0))) {
                        if (!stream.eos() && stream.peek() === '!') {
                            if (stream.advanceIfString2('!--')) {
                                this.kind = States.WithinComment;
                                return { type: htmlTokenTypes.DELIM_COMMENT, dontMergeWithPrev: true };
                            }
                            if (stream.advanceIfStringCaseInsensitive2('!DOCTYPE')) {
                                this.kind = States.WithinDoctype;
                                return { type: htmlTokenTypes.DELIM_DOCTYPE, dontMergeWithPrev: true };
                            }
                        }
                        if (stream.advanceIfCharCode2('/'.charCodeAt(0))) {
                            this.kind = States.OpeningEndTag;
                            return { type: htmlTokenTypes.DELIM_END, dontMergeWithPrev: true };
                        }
                        this.kind = States.OpeningStartTag;
                        return { type: htmlTokenTypes.DELIM_START, dontMergeWithPrev: true };
                    }
                    break;
                case States.OpeningEndTag:
                    var tagName = this.nextElementName(stream);
                    if (tagName.length > 0) {
                        return {
                            type: State.escapeTagName(tagName),
                        };
                    }
                    else if (stream.advanceIfString2('>')) {
                        this.kind = States.Content;
                        return { type: htmlTokenTypes.DELIM_END, dontMergeWithPrev: true };
                    }
                    else {
                        stream.advanceUntilString2('>', false);
                        return { type: '' };
                    }
                case States.OpeningStartTag:
                    this.lastTagName = this.nextElementName(stream);
                    if (this.lastTagName.length > 0) {
                        this.lastAttributeName = null;
                        if ('script' === this.lastTagName || 'style' === this.lastTagName) {
                            this.lastAttributeName = null;
                            this.embeddedContentType = null;
                        }
                        this.kind = States.WithinTag;
                        return {
                            type: State.escapeTagName(this.lastTagName),
                        };
                    }
                    break;
                case States.WithinTag:
                    if (stream.skipWhitespace2() || stream.eos()) {
                        this.lastAttributeName = ''; // remember that we have seen a whitespace
                        return { type: '' };
                    }
                    else {
                        if (this.lastAttributeName === '') {
                            var name = this.nextAttributeName(stream);
                            if (name.length > 0) {
                                this.lastAttributeName = name;
                                this.kind = States.AttributeName;
                                return { type: htmlTokenTypes.ATTRIB_NAME };
                            }
                        }
                        if (stream.advanceIfString2('/>')) {
                            this.kind = States.Content;
                            return { type: htmlTokenTypes.DELIM_START, dontMergeWithPrev: true };
                        }
                        if (stream.advanceIfCharCode2('>'.charCodeAt(0))) {
                            if (tagsEmbeddingContent.indexOf(this.lastTagName) !== -1) {
                                this.kind = States.WithinEmbeddedContent;
                                return { type: htmlTokenTypes.DELIM_START, dontMergeWithPrev: true };
                            }
                            else {
                                this.kind = States.Content;
                                return { type: htmlTokenTypes.DELIM_START, dontMergeWithPrev: true };
                            }
                        }
                        else {
                            stream.next2();
                            return { type: '' };
                        }
                    }
                case States.AttributeName:
                    if (stream.skipWhitespace2() || stream.eos()) {
                        return { type: '' };
                    }
                    if (stream.advanceIfCharCode2('='.charCodeAt(0))) {
                        this.kind = States.AttributeValue;
                        return { type: htmlTokenTypes.DELIM_ASSIGN };
                    }
                    else {
                        this.kind = States.WithinTag;
                        this.lastAttributeName = '';
                        return this.tokenize(stream); // no advance yet - jump to WithinTag
                    }
                case States.AttributeValue:
                    if (stream.eos()) {
                        return { type: '' };
                    }
                    if (stream.skipWhitespace2()) {
                        if (this.attributeValueQuote === '"' || this.attributeValueQuote === '\'') {
                            // We are inside the quotes of an attribute value
                            return { type: htmlTokenTypes.ATTRIB_VALUE };
                        }
                        return { type: '' };
                    }
                    // We are in a attribute value
                    if (this.attributeValueQuote === '"' || this.attributeValueQuote === '\'') {
                        if (this.attributeValue === this.attributeValueQuote && ('script' === this.lastTagName || 'style' === this.lastTagName) && 'type' === this.lastAttributeName) {
                            this.attributeValue = stream.advanceUntilString(this.attributeValueQuote, true);
                            if (this.attributeValue.length > 0) {
                                this.embeddedContentType = this.unquote(this.attributeValue);
                                this.kind = States.WithinTag;
                                this.attributeValue = '';
                                this.attributeValueQuote = '';
                                return { type: htmlTokenTypes.ATTRIB_VALUE };
                            }
                        }
                        else {
                            if (stream.advanceIfCharCode2(this.attributeValueQuote.charCodeAt(0))) {
                                this.kind = States.WithinTag;
                                this.attributeValue = '';
                                this.attributeValueQuote = '';
                                this.lastAttributeName = null;
                            }
                            else {
                                var part = stream.next();
                                this.attributeValue += part;
                            }
                            return { type: htmlTokenTypes.ATTRIB_VALUE };
                        }
                    }
                    else {
                        var attributeValue = stream.advanceIfRegExp(/^[^\s"'`=<>]+/);
                        if (attributeValue.length > 0) {
                            this.kind = States.WithinTag;
                            this.lastAttributeName = null;
                            return { type: htmlTokenTypes.ATTRIB_VALUE };
                        }
                        var ch = stream.peek();
                        if (ch === '\'' || ch === '"') {
                            this.attributeValueQuote = ch;
                            this.attributeValue = ch;
                            stream.next2();
                            return { type: htmlTokenTypes.ATTRIB_VALUE };
                        }
                        else {
                            this.kind = States.WithinTag;
                            this.lastAttributeName = null;
                            return this.tokenize(stream); // no advance yet - jump to WithinTag
                        }
                    }
            }
            stream.next2();
            this.kind = States.Content;
            return { type: '' };
        };
        State.prototype.unquote = function (value) {
            var start = 0;
            var end = value.length;
            if ('"' === value[0]) {
                start++;
            }
            if ('"' === value[end - 1]) {
                end--;
            }
            return value.substring(start, end);
        };
        return State;
    }(abstractState_1.AbstractState));
    exports.State = State;
    var HTMLMode = (function (_super) {
        __extends(HTMLMode, _super);
        function HTMLMode(descriptor, instantiationService, modeService, compatWorkerService) {
            _super.call(this, descriptor.id, compatWorkerService);
            this._modeWorkerManager = this._createModeWorkerManager(descriptor, instantiationService);
            this.modeService = modeService;
            this.tokenizationSupport = new tokenizationSupport_1.TokenizationSupport(this, this, true);
            this.configSupport = this;
            this._registerSupports();
        }
        HTMLMode.prototype._registerSupports = function () {
            var _this = this;
            if (this.getId() !== 'html') {
                throw new Error('This method must be overwritten!');
            }
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
            modes.DocumentRangeFormattingEditProviderRegistry.register(this.getId(), {
                provideDocumentRangeFormattingEdits: function (model, range, options, token) {
                    return async_1.wireCancellationToken(token, _this._provideDocumentRangeFormattingEdits(model.uri, range, options));
                }
            }, true);
            modes.LinkProviderRegistry.register(this.getId(), {
                provideLinks: function (model, token) {
                    return async_1.wireCancellationToken(token, _this._provideLinks(model.uri));
                }
            }, true);
            languageConfigurationRegistry_1.LanguageConfigurationRegistry.register(this.getId(), HTMLMode.LANG_CONFIG);
        };
        HTMLMode.prototype._createModeWorkerManager = function (descriptor, instantiationService) {
            return new abstractMode_1.ModeWorkerManager(descriptor, 'vs/languages/html/common/htmlWorker', 'HTMLWorker', null, instantiationService);
        };
        HTMLMode.prototype._worker = function (runner) {
            return this._modeWorkerManager.worker(runner);
        };
        // TokenizationSupport
        HTMLMode.prototype.getInitialState = function () {
            return new State(this, States.Content, '', '', '', '', '');
        };
        HTMLMode.prototype.enterNestedMode = function (state) {
            return state instanceof State && state.kind === States.WithinEmbeddedContent;
        };
        HTMLMode.prototype.getNestedMode = function (state) {
            var result = null;
            var htmlState = state;
            var missingModePromise = null;
            if (htmlState.embeddedContentType !== null) {
                if (this.modeService.isRegisteredMode(htmlState.embeddedContentType)) {
                    result = this.modeService.getMode(htmlState.embeddedContentType);
                    if (!result) {
                        missingModePromise = this.modeService.getOrCreateMode(htmlState.embeddedContentType);
                    }
                }
            }
            else {
                var mimeType = null;
                if ('script' === htmlState.lastTagName) {
                    mimeType = 'text/javascript';
                }
                else if ('style' === htmlState.lastTagName) {
                    mimeType = 'text/css';
                }
                else {
                    mimeType = 'text/plain';
                }
                result = this.modeService.getMode(mimeType);
            }
            if (result === null) {
                result = this.modeService.getMode('text/plain');
            }
            return {
                mode: result,
                missingModePromise: missingModePromise
            };
        };
        HTMLMode.prototype.getLeavingNestedModeData = function (line, state) {
            var tagName = state.lastTagName;
            var regexp = new RegExp('<\\/' + tagName + '\\s*>', 'i');
            var match = regexp.exec(line);
            if (match !== null) {
                return {
                    nestedModeBuffer: line.substring(0, match.index),
                    bufferAfterNestedMode: line.substring(match.index),
                    stateAfterNestedMode: new State(this, States.Content, '', '', '', '', '')
                };
            }
            return null;
        };
        HTMLMode.prototype.configure = function (options) {
            if (!this.compatWorkerService) {
                return;
            }
            if (this.compatWorkerService.isInMainThread) {
                return this._configureWorker(options);
            }
            else {
                return this._worker(function (w) { return w._doConfigure(options); });
            }
        };
        HTMLMode.prototype._configureWorker = function (options) {
            return this._worker(function (w) { return w._doConfigure(options); });
        };
        HTMLMode.prototype._provideLinks = function (resource) {
            return this._worker(function (w) { return w.provideLinks(resource); });
        };
        HTMLMode.prototype._provideDocumentRangeFormattingEdits = function (resource, range, options) {
            return this._worker(function (w) { return w.provideDocumentRangeFormattingEdits(resource, range, options); });
        };
        HTMLMode.prototype._provideDocumentHighlights = function (resource, position, strict) {
            if (strict === void 0) { strict = false; }
            return this._worker(function (w) { return w.provideDocumentHighlights(resource, position, strict); });
        };
        HTMLMode.prototype._provideCompletionItems = function (resource, position) {
            return this._worker(function (w) { return w.provideCompletionItems(resource, position); });
        };
        HTMLMode.LANG_CONFIG = {
            wordPattern: abstractMode_1.createWordRegExp('#-?%'),
            comments: {
                blockComment: ['<!--', '-->']
            },
            brackets: [
                ['<!--', '-->'],
                ['<', '>'],
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
                    beforeText: new RegExp("<(?!(?:" + htmlEmptyTagsShared_1.EMPTY_ELEMENTS.join('|') + "))([_:\\w][_:\\w-.\\d]*)([^/>]*(?!/)>)[^<]*$", 'i'),
                    afterText: /^<\/([_:\w][_:\w-.\d]*)\s*>$/i,
                    action: { indentAction: modes.IndentAction.IndentOutdent }
                },
                {
                    beforeText: new RegExp("<(?!(?:" + htmlEmptyTagsShared_1.EMPTY_ELEMENTS.join('|') + "))(\\w[\\w\\d]*)([^/>]*(?!/)>)[^<]*$", 'i'),
                    action: { indentAction: modes.IndentAction.Indent }
                }
            ],
        };
        HTMLMode.$_configureWorker = compatWorkerService_1.CompatWorkerAttr(HTMLMode, HTMLMode.prototype._configureWorker);
        HTMLMode.$_provideLinks = compatWorkerService_1.CompatWorkerAttr(HTMLMode, HTMLMode.prototype._provideLinks);
        HTMLMode.$_provideDocumentRangeFormattingEdits = compatWorkerService_1.CompatWorkerAttr(HTMLMode, HTMLMode.prototype._provideDocumentRangeFormattingEdits);
        HTMLMode.$_provideDocumentHighlights = compatWorkerService_1.CompatWorkerAttr(HTMLMode, HTMLMode.prototype._provideDocumentHighlights);
        HTMLMode.$_provideCompletionItems = compatWorkerService_1.CompatWorkerAttr(HTMLMode, HTMLMode.prototype._provideCompletionItems);
        HTMLMode = __decorate([
            __param(1, instantiation_1.IInstantiationService),
            __param(2, modeService_1.IModeService),
            __param(3, compatWorkerService_1.ICompatWorkerService)
        ], HTMLMode);
        return HTMLMode;
    }(abstractMode_1.CompatMode));
    exports.HTMLMode = HTMLMode;
});

}).call(this);
//# sourceMappingURL=html.js.map
