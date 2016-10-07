/*!-----------------------------------------------------------
 * Copyright (c) Microsoft Corporation. All rights reserved.
 * Version: 0.5.3(793ede49d53dba79d39e52205f16321278f5183c)
 * Released under the MIT license
 * https://github.com/Microsoft/vscode/blob/master/LICENSE.txt
 *-----------------------------------------------------------*/

(function() {
var __m = ["exports","require","vs/languages/html/common/htmlScanner","vs/languages/html/common/htmlTokenTypes","vs/nls!vs/languages/html/common/htmlTags","vs/languages/html/common/htmlTags","vs/base/common/strings","vs/base/common/paths","vs/nls","vs/languages/html/common/htmlWorker","vs/nls!vs/languages/html/common/htmlWorker","vs/base/common/winjs.base","vs/languages/lib/common/beautify-html","vs/base/common/network","vs/editor/common/modes","vs/platform/workspace/common/workspace","vs/editor/common/services/resourceService","vs/languages/html/common/htmlEmptyTagsShared","vs/editor/common/modes/supports/suggestSupport","vs/base/common/uri"];
var __M = function(deps) {
  var result = [];
  for (var i = 0, len = deps.length; i < len; i++) {
    result[i] = __m[deps[i]];
  }
  return result;
};
define(__m[2], __M([1,0,3]), function (require, exports, htmlTokenTypes_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    function isDelimiter(tokenType) {
        switch (tokenType) {
            case htmlTokenTypes_1.DELIM_START:
            case htmlTokenTypes_1.DELIM_END:
            case htmlTokenTypes_1.DELIM_ASSIGN:
                return true;
        }
        return false;
    }
    function isInterestingToken(tokenType) {
        switch (tokenType) {
            case htmlTokenTypes_1.DELIM_START:
            case htmlTokenTypes_1.DELIM_END:
            case htmlTokenTypes_1.DELIM_ASSIGN:
            case htmlTokenTypes_1.ATTRIB_NAME:
            case htmlTokenTypes_1.ATTRIB_VALUE:
                return true;
        }
        return htmlTokenTypes_1.isTag(tokenType);
    }
    function getScanner(model, position) {
        var lineOffset = position.column - 1;
        var currentLine = position.lineNumber;
        var tokens = model.getLineTokens(currentLine);
        var lineContent = model.getLineContent(currentLine);
        var tokenIndex = tokens.findIndexOfOffset(lineOffset);
        var tokensOnLine = tokens.getTokenCount();
        var tokenType = tokens.getTokenType(tokenIndex);
        var tokenStart = tokens.getTokenStartIndex(tokenIndex);
        var tokenEnd = tokens.getTokenEndIndex(tokenIndex, lineContent.length);
        if ((tokenType === '' || isDelimiter(tokenType)) && tokenStart === lineOffset) {
            tokenIndex--;
            if (tokenIndex >= 0) {
                // we are at the end of a different token
                tokenType = tokens.getTokenType(tokenIndex);
                tokenStart = tokens.getTokenStartIndex(tokenIndex);
                tokenEnd = tokens.getTokenEndIndex(tokenIndex, lineContent.length);
            }
            else {
                tokenType = '';
                tokenStart = tokenEnd = 0;
            }
        }
        return {
            getTokenType: function () { return tokenType; },
            isAtTokenEnd: function () { return lineOffset === tokenEnd; },
            isAtTokenStart: function () { return lineOffset === tokenStart; },
            getTokenContent: function () { return lineContent.substring(tokenStart, tokenEnd); },
            isOpenBrace: function () { return tokenStart < tokenEnd && lineContent.charAt(tokenStart) === '<'; },
            getTokenPosition: function () { return { lineNumber: currentLine, column: tokenStart + 1 }; },
            getTokenRange: function () { return { startLineNumber: currentLine, startColumn: tokenStart + 1, endLineNumber: currentLine, endColumn: tokenEnd + 1 }; },
            getModel: function () { return model; },
            scanBack: function () {
                if (currentLine <= 0) {
                    return false;
                }
                tokenIndex--;
                do {
                    while (tokenIndex >= 0) {
                        tokenType = tokens.getTokenType(tokenIndex);
                        tokenStart = tokens.getTokenStartIndex(tokenIndex);
                        tokenEnd = tokens.getTokenEndIndex(tokenIndex, lineContent.length);
                        if (isInterestingToken(tokenType)) {
                            return true;
                        }
                        tokenIndex--;
                    }
                    currentLine--;
                    if (currentLine > 0) {
                        tokens = model.getLineTokens(currentLine);
                        lineContent = model.getLineContent(currentLine);
                        tokensOnLine = tokens.getTokenCount();
                        tokenIndex = tokensOnLine - 1;
                    }
                } while (currentLine > 0);
                tokens = null;
                tokenType = lineContent = '';
                tokenStart = tokenEnd = tokensOnLine = 0;
                return false;
            },
            scanForward: function () {
                if (currentLine > model.getLineCount()) {
                    return false;
                }
                tokenIndex++;
                do {
                    while (tokenIndex < tokensOnLine) {
                        tokenType = tokens.getTokenType(tokenIndex);
                        tokenStart = tokens.getTokenStartIndex(tokenIndex);
                        tokenEnd = tokens.getTokenEndIndex(tokenIndex, lineContent.length);
                        if (isInterestingToken(tokenType)) {
                            return true;
                        }
                        tokenIndex++;
                    }
                    currentLine++;
                    tokenIndex = 0;
                    if (currentLine <= model.getLineCount()) {
                        tokens = model.getLineTokens(currentLine);
                        lineContent = model.getLineContent(currentLine);
                        tokensOnLine = tokens.getTokenCount();
                    }
                } while (currentLine <= model.getLineCount());
                tokenType = lineContent = '';
                tokenStart = tokenEnd = tokensOnLine = 0;
                return false;
            }
        };
    }
    exports.getScanner = getScanner;
});

define(__m[4], __M([8,10]), function(nls, data) { return nls.create("vs/languages/html/common/htmlTags", data); });
/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Microsoft Corporation. All rights reserved.
 *  Licensed under the MIT License. See License.txt in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
/*!
BEGIN THIRD PARTY
*/
/*--------------------------------------------------------------------------------------------
 *  This file is based on or incorporates material from the projects listed below (Third Party IP).
 *  The original copyright notice and the license under which Microsoft received such Third Party IP,
 *  are set forth below. Such licenses and notices are provided for informational purposes only.
 *  Microsoft licenses the Third Party IP to you under the licensing terms for the Microsoft product.
 *  Microsoft reserves all other rights not expressly granted under this agreement, whether by implication,
 *  estoppel or otherwise.
 *--------------------------------------------------------------------------------------------*/
/*---------------------------------------------------------------------------------------------
 *  Copyright © 2015 W3C® (MIT, ERCIM, Keio, Beihang). This software or document includes includes material copied
 *  from or derived from HTML 5.1 W3C Working Draft (http://www.w3.org/TR/2015/WD-html51-20151008/.)"
 *--------------------------------------------------------------------------------------------*/
/*---------------------------------------------------------------------------------------------
 *  Ionic Main Site (https://github.com/driftyco/ionic-site).
 *  Copyright Drifty Co. http://drifty.com/.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
 *  except in compliance with the License. You may obtain a copy of the License at
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  THIS CODE IS PROVIDED ON AN *AS IS* BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 *  KIND, EITHER EXPRESS OR IMPLIED, INCLUDING WITHOUT LIMITATION ANY IMPLIED
 *  WARRANTIES OR CONDITIONS OF TITLE, FITNESS FOR A PARTICULAR PURPOSE,
 *  MERCHANTABLITY OR NON-INFRINGEMENT.
 *
 *  See the Apache Version 2.0 License for specific language governing permissions
 *  and limitations under the License.
 *--------------------------------------------------------------------------------------------*/
define(__m[5], __M([1,0,6,4]), function (require, exports, strings, nls) {
    "use strict";
    var HTMLTagSpecification = (function () {
        function HTMLTagSpecification(label, attributes) {
            if (attributes === void 0) { attributes = []; }
            this.label = label;
            this.attributes = attributes;
        }
        return HTMLTagSpecification;
    }());
    exports.HTMLTagSpecification = HTMLTagSpecification;
    // HTML tag information sourced from http://www.w3.org/TR/2015/WD-html51-20151008/
    exports.HTML_TAGS = {
        // The root element
        html: new HTMLTagSpecification(nls.localize(0, null), ['manifest']),
        // Document metadata
        head: new HTMLTagSpecification(nls.localize(1, null)),
        title: new HTMLTagSpecification(nls.localize(2, null)),
        base: new HTMLTagSpecification(nls.localize(3, null), ['href', 'target']),
        link: new HTMLTagSpecification(nls.localize(4, null), ['href', 'crossorigin:xo', 'rel', 'media', 'hreflang', 'type', 'sizes']),
        meta: new HTMLTagSpecification(nls.localize(5, null), ['name', 'http-equiv', 'content', 'charset']),
        style: new HTMLTagSpecification(nls.localize(6, null), ['media', 'nonce', 'type', 'scoped:v']),
        // Sections
        body: new HTMLTagSpecification(nls.localize(7, null), ['onafterprint', 'onbeforeprint', 'onbeforeunload', 'onhashchange', 'onlanguagechange', 'onmessage', 'onoffline', 'ononline', 'onpagehide', 'onpageshow', 'onpopstate', 'onstorage', 'onunload']),
        article: new HTMLTagSpecification(nls.localize(8, null)),
        section: new HTMLTagSpecification(nls.localize(9, null)),
        nav: new HTMLTagSpecification(nls.localize(10, null)),
        aside: new HTMLTagSpecification(nls.localize(11, null)),
        h1: new HTMLTagSpecification(nls.localize(12, null)),
        h2: new HTMLTagSpecification(nls.localize(13, null)),
        h3: new HTMLTagSpecification(nls.localize(14, null)),
        h4: new HTMLTagSpecification(nls.localize(15, null)),
        h5: new HTMLTagSpecification(nls.localize(16, null)),
        h6: new HTMLTagSpecification(nls.localize(17, null)),
        header: new HTMLTagSpecification(nls.localize(18, null)),
        footer: new HTMLTagSpecification(nls.localize(19, null)),
        address: new HTMLTagSpecification(nls.localize(20, null)),
        // Grouping content
        p: new HTMLTagSpecification(nls.localize(21, null)),
        hr: new HTMLTagSpecification(nls.localize(22, null)),
        pre: new HTMLTagSpecification(nls.localize(23, null)),
        blockquote: new HTMLTagSpecification(nls.localize(24, null), ['cite']),
        ol: new HTMLTagSpecification(nls.localize(25, null), ['reversed:v', 'start', 'type:lt']),
        ul: new HTMLTagSpecification(nls.localize(26, null)),
        li: new HTMLTagSpecification(nls.localize(27, null), ['value']),
        dl: new HTMLTagSpecification(nls.localize(28, null)),
        dt: new HTMLTagSpecification(nls.localize(29, null)),
        dd: new HTMLTagSpecification(nls.localize(30, null)),
        figure: new HTMLTagSpecification(nls.localize(31, null)),
        figcaption: new HTMLTagSpecification(nls.localize(32, null)),
        main: new HTMLTagSpecification(nls.localize(33, null)),
        div: new HTMLTagSpecification(nls.localize(34, null)),
        // Text-level semantics
        a: new HTMLTagSpecification(nls.localize(35, null), ['href', 'target', 'download', 'ping', 'rel', 'hreflang', 'type']),
        em: new HTMLTagSpecification(nls.localize(36, null)),
        strong: new HTMLTagSpecification(nls.localize(37, null)),
        small: new HTMLTagSpecification(nls.localize(38, null)),
        s: new HTMLTagSpecification(nls.localize(39, null)),
        cite: new HTMLTagSpecification(nls.localize(40, null)),
        q: new HTMLTagSpecification(nls.localize(41, null), ['cite']),
        dfn: new HTMLTagSpecification(nls.localize(42, null)),
        abbr: new HTMLTagSpecification(nls.localize(43, null)),
        ruby: new HTMLTagSpecification(nls.localize(44, null)),
        rb: new HTMLTagSpecification(nls.localize(45, null)),
        rt: new HTMLTagSpecification(nls.localize(46, null)),
        // <rtc> is not yet supported by 2+ browsers
        //rtc: new HTMLTagSpecification(
        //	nls.localize('tags.rtc', 'The rtc element marks a ruby text container for ruby text components in a ruby annotation. When it is the child of a ruby element it doesn\'t represent anything itself, but its parent ruby element uses it as part of determining what it represents.')),
        rp: new HTMLTagSpecification(nls.localize(47, null)),
        // <data> is not yet supported by 2+ browsers
        //data: new HTMLTagSpecification(
        //	nls.localize('tags.data', 'The data element represents its contents, along with a machine-readable form of those contents in the value attribute.')),
        time: new HTMLTagSpecification(nls.localize(48, null), ['datetime']),
        code: new HTMLTagSpecification(nls.localize(49, null)),
        var: new HTMLTagSpecification(nls.localize(50, null)),
        samp: new HTMLTagSpecification(nls.localize(51, null)),
        kbd: new HTMLTagSpecification(nls.localize(52, null)),
        sub: new HTMLTagSpecification(nls.localize(53, null)),
        sup: new HTMLTagSpecification(nls.localize(54, null)),
        i: new HTMLTagSpecification(nls.localize(55, null)),
        b: new HTMLTagSpecification(nls.localize(56, null)),
        u: new HTMLTagSpecification(nls.localize(57, null)),
        mark: new HTMLTagSpecification(nls.localize(58, null)),
        bdi: new HTMLTagSpecification(nls.localize(59, null)),
        bdo: new HTMLTagSpecification(nls.localize(60, null)),
        span: new HTMLTagSpecification(nls.localize(61, null)),
        br: new HTMLTagSpecification(nls.localize(62, null)),
        wbr: new HTMLTagSpecification(nls.localize(63, null)),
        // Edits
        ins: new HTMLTagSpecification(nls.localize(64, null)),
        del: new HTMLTagSpecification(nls.localize(65, null), ['cite', 'datetime']),
        // Embedded content
        picture: new HTMLTagSpecification(nls.localize(66, null)),
        img: new HTMLTagSpecification(nls.localize(67, null), ['alt', 'src', 'srcset', 'crossorigin:xo', 'usemap', 'ismap:v', 'width', 'height']),
        iframe: new HTMLTagSpecification(nls.localize(68, null), ['src', 'srcdoc', 'name', 'sandbox:sb', 'seamless:v', 'allowfullscreen:v', 'width', 'height']),
        embed: new HTMLTagSpecification(nls.localize(69, null), ['src', 'type', 'width', 'height']),
        object: new HTMLTagSpecification(nls.localize(70, null), ['data', 'type', 'typemustmatch:v', 'name', 'usemap', 'form', 'width', 'height']),
        param: new HTMLTagSpecification(nls.localize(71, null), ['name', 'value']),
        video: new HTMLTagSpecification(nls.localize(72, null), ['src', 'crossorigin:xo', 'poster', 'preload:pl', 'autoplay:v', 'mediagroup', 'loop:v', 'muted:v', 'controls:v', 'width', 'height']),
        audio: new HTMLTagSpecification(nls.localize(73, null), ['src', 'crossorigin:xo', 'preload:pl', 'autoplay:v', 'mediagroup', 'loop:v', 'muted:v', 'controls:v']),
        source: new HTMLTagSpecification(nls.localize(74, null), 
        // 'When the source element has a parent that is a picture element, the source element allows authors to specify multiple alternative source sets for img elements.'
        ['src', 'type']),
        track: new HTMLTagSpecification(nls.localize(75, null), ['default:v', 'kind:tk', 'label', 'src', 'srclang']),
        map: new HTMLTagSpecification(nls.localize(76, null), ['name']),
        area: new HTMLTagSpecification(nls.localize(77, null), ['alt', 'coords', 'shape:sh', 'href', 'target', 'download', 'ping', 'rel', 'hreflang', 'type']),
        // Tabular data
        table: new HTMLTagSpecification(nls.localize(78, null), ['sortable:v', 'border']),
        caption: new HTMLTagSpecification(nls.localize(79, null)),
        colgroup: new HTMLTagSpecification(nls.localize(80, null), ['span']),
        col: new HTMLTagSpecification(nls.localize(81, null), ['span']),
        tbody: new HTMLTagSpecification(nls.localize(82, null)),
        thead: new HTMLTagSpecification(nls.localize(83, null)),
        tfoot: new HTMLTagSpecification(nls.localize(84, null)),
        tr: new HTMLTagSpecification(nls.localize(85, null)),
        td: new HTMLTagSpecification(nls.localize(86, null), ['colspan', 'rowspan', 'headers']),
        th: new HTMLTagSpecification(nls.localize(87, null), ['colspan', 'rowspan', 'headers', 'scope:s', 'sorted', 'abbr']),
        // Forms
        form: new HTMLTagSpecification(nls.localize(88, null), ['accept-charset', 'action', 'autocomplete:o', 'enctype:et', 'method:m', 'name', 'novalidate:v', 'target']),
        label: new HTMLTagSpecification(nls.localize(89, null), ['form', 'for']),
        input: new HTMLTagSpecification(nls.localize(90, null), ['accept', 'alt', 'autocomplete:inputautocomplete', 'autofocus:v', 'checked:v', 'dirname', 'disabled:v', 'form', 'formaction', 'formenctype:et', 'formmethod:fm', 'formnovalidate:v', 'formtarget', 'height', 'inputmode:im', 'list', 'max', 'maxlength', 'min', 'minlength', 'multiple:v', 'name', 'pattern', 'placeholder', 'readonly:v', 'required:v', 'size', 'src', 'step', 'type:t', 'value', 'width']),
        button: new HTMLTagSpecification(nls.localize(91, null), ['autofocus:v', 'disabled:v', 'form', 'formaction', 'formenctype:et', 'formmethod:fm', 'formnovalidate:v', 'formtarget', 'name', 'type:bt', 'value']),
        select: new HTMLTagSpecification(nls.localize(92, null), ['autocomplete:inputautocomplete', 'autofocus:v', 'disabled:v', 'form', 'multiple:v', 'name', 'required:v', 'size']),
        datalist: new HTMLTagSpecification(nls.localize(93, null)),
        optgroup: new HTMLTagSpecification(nls.localize(94, null), ['disabled:v', 'label']),
        option: new HTMLTagSpecification(nls.localize(95, null), ['disabled:v', 'label', 'selected:v', 'value']),
        textarea: new HTMLTagSpecification(nls.localize(96, null), ['autocomplete:inputautocomplete', 'autofocus:v', 'cols', 'dirname', 'disabled:v', 'form', 'inputmode:im', 'maxlength', 'minlength', 'name', 'placeholder', 'readonly:v', 'required:v', 'rows', 'wrap:w']),
        output: new HTMLTagSpecification(nls.localize(97, null), ['for', 'form', 'name']),
        progress: new HTMLTagSpecification(nls.localize(98, null), ['value', 'max']),
        meter: new HTMLTagSpecification(nls.localize(99, null), ['value', 'min', 'max', 'low', 'high', 'optimum']),
        fieldset: new HTMLTagSpecification(nls.localize(100, null), ['disabled:v', 'form', 'name']),
        legend: new HTMLTagSpecification(nls.localize(101, null)),
        // Interactive elements
        details: new HTMLTagSpecification(nls.localize(102, null), ['open:v']),
        summary: new HTMLTagSpecification(nls.localize(103, null)),
        // <menu> and <menuitem> are not yet supported by 2+ browsers
        //menu: new HTMLTagSpecification(
        //	nls.localize('tags.menu', 'The menu element represents a list of commands.'),
        //	['type:mt', 'label']),
        //menuitem: new HTMLTagSpecification(
        //	nls.localize('tags.menuitem', 'The menuitem element represents a command that the user can invoke from a popup menu (either a context menu or the menu of a menu button).')),
        dialog: new HTMLTagSpecification(nls.localize(104, null)),
        // Scripting
        script: new HTMLTagSpecification(nls.localize(105, null), ['src', 'type', 'charset', 'async:v', 'defer:v', 'crossorigin:xo', 'nonce']),
        noscript: new HTMLTagSpecification(nls.localize(106, null)),
        template: new HTMLTagSpecification(nls.localize(107, null)),
        canvas: new HTMLTagSpecification(nls.localize(108, null), ['width', 'height'])
    };
    // Ionic tag information sourced from Ionic main website (https://github.com/driftyco/ionic-site)
    exports.IONIC_TAGS = {
        'ion-checkbox': new HTMLTagSpecification(nls.localize(109, null), ['name', 'ng-false-value', 'ng-model', 'ng-true-value']),
        'ion-content': new HTMLTagSpecification(nls.localize(110, null), ['delegate-handle', 'direction:scrolldir', 'has-bouncing:b', 'locking:b', 'on-scroll', 'on-scroll-complete', 'overflow-scroll:b', 'padding:b', 'scroll:b', 'scrollbar-x:b', 'scrollbar-y:b', 'start-x', 'start-y']),
        'ion-delete-button': new HTMLTagSpecification(nls.localize(111, null), []),
        'ion-footer-bar': new HTMLTagSpecification(nls.localize(112, null), ['align-title:align', 'keyboard-attach:v']),
        'ion-header-bar': new HTMLTagSpecification(nls.localize(113, null), ['align-title:align', 'no-tap-scroll:b']),
        'ion-infinite-scroll': new HTMLTagSpecification(nls.localize(114, null), ['distance', 'icon', 'immediate-check:b', 'on-infinite', 'spinner']),
        'ion-input': new HTMLTagSpecification(nls.localize(115, null), ['type:inputtype', 'clearInput:v']),
        'ion-item': new HTMLTagSpecification(nls.localize(116, null), []),
        'ion-list': new HTMLTagSpecification(nls.localize(117, null), ['can-swipe:b', 'delegate-handle', 'show-delete:b', 'show-reorder:b', 'type:listtype']),
        'ion-modal-view': new HTMLTagSpecification(nls.localize(118, null), []),
        'ion-nav-back-button': new HTMLTagSpecification(nls.localize(119, null), []),
        'ion-nav-bar': new HTMLTagSpecification(nls.localize(120, null), ['align-title:align', 'delegate-handle', 'no-tap-scroll:b']),
        'ion-nav-buttons': new HTMLTagSpecification(nls.localize(121, null), ['side:navsides']),
        'ion-nav-title': new HTMLTagSpecification(nls.localize(122, null), []),
        'ion-nav-view': new HTMLTagSpecification(nls.localize(123, null), ['name']),
        'ion-option-button': new HTMLTagSpecification(nls.localize(124, null), []),
        'ion-pane': new HTMLTagSpecification(nls.localize(125, null), []),
        'ion-popover-view': new HTMLTagSpecification(nls.localize(126, null), []),
        'ion-radio': new HTMLTagSpecification(nls.localize(127, null), ['disabled:b', 'icon', 'name', 'ng-disabled:b', 'ng-model', 'ng-value', 'value']),
        'ion-refresher': new HTMLTagSpecification(nls.localize(128, null), ['disable-pulling-rotation:b', 'on-pulling', 'on-refresh', 'pulling-icon', 'pulling-text', 'refreshing-icon', 'spinner']),
        'ion-reorder-button': new HTMLTagSpecification(nls.localize(129, null), ['on-reorder']),
        'ion-scroll': new HTMLTagSpecification(nls.localize(130, null), ['delegate-handle', 'direction:scrolldir', 'has-bouncing:b', 'locking:b', 'max-zoom', 'min-zoom', 'on-refresh', 'on-scroll', 'paging:b', 'scrollbar-x:b', 'scrollbar-y:b', 'zooming:b']),
        'ion-side-menu': new HTMLTagSpecification(nls.localize(131, null), ['is-enabled:b', 'expose-aside-when', 'side:navsides', 'width']),
        'ion-side-menu-content': new HTMLTagSpecification(nls.localize(132, null), ['drag-content:b', 'edge-drag-threshold']),
        'ion-side-menus': new HTMLTagSpecification(nls.localize(133, null), ['delegate-handle', 'enable-menu-with-back-views:b']),
        'ion-slide': new HTMLTagSpecification(nls.localize(134, null), []),
        'ion-slide-box': new HTMLTagSpecification(nls.localize(135, null), ['active-slide', 'auto-play:b', 'delegate-handle', 'does-continue:b', 'on-slide-changed', 'pager-click', 'show-pager:b', 'slide-interval']),
        'ion-spinner': new HTMLTagSpecification(nls.localize(136, null), ['icon']),
        'ion-tab': new HTMLTagSpecification(nls.localize(137, null), ['badge', 'badge-style', 'disabled', 'hidden', 'href', 'icon', 'icon-off', 'icon-on', 'ng-click', 'on-deselect', 'on-select', 'title']),
        'ion-tabs': new HTMLTagSpecification(nls.localize(138, null), ['delegate-handle']),
        'ion-title': new HTMLTagSpecification(nls.localize(139, null), []),
        'ion-toggle': new HTMLTagSpecification(nls.localize(140, null), ['name', 'ng-false-value', 'ng-model', 'ng-true-value', 'toggle-class']),
        'ion-view ': new HTMLTagSpecification(nls.localize(141, null), ['cache-view:b', 'can-swipe-back:b', 'hide-back-button:b', 'hide-nav-bar:b', 'view-title'])
    };
    function getHTML5TagProvider() {
        var globalAttributes = [
            'aria-activedescendant', 'aria-atomic:b', 'aria-autocomplete:autocomplete', 'aria-busy:b', 'aria-checked:tristate', 'aria-colcount', 'aria-colindex', 'aria-colspan', 'aria-controls', 'aria-current:current', 'aria-describedat',
            'aria-describedby', 'aria-disabled:b', 'aria-dropeffect:dropeffect', 'aria-errormessage', 'aria-expanded:u', 'aria-flowto', 'aria-grabbed:u', 'aria-haspopup:b', 'aria-hidden:b', 'aria-invalid:invalid', 'aria-kbdshortcuts',
            'aria-label', 'aria-labelledby', 'aria-level', 'aria-live:live', 'aria-modal:b', 'aria-multiline:b', 'aria-multiselectable:b', 'aria-orientation:orientation', 'aria-owns', 'aria-placeholder', 'aria-posinset', 'aria-pressed:tristate',
            'aria-readonly:b', 'aria-relevant:relevant', 'aria-required:b', 'aria-roledescription', 'aria-rowcount', 'aria-rowindex', 'aria-rowspan', 'aria-selected:u', 'aria-setsize', 'aria-sort:sort', 'aria-valuemax', 'aria-valuemin', 'aria-valuenow', 'aria-valuetext',
            'accesskey', 'class', 'contenteditable:b', 'contextmenu', 'dir:d', 'draggable:b', 'dropzone', 'hidden:v', 'id', 'itemid', 'itemprop', 'itemref', 'itemscope:v', 'itemtype', 'lang', 'role:roles', 'spellcheck:b', 'style', 'tabindex',
            'title', 'translate:y'];
        var eventHandlers = ['onabort', 'onblur', 'oncanplay', 'oncanplaythrough', 'onchange', 'onclick', 'oncontextmenu', 'ondblclick', 'ondrag', 'ondragend', 'ondragenter', 'ondragleave', 'ondragover', 'ondragstart',
            'ondrop', 'ondurationchange', 'onemptied', 'onended', 'onerror', 'onfocus', 'onformchange', 'onforminput', 'oninput', 'oninvalid', 'onkeydown', 'onkeypress', 'onkeyup', 'onload', 'onloadeddata', 'onloadedmetadata',
            'onloadstart', 'onmousedown', 'onmousemove', 'onmouseout', 'onmouseover', 'onmouseup', 'onmousewheel', 'onpause', 'onplay', 'onplaying', 'onprogress', 'onratechange', 'onreset', 'onresize', 'onreadystatechange', 'onscroll',
            'onseeked', 'onseeking', 'onselect', 'onshow', 'onstalled', 'onsubmit', 'onsuspend', 'ontimeupdate', 'onvolumechange', 'onwaiting'];
        var valueSets = {
            b: ['true', 'false'],
            u: ['true', 'false', 'undefined'],
            o: ['on', 'off'],
            y: ['yes', 'no'],
            w: ['soft', 'hard'],
            d: ['ltr', 'rtl', 'auto'],
            m: ['GET', 'POST', 'dialog'],
            fm: ['GET', 'POST'],
            s: ['row', 'col', 'rowgroup', 'colgroup'],
            t: ['hidden', 'text', 'search', 'tel', 'url', 'email', 'password', 'datetime', 'date', 'month', 'week', 'time', 'datetime-local', 'number', 'range', 'color', 'checkbox', 'radio', 'file', 'submit', 'image', 'reset', 'button'],
            im: ['verbatim', 'latin', 'latin-name', 'latin-prose', 'full-width-latin', 'kana', 'kana-name', 'katakana', 'numeric', 'tel', 'email', 'url'],
            bt: ['button', 'submit', 'reset', 'menu'],
            lt: ['1', 'a', 'A', 'i', 'I'],
            mt: ['context', 'toolbar'],
            mit: ['command', 'checkbox', 'radio'],
            et: ['application/x-www-form-urlencoded', 'multipart/form-data', 'text/plain'],
            tk: ['subtitles', 'captions', 'descriptions', 'chapters', 'metadata'],
            pl: ['none', 'metadata', 'auto'],
            sh: ['circle', 'default', 'poly', 'rect'],
            xo: ['anonymous', 'use-credentials'],
            sb: ['allow-forms', 'allow-modals', 'allow-pointer-lock', 'allow-popups', 'allow-popups-to-escape-sandbox', 'allow-same-origin', 'allow-scripts', 'allow-top-navigation'],
            tristate: ['true', 'false', 'mixed', 'undefined'],
            inputautocomplete: ['additional-name', 'address-level1', 'address-level2', 'address-level3', 'address-level4', 'address-line1', 'address-line2', 'address-line3', 'bday', 'bday-year', 'bday-day', 'bday-month', 'billing', 'cc-additional-name', 'cc-csc', 'cc-exp', 'cc-exp-month', 'cc-exp-year', 'cc-family-name', 'cc-given-name', 'cc-name', 'cc-number', 'cc-type', 'country', 'country-name', 'current-password', 'email', 'family-name', 'fax', 'given-name', 'home', 'honorific-prefix', 'honorific-suffix', 'impp', 'language', 'mobile', 'name', 'new-password', 'nickname', 'organization', 'organization-title', 'pager', 'photo', 'postal-code', 'sex', 'shipping', 'street-address', 'tel-area-code', 'tel', 'tel-country-code', 'tel-extension', 'tel-local', 'tel-local-prefix', 'tel-local-suffix', 'tel-national', 'transaction-amount', 'transaction-currency', 'url', 'username', 'work'],
            autocomplete: ['inline', 'list', 'both', 'none'],
            current: ['page', 'step', 'location', 'date', 'time', 'true', 'false'],
            dropeffect: ['copy', 'move', 'link', 'execute', 'popup', 'none'],
            invalid: ['grammar', 'false', 'spelling', 'true'],
            live: ['off', 'polite', 'assertive'],
            orientation: ['vertical', 'horizontal', 'undefined'],
            relevant: ['additions', 'removals', 'text', 'all', 'additions text'],
            sort: ['ascending', 'descending', 'none', 'other'],
            roles: ['alert', 'alertdialog', 'button', 'checkbox', 'dialog', 'gridcell', 'link', 'log', 'marquee', 'menuitem', 'menuitemcheckbox', 'menuitemradio', 'option', 'progressbar', 'radio', 'scrollbar', 'searchbox', 'slider',
                'spinbutton', 'status', 'switch', 'tab', 'tabpanel', 'textbox', 'timer', 'tooltip', 'treeitem', 'combobox', 'grid', 'listbox', 'menu', 'menubar', 'radiogroup', 'tablist', 'tree', 'treegrid',
                'application', 'article', 'cell', 'columnheader', 'definition', 'directory', 'document', 'feed', 'figure', 'group', 'heading', 'img', 'list', 'listitem', 'math', 'none', 'note', 'presentation', 'region', 'row', 'rowgroup',
                'rowheader', 'separator', 'table', 'term', 'text', 'toolbar',
                'banner', 'complementary', 'contentinfo', 'form', 'main', 'navigation', 'region', 'search']
        };
        return {
            collectTags: function (collector) { return collectTagsDefault(collector, exports.HTML_TAGS); },
            collectAttributes: function (tag, collector) {
                collectAttributesDefault(tag, collector, exports.HTML_TAGS, globalAttributes);
                eventHandlers.forEach(function (handler) {
                    collector(handler, 'event');
                });
            },
            collectValues: function (tag, attribute, collector) { return collectValuesDefault(tag, attribute, collector, exports.HTML_TAGS, globalAttributes, valueSets); }
        };
    }
    exports.getHTML5TagProvider = getHTML5TagProvider;
    function getAngularTagProvider() {
        var customTags = {
            input: ['ng-model', 'ng-required', 'ng-minlength', 'ng-maxlength', 'ng-pattern', 'ng-trim'],
            select: ['ng-model'],
            textarea: ['ng-model', 'ng-required', 'ng-minlength', 'ng-maxlength', 'ng-pattern', 'ng-trim']
        };
        var globalAttributes = ['ng-app', 'ng-bind', 'ng-bind-html', 'ng-bind-template', 'ng-blur', 'ng-change', 'ng-checked', 'ng-class', 'ng-class-even', 'ng-class-odd',
            'ng-click', 'ng-cloak', 'ng-controller', 'ng-copy', 'ng-csp', 'ng-cut', 'ng-dblclick', 'ng-disabled', 'ng-focus', 'ng-form', 'ng-hide', 'ng-href', 'ng-if',
            'ng-include', 'ng-init', 'ng-jq', 'ng-keydown', 'ng-keypress', 'ng-keyup', 'ng-list', 'ng-model-options', 'ng-mousedown', 'ng-mouseenter', 'ng-mouseleave',
            'ng-mousemove', 'ng-mouseover', 'ng-mouseup', 'ng-non-bindable', 'ng-open', 'ng-options', 'ng-paste', 'ng-pluralize', 'ng-readonly', 'ng-repeat', 'ng-selected',
            'ng-show', 'ng-src', 'ng-srcset', 'ng-style', 'ng-submit', 'ng-switch', 'ng-transclude', 'ng-value'
        ];
        return {
            collectTags: function (collector) {
                // no extra tags
            },
            collectAttributes: function (tag, collector) {
                if (tag) {
                    var attributes = customTags[tag];
                    if (attributes) {
                        attributes.forEach(function (a) {
                            collector(a, null);
                            collector('data-' + a, null);
                        });
                    }
                }
                globalAttributes.forEach(function (a) {
                    collector(a, null);
                    collector('data-' + a, null);
                });
            },
            collectValues: function (tag, attribute, collector) {
                // no values
            }
        };
    }
    exports.getAngularTagProvider = getAngularTagProvider;
    function getIonicTagProvider() {
        var customTags = {
            a: ['nav-direction:navdir', 'nav-transition:trans'],
            button: ['menu-toggle:menusides']
        };
        var globalAttributes = ['collection-repeat', 'force-refresh-images:b', 'ion-stop-event', 'item-height', 'item-render-buffer', 'item-width', 'menu-close:v',
            'on-double-tap', 'on-drag', 'on-drag-down', 'on-drag-left', 'on-drag-right', 'on-drag-up', 'on-hold', 'on-release', 'on-swipe', 'on-swipe-down', 'on-swipe-left',
            'on-swipe-right', 'on-swipe-up', 'on-tap', 'on-touch'];
        var valueSets = {
            align: ['center', 'left', 'right'],
            b: ['true', 'false'],
            inputtype: ['email', 'number', 'password', 'search', 'tel', 'text', 'url'],
            listtype: ['card', 'list-inset'],
            menusides: ['left', 'right'],
            navdir: ['back', 'enter', 'exit', 'forward', 'swap'],
            navsides: ['left', 'primary', 'right', 'secondary'],
            scrolldir: ['x', 'xy', 'y'],
            trans: ['android', 'ios', 'none']
        };
        return {
            collectTags: function (collector) { return collectTagsDefault(collector, exports.IONIC_TAGS); },
            collectAttributes: function (tag, collector) {
                collectAttributesDefault(tag, collector, exports.IONIC_TAGS, globalAttributes);
                if (tag) {
                    var attributes = customTags[tag];
                    if (attributes) {
                        attributes.forEach(function (a) {
                            var segments = a.split(':');
                            collector(segments[0], segments[1]);
                        });
                    }
                }
            },
            collectValues: function (tag, attribute, collector) { return collectValuesDefault(tag, attribute, collector, exports.IONIC_TAGS, globalAttributes, valueSets, customTags); }
        };
    }
    exports.getIonicTagProvider = getIonicTagProvider;
    function collectTagsDefault(collector, tagSet) {
        for (var tag in tagSet) {
            collector(tag, tagSet[tag].label);
        }
    }
    function collectAttributesDefault(tag, collector, tagSet, globalAttributes) {
        globalAttributes.forEach(function (attr) {
            var segments = attr.split(':');
            collector(segments[0], segments[1]);
        });
        if (tag) {
            var tags = tagSet[tag];
            if (tags) {
                var attributes = tags.attributes;
                if (attributes) {
                    attributes.forEach(function (attr) {
                        var segments = attr.split(':');
                        collector(segments[0], segments[1]);
                    });
                }
            }
        }
    }
    function collectValuesDefault(tag, attribute, collector, tagSet, globalAttributes, valueSets, customTags) {
        var prefix = attribute + ':';
        var processAttributes = function (attributes) {
            attributes.forEach(function (attr) {
                if (attr.length > prefix.length && strings.startsWith(attr, prefix)) {
                    var typeInfo = attr.substr(prefix.length);
                    if (typeInfo === 'v') {
                        collector(attribute);
                    }
                    else {
                        var values = valueSets[typeInfo];
                        if (values) {
                            values.forEach(collector);
                        }
                    }
                }
            });
        };
        if (tag) {
            var tags = tagSet[tag];
            if (tags) {
                var attributes = tags.attributes;
                if (attributes) {
                    processAttributes(attributes);
                }
            }
        }
        processAttributes(globalAttributes);
        if (customTags) {
            var customTagAttributes = customTags[tag];
            if (customTagAttributes) {
                processAttributes(customTagAttributes);
            }
        }
    }
});
/*!
END THIRD PARTY
*/ 

var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __param = (this && this.__param) || function (paramIndex, decorator) {
    return function (target, key) { decorator(target, key, paramIndex); }
};
define(__m[9], __M([1,0,19,11,12,5,13,14,6,15,16,2,3,17,18,7]), function (require, exports, uri_1, winjs, beautifyHTML, htmlTags, network, modes, strings, workspace_1, resourceService_1, htmlScanner_1, htmlTokenTypes_1, htmlEmptyTagsShared_1, suggestSupport_1, paths) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var LinkDetectionState;
    (function (LinkDetectionState) {
        LinkDetectionState[LinkDetectionState["LOOKING_FOR_HREF_OR_SRC"] = 1] = "LOOKING_FOR_HREF_OR_SRC";
        LinkDetectionState[LinkDetectionState["AFTER_HREF_OR_SRC"] = 2] = "AFTER_HREF_OR_SRC";
    })(LinkDetectionState || (LinkDetectionState = {}));
    var HTMLWorker = (function () {
        function HTMLWorker(modeId, resourceService, contextService) {
            this._modeId = modeId;
            this.resourceService = resourceService;
            this._contextService = contextService;
            this._tagProviders = [];
            this._tagProviders.push(htmlTags.getHTML5TagProvider());
            this.addCustomTagProviders(this._tagProviders);
        }
        HTMLWorker.prototype.addCustomTagProviders = function (providers) {
            providers.push(htmlTags.getAngularTagProvider());
            providers.push(htmlTags.getIonicTagProvider());
        };
        HTMLWorker.prototype.provideDocumentRangeFormattingEdits = function (resource, range, options) {
            return this.formatHTML(resource, range, options);
        };
        HTMLWorker.prototype.formatHTML = function (resource, range, options) {
            var model = this.resourceService.get(resource);
            var value = range ? model.getValueInRange(range) : model.getValue();
            var htmlOptions = {
                indent_size: options.insertSpaces ? options.tabSize : 1,
                indent_char: options.insertSpaces ? ' ' : '\t',
                wrap_line_length: this.getFormatOption('wrapLineLength', 120),
                unformatted: this.getTagsFormatOption('unformatted', void 0),
                indent_inner_html: this.getFormatOption('indentInnerHtml', false),
                preserve_newlines: this.getFormatOption('preserveNewLines', false),
                max_preserve_newlines: this.getFormatOption('maxPreserveNewLines', void 0),
                indent_handlebars: this.getFormatOption('indentHandlebars', false),
                end_with_newline: this.getFormatOption('endWithNewline', false),
                extra_liners: this.getTagsFormatOption('extraLiners', void 0),
            };
            var result = beautifyHTML.html_beautify(value, htmlOptions);
            return winjs.TPromise.as([{
                    range: range,
                    text: result
                }]);
        };
        HTMLWorker.prototype.getFormatOption = function (key, dflt) {
            if (this.formatSettings && this.formatSettings.hasOwnProperty(key)) {
                var value = this.formatSettings[key];
                if (value !== null) {
                    return value;
                }
            }
            return dflt;
        };
        HTMLWorker.prototype.getTagsFormatOption = function (key, dflt) {
            var list = this.getFormatOption(key, null);
            if (typeof list === 'string') {
                if (list.length > 0) {
                    return list.split(',').map(function (t) { return t.trim().toLowerCase(); });
                }
                return [];
            }
            return dflt;
        };
        HTMLWorker.prototype._doConfigure = function (options) {
            this.formatSettings = options && options.format;
            return winjs.TPromise.as(null);
        };
        HTMLWorker.prototype.findMatchingOpenTag = function (scanner) {
            var closedTags = {};
            var tagClosed = false;
            while (scanner.scanBack()) {
                if (htmlTokenTypes_1.isTag(scanner.getTokenType()) && !tagClosed) {
                    var tag = scanner.getTokenContent();
                    scanner.scanBack();
                    if (scanner.getTokenType() === htmlTokenTypes_1.DELIM_END) {
                        closedTags[tag] = (closedTags[tag] || 0) + 1;
                    }
                    else if (!htmlEmptyTagsShared_1.isEmptyElement(tag)) {
                        if (closedTags[tag]) {
                            closedTags[tag]--;
                        }
                        else {
                            return tag;
                        }
                    }
                }
                else if (scanner.getTokenType() === htmlTokenTypes_1.DELIM_START) {
                    tagClosed = scanner.getTokenContent() === '/>';
                }
            }
            return null;
        };
        HTMLWorker.prototype.collectTagSuggestions = function (scanner, position, suggestions) {
            var _this = this;
            var model = scanner.getModel();
            var currentLine = model.getLineContent(position.lineNumber);
            var contentAfter = currentLine.substr(position.column - 1);
            var closeTag = isWhiteSpace(contentAfter) || strings.startsWith(contentAfter, '<') ? '>' : '';
            var collectClosingTagSuggestion = function (overwriteBefore) {
                var endPosition = scanner.getTokenPosition();
                var matchingTag = _this.findMatchingOpenTag(scanner);
                if (matchingTag) {
                    var suggestion = {
                        label: '/' + matchingTag,
                        codeSnippet: '/' + matchingTag + closeTag,
                        overwriteBefore: overwriteBefore,
                        type: 'property'
                    };
                    suggestions.suggestions.push(suggestion);
                    // use indent from start tag
                    var startPosition = scanner.getTokenPosition();
                    if (endPosition.lineNumber !== startPosition.lineNumber) {
                        var startIndent = model.getLineContent(startPosition.lineNumber).substring(0, startPosition.column - 1);
                        var endIndent = model.getLineContent(endPosition.lineNumber).substring(0, endPosition.column - 1);
                        if (isWhiteSpace(startIndent) && isWhiteSpace(endIndent)) {
                            suggestion.overwriteBefore = position.column - 1; // replace from start of line
                            suggestion.codeSnippet = startIndent + '</' + matchingTag + closeTag;
                            suggestion.filterText = currentLine.substring(0, position.column - 1);
                        }
                    }
                    return true;
                }
                return false;
            };
            if (scanner.getTokenType() === htmlTokenTypes_1.DELIM_END) {
                var hasClose = collectClosingTagSuggestion(suggestions.currentWord.length + 1);
                if (!hasClose) {
                    this._tagProviders.forEach(function (provider) {
                        provider.collectTags(function (tag, label) {
                            suggestions.suggestions.push({
                                label: '/' + tag,
                                overwriteBefore: suggestions.currentWord.length + 1,
                                codeSnippet: '/' + tag + closeTag,
                                type: 'property',
                                documentationLabel: label
                            });
                        });
                    });
                }
            }
            else {
                collectClosingTagSuggestion(suggestions.currentWord.length);
                this._tagProviders.forEach(function (provider) {
                    provider.collectTags(function (tag, label) {
                        suggestions.suggestions.push({
                            label: tag,
                            codeSnippet: tag,
                            type: 'property',
                            documentationLabel: label
                        });
                    });
                });
            }
        };
        HTMLWorker.prototype.collectContentSuggestions = function (suggestions) {
            // disable the simple snippets in favor of the emmet templates
        };
        HTMLWorker.prototype.collectAttributeSuggestions = function (scanner, suggestions) {
            var parentTag = null;
            do {
                if (htmlTokenTypes_1.isTag(scanner.getTokenType())) {
                    parentTag = scanner.getTokenContent();
                    break;
                }
                if (scanner.getTokenType() === htmlTokenTypes_1.DELIM_START) {
                    break;
                }
            } while (scanner.scanBack());
            this._tagProviders.forEach(function (provider) {
                provider.collectAttributes(parentTag, function (attribute, type) {
                    var codeSnippet = attribute;
                    if (type !== 'v') {
                        codeSnippet = codeSnippet + '="{{}}"';
                    }
                    suggestions.suggestions.push({
                        label: attribute,
                        codeSnippet: codeSnippet,
                        type: type === 'handler' ? 'function' : 'value'
                    });
                });
            });
        };
        HTMLWorker.prototype.collectAttributeValueSuggestions = function (scanner, suggestions) {
            var needsQuotes = scanner.getTokenType() === htmlTokenTypes_1.DELIM_ASSIGN;
            var attribute = null;
            var parentTag = null;
            while (scanner.scanBack()) {
                if (scanner.getTokenType() === htmlTokenTypes_1.ATTRIB_NAME) {
                    attribute = scanner.getTokenContent();
                    break;
                }
            }
            while (scanner.scanBack()) {
                if (htmlTokenTypes_1.isTag(scanner.getTokenType())) {
                    parentTag = scanner.getTokenContent();
                    break;
                }
                if (scanner.getTokenType() === htmlTokenTypes_1.DELIM_START) {
                    return;
                }
            }
            this._tagProviders.forEach(function (provider) {
                provider.collectValues(parentTag, attribute, function (value) {
                    suggestions.suggestions.push({
                        label: value,
                        codeSnippet: needsQuotes ? '"' + value + '"' : value,
                        type: 'unit'
                    });
                });
            });
        };
        HTMLWorker.prototype.provideCompletionItems = function (resource, position) {
            var model = this.resourceService.get(resource);
            var modeIdAtPosition = model.getModeIdAtPosition(position.lineNumber, position.column);
            if (modeIdAtPosition === this._modeId) {
                return this.suggestHTML(resource, position);
            }
            else {
                return winjs.TPromise.as([]);
            }
        };
        HTMLWorker.prototype.suggestHTML = function (resource, position) {
            return this.doSuggest(resource, position).then(function (value) { return suggestSupport_1.filterSuggestions(value); });
        };
        HTMLWorker.prototype.doSuggest = function (resource, position) {
            var model = this.resourceService.get(resource), currentWord = model.getWordUntilPosition(position).word;
            var suggestions = {
                currentWord: currentWord,
                suggestions: [],
            };
            var scanner = htmlScanner_1.getScanner(model, position);
            switch (scanner.getTokenType()) {
                case htmlTokenTypes_1.DELIM_START:
                case htmlTokenTypes_1.DELIM_END:
                    if (scanner.isOpenBrace()) {
                        this.collectTagSuggestions(scanner, position, suggestions);
                    }
                    else {
                        this.collectContentSuggestions(suggestions);
                    }
                    break;
                case htmlTokenTypes_1.ATTRIB_NAME:
                    this.collectAttributeSuggestions(scanner, suggestions);
                    break;
                case htmlTokenTypes_1.ATTRIB_VALUE:
                    this.collectAttributeValueSuggestions(scanner, suggestions);
                    break;
                case htmlTokenTypes_1.DELIM_ASSIGN:
                    if (scanner.isAtTokenEnd()) {
                        this.collectAttributeValueSuggestions(scanner, suggestions);
                    }
                    break;
                case '':
                    if (isWhiteSpace(scanner.getTokenContent()) && scanner.scanBack()) {
                        switch (scanner.getTokenType()) {
                            case htmlTokenTypes_1.ATTRIB_VALUE:
                            case htmlTokenTypes_1.ATTRIB_NAME:
                                this.collectAttributeSuggestions(scanner, suggestions);
                                break;
                            case htmlTokenTypes_1.DELIM_ASSIGN:
                                this.collectAttributeValueSuggestions(scanner, suggestions);
                                break;
                            case htmlTokenTypes_1.DELIM_START:
                            case htmlTokenTypes_1.DELIM_END:
                                if (scanner.isOpenBrace()) {
                                    this.collectTagSuggestions(scanner, position, suggestions);
                                }
                                else {
                                    this.collectContentSuggestions(suggestions);
                                }
                                break;
                            default:
                                if (htmlTokenTypes_1.isTag(scanner.getTokenType())) {
                                    this.collectAttributeSuggestions(scanner, suggestions);
                                }
                        }
                    }
                    else {
                        this.collectContentSuggestions(suggestions);
                    }
                    break;
                default:
                    if (htmlTokenTypes_1.isTag(scanner.getTokenType())) {
                        scanner.scanBack(); // one back to the end/start bracket
                        this.collectTagSuggestions(scanner, position, suggestions);
                    }
            }
            return winjs.TPromise.as(suggestions);
        };
        HTMLWorker.prototype.findMatchingBracket = function (tagname, scanner) {
            if (htmlEmptyTagsShared_1.isEmptyElement(tagname)) {
                return null;
            }
            var tagCount = 0;
            scanner.scanBack(); // one back to the end/start bracket
            if (scanner.getTokenType() === htmlTokenTypes_1.DELIM_END) {
                // find the opening tag
                var tagClosed = false;
                while (scanner.scanBack()) {
                    if (htmlTokenTypes_1.isTag(scanner.getTokenType()) && scanner.getTokenContent() === tagname && !tagClosed) {
                        var range = scanner.getTokenRange();
                        scanner.scanBack(); // one back to the end/start bracket
                        if (scanner.getTokenType() === htmlTokenTypes_1.DELIM_START) {
                            if (tagCount === 0) {
                                return range;
                            }
                            else {
                                tagCount--;
                            }
                        }
                        else {
                            tagCount++;
                        }
                    }
                    else if (scanner.getTokenType() === htmlTokenTypes_1.DELIM_START) {
                        tagClosed = scanner.getTokenContent() === '/>';
                    }
                }
            }
            else {
                var isTagEnd = false;
                while (scanner.scanForward()) {
                    if (htmlTokenTypes_1.isTag(scanner.getTokenType()) && scanner.getTokenContent() === tagname) {
                        if (!isTagEnd) {
                            scanner.scanForward();
                            if (scanner.getTokenType() === htmlTokenTypes_1.DELIM_START && scanner.getTokenContent() === '/>') {
                                if (tagCount <= 0) {
                                    return null;
                                }
                            }
                            else {
                                tagCount++;
                            }
                        }
                        else {
                            tagCount--;
                            if (tagCount <= 0) {
                                return scanner.getTokenRange();
                            }
                        }
                    }
                    else if (scanner.getTokenType() === htmlTokenTypes_1.DELIM_START) {
                        isTagEnd = false;
                    }
                    else if (scanner.getTokenType() === htmlTokenTypes_1.DELIM_END) {
                        isTagEnd = true;
                    }
                }
            }
            return null;
        };
        HTMLWorker.prototype.provideDocumentHighlights = function (resource, position, strict) {
            if (strict === void 0) { strict = false; }
            var model = this.resourceService.get(resource), wordAtPosition = model.getWordAtPosition(position), currentWord = (wordAtPosition ? wordAtPosition.word : ''), result = [];
            var scanner = htmlScanner_1.getScanner(model, position);
            if (htmlTokenTypes_1.isTag(scanner.getTokenType())) {
                var tagname = scanner.getTokenContent();
                result.push({
                    range: scanner.getTokenRange(),
                    kind: modes.DocumentHighlightKind.Read
                });
                var range = this.findMatchingBracket(tagname, scanner);
                if (range) {
                    result.push({
                        range: range,
                        kind: modes.DocumentHighlightKind.Read
                    });
                }
            }
            else {
                var words = model.getAllWordsWithRange(), upperBound = Math.min(1000, words.length); // Limit find occurences to 1000 occurences
                for (var i = 0; i < upperBound; i++) {
                    if (words[i].text === currentWord) {
                        result.push({
                            range: words[i].range,
                            kind: modes.DocumentHighlightKind.Read
                        });
                    }
                }
            }
            return winjs.TPromise.as(result);
        };
        HTMLWorker._stripQuotes = function (url) {
            return url
                .replace(/^'([^']+)'$/, function (substr, match1) { return match1; })
                .replace(/^"([^"]+)"$/, function (substr, match1) { return match1; });
        };
        HTMLWorker._getWorkspaceUrl = function (modelAbsoluteUri, rootAbsoluteUri, tokenContent) {
            tokenContent = HTMLWorker._stripQuotes(tokenContent);
            if (/^\s*javascript\:/i.test(tokenContent) || /^\s*\#/i.test(tokenContent)) {
                return null;
            }
            if (/^\s*https?:\/\//i.test(tokenContent) || /^\s*file:\/\//i.test(tokenContent)) {
                // Absolute link that needs no treatment
                return tokenContent.replace(/^\s*/g, '');
            }
            if (/^\s*\/\//i.test(tokenContent)) {
                // Absolute link (that does not name the protocol)
                var pickedScheme = network.Schemas.http;
                if (modelAbsoluteUri.scheme === network.Schemas.https) {
                    pickedScheme = network.Schemas.https;
                }
                return pickedScheme + ':' + tokenContent.replace(/^\s*/g, '');
            }
            var modelPath = paths.dirname(modelAbsoluteUri.path);
            var alternativeResultPath = null;
            if (tokenContent.length > 0 && tokenContent.charAt(0) === '/') {
                alternativeResultPath = tokenContent;
            }
            else {
                alternativeResultPath = paths.join(modelPath, tokenContent);
                alternativeResultPath = alternativeResultPath.replace(/^(\/\.\.)+/, '');
            }
            var potentialResult = modelAbsoluteUri.with({ path: alternativeResultPath }).toString();
            var rootAbsoluteUrlStr = (rootAbsoluteUri ? rootAbsoluteUri.toString() : null);
            if (rootAbsoluteUrlStr && strings.startsWith(modelAbsoluteUri.toString(), rootAbsoluteUrlStr)) {
                // The `rootAbsoluteUrl` is set and matches our current model
                // We need to ensure that this `potentialResult` does not escape `rootAbsoluteUrl`
                var commonPrefixLength = strings.commonPrefixLength(rootAbsoluteUrlStr, potentialResult);
                if (strings.endsWith(rootAbsoluteUrlStr, '/')) {
                    commonPrefixLength = potentialResult.lastIndexOf('/', commonPrefixLength) + 1;
                }
                return rootAbsoluteUrlStr + potentialResult.substr(commonPrefixLength);
            }
            return potentialResult;
        };
        HTMLWorker.prototype.createLink = function (modelAbsoluteUrl, rootAbsoluteUrl, tokenContent, lineNumber, startColumn, endColumn) {
            var workspaceUrl = HTMLWorker._getWorkspaceUrl(modelAbsoluteUrl, rootAbsoluteUrl, tokenContent);
            if (!workspaceUrl) {
                return null;
            }
            //		console.info('workspaceUrl: ' + workspaceUrl);
            return {
                range: {
                    startLineNumber: lineNumber,
                    startColumn: startColumn,
                    endLineNumber: lineNumber,
                    endColumn: endColumn
                },
                url: workspaceUrl
            };
        };
        HTMLWorker.prototype._computeHTMLLinks = function (model) {
            var lineCount = model.getLineCount(), newLinks = [], state = LinkDetectionState.LOOKING_FOR_HREF_OR_SRC, modelAbsoluteUrl = model.uri, lineNumber, lineContent, lineContentLength, tokens, tokenType, tokensLength, i, nextTokenEndIndex, tokenContent, link;
            var rootAbsoluteUrl = null;
            var workspace = this._contextService.getWorkspace();
            if (workspace) {
                // The workspace can be null in the no folder opened case
                var strRootAbsoluteUrl = String(workspace.resource);
                if (strRootAbsoluteUrl.charAt(strRootAbsoluteUrl.length - 1) === '/') {
                    rootAbsoluteUrl = uri_1.default.parse(strRootAbsoluteUrl);
                }
                else {
                    rootAbsoluteUrl = uri_1.default.parse(strRootAbsoluteUrl + '/');
                }
            }
            for (lineNumber = 1; lineNumber <= lineCount; lineNumber++) {
                lineContent = model.getLineContent(lineNumber);
                lineContentLength = lineContent.length;
                tokens = model.getLineTokens(lineNumber);
                for (i = 0, tokensLength = tokens.getTokenCount(); i < tokensLength; i++) {
                    tokenType = tokens.getTokenType(i);
                    switch (tokenType) {
                        case htmlTokenTypes_1.DELIM_ASSIGN:
                        case '':
                            break;
                        case htmlTokenTypes_1.ATTRIB_NAME:
                            nextTokenEndIndex = tokens.getTokenEndIndex(i, lineContentLength);
                            tokenContent = lineContent.substring(tokens.getTokenStartIndex(i), nextTokenEndIndex).toLowerCase();
                            if (tokenContent === 'src' || tokenContent === 'href') {
                                state = LinkDetectionState.AFTER_HREF_OR_SRC;
                            }
                            else {
                                state = LinkDetectionState.LOOKING_FOR_HREF_OR_SRC;
                            }
                            break;
                        case htmlTokenTypes_1.ATTRIB_VALUE:
                            if (state === LinkDetectionState.AFTER_HREF_OR_SRC) {
                                nextTokenEndIndex = tokens.getTokenEndIndex(i, lineContentLength);
                                tokenContent = lineContent.substring(tokens.getTokenStartIndex(i), nextTokenEndIndex);
                                link = this.createLink(modelAbsoluteUrl, rootAbsoluteUrl, tokenContent, lineNumber, tokens.getTokenStartIndex(i) + 2, nextTokenEndIndex);
                                if (link) {
                                    newLinks.push(link);
                                }
                                state = LinkDetectionState.LOOKING_FOR_HREF_OR_SRC;
                            }
                        default:
                            if (htmlTokenTypes_1.isTag(tokenType)) {
                                state = LinkDetectionState.LOOKING_FOR_HREF_OR_SRC;
                            }
                            else if (state === LinkDetectionState.AFTER_HREF_OR_SRC) {
                                state = LinkDetectionState.LOOKING_FOR_HREF_OR_SRC;
                            }
                    }
                }
            }
            return newLinks;
        };
        HTMLWorker.prototype.provideLinks = function (resource) {
            var model = this.resourceService.get(resource);
            return winjs.TPromise.as(this._computeHTMLLinks(model));
        };
        HTMLWorker = __decorate([
            __param(1, resourceService_1.IResourceService),
            __param(2, workspace_1.IWorkspaceContextService)
        ], HTMLWorker);
        return HTMLWorker;
    }());
    exports.HTMLWorker = HTMLWorker;
    function isWhiteSpace(s) {
        return /^\s*$/.test(s);
    }
});

}).call(this);
//# sourceMappingURL=htmlWorker.js.map
