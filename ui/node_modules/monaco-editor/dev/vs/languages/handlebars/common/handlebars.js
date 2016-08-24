/*!-----------------------------------------------------------
 * Copyright (c) Microsoft Corporation. All rights reserved.
 * Version: 0.5.3(793ede49d53dba79d39e52205f16321278f5183c)
 * Released under the MIT license
 * https://github.com/Microsoft/vscode/blob/master/LICENSE.txt
 *-----------------------------------------------------------*/

(function() {
var __m = ["vs/languages/handlebars/common/handlebarsTokenTypes","exports","require","vs/languages/handlebars/common/handlebars","vs/editor/common/modes","vs/languages/html/common/html","vs/editor/common/services/compatWorkerService","vs/editor/common/services/modeService","vs/editor/common/modes/languageConfigurationRegistry","vs/editor/common/modes/abstractMode","vs/base/common/async","vs/platform/instantiation/common/instantiation"];
var __M = function(deps) {
  var result = [];
  for (var i = 0, len = deps.length; i < len; i++) {
    result[i] = __m[deps[i]];
  }
  return result;
};
define(__m[0], __M([2,1]), function (require, exports) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    exports.EMBED = 'punctuation.expression.unescaped.handlebars';
    exports.EMBED_UNESCAPED = 'punctuation.expression.handlebars';
    exports.KEYWORD = 'keyword.helper.handlebars';
    exports.VARIABLE = 'variable.parameter.handlebars';
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
define(__m[3], __M([2,1,4,5,0,11,7,8,9,10,6]), function (require, exports, modes, htmlMode, handlebarsTokenTypes, instantiation_1, modeService_1, languageConfigurationRegistry_1, abstractMode_1, async_1, compatWorkerService_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    (function (States) {
        States[States["HTML"] = 0] = "HTML";
        States[States["Expression"] = 1] = "Expression";
        States[States["UnescapedExpression"] = 2] = "UnescapedExpression";
    })(exports.States || (exports.States = {}));
    var States = exports.States;
    var HandlebarsState = (function (_super) {
        __extends(HandlebarsState, _super);
        function HandlebarsState(mode, kind, handlebarsKind, lastTagName, lastAttributeName, embeddedContentType, attributeValueQuote, attributeValue) {
            _super.call(this, mode, kind, lastTagName, lastAttributeName, embeddedContentType, attributeValueQuote, attributeValue);
            this.kind = kind;
            this.handlebarsKind = handlebarsKind;
            this.lastTagName = lastTagName;
            this.lastAttributeName = lastAttributeName;
            this.embeddedContentType = embeddedContentType;
            this.attributeValueQuote = attributeValueQuote;
            this.attributeValue = attributeValue;
        }
        HandlebarsState.prototype.makeClone = function () {
            return new HandlebarsState(this.getMode(), this.kind, this.handlebarsKind, this.lastTagName, this.lastAttributeName, this.embeddedContentType, this.attributeValueQuote, this.attributeValue);
        };
        HandlebarsState.prototype.equals = function (other) {
            if (other instanceof HandlebarsState) {
                return (_super.prototype.equals.call(this, other));
            }
            return false;
        };
        HandlebarsState.prototype.tokenize = function (stream) {
            switch (this.handlebarsKind) {
                case States.HTML:
                    if (stream.advanceIfString('{{{').length > 0) {
                        this.handlebarsKind = States.UnescapedExpression;
                        return { type: handlebarsTokenTypes.EMBED_UNESCAPED };
                    }
                    else if (stream.advanceIfString('{{').length > 0) {
                        this.handlebarsKind = States.Expression;
                        return { type: handlebarsTokenTypes.EMBED };
                    }
                    break;
                case States.Expression:
                case States.UnescapedExpression:
                    if (this.handlebarsKind === States.Expression && stream.advanceIfString('}}').length > 0) {
                        this.handlebarsKind = States.HTML;
                        return { type: handlebarsTokenTypes.EMBED };
                    }
                    else if (this.handlebarsKind === States.UnescapedExpression && stream.advanceIfString('}}}').length > 0) {
                        this.handlebarsKind = States.HTML;
                        return { type: handlebarsTokenTypes.EMBED_UNESCAPED };
                    }
                    else if (stream.skipWhitespace().length > 0) {
                        return { type: '' };
                    }
                    if (stream.peek() === '#') {
                        stream.advanceWhile(/^[^\s}]/);
                        return { type: handlebarsTokenTypes.KEYWORD };
                    }
                    if (stream.peek() === '/') {
                        stream.advanceWhile(/^[^\s}]/);
                        return { type: handlebarsTokenTypes.KEYWORD };
                    }
                    if (stream.advanceIfString('else')) {
                        var next = stream.peek();
                        if (next === ' ' || next === '\t' || next === '}') {
                            return { type: handlebarsTokenTypes.KEYWORD };
                        }
                        else {
                            stream.goBack(4);
                        }
                    }
                    if (stream.advanceWhile(/^[^\s}]/).length > 0) {
                        return { type: handlebarsTokenTypes.VARIABLE };
                    }
                    break;
            }
            return _super.prototype.tokenize.call(this, stream);
        };
        return HandlebarsState;
    }(htmlMode.State));
    exports.HandlebarsState = HandlebarsState;
    var HandlebarsMode = (function (_super) {
        __extends(HandlebarsMode, _super);
        function HandlebarsMode(descriptor, instantiationService, modeService, compatWorkerService) {
            _super.call(this, descriptor, instantiationService, modeService, compatWorkerService);
        }
        HandlebarsMode.prototype._registerSupports = function () {
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
            languageConfigurationRegistry_1.LanguageConfigurationRegistry.register(this.getId(), HandlebarsMode.LANG_CONFIG);
        };
        HandlebarsMode.prototype.getInitialState = function () {
            return new HandlebarsState(this, htmlMode.States.Content, States.HTML, '', '', '', '', '');
        };
        HandlebarsMode.prototype.getLeavingNestedModeData = function (line, state) {
            var leavingNestedModeData = _super.prototype.getLeavingNestedModeData.call(this, line, state);
            if (leavingNestedModeData) {
                leavingNestedModeData.stateAfterNestedMode = new HandlebarsState(this, htmlMode.States.Content, States.HTML, '', '', '', '', '');
            }
            return leavingNestedModeData;
        };
        HandlebarsMode.LANG_CONFIG = {
            wordPattern: abstractMode_1.createWordRegExp('#-?%'),
            comments: {
                blockComment: ['<!--', '-->']
            },
            brackets: [
                ['<!--', '-->'],
                ['{{', '}}']
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
                { open: '<', close: '>' },
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
        HandlebarsMode = __decorate([
            __param(1, instantiation_1.IInstantiationService),
            __param(2, modeService_1.IModeService),
            __param(3, compatWorkerService_1.ICompatWorkerService)
        ], HandlebarsMode);
        return HandlebarsMode;
    }(htmlMode.HTMLMode));
    exports.HandlebarsMode = HandlebarsMode;
});

}).call(this);
//# sourceMappingURL=handlebars.js.map
