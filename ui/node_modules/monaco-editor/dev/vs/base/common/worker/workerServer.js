/*!-----------------------------------------------------------
 * Copyright (c) Microsoft Corporation. All rights reserved.
 * Version: 0.5.3(793ede49d53dba79d39e52205f16321278f5183c)
 * Released under the MIT license
 * https://github.com/Microsoft/vscode/blob/master/LICENSE.txt
 *-----------------------------------------------------------*/

(function() {
var __m = ["exports","require","vs/base/common/errors","vs/base/common/strings","vs/platform/instantiation/common/instantiation","vs/base/common/winjs.base","vs/nls!vs/base/common/worker/workerServer","vs/nls","vs/base/common/event","vs/base/common/types","vs/base/common/platform","vs/platform/platform","vs/base/common/objects","vs/base/common/paths","vs/editor/common/modes/supports","vs/base/common/eventEmitter","vs/base/common/lifecycle","vs/editor/common/model/wordHelper","vs/editor/common/core/modeTransition","vs/editor/common/editorCommon","vs/base/common/uri","vs/editor/common/modes","vs/editor/common/core/position","vs/editor/common/services/modeService","vs/base/common/async","vs/editor/common/modes/monarch/monarchCommon","vs/base/common/timer","vs/platform/workspace/common/workspace","vs/editor/common/modes/supports/richEditBrackets","vs/editor/common/core/range","vs/editor/common/core/arrays","vs/editor/common/modes/languageConfigurationRegistry","vs/base/common/severity","vs/platform/instantiation/common/descriptors","vs/base/common/assert","vs/editor/common/services/compatWorkerService","vs/base/common/map","vs/editor/common/services/resourceService","vs/platform/extensions/common/extensions","vs/editor/common/modes/abstractState","vs/editor/common/modes/modesRegistry","vs/platform/jsonschemas/common/jsonContributionRegistry","vs/platform/extensions/common/extensionsRegistry","vs/platform/request/common/request","vs/platform/telemetry/common/telemetry","vs/editor/common/model/textModel","vs/base/common/mime","vs/base/common/cancellation","vs/editor/common/modes/nullMode","vs/base/common/collections","vs/editor/common/core/viewLineToken","vs/platform/configuration/common/configuration","vs/platform/event/common/event","vs/editor/common/modes/supports/tokenizationSupport","vs/platform/instantiation/common/serviceCollection","vs/base/common/glob","vs/editor/common/model/tokensBinaryEncoding","vs/editor/common/model/modelLine","vs/base/common/arrays","vs/editor/common/modes/supports/suggestSupport","vs/editor/common/modes/abstractMode","vs/base/common/stopwatch","vs/editor/common/modes/lineStream","vs/base/common/callbackList","vs/editor/common/model/textModelWithTokensHelpers","vs/nls!vs/base/common/severity","vs/nls!vs/base/common/errors","vs/nls!vs/editor/common/config/defaultConfig","vs/editor/common/config/defaultConfig","vs/editor/common/model/tokenIterator","vs/nls!vs/editor/common/model/textModelWithTokens","vs/editor/common/model/textModelWithTokens","vs/editor/common/model/mirrorModel","vs/nls!vs/editor/common/modes/modesRegistry","vs/nls!vs/editor/common/modes/supports/suggestSupport","vs/nls!vs/editor/common/services/modeServiceImpl","vs/nls!vs/platform/configuration/common/configurationRegistry","vs/nls!vs/platform/extensions/common/abstractExtensionService","vs/nls!vs/platform/extensions/common/extensionsRegistry","vs/nls!vs/platform/jsonschemas/common/jsonContributionRegistry","vs/base/common/filters","vs/base/common/events","vs/editor/common/modes/languageSelector","vs/editor/common/services/editorWorkerService","vs/base/common/network","vs/editor/common/services/modelService","vs/base/common/graph","vs/editor/common/services/resourceServiceImpl","vs/editor/common/modes/monarch/monarchCompile","vs/base/common/marshalling","vs/platform/event/common/eventService","vs/base/common/worker/workerProtocol","vs/platform/workspace/common/baseWorkspaceContextService","vs/platform/instantiation/common/instantiationService","vs/editor/common/modes/supports/characterPair","vs/editor/common/model/indentationGuesser","vs/editor/common/services/compatWorkerServiceWorker","vs/editor/common/services/languagesRegistry","vs/editor/common/languages.common","vs/editor/common/modes/supports/electricCharacter","vs/platform/configuration/common/configurationRegistry","vs/editor/common/model/lineToken","vs/editor/common/modes/languageFeatureRegistry","vs/editor/common/services/modeServiceImpl","vs/platform/extensions/common/abstractExtensionService","vs/editor/common/modes/monarch/monarchLexer","vs/editor/common/modes/supports/onEnter","vs/platform/request/common/baseRequestService","vs/editor/common/viewModel/prefixSumComputer","vs/base/common/worker/workerServer","vs/base/common/winjs.base.raw","vs/editor/common/worker/editorWorkerServer"];
var __M = function(deps) {
  var result = [];
  for (var i = 0, len = deps.length; i < len; i++) {
    result[i] = __m[deps[i]];
  }
  return result;
};
define(__m[58], __M([1,0]), function (require, exports) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    /**
     * Returns the last element of an array.
     * @param array The array.
     * @param n Which element from the end (default ist zero).
     */
    function tail(array, n) {
        if (n === void 0) { n = 0; }
        return array[array.length - (1 + n)];
    }
    exports.tail = tail;
    /**
     * Iterates the provided array and allows to remove
     * elements while iterating.
     */
    function forEach(array, callback) {
        for (var i = 0, len = array.length; i < len; i++) {
            callback(array[i], function () {
                array.splice(i, 1);
                i--;
                len--;
            });
        }
    }
    exports.forEach = forEach;
    function equals(one, other, itemEquals) {
        if (itemEquals === void 0) { itemEquals = function (a, b) { return a === b; }; }
        if (one.length !== other.length) {
            return false;
        }
        for (var i = 0, len = one.length; i < len; i++) {
            if (!itemEquals(one[i], other[i])) {
                return false;
            }
        }
        return true;
    }
    exports.equals = equals;
    function binarySearch(array, key, comparator) {
        var low = 0, high = array.length - 1;
        while (low <= high) {
            var mid = ((low + high) / 2) | 0;
            var comp = comparator(array[mid], key);
            if (comp < 0) {
                low = mid + 1;
            }
            else if (comp > 0) {
                high = mid - 1;
            }
            else {
                return mid;
            }
        }
        return -(low + 1);
    }
    exports.binarySearch = binarySearch;
    /**
     * Takes a sorted array and a function p. The array is sorted in such a way that all elements where p(x) is false
     * are located before all elements where p(x) is true.
     * @returns the least x for which p(x) is true or array.length if no element fullfills the given function.
     */
    function findFirst(array, p) {
        var low = 0, high = array.length;
        if (high === 0) {
            return 0; // no children
        }
        while (low < high) {
            var mid = Math.floor((low + high) / 2);
            if (p(array[mid])) {
                high = mid;
            }
            else {
                low = mid + 1;
            }
        }
        return low;
    }
    exports.findFirst = findFirst;
    function merge(arrays, hashFn) {
        var result = new Array();
        if (!hashFn) {
            for (var i = 0, len = arrays.length; i < len; i++) {
                result.push.apply(result, arrays[i]);
            }
        }
        else {
            var map = {};
            for (var i = 0; i < arrays.length; i++) {
                for (var j = 0; j < arrays[i].length; j++) {
                    var element = arrays[i][j], hash = hashFn(element);
                    if (!map.hasOwnProperty(hash)) {
                        map[hash] = true;
                        result.push(element);
                    }
                }
            }
        }
        return result;
    }
    exports.merge = merge;
    /**
     * @returns a new array with all undefined or null values removed. The original array is not modified at all.
     */
    function coalesce(array) {
        if (!array) {
            return array;
        }
        return array.filter(function (e) { return !!e; });
    }
    exports.coalesce = coalesce;
    /**
     * @returns true if the given item is contained in the array.
     */
    function contains(array, item) {
        return array.indexOf(item) >= 0;
    }
    exports.contains = contains;
    /**
     * Swaps the elements in the array for the provided positions.
     */
    function swap(array, pos1, pos2) {
        var element1 = array[pos1];
        var element2 = array[pos2];
        array[pos1] = element2;
        array[pos2] = element1;
    }
    exports.swap = swap;
    /**
     * Moves the element in the array for the provided positions.
     */
    function move(array, from, to) {
        array.splice(to, 0, array.splice(from, 1)[0]);
    }
    exports.move = move;
    /**
     * @returns {{false}} if the provided object is an array
     * 	and not empty.
     */
    function isFalsyOrEmpty(obj) {
        return !Array.isArray(obj) || obj.length === 0;
    }
    exports.isFalsyOrEmpty = isFalsyOrEmpty;
    /**
     * Removes duplicates from the given array. The optional keyFn allows to specify
     * how elements are checked for equalness by returning a unique string for each.
     */
    function distinct(array, keyFn) {
        if (!keyFn) {
            return array.filter(function (element, position) {
                return array.indexOf(element) === position;
            });
        }
        var seen = Object.create(null);
        return array.filter(function (elem) {
            var key = keyFn(elem);
            if (seen[key]) {
                return false;
            }
            seen[key] = true;
            return true;
        });
    }
    exports.distinct = distinct;
    function uniqueFilter(keyFn) {
        var seen = Object.create(null);
        return function (element) {
            var key = keyFn(element);
            if (seen[key]) {
                return false;
            }
            seen[key] = true;
            return true;
        };
    }
    exports.uniqueFilter = uniqueFilter;
    function firstIndex(array, fn) {
        for (var i = 0; i < array.length; i++) {
            var element = array[i];
            if (fn(element)) {
                return i;
            }
        }
        return -1;
    }
    exports.firstIndex = firstIndex;
    function first(array, fn, notFoundValue) {
        if (notFoundValue === void 0) { notFoundValue = null; }
        var index = firstIndex(array, fn);
        return index < 0 ? notFoundValue : array[index];
    }
    exports.first = first;
    function commonPrefixLength(one, other, equals) {
        if (equals === void 0) { equals = function (a, b) { return a === b; }; }
        var result = 0;
        for (var i = 0, len = Math.min(one.length, other.length); i < len && equals(one[i], other[i]); i++) {
            result++;
        }
        return result;
    }
    exports.commonPrefixLength = commonPrefixLength;
    function flatten(arr) {
        return arr.reduce(function (r, v) { return r.concat(v); }, []);
    }
    exports.flatten = flatten;
    function range(to, from) {
        if (from === void 0) { from = 0; }
        var result = [];
        for (var i = from; i < to; i++) {
            result.push(i);
        }
        return result;
    }
    exports.range = range;
    function fill(num, valueFn, arr) {
        if (arr === void 0) { arr = []; }
        for (var i = 0; i < num; i++) {
            arr[i] = valueFn();
        }
        return arr;
    }
    exports.fill = fill;
    function index(array, indexer) {
        var result = Object.create(null);
        array.forEach(function (t) { return result[indexer(t)] = t; });
        return result;
    }
    exports.index = index;
});

define(__m[34], __M([1,0]), function (require, exports) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    /**
     * Throws an error with the provided message if the provided value does not evaluate to a true Javascript value.
     */
    function ok(value, message) {
        if (!value || value === null) {
            throw new Error(message ? 'Assertion failed (' + message + ')' : 'Assertion Failed');
        }
    }
    exports.ok = ok;
});

define(__m[49], __M([1,0]), function (require, exports) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    function createStringDictionary() {
        return Object.create(null);
    }
    exports.createStringDictionary = createStringDictionary;
    function createNumberDictionary() {
        return Object.create(null);
    }
    exports.createNumberDictionary = createNumberDictionary;
    function lookup(from, what, alternate) {
        if (alternate === void 0) { alternate = null; }
        var key = String(what);
        if (contains(from, key)) {
            return from[key];
        }
        return alternate;
    }
    exports.lookup = lookup;
    function lookupOrInsert(from, stringOrNumber, alternate) {
        var key = String(stringOrNumber);
        if (contains(from, key)) {
            return from[key];
        }
        else {
            if (typeof alternate === 'function') {
                alternate = alternate();
            }
            from[key] = alternate;
            return alternate;
        }
    }
    exports.lookupOrInsert = lookupOrInsert;
    function insert(into, data, hashFn) {
        into[hashFn(data)] = data;
    }
    exports.insert = insert;
    var hasOwnProperty = Object.prototype.hasOwnProperty;
    function contains(from, what) {
        return hasOwnProperty.call(from, what);
    }
    exports.contains = contains;
    function values(from) {
        var result = [];
        for (var key in from) {
            if (hasOwnProperty.call(from, key)) {
                result.push(from[key]);
            }
        }
        return result;
    }
    exports.values = values;
    function forEach(from, callback) {
        for (var key in from) {
            if (hasOwnProperty.call(from, key)) {
                var result = callback({ key: key, value: from[key] }, function () {
                    delete from[key];
                });
                if (result === false) {
                    return;
                }
            }
        }
    }
    exports.forEach = forEach;
    function remove(from, key) {
        if (!hasOwnProperty.call(from, key)) {
            return false;
        }
        delete from[key];
        return true;
    }
    exports.remove = remove;
    /**
     * Groups the collection into a dictionary based on the provided
     * group function.
     */
    function groupBy(data, groupFn) {
        var result = createStringDictionary();
        data.forEach(function (element) { return lookupOrInsert(result, groupFn(element), []).push(element); });
        return result;
    }
    exports.groupBy = groupBy;
});

var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
define(__m[81], __M([1,0]), function (require, exports) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var Event = (function () {
        function Event(originalEvent) {
            this.time = (new Date()).getTime();
            this.originalEvent = originalEvent;
            this.source = null;
        }
        return Event;
    }());
    exports.Event = Event;
    var PropertyChangeEvent = (function (_super) {
        __extends(PropertyChangeEvent, _super);
        function PropertyChangeEvent(key, oldValue, newValue, originalEvent) {
            _super.call(this, originalEvent);
            this.key = key;
            this.oldValue = oldValue;
            this.newValue = newValue;
        }
        return PropertyChangeEvent;
    }(Event));
    exports.PropertyChangeEvent = PropertyChangeEvent;
    var ViewerEvent = (function (_super) {
        __extends(ViewerEvent, _super);
        function ViewerEvent(element, originalEvent) {
            _super.call(this, originalEvent);
            this.element = element;
        }
        return ViewerEvent;
    }(Event));
    exports.ViewerEvent = ViewerEvent;
    exports.EventType = {
        PROPERTY_CHANGED: 'propertyChanged',
        SELECTION: 'selection',
        FOCUS: 'focus',
        BLUR: 'blur',
        HIGHLIGHT: 'highlight',
        EXPAND: 'expand',
        COLLAPSE: 'collapse',
        TOGGLE: 'toggle',
        CONTENTS_CHANGED: 'contentsChanged',
        BEFORE_RUN: 'beforeRun',
        RUN: 'run',
        EDIT: 'edit',
        SAVE: 'save',
        CANCEL: 'cancel',
        CHANGE: 'change',
        DISPOSE: 'dispose',
    };
});

/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Microsoft Corporation. All rights reserved.
 *  Licensed under the MIT License. See License.txt in the project root for license information.
 *--------------------------------------------------------------------------------------------*/





define(__m[36], __M([1,0]), function (require, exports) {
    'use strict';
    /**
     * A simple map to store value by a key object. Key can be any object that has toString() function to get
     * string value of the key.
     */
    var SimpleMap = (function () {
        function SimpleMap() {
            this.map = Object.create(null);
            this._size = 0;
        }
        Object.defineProperty(SimpleMap.prototype, "size", {
            get: function () {
                return this._size;
            },
            enumerable: true,
            configurable: true
        });
        SimpleMap.prototype.get = function (k) {
            var value = this.peek(k);
            return value ? value : null;
        };
        SimpleMap.prototype.keys = function () {
            var keys = [];
            for (var key in this.map) {
                keys.push(this.map[key].key);
            }
            return keys;
        };
        SimpleMap.prototype.entries = function () {
            var entries = [];
            for (var key in this.map) {
                entries.push(this.map[key]);
            }
            return entries;
        };
        SimpleMap.prototype.set = function (k, t) {
            if (this.get(k)) {
                return false; // already present!
            }
            this.push(k, t);
            return true;
        };
        SimpleMap.prototype.delete = function (k) {
            var value = this.get(k);
            if (value) {
                this.pop(k);
                return value;
            }
            return null;
        };
        SimpleMap.prototype.has = function (k) {
            return !!this.get(k);
        };
        SimpleMap.prototype.clear = function () {
            this.map = Object.create(null);
            this._size = 0;
        };
        SimpleMap.prototype.push = function (key, value) {
            var entry = { key: key, value: value };
            this.map[key.toString()] = entry;
            this._size++;
        };
        SimpleMap.prototype.pop = function (k) {
            delete this.map[k.toString()];
            this._size--;
        };
        SimpleMap.prototype.peek = function (k) {
            var entry = this.map[k.toString()];
            return entry ? entry.value : null;
        };
        return SimpleMap;
    }());
    exports.SimpleMap = SimpleMap;
    /**
     * A simple Map<T> that optionally allows to set a limit of entries to store. Once the limit is hit,
     * the cache will remove the entry that was last recently added. Or, if a ratio is provided below 1,
     * all elements will be removed until the ratio is full filled (e.g. 0.75 to remove 25% of old elements).
     */
    var LinkedMap = (function () {
        function LinkedMap(limit, ratio) {
            if (limit === void 0) { limit = Number.MAX_VALUE; }
            if (ratio === void 0) { ratio = 1; }
            this.limit = limit;
            this.map = Object.create(null);
            this._size = 0;
            this.ratio = limit * ratio;
        }
        Object.defineProperty(LinkedMap.prototype, "size", {
            get: function () {
                return this._size;
            },
            enumerable: true,
            configurable: true
        });
        LinkedMap.prototype.set = function (key, value) {
            if (this.map[key]) {
                return false; // already present!
            }
            var entry = { key: key, value: value };
            this.push(entry);
            if (this._size > this.limit) {
                this.trim();
            }
            return true;
        };
        LinkedMap.prototype.get = function (key) {
            var entry = this.map[key];
            return entry ? entry.value : null;
        };
        LinkedMap.prototype.delete = function (key) {
            var entry = this.map[key];
            if (entry) {
                this.map[key] = void 0;
                this._size--;
                if (entry.next) {
                    entry.next.prev = entry.prev; // [A]<-[x]<-[C] = [A]<-[C]
                }
                else {
                    this.head = entry.prev; // [A]-[x] = [A]
                }
                if (entry.prev) {
                    entry.prev.next = entry.next; // [A]->[x]->[C] = [A]->[C]
                }
                else {
                    this.tail = entry.next; // [x]-[A] = [A]
                }
                return entry.value;
            }
            return null;
        };
        LinkedMap.prototype.has = function (key) {
            return !!this.map[key];
        };
        LinkedMap.prototype.clear = function () {
            this.map = Object.create(null);
            this._size = 0;
            this.head = null;
            this.tail = null;
        };
        LinkedMap.prototype.push = function (entry) {
            if (this.head) {
                // [A]-[B] = [A]-[B]->[X]
                entry.prev = this.head;
                this.head.next = entry;
            }
            if (!this.tail) {
                this.tail = entry;
            }
            this.head = entry;
            this.map[entry.key] = entry;
            this._size++;
        };
        LinkedMap.prototype.trim = function () {
            if (this.tail) {
                // Remove all elements until ratio is reached
                if (this.ratio < this.limit) {
                    var index = 0;
                    var current = this.tail;
                    while (current.next) {
                        // Remove the entry
                        this.map[current.key] = void 0;
                        this._size--;
                        // if we reached the element that overflows our ratio condition
                        // make its next element the new tail of the Map and adjust the size
                        if (index === this.ratio) {
                            this.tail = current.next;
                            this.tail.prev = null;
                            break;
                        }
                        // Move on
                        current = current.next;
                        index++;
                    }
                }
                else {
                    this.map[this.tail.key] = void 0;
                    this._size--;
                    // [x]-[B] = [B]
                    this.tail = this.tail.next;
                    this.tail.prev = null;
                }
            }
        };
        return LinkedMap;
    }());
    exports.LinkedMap = LinkedMap;
    /**
     * A subclass of Map<T> that makes an entry the MRU entry as soon
     * as it is being accessed. In combination with the limit for the
     * maximum number of elements in the cache, it helps to remove those
     * entries from the cache that are LRU.
     */
    var LRUCache = (function (_super) {
        __extends(LRUCache, _super);
        function LRUCache(limit) {
            _super.call(this, limit);
        }
        LRUCache.prototype.get = function (key) {
            // Upon access of an entry, make it the head of
            // the linked map so that it is the MRU element
            var entry = this.map[key];
            if (entry) {
                this.delete(key);
                this.push(entry);
                return entry.value;
            }
            return null;
        };
        return LRUCache;
    }(LinkedMap));
    exports.LRUCache = LRUCache;
});

define(__m[10], __M([1,0]), function (require, exports) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    // --- THIS FILE IS TEMPORARY UNTIL ENV.TS IS CLEANED UP. IT CAN SAFELY BE USED IN ALL TARGET EXECUTION ENVIRONMENTS (node & dom) ---
    var _isWindows = false;
    var _isMacintosh = false;
    var _isLinux = false;
    var _isRootUser = false;
    var _isNative = false;
    var _isWeb = false;
    var _isQunit = false;
    var _locale = undefined;
    var _language = undefined;
    exports.LANGUAGE_DEFAULT = 'en';
    // OS detection
    if (typeof process === 'object') {
        _isWindows = (process.platform === 'win32');
        _isMacintosh = (process.platform === 'darwin');
        _isLinux = (process.platform === 'linux');
        _isRootUser = !_isWindows && (process.getuid() === 0);
        var vscode_nls_config = process.env['VSCODE_NLS_CONFIG'];
        if (vscode_nls_config) {
            try {
                var nlsConfig = JSON.parse(vscode_nls_config);
                var resolved = nlsConfig.availableLanguages['*'];
                _locale = nlsConfig.locale;
                // VSCode's default language is 'en'
                _language = resolved ? resolved : exports.LANGUAGE_DEFAULT;
            }
            catch (e) {
            }
        }
        _isNative = true;
    }
    else if (typeof navigator === 'object') {
        var userAgent = navigator.userAgent;
        _isWindows = userAgent.indexOf('Windows') >= 0;
        _isMacintosh = userAgent.indexOf('Macintosh') >= 0;
        _isLinux = userAgent.indexOf('Linux') >= 0;
        _isWeb = true;
        _locale = navigator.language;
        _language = _locale;
        _isQunit = !!self.QUnit;
    }
    (function (Platform) {
        Platform[Platform["Web"] = 0] = "Web";
        Platform[Platform["Mac"] = 1] = "Mac";
        Platform[Platform["Linux"] = 2] = "Linux";
        Platform[Platform["Windows"] = 3] = "Windows";
    })(exports.Platform || (exports.Platform = {}));
    var Platform = exports.Platform;
    exports._platform = Platform.Web;
    if (_isNative) {
        if (_isMacintosh) {
            exports._platform = Platform.Mac;
        }
        else if (_isWindows) {
            exports._platform = Platform.Windows;
        }
        else if (_isLinux) {
            exports._platform = Platform.Linux;
        }
    }
    exports.isWindows = _isWindows;
    exports.isMacintosh = _isMacintosh;
    exports.isLinux = _isLinux;
    exports.isRootUser = _isRootUser;
    exports.isNative = _isNative;
    exports.isWeb = _isWeb;
    exports.isQunit = _isQunit;
    exports.platform = exports._platform;
    /**
     * The language used for the user interface. The format of
     * the string is all lower case (e.g. zh-tw for Traditional
     * Chinese)
     */
    exports.language = _language;
    /**
     * The OS locale or the locale specified by --locale. The format of
     * the string is all lower case (e.g. zh-tw for Traditional
     * Chinese). The UI is not necessarily shown in the provided locale.
     */
    exports.locale = _locale;
    var _globals = (typeof self === 'object' ? self : global);
    exports.globals = _globals;
    function hasWebWorkerSupport() {
        return typeof _globals.Worker !== 'undefined';
    }
    exports.hasWebWorkerSupport = hasWebWorkerSupport;
    exports.setTimeout = _globals.setTimeout.bind(_globals);
    exports.clearTimeout = _globals.clearTimeout.bind(_globals);
    exports.setInterval = _globals.setInterval.bind(_globals);
    exports.clearInterval = _globals.clearInterval.bind(_globals);
});

define(__m[13], __M([1,0,10]), function (require, exports, platform_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    /**
     * The forward slash path separator.
     */
    exports.sep = '/';
    /**
     * The native path separator depending on the OS.
     */
    exports.nativeSep = platform_1.isWindows ? '\\' : '/';
    function relative(from, to) {
        from = normalize(from);
        to = normalize(to);
        var fromParts = from.split(exports.sep), toParts = to.split(exports.sep);
        while (fromParts.length > 0 && toParts.length > 0) {
            if (fromParts[0] === toParts[0]) {
                fromParts.shift();
                toParts.shift();
            }
            else {
                break;
            }
        }
        for (var i = 0, len = fromParts.length; i < len; i++) {
            toParts.unshift('..');
        }
        return toParts.join(exports.sep);
    }
    exports.relative = relative;
    /**
     * @returns the directory name of a path.
     */
    function dirname(path) {
        var idx = ~path.lastIndexOf('/') || ~path.lastIndexOf('\\');
        if (idx === 0) {
            return '.';
        }
        else if (~idx === 0) {
            return path[0];
        }
        else {
            return path.substring(0, ~idx);
        }
    }
    exports.dirname = dirname;
    /**
     * @returns the base name of a path.
     */
    function basename(path) {
        var idx = ~path.lastIndexOf('/') || ~path.lastIndexOf('\\');
        if (idx === 0) {
            return path;
        }
        else if (~idx === path.length - 1) {
            return basename(path.substring(0, path.length - 1));
        }
        else {
            return path.substr(~idx + 1);
        }
    }
    exports.basename = basename;
    /**
     * @returns {{.far}} from boo.far or the empty string.
     */
    function extname(path) {
        path = basename(path);
        var idx = ~path.lastIndexOf('.');
        return idx ? path.substring(~idx) : '';
    }
    exports.extname = extname;
    var _posixBadPath = /(\/\.\.?\/)|(\/\.\.?)$|^(\.\.?\/)|(\/\/+)|(\\)/;
    var _winBadPath = /(\\\.\.?\\)|(\\\.\.?)$|^(\.\.?\\)|(\\\\+)|(\/)/;
    function _isNormal(path, win) {
        return win
            ? !_winBadPath.test(path)
            : !_posixBadPath.test(path);
    }
    function normalize(path, toOSPath) {
        if (path === null || path === void 0) {
            return path;
        }
        var len = path.length;
        if (len === 0) {
            return '.';
        }
        var wantsBackslash = platform_1.isWindows && toOSPath;
        if (_isNormal(path, wantsBackslash)) {
            return path;
        }
        var sep = wantsBackslash ? '\\' : '/';
        var root = getRoot(path, sep);
        // skip the root-portion of the path
        var start = root.length;
        var skip = false;
        var res = '';
        for (var end = root.length; end <= len; end++) {
            // either at the end or at a path-separator character
            if (end === len || path.charCodeAt(end) === _slash || path.charCodeAt(end) === _backslash) {
                if (streql(path, start, end, '..')) {
                    // skip current and remove parent (if there is already something)
                    var prev_start = res.lastIndexOf(sep);
                    var prev_part = res.slice(prev_start + 1);
                    if ((root || prev_part.length > 0) && prev_part !== '..') {
                        res = prev_start === -1 ? '' : res.slice(0, prev_start);
                        skip = true;
                    }
                }
                else if (streql(path, start, end, '.') && (root || res || end < len - 1)) {
                    // skip current (if there is already something or if there is more to come)
                    skip = true;
                }
                if (!skip) {
                    var part = path.slice(start, end);
                    if (res !== '' && res[res.length - 1] !== sep) {
                        res += sep;
                    }
                    res += part;
                }
                start = end + 1;
                skip = false;
            }
        }
        return root + res;
    }
    exports.normalize = normalize;
    function streql(value, start, end, other) {
        return start + other.length === end && value.indexOf(other, start) === start;
    }
    /**
     * Computes the _root_ this path, like `getRoot('c:\files') === c:\`,
     * `getRoot('files:///files/path') === files:///`,
     * or `getRoot('\\server\shares\path') === \\server\shares\`
     */
    function getRoot(path, sep) {
        if (sep === void 0) { sep = '/'; }
        if (!path) {
            return '';
        }
        var len = path.length;
        var code = path.charCodeAt(0);
        if (code === _slash || code === _backslash) {
            code = path.charCodeAt(1);
            if (code === _slash || code === _backslash) {
                // UNC candidate \\localhost\shares\ddd
                //               ^^^^^^^^^^^^^^^^^^^
                code = path.charCodeAt(2);
                if (code !== _slash && code !== _backslash) {
                    var pos_1 = 3;
                    var start = pos_1;
                    for (; pos_1 < len; pos_1++) {
                        code = path.charCodeAt(pos_1);
                        if (code === _slash || code === _backslash) {
                            break;
                        }
                    }
                    code = path.charCodeAt(pos_1 + 1);
                    if (start !== pos_1 && code !== _slash && code !== _backslash) {
                        pos_1 += 1;
                        for (; pos_1 < len; pos_1++) {
                            code = path.charCodeAt(pos_1);
                            if (code === _slash || code === _backslash) {
                                return path.slice(0, pos_1 + 1) // consume this separator
                                    .replace(/[\\/]/g, sep);
                            }
                        }
                    }
                }
            }
            // /user/far
            // ^
            return sep;
        }
        else if ((code >= _A && code <= _Z) || (code >= _a && code <= _z)) {
            // check for windows drive letter c:\ or c:
            if (path.charCodeAt(1) === _colon) {
                code = path.charCodeAt(2);
                if (code === _slash || code === _backslash) {
                    // C:\fff
                    // ^^^
                    return path.slice(0, 2) + sep;
                }
                else {
                    // C:
                    // ^^
                    return path.slice(0, 2);
                }
            }
        }
        // check for URI
        // scheme://authority/path
        // ^^^^^^^^^^^^^^^^^^^
        var pos = path.indexOf('://');
        if (pos !== -1) {
            pos += 3; // 3 -> "://".length
            for (; pos < len; pos++) {
                code = path.charCodeAt(pos);
                if (code === _slash || code === _backslash) {
                    return path.slice(0, pos + 1); // consume this separator
                }
            }
        }
        return '';
    }
    exports.getRoot = getRoot;
    exports.join = function () {
        var value = '';
        for (var i = 0; i < arguments.length; i++) {
            var part = arguments[i];
            if (i > 0) {
                // add the separater between two parts unless
                // there already is one
                var last = value.charCodeAt(value.length - 1);
                if (last !== _slash && last !== _backslash) {
                    var next = part.charCodeAt(0);
                    if (next !== _slash && next !== _backslash) {
                        value += exports.sep;
                    }
                }
            }
            value += part;
        }
        return normalize(value);
    };
    /**
     * Check if the path follows this pattern: `\\hostname\sharename`.
     *
     * @see https://msdn.microsoft.com/en-us/library/gg465305.aspx
     * @return A boolean indication if the path is a UNC path, on none-windows
     * always false.
     */
    function isUNC(path) {
        if (!platform_1.isWindows) {
            // UNC is a windows concept
            return false;
        }
        if (!path || path.length < 5) {
            // at least \\a\b
            return false;
        }
        var code = path.charCodeAt(0);
        if (code !== _backslash) {
            return false;
        }
        code = path.charCodeAt(1);
        if (code !== _backslash) {
            return false;
        }
        var pos = 2;
        var start = pos;
        for (; pos < path.length; pos++) {
            code = path.charCodeAt(pos);
            if (code === _backslash) {
                break;
            }
        }
        if (start === pos) {
            return false;
        }
        code = path.charCodeAt(pos + 1);
        if (isNaN(code) || code === _backslash) {
            return false;
        }
        return true;
    }
    exports.isUNC = isUNC;
    function isPosixAbsolute(path) {
        return path && path[0] === '/';
    }
    function makePosixAbsolute(path) {
        return isPosixAbsolute(normalize(path)) ? path : exports.sep + path;
    }
    exports.makePosixAbsolute = makePosixAbsolute;
    var _slash = '/'.charCodeAt(0);
    var _backslash = '\\'.charCodeAt(0);
    var _colon = ':'.charCodeAt(0);
    var _a = 'a'.charCodeAt(0);
    var _A = 'A'.charCodeAt(0);
    var _z = 'z'.charCodeAt(0);
    var _Z = 'Z'.charCodeAt(0);
    function isEqualOrParent(path, candidate) {
        if (path === candidate) {
            return true;
        }
        path = normalize(path);
        candidate = normalize(candidate);
        var candidateLen = candidate.length;
        var lastCandidateChar = candidate.charCodeAt(candidateLen - 1);
        if (lastCandidateChar === _slash) {
            candidate = candidate.substring(0, candidateLen - 1);
            candidateLen -= 1;
        }
        if (path === candidate) {
            return true;
        }
        if (!platform_1.isLinux) {
            // case insensitive
            path = path.toLowerCase();
            candidate = candidate.toLowerCase();
        }
        if (path === candidate) {
            return true;
        }
        if (path.indexOf(candidate) !== 0) {
            return false;
        }
        var char = path.charCodeAt(candidateLen);
        return char === _slash;
    }
    exports.isEqualOrParent = isEqualOrParent;
    // Reference: https://en.wikipedia.org/wiki/Filename
    var INVALID_FILE_CHARS = platform_1.isWindows ? /[\\/:\*\?"<>\|]/g : /[\\/]/g;
    var WINDOWS_FORBIDDEN_NAMES = /^(con|prn|aux|clock\$|nul|lpt[0-9]|com[0-9])$/i;
    function isValidBasename(name) {
        if (!name || name.length === 0 || /^\s+$/.test(name)) {
            return false; // require a name that is not just whitespace
        }
        INVALID_FILE_CHARS.lastIndex = 0; // the holy grail of software development
        if (INVALID_FILE_CHARS.test(name)) {
            return false; // check for certain invalid file characters
        }
        if (platform_1.isWindows && WINDOWS_FORBIDDEN_NAMES.test(name)) {
            return false; // check for certain invalid file names
        }
        if (name === '.' || name === '..') {
            return false; // check for reserved values
        }
        if (platform_1.isWindows && name[name.length - 1] === '.') {
            return false; // Windows: file cannot end with a "."
        }
        if (platform_1.isWindows && name.length !== name.trim().length) {
            return false; // Windows: file cannot end with a whitespace
        }
        return true;
    }
    exports.isValidBasename = isValidBasename;
    exports.isAbsoluteRegex = /^((\/|[a-zA-Z]:\\)[^\(\)<>\\'\"\[\]]+)/;
    /**
     * If you have access to node, it is recommended to use node's path.isAbsolute().
     * This is a simple regex based approach.
     */
    function isAbsolute(path) {
        return exports.isAbsoluteRegex.test(path);
    }
    exports.isAbsolute = isAbsolute;
});

define(__m[61], __M([1,0,10]), function (require, exports, platform_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var hasPerformanceNow = (platform_1.globals.performance && typeof platform_1.globals.performance.now === 'function');
    var StopWatch = (function () {
        function StopWatch(highResolution) {
            this._highResolution = hasPerformanceNow && highResolution;
            this._startTime = this._now();
            this._stopTime = -1;
        }
        StopWatch.create = function (highResolution) {
            if (highResolution === void 0) { highResolution = true; }
            return new StopWatch(highResolution);
        };
        StopWatch.prototype.stop = function () {
            this._stopTime = this._now();
        };
        StopWatch.prototype.elapsed = function () {
            if (this._stopTime !== -1) {
                return this._stopTime - this._startTime;
            }
            return this._now() - this._startTime;
        };
        StopWatch.prototype._now = function () {
            return this._highResolution ? platform_1.globals.performance.now() : new Date().getTime();
        };
        return StopWatch;
    }());
    exports.StopWatch = StopWatch;
});

define(__m[3], __M([1,0,36]), function (require, exports, map_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    /**
     * The empty string.
     */
    exports.empty = '';
    /**
     * @returns the provided number with the given number of preceding zeros.
     */
    function pad(n, l, char) {
        if (char === void 0) { char = '0'; }
        var str = '' + n;
        var r = [str];
        for (var i = str.length; i < l; i++) {
            r.push(char);
        }
        return r.reverse().join('');
    }
    exports.pad = pad;
    var _formatRegexp = /{(\d+)}/g;
    /**
     * Helper to produce a string with a variable number of arguments. Insert variable segments
     * into the string using the {n} notation where N is the index of the argument following the string.
     * @param value string to which formatting is applied
     * @param args replacements for {n}-entries
     */
    function format(value) {
        var args = [];
        for (var _i = 1; _i < arguments.length; _i++) {
            args[_i - 1] = arguments[_i];
        }
        if (args.length === 0) {
            return value;
        }
        return value.replace(_formatRegexp, function (match, group) {
            var idx = parseInt(group, 10);
            return isNaN(idx) || idx < 0 || idx >= args.length ?
                match :
                args[idx];
        });
    }
    exports.format = format;
    /**
     * Converts HTML characters inside the string to use entities instead. Makes the string safe from
     * being used e.g. in HTMLElement.innerHTML.
     */
    function escape(html) {
        return html.replace(/[<|>|&]/g, function (match) {
            switch (match) {
                case '<': return '&lt;';
                case '>': return '&gt;';
                case '&': return '&amp;';
                default: return match;
            }
        });
    }
    exports.escape = escape;
    /**
     * Escapes regular expression characters in a given string
     */
    function escapeRegExpCharacters(value) {
        return value.replace(/[\-\\\{\}\*\+\?\|\^\$\.\,\[\]\(\)\#\s]/g, '\\$&');
    }
    exports.escapeRegExpCharacters = escapeRegExpCharacters;
    /**
     * Removes all occurrences of needle from the beginning and end of haystack.
     * @param haystack string to trim
     * @param needle the thing to trim (default is a blank)
     */
    function trim(haystack, needle) {
        if (needle === void 0) { needle = ' '; }
        var trimmed = ltrim(haystack, needle);
        return rtrim(trimmed, needle);
    }
    exports.trim = trim;
    /**
     * Removes all occurrences of needle from the beginning of haystack.
     * @param haystack string to trim
     * @param needle the thing to trim
     */
    function ltrim(haystack, needle) {
        if (!haystack || !needle) {
            return haystack;
        }
        var needleLen = needle.length;
        if (needleLen === 0 || haystack.length === 0) {
            return haystack;
        }
        var offset = 0, idx = -1;
        while ((idx = haystack.indexOf(needle, offset)) === offset) {
            offset = offset + needleLen;
        }
        return haystack.substring(offset);
    }
    exports.ltrim = ltrim;
    /**
     * Removes all occurrences of needle from the end of haystack.
     * @param haystack string to trim
     * @param needle the thing to trim
     */
    function rtrim(haystack, needle) {
        if (!haystack || !needle) {
            return haystack;
        }
        var needleLen = needle.length, haystackLen = haystack.length;
        if (needleLen === 0 || haystackLen === 0) {
            return haystack;
        }
        var offset = haystackLen, idx = -1;
        while (true) {
            idx = haystack.lastIndexOf(needle, offset - 1);
            if (idx === -1 || idx + needleLen !== offset) {
                break;
            }
            if (idx === 0) {
                return '';
            }
            offset = idx;
        }
        return haystack.substring(0, offset);
    }
    exports.rtrim = rtrim;
    function convertSimple2RegExpPattern(pattern) {
        return pattern.replace(/[\-\\\{\}\+\?\|\^\$\.\,\[\]\(\)\#\s]/g, '\\$&').replace(/[\*]/g, '.*');
    }
    exports.convertSimple2RegExpPattern = convertSimple2RegExpPattern;
    function stripWildcards(pattern) {
        return pattern.replace(/\*/g, '');
    }
    exports.stripWildcards = stripWildcards;
    /**
     * Determines if haystack starts with needle.
     */
    function startsWith(haystack, needle) {
        if (haystack.length < needle.length) {
            return false;
        }
        for (var i = 0; i < needle.length; i++) {
            if (haystack[i] !== needle[i]) {
                return false;
            }
        }
        return true;
    }
    exports.startsWith = startsWith;
    /**
     * Determines if haystack ends with needle.
     */
    function endsWith(haystack, needle) {
        var diff = haystack.length - needle.length;
        if (diff > 0) {
            return haystack.lastIndexOf(needle) === diff;
        }
        else if (diff === 0) {
            return haystack === needle;
        }
        else {
            return false;
        }
    }
    exports.endsWith = endsWith;
    function createRegExp(searchString, isRegex, matchCase, wholeWord, global) {
        if (searchString === '') {
            throw new Error('Cannot create regex from empty string');
        }
        if (!isRegex) {
            searchString = searchString.replace(/[\-\\\{\}\*\+\?\|\^\$\.\,\[\]\(\)\#\s]/g, '\\$&');
        }
        if (wholeWord) {
            if (!/\B/.test(searchString.charAt(0))) {
                searchString = '\\b' + searchString;
            }
            if (!/\B/.test(searchString.charAt(searchString.length - 1))) {
                searchString = searchString + '\\b';
            }
        }
        var modifiers = '';
        if (global) {
            modifiers += 'g';
        }
        if (!matchCase) {
            modifiers += 'i';
        }
        return new RegExp(searchString, modifiers);
    }
    exports.createRegExp = createRegExp;
    /**
     * Create a regular expression only if it is valid and it doesn't lead to endless loop.
     */
    function createSafeRegExp(searchString, isRegex, matchCase, wholeWord) {
        if (searchString === '') {
            return null;
        }
        // Try to create a RegExp out of the params
        var regex = null;
        try {
            regex = createRegExp(searchString, isRegex, matchCase, wholeWord, true);
        }
        catch (err) {
            return null;
        }
        // Guard against endless loop RegExps & wrap around try-catch as very long regexes produce an exception when executed the first time
        try {
            if (regExpLeadsToEndlessLoop(regex)) {
                return null;
            }
        }
        catch (err) {
            return null;
        }
        return regex;
    }
    exports.createSafeRegExp = createSafeRegExp;
    function regExpLeadsToEndlessLoop(regexp) {
        // Exit early if it's one of these special cases which are meant to match
        // against an empty string
        if (regexp.source === '^' || regexp.source === '^$' || regexp.source === '$') {
            return false;
        }
        // We check against an empty string. If the regular expression doesn't advance
        // (e.g. ends in an endless loop) it will match an empty string.
        var match = regexp.exec('');
        return (match && regexp.lastIndex === 0);
    }
    exports.regExpLeadsToEndlessLoop = regExpLeadsToEndlessLoop;
    /**
     * The normalize() method returns the Unicode Normalization Form of a given string. The form will be
     * the Normalization Form Canonical Composition.
     *
     * @see {@link https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/String/normalize}
     */
    exports.canNormalize = typeof (''.normalize) === 'function';
    var nonAsciiCharactersPattern = /[^\u0000-\u0080]/;
    var normalizedCache = new map_1.LinkedMap(10000); // bounded to 10000 elements
    function normalizeNFC(str) {
        if (!exports.canNormalize || !str) {
            return str;
        }
        var cached = normalizedCache.get(str);
        if (cached) {
            return cached;
        }
        var res;
        if (nonAsciiCharactersPattern.test(str)) {
            res = str.normalize('NFC');
        }
        else {
            res = str;
        }
        // Use the cache for fast lookup
        normalizedCache.set(str, res);
        return res;
    }
    exports.normalizeNFC = normalizeNFC;
    /**
     * Returns first index of the string that is not whitespace.
     * If string is empty or contains only whitespaces, returns -1
     */
    function firstNonWhitespaceIndex(str) {
        for (var i = 0, len = str.length; i < len; i++) {
            if (str.charAt(i) !== ' ' && str.charAt(i) !== '\t') {
                return i;
            }
        }
        return -1;
    }
    exports.firstNonWhitespaceIndex = firstNonWhitespaceIndex;
    /**
     * Returns the leading whitespace of the string.
     * If the string contains only whitespaces, returns entire string
     */
    function getLeadingWhitespace(str) {
        for (var i = 0, len = str.length; i < len; i++) {
            if (str.charAt(i) !== ' ' && str.charAt(i) !== '\t') {
                return str.substring(0, i);
            }
        }
        return str;
    }
    exports.getLeadingWhitespace = getLeadingWhitespace;
    /**
     * Returns last index of the string that is not whitespace.
     * If string is empty or contains only whitespaces, returns -1
     */
    function lastNonWhitespaceIndex(str, startIndex) {
        if (startIndex === void 0) { startIndex = str.length - 1; }
        for (var i = startIndex; i >= 0; i--) {
            if (str.charAt(i) !== ' ' && str.charAt(i) !== '\t') {
                return i;
            }
        }
        return -1;
    }
    exports.lastNonWhitespaceIndex = lastNonWhitespaceIndex;
    function localeCompare(strA, strB) {
        return strA.localeCompare(strB);
    }
    exports.localeCompare = localeCompare;
    function isAsciiChar(code) {
        return (code >= 97 && code <= 122) || (code >= 65 && code <= 90);
    }
    function equalsIgnoreCase(a, b) {
        var len1 = a.length, len2 = b.length;
        if (len1 !== len2) {
            return false;
        }
        for (var i = 0; i < len1; i++) {
            var codeA = a.charCodeAt(i), codeB = b.charCodeAt(i);
            if (codeA === codeB) {
                continue;
            }
            else if (isAsciiChar(codeA) && isAsciiChar(codeB)) {
                var diff = Math.abs(codeA - codeB);
                if (diff !== 0 && diff !== 32) {
                    return false;
                }
            }
            else {
                if (String.fromCharCode(codeA).toLocaleLowerCase() !== String.fromCharCode(codeB).toLocaleLowerCase()) {
                    return false;
                }
            }
        }
        return true;
    }
    exports.equalsIgnoreCase = equalsIgnoreCase;
    /**
     * @returns the length of the common prefix of the two strings.
     */
    function commonPrefixLength(a, b) {
        var i, len = Math.min(a.length, b.length);
        for (i = 0; i < len; i++) {
            if (a.charCodeAt(i) !== b.charCodeAt(i)) {
                return i;
            }
        }
        return len;
    }
    exports.commonPrefixLength = commonPrefixLength;
    /**
     * @returns the length of the common suffix of the two strings.
     */
    function commonSuffixLength(a, b) {
        var i, len = Math.min(a.length, b.length);
        var aLastIndex = a.length - 1;
        var bLastIndex = b.length - 1;
        for (i = 0; i < len; i++) {
            if (a.charCodeAt(aLastIndex - i) !== b.charCodeAt(bLastIndex - i)) {
                return i;
            }
        }
        return len;
    }
    exports.commonSuffixLength = commonSuffixLength;
    // --- unicode
    // http://en.wikipedia.org/wiki/Surrogate_pair
    // Returns the code point starting at a specified index in a string
    // Code points U+0000 to U+D7FF and U+E000 to U+FFFF are represented on a single character
    // Code points U+10000 to U+10FFFF are represented on two consecutive characters
    //export function getUnicodePoint(str:string, index:number, len:number):number {
    //	let chrCode = str.charCodeAt(index);
    //	if (0xD800 <= chrCode && chrCode <= 0xDBFF && index + 1 < len) {
    //		let nextChrCode = str.charCodeAt(index + 1);
    //		if (0xDC00 <= nextChrCode && nextChrCode <= 0xDFFF) {
    //			return (chrCode - 0xD800) << 10 + (nextChrCode - 0xDC00) + 0x10000;
    //		}
    //	}
    //	return chrCode;
    //}
    //export function isLeadSurrogate(chr:string) {
    //	let chrCode = chr.charCodeAt(0);
    //	return ;
    //}
    //
    //export function isTrailSurrogate(chr:string) {
    //	let chrCode = chr.charCodeAt(0);
    //	return 0xDC00 <= chrCode && chrCode <= 0xDFFF;
    //}
    function isFullWidthCharacter(charCode) {
        // Do a cheap trick to better support wrapping of wide characters, treat them as 2 columns
        // http://jrgraphix.net/research/unicode_blocks.php
        //          2E80 — 2EFF   CJK Radicals Supplement
        //          2F00 — 2FDF   Kangxi Radicals
        //          2FF0 — 2FFF   Ideographic Description Characters
        //          3000 — 303F   CJK Symbols and Punctuation
        //          3040 — 309F   Hiragana
        //          30A0 — 30FF   Katakana
        //          3100 — 312F   Bopomofo
        //          3130 — 318F   Hangul Compatibility Jamo
        //          3190 — 319F   Kanbun
        //          31A0 — 31BF   Bopomofo Extended
        //          31F0 — 31FF   Katakana Phonetic Extensions
        //          3200 — 32FF   Enclosed CJK Letters and Months
        //          3300 — 33FF   CJK Compatibility
        //          3400 — 4DBF   CJK Unified Ideographs Extension A
        //          4DC0 — 4DFF   Yijing Hexagram Symbols
        //          4E00 — 9FFF   CJK Unified Ideographs
        //          A000 — A48F   Yi Syllables
        //          A490 — A4CF   Yi Radicals
        //          AC00 — D7AF   Hangul Syllables
        // [IGNORE] D800 — DB7F   High Surrogates
        // [IGNORE] DB80 — DBFF   High Private Use Surrogates
        // [IGNORE] DC00 — DFFF   Low Surrogates
        // [IGNORE] E000 — F8FF   Private Use Area
        //          F900 — FAFF   CJK Compatibility Ideographs
        // [IGNORE] FB00 — FB4F   Alphabetic Presentation Forms
        // [IGNORE] FB50 — FDFF   Arabic Presentation Forms-A
        // [IGNORE] FE00 — FE0F   Variation Selectors
        // [IGNORE] FE20 — FE2F   Combining Half Marks
        // [IGNORE] FE30 — FE4F   CJK Compatibility Forms
        // [IGNORE] FE50 — FE6F   Small Form Variants
        // [IGNORE] FE70 — FEFF   Arabic Presentation Forms-B
        //          FF00 — FFEF   Halfwidth and Fullwidth Forms
        //               [https://en.wikipedia.org/wiki/Halfwidth_and_fullwidth_forms]
        //               of which FF01 - FF5E fullwidth ASCII of 21 to 7E
        // [IGNORE]    and FF65 - FFDC halfwidth of Katakana and Hangul
        // [IGNORE] FFF0 — FFFF   Specials
        charCode = +charCode; // @perf
        return ((charCode >= 0x2E80 && charCode <= 0xD7AF)
            || (charCode >= 0xF900 && charCode <= 0xFAFF)
            || (charCode >= 0xFF01 && charCode <= 0xFF5E));
    }
    exports.isFullWidthCharacter = isFullWidthCharacter;
    /**
     * Computes the difference score for two strings. More similar strings have a higher score.
     * We use largest common subsequence dynamic programming approach but penalize in the end for length differences.
     * Strings that have a large length difference will get a bad default score 0.
     * Complexity - both time and space O(first.length * second.length)
     * Dynamic programming LCS computation http://en.wikipedia.org/wiki/Longest_common_subsequence_problem
     *
     * @param first a string
     * @param second a string
     */
    function difference(first, second, maxLenDelta) {
        if (maxLenDelta === void 0) { maxLenDelta = 4; }
        var lengthDifference = Math.abs(first.length - second.length);
        // We only compute score if length of the currentWord and length of entry.name are similar.
        if (lengthDifference > maxLenDelta) {
            return 0;
        }
        // Initialize LCS (largest common subsequence) matrix.
        var LCS = [];
        var zeroArray = [];
        var i, j;
        for (i = 0; i < second.length + 1; ++i) {
            zeroArray.push(0);
        }
        for (i = 0; i < first.length + 1; ++i) {
            LCS.push(zeroArray);
        }
        for (i = 1; i < first.length + 1; ++i) {
            for (j = 1; j < second.length + 1; ++j) {
                if (first[i - 1] === second[j - 1]) {
                    LCS[i][j] = LCS[i - 1][j - 1] + 1;
                }
                else {
                    LCS[i][j] = Math.max(LCS[i - 1][j], LCS[i][j - 1]);
                }
            }
        }
        return LCS[first.length][second.length] - Math.sqrt(lengthDifference);
    }
    exports.difference = difference;
    /**
     * Returns an array in which every entry is the offset of a
     * line. There is always one entry which is zero.
     */
    function computeLineStarts(text) {
        var regexp = /\r\n|\r|\n/g, ret = [0], match;
        while ((match = regexp.exec(text))) {
            ret.push(regexp.lastIndex);
        }
        return ret;
    }
    exports.computeLineStarts = computeLineStarts;
    /**
     * Given a string and a max length returns a shorted version. Shorting
     * happens at favorable positions - such as whitespace or punctuation characters.
     */
    function lcut(text, n) {
        if (text.length < n) {
            return text;
        }
        var segments = text.split(/\b/), count = 0;
        for (var i = segments.length - 1; i >= 0; i--) {
            count += segments[i].length;
            if (count > n) {
                segments.splice(0, i);
                break;
            }
        }
        return segments.join(exports.empty).replace(/^\s/, exports.empty);
    }
    exports.lcut = lcut;
    // Escape codes
    // http://en.wikipedia.org/wiki/ANSI_escape_code
    var EL = /\x1B\x5B[12]?K/g; // Erase in line
    var LF = /\xA/g; // line feed
    var COLOR_START = /\x1b\[\d+m/g; // Color
    var COLOR_END = /\x1b\[0?m/g; // Color
    function removeAnsiEscapeCodes(str) {
        if (str) {
            str = str.replace(EL, '');
            str = str.replace(LF, '\n');
            str = str.replace(COLOR_START, '');
            str = str.replace(COLOR_END, '');
        }
        return str;
    }
    exports.removeAnsiEscapeCodes = removeAnsiEscapeCodes;
    // -- UTF-8 BOM
    var __utf8_bom = 65279;
    exports.UTF8_BOM_CHARACTER = String.fromCharCode(__utf8_bom);
    function startsWithUTF8BOM(str) {
        return (str && str.length > 0 && str.charCodeAt(0) === __utf8_bom);
    }
    exports.startsWithUTF8BOM = startsWithUTF8BOM;
    /**
     * Appends two strings. If the appended result is longer than maxLength,
     * trims the start of the result and replaces it with '...'.
     */
    function appendWithLimit(first, second, maxLength) {
        var newLength = first.length + second.length;
        if (newLength > maxLength) {
            first = '...' + first.substr(newLength - maxLength);
        }
        if (second.length > maxLength) {
            first += second.substr(second.length - maxLength);
        }
        else {
            first += second;
        }
        return first;
    }
    exports.appendWithLimit = appendWithLimit;
    function safeBtoa(str) {
        return btoa(encodeURIComponent(str)); // we use encodeURIComponent because btoa fails for non Latin 1 values
    }
    exports.safeBtoa = safeBtoa;
    function repeat(s, count) {
        var result = '';
        for (var i = 0; i < count; i++) {
            result += s;
        }
        return result;
    }
    exports.repeat = repeat;
});

define(__m[80], __M([1,0,3,36]), function (require, exports, strings, map_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    // Combined filters
    /**
     * @returns A filter which combines the provided set
     * of filters with an or. The *first* filters that
     * matches defined the return value of the returned
     * filter.
     */
    function or() {
        var filter = [];
        for (var _i = 0; _i < arguments.length; _i++) {
            filter[_i - 0] = arguments[_i];
        }
        return function (word, wordToMatchAgainst) {
            for (var i = 0, len = filter.length; i < len; i++) {
                var match = filter[i](word, wordToMatchAgainst);
                if (match) {
                    return match;
                }
            }
            return null;
        };
    }
    exports.or = or;
    /**
     * @returns A filter which combines the provided set
     * of filters with an and. The combines matches are
     * returned if *all* filters match.
     */
    function and() {
        var filter = [];
        for (var _i = 0; _i < arguments.length; _i++) {
            filter[_i - 0] = arguments[_i];
        }
        return function (word, wordToMatchAgainst) {
            var result = [];
            for (var i = 0, len = filter.length; i < len; i++) {
                var match = filter[i](word, wordToMatchAgainst);
                if (!match) {
                    return null;
                }
                result = result.concat(match);
            }
            return result;
        };
    }
    exports.and = and;
    // Prefix
    exports.matchesStrictPrefix = function (word, wordToMatchAgainst) { return _matchesPrefix(false, word, wordToMatchAgainst); };
    exports.matchesPrefix = function (word, wordToMatchAgainst) { return _matchesPrefix(true, word, wordToMatchAgainst); };
    function _matchesPrefix(ignoreCase, word, wordToMatchAgainst) {
        if (!wordToMatchAgainst || wordToMatchAgainst.length === 0 || wordToMatchAgainst.length < word.length) {
            return null;
        }
        if (ignoreCase) {
            word = word.toLowerCase();
            wordToMatchAgainst = wordToMatchAgainst.toLowerCase();
        }
        for (var i = 0; i < word.length; i++) {
            if (word[i] !== wordToMatchAgainst[i]) {
                return null;
            }
        }
        return word.length > 0 ? [{ start: 0, end: word.length }] : [];
    }
    // Contiguous Substring
    function matchesContiguousSubString(word, wordToMatchAgainst) {
        var index = wordToMatchAgainst.toLowerCase().indexOf(word.toLowerCase());
        if (index === -1) {
            return null;
        }
        return [{ start: index, end: index + word.length }];
    }
    exports.matchesContiguousSubString = matchesContiguousSubString;
    // Substring
    function matchesSubString(word, wordToMatchAgainst) {
        return _matchesSubString(word.toLowerCase(), wordToMatchAgainst.toLowerCase(), 0, 0);
    }
    exports.matchesSubString = matchesSubString;
    function _matchesSubString(word, wordToMatchAgainst, i, j) {
        if (i === word.length) {
            return [];
        }
        else if (j === wordToMatchAgainst.length) {
            return null;
        }
        else {
            if (word[i] === wordToMatchAgainst[j]) {
                var result = null;
                if (result = _matchesSubString(word, wordToMatchAgainst, i + 1, j + 1)) {
                    return join({ start: j, end: j + 1 }, result);
                }
            }
            return _matchesSubString(word, wordToMatchAgainst, i, j + 1);
        }
    }
    // CamelCase
    function isLower(code) {
        return 97 <= code && code <= 122;
    }
    function isUpper(code) {
        return 65 <= code && code <= 90;
    }
    function isNumber(code) {
        return 48 <= code && code <= 57;
    }
    function isWhitespace(code) {
        return [32, 9, 10, 13].indexOf(code) > -1;
    }
    function isAlphanumeric(code) {
        return isLower(code) || isUpper(code) || isNumber(code);
    }
    function join(head, tail) {
        if (tail.length === 0) {
            tail = [head];
        }
        else if (head.end === tail[0].start) {
            tail[0].start = head.start;
        }
        else {
            tail.unshift(head);
        }
        return tail;
    }
    function nextAnchor(camelCaseWord, start) {
        for (var i = start; i < camelCaseWord.length; i++) {
            var c = camelCaseWord.charCodeAt(i);
            if (isUpper(c) || isNumber(c) || (i > 0 && !isAlphanumeric(camelCaseWord.charCodeAt(i - 1)))) {
                return i;
            }
        }
        return camelCaseWord.length;
    }
    function _matchesCamelCase(word, camelCaseWord, i, j) {
        if (i === word.length) {
            return [];
        }
        else if (j === camelCaseWord.length) {
            return null;
        }
        else if (word[i] !== camelCaseWord[j].toLowerCase()) {
            return null;
        }
        else {
            var result = null;
            var nextUpperIndex = j + 1;
            result = _matchesCamelCase(word, camelCaseWord, i + 1, j + 1);
            while (!result && (nextUpperIndex = nextAnchor(camelCaseWord, nextUpperIndex)) < camelCaseWord.length) {
                result = _matchesCamelCase(word, camelCaseWord, i + 1, nextUpperIndex);
                nextUpperIndex++;
            }
            return result === null ? null : join({ start: j, end: j + 1 }, result);
        }
    }
    // Heuristic to avoid computing camel case matcher for words that don't
    // look like camelCaseWords.
    function isCamelCaseWord(word) {
        if (word.length > 60) {
            return false;
        }
        var upper = 0, lower = 0, alpha = 0, numeric = 0, code = 0;
        for (var i = 0; i < word.length; i++) {
            code = word.charCodeAt(i);
            if (isUpper(code)) {
                upper++;
            }
            if (isLower(code)) {
                lower++;
            }
            if (isAlphanumeric(code)) {
                alpha++;
            }
            if (isNumber(code)) {
                numeric++;
            }
        }
        var upperPercent = upper / word.length;
        var lowerPercent = lower / word.length;
        var alphaPercent = alpha / word.length;
        var numericPercent = numeric / word.length;
        return lowerPercent > 0.2 && upperPercent < 0.8 && alphaPercent > 0.6 && numericPercent < 0.2;
    }
    // Heuristic to avoid computing camel case matcher for words that don't
    // look like camel case patterns.
    function isCamelCasePattern(word) {
        var upper = 0, lower = 0, code = 0, whitespace = 0;
        for (var i = 0; i < word.length; i++) {
            code = word.charCodeAt(i);
            if (isUpper(code)) {
                upper++;
            }
            if (isLower(code)) {
                lower++;
            }
            if (isWhitespace(code)) {
                whitespace++;
            }
        }
        if ((upper === 0 || lower === 0) && whitespace === 0) {
            return word.length <= 30;
        }
        else {
            return upper <= 5;
        }
    }
    function matchesCamelCase(word, camelCaseWord) {
        if (!camelCaseWord || camelCaseWord.length === 0) {
            return null;
        }
        if (!isCamelCasePattern(word)) {
            return null;
        }
        if (!isCamelCaseWord(camelCaseWord)) {
            return null;
        }
        var result = null;
        var i = 0;
        while (i < camelCaseWord.length && (result = _matchesCamelCase(word.toLowerCase(), camelCaseWord, 0, i)) === null) {
            i = nextAnchor(camelCaseWord, i + 1);
        }
        return result;
    }
    exports.matchesCamelCase = matchesCamelCase;
    // Matches beginning of words supporting non-ASCII languages
    // E.g. "gp" or "g p" will match "Git: Pull"
    // Useful in cases where the target is words (e.g. command labels)
    function matchesWords(word, target) {
        if (!target || target.length === 0) {
            return null;
        }
        var result = null;
        var i = 0;
        while (i < target.length && (result = _matchesWords(word.toLowerCase(), target, 0, i)) === null) {
            i = nextWord(target, i + 1);
        }
        return result;
    }
    exports.matchesWords = matchesWords;
    function _matchesWords(word, target, i, j) {
        if (i === word.length) {
            return [];
        }
        else if (j === target.length) {
            return null;
        }
        else if (word[i] !== target[j].toLowerCase()) {
            return null;
        }
        else {
            var result = null;
            var nextWordIndex = j + 1;
            result = _matchesWords(word, target, i + 1, j + 1);
            while (!result && (nextWordIndex = nextWord(target, nextWordIndex)) < target.length) {
                result = _matchesWords(word, target, i + 1, nextWordIndex);
                nextWordIndex++;
            }
            return result === null ? null : join({ start: j, end: j + 1 }, result);
        }
    }
    function nextWord(word, start) {
        for (var i = start; i < word.length; i++) {
            var c = word.charCodeAt(i);
            if (isWhitespace(c) || (i > 0 && isWhitespace(word.charCodeAt(i - 1)))) {
                return i;
            }
        }
        return word.length;
    }
    // Fuzzy
    (function (SubstringMatching) {
        SubstringMatching[SubstringMatching["Contiguous"] = 0] = "Contiguous";
        SubstringMatching[SubstringMatching["Separate"] = 1] = "Separate";
    })(exports.SubstringMatching || (exports.SubstringMatching = {}));
    var SubstringMatching = exports.SubstringMatching;
    exports.fuzzyContiguousFilter = or(exports.matchesPrefix, matchesCamelCase, matchesContiguousSubString);
    var fuzzySeparateFilter = or(exports.matchesPrefix, matchesCamelCase, matchesSubString);
    var fuzzyRegExpCache = new map_1.LinkedMap(10000); // bounded to 10000 elements
    function matchesFuzzy(word, wordToMatchAgainst, enableSeparateSubstringMatching) {
        if (enableSeparateSubstringMatching === void 0) { enableSeparateSubstringMatching = false; }
        if (typeof word !== 'string' || typeof wordToMatchAgainst !== 'string') {
            return null; // return early for invalid input
        }
        // Form RegExp for wildcard matches
        var regexp = fuzzyRegExpCache.get(word);
        if (!regexp) {
            regexp = new RegExp(strings.convertSimple2RegExpPattern(word), 'i');
            fuzzyRegExpCache.set(word, regexp);
        }
        // RegExp Filter
        var match = regexp.exec(wordToMatchAgainst);
        if (match) {
            return [{ start: match.index, end: match.index + match[0].length }];
        }
        // Default Filter
        return enableSeparateSubstringMatching ? fuzzySeparateFilter(word, wordToMatchAgainst) : exports.fuzzyContiguousFilter(word, wordToMatchAgainst);
    }
    exports.matchesFuzzy = matchesFuzzy;
});

define(__m[55], __M([1,0,3,13,36]), function (require, exports, strings, paths, map_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var PATH_REGEX = '[/\\\\]'; // any slash or backslash
    var NO_PATH_REGEX = '[^/\\\\]'; // any non-slash and non-backslash
    function starsToRegExp(starCount) {
        switch (starCount) {
            case 0:
                return '';
            case 1:
                return NO_PATH_REGEX + "*?"; // 1 star matches any number of characters except path separator (/ and \) - non greedy (?)
            default:
                // Matches:  (Path Sep OR Path Val followed by Path Sep OR Path Sep followed by Path Val) 0-many times
                // Group is non capturing because we don't need to capture at all (?:...)
                // Overall we use non-greedy matching because it could be that we match too much
                return "(?:" + PATH_REGEX + "|" + NO_PATH_REGEX + "+" + PATH_REGEX + "|" + PATH_REGEX + NO_PATH_REGEX + "+)*?";
        }
    }
    function splitGlobAware(pattern, splitChar) {
        if (!pattern) {
            return [];
        }
        var segments = [];
        var inBraces = false;
        var inBrackets = false;
        var char;
        var curVal = '';
        for (var i = 0; i < pattern.length; i++) {
            char = pattern[i];
            switch (char) {
                case splitChar:
                    if (!inBraces && !inBrackets) {
                        segments.push(curVal);
                        curVal = '';
                        continue;
                    }
                    break;
                case '{':
                    inBraces = true;
                    break;
                case '}':
                    inBraces = false;
                    break;
                case '[':
                    inBrackets = true;
                    break;
                case ']':
                    inBrackets = false;
                    break;
            }
            curVal += char;
        }
        // Tail
        if (curVal) {
            segments.push(curVal);
        }
        return segments;
    }
    exports.splitGlobAware = splitGlobAware;
    function parseRegExp(pattern) {
        if (!pattern) {
            return '';
        }
        var regEx = '';
        // Split up into segments for each slash found
        var segments = splitGlobAware(pattern, '/');
        // Special case where we only have globstars
        if (segments.every(function (s) { return s === '**'; })) {
            regEx = '.*';
        }
        else {
            var previousSegmentWasGlobStar_1 = false;
            segments.forEach(function (segment, index) {
                // Globstar is special
                if (segment === '**') {
                    // if we have more than one globstar after another, just ignore it
                    if (!previousSegmentWasGlobStar_1) {
                        regEx += starsToRegExp(2);
                        previousSegmentWasGlobStar_1 = true;
                    }
                    return;
                }
                // States
                var inBraces = false;
                var braceVal = '';
                var inBrackets = false;
                var bracketVal = '';
                var char;
                for (var i = 0; i < segment.length; i++) {
                    char = segment[i];
                    // Support brace expansion
                    if (char !== '}' && inBraces) {
                        braceVal += char;
                        continue;
                    }
                    // Support brackets
                    if (char !== ']' && inBrackets) {
                        var res = void 0;
                        switch (char) {
                            case '-':
                                res = char;
                                break;
                            case '^':
                                res = char;
                                break;
                            default:
                                res = strings.escapeRegExpCharacters(char);
                        }
                        bracketVal += res;
                        continue;
                    }
                    switch (char) {
                        case '{':
                            inBraces = true;
                            continue;
                        case '[':
                            inBrackets = true;
                            continue;
                        case '}':
                            var choices = splitGlobAware(braceVal, ',');
                            // Converts {foo,bar} => [foo|bar]
                            var braceRegExp = "(?:" + choices.map(function (c) { return parseRegExp(c); }).join('|') + ")";
                            regEx += braceRegExp;
                            inBraces = false;
                            braceVal = '';
                            break;
                        case ']':
                            regEx += ('[' + bracketVal + ']');
                            inBrackets = false;
                            bracketVal = '';
                            break;
                        case '?':
                            regEx += NO_PATH_REGEX; // 1 ? matches any single character except path separator (/ and \)
                            continue;
                        case '*':
                            regEx += starsToRegExp(1);
                            continue;
                        default:
                            regEx += strings.escapeRegExpCharacters(char);
                    }
                }
                // Tail: Add the slash we had split on if there is more to come and the next one is not a globstar
                if (index < segments.length - 1 && segments[index + 1] !== '**') {
                    regEx += PATH_REGEX;
                }
                // reset state
                previousSegmentWasGlobStar_1 = false;
            });
        }
        return regEx;
    }
    // regexes to check for trival glob patterns that just check for String#endsWith
    var T1 = /^\*\*\/\*\.[\w\.-]+$/; // **/*.something
    var T2 = /^\*\*\/[\w\.-]+$/; // **/something
    var T3 = /^{\*\*\/\*\.[\w\.-]+(,\*\*\/\*\.[\w\.-]+)*}$/; // {**/*.something,**/*.else}
    var Trivia;
    (function (Trivia) {
        Trivia[Trivia["T1"] = 0] = "T1";
        Trivia[Trivia["T2"] = 1] = "T2";
        Trivia[Trivia["T3"] = 2] = "T3"; // {**/*.something,**/*.else}
    })(Trivia || (Trivia = {}));
    var CACHE = new map_1.LinkedMap(10000); // bounded to 10000 elements
    function parsePattern(pattern) {
        if (!pattern) {
            return null;
        }
        // Whitespace trimming
        pattern = pattern.trim();
        // Check cache
        var parsedPattern = CACHE.get(pattern);
        if (parsedPattern) {
            if (parsedPattern.regexp) {
                parsedPattern.regexp.lastIndex = 0; // reset RegExp to its initial state to reuse it!
            }
            return parsedPattern;
        }
        parsedPattern = Object.create(null);
        // Check for Trivias
        if (T1.test(pattern)) {
            parsedPattern.trivia = Trivia.T1;
        }
        else if (T2.test(pattern)) {
            parsedPattern.trivia = Trivia.T2;
        }
        else if (T3.test(pattern)) {
            parsedPattern.trivia = Trivia.T3;
        }
        else {
            parsedPattern.regexp = toRegExp("^" + parseRegExp(pattern) + "$");
        }
        // Cache
        CACHE.set(pattern, parsedPattern);
        return parsedPattern;
    }
    function toRegExp(regEx) {
        try {
            return new RegExp(regEx);
        }
        catch (error) {
            return /.^/; // create a regex that matches nothing if we cannot parse the pattern
        }
    }
    function match(arg1, path, siblings) {
        if (!arg1 || !path) {
            return false;
        }
        // Glob with String
        if (typeof arg1 === 'string') {
            var parsedPattern = parsePattern(arg1);
            if (!parsedPattern) {
                return false;
            }
            // common pattern: **/*.txt just need endsWith check
            if (parsedPattern.trivia === Trivia.T1) {
                return strings.endsWith(path, arg1.substr(4)); // '**/*'.length === 4
            }
            // common pattern: **/some.txt just need basename check
            if (parsedPattern.trivia === Trivia.T2) {
                var base = arg1.substr(3); // '**/'.length === 3
                return path === base || strings.endsWith(path, "/" + base) || strings.endsWith(path, "\\" + base);
            }
            // repetition of common patterns (see above) {**/*.txt,**/*.png}
            if (parsedPattern.trivia === Trivia.T3) {
                return arg1.slice(1, -1).split(',').some(function (pattern) { return match(pattern, path); });
            }
            return parsedPattern.regexp.test(path);
        }
        // Glob with Expression
        return matchExpression(arg1, path, siblings);
    }
    exports.match = match;
    function matchExpression(expression, path, siblings) {
        var patterns = Object.getOwnPropertyNames(expression);
        var basename;
        var _loop_1 = function(i) {
            var pattern = patterns[i];
            var value = expression[pattern];
            if (value === false) {
                return "continue"; // pattern is disabled
            }
            // Pattern matches path
            if (match(pattern, path)) {
                // Expression Pattern is <boolean>
                if (typeof value === 'boolean') {
                    return { value: pattern };
                }
                // Expression Pattern is <SiblingClause>
                if (value && typeof value.when === 'string') {
                    if (!siblings || !siblings.length) {
                        return "continue"; // pattern is malformed or we don't have siblings
                    }
                    if (!basename) {
                        basename = strings.rtrim(paths.basename(path), paths.extname(path));
                    }
                    var clause = value;
                    var clausePattern_1 = clause.when.replace('$(basename)', basename);
                    if (siblings.some(function (sibling) { return sibling === clausePattern_1; })) {
                        return { value: pattern };
                    }
                    else {
                        return "continue"; // pattern does not match in the end because the when clause is not satisfied
                    }
                }
                // Expression is Anything
                return { value: pattern };
            }
        };
        for (var i = 0; i < patterns.length; i++) {
            var state_1 = _loop_1(i);
            if (typeof state_1 === "object") return state_1.value;
            if (state_1 === "continue") continue;
        }
        return null;
    }
});

define(__m[9], __M([1,0]), function (require, exports) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var _typeof = {
        number: 'number',
        string: 'string',
        undefined: 'undefined',
        object: 'object',
        function: 'function'
    };
    /**
     * @returns whether the provided parameter is a JavaScript Array or not.
     */
    function isArray(array) {
        if (Array.isArray) {
            return Array.isArray(array);
        }
        if (array && typeof (array.length) === _typeof.number && array.constructor === Array) {
            return true;
        }
        return false;
    }
    exports.isArray = isArray;
    /**
     * @returns whether the provided parameter is a JavaScript String or not.
     */
    function isString(str) {
        if (typeof (str) === _typeof.string || str instanceof String) {
            return true;
        }
        return false;
    }
    exports.isString = isString;
    /**
     * @returns whether the provided parameter is a JavaScript Array and each element in the array is a string.
     */
    function isStringArray(value) {
        return isArray(value) && value.every(function (elem) { return isString(elem); });
    }
    exports.isStringArray = isStringArray;
    /**
     *
     * @returns whether the provided parameter is of type `object` but **not**
     *	`null`, an `array`, a `regexp`, nor a `date`.
     */
    function isObject(obj) {
        return typeof obj === _typeof.object
            && obj !== null
            && !Array.isArray(obj)
            && !(obj instanceof RegExp)
            && !(obj instanceof Date);
    }
    exports.isObject = isObject;
    /**
     * In **contrast** to just checking `typeof` this will return `false` for `NaN`.
     * @returns whether the provided parameter is a JavaScript Number or not.
     */
    function isNumber(obj) {
        if ((typeof (obj) === _typeof.number || obj instanceof Number) && !isNaN(obj)) {
            return true;
        }
        return false;
    }
    exports.isNumber = isNumber;
    /**
     * @returns whether the provided parameter is a JavaScript Boolean or not.
     */
    function isBoolean(obj) {
        return obj === true || obj === false;
    }
    exports.isBoolean = isBoolean;
    /**
     * @returns whether the provided parameter is undefined.
     */
    function isUndefined(obj) {
        return typeof (obj) === _typeof.undefined;
    }
    exports.isUndefined = isUndefined;
    /**
     * @returns whether the provided parameter is undefined or null.
     */
    function isUndefinedOrNull(obj) {
        return isUndefined(obj) || obj === null;
    }
    exports.isUndefinedOrNull = isUndefinedOrNull;
    var hasOwnProperty = Object.prototype.hasOwnProperty;
    /**
     * @returns whether the provided parameter is an empty JavaScript Object or not.
     */
    function isEmptyObject(obj) {
        if (!isObject(obj)) {
            return false;
        }
        for (var key in obj) {
            if (hasOwnProperty.call(obj, key)) {
                return false;
            }
        }
        return true;
    }
    exports.isEmptyObject = isEmptyObject;
    /**
     * @returns whether the provided parameter is a JavaScript Function or not.
     */
    function isFunction(obj) {
        return typeof obj === _typeof.function;
    }
    exports.isFunction = isFunction;
    /**
     * @returns whether the provided parameters is are JavaScript Function or not.
     */
    function areFunctions() {
        var objects = [];
        for (var _i = 0; _i < arguments.length; _i++) {
            objects[_i - 0] = arguments[_i];
        }
        return objects && objects.length > 0 && objects.every(isFunction);
    }
    exports.areFunctions = areFunctions;
    function validateConstraints(args, constraints) {
        var len = Math.min(args.length, constraints.length);
        for (var i = 0; i < len; i++) {
            validateConstraint(args[i], constraints[i]);
        }
    }
    exports.validateConstraints = validateConstraints;
    function validateConstraint(arg, constraint) {
        if (isString(constraint)) {
            if (typeof arg !== constraint) {
                throw new Error("argument does not match constraint: typeof " + constraint);
            }
        }
        else if (isFunction(constraint)) {
            if (arg instanceof constraint) {
                return;
            }
            if (arg && arg.constructor === constraint) {
                return;
            }
            if (constraint.length === 1 && constraint.call(undefined, arg) === true) {
                return;
            }
            throw new Error("argument does not match one of these constraints: arg instanceof constraint, arg.constructor === constraint, nor constraint(arg) === true");
        }
    }
    exports.validateConstraint = validateConstraint;
    /**
     * Creates a new object of the provided class and will call the constructor with
     * any additional argument supplied.
     */
    function create(ctor) {
        var args = [];
        for (var _i = 1; _i < arguments.length; _i++) {
            args[_i - 1] = arguments[_i];
        }
        var obj = Object.create(ctor.prototype);
        ctor.apply(obj, args);
        return obj;
    }
    exports.create = create;
});

define(__m[86], __M([1,0,9,49]), function (require, exports, types_1, collections_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    function newNode(data) {
        return {
            data: data,
            incoming: Object.create(null),
            outgoing: Object.create(null)
        };
    }
    var Graph = (function () {
        function Graph(_hashFn) {
            this._hashFn = _hashFn;
            this._nodes = Object.create(null);
            // empty
        }
        Graph.prototype.roots = function () {
            var ret = [];
            collections_1.forEach(this._nodes, function (entry) {
                if (types_1.isEmptyObject(entry.value.outgoing)) {
                    ret.push(entry.value);
                }
            });
            return ret;
        };
        Graph.prototype.traverse = function (start, inwards, callback) {
            var startNode = this.lookup(start);
            if (!startNode) {
                return;
            }
            this._traverse(startNode, inwards, Object.create(null), callback);
        };
        Graph.prototype._traverse = function (node, inwards, seen, callback) {
            var _this = this;
            var key = this._hashFn(node.data);
            if (collections_1.contains(seen, key)) {
                return;
            }
            seen[key] = true;
            callback(node.data);
            var nodes = inwards ? node.outgoing : node.incoming;
            collections_1.forEach(nodes, function (entry) { return _this._traverse(entry.value, inwards, seen, callback); });
        };
        Graph.prototype.insertEdge = function (from, to) {
            var fromNode = this.lookupOrInsertNode(from), toNode = this.lookupOrInsertNode(to);
            fromNode.outgoing[this._hashFn(to)] = toNode;
            toNode.incoming[this._hashFn(from)] = fromNode;
        };
        Graph.prototype.removeNode = function (data) {
            var key = this._hashFn(data);
            delete this._nodes[key];
            collections_1.forEach(this._nodes, function (entry) {
                delete entry.value.outgoing[key];
                delete entry.value.incoming[key];
            });
        };
        Graph.prototype.lookupOrInsertNode = function (data) {
            var key = this._hashFn(data), node = collections_1.lookup(this._nodes, key);
            if (!node) {
                node = newNode(data);
                this._nodes[key] = node;
            }
            return node;
        };
        Graph.prototype.lookup = function (data) {
            return collections_1.lookup(this._nodes, this._hashFn(data));
        };
        Object.defineProperty(Graph.prototype, "length", {
            get: function () {
                return Object.keys(this._nodes).length;
            },
            enumerable: true,
            configurable: true
        });
        Graph.prototype.toString = function () {
            var data = [];
            collections_1.forEach(this._nodes, function (entry) {
                data.push(entry.key + ", (incoming)[" + Object.keys(entry.value.incoming).join(', ') + "], (outgoing)[" + Object.keys(entry.value.outgoing).join(',') + "]");
            });
            return data.join('\n');
        };
        return Graph;
    }());
    exports.Graph = Graph;
});






define(__m[16], __M([1,0,9]), function (require, exports, types_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    exports.empty = Object.freeze({
        dispose: function () { }
    });
    function dispose() {
        var disposables = [];
        for (var _i = 0; _i < arguments.length; _i++) {
            disposables[_i - 0] = arguments[_i];
        }
        var first = disposables[0];
        if (types_1.isArray(first)) {
            disposables = first;
        }
        disposables.forEach(function (d) { return d && d.dispose(); });
        return [];
    }
    exports.dispose = dispose;
    function combinedDisposable(disposables) {
        return { dispose: function () { return dispose(disposables); } };
    }
    exports.combinedDisposable = combinedDisposable;
    function toDisposable() {
        var fns = [];
        for (var _i = 0; _i < arguments.length; _i++) {
            fns[_i - 0] = arguments[_i];
        }
        return combinedDisposable(fns.map(function (fn) { return ({ dispose: fn }); }));
    }
    exports.toDisposable = toDisposable;
    var Disposable = (function () {
        function Disposable() {
            this._toDispose = [];
        }
        Disposable.prototype.dispose = function () {
            this._toDispose = dispose(this._toDispose);
        };
        Disposable.prototype._register = function (t) {
            this._toDispose.push(t);
            return t;
        };
        return Disposable;
    }());
    exports.Disposable = Disposable;
    var Disposables = (function (_super) {
        __extends(Disposables, _super);
        function Disposables() {
            _super.apply(this, arguments);
        }
        Disposables.prototype.add = function (arg) {
            if (!Array.isArray(arg)) {
                return this._register(arg);
            }
            else {
                for (var _i = 0, arg_1 = arg; _i < arg_1.length; _i++) {
                    var element = arg_1[_i];
                    return this._register(element);
                }
            }
        };
        return Disposables;
    }(Disposable));
    exports.Disposables = Disposables;
});

define(__m[46], __M([1,0,13,9,3,55]), function (require, exports, paths, types, strings, glob_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    exports.MIME_TEXT = 'text/plain';
    exports.MIME_BINARY = 'application/octet-stream';
    exports.MIME_UNKNOWN = 'application/unknown';
    var registeredAssociations = [];
    /**
     * Associate a text mime to the registry.
     */
    function registerTextMime(association) {
        // Register
        registeredAssociations.push(association);
        // Check for conflicts unless this is a user configured association
        if (!association.userConfigured) {
            registeredAssociations.forEach(function (a) {
                if (a.mime === association.mime || a.userConfigured) {
                    return; // same mime or userConfigured is ok
                }
                if (association.extension && a.extension === association.extension) {
                    console.warn("Overwriting extension <<" + association.extension + ">> to now point to mime <<" + association.mime + ">>");
                }
                if (association.filename && a.filename === association.filename) {
                    console.warn("Overwriting filename <<" + association.filename + ">> to now point to mime <<" + association.mime + ">>");
                }
                if (association.filepattern && a.filepattern === association.filepattern) {
                    console.warn("Overwriting filepattern <<" + association.filepattern + ">> to now point to mime <<" + association.mime + ">>");
                }
                if (association.firstline && a.firstline === association.firstline) {
                    console.warn("Overwriting firstline <<" + association.firstline + ">> to now point to mime <<" + association.mime + ">>");
                }
            });
        }
    }
    exports.registerTextMime = registerTextMime;
    /**
     * Clear text mimes from the registry.
     */
    function clearTextMimes(onlyUserConfigured) {
        if (!onlyUserConfigured) {
            registeredAssociations = [];
        }
        else {
            registeredAssociations = registeredAssociations.filter(function (a) { return !a.userConfigured; });
        }
    }
    exports.clearTextMimes = clearTextMimes;
    /**
     * Given a file, return the best matching mime type for it
     */
    function guessMimeTypes(path, firstLine) {
        if (!path) {
            return [exports.MIME_UNKNOWN];
        }
        path = path.toLowerCase();
        // 1.) User configured mappings have highest priority
        var configuredMime = guessMimeTypeByPath(path, registeredAssociations.filter(function (a) { return a.userConfigured; }));
        if (configuredMime) {
            return [configuredMime, exports.MIME_TEXT];
        }
        // 2.) Registered mappings have middle priority
        var registeredMime = guessMimeTypeByPath(path, registeredAssociations.filter(function (a) { return !a.userConfigured; }));
        if (registeredMime) {
            return [registeredMime, exports.MIME_TEXT];
        }
        // 3.) Firstline has lowest priority
        if (firstLine) {
            var firstlineMime = guessMimeTypeByFirstline(firstLine);
            if (firstlineMime) {
                return [firstlineMime, exports.MIME_TEXT];
            }
        }
        return [exports.MIME_UNKNOWN];
    }
    exports.guessMimeTypes = guessMimeTypes;
    function guessMimeTypeByPath(path, associations) {
        var filename = paths.basename(path);
        var filenameMatch;
        var patternMatch;
        var extensionMatch;
        for (var i = 0; i < associations.length; i++) {
            var association = associations[i];
            // First exact name match
            if (association.filename && filename === association.filename.toLowerCase()) {
                filenameMatch = association;
                break; // take it!
            }
            // Longest pattern match
            if (association.filepattern) {
                var target = association.filepattern.indexOf(paths.sep) >= 0 ? path : filename; // match on full path if pattern contains path separator
                if (glob_1.match(association.filepattern.toLowerCase(), target)) {
                    if (!patternMatch || association.filepattern.length > patternMatch.filepattern.length) {
                        patternMatch = association;
                    }
                }
            }
            // Longest extension match
            if (association.extension) {
                if (strings.endsWith(filename, association.extension.toLowerCase())) {
                    if (!extensionMatch || association.extension.length > extensionMatch.extension.length) {
                        extensionMatch = association;
                    }
                }
            }
        }
        // 1.) Exact name match has second highest prio
        if (filenameMatch) {
            return filenameMatch.mime;
        }
        // 2.) Match on pattern
        if (patternMatch) {
            return patternMatch.mime;
        }
        // 3.) Match on extension comes next
        if (extensionMatch) {
            return extensionMatch.mime;
        }
        return null;
    }
    function guessMimeTypeByFirstline(firstLine) {
        if (strings.startsWithUTF8BOM(firstLine)) {
            firstLine = firstLine.substr(1);
        }
        if (firstLine.length > 0) {
            for (var i = 0; i < registeredAssociations.length; ++i) {
                var association = registeredAssociations[i];
                if (!association.firstline) {
                    continue;
                }
                // Make sure the entire line matches, not just a subpart.
                var matches = firstLine.match(association.firstline);
                if (matches && matches.length > 0 && matches[0].length === firstLine.length) {
                    return association.mime;
                }
            }
        }
        return null;
    }
    function isBinaryMime(mimes) {
        if (!mimes) {
            return false;
        }
        var mimeVals;
        if (types.isArray(mimes)) {
            mimeVals = mimes;
        }
        else {
            mimeVals = mimes.split(',').map(function (mime) { return mime.trim(); });
        }
        return mimeVals.indexOf(exports.MIME_BINARY) >= 0;
    }
    exports.isBinaryMime = isBinaryMime;
    function isUnspecific(mime) {
        if (!mime) {
            return true;
        }
        if (typeof mime === 'string') {
            return mime === exports.MIME_BINARY || mime === exports.MIME_TEXT || mime === exports.MIME_UNKNOWN;
        }
        return mime.length === 1 && isUnspecific(mime[0]);
    }
    exports.isUnspecific = isUnspecific;
    function suggestFilename(theMime, prefix) {
        for (var i = 0; i < registeredAssociations.length; i++) {
            var association = registeredAssociations[i];
            if (association.userConfigured) {
                continue; // only support registered ones
            }
            if (association.mime === theMime && association.extension) {
                return prefix + association.extension;
            }
        }
        return null;
    }
    exports.suggestFilename = suggestFilename;
});

define(__m[12], __M([1,0,9]), function (require, exports, Types) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    function clone(obj) {
        if (!obj || typeof obj !== 'object') {
            return obj;
        }
        if (obj instanceof RegExp) {
            return obj;
        }
        var result = (Array.isArray(obj)) ? [] : {};
        Object.keys(obj).forEach(function (key) {
            if (obj[key] && typeof obj[key] === 'object') {
                result[key] = clone(obj[key]);
            }
            else {
                result[key] = obj[key];
            }
        });
        return result;
    }
    exports.clone = clone;
    function deepClone(obj) {
        if (!obj || typeof obj !== 'object') {
            return obj;
        }
        var result = (Array.isArray(obj)) ? [] : {};
        Object.getOwnPropertyNames(obj).forEach(function (key) {
            if (obj[key] && typeof obj[key] === 'object') {
                result[key] = deepClone(obj[key]);
            }
            else {
                result[key] = obj[key];
            }
        });
        return result;
    }
    exports.deepClone = deepClone;
    var hasOwnProperty = Object.prototype.hasOwnProperty;
    function cloneAndChange(obj, changer) {
        return _cloneAndChange(obj, changer, []);
    }
    exports.cloneAndChange = cloneAndChange;
    function _cloneAndChange(obj, changer, encounteredObjects) {
        if (Types.isUndefinedOrNull(obj)) {
            return obj;
        }
        var changed = changer(obj);
        if (typeof changed !== 'undefined') {
            return changed;
        }
        if (Types.isArray(obj)) {
            var r1 = [];
            for (var i1 = 0; i1 < obj.length; i1++) {
                r1.push(_cloneAndChange(obj[i1], changer, encounteredObjects));
            }
            return r1;
        }
        if (Types.isObject(obj)) {
            if (encounteredObjects.indexOf(obj) >= 0) {
                throw new Error('Cannot clone recursive data-structure');
            }
            encounteredObjects.push(obj);
            var r2 = {};
            for (var i2 in obj) {
                if (hasOwnProperty.call(obj, i2)) {
                    r2[i2] = _cloneAndChange(obj[i2], changer, encounteredObjects);
                }
            }
            encounteredObjects.pop();
            return r2;
        }
        return obj;
    }
    // DON'T USE THESE FUNCTION UNLESS YOU KNOW HOW CHROME
    // WORKS... WE HAVE SEEN VERY WEIRD BEHAVIOUR WITH CHROME >= 37
    ///**
    // * Recursively call Object.freeze on object and any properties that are objects.
    // */
    //export function deepFreeze(obj:any):void {
    //	Object.freeze(obj);
    //	Object.keys(obj).forEach((key) => {
    //		if(!(typeof obj[key] === 'object') || Object.isFrozen(obj[key])) {
    //			return;
    //		}
    //
    //		deepFreeze(obj[key]);
    //	});
    //	if(!Object.isFrozen(obj)) {
    //		console.log('too warm');
    //	}
    //}
    //
    //export function deepSeal(obj:any):void {
    //	Object.seal(obj);
    //	Object.keys(obj).forEach((key) => {
    //		if(!(typeof obj[key] === 'object') || Object.isSealed(obj[key])) {
    //			return;
    //		}
    //
    //		deepSeal(obj[key]);
    //	});
    //	if(!Object.isSealed(obj)) {
    //		console.log('NOT sealed');
    //	}
    //}
    /**
     * Copies all properties of source into destination. The optional parameter "overwrite" allows to control
     * if existing properties on the destination should be overwritten or not. Defaults to true (overwrite).
     */
    function mixin(destination, source, overwrite) {
        if (overwrite === void 0) { overwrite = true; }
        if (!Types.isObject(destination)) {
            return source;
        }
        if (Types.isObject(source)) {
            Object.keys(source).forEach(function (key) {
                if (key in destination) {
                    if (overwrite) {
                        if (Types.isObject(destination[key]) && Types.isObject(source[key])) {
                            mixin(destination[key], source[key], overwrite);
                        }
                        else {
                            destination[key] = source[key];
                        }
                    }
                }
                else {
                    destination[key] = source[key];
                }
            });
        }
        return destination;
    }
    exports.mixin = mixin;
    function assign(destination) {
        var sources = [];
        for (var _i = 1; _i < arguments.length; _i++) {
            sources[_i - 1] = arguments[_i];
        }
        sources.forEach(function (source) { return Object.keys(source).forEach(function (key) { return destination[key] = source[key]; }); });
        return destination;
    }
    exports.assign = assign;
    function toObject(arr, keyMap, valueMap) {
        if (valueMap === void 0) { valueMap = function (x) { return x; }; }
        return arr.reduce(function (o, d) { return assign(o, (_a = {}, _a[keyMap(d)] = valueMap(d), _a)); var _a; }, Object.create(null));
    }
    exports.toObject = toObject;
    function equals(one, other) {
        if (one === other) {
            return true;
        }
        if (one === null || one === undefined || other === null || other === undefined) {
            return false;
        }
        if (typeof one !== typeof other) {
            return false;
        }
        if (typeof one !== 'object') {
            return false;
        }
        if ((Array.isArray(one)) !== (Array.isArray(other))) {
            return false;
        }
        var i, key;
        if (Array.isArray(one)) {
            if (one.length !== other.length) {
                return false;
            }
            for (i = 0; i < one.length; i++) {
                if (!equals(one[i], other[i])) {
                    return false;
                }
            }
        }
        else {
            var oneKeys = [];
            for (key in one) {
                oneKeys.push(key);
            }
            oneKeys.sort();
            var otherKeys = [];
            for (key in other) {
                otherKeys.push(key);
            }
            otherKeys.sort();
            if (!equals(oneKeys, otherKeys)) {
                return false;
            }
            for (i = 0; i < oneKeys.length; i++) {
                if (!equals(one[oneKeys[i]], other[oneKeys[i]])) {
                    return false;
                }
            }
        }
        return true;
    }
    exports.equals = equals;
    function ensureProperty(obj, property, defaultValue) {
        if (typeof obj[property] === 'undefined') {
            obj[property] = defaultValue;
        }
    }
    exports.ensureProperty = ensureProperty;
    function arrayToHash(array) {
        var result = {};
        for (var i = 0; i < array.length; ++i) {
            result[array[i]] = true;
        }
        return result;
    }
    exports.arrayToHash = arrayToHash;
    /**
     * Given an array of strings, returns a function which, given a string
     * returns true or false whether the string is in that array.
     */
    function createKeywordMatcher(arr, caseInsensitive) {
        if (caseInsensitive === void 0) { caseInsensitive = false; }
        if (caseInsensitive) {
            arr = arr.map(function (x) { return x.toLowerCase(); });
        }
        var hash = arrayToHash(arr);
        if (caseInsensitive) {
            return function (word) {
                return hash[word.toLowerCase()] !== undefined && hash.hasOwnProperty(word.toLowerCase());
            };
        }
        else {
            return function (word) {
                return hash[word] !== undefined && hash.hasOwnProperty(word);
            };
        }
    }
    exports.createKeywordMatcher = createKeywordMatcher;
    /**
     * Started from TypeScript's __extends function to make a type a subclass of a specific class.
     * Modified to work with properties already defined on the derivedClass, since we can't get TS
     * to call this method before the constructor definition.
     */
    function derive(baseClass, derivedClass) {
        for (var prop in baseClass) {
            if (baseClass.hasOwnProperty(prop)) {
                derivedClass[prop] = baseClass[prop];
            }
        }
        derivedClass = derivedClass || function () { };
        var basePrototype = baseClass.prototype;
        var derivedPrototype = derivedClass.prototype;
        derivedClass.prototype = Object.create(basePrototype);
        for (var prop in derivedPrototype) {
            if (derivedPrototype.hasOwnProperty(prop)) {
                // handle getters and setters properly
                Object.defineProperty(derivedClass.prototype, prop, Object.getOwnPropertyDescriptor(derivedPrototype, prop));
            }
        }
        // Cast to any due to Bug 16188:PropertyDescriptor set and get function should be optional.
        Object.defineProperty(derivedClass.prototype, 'constructor', { value: derivedClass, writable: true, configurable: true, enumerable: true });
    }
    exports.derive = derive;
    /**
     * Calls JSON.Stringify with a replacer to break apart any circular references.
     * This prevents JSON.stringify from throwing the exception
     *  "Uncaught TypeError: Converting circular structure to JSON"
     */
    function safeStringify(obj) {
        var seen = [];
        return JSON.stringify(obj, function (key, value) {
            if (Types.isObject(value) || Array.isArray(value)) {
                if (seen.indexOf(value) !== -1) {
                    return '[Circular]';
                }
                else {
                    seen.push(value);
                }
            }
            return value;
        });
    }
    exports.safeStringify = safeStringify;
    function getOrDefault(obj, fn, defaultValue) {
        if (defaultValue === void 0) { defaultValue = null; }
        var result = fn(obj);
        return typeof result === 'undefined' ? defaultValue : result;
    }
    exports.getOrDefault = getOrDefault;
});

define(__m[20], __M([1,0,10]), function (require, exports, platform) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    function _encode(ch) {
        return '%' + ch.charCodeAt(0).toString(16).toUpperCase();
    }
    // see https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/encodeURIComponent
    function encodeURIComponent2(str) {
        return encodeURIComponent(str).replace(/[!'()*]/g, _encode);
    }
    function encodeNoop(str) {
        return str;
    }
    /**
     * Uniform Resource Identifier (URI) http://tools.ietf.org/html/rfc3986.
     * This class is a simple parser which creates the basic component paths
     * (http://tools.ietf.org/html/rfc3986#section-3) with minimal validation
     * and encoding.
     *
     *       foo://example.com:8042/over/there?name=ferret#nose
     *       \_/   \______________/\_________/ \_________/ \__/
     *        |           |            |            |        |
     *     scheme     authority       path        query   fragment
     *        |   _____________________|__
     *       / \ /                        \
     *       urn:example:animal:ferret:nose
     *
     *
     */
    var URI = (function () {
        function URI() {
            this._scheme = URI._empty;
            this._authority = URI._empty;
            this._path = URI._empty;
            this._query = URI._empty;
            this._fragment = URI._empty;
            this._formatted = null;
            this._fsPath = null;
        }
        Object.defineProperty(URI.prototype, "scheme", {
            /**
             * scheme is the 'http' part of 'http://www.msft.com/some/path?query#fragment'.
             * The part before the first colon.
             */
            get: function () {
                return this._scheme;
            },
            enumerable: true,
            configurable: true
        });
        Object.defineProperty(URI.prototype, "authority", {
            /**
             * authority is the 'www.msft.com' part of 'http://www.msft.com/some/path?query#fragment'.
             * The part between the first double slashes and the next slash.
             */
            get: function () {
                return this._authority;
            },
            enumerable: true,
            configurable: true
        });
        Object.defineProperty(URI.prototype, "path", {
            /**
             * path is the '/some/path' part of 'http://www.msft.com/some/path?query#fragment'.
             */
            get: function () {
                return this._path;
            },
            enumerable: true,
            configurable: true
        });
        Object.defineProperty(URI.prototype, "query", {
            /**
             * query is the 'query' part of 'http://www.msft.com/some/path?query#fragment'.
             */
            get: function () {
                return this._query;
            },
            enumerable: true,
            configurable: true
        });
        Object.defineProperty(URI.prototype, "fragment", {
            /**
             * fragment is the 'fragment' part of 'http://www.msft.com/some/path?query#fragment'.
             */
            get: function () {
                return this._fragment;
            },
            enumerable: true,
            configurable: true
        });
        Object.defineProperty(URI.prototype, "fsPath", {
            // ---- filesystem path -----------------------
            /**
             * Returns a string representing the corresponding file system path of this URI.
             * Will handle UNC paths and normalize windows drive letters to lower-case. Also
             * uses the platform specific path separator. Will *not* validate the path for
             * invalid characters and semantics. Will *not* look at the scheme of this URI.
             */
            get: function () {
                if (!this._fsPath) {
                    var value;
                    if (this._authority && this.scheme === 'file') {
                        // unc path: file://shares/c$/far/boo
                        value = "//" + this._authority + this._path;
                    }
                    else if (URI._driveLetterPath.test(this._path)) {
                        // windows drive letter: file:///c:/far/boo
                        value = this._path[1].toLowerCase() + this._path.substr(2);
                    }
                    else {
                        // other path
                        value = this._path;
                    }
                    if (platform.isWindows) {
                        value = value.replace(/\//g, '\\');
                    }
                    this._fsPath = value;
                }
                return this._fsPath;
            },
            enumerable: true,
            configurable: true
        });
        // ---- modify to new -------------------------
        URI.prototype.with = function (change) {
            if (!change) {
                return this;
            }
            var scheme = change.scheme || this.scheme;
            var authority = change.authority || this.authority;
            var path = change.path || this.path;
            var query = change.query || this.query;
            var fragment = change.fragment || this.fragment;
            if (scheme === this.scheme
                && authority === this.authority
                && path === this.path
                && query === this.query
                && fragment === this.fragment) {
                return this;
            }
            var ret = new URI();
            ret._scheme = scheme;
            ret._authority = authority;
            ret._path = path;
            ret._query = query;
            ret._fragment = fragment;
            URI._validate(ret);
            return ret;
        };
        // ---- parse & validate ------------------------
        URI.parse = function (value) {
            var ret = new URI();
            var data = URI._parseComponents(value);
            ret._scheme = data.scheme;
            ret._authority = decodeURIComponent(data.authority);
            ret._path = decodeURIComponent(data.path);
            ret._query = decodeURIComponent(data.query);
            ret._fragment = decodeURIComponent(data.fragment);
            URI._validate(ret);
            return ret;
        };
        URI.file = function (path) {
            var ret = new URI();
            ret._scheme = 'file';
            // normalize to fwd-slashes
            path = path.replace(/\\/g, URI._slash);
            // check for authority as used in UNC shares
            // or use the path as given
            if (path[0] === URI._slash && path[0] === path[1]) {
                var idx = path.indexOf(URI._slash, 2);
                if (idx === -1) {
                    ret._authority = path.substring(2);
                }
                else {
                    ret._authority = path.substring(2, idx);
                    ret._path = path.substring(idx);
                }
            }
            else {
                ret._path = path;
            }
            // Ensure that path starts with a slash
            // or that it is at least a slash
            if (ret._path[0] !== URI._slash) {
                ret._path = URI._slash + ret._path;
            }
            URI._validate(ret);
            return ret;
        };
        URI._parseComponents = function (value) {
            var ret = {
                scheme: URI._empty,
                authority: URI._empty,
                path: URI._empty,
                query: URI._empty,
                fragment: URI._empty,
            };
            var match = URI._regexp.exec(value);
            if (match) {
                ret.scheme = match[2] || ret.scheme;
                ret.authority = match[4] || ret.authority;
                ret.path = match[5] || ret.path;
                ret.query = match[7] || ret.query;
                ret.fragment = match[9] || ret.fragment;
            }
            return ret;
        };
        URI.from = function (components) {
            return new URI().with(components);
        };
        URI._validate = function (ret) {
            // validation
            // path, http://tools.ietf.org/html/rfc3986#section-3.3
            // If a URI contains an authority component, then the path component
            // must either be empty or begin with a slash ("/") character.  If a URI
            // does not contain an authority component, then the path cannot begin
            // with two slash characters ("//").
            if (ret.authority && ret.path && ret.path[0] !== '/') {
                throw new Error('[UriError]: If a URI contains an authority component, then the path component must either be empty or begin with a slash ("/") character');
            }
            if (!ret.authority && ret.path.indexOf('//') === 0) {
                throw new Error('[UriError]: If a URI does not contain an authority component, then the path cannot begin with two slash characters ("//")');
            }
        };
        // ---- printing/externalize ---------------------------
        /**
         *
         * @param skipEncoding Do not encode the result, default is `false`
         */
        URI.prototype.toString = function (skipEncoding) {
            if (skipEncoding === void 0) { skipEncoding = false; }
            if (!skipEncoding) {
                if (!this._formatted) {
                    this._formatted = URI._asFormatted(this, false);
                }
                return this._formatted;
            }
            else {
                // we don't cache that
                return URI._asFormatted(this, true);
            }
        };
        URI._asFormatted = function (uri, skipEncoding) {
            var encoder = !skipEncoding
                ? encodeURIComponent2
                : encodeNoop;
            var parts = [];
            var scheme = uri.scheme, authority = uri.authority, path = uri.path, query = uri.query, fragment = uri.fragment;
            if (scheme) {
                parts.push(scheme, ':');
            }
            if (authority || scheme === 'file') {
                parts.push('//');
            }
            if (authority) {
                authority = authority.toLowerCase();
                var idx = authority.indexOf(':');
                if (idx === -1) {
                    parts.push(encoder(authority));
                }
                else {
                    parts.push(encoder(authority.substr(0, idx)), authority.substr(idx));
                }
            }
            if (path) {
                // lower-case windown drive letters in /C:/fff
                var m = URI._upperCaseDrive.exec(path);
                if (m) {
                    path = m[1] + m[2].toLowerCase() + path.substr(m[1].length + m[2].length);
                }
                // encode every segement but not slashes
                // make sure that # and ? are always encoded
                // when occurring in paths - otherwise the result
                // cannot be parsed back again
                var lastIdx = 0;
                while (true) {
                    var idx = path.indexOf(URI._slash, lastIdx);
                    if (idx === -1) {
                        parts.push(encoder(path.substring(lastIdx)).replace(/[#?]/, _encode));
                        break;
                    }
                    parts.push(encoder(path.substring(lastIdx, idx)).replace(/[#?]/, _encode), URI._slash);
                    lastIdx = idx + 1;
                }
                ;
            }
            if (query) {
                parts.push('?', encoder(query));
            }
            if (fragment) {
                parts.push('#', encoder(fragment));
            }
            return parts.join(URI._empty);
        };
        URI.prototype.toJSON = function () {
            return {
                scheme: this.scheme,
                authority: this.authority,
                path: this.path,
                fsPath: this.fsPath,
                query: this.query,
                fragment: this.fragment,
                external: this.toString(),
                $mid: 1
            };
        };
        URI.revive = function (data) {
            var result = new URI();
            result._scheme = data.scheme;
            result._authority = data.authority;
            result._path = data.path;
            result._query = data.query;
            result._fragment = data.fragment;
            result._fsPath = data.fsPath;
            result._formatted = data.external;
            URI._validate(result);
            return result;
        };
        URI._empty = '';
        URI._slash = '/';
        URI._regexp = /^(([^:/?#]+?):)?(\/\/([^/?#]*))?([^?#]*)(\?([^#]*))?(#(.*))?/;
        URI._driveLetterPath = /^\/[a-zA-z]:/;
        URI._upperCaseDrive = /^(\/)?([A-Z]:)/;
        return URI;
    }());
    Object.defineProperty(exports, "__esModule", { value: true });
    exports.default = URI;
});

define(__m[89], __M([1,0,20]), function (require, exports, uri_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    function stringify(obj) {
        return JSON.stringify(obj, replacer);
    }
    exports.stringify = stringify;
    function parse(text) {
        return JSON.parse(text, reviver);
    }
    exports.parse = parse;
    function replacer(key, value) {
        // URI is done via toJSON-member
        if (value instanceof RegExp) {
            return {
                $mid: 2,
                source: value.source,
                flags: (value.global ? 'g' : '') + (value.ignoreCase ? 'i' : '') + (value.multiline ? 'm' : ''),
            };
        }
        return value;
    }
    function reviver(key, value) {
        var marshallingConst;
        if (value !== void 0 && value !== null) {
            marshallingConst = value.$mid;
        }
        if (marshallingConst === 1) {
            return uri_1.default.revive(value);
        }
        else if (marshallingConst === 2) {
            return new RegExp(value.source, value.flags);
        }
        else {
            return value;
        }
    }
});

/**
 * Extracted from https://github.com/winjs/winjs
 * Version: 4.4.0(ec3258a9f3a36805a187848984e3bb938044178d)
 * Copyright (c) Microsoft Corporation.
 * All Rights Reserved.
 * Licensed under the MIT License.
 */
(function() {

var _modules = {};
_modules["WinJS/Core/_WinJS"] = {};

var _winjs = function(moduleId, deps, factory) {
    var exports = {};
    var exportsPassedIn = false;

    var depsValues = deps.map(function(dep) {
        if (dep === 'exports') {
            exportsPassedIn = true;
            return exports;
        }
        return _modules[dep];
    });

    var result = factory.apply({}, depsValues);

    _modules[moduleId] = exportsPassedIn ? exports : result;
};


_winjs("WinJS/Core/_Global", [], function () {
    "use strict";

    // Appease jshint
    /* global window, self, global */

    var globalObject =
        typeof window !== 'undefined' ? window :
        typeof self !== 'undefined' ? self :
        typeof global !== 'undefined' ? global :
        {};
    return globalObject;
});

_winjs("WinJS/Core/_BaseCoreUtils", ["WinJS/Core/_Global"], function baseCoreUtilsInit(_Global) {
    "use strict";

    var hasWinRT = !!_Global.Windows;

    function markSupportedForProcessing(func) {
        /// <signature helpKeyword="WinJS.Utilities.markSupportedForProcessing">
        /// <summary locid="WinJS.Utilities.markSupportedForProcessing">
        /// Marks a function as being compatible with declarative processing, such as WinJS.UI.processAll
        /// or WinJS.Binding.processAll.
        /// </summary>
        /// <param name="func" type="Function" locid="WinJS.Utilities.markSupportedForProcessing_p:func">
        /// The function to be marked as compatible with declarative processing.
        /// </param>
        /// <returns type="Function" locid="WinJS.Utilities.markSupportedForProcessing_returnValue">
        /// The input function.
        /// </returns>
        /// </signature>
        func.supportedForProcessing = true;
        return func;
    }

    return {
        hasWinRT: hasWinRT,
        markSupportedForProcessing: markSupportedForProcessing,
        _setImmediate: _Global.setImmediate ? _Global.setImmediate.bind(_Global) : function (handler) {
            _Global.setTimeout(handler, 0);
        }
    };
});
_winjs("WinJS/Core/_WriteProfilerMark", ["WinJS/Core/_Global"], function profilerInit(_Global) {
    "use strict";

    return _Global.msWriteProfilerMark || function () { };
});
_winjs("WinJS/Core/_Base", ["WinJS/Core/_WinJS","WinJS/Core/_Global","WinJS/Core/_BaseCoreUtils","WinJS/Core/_WriteProfilerMark"], function baseInit(_WinJS, _Global, _BaseCoreUtils, _WriteProfilerMark) {
    "use strict";

    function initializeProperties(target, members, prefix) {
        var keys = Object.keys(members);
        var isArray = Array.isArray(target);
        var properties;
        var i, len;
        for (i = 0, len = keys.length; i < len; i++) {
            var key = keys[i];
            var enumerable = key.charCodeAt(0) !== /*_*/95;
            var member = members[key];
            if (member && typeof member === 'object') {
                if (member.value !== undefined || typeof member.get === 'function' || typeof member.set === 'function') {
                    if (member.enumerable === undefined) {
                        member.enumerable = enumerable;
                    }
                    if (prefix && member.setName && typeof member.setName === 'function') {
                        member.setName(prefix + "." + key);
                    }
                    properties = properties || {};
                    properties[key] = member;
                    continue;
                }
            }
            if (!enumerable) {
                properties = properties || {};
                properties[key] = { value: member, enumerable: enumerable, configurable: true, writable: true };
                continue;
            }
            if (isArray) {
                target.forEach(function (target) {
                    target[key] = member;
                });
            } else {
                target[key] = member;
            }
        }
        if (properties) {
            if (isArray) {
                target.forEach(function (target) {
                    Object.defineProperties(target, properties);
                });
            } else {
                Object.defineProperties(target, properties);
            }
        }
    }

    (function () {

        var _rootNamespace = _WinJS;
        if (!_rootNamespace.Namespace) {
            _rootNamespace.Namespace = Object.create(Object.prototype);
        }

        function createNamespace(parentNamespace, name) {
            var currentNamespace = parentNamespace || {};
            if (name) {
                var namespaceFragments = name.split(".");
                if (currentNamespace === _Global && namespaceFragments[0] === "WinJS") {
                    currentNamespace = _WinJS;
                    namespaceFragments.splice(0, 1);
                }
                for (var i = 0, len = namespaceFragments.length; i < len; i++) {
                    var namespaceName = namespaceFragments[i];
                    if (!currentNamespace[namespaceName]) {
                        Object.defineProperty(currentNamespace, namespaceName,
                            { value: {}, writable: false, enumerable: true, configurable: true }
                        );
                    }
                    currentNamespace = currentNamespace[namespaceName];
                }
            }
            return currentNamespace;
        }

        function defineWithParent(parentNamespace, name, members) {
            /// <signature helpKeyword="WinJS.Namespace.defineWithParent">
            /// <summary locid="WinJS.Namespace.defineWithParent">
            /// Defines a new namespace with the specified name under the specified parent namespace.
            /// </summary>
            /// <param name="parentNamespace" type="Object" locid="WinJS.Namespace.defineWithParent_p:parentNamespace">
            /// The parent namespace.
            /// </param>
            /// <param name="name" type="String" locid="WinJS.Namespace.defineWithParent_p:name">
            /// The name of the new namespace.
            /// </param>
            /// <param name="members" type="Object" locid="WinJS.Namespace.defineWithParent_p:members">
            /// The members of the new namespace.
            /// </param>
            /// <returns type="Object" locid="WinJS.Namespace.defineWithParent_returnValue">
            /// The newly-defined namespace.
            /// </returns>
            /// </signature>
            var currentNamespace = createNamespace(parentNamespace, name);

            if (members) {
                initializeProperties(currentNamespace, members, name || "<ANONYMOUS>");
            }

            return currentNamespace;
        }

        function define(name, members) {
            /// <signature helpKeyword="WinJS.Namespace.define">
            /// <summary locid="WinJS.Namespace.define">
            /// Defines a new namespace with the specified name.
            /// </summary>
            /// <param name="name" type="String" locid="WinJS.Namespace.define_p:name">
            /// The name of the namespace. This could be a dot-separated name for nested namespaces.
            /// </param>
            /// <param name="members" type="Object" locid="WinJS.Namespace.define_p:members">
            /// The members of the new namespace.
            /// </param>
            /// <returns type="Object" locid="WinJS.Namespace.define_returnValue">
            /// The newly-defined namespace.
            /// </returns>
            /// </signature>
            return defineWithParent(_Global, name, members);
        }

        var LazyStates = {
            uninitialized: 1,
            working: 2,
            initialized: 3,
        };

        function lazy(f) {
            var name;
            var state = LazyStates.uninitialized;
            var result;
            return {
                setName: function (value) {
                    name = value;
                },
                get: function () {
                    switch (state) {
                        case LazyStates.initialized:
                            return result;

                        case LazyStates.uninitialized:
                            state = LazyStates.working;
                            try {
                                _WriteProfilerMark("WinJS.Namespace._lazy:" + name + ",StartTM");
                                result = f();
                            } finally {
                                _WriteProfilerMark("WinJS.Namespace._lazy:" + name + ",StopTM");
                                state = LazyStates.uninitialized;
                            }
                            f = null;
                            state = LazyStates.initialized;
                            return result;

                        case LazyStates.working:
                            throw "Illegal: reentrancy on initialization";

                        default:
                            throw "Illegal";
                    }
                },
                set: function (value) {
                    switch (state) {
                        case LazyStates.working:
                            throw "Illegal: reentrancy on initialization";

                        default:
                            state = LazyStates.initialized;
                            result = value;
                            break;
                    }
                },
                enumerable: true,
                configurable: true,
            };
        }

        // helper for defining AMD module members
        function moduleDefine(exports, name, members) {
            var target = [exports];
            var publicNS = null;
            if (name) {
                publicNS = createNamespace(_Global, name);
                target.push(publicNS);
            }
            initializeProperties(target, members, name || "<ANONYMOUS>");
            return publicNS;
        }

        // Establish members of the "WinJS.Namespace" namespace
        Object.defineProperties(_rootNamespace.Namespace, {

            defineWithParent: { value: defineWithParent, writable: true, enumerable: true, configurable: true },

            define: { value: define, writable: true, enumerable: true, configurable: true },

            _lazy: { value: lazy, writable: true, enumerable: true, configurable: true },

            _moduleDefine: { value: moduleDefine, writable: true, enumerable: true, configurable: true }

        });

    })();

    (function () {

        function define(constructor, instanceMembers, staticMembers) {
            /// <signature helpKeyword="WinJS.Class.define">
            /// <summary locid="WinJS.Class.define">
            /// Defines a class using the given constructor and the specified instance members.
            /// </summary>
            /// <param name="constructor" type="Function" locid="WinJS.Class.define_p:constructor">
            /// A constructor function that is used to instantiate this class.
            /// </param>
            /// <param name="instanceMembers" type="Object" locid="WinJS.Class.define_p:instanceMembers">
            /// The set of instance fields, properties, and methods made available on the class.
            /// </param>
            /// <param name="staticMembers" type="Object" locid="WinJS.Class.define_p:staticMembers">
            /// The set of static fields, properties, and methods made available on the class.
            /// </param>
            /// <returns type="Function" locid="WinJS.Class.define_returnValue">
            /// The newly-defined class.
            /// </returns>
            /// </signature>
            constructor = constructor || function () { };
            _BaseCoreUtils.markSupportedForProcessing(constructor);
            if (instanceMembers) {
                initializeProperties(constructor.prototype, instanceMembers);
            }
            if (staticMembers) {
                initializeProperties(constructor, staticMembers);
            }
            return constructor;
        }

        function derive(baseClass, constructor, instanceMembers, staticMembers) {
            /// <signature helpKeyword="WinJS.Class.derive">
            /// <summary locid="WinJS.Class.derive">
            /// Creates a sub-class based on the supplied baseClass parameter, using prototypal inheritance.
            /// </summary>
            /// <param name="baseClass" type="Function" locid="WinJS.Class.derive_p:baseClass">
            /// The class to inherit from.
            /// </param>
            /// <param name="constructor" type="Function" locid="WinJS.Class.derive_p:constructor">
            /// A constructor function that is used to instantiate this class.
            /// </param>
            /// <param name="instanceMembers" type="Object" locid="WinJS.Class.derive_p:instanceMembers">
            /// The set of instance fields, properties, and methods to be made available on the class.
            /// </param>
            /// <param name="staticMembers" type="Object" locid="WinJS.Class.derive_p:staticMembers">
            /// The set of static fields, properties, and methods to be made available on the class.
            /// </param>
            /// <returns type="Function" locid="WinJS.Class.derive_returnValue">
            /// The newly-defined class.
            /// </returns>
            /// </signature>
            if (baseClass) {
                constructor = constructor || function () { };
                var basePrototype = baseClass.prototype;
                constructor.prototype = Object.create(basePrototype);
                _BaseCoreUtils.markSupportedForProcessing(constructor);
                Object.defineProperty(constructor.prototype, "constructor", { value: constructor, writable: true, configurable: true, enumerable: true });
                if (instanceMembers) {
                    initializeProperties(constructor.prototype, instanceMembers);
                }
                if (staticMembers) {
                    initializeProperties(constructor, staticMembers);
                }
                return constructor;
            } else {
                return define(constructor, instanceMembers, staticMembers);
            }
        }

        function mix(constructor) {
            /// <signature helpKeyword="WinJS.Class.mix">
            /// <summary locid="WinJS.Class.mix">
            /// Defines a class using the given constructor and the union of the set of instance members
            /// specified by all the mixin objects. The mixin parameter list is of variable length.
            /// </summary>
            /// <param name="constructor" locid="WinJS.Class.mix_p:constructor">
            /// A constructor function that is used to instantiate this class.
            /// </param>
            /// <returns type="Function" locid="WinJS.Class.mix_returnValue">
            /// The newly-defined class.
            /// </returns>
            /// </signature>
            constructor = constructor || function () { };
            var i, len;
            for (i = 1, len = arguments.length; i < len; i++) {
                initializeProperties(constructor.prototype, arguments[i]);
            }
            return constructor;
        }

        // Establish members of "WinJS.Class" namespace
        _WinJS.Namespace.define("WinJS.Class", {
            define: define,
            derive: derive,
            mix: mix
        });

    })();

    return {
        Namespace: _WinJS.Namespace,
        Class: _WinJS.Class
    };

});
_winjs("WinJS/Core/_ErrorFromName", ["WinJS/Core/_Base"], function errorsInit(_Base) {
    "use strict";

    var ErrorFromName = _Base.Class.derive(Error, function (name, message) {
        /// <signature helpKeyword="WinJS.ErrorFromName">
        /// <summary locid="WinJS.ErrorFromName">
        /// Creates an Error object with the specified name and message properties.
        /// </summary>
        /// <param name="name" type="String" locid="WinJS.ErrorFromName_p:name">The name of this error. The name is meant to be consumed programmatically and should not be localized.</param>
        /// <param name="message" type="String" optional="true" locid="WinJS.ErrorFromName_p:message">The message for this error. The message is meant to be consumed by humans and should be localized.</param>
        /// <returns type="Error" locid="WinJS.ErrorFromName_returnValue">Error instance with .name and .message properties populated</returns>
        /// </signature>
        this.name = name;
        this.message = message || name;
    }, {
        /* empty */
    }, {
        supportedForProcessing: false,
    });

    _Base.Namespace.define("WinJS", {
        // ErrorFromName establishes a simple pattern for returning error codes.
        //
        ErrorFromName: ErrorFromName
    });

    return ErrorFromName;

});


_winjs("WinJS/Core/_Events", ["exports","WinJS/Core/_Base"], function eventsInit(exports, _Base) {
    "use strict";


    function createEventProperty(name) {
        var eventPropStateName = "_on" + name + "state";

        return {
            get: function () {
                var state = this[eventPropStateName];
                return state && state.userHandler;
            },
            set: function (handler) {
                var state = this[eventPropStateName];
                if (handler) {
                    if (!state) {
                        state = { wrapper: function (evt) { return state.userHandler(evt); }, userHandler: handler };
                        Object.defineProperty(this, eventPropStateName, { value: state, enumerable: false, writable:true, configurable: true });
                        this.addEventListener(name, state.wrapper, false);
                    }
                    state.userHandler = handler;
                } else if (state) {
                    this.removeEventListener(name, state.wrapper, false);
                    this[eventPropStateName] = null;
                }
            },
            enumerable: true
        };
    }

    function createEventProperties() {
        /// <signature helpKeyword="WinJS.Utilities.createEventProperties">
        /// <summary locid="WinJS.Utilities.createEventProperties">
        /// Creates an object that has one property for each name passed to the function.
        /// </summary>
        /// <param name="events" locid="WinJS.Utilities.createEventProperties_p:events">
        /// A variable list of property names.
        /// </param>
        /// <returns type="Object" locid="WinJS.Utilities.createEventProperties_returnValue">
        /// The object with the specified properties. The names of the properties are prefixed with 'on'.
        /// </returns>
        /// </signature>
        var props = {};
        for (var i = 0, len = arguments.length; i < len; i++) {
            var name = arguments[i];
            props["on" + name] = createEventProperty(name);
        }
        return props;
    }

    var EventMixinEvent = _Base.Class.define(
        function EventMixinEvent_ctor(type, detail, target) {
            this.detail = detail;
            this.target = target;
            this.timeStamp = Date.now();
            this.type = type;
        },
        {
            bubbles: { value: false, writable: false },
            cancelable: { value: false, writable: false },
            currentTarget: {
                get: function () { return this.target; }
            },
            defaultPrevented: {
                get: function () { return this._preventDefaultCalled; }
            },
            trusted: { value: false, writable: false },
            eventPhase: { value: 0, writable: false },
            target: null,
            timeStamp: null,
            type: null,

            preventDefault: function () {
                this._preventDefaultCalled = true;
            },
            stopImmediatePropagation: function () {
                this._stopImmediatePropagationCalled = true;
            },
            stopPropagation: function () {
            }
        }, {
            supportedForProcessing: false,
        }
    );

    var eventMixin = {
        _listeners: null,

        addEventListener: function (type, listener, useCapture) {
            /// <signature helpKeyword="WinJS.Utilities.eventMixin.addEventListener">
            /// <summary locid="WinJS.Utilities.eventMixin.addEventListener">
            /// Adds an event listener to the control.
            /// </summary>
            /// <param name="type" locid="WinJS.Utilities.eventMixin.addEventListener_p:type">
            /// The type (name) of the event.
            /// </param>
            /// <param name="listener" locid="WinJS.Utilities.eventMixin.addEventListener_p:listener">
            /// The listener to invoke when the event is raised.
            /// </param>
            /// <param name="useCapture" locid="WinJS.Utilities.eventMixin.addEventListener_p:useCapture">
            /// if true initiates capture, otherwise false.
            /// </param>
            /// </signature>
            useCapture = useCapture || false;
            this._listeners = this._listeners || {};
            var eventListeners = (this._listeners[type] = this._listeners[type] || []);
            for (var i = 0, len = eventListeners.length; i < len; i++) {
                var l = eventListeners[i];
                if (l.useCapture === useCapture && l.listener === listener) {
                    return;
                }
            }
            eventListeners.push({ listener: listener, useCapture: useCapture });
        },
        dispatchEvent: function (type, details) {
            /// <signature helpKeyword="WinJS.Utilities.eventMixin.dispatchEvent">
            /// <summary locid="WinJS.Utilities.eventMixin.dispatchEvent">
            /// Raises an event of the specified type and with the specified additional properties.
            /// </summary>
            /// <param name="type" locid="WinJS.Utilities.eventMixin.dispatchEvent_p:type">
            /// The type (name) of the event.
            /// </param>
            /// <param name="details" locid="WinJS.Utilities.eventMixin.dispatchEvent_p:details">
            /// The set of additional properties to be attached to the event object when the event is raised.
            /// </param>
            /// <returns type="Boolean" locid="WinJS.Utilities.eventMixin.dispatchEvent_returnValue">
            /// true if preventDefault was called on the event.
            /// </returns>
            /// </signature>
            var listeners = this._listeners && this._listeners[type];
            if (listeners) {
                var eventValue = new EventMixinEvent(type, details, this);
                // Need to copy the array to protect against people unregistering while we are dispatching
                listeners = listeners.slice(0, listeners.length);
                for (var i = 0, len = listeners.length; i < len && !eventValue._stopImmediatePropagationCalled; i++) {
                    listeners[i].listener(eventValue);
                }
                return eventValue.defaultPrevented || false;
            }
            return false;
        },
        removeEventListener: function (type, listener, useCapture) {
            /// <signature helpKeyword="WinJS.Utilities.eventMixin.removeEventListener">
            /// <summary locid="WinJS.Utilities.eventMixin.removeEventListener">
            /// Removes an event listener from the control.
            /// </summary>
            /// <param name="type" locid="WinJS.Utilities.eventMixin.removeEventListener_p:type">
            /// The type (name) of the event.
            /// </param>
            /// <param name="listener" locid="WinJS.Utilities.eventMixin.removeEventListener_p:listener">
            /// The listener to remove.
            /// </param>
            /// <param name="useCapture" locid="WinJS.Utilities.eventMixin.removeEventListener_p:useCapture">
            /// Specifies whether to initiate capture.
            /// </param>
            /// </signature>
            useCapture = useCapture || false;
            var listeners = this._listeners && this._listeners[type];
            if (listeners) {
                for (var i = 0, len = listeners.length; i < len; i++) {
                    var l = listeners[i];
                    if (l.listener === listener && l.useCapture === useCapture) {
                        listeners.splice(i, 1);
                        if (listeners.length === 0) {
                            delete this._listeners[type];
                        }
                        // Only want to remove one element for each call to removeEventListener
                        break;
                    }
                }
            }
        }
    };

    _Base.Namespace._moduleDefine(exports, "WinJS.Utilities", {
        _createEventProperty: createEventProperty,
        createEventProperties: createEventProperties,
        eventMixin: eventMixin
    });

});


_winjs("WinJS/Core/_Trace", ["WinJS/Core/_Global"], function traceInit(_Global) {
    "use strict";

    function nop(v) {
        return v;
    }

    return {
        _traceAsyncOperationStarting: (_Global.Debug && _Global.Debug.msTraceAsyncOperationStarting && _Global.Debug.msTraceAsyncOperationStarting.bind(_Global.Debug)) || nop,
        _traceAsyncOperationCompleted: (_Global.Debug && _Global.Debug.msTraceAsyncOperationCompleted && _Global.Debug.msTraceAsyncOperationCompleted.bind(_Global.Debug)) || nop,
        _traceAsyncCallbackStarting: (_Global.Debug && _Global.Debug.msTraceAsyncCallbackStarting && _Global.Debug.msTraceAsyncCallbackStarting.bind(_Global.Debug)) || nop,
        _traceAsyncCallbackCompleted: (_Global.Debug && _Global.Debug.msTraceAsyncCallbackCompleted && _Global.Debug.msTraceAsyncCallbackCompleted.bind(_Global.Debug)) || nop
    };
});
_winjs("WinJS/Promise/_StateMachine", ["WinJS/Core/_Global","WinJS/Core/_BaseCoreUtils","WinJS/Core/_Base","WinJS/Core/_ErrorFromName","WinJS/Core/_Events","WinJS/Core/_Trace"], function promiseStateMachineInit(_Global, _BaseCoreUtils, _Base, _ErrorFromName, _Events, _Trace) {
    "use strict";

    _Global.Debug && (_Global.Debug.setNonUserCodeExceptions = true);

    var ListenerType = _Base.Class.mix(_Base.Class.define(null, { /*empty*/ }, { supportedForProcessing: false }), _Events.eventMixin);
    var promiseEventListeners = new ListenerType();
    // make sure there is a listeners collection so that we can do a more trivial check below
    promiseEventListeners._listeners = {};
    var errorET = "error";
    var canceledName = "Canceled";
    var tagWithStack = false;
    var tag = {
        promise: 0x01,
        thenPromise: 0x02,
        errorPromise: 0x04,
        exceptionPromise: 0x08,
        completePromise: 0x10,
    };
    tag.all = tag.promise | tag.thenPromise | tag.errorPromise | tag.exceptionPromise | tag.completePromise;

    //
    // Global error counter, for each error which enters the system we increment this once and then
    // the error number travels with the error as it traverses the tree of potential handlers.
    //
    // When someone has registered to be told about errors (WinJS.Promise.callonerror) promises
    // which are in error will get tagged with a ._errorId field. This tagged field is the
    // contract by which nested promises with errors will be identified as chaining for the
    // purposes of the callonerror semantics. If a nested promise in error is encountered without
    // a ._errorId it will be assumed to be foreign and treated as an interop boundary and
    // a new error id will be minted.
    //
    var error_number = 1;

    //
    // The state machine has a interesting hiccup in it with regards to notification, in order
    // to flatten out notification and avoid recursion for synchronous completion we have an
    // explicit set of *_notify states which are responsible for notifying their entire tree
    // of children. They can do this because they know that immediate children are always
    // ThenPromise instances and we can therefore reach into their state to access the
    // _listeners collection.
    //
    // So, what happens is that a Promise will be fulfilled through the _completed or _error
    // messages at which point it will enter a *_notify state and be responsible for to move
    // its children into an (as appropriate) success or error state and also notify that child's
    // listeners of the state transition, until leaf notes are reached.
    //

    var state_created,              // -> working
        state_working,              // -> error | error_notify | success | success_notify | canceled | waiting
        state_waiting,              // -> error | error_notify | success | success_notify | waiting_canceled
        state_waiting_canceled,     // -> error | error_notify | success | success_notify | canceling
        state_canceled,             // -> error | error_notify | success | success_notify | canceling
        state_canceling,            // -> error_notify
        state_success_notify,       // -> success
        state_success,              // -> .
        state_error_notify,         // -> error
        state_error;                // -> .

    // Noop function, used in the various states to indicate that they don't support a given
    // message. Named with the somewhat cute name '_' because it reads really well in the states.

    function _() { }

    // Initial state
    //
    state_created = {
        name: "created",
        enter: function (promise) {
            promise._setState(state_working);
        },
        cancel: _,
        done: _,
        then: _,
        _completed: _,
        _error: _,
        _notify: _,
        _progress: _,
        _setCompleteValue: _,
        _setErrorValue: _
    };

    // Ready state, waiting for a message (completed/error/progress), able to be canceled
    //
    state_working = {
        name: "working",
        enter: _,
        cancel: function (promise) {
            promise._setState(state_canceled);
        },
        done: done,
        then: then,
        _completed: completed,
        _error: error,
        _notify: _,
        _progress: progress,
        _setCompleteValue: setCompleteValue,
        _setErrorValue: setErrorValue
    };

    // Waiting state, if a promise is completed with a value which is itself a promise
    // (has a then() method) it signs up to be informed when that child promise is
    // fulfilled at which point it will be fulfilled with that value.
    //
    state_waiting = {
        name: "waiting",
        enter: function (promise) {
            var waitedUpon = promise._value;
            // We can special case our own intermediate promises which are not in a
            //  terminal state by just pushing this promise as a listener without
            //  having to create new indirection functions
            if (waitedUpon instanceof ThenPromise &&
                waitedUpon._state !== state_error &&
                waitedUpon._state !== state_success) {
                pushListener(waitedUpon, { promise: promise });
            } else {
                var error = function (value) {
                    if (waitedUpon._errorId) {
                        promise._chainedError(value, waitedUpon);
                    } else {
                        // Because this is an interop boundary we want to indicate that this
                        //  error has been handled by the promise infrastructure before we
                        //  begin a new handling chain.
                        //
                        callonerror(promise, value, detailsForHandledError, waitedUpon, error);
                        promise._error(value);
                    }
                };
                error.handlesOnError = true;
                waitedUpon.then(
                    promise._completed.bind(promise),
                    error,
                    promise._progress.bind(promise)
                );
            }
        },
        cancel: function (promise) {
            promise._setState(state_waiting_canceled);
        },
        done: done,
        then: then,
        _completed: completed,
        _error: error,
        _notify: _,
        _progress: progress,
        _setCompleteValue: setCompleteValue,
        _setErrorValue: setErrorValue
    };

    // Waiting canceled state, when a promise has been in a waiting state and receives a
    // request to cancel its pending work it will forward that request to the child promise
    // and then waits to be informed of the result. This promise moves itself into the
    // canceling state but understands that the child promise may instead push it to a
    // different state.
    //
    state_waiting_canceled = {
        name: "waiting_canceled",
        enter: function (promise) {
            // Initiate a transition to canceling. Triggering a cancel on the promise
            // that we are waiting upon may result in a different state transition
            // before the state machine pump runs again.
            promise._setState(state_canceling);
            var waitedUpon = promise._value;
            if (waitedUpon.cancel) {
                waitedUpon.cancel();
            }
        },
        cancel: _,
        done: done,
        then: then,
        _completed: completed,
        _error: error,
        _notify: _,
        _progress: progress,
        _setCompleteValue: setCompleteValue,
        _setErrorValue: setErrorValue
    };

    // Canceled state, moves to the canceling state and then tells the promise to do
    // whatever it might need to do on cancelation.
    //
    state_canceled = {
        name: "canceled",
        enter: function (promise) {
            // Initiate a transition to canceling. The _cancelAction may change the state
            // before the state machine pump runs again.
            promise._setState(state_canceling);
            promise._cancelAction();
        },
        cancel: _,
        done: done,
        then: then,
        _completed: completed,
        _error: error,
        _notify: _,
        _progress: progress,
        _setCompleteValue: setCompleteValue,
        _setErrorValue: setErrorValue
    };

    // Canceling state, commits to the promise moving to an error state with an error
    // object whose 'name' and 'message' properties contain the string "Canceled"
    //
    state_canceling = {
        name: "canceling",
        enter: function (promise) {
            var error = new Error(canceledName);
            error.name = error.message;
            promise._value = error;
            promise._setState(state_error_notify);
        },
        cancel: _,
        done: _,
        then: _,
        _completed: _,
        _error: _,
        _notify: _,
        _progress: _,
        _setCompleteValue: _,
        _setErrorValue: _
    };

    // Success notify state, moves a promise to the success state and notifies all children
    //
    state_success_notify = {
        name: "complete_notify",
        enter: function (promise) {
            promise.done = CompletePromise.prototype.done;
            promise.then = CompletePromise.prototype.then;
            if (promise._listeners) {
                var queue = [promise];
                var p;
                while (queue.length) {
                    p = queue.shift();
                    p._state._notify(p, queue);
                }
            }
            promise._setState(state_success);
        },
        cancel: _,
        done: null, /*error to get here */
        then: null, /*error to get here */
        _completed: _,
        _error: _,
        _notify: notifySuccess,
        _progress: _,
        _setCompleteValue: _,
        _setErrorValue: _
    };

    // Success state, moves a promise to the success state and does NOT notify any children.
    // Some upstream promise is owning the notification pass.
    //
    state_success = {
        name: "success",
        enter: function (promise) {
            promise.done = CompletePromise.prototype.done;
            promise.then = CompletePromise.prototype.then;
            promise._cleanupAction();
        },
        cancel: _,
        done: null, /*error to get here */
        then: null, /*error to get here */
        _completed: _,
        _error: _,
        _notify: notifySuccess,
        _progress: _,
        _setCompleteValue: _,
        _setErrorValue: _
    };

    // Error notify state, moves a promise to the error state and notifies all children
    //
    state_error_notify = {
        name: "error_notify",
        enter: function (promise) {
            promise.done = ErrorPromise.prototype.done;
            promise.then = ErrorPromise.prototype.then;
            if (promise._listeners) {
                var queue = [promise];
                var p;
                while (queue.length) {
                    p = queue.shift();
                    p._state._notify(p, queue);
                }
            }
            promise._setState(state_error);
        },
        cancel: _,
        done: null, /*error to get here*/
        then: null, /*error to get here*/
        _completed: _,
        _error: _,
        _notify: notifyError,
        _progress: _,
        _setCompleteValue: _,
        _setErrorValue: _
    };

    // Error state, moves a promise to the error state and does NOT notify any children.
    // Some upstream promise is owning the notification pass.
    //
    state_error = {
        name: "error",
        enter: function (promise) {
            promise.done = ErrorPromise.prototype.done;
            promise.then = ErrorPromise.prototype.then;
            promise._cleanupAction();
        },
        cancel: _,
        done: null, /*error to get here*/
        then: null, /*error to get here*/
        _completed: _,
        _error: _,
        _notify: notifyError,
        _progress: _,
        _setCompleteValue: _,
        _setErrorValue: _
    };

    //
    // The statemachine implementation follows a very particular pattern, the states are specified
    // as static stateless bags of functions which are then indirected through the state machine
    // instance (a Promise). As such all of the functions on each state have the promise instance
    // passed to them explicitly as a parameter and the Promise instance members do a little
    // dance where they indirect through the state and insert themselves in the argument list.
    //
    // We could instead call directly through the promise states however then every caller
    // would have to remember to do things like pumping the state machine to catch state transitions.
    //

    var PromiseStateMachine = _Base.Class.define(null, {
        _listeners: null,
        _nextState: null,
        _state: null,
        _value: null,

        cancel: function () {
            /// <signature helpKeyword="WinJS.PromiseStateMachine.cancel">
            /// <summary locid="WinJS.PromiseStateMachine.cancel">
            /// Attempts to cancel the fulfillment of a promised value. If the promise hasn't
            /// already been fulfilled and cancellation is supported, the promise enters
            /// the error state with a value of Error("Canceled").
            /// </summary>
            /// </signature>
            this._state.cancel(this);
            this._run();
        },
        done: function Promise_done(onComplete, onError, onProgress) {
            /// <signature helpKeyword="WinJS.PromiseStateMachine.done">
            /// <summary locid="WinJS.PromiseStateMachine.done">
            /// Allows you to specify the work to be done on the fulfillment of the promised value,
            /// the error handling to be performed if the promise fails to fulfill
            /// a value, and the handling of progress notifications along the way.
            ///
            /// After the handlers have finished executing, this function throws any error that would have been returned
            /// from then() as a promise in the error state.
            /// </summary>
            /// <param name='onComplete' type='Function' locid="WinJS.PromiseStateMachine.done_p:onComplete">
            /// The function to be called if the promise is fulfilled successfully with a value.
            /// The fulfilled value is passed as the single argument. If the value is null,
            /// the fulfilled value is returned. The value returned
            /// from the function becomes the fulfilled value of the promise returned by
            /// then(). If an exception is thrown while executing the function, the promise returned
            /// by then() moves into the error state.
            /// </param>
            /// <param name='onError' type='Function' optional='true' locid="WinJS.PromiseStateMachine.done_p:onError">
            /// The function to be called if the promise is fulfilled with an error. The error
            /// is passed as the single argument. If it is null, the error is forwarded.
            /// The value returned from the function is the fulfilled value of the promise returned by then().
            /// </param>
            /// <param name='onProgress' type='Function' optional='true' locid="WinJS.PromiseStateMachine.done_p:onProgress">
            /// the function to be called if the promise reports progress. Data about the progress
            /// is passed as the single argument. Promises are not required to support
            /// progress.
            /// </param>
            /// </signature>
            this._state.done(this, onComplete, onError, onProgress);
        },
        then: function Promise_then(onComplete, onError, onProgress) {
            /// <signature helpKeyword="WinJS.PromiseStateMachine.then">
            /// <summary locid="WinJS.PromiseStateMachine.then">
            /// Allows you to specify the work to be done on the fulfillment of the promised value,
            /// the error handling to be performed if the promise fails to fulfill
            /// a value, and the handling of progress notifications along the way.
            /// </summary>
            /// <param name='onComplete' type='Function' locid="WinJS.PromiseStateMachine.then_p:onComplete">
            /// The function to be called if the promise is fulfilled successfully with a value.
            /// The value is passed as the single argument. If the value is null, the value is returned.
            /// The value returned from the function becomes the fulfilled value of the promise returned by
            /// then(). If an exception is thrown while this function is being executed, the promise returned
            /// by then() moves into the error state.
            /// </param>
            /// <param name='onError' type='Function' optional='true' locid="WinJS.PromiseStateMachine.then_p:onError">
            /// The function to be called if the promise is fulfilled with an error. The error
            /// is passed as the single argument. If it is null, the error is forwarded.
            /// The value returned from the function becomes the fulfilled value of the promise returned by then().
            /// </param>
            /// <param name='onProgress' type='Function' optional='true' locid="WinJS.PromiseStateMachine.then_p:onProgress">
            /// The function to be called if the promise reports progress. Data about the progress
            /// is passed as the single argument. Promises are not required to support
            /// progress.
            /// </param>
            /// <returns type="WinJS.Promise" locid="WinJS.PromiseStateMachine.then_returnValue">
            /// The promise whose value is the result of executing the complete or
            /// error function.
            /// </returns>
            /// </signature>
            return this._state.then(this, onComplete, onError, onProgress);
        },

        _chainedError: function (value, context) {
            var result = this._state._error(this, value, detailsForChainedError, context);
            this._run();
            return result;
        },
        _completed: function (value) {
            var result = this._state._completed(this, value);
            this._run();
            return result;
        },
        _error: function (value) {
            var result = this._state._error(this, value, detailsForError);
            this._run();
            return result;
        },
        _progress: function (value) {
            this._state._progress(this, value);
        },
        _setState: function (state) {
            this._nextState = state;
        },
        _setCompleteValue: function (value) {
            this._state._setCompleteValue(this, value);
            this._run();
        },
        _setChainedErrorValue: function (value, context) {
            var result = this._state._setErrorValue(this, value, detailsForChainedError, context);
            this._run();
            return result;
        },
        _setExceptionValue: function (value) {
            var result = this._state._setErrorValue(this, value, detailsForException);
            this._run();
            return result;
        },
        _run: function () {
            while (this._nextState) {
                this._state = this._nextState;
                this._nextState = null;
                this._state.enter(this);
            }
        }
    }, {
        supportedForProcessing: false
    });

    //
    // Implementations of shared state machine code.
    //

    function completed(promise, value) {
        var targetState;
        if (value && typeof value === "object" && typeof value.then === "function") {
            targetState = state_waiting;
        } else {
            targetState = state_success_notify;
        }
        promise._value = value;
        promise._setState(targetState);
    }
    function createErrorDetails(exception, error, promise, id, parent, handler) {
        return {
            exception: exception,
            error: error,
            promise: promise,
            handler: handler,
            id: id,
            parent: parent
        };
    }
    function detailsForHandledError(promise, errorValue, context, handler) {
        var exception = context._isException;
        var errorId = context._errorId;
        return createErrorDetails(
            exception ? errorValue : null,
            exception ? null : errorValue,
            promise,
            errorId,
            context,
            handler
        );
    }
    function detailsForChainedError(promise, errorValue, context) {
        var exception = context._isException;
        var errorId = context._errorId;
        setErrorInfo(promise, errorId, exception);
        return createErrorDetails(
            exception ? errorValue : null,
            exception ? null : errorValue,
            promise,
            errorId,
            context
        );
    }
    function detailsForError(promise, errorValue) {
        var errorId = ++error_number;
        setErrorInfo(promise, errorId);
        return createErrorDetails(
            null,
            errorValue,
            promise,
            errorId
        );
    }
    function detailsForException(promise, exceptionValue) {
        var errorId = ++error_number;
        setErrorInfo(promise, errorId, true);
        return createErrorDetails(
            exceptionValue,
            null,
            promise,
            errorId
        );
    }
    function done(promise, onComplete, onError, onProgress) {
        var asyncOpID = _Trace._traceAsyncOperationStarting("WinJS.Promise.done");
        pushListener(promise, { c: onComplete, e: onError, p: onProgress, asyncOpID: asyncOpID });
    }
    function error(promise, value, onerrorDetails, context) {
        promise._value = value;
        callonerror(promise, value, onerrorDetails, context);
        promise._setState(state_error_notify);
    }
    function notifySuccess(promise, queue) {
        var value = promise._value;
        var listeners = promise._listeners;
        if (!listeners) {
            return;
        }
        promise._listeners = null;
        var i, len;
        for (i = 0, len = Array.isArray(listeners) ? listeners.length : 1; i < len; i++) {
            var listener = len === 1 ? listeners : listeners[i];
            var onComplete = listener.c;
            var target = listener.promise;

            _Trace._traceAsyncOperationCompleted(listener.asyncOpID, _Global.Debug && _Global.Debug.MS_ASYNC_OP_STATUS_SUCCESS);

            if (target) {
                _Trace._traceAsyncCallbackStarting(listener.asyncOpID);
                try {
                    target._setCompleteValue(onComplete ? onComplete(value) : value);
                } catch (ex) {
                    target._setExceptionValue(ex);
                } finally {
                    _Trace._traceAsyncCallbackCompleted();
                }
                if (target._state !== state_waiting && target._listeners) {
                    queue.push(target);
                }
            } else {
                CompletePromise.prototype.done.call(promise, onComplete);
            }
        }
    }
    function notifyError(promise, queue) {
        var value = promise._value;
        var listeners = promise._listeners;
        if (!listeners) {
            return;
        }
        promise._listeners = null;
        var i, len;
        for (i = 0, len = Array.isArray(listeners) ? listeners.length : 1; i < len; i++) {
            var listener = len === 1 ? listeners : listeners[i];
            var onError = listener.e;
            var target = listener.promise;

            var errorID = _Global.Debug && (value && value.name === canceledName ? _Global.Debug.MS_ASYNC_OP_STATUS_CANCELED : _Global.Debug.MS_ASYNC_OP_STATUS_ERROR);
            _Trace._traceAsyncOperationCompleted(listener.asyncOpID, errorID);

            if (target) {
                var asyncCallbackStarted = false;
                try {
                    if (onError) {
                        _Trace._traceAsyncCallbackStarting(listener.asyncOpID);
                        asyncCallbackStarted = true;
                        if (!onError.handlesOnError) {
                            callonerror(target, value, detailsForHandledError, promise, onError);
                        }
                        target._setCompleteValue(onError(value));
                    } else {
                        target._setChainedErrorValue(value, promise);
                    }
                } catch (ex) {
                    target._setExceptionValue(ex);
                } finally {
                    if (asyncCallbackStarted) {
                        _Trace._traceAsyncCallbackCompleted();
                    }
                }
                if (target._state !== state_waiting && target._listeners) {
                    queue.push(target);
                }
            } else {
                ErrorPromise.prototype.done.call(promise, null, onError);
            }
        }
    }
    function callonerror(promise, value, onerrorDetailsGenerator, context, handler) {
        if (promiseEventListeners._listeners[errorET]) {
            if (value instanceof Error && value.message === canceledName) {
                return;
            }
            promiseEventListeners.dispatchEvent(errorET, onerrorDetailsGenerator(promise, value, context, handler));
        }
    }
    function progress(promise, value) {
        var listeners = promise._listeners;
        if (listeners) {
            var i, len;
            for (i = 0, len = Array.isArray(listeners) ? listeners.length : 1; i < len; i++) {
                var listener = len === 1 ? listeners : listeners[i];
                var onProgress = listener.p;
                if (onProgress) {
                    try { onProgress(value); } catch (ex) { }
                }
                if (!(listener.c || listener.e) && listener.promise) {
                    listener.promise._progress(value);
                }
            }
        }
    }
    function pushListener(promise, listener) {
        var listeners = promise._listeners;
        if (listeners) {
            // We may have either a single listener (which will never be wrapped in an array)
            // or 2+ listeners (which will be wrapped). Since we are now adding one more listener
            // we may have to wrap the single listener before adding the second.
            listeners = Array.isArray(listeners) ? listeners : [listeners];
            listeners.push(listener);
        } else {
            listeners = listener;
        }
        promise._listeners = listeners;
    }
    // The difference beween setCompleteValue()/setErrorValue() and complete()/error() is that setXXXValue() moves
    // a promise directly to the success/error state without starting another notification pass (because one
    // is already ongoing).
    function setErrorInfo(promise, errorId, isException) {
        promise._isException = isException || false;
        promise._errorId = errorId;
    }
    function setErrorValue(promise, value, onerrorDetails, context) {
        promise._value = value;
        callonerror(promise, value, onerrorDetails, context);
        promise._setState(state_error);
    }
    function setCompleteValue(promise, value) {
        var targetState;
        if (value && typeof value === "object" && typeof value.then === "function") {
            targetState = state_waiting;
        } else {
            targetState = state_success;
        }
        promise._value = value;
        promise._setState(targetState);
    }
    function then(promise, onComplete, onError, onProgress) {
        var result = new ThenPromise(promise);
        var asyncOpID = _Trace._traceAsyncOperationStarting("WinJS.Promise.then");
        pushListener(promise, { promise: result, c: onComplete, e: onError, p: onProgress, asyncOpID: asyncOpID });
        return result;
    }

    //
    // Internal implementation detail promise, ThenPromise is created when a promise needs
    // to be returned from a then() method.
    //
    var ThenPromise = _Base.Class.derive(PromiseStateMachine,
        function (creator) {

            if (tagWithStack && (tagWithStack === true || (tagWithStack & tag.thenPromise))) {
                this._stack = Promise._getStack();
            }

            this._creator = creator;
            this._setState(state_created);
            this._run();
        }, {
            _creator: null,

            _cancelAction: function () { if (this._creator) { this._creator.cancel(); } },
            _cleanupAction: function () { this._creator = null; }
        }, {
            supportedForProcessing: false
        }
    );

    //
    // Slim promise implementations for already completed promises, these are created
    // under the hood on synchronous completion paths as well as by WinJS.Promise.wrap
    // and WinJS.Promise.wrapError.
    //

    var ErrorPromise = _Base.Class.define(
        function ErrorPromise_ctor(value) {

            if (tagWithStack && (tagWithStack === true || (tagWithStack & tag.errorPromise))) {
                this._stack = Promise._getStack();
            }

            this._value = value;
            callonerror(this, value, detailsForError);
        }, {
            cancel: function () {
                /// <signature helpKeyword="WinJS.PromiseStateMachine.cancel">
                /// <summary locid="WinJS.PromiseStateMachine.cancel">
                /// Attempts to cancel the fulfillment of a promised value. If the promise hasn't
                /// already been fulfilled and cancellation is supported, the promise enters
                /// the error state with a value of Error("Canceled").
                /// </summary>
                /// </signature>
            },
            done: function ErrorPromise_done(unused, onError) {
                /// <signature helpKeyword="WinJS.PromiseStateMachine.done">
                /// <summary locid="WinJS.PromiseStateMachine.done">
                /// Allows you to specify the work to be done on the fulfillment of the promised value,
                /// the error handling to be performed if the promise fails to fulfill
                /// a value, and the handling of progress notifications along the way.
                ///
                /// After the handlers have finished executing, this function throws any error that would have been returned
                /// from then() as a promise in the error state.
                /// </summary>
                /// <param name='onComplete' type='Function' locid="WinJS.PromiseStateMachine.done_p:onComplete">
                /// The function to be called if the promise is fulfilled successfully with a value.
                /// The fulfilled value is passed as the single argument. If the value is null,
                /// the fulfilled value is returned. The value returned
                /// from the function becomes the fulfilled value of the promise returned by
                /// then(). If an exception is thrown while executing the function, the promise returned
                /// by then() moves into the error state.
                /// </param>
                /// <param name='onError' type='Function' optional='true' locid="WinJS.PromiseStateMachine.done_p:onError">
                /// The function to be called if the promise is fulfilled with an error. The error
                /// is passed as the single argument. If it is null, the error is forwarded.
                /// The value returned from the function is the fulfilled value of the promise returned by then().
                /// </param>
                /// <param name='onProgress' type='Function' optional='true' locid="WinJS.PromiseStateMachine.done_p:onProgress">
                /// the function to be called if the promise reports progress. Data about the progress
                /// is passed as the single argument. Promises are not required to support
                /// progress.
                /// </param>
                /// </signature>
                var value = this._value;
                if (onError) {
                    try {
                        if (!onError.handlesOnError) {
                            callonerror(null, value, detailsForHandledError, this, onError);
                        }
                        var result = onError(value);
                        if (result && typeof result === "object" && typeof result.done === "function") {
                            // If a promise is returned we need to wait on it.
                            result.done();
                        }
                        return;
                    } catch (ex) {
                        value = ex;
                    }
                }
                if (value instanceof Error && value.message === canceledName) {
                    // suppress cancel
                    return;
                }
                // force the exception to be thrown asyncronously to avoid any try/catch blocks
                //
                Promise._doneHandler(value);
            },
            then: function ErrorPromise_then(unused, onError) {
                /// <signature helpKeyword="WinJS.PromiseStateMachine.then">
                /// <summary locid="WinJS.PromiseStateMachine.then">
                /// Allows you to specify the work to be done on the fulfillment of the promised value,
                /// the error handling to be performed if the promise fails to fulfill
                /// a value, and the handling of progress notifications along the way.
                /// </summary>
                /// <param name='onComplete' type='Function' locid="WinJS.PromiseStateMachine.then_p:onComplete">
                /// The function to be called if the promise is fulfilled successfully with a value.
                /// The value is passed as the single argument. If the value is null, the value is returned.
                /// The value returned from the function becomes the fulfilled value of the promise returned by
                /// then(). If an exception is thrown while this function is being executed, the promise returned
                /// by then() moves into the error state.
                /// </param>
                /// <param name='onError' type='Function' optional='true' locid="WinJS.PromiseStateMachine.then_p:onError">
                /// The function to be called if the promise is fulfilled with an error. The error
                /// is passed as the single argument. If it is null, the error is forwarded.
                /// The value returned from the function becomes the fulfilled value of the promise returned by then().
                /// </param>
                /// <param name='onProgress' type='Function' optional='true' locid="WinJS.PromiseStateMachine.then_p:onProgress">
                /// The function to be called if the promise reports progress. Data about the progress
                /// is passed as the single argument. Promises are not required to support
                /// progress.
                /// </param>
                /// <returns type="WinJS.Promise" locid="WinJS.PromiseStateMachine.then_returnValue">
                /// The promise whose value is the result of executing the complete or
                /// error function.
                /// </returns>
                /// </signature>

                // If the promise is already in a error state and no error handler is provided
                // we optimize by simply returning the promise instead of creating a new one.
                //
                if (!onError) { return this; }
                var result;
                var value = this._value;
                try {
                    if (!onError.handlesOnError) {
                        callonerror(null, value, detailsForHandledError, this, onError);
                    }
                    result = new CompletePromise(onError(value));
                } catch (ex) {
                    // If the value throw from the error handler is the same as the value
                    // provided to the error handler then there is no need for a new promise.
                    //
                    if (ex === value) {
                        result = this;
                    } else {
                        result = new ExceptionPromise(ex);
                    }
                }
                return result;
            }
        }, {
            supportedForProcessing: false
        }
    );

    var ExceptionPromise = _Base.Class.derive(ErrorPromise,
        function ExceptionPromise_ctor(value) {

            if (tagWithStack && (tagWithStack === true || (tagWithStack & tag.exceptionPromise))) {
                this._stack = Promise._getStack();
            }

            this._value = value;
            callonerror(this, value, detailsForException);
        }, {
            /* empty */
        }, {
            supportedForProcessing: false
        }
    );

    var CompletePromise = _Base.Class.define(
        function CompletePromise_ctor(value) {

            if (tagWithStack && (tagWithStack === true || (tagWithStack & tag.completePromise))) {
                this._stack = Promise._getStack();
            }

            if (value && typeof value === "object" && typeof value.then === "function") {
                var result = new ThenPromise(null);
                result._setCompleteValue(value);
                return result;
            }
            this._value = value;
        }, {
            cancel: function () {
                /// <signature helpKeyword="WinJS.PromiseStateMachine.cancel">
                /// <summary locid="WinJS.PromiseStateMachine.cancel">
                /// Attempts to cancel the fulfillment of a promised value. If the promise hasn't
                /// already been fulfilled and cancellation is supported, the promise enters
                /// the error state with a value of Error("Canceled").
                /// </summary>
                /// </signature>
            },
            done: function CompletePromise_done(onComplete) {
                /// <signature helpKeyword="WinJS.PromiseStateMachine.done">
                /// <summary locid="WinJS.PromiseStateMachine.done">
                /// Allows you to specify the work to be done on the fulfillment of the promised value,
                /// the error handling to be performed if the promise fails to fulfill
                /// a value, and the handling of progress notifications along the way.
                ///
                /// After the handlers have finished executing, this function throws any error that would have been returned
                /// from then() as a promise in the error state.
                /// </summary>
                /// <param name='onComplete' type='Function' locid="WinJS.PromiseStateMachine.done_p:onComplete">
                /// The function to be called if the promise is fulfilled successfully with a value.
                /// The fulfilled value is passed as the single argument. If the value is null,
                /// the fulfilled value is returned. The value returned
                /// from the function becomes the fulfilled value of the promise returned by
                /// then(). If an exception is thrown while executing the function, the promise returned
                /// by then() moves into the error state.
                /// </param>
                /// <param name='onError' type='Function' optional='true' locid="WinJS.PromiseStateMachine.done_p:onError">
                /// The function to be called if the promise is fulfilled with an error. The error
                /// is passed as the single argument. If it is null, the error is forwarded.
                /// The value returned from the function is the fulfilled value of the promise returned by then().
                /// </param>
                /// <param name='onProgress' type='Function' optional='true' locid="WinJS.PromiseStateMachine.done_p:onProgress">
                /// the function to be called if the promise reports progress. Data about the progress
                /// is passed as the single argument. Promises are not required to support
                /// progress.
                /// </param>
                /// </signature>
                if (!onComplete) { return; }
                try {
                    var result = onComplete(this._value);
                    if (result && typeof result === "object" && typeof result.done === "function") {
                        result.done();
                    }
                } catch (ex) {
                    // force the exception to be thrown asynchronously to avoid any try/catch blocks
                    Promise._doneHandler(ex);
                }
            },
            then: function CompletePromise_then(onComplete) {
                /// <signature helpKeyword="WinJS.PromiseStateMachine.then">
                /// <summary locid="WinJS.PromiseStateMachine.then">
                /// Allows you to specify the work to be done on the fulfillment of the promised value,
                /// the error handling to be performed if the promise fails to fulfill
                /// a value, and the handling of progress notifications along the way.
                /// </summary>
                /// <param name='onComplete' type='Function' locid="WinJS.PromiseStateMachine.then_p:onComplete">
                /// The function to be called if the promise is fulfilled successfully with a value.
                /// The value is passed as the single argument. If the value is null, the value is returned.
                /// The value returned from the function becomes the fulfilled value of the promise returned by
                /// then(). If an exception is thrown while this function is being executed, the promise returned
                /// by then() moves into the error state.
                /// </param>
                /// <param name='onError' type='Function' optional='true' locid="WinJS.PromiseStateMachine.then_p:onError">
                /// The function to be called if the promise is fulfilled with an error. The error
                /// is passed as the single argument. If it is null, the error is forwarded.
                /// The value returned from the function becomes the fulfilled value of the promise returned by then().
                /// </param>
                /// <param name='onProgress' type='Function' optional='true' locid="WinJS.PromiseStateMachine.then_p:onProgress">
                /// The function to be called if the promise reports progress. Data about the progress
                /// is passed as the single argument. Promises are not required to support
                /// progress.
                /// </param>
                /// <returns type="WinJS.Promise" locid="WinJS.PromiseStateMachine.then_returnValue">
                /// The promise whose value is the result of executing the complete or
                /// error function.
                /// </returns>
                /// </signature>
                try {
                    // If the value returned from the completion handler is the same as the value
                    // provided to the completion handler then there is no need for a new promise.
                    //
                    var newValue = onComplete ? onComplete(this._value) : this._value;
                    return newValue === this._value ? this : new CompletePromise(newValue);
                } catch (ex) {
                    return new ExceptionPromise(ex);
                }
            }
        }, {
            supportedForProcessing: false
        }
    );

    //
    // Promise is the user-creatable WinJS.Promise object.
    //

    function timeout(timeoutMS) {
        var id;
        return new Promise(
            function (c) {
                if (timeoutMS) {
                    id = _Global.setTimeout(c, timeoutMS);
                } else {
                    _BaseCoreUtils._setImmediate(c);
                }
            },
            function () {
                if (id) {
                    _Global.clearTimeout(id);
                }
            }
        );
    }

    function timeoutWithPromise(timeout, promise) {
        var cancelPromise = function () { promise.cancel(); };
        var cancelTimeout = function () { timeout.cancel(); };
        timeout.then(cancelPromise);
        promise.then(cancelTimeout, cancelTimeout);
        return promise;
    }

    var staticCanceledPromise;

    var Promise = _Base.Class.derive(PromiseStateMachine,
        function Promise_ctor(init, oncancel) {
            /// <signature helpKeyword="WinJS.Promise">
            /// <summary locid="WinJS.Promise">
            /// A promise provides a mechanism to schedule work to be done on a value that
            /// has not yet been computed. It is a convenient abstraction for managing
            /// interactions with asynchronous APIs.
            /// </summary>
            /// <param name="init" type="Function" locid="WinJS.Promise_p:init">
            /// The function that is called during construction of the  promise. The function
            /// is given three arguments (complete, error, progress). Inside this function
            /// you should add event listeners for the notifications supported by this value.
            /// </param>
            /// <param name="oncancel" optional="true" locid="WinJS.Promise_p:oncancel">
            /// The function to call if a consumer of this promise wants
            /// to cancel its undone work. Promises are not required to
            /// support cancellation.
            /// </param>
            /// </signature>

            if (tagWithStack && (tagWithStack === true || (tagWithStack & tag.promise))) {
                this._stack = Promise._getStack();
            }

            this._oncancel = oncancel;
            this._setState(state_created);
            this._run();

            try {
                var complete = this._completed.bind(this);
                var error = this._error.bind(this);
                var progress = this._progress.bind(this);
                init(complete, error, progress);
            } catch (ex) {
                this._setExceptionValue(ex);
            }
        }, {
            _oncancel: null,

            _cancelAction: function () {
                // BEGIN monaco change
                try {
                    if (this._oncancel) {
                        this._oncancel();
                    } else {
                        throw new Error('Promise did not implement oncancel');
                    }
                } catch (ex) {
                    // Access fields to get them created
                    var msg = ex.message;
                    var stack = ex.stack;
                    promiseEventListeners.dispatchEvent('error', ex);
                }
                // END monaco change
            },
            _cleanupAction: function () { this._oncancel = null; }
        }, {

            addEventListener: function Promise_addEventListener(eventType, listener, capture) {
                /// <signature helpKeyword="WinJS.Promise.addEventListener">
                /// <summary locid="WinJS.Promise.addEventListener">
                /// Adds an event listener to the control.
                /// </summary>
                /// <param name="eventType" locid="WinJS.Promise.addEventListener_p:eventType">
                /// The type (name) of the event.
                /// </param>
                /// <param name="listener" locid="WinJS.Promise.addEventListener_p:listener">
                /// The listener to invoke when the event is raised.
                /// </param>
                /// <param name="capture" locid="WinJS.Promise.addEventListener_p:capture">
                /// Specifies whether or not to initiate capture.
                /// </param>
                /// </signature>
                promiseEventListeners.addEventListener(eventType, listener, capture);
            },
            any: function Promise_any(values) {
                /// <signature helpKeyword="WinJS.Promise.any">
                /// <summary locid="WinJS.Promise.any">
                /// Returns a promise that is fulfilled when one of the input promises
                /// has been fulfilled.
                /// </summary>
                /// <param name="values" type="Array" locid="WinJS.Promise.any_p:values">
                /// An array that contains promise objects or objects whose property
                /// values include promise objects.
                /// </param>
                /// <returns type="WinJS.Promise" locid="WinJS.Promise.any_returnValue">
                /// A promise that on fulfillment yields the value of the input (complete or error).
                /// </returns>
                /// </signature>
                return new Promise(
                    function (complete, error) {
                        var keys = Object.keys(values);
                        if (keys.length === 0) {
                            complete();
                        }
                        var canceled = 0;
                        keys.forEach(function (key) {
                            Promise.as(values[key]).then(
                                function () { complete({ key: key, value: values[key] }); },
                                function (e) {
                                    if (e instanceof Error && e.name === canceledName) {
                                        if ((++canceled) === keys.length) {
                                            complete(Promise.cancel);
                                        }
                                        return;
                                    }
                                    error({ key: key, value: values[key] });
                                }
                            );
                        });
                    },
                    function () {
                        var keys = Object.keys(values);
                        keys.forEach(function (key) {
                            var promise = Promise.as(values[key]);
                            if (typeof promise.cancel === "function") {
                                promise.cancel();
                            }
                        });
                    }
                );
            },
            as: function Promise_as(value) {
                /// <signature helpKeyword="WinJS.Promise.as">
                /// <summary locid="WinJS.Promise.as">
                /// Returns a promise. If the object is already a promise it is returned;
                /// otherwise the object is wrapped in a promise.
                /// </summary>
                /// <param name="value" locid="WinJS.Promise.as_p:value">
                /// The value to be treated as a promise.
                /// </param>
                /// <returns type="WinJS.Promise" locid="WinJS.Promise.as_returnValue">
                /// A promise.
                /// </returns>
                /// </signature>
                if (value && typeof value === "object" && typeof value.then === "function") {
                    return value;
                }
                return new CompletePromise(value);
            },
            /// <field type="WinJS.Promise" helpKeyword="WinJS.Promise.cancel" locid="WinJS.Promise.cancel">
            /// Canceled promise value, can be returned from a promise completion handler
            /// to indicate cancelation of the promise chain.
            /// </field>
            cancel: {
                get: function () {
                    return (staticCanceledPromise = staticCanceledPromise || new ErrorPromise(new _ErrorFromName(canceledName)));
                }
            },
            dispatchEvent: function Promise_dispatchEvent(eventType, details) {
                /// <signature helpKeyword="WinJS.Promise.dispatchEvent">
                /// <summary locid="WinJS.Promise.dispatchEvent">
                /// Raises an event of the specified type and properties.
                /// </summary>
                /// <param name="eventType" locid="WinJS.Promise.dispatchEvent_p:eventType">
                /// The type (name) of the event.
                /// </param>
                /// <param name="details" locid="WinJS.Promise.dispatchEvent_p:details">
                /// The set of additional properties to be attached to the event object.
                /// </param>
                /// <returns type="Boolean" locid="WinJS.Promise.dispatchEvent_returnValue">
                /// Specifies whether preventDefault was called on the event.
                /// </returns>
                /// </signature>
                return promiseEventListeners.dispatchEvent(eventType, details);
            },
            is: function Promise_is(value) {
                /// <signature helpKeyword="WinJS.Promise.is">
                /// <summary locid="WinJS.Promise.is">
                /// Determines whether a value fulfills the promise contract.
                /// </summary>
                /// <param name="value" locid="WinJS.Promise.is_p:value">
                /// A value that may be a promise.
                /// </param>
                /// <returns type="Boolean" locid="WinJS.Promise.is_returnValue">
                /// true if the specified value is a promise, otherwise false.
                /// </returns>
                /// </signature>
                return value && typeof value === "object" && typeof value.then === "function";
            },
            join: function Promise_join(values) {
                /// <signature helpKeyword="WinJS.Promise.join">
                /// <summary locid="WinJS.Promise.join">
                /// Creates a promise that is fulfilled when all the values are fulfilled.
                /// </summary>
                /// <param name="values" type="Object" locid="WinJS.Promise.join_p:values">
                /// An object whose fields contain values, some of which may be promises.
                /// </param>
                /// <returns type="WinJS.Promise" locid="WinJS.Promise.join_returnValue">
                /// A promise whose value is an object with the same field names as those of the object in the values parameter, where
                /// each field value is the fulfilled value of a promise.
                /// </returns>
                /// </signature>
                return new Promise(
                    function (complete, error, progress) {
                        var keys = Object.keys(values);
                        var errors = Array.isArray(values) ? [] : {};
                        var results = Array.isArray(values) ? [] : {};
                        var undefineds = 0;
                        var pending = keys.length;
                        var argDone = function (key) {
                            if ((--pending) === 0) {
                                var errorCount = Object.keys(errors).length;
                                if (errorCount === 0) {
                                    complete(results);
                                } else {
                                    var canceledCount = 0;
                                    keys.forEach(function (key) {
                                        var e = errors[key];
                                        if (e instanceof Error && e.name === canceledName) {
                                            canceledCount++;
                                        }
                                    });
                                    if (canceledCount === errorCount) {
                                        complete(Promise.cancel);
                                    } else {
                                        error(errors);
                                    }
                                }
                            } else {
                                progress({ Key: key, Done: true });
                            }
                        };
                        keys.forEach(function (key) {
                            var value = values[key];
                            if (value === undefined) {
                                undefineds++;
                            } else {
                                Promise.then(value,
                                    function (value) { results[key] = value; argDone(key); },
                                    function (value) { errors[key] = value; argDone(key); }
                                );
                            }
                        });
                        pending -= undefineds;
                        if (pending === 0) {
                            complete(results);
                            return;
                        }
                    },
                    function () {
                        Object.keys(values).forEach(function (key) {
                            var promise = Promise.as(values[key]);
                            if (typeof promise.cancel === "function") {
                                promise.cancel();
                            }
                        });
                    }
                );
            },
            removeEventListener: function Promise_removeEventListener(eventType, listener, capture) {
                /// <signature helpKeyword="WinJS.Promise.removeEventListener">
                /// <summary locid="WinJS.Promise.removeEventListener">
                /// Removes an event listener from the control.
                /// </summary>
                /// <param name='eventType' locid="WinJS.Promise.removeEventListener_eventType">
                /// The type (name) of the event.
                /// </param>
                /// <param name='listener' locid="WinJS.Promise.removeEventListener_listener">
                /// The listener to remove.
                /// </param>
                /// <param name='capture' locid="WinJS.Promise.removeEventListener_capture">
                /// Specifies whether or not to initiate capture.
                /// </param>
                /// </signature>
                promiseEventListeners.removeEventListener(eventType, listener, capture);
            },
            supportedForProcessing: false,
            then: function Promise_then(value, onComplete, onError, onProgress) {
                /// <signature helpKeyword="WinJS.Promise.then">
                /// <summary locid="WinJS.Promise.then">
                /// A static version of the promise instance method then().
                /// </summary>
                /// <param name="value" locid="WinJS.Promise.then_p:value">
                /// the value to be treated as a promise.
                /// </param>
                /// <param name="onComplete" type="Function" locid="WinJS.Promise.then_p:complete">
                /// The function to be called if the promise is fulfilled with a value.
                /// If it is null, the promise simply
                /// returns the value. The value is passed as the single argument.
                /// </param>
                /// <param name="onError" type="Function" optional="true" locid="WinJS.Promise.then_p:error">
                /// The function to be called if the promise is fulfilled with an error. The error
                /// is passed as the single argument.
                /// </param>
                /// <param name="onProgress" type="Function" optional="true" locid="WinJS.Promise.then_p:progress">
                /// The function to be called if the promise reports progress. Data about the progress
                /// is passed as the single argument. Promises are not required to support
                /// progress.
                /// </param>
                /// <returns type="WinJS.Promise" locid="WinJS.Promise.then_returnValue">
                /// A promise whose value is the result of executing the provided complete function.
                /// </returns>
                /// </signature>
                return Promise.as(value).then(onComplete, onError, onProgress);
            },
            thenEach: function Promise_thenEach(values, onComplete, onError, onProgress) {
                /// <signature helpKeyword="WinJS.Promise.thenEach">
                /// <summary locid="WinJS.Promise.thenEach">
                /// Performs an operation on all the input promises and returns a promise
                /// that has the shape of the input and contains the result of the operation
                /// that has been performed on each input.
                /// </summary>
                /// <param name="values" locid="WinJS.Promise.thenEach_p:values">
                /// A set of values (which could be either an array or an object) of which some or all are promises.
                /// </param>
                /// <param name="onComplete" type="Function" locid="WinJS.Promise.thenEach_p:complete">
                /// The function to be called if the promise is fulfilled with a value.
                /// If the value is null, the promise returns the value.
                /// The value is passed as the single argument.
                /// </param>
                /// <param name="onError" type="Function" optional="true" locid="WinJS.Promise.thenEach_p:error">
                /// The function to be called if the promise is fulfilled with an error. The error
                /// is passed as the single argument.
                /// </param>
                /// <param name="onProgress" type="Function" optional="true" locid="WinJS.Promise.thenEach_p:progress">
                /// The function to be called if the promise reports progress. Data about the progress
                /// is passed as the single argument. Promises are not required to support
                /// progress.
                /// </param>
                /// <returns type="WinJS.Promise" locid="WinJS.Promise.thenEach_returnValue">
                /// A promise that is the result of calling Promise.join on the values parameter.
                /// </returns>
                /// </signature>
                var result = Array.isArray(values) ? [] : {};
                Object.keys(values).forEach(function (key) {
                    result[key] = Promise.as(values[key]).then(onComplete, onError, onProgress);
                });
                return Promise.join(result);
            },
            timeout: function Promise_timeout(time, promise) {
                /// <signature helpKeyword="WinJS.Promise.timeout">
                /// <summary locid="WinJS.Promise.timeout">
                /// Creates a promise that is fulfilled after a timeout.
                /// </summary>
                /// <param name="timeout" type="Number" optional="true" locid="WinJS.Promise.timeout_p:timeout">
                /// The timeout period in milliseconds. If this value is zero or not specified
                /// setImmediate is called, otherwise setTimeout is called.
                /// </param>
                /// <param name="promise" type="Promise" optional="true" locid="WinJS.Promise.timeout_p:promise">
                /// A promise that will be canceled if it doesn't complete before the
                /// timeout has expired.
                /// </param>
                /// <returns type="WinJS.Promise" locid="WinJS.Promise.timeout_returnValue">
                /// A promise that is completed asynchronously after the specified timeout.
                /// </returns>
                /// </signature>
                var to = timeout(time);
                return promise ? timeoutWithPromise(to, promise) : to;
            },
            wrap: function Promise_wrap(value) {
                /// <signature helpKeyword="WinJS.Promise.wrap">
                /// <summary locid="WinJS.Promise.wrap">
                /// Wraps a non-promise value in a promise. You can use this function if you need
                /// to pass a value to a function that requires a promise.
                /// </summary>
                /// <param name="value" locid="WinJS.Promise.wrap_p:value">
                /// Some non-promise value to be wrapped in a promise.
                /// </param>
                /// <returns type="WinJS.Promise" locid="WinJS.Promise.wrap_returnValue">
                /// A promise that is successfully fulfilled with the specified value
                /// </returns>
                /// </signature>
                return new CompletePromise(value);
            },
            wrapError: function Promise_wrapError(error) {
                /// <signature helpKeyword="WinJS.Promise.wrapError">
                /// <summary locid="WinJS.Promise.wrapError">
                /// Wraps a non-promise error value in a promise. You can use this function if you need
                /// to pass an error to a function that requires a promise.
                /// </summary>
                /// <param name="error" locid="WinJS.Promise.wrapError_p:error">
                /// A non-promise error value to be wrapped in a promise.
                /// </param>
                /// <returns type="WinJS.Promise" locid="WinJS.Promise.wrapError_returnValue">
                /// A promise that is in an error state with the specified value.
                /// </returns>
                /// </signature>
                return new ErrorPromise(error);
            },

            _veryExpensiveTagWithStack: {
                get: function () { return tagWithStack; },
                set: function (value) { tagWithStack = value; }
            },
            _veryExpensiveTagWithStack_tag: tag,
            _getStack: function () {
                if (_Global.Debug && _Global.Debug.debuggerEnabled) {
                    try { throw new Error(); } catch (e) { return e.stack; }
                }
            },

            _cancelBlocker: function Promise__cancelBlocker(input, oncancel) {
                //
                // Returns a promise which on cancelation will still result in downstream cancelation while
                //  protecting the promise 'input' from being  canceled which has the effect of allowing
                //  'input' to be shared amoung various consumers.
                //
                if (!Promise.is(input)) {
                    return Promise.wrap(input);
                }
                var complete;
                var error;
                var output = new Promise(
                    function (c, e) {
                        complete = c;
                        error = e;
                    },
                    function () {
                        complete = null;
                        error = null;
                        oncancel && oncancel();
                    }
                );
                input.then(
                    function (v) { complete && complete(v); },
                    function (e) { error && error(e); }
                );
                return output;
            },

        }
    );
    Object.defineProperties(Promise, _Events.createEventProperties(errorET));

    Promise._doneHandler = function (value) {
        _BaseCoreUtils._setImmediate(function Promise_done_rethrow() {
            throw value;
        });
    };

    return {
        PromiseStateMachine: PromiseStateMachine,
        Promise: Promise,
        state_created: state_created
    };
});

_winjs("WinJS/Promise", ["WinJS/Core/_Base","WinJS/Promise/_StateMachine"], function promiseInit( _Base, _StateMachine) {
    "use strict";

    _Base.Namespace.define("WinJS", {
        Promise: _StateMachine.Promise
    });

    return _StateMachine.Promise;
});

var exported = _modules["WinJS/Core/_WinJS"];

if (typeof exports === 'undefined' && typeof define === 'function' && define.amd) {
    define("vs/base/common/winjs.base.raw", exported);
} else {
    module.exports = exported;
}

if (typeof process !== 'undefined' && typeof process.nextTick === 'function') {
    _modules["WinJS/Core/_BaseCoreUtils"]._setImmediate = function(handler) {
        return process.nextTick(handler);
    };
}

})();
define(__m[91], __M([1,0]), function (require, exports) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    exports.MessageType = {
        INITIALIZE: '$initialize',
        REPLY: '$reply',
        PRINT: '$print'
    };
    exports.ReplyType = {
        COMPLETE: 'complete',
        ERROR: 'error',
        PROGRESS: 'progress'
    };
    exports.PrintType = {
        LOG: 'log',
        DEBUG: 'debug',
        INFO: 'info',
        WARN: 'warn',
        ERROR: 'error'
    };
});

define(__m[30], __M([1,0]), function (require, exports) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var Arrays;
    (function (Arrays) {
        /**
         * Given a sorted array of natural number segments, find the segment containing a natural number.
         *    For example, the segments [0, 5), [5, 9), [9, infinity) will be represented in the following manner:
         *       [{ startIndex: 0 }, { startIndex: 5 }, { startIndex: 9 }]
         *    Searching for 0, 1, 2, 3 or 4 will return 0.
         *    Searching for 5, 6, 7 or 8 will return 1.
         *    Searching for 9, 10, 11, ... will return 2.
         * @param arr A sorted array representing natural number segments
         * @param desiredIndex The search
         * @return The index of the containing segment in the array.
         */
        function findIndexInSegmentsArray(arr, desiredIndex) {
            var low = 0;
            var high = arr.length - 1;
            if (high <= 0) {
                return 0;
            }
            while (low < high) {
                var mid = low + Math.ceil((high - low) / 2);
                if (arr[mid].startIndex > desiredIndex) {
                    high = mid - 1;
                }
                else {
                    low = mid;
                }
            }
            return low;
        }
        Arrays.findIndexInSegmentsArray = findIndexInSegmentsArray;
    })(Arrays = exports.Arrays || (exports.Arrays = {}));
});

define(__m[18], __M([1,0,30]), function (require, exports, arrays_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var ModeTransition = (function () {
        function ModeTransition(startIndex, mode) {
            this.startIndex = startIndex | 0;
            this.mode = mode;
            this.modeId = mode.getId();
        }
        ModeTransition.findIndexInSegmentsArray = function (arr, desiredIndex) {
            return arrays_1.Arrays.findIndexInSegmentsArray(arr, desiredIndex);
        };
        ModeTransition.create = function (modeTransitions) {
            var result = [];
            for (var i = 0, len = modeTransitions.length; i < len; i++) {
                var modeTransition = modeTransitions[i];
                result.push(new ModeTransition(modeTransition.startIndex, modeTransition.mode));
            }
            return result;
        };
        return ModeTransition;
    }());
    exports.ModeTransition = ModeTransition;
});

define(__m[22], __M([1,0]), function (require, exports) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    /**
     * A position in the editor.
     */
    var Position = (function () {
        function Position(lineNumber, column) {
            this.lineNumber = lineNumber;
            this.column = column;
        }
        /**
         * Test if this position equals other position
         */
        Position.prototype.equals = function (other) {
            return Position.equals(this, other);
        };
        /**
         * Test if position `a` equals position `b`
         */
        Position.equals = function (a, b) {
            if (!a && !b) {
                return true;
            }
            return (!!a &&
                !!b &&
                a.lineNumber === b.lineNumber &&
                a.column === b.column);
        };
        /**
         * Test if this position is before other position.
         * If the two positions are equal, the result will be false.
         */
        Position.prototype.isBefore = function (other) {
            return Position.isBefore(this, other);
        };
        /**
         * Test if position `a` is before position `b`.
         * If the two positions are equal, the result will be false.
         */
        Position.isBefore = function (a, b) {
            if (a.lineNumber < b.lineNumber) {
                return true;
            }
            if (b.lineNumber < a.lineNumber) {
                return false;
            }
            return a.column < b.column;
        };
        /**
         * Test if this position is before other position.
         * If the two positions are equal, the result will be true.
         */
        Position.prototype.isBeforeOrEqual = function (other) {
            return Position.isBeforeOrEqual(this, other);
        };
        /**
         * Test if position `a` is before position `b`.
         * If the two positions are equal, the result will be true.
         */
        Position.isBeforeOrEqual = function (a, b) {
            if (a.lineNumber < b.lineNumber) {
                return true;
            }
            if (b.lineNumber < a.lineNumber) {
                return false;
            }
            return a.column <= b.column;
        };
        /**
         * Clone this position.
         */
        Position.prototype.clone = function () {
            return new Position(this.lineNumber, this.column);
        };
        /**
         * Convert to a human-readable representation.
         */
        Position.prototype.toString = function () {
            return '(' + this.lineNumber + ',' + this.column + ')';
        };
        // ---
        /**
         * Create a `Position` from an `IPosition`.
         */
        Position.lift = function (pos) {
            return new Position(pos.lineNumber, pos.column);
        };
        /**
         * Test if `obj` is an `IPosition`.
         */
        Position.isIPosition = function (obj) {
            return (obj
                && (typeof obj.lineNumber === 'number')
                && (typeof obj.column === 'number'));
        };
        return Position;
    }());
    exports.Position = Position;
});

/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Microsoft Corporation. All rights reserved.
 *  Licensed under the MIT License. See License.txt in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
define(__m[29], __M([1,0,22]), function (require, exports, position_1) {
    'use strict';
    /**
     * A range in the editor. (startLineNumber,startColumn) is <= (endLineNumber,endColumn)
     */
    var Range = (function () {
        function Range(startLineNumber, startColumn, endLineNumber, endColumn) {
            if ((startLineNumber > endLineNumber) || (startLineNumber === endLineNumber && startColumn > endColumn)) {
                this.startLineNumber = endLineNumber;
                this.startColumn = endColumn;
                this.endLineNumber = startLineNumber;
                this.endColumn = startColumn;
            }
            else {
                this.startLineNumber = startLineNumber;
                this.startColumn = startColumn;
                this.endLineNumber = endLineNumber;
                this.endColumn = endColumn;
            }
        }
        /**
         * Test if this range is empty.
         */
        Range.prototype.isEmpty = function () {
            return Range.isEmpty(this);
        };
        /**
         * Test if `range` is empty.
         */
        Range.isEmpty = function (range) {
            return (range.startLineNumber === range.endLineNumber && range.startColumn === range.endColumn);
        };
        /**
         * Test if position is in this range. If the position is at the edges, will return true.
         */
        Range.prototype.containsPosition = function (position) {
            return Range.containsPosition(this, position);
        };
        /**
         * Test if `position` is in `range`. If the position is at the edges, will return true.
         */
        Range.containsPosition = function (range, position) {
            if (position.lineNumber < range.startLineNumber || position.lineNumber > range.endLineNumber) {
                return false;
            }
            if (position.lineNumber === range.startLineNumber && position.column < range.startColumn) {
                return false;
            }
            if (position.lineNumber === range.endLineNumber && position.column > range.endColumn) {
                return false;
            }
            return true;
        };
        /**
         * Test if range is in this range. If the range is equal to this range, will return true.
         */
        Range.prototype.containsRange = function (range) {
            return Range.containsRange(this, range);
        };
        /**
         * Test if `otherRange` is in `range`. If the ranges are equal, will return true.
         */
        Range.containsRange = function (range, otherRange) {
            if (otherRange.startLineNumber < range.startLineNumber || otherRange.endLineNumber < range.startLineNumber) {
                return false;
            }
            if (otherRange.startLineNumber > range.endLineNumber || otherRange.endLineNumber > range.endLineNumber) {
                return false;
            }
            if (otherRange.startLineNumber === range.startLineNumber && otherRange.startColumn < range.startColumn) {
                return false;
            }
            if (otherRange.endLineNumber === range.endLineNumber && otherRange.endColumn > range.endColumn) {
                return false;
            }
            return true;
        };
        /**
         * A reunion of the two ranges.
         * The smallest position will be used as the start point, and the largest one as the end point.
         */
        Range.prototype.plusRange = function (range) {
            return Range.plusRange(this, range);
        };
        /**
         * A reunion of the two ranges.
         * The smallest position will be used as the start point, and the largest one as the end point.
         */
        Range.plusRange = function (a, b) {
            var startLineNumber, startColumn, endLineNumber, endColumn;
            if (b.startLineNumber < a.startLineNumber) {
                startLineNumber = b.startLineNumber;
                startColumn = b.startColumn;
            }
            else if (b.startLineNumber === a.startLineNumber) {
                startLineNumber = b.startLineNumber;
                startColumn = Math.min(b.startColumn, a.startColumn);
            }
            else {
                startLineNumber = a.startLineNumber;
                startColumn = a.startColumn;
            }
            if (b.endLineNumber > a.endLineNumber) {
                endLineNumber = b.endLineNumber;
                endColumn = b.endColumn;
            }
            else if (b.endLineNumber === a.endLineNumber) {
                endLineNumber = b.endLineNumber;
                endColumn = Math.max(b.endColumn, a.endColumn);
            }
            else {
                endLineNumber = a.endLineNumber;
                endColumn = a.endColumn;
            }
            return new Range(startLineNumber, startColumn, endLineNumber, endColumn);
        };
        /**
         * A intersection of the two ranges.
         */
        Range.prototype.intersectRanges = function (range) {
            return Range.intersectRanges(this, range);
        };
        /**
         * A intersection of the two ranges.
         */
        Range.intersectRanges = function (a, b) {
            var resultStartLineNumber = a.startLineNumber, resultStartColumn = a.startColumn, resultEndLineNumber = a.endLineNumber, resultEndColumn = a.endColumn, otherStartLineNumber = b.startLineNumber, otherStartColumn = b.startColumn, otherEndLineNumber = b.endLineNumber, otherEndColumn = b.endColumn;
            if (resultStartLineNumber < otherStartLineNumber) {
                resultStartLineNumber = otherStartLineNumber;
                resultStartColumn = otherStartColumn;
            }
            else if (resultStartLineNumber === otherStartLineNumber) {
                resultStartColumn = Math.max(resultStartColumn, otherStartColumn);
            }
            if (resultEndLineNumber > otherEndLineNumber) {
                resultEndLineNumber = otherEndLineNumber;
                resultEndColumn = otherEndColumn;
            }
            else if (resultEndLineNumber === otherEndLineNumber) {
                resultEndColumn = Math.min(resultEndColumn, otherEndColumn);
            }
            // Check if selection is now empty
            if (resultStartLineNumber > resultEndLineNumber) {
                return null;
            }
            if (resultStartLineNumber === resultEndLineNumber && resultStartColumn > resultEndColumn) {
                return null;
            }
            return new Range(resultStartLineNumber, resultStartColumn, resultEndLineNumber, resultEndColumn);
        };
        /**
         * Test if this range equals other.
         */
        Range.prototype.equalsRange = function (other) {
            return Range.equalsRange(this, other);
        };
        /**
         * Test if range `a` equals `b`.
         */
        Range.equalsRange = function (a, b) {
            return (!!a &&
                !!b &&
                a.startLineNumber === b.startLineNumber &&
                a.startColumn === b.startColumn &&
                a.endLineNumber === b.endLineNumber &&
                a.endColumn === b.endColumn);
        };
        /**
         * Return the end position (which will be after or equal to the start position)
         */
        Range.prototype.getEndPosition = function () {
            return new position_1.Position(this.endLineNumber, this.endColumn);
        };
        /**
         * Return the start position (which will be before or equal to the end position)
         */
        Range.prototype.getStartPosition = function () {
            return new position_1.Position(this.startLineNumber, this.startColumn);
        };
        /**
         * Clone this range.
         */
        Range.prototype.cloneRange = function () {
            return new Range(this.startLineNumber, this.startColumn, this.endLineNumber, this.endColumn);
        };
        /**
         * Transform to a user presentable string representation.
         */
        Range.prototype.toString = function () {
            return '[' + this.startLineNumber + ',' + this.startColumn + ' -> ' + this.endLineNumber + ',' + this.endColumn + ']';
        };
        /**
         * Create a new range using this range's start position, and using endLineNumber and endColumn as the end position.
         */
        Range.prototype.setEndPosition = function (endLineNumber, endColumn) {
            return new Range(this.startLineNumber, this.startColumn, endLineNumber, endColumn);
        };
        /**
         * Create a new range using this range's end position, and using startLineNumber and startColumn as the start position.
         */
        Range.prototype.setStartPosition = function (startLineNumber, startColumn) {
            return new Range(startLineNumber, startColumn, this.endLineNumber, this.endColumn);
        };
        /**
         * Create a new empty range using this range's start position.
         */
        Range.prototype.collapseToStart = function () {
            return Range.collapseToStart(this);
        };
        /**
         * Create a new empty range using this range's start position.
         */
        Range.collapseToStart = function (range) {
            return new Range(range.startLineNumber, range.startColumn, range.startLineNumber, range.startColumn);
        };
        // ---
        /**
         * Create a `Range` from an `IRange`.
         */
        Range.lift = function (range) {
            if (!range) {
                return null;
            }
            return new Range(range.startLineNumber, range.startColumn, range.endLineNumber, range.endColumn);
        };
        /**
         * Test if `obj` is an `IRange`.
         */
        Range.isIRange = function (obj) {
            return (obj
                && (typeof obj.startLineNumber === 'number')
                && (typeof obj.startColumn === 'number')
                && (typeof obj.endLineNumber === 'number')
                && (typeof obj.endColumn === 'number'));
        };
        /**
         * Test if the two ranges are touching in any way.
         */
        Range.areIntersectingOrTouching = function (a, b) {
            // Check if `a` is before `b`
            if (a.endLineNumber < b.startLineNumber || (a.endLineNumber === b.startLineNumber && a.endColumn < b.startColumn)) {
                return false;
            }
            // Check if `b` is before `a`
            if (b.endLineNumber < a.startLineNumber || (b.endLineNumber === a.startLineNumber && b.endColumn < a.startColumn)) {
                return false;
            }
            // These ranges must intersect
            return true;
        };
        /**
         * A function that compares ranges, useful for sorting ranges
         * It will first compare ranges on the startPosition and then on the endPosition
         */
        Range.compareRangesUsingStarts = function (a, b) {
            var aStartLineNumber = a.startLineNumber | 0;
            var bStartLineNumber = b.startLineNumber | 0;
            var aStartColumn = a.startColumn | 0;
            var bStartColumn = b.startColumn | 0;
            var aEndLineNumber = a.endLineNumber | 0;
            var bEndLineNumber = b.endLineNumber | 0;
            var aEndColumn = a.endColumn | 0;
            var bEndColumn = b.endColumn | 0;
            if (aStartLineNumber === bStartLineNumber) {
                if (aStartColumn === bStartColumn) {
                    if (aEndLineNumber === bEndLineNumber) {
                        return aEndColumn - bEndColumn;
                    }
                    return aEndLineNumber - bEndLineNumber;
                }
                return aStartColumn - bStartColumn;
            }
            return aStartLineNumber - bStartLineNumber;
        };
        /**
         * A function that compares ranges, useful for sorting ranges
         * It will first compare ranges on the endPosition and then on the startPosition
         */
        Range.compareRangesUsingEnds = function (a, b) {
            if (a.endLineNumber === b.endLineNumber) {
                if (a.endColumn === b.endColumn) {
                    if (a.startLineNumber === b.startLineNumber) {
                        return a.startColumn - b.startColumn;
                    }
                    return a.startLineNumber - b.startLineNumber;
                }
                return a.endColumn - b.endColumn;
            }
            return a.endLineNumber - b.endLineNumber;
        };
        /**
         * Test if the range spans multiple lines.
         */
        Range.spansMultipleLines = function (range) {
            return range.endLineNumber > range.startLineNumber;
        };
        return Range;
    }());
    exports.Range = Range;
});

define(__m[50], __M([1,0,30]), function (require, exports, arrays_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    /**
     * A token on a line.
     */
    var ViewLineToken = (function () {
        function ViewLineToken(startIndex, type) {
            this.startIndex = startIndex | 0; // @perf
            this.type = type.replace(/[^a-z0-9\-]/gi, ' ');
        }
        ViewLineToken.prototype.equals = function (other) {
            return (this.startIndex === other.startIndex
                && this.type === other.type);
        };
        ViewLineToken.findIndexInSegmentsArray = function (arr, desiredIndex) {
            return arrays_1.Arrays.findIndexInSegmentsArray(arr, desiredIndex);
        };
        ViewLineToken.equalsArray = function (a, b) {
            var aLen = a.length;
            var bLen = b.length;
            if (aLen !== bLen) {
                return false;
            }
            for (var i = 0; i < aLen; i++) {
                if (!a[i].equals(b[i])) {
                    return false;
                }
            }
            return true;
        };
        return ViewLineToken;
    }());
    exports.ViewLineToken = ViewLineToken;
    var ViewLineTokens = (function () {
        function ViewLineTokens(lineTokens, fauxIndentLength, textLength) {
            this._lineTokens = lineTokens;
            this._fauxIndentLength = fauxIndentLength | 0;
            this._textLength = textLength | 0;
        }
        ViewLineTokens.prototype.getTokens = function () {
            return this._lineTokens;
        };
        ViewLineTokens.prototype.getFauxIndentLength = function () {
            return this._fauxIndentLength;
        };
        ViewLineTokens.prototype.getTextLength = function () {
            return this._textLength;
        };
        ViewLineTokens.prototype.equals = function (other) {
            return (this._fauxIndentLength === other._fauxIndentLength
                && this._textLength === other._textLength
                && ViewLineToken.equalsArray(this._lineTokens, other._lineTokens));
        };
        ViewLineTokens.prototype.findIndexOfOffset = function (offset) {
            return ViewLineToken.findIndexInSegmentsArray(this._lineTokens, offset);
        };
        return ViewLineTokens;
    }());
    exports.ViewLineTokens = ViewLineTokens;
});






define(__m[19], __M([1,0]), function (require, exports) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    /**
     * Describes how to indent wrapped lines.
     */
    (function (WrappingIndent) {
        /**
         * No indentation => wrapped lines begin at column 1.
         */
        WrappingIndent[WrappingIndent["None"] = 0] = "None";
        /**
         * Same => wrapped lines get the same indentation as the parent.
         */
        WrappingIndent[WrappingIndent["Same"] = 1] = "Same";
        /**
         * Indent => wrapped lines get +1 indentation as the parent.
         */
        WrappingIndent[WrappingIndent["Indent"] = 2] = "Indent";
    })(exports.WrappingIndent || (exports.WrappingIndent = {}));
    var WrappingIndent = exports.WrappingIndent;
    var InternalEditorScrollbarOptions = (function () {
        /**
         * @internal
         */
        function InternalEditorScrollbarOptions(source) {
            this.arrowSize = source.arrowSize | 0;
            this.vertical = source.vertical | 0;
            this.horizontal = source.horizontal | 0;
            this.useShadows = Boolean(source.useShadows);
            this.verticalHasArrows = Boolean(source.verticalHasArrows);
            this.horizontalHasArrows = Boolean(source.horizontalHasArrows);
            this.handleMouseWheel = Boolean(source.handleMouseWheel);
            this.horizontalScrollbarSize = source.horizontalScrollbarSize | 0;
            this.horizontalSliderSize = source.horizontalSliderSize | 0;
            this.verticalScrollbarSize = source.verticalScrollbarSize | 0;
            this.verticalSliderSize = source.verticalSliderSize | 0;
            this.mouseWheelScrollSensitivity = Number(source.mouseWheelScrollSensitivity);
        }
        /**
         * @internal
         */
        InternalEditorScrollbarOptions.prototype.equals = function (other) {
            return (this.arrowSize === other.arrowSize
                && this.vertical === other.vertical
                && this.horizontal === other.horizontal
                && this.useShadows === other.useShadows
                && this.verticalHasArrows === other.verticalHasArrows
                && this.horizontalHasArrows === other.horizontalHasArrows
                && this.handleMouseWheel === other.handleMouseWheel
                && this.horizontalScrollbarSize === other.horizontalScrollbarSize
                && this.horizontalSliderSize === other.horizontalSliderSize
                && this.verticalScrollbarSize === other.verticalScrollbarSize
                && this.verticalSliderSize === other.verticalSliderSize
                && this.mouseWheelScrollSensitivity === other.mouseWheelScrollSensitivity);
        };
        /**
         * @internal
         */
        InternalEditorScrollbarOptions.prototype.clone = function () {
            return new InternalEditorScrollbarOptions(this);
        };
        return InternalEditorScrollbarOptions;
    }());
    exports.InternalEditorScrollbarOptions = InternalEditorScrollbarOptions;
    var EditorWrappingInfo = (function () {
        /**
         * @internal
         */
        function EditorWrappingInfo(source) {
            this.isViewportWrapping = Boolean(source.isViewportWrapping);
            this.wrappingColumn = source.wrappingColumn | 0;
            this.wrappingIndent = source.wrappingIndent | 0;
            this.wordWrapBreakBeforeCharacters = String(source.wordWrapBreakBeforeCharacters);
            this.wordWrapBreakAfterCharacters = String(source.wordWrapBreakAfterCharacters);
            this.wordWrapBreakObtrusiveCharacters = String(source.wordWrapBreakObtrusiveCharacters);
        }
        /**
         * @internal
         */
        EditorWrappingInfo.prototype.equals = function (other) {
            return (this.isViewportWrapping === other.isViewportWrapping
                && this.wrappingColumn === other.wrappingColumn
                && this.wrappingIndent === other.wrappingIndent
                && this.wordWrapBreakBeforeCharacters === other.wordWrapBreakBeforeCharacters
                && this.wordWrapBreakAfterCharacters === other.wordWrapBreakAfterCharacters
                && this.wordWrapBreakObtrusiveCharacters === other.wordWrapBreakObtrusiveCharacters);
        };
        /**
         * @internal
         */
        EditorWrappingInfo.prototype.clone = function () {
            return new EditorWrappingInfo(this);
        };
        return EditorWrappingInfo;
    }());
    exports.EditorWrappingInfo = EditorWrappingInfo;
    var InternalEditorViewOptions = (function () {
        /**
         * @internal
         */
        function InternalEditorViewOptions(source) {
            this.theme = String(source.theme);
            this.canUseTranslate3d = Boolean(source.canUseTranslate3d);
            this.experimentalScreenReader = Boolean(source.experimentalScreenReader);
            this.rulers = InternalEditorViewOptions._toSortedIntegerArray(source.rulers);
            this.ariaLabel = String(source.ariaLabel);
            this.lineNumbers = source.lineNumbers;
            this.selectOnLineNumbers = Boolean(source.selectOnLineNumbers);
            this.glyphMargin = Boolean(source.glyphMargin);
            this.revealHorizontalRightPadding = source.revealHorizontalRightPadding | 0;
            this.roundedSelection = Boolean(source.roundedSelection);
            this.overviewRulerLanes = source.overviewRulerLanes | 0;
            this.cursorBlinking = String(source.cursorBlinking);
            this.cursorStyle = source.cursorStyle | 0;
            this.hideCursorInOverviewRuler = Boolean(source.hideCursorInOverviewRuler);
            this.scrollBeyondLastLine = Boolean(source.scrollBeyondLastLine);
            this.editorClassName = String(source.editorClassName);
            this.stopRenderingLineAfter = source.stopRenderingLineAfter | 0;
            this.renderWhitespace = Boolean(source.renderWhitespace);
            this.indentGuides = Boolean(source.indentGuides);
            this.scrollbar = source.scrollbar.clone();
        }
        InternalEditorViewOptions._toSortedIntegerArray = function (source) {
            if (!Array.isArray(source)) {
                return [];
            }
            var arrSource = source;
            var result = arrSource.map(function (el) {
                var r = parseInt(el, 10);
                if (isNaN(r)) {
                    return 0;
                }
                return r;
            });
            result.sort();
            return result;
        };
        InternalEditorViewOptions._numberArraysEqual = function (a, b) {
            if (a.length !== b.length) {
                return false;
            }
            for (var i = 0; i < a.length; i++) {
                if (a[i] !== b[i]) {
                    return false;
                }
            }
            return true;
        };
        /**
         * @internal
         */
        InternalEditorViewOptions.prototype.equals = function (other) {
            return (this.theme === other.theme
                && this.canUseTranslate3d === other.canUseTranslate3d
                && this.experimentalScreenReader === other.experimentalScreenReader
                && InternalEditorViewOptions._numberArraysEqual(this.rulers, other.rulers)
                && this.ariaLabel === other.ariaLabel
                && this.lineNumbers === other.lineNumbers
                && this.selectOnLineNumbers === other.selectOnLineNumbers
                && this.glyphMargin === other.glyphMargin
                && this.revealHorizontalRightPadding === other.revealHorizontalRightPadding
                && this.roundedSelection === other.roundedSelection
                && this.overviewRulerLanes === other.overviewRulerLanes
                && this.cursorBlinking === other.cursorBlinking
                && this.cursorStyle === other.cursorStyle
                && this.hideCursorInOverviewRuler === other.hideCursorInOverviewRuler
                && this.scrollBeyondLastLine === other.scrollBeyondLastLine
                && this.editorClassName === other.editorClassName
                && this.stopRenderingLineAfter === other.stopRenderingLineAfter
                && this.renderWhitespace === other.renderWhitespace
                && this.indentGuides === other.indentGuides
                && this.scrollbar.equals(other.scrollbar));
        };
        /**
         * @internal
         */
        InternalEditorViewOptions.prototype.createChangeEvent = function (newOpts) {
            return {
                theme: this.theme !== newOpts.theme,
                canUseTranslate3d: this.canUseTranslate3d !== newOpts.canUseTranslate3d,
                experimentalScreenReader: this.experimentalScreenReader !== newOpts.experimentalScreenReader,
                rulers: (!InternalEditorViewOptions._numberArraysEqual(this.rulers, newOpts.rulers)),
                ariaLabel: this.ariaLabel !== newOpts.ariaLabel,
                lineNumbers: this.lineNumbers !== newOpts.lineNumbers,
                selectOnLineNumbers: this.selectOnLineNumbers !== newOpts.selectOnLineNumbers,
                glyphMargin: this.glyphMargin !== newOpts.glyphMargin,
                revealHorizontalRightPadding: this.revealHorizontalRightPadding !== newOpts.revealHorizontalRightPadding,
                roundedSelection: this.roundedSelection !== newOpts.roundedSelection,
                overviewRulerLanes: this.overviewRulerLanes !== newOpts.overviewRulerLanes,
                cursorBlinking: this.cursorBlinking !== newOpts.cursorBlinking,
                cursorStyle: this.cursorStyle !== newOpts.cursorStyle,
                hideCursorInOverviewRuler: this.hideCursorInOverviewRuler !== newOpts.hideCursorInOverviewRuler,
                scrollBeyondLastLine: this.scrollBeyondLastLine !== newOpts.scrollBeyondLastLine,
                editorClassName: this.editorClassName !== newOpts.editorClassName,
                stopRenderingLineAfter: this.stopRenderingLineAfter !== newOpts.stopRenderingLineAfter,
                renderWhitespace: this.renderWhitespace !== newOpts.renderWhitespace,
                indentGuides: this.indentGuides !== newOpts.indentGuides,
                scrollbar: (!this.scrollbar.equals(newOpts.scrollbar)),
            };
        };
        /**
         * @internal
         */
        InternalEditorViewOptions.prototype.clone = function () {
            return new InternalEditorViewOptions(this);
        };
        return InternalEditorViewOptions;
    }());
    exports.InternalEditorViewOptions = InternalEditorViewOptions;
    var EditorContribOptions = (function () {
        /**
         * @internal
         */
        function EditorContribOptions(source) {
            this.selectionClipboard = Boolean(source.selectionClipboard);
            this.hover = Boolean(source.hover);
            this.contextmenu = Boolean(source.contextmenu);
            this.quickSuggestions = Boolean(source.quickSuggestions);
            this.quickSuggestionsDelay = source.quickSuggestionsDelay || 0;
            this.parameterHints = Boolean(source.parameterHints);
            this.iconsInSuggestions = Boolean(source.iconsInSuggestions);
            this.formatOnType = Boolean(source.formatOnType);
            this.suggestOnTriggerCharacters = Boolean(source.suggestOnTriggerCharacters);
            this.acceptSuggestionOnEnter = Boolean(source.acceptSuggestionOnEnter);
            this.selectionHighlight = Boolean(source.selectionHighlight);
            this.outlineMarkers = Boolean(source.outlineMarkers);
            this.referenceInfos = Boolean(source.referenceInfos);
            this.folding = Boolean(source.folding);
        }
        /**
         * @internal
         */
        EditorContribOptions.prototype.equals = function (other) {
            return (this.selectionClipboard === other.selectionClipboard
                && this.hover === other.hover
                && this.contextmenu === other.contextmenu
                && this.quickSuggestions === other.quickSuggestions
                && this.quickSuggestionsDelay === other.quickSuggestionsDelay
                && this.parameterHints === other.parameterHints
                && this.iconsInSuggestions === other.iconsInSuggestions
                && this.formatOnType === other.formatOnType
                && this.suggestOnTriggerCharacters === other.suggestOnTriggerCharacters
                && this.acceptSuggestionOnEnter === other.acceptSuggestionOnEnter
                && this.selectionHighlight === other.selectionHighlight
                && this.outlineMarkers === other.outlineMarkers
                && this.referenceInfos === other.referenceInfos
                && this.folding === other.folding);
        };
        /**
         * @internal
         */
        EditorContribOptions.prototype.clone = function () {
            return new EditorContribOptions(this);
        };
        return EditorContribOptions;
    }());
    exports.EditorContribOptions = EditorContribOptions;
    /**
     * Internal configuration options (transformed or computed) for the editor.
     */
    var InternalEditorOptions = (function () {
        /**
         * @internal
         */
        function InternalEditorOptions(source) {
            this.lineHeight = source.lineHeight | 0;
            this.readOnly = Boolean(source.readOnly);
            this.wordSeparators = String(source.wordSeparators);
            this.autoClosingBrackets = Boolean(source.autoClosingBrackets);
            this.useTabStops = Boolean(source.useTabStops);
            this.tabFocusMode = Boolean(source.tabFocusMode);
            this.layoutInfo = source.layoutInfo.clone();
            this.fontInfo = source.fontInfo.clone();
            this.viewInfo = source.viewInfo.clone();
            this.wrappingInfo = source.wrappingInfo.clone();
            this.contribInfo = source.contribInfo.clone();
        }
        /**
         * @internal
         */
        InternalEditorOptions.prototype.equals = function (other) {
            return (this.lineHeight === other.lineHeight
                && this.readOnly === other.readOnly
                && this.wordSeparators === other.wordSeparators
                && this.autoClosingBrackets === other.autoClosingBrackets
                && this.useTabStops === other.useTabStops
                && this.tabFocusMode === other.tabFocusMode
                && this.layoutInfo.equals(other.layoutInfo)
                && this.fontInfo.equals(other.fontInfo)
                && this.viewInfo.equals(other.viewInfo)
                && this.wrappingInfo.equals(other.wrappingInfo)
                && this.contribInfo.equals(other.contribInfo));
        };
        /**
         * @internal
         */
        InternalEditorOptions.prototype.createChangeEvent = function (newOpts) {
            return {
                lineHeight: (this.lineHeight !== newOpts.lineHeight),
                readOnly: (this.readOnly !== newOpts.readOnly),
                wordSeparators: (this.wordSeparators !== newOpts.wordSeparators),
                autoClosingBrackets: (this.autoClosingBrackets !== newOpts.autoClosingBrackets),
                useTabStops: (this.useTabStops !== newOpts.useTabStops),
                tabFocusMode: (this.tabFocusMode !== newOpts.tabFocusMode),
                layoutInfo: (!this.layoutInfo.equals(newOpts.layoutInfo)),
                fontInfo: (!this.fontInfo.equals(newOpts.fontInfo)),
                viewInfo: this.viewInfo.createChangeEvent(newOpts.viewInfo),
                wrappingInfo: (!this.wrappingInfo.equals(newOpts.wrappingInfo)),
                contribInfo: (!this.contribInfo.equals(newOpts.contribInfo)),
            };
        };
        /**
         * @internal
         */
        InternalEditorOptions.prototype.clone = function () {
            return new InternalEditorOptions(this);
        };
        return InternalEditorOptions;
    }());
    exports.InternalEditorOptions = InternalEditorOptions;
    /**
     * Vertical Lane in the overview ruler of the editor.
     */
    (function (OverviewRulerLane) {
        OverviewRulerLane[OverviewRulerLane["Left"] = 1] = "Left";
        OverviewRulerLane[OverviewRulerLane["Center"] = 2] = "Center";
        OverviewRulerLane[OverviewRulerLane["Right"] = 4] = "Right";
        OverviewRulerLane[OverviewRulerLane["Full"] = 7] = "Full";
    })(exports.OverviewRulerLane || (exports.OverviewRulerLane = {}));
    var OverviewRulerLane = exports.OverviewRulerLane;
    /**
     * End of line character preference.
     */
    (function (EndOfLinePreference) {
        /**
         * Use the end of line character identified in the text buffer.
         */
        EndOfLinePreference[EndOfLinePreference["TextDefined"] = 0] = "TextDefined";
        /**
         * Use line feed (\n) as the end of line character.
         */
        EndOfLinePreference[EndOfLinePreference["LF"] = 1] = "LF";
        /**
         * Use carriage return and line feed (\r\n) as the end of line character.
         */
        EndOfLinePreference[EndOfLinePreference["CRLF"] = 2] = "CRLF";
    })(exports.EndOfLinePreference || (exports.EndOfLinePreference = {}));
    var EndOfLinePreference = exports.EndOfLinePreference;
    /**
     * The default end of line to use when instantiating models.
     */
    (function (DefaultEndOfLine) {
        /**
         * Use line feed (\n) as the end of line character.
         */
        DefaultEndOfLine[DefaultEndOfLine["LF"] = 1] = "LF";
        /**
         * Use carriage return and line feed (\r\n) as the end of line character.
         */
        DefaultEndOfLine[DefaultEndOfLine["CRLF"] = 2] = "CRLF";
    })(exports.DefaultEndOfLine || (exports.DefaultEndOfLine = {}));
    var DefaultEndOfLine = exports.DefaultEndOfLine;
    /**
     * End of line character preference.
     */
    (function (EndOfLineSequence) {
        /**
         * Use line feed (\n) as the end of line character.
         */
        EndOfLineSequence[EndOfLineSequence["LF"] = 0] = "LF";
        /**
         * Use carriage return and line feed (\r\n) as the end of line character.
         */
        EndOfLineSequence[EndOfLineSequence["CRLF"] = 1] = "CRLF";
    })(exports.EndOfLineSequence || (exports.EndOfLineSequence = {}));
    var EndOfLineSequence = exports.EndOfLineSequence;
    /**
     * Describes the behaviour of decorations when typing/editing near their edges.
     */
    (function (TrackedRangeStickiness) {
        TrackedRangeStickiness[TrackedRangeStickiness["AlwaysGrowsWhenTypingAtEdges"] = 0] = "AlwaysGrowsWhenTypingAtEdges";
        TrackedRangeStickiness[TrackedRangeStickiness["NeverGrowsWhenTypingAtEdges"] = 1] = "NeverGrowsWhenTypingAtEdges";
        TrackedRangeStickiness[TrackedRangeStickiness["GrowsOnlyWhenTypingBefore"] = 2] = "GrowsOnlyWhenTypingBefore";
        TrackedRangeStickiness[TrackedRangeStickiness["GrowsOnlyWhenTypingAfter"] = 3] = "GrowsOnlyWhenTypingAfter";
    })(exports.TrackedRangeStickiness || (exports.TrackedRangeStickiness = {}));
    var TrackedRangeStickiness = exports.TrackedRangeStickiness;
    /**
     * Describes the reason the cursor has changed its position.
     */
    (function (CursorChangeReason) {
        /**
         * Unknown or not set.
         */
        CursorChangeReason[CursorChangeReason["NotSet"] = 0] = "NotSet";
        /**
         * A `model.setValue()` was called.
         */
        CursorChangeReason[CursorChangeReason["ContentFlush"] = 1] = "ContentFlush";
        /**
         * The `model` has been changed outside of this cursor and the cursor recovers its position from associated markers.
         */
        CursorChangeReason[CursorChangeReason["RecoverFromMarkers"] = 2] = "RecoverFromMarkers";
        /**
         * There was an explicit user gesture.
         */
        CursorChangeReason[CursorChangeReason["Explicit"] = 3] = "Explicit";
        /**
         * There was a Paste.
         */
        CursorChangeReason[CursorChangeReason["Paste"] = 4] = "Paste";
        /**
         * There was an Undo.
         */
        CursorChangeReason[CursorChangeReason["Undo"] = 5] = "Undo";
        /**
         * There was a Redo.
         */
        CursorChangeReason[CursorChangeReason["Redo"] = 6] = "Redo";
    })(exports.CursorChangeReason || (exports.CursorChangeReason = {}));
    var CursorChangeReason = exports.CursorChangeReason;
    /**
     * @internal
     */
    (function (VerticalRevealType) {
        VerticalRevealType[VerticalRevealType["Simple"] = 0] = "Simple";
        VerticalRevealType[VerticalRevealType["Center"] = 1] = "Center";
        VerticalRevealType[VerticalRevealType["CenterIfOutsideViewport"] = 2] = "CenterIfOutsideViewport";
    })(exports.VerticalRevealType || (exports.VerticalRevealType = {}));
    var VerticalRevealType = exports.VerticalRevealType;
    /**
     * A description for the overview ruler position.
     */
    var OverviewRulerPosition = (function () {
        /**
         * @internal
         */
        function OverviewRulerPosition(source) {
            this.width = source.width | 0;
            this.height = source.height | 0;
            this.top = source.top | 0;
            this.right = source.right | 0;
        }
        /**
         * @internal
         */
        OverviewRulerPosition.prototype.equals = function (other) {
            return (this.width === other.width
                && this.height === other.height
                && this.top === other.top
                && this.right === other.right);
        };
        /**
         * @internal
         */
        OverviewRulerPosition.prototype.clone = function () {
            return new OverviewRulerPosition(this);
        };
        return OverviewRulerPosition;
    }());
    exports.OverviewRulerPosition = OverviewRulerPosition;
    /**
     * The internal layout details of the editor.
     */
    var EditorLayoutInfo = (function () {
        /**
         * @internal
         */
        function EditorLayoutInfo(source) {
            this.width = source.width | 0;
            this.height = source.height | 0;
            this.glyphMarginLeft = source.glyphMarginLeft | 0;
            this.glyphMarginWidth = source.glyphMarginWidth | 0;
            this.glyphMarginHeight = source.glyphMarginHeight | 0;
            this.lineNumbersLeft = source.lineNumbersLeft | 0;
            this.lineNumbersWidth = source.lineNumbersWidth | 0;
            this.lineNumbersHeight = source.lineNumbersHeight | 0;
            this.decorationsLeft = source.decorationsLeft | 0;
            this.decorationsWidth = source.decorationsWidth | 0;
            this.decorationsHeight = source.decorationsHeight | 0;
            this.contentLeft = source.contentLeft | 0;
            this.contentWidth = source.contentWidth | 0;
            this.contentHeight = source.contentHeight | 0;
            this.verticalScrollbarWidth = source.verticalScrollbarWidth | 0;
            this.horizontalScrollbarHeight = source.horizontalScrollbarHeight | 0;
            this.overviewRuler = source.overviewRuler.clone();
        }
        /**
         * @internal
         */
        EditorLayoutInfo.prototype.equals = function (other) {
            return (this.width === other.width
                && this.height === other.height
                && this.glyphMarginLeft === other.glyphMarginLeft
                && this.glyphMarginWidth === other.glyphMarginWidth
                && this.glyphMarginHeight === other.glyphMarginHeight
                && this.lineNumbersLeft === other.lineNumbersLeft
                && this.lineNumbersWidth === other.lineNumbersWidth
                && this.lineNumbersHeight === other.lineNumbersHeight
                && this.decorationsLeft === other.decorationsLeft
                && this.decorationsWidth === other.decorationsWidth
                && this.decorationsHeight === other.decorationsHeight
                && this.contentLeft === other.contentLeft
                && this.contentWidth === other.contentWidth
                && this.contentHeight === other.contentHeight
                && this.verticalScrollbarWidth === other.verticalScrollbarWidth
                && this.horizontalScrollbarHeight === other.horizontalScrollbarHeight
                && this.overviewRuler.equals(other.overviewRuler));
        };
        /**
         * @internal
         */
        EditorLayoutInfo.prototype.clone = function () {
            return new EditorLayoutInfo(this);
        };
        return EditorLayoutInfo;
    }());
    exports.EditorLayoutInfo = EditorLayoutInfo;
    /**
     * Type of hit element with the mouse in the editor.
     */
    (function (MouseTargetType) {
        /**
         * Mouse is on top of an unknown element.
         */
        MouseTargetType[MouseTargetType["UNKNOWN"] = 0] = "UNKNOWN";
        /**
         * Mouse is on top of the textarea used for input.
         */
        MouseTargetType[MouseTargetType["TEXTAREA"] = 1] = "TEXTAREA";
        /**
         * Mouse is on top of the glyph margin
         */
        MouseTargetType[MouseTargetType["GUTTER_GLYPH_MARGIN"] = 2] = "GUTTER_GLYPH_MARGIN";
        /**
         * Mouse is on top of the line numbers
         */
        MouseTargetType[MouseTargetType["GUTTER_LINE_NUMBERS"] = 3] = "GUTTER_LINE_NUMBERS";
        /**
         * Mouse is on top of the line decorations
         */
        MouseTargetType[MouseTargetType["GUTTER_LINE_DECORATIONS"] = 4] = "GUTTER_LINE_DECORATIONS";
        /**
         * Mouse is on top of the whitespace left in the gutter by a view zone.
         */
        MouseTargetType[MouseTargetType["GUTTER_VIEW_ZONE"] = 5] = "GUTTER_VIEW_ZONE";
        /**
         * Mouse is on top of text in the content.
         */
        MouseTargetType[MouseTargetType["CONTENT_TEXT"] = 6] = "CONTENT_TEXT";
        /**
         * Mouse is on top of empty space in the content (e.g. after line text or below last line)
         */
        MouseTargetType[MouseTargetType["CONTENT_EMPTY"] = 7] = "CONTENT_EMPTY";
        /**
         * Mouse is on top of a view zone in the content.
         */
        MouseTargetType[MouseTargetType["CONTENT_VIEW_ZONE"] = 8] = "CONTENT_VIEW_ZONE";
        /**
         * Mouse is on top of a content widget.
         */
        MouseTargetType[MouseTargetType["CONTENT_WIDGET"] = 9] = "CONTENT_WIDGET";
        /**
         * Mouse is on top of the decorations overview ruler.
         */
        MouseTargetType[MouseTargetType["OVERVIEW_RULER"] = 10] = "OVERVIEW_RULER";
        /**
         * Mouse is on top of a scrollbar.
         */
        MouseTargetType[MouseTargetType["SCROLLBAR"] = 11] = "SCROLLBAR";
        /**
         * Mouse is on top of an overlay widget.
         */
        MouseTargetType[MouseTargetType["OVERLAY_WIDGET"] = 12] = "OVERLAY_WIDGET";
    })(exports.MouseTargetType || (exports.MouseTargetType = {}));
    var MouseTargetType = exports.MouseTargetType;
    /**
     * A context key that is set when the editor's text has focus (cursor is blinking).
     */
    exports.KEYBINDING_CONTEXT_EDITOR_TEXT_FOCUS = 'editorTextFocus';
    /**
     * A context key that is set when the editor's text or an editor's widget has focus.
     */
    exports.KEYBINDING_CONTEXT_EDITOR_FOCUS = 'editorFocus';
    /**
     * @internal
     */
    exports.KEYBINDING_CONTEXT_EDITOR_TAB_MOVES_FOCUS = 'editorTabMovesFocus';
    /**
     * A context key that is set when the editor has multiple selections (multiple cursors).
     */
    exports.KEYBINDING_CONTEXT_EDITOR_HAS_MULTIPLE_SELECTIONS = 'editorHasMultipleSelections';
    /**
     * A context key that is set when the editor has a non-collapsed selection.
     */
    exports.KEYBINDING_CONTEXT_EDITOR_HAS_NON_EMPTY_SELECTION = 'editorHasSelection';
    /**
     * A context key that is set to the language associated with the model associated with the editor.
     */
    exports.KEYBINDING_CONTEXT_EDITOR_LANGUAGE_ID = 'editorLangId';
    /**
     * @internal
     */
    exports.SHOW_ACCESSIBILITY_HELP_ACTION_ID = 'editor.action.showAccessibilityHelp';
    var BareFontInfo = (function () {
        /**
         * @internal
         */
        function BareFontInfo(opts) {
            this.fontFamily = String(opts.fontFamily);
            this.fontSize = opts.fontSize | 0;
            this.lineHeight = opts.lineHeight | 0;
        }
        /**
         * @internal
         */
        BareFontInfo.prototype.getId = function () {
            return this.fontFamily + '-' + this.fontSize + '-' + this.lineHeight;
        };
        return BareFontInfo;
    }());
    exports.BareFontInfo = BareFontInfo;
    var FontInfo = (function (_super) {
        __extends(FontInfo, _super);
        /**
         * @internal
         */
        function FontInfo(opts) {
            _super.call(this, opts);
            this.typicalHalfwidthCharacterWidth = opts.typicalHalfwidthCharacterWidth;
            this.typicalFullwidthCharacterWidth = opts.typicalFullwidthCharacterWidth;
            this.spaceWidth = opts.spaceWidth;
            this.maxDigitWidth = opts.maxDigitWidth;
        }
        /**
         * @internal
         */
        FontInfo.prototype.equals = function (other) {
            return (this.fontFamily === other.fontFamily
                && this.fontSize === other.fontSize
                && this.lineHeight === other.lineHeight
                && this.typicalHalfwidthCharacterWidth === other.typicalHalfwidthCharacterWidth
                && this.typicalFullwidthCharacterWidth === other.typicalFullwidthCharacterWidth
                && this.spaceWidth === other.spaceWidth
                && this.maxDigitWidth === other.maxDigitWidth);
        };
        /**
         * @internal
         */
        FontInfo.prototype.clone = function () {
            return new FontInfo(this);
        };
        return FontInfo;
    }(BareFontInfo));
    exports.FontInfo = FontInfo;
    /**
     * @internal
     */
    exports.ViewEventNames = {
        ModelFlushedEvent: 'modelFlushedEvent',
        LinesDeletedEvent: 'linesDeletedEvent',
        LinesInsertedEvent: 'linesInsertedEvent',
        LineChangedEvent: 'lineChangedEvent',
        TokensChangedEvent: 'tokensChangedEvent',
        DecorationsChangedEvent: 'decorationsChangedEvent',
        CursorPositionChangedEvent: 'cursorPositionChangedEvent',
        CursorSelectionChangedEvent: 'cursorSelectionChangedEvent',
        RevealRangeEvent: 'revealRangeEvent',
        LineMappingChangedEvent: 'lineMappingChangedEvent',
        ScrollRequestEvent: 'scrollRequestEvent'
    };
    /**
     * @internal
     */
    var Viewport = (function () {
        function Viewport(top, left, width, height) {
            this.top = top | 0;
            this.left = left | 0;
            this.width = width | 0;
            this.height = height | 0;
        }
        return Viewport;
    }());
    exports.Viewport = Viewport;
    /**
     * @internal
     */
    (function (CodeEditorStateFlag) {
        CodeEditorStateFlag[CodeEditorStateFlag["Value"] = 0] = "Value";
        CodeEditorStateFlag[CodeEditorStateFlag["Selection"] = 1] = "Selection";
        CodeEditorStateFlag[CodeEditorStateFlag["Position"] = 2] = "Position";
        CodeEditorStateFlag[CodeEditorStateFlag["Scroll"] = 3] = "Scroll";
    })(exports.CodeEditorStateFlag || (exports.CodeEditorStateFlag = {}));
    var CodeEditorStateFlag = exports.CodeEditorStateFlag;
    /**
     * The type of the `IEditor`.
     */
    exports.EditorType = {
        ICodeEditor: 'vs.editor.ICodeEditor',
        IDiffEditor: 'vs.editor.IDiffEditor'
    };
    /**
     * @internal
     */
    exports.ClassName = {
        EditorWarningDecoration: 'greensquiggly',
        EditorErrorDecoration: 'redsquiggly'
    };
    /**
     * @internal
     */
    exports.EventType = {
        Disposed: 'disposed',
        ConfigurationChanged: 'configurationChanged',
        ModelDispose: 'modelDispose',
        ModelChanged: 'modelChanged',
        ModelTokensChanged: 'modelTokensChanged',
        ModelModeChanged: 'modelsModeChanged',
        ModelModeSupportChanged: 'modelsModeSupportChanged',
        ModelOptionsChanged: 'modelOptionsChanged',
        ModelRawContentChanged: 'contentChanged',
        ModelContentChanged2: 'contentChanged2',
        ModelRawContentChangedFlush: 'flush',
        ModelRawContentChangedLinesDeleted: 'linesDeleted',
        ModelRawContentChangedLinesInserted: 'linesInserted',
        ModelRawContentChangedLineChanged: 'lineChanged',
        EditorTextBlur: 'blur',
        EditorTextFocus: 'focus',
        EditorFocus: 'widgetFocus',
        EditorBlur: 'widgetBlur',
        ModelDecorationsChanged: 'decorationsChanged',
        CursorPositionChanged: 'positionChanged',
        CursorSelectionChanged: 'selectionChanged',
        CursorRevealRange: 'revealRange',
        CursorScrollRequest: 'scrollRequest',
        ViewFocusGained: 'focusGained',
        ViewFocusLost: 'focusLost',
        ViewFocusChanged: 'focusChanged',
        ViewScrollChanged: 'scrollChanged',
        ViewZonesChanged: 'zonesChanged',
        ViewLayoutChanged: 'viewLayoutChanged',
        ContextMenu: 'contextMenu',
        MouseDown: 'mousedown',
        MouseUp: 'mouseup',
        MouseMove: 'mousemove',
        MouseLeave: 'mouseleave',
        KeyDown: 'keydown',
        KeyUp: 'keyup',
        EditorLayout: 'editorLayout',
        DiffUpdated: 'diffUpdated'
    };
    /**
     * Built-in commands.
     */
    exports.Handler = {
        ExecuteCommand: 'executeCommand',
        ExecuteCommands: 'executeCommands',
        CursorLeft: 'cursorLeft',
        CursorLeftSelect: 'cursorLeftSelect',
        CursorWordLeft: 'cursorWordLeft',
        CursorWordStartLeft: 'cursorWordStartLeft',
        CursorWordEndLeft: 'cursorWordEndLeft',
        CursorWordLeftSelect: 'cursorWordLeftSelect',
        CursorWordStartLeftSelect: 'cursorWordStartLeftSelect',
        CursorWordEndLeftSelect: 'cursorWordEndLeftSelect',
        CursorRight: 'cursorRight',
        CursorRightSelect: 'cursorRightSelect',
        CursorWordRight: 'cursorWordRight',
        CursorWordStartRight: 'cursorWordStartRight',
        CursorWordEndRight: 'cursorWordEndRight',
        CursorWordRightSelect: 'cursorWordRightSelect',
        CursorWordStartRightSelect: 'cursorWordStartRightSelect',
        CursorWordEndRightSelect: 'cursorWordEndRightSelect',
        CursorUp: 'cursorUp',
        CursorUpSelect: 'cursorUpSelect',
        CursorDown: 'cursorDown',
        CursorDownSelect: 'cursorDownSelect',
        CursorPageUp: 'cursorPageUp',
        CursorPageUpSelect: 'cursorPageUpSelect',
        CursorPageDown: 'cursorPageDown',
        CursorPageDownSelect: 'cursorPageDownSelect',
        CursorHome: 'cursorHome',
        CursorHomeSelect: 'cursorHomeSelect',
        CursorEnd: 'cursorEnd',
        CursorEndSelect: 'cursorEndSelect',
        ExpandLineSelection: 'expandLineSelection',
        CursorTop: 'cursorTop',
        CursorTopSelect: 'cursorTopSelect',
        CursorBottom: 'cursorBottom',
        CursorBottomSelect: 'cursorBottomSelect',
        CursorColumnSelectLeft: 'cursorColumnSelectLeft',
        CursorColumnSelectRight: 'cursorColumnSelectRight',
        CursorColumnSelectUp: 'cursorColumnSelectUp',
        CursorColumnSelectPageUp: 'cursorColumnSelectPageUp',
        CursorColumnSelectDown: 'cursorColumnSelectDown',
        CursorColumnSelectPageDown: 'cursorColumnSelectPageDown',
        AddCursorDown: 'addCursorDown',
        AddCursorUp: 'addCursorUp',
        CursorUndo: 'cursorUndo',
        MoveTo: 'moveTo',
        MoveToSelect: 'moveToSelect',
        ColumnSelect: 'columnSelect',
        CreateCursor: 'createCursor',
        LastCursorMoveToSelect: 'lastCursorMoveToSelect',
        JumpToBracket: 'jumpToBracket',
        Type: 'type',
        ReplacePreviousChar: 'replacePreviousChar',
        Paste: 'paste',
        Tab: 'tab',
        Indent: 'indent',
        Outdent: 'outdent',
        DeleteLeft: 'deleteLeft',
        DeleteRight: 'deleteRight',
        DeleteWordLeft: 'deleteWordLeft',
        DeleteWordStartLeft: 'deleteWordStartLeft',
        DeleteWordEndLeft: 'deleteWordEndLeft',
        DeleteWordRight: 'deleteWordRight',
        DeleteWordStartRight: 'deleteWordStartRight',
        DeleteWordEndRight: 'deleteWordEndRight',
        DeleteAllLeft: 'deleteAllLeft',
        DeleteAllRight: 'deleteAllRight',
        RemoveSecondaryCursors: 'removeSecondaryCursors',
        CancelSelection: 'cancelSelection',
        Cut: 'cut',
        Undo: 'undo',
        Redo: 'redo',
        WordSelect: 'wordSelect',
        WordSelectDrag: 'wordSelectDrag',
        LastCursorWordSelect: 'lastCursorWordSelect',
        LineSelect: 'lineSelect',
        LineSelectDrag: 'lineSelectDrag',
        LastCursorLineSelect: 'lastCursorLineSelect',
        LastCursorLineSelectDrag: 'lastCursorLineSelectDrag',
        LineInsertBefore: 'lineInsertBefore',
        LineInsertAfter: 'lineInsertAfter',
        LineBreakInsert: 'lineBreakInsert',
        SelectAll: 'selectAll',
        ScrollLineUp: 'scrollLineUp',
        ScrollLineDown: 'scrollLineDown',
        ScrollPageUp: 'scrollPageUp',
        ScrollPageDown: 'scrollPageDown'
    };
    /**
     * The style in which the editor's cursor should be rendered.
     */
    (function (TextEditorCursorStyle) {
        /**
         * As a vertical line (sitting between two characters).
         */
        TextEditorCursorStyle[TextEditorCursorStyle["Line"] = 1] = "Line";
        /**
         * As a block (sitting on top of a character).
         */
        TextEditorCursorStyle[TextEditorCursorStyle["Block"] = 2] = "Block";
        /**
         * As a horizontal line (sitting under a character).
         */
        TextEditorCursorStyle[TextEditorCursorStyle["Underline"] = 3] = "Underline";
    })(exports.TextEditorCursorStyle || (exports.TextEditorCursorStyle = {}));
    var TextEditorCursorStyle = exports.TextEditorCursorStyle;
    /**
     * @internal
     */
    function cursorStyleToString(cursorStyle) {
        if (cursorStyle === TextEditorCursorStyle.Line) {
            return 'line';
        }
        else if (cursorStyle === TextEditorCursorStyle.Block) {
            return 'block';
        }
        else if (cursorStyle === TextEditorCursorStyle.Underline) {
            return 'underline';
        }
        else {
            throw new Error('cursorStyleToString: Unknown cursorStyle');
        }
    }
    exports.cursorStyleToString = cursorStyleToString;
    /**
     * @internal
     */
    var ColorZone = (function () {
        function ColorZone(from, to, colorId, position) {
            this.from = from | 0;
            this.to = to | 0;
            this.colorId = colorId | 0;
            this.position = position | 0;
        }
        return ColorZone;
    }());
    exports.ColorZone = ColorZone;
    /**
     * A zone in the overview ruler
     * @internal
     */
    var OverviewRulerZone = (function () {
        function OverviewRulerZone(startLineNumber, endLineNumber, position, forceHeight, color, darkColor) {
            this.startLineNumber = startLineNumber;
            this.endLineNumber = endLineNumber;
            this.position = position;
            this.forceHeight = forceHeight;
            this._color = color;
            this._darkColor = darkColor;
            this._colorZones = null;
        }
        OverviewRulerZone.prototype.getColor = function (useDarkColor) {
            if (useDarkColor) {
                return this._darkColor;
            }
            return this._color;
        };
        OverviewRulerZone.prototype.equals = function (other) {
            return (this.startLineNumber === other.startLineNumber
                && this.endLineNumber === other.endLineNumber
                && this.position === other.position
                && this.forceHeight === other.forceHeight
                && this._color === other._color
                && this._darkColor === other._darkColor);
        };
        OverviewRulerZone.prototype.compareTo = function (other) {
            if (this.startLineNumber === other.startLineNumber) {
                if (this.endLineNumber === other.endLineNumber) {
                    if (this.forceHeight === other.forceHeight) {
                        if (this.position === other.position) {
                            if (this._darkColor === other._darkColor) {
                                if (this._color === other._color) {
                                    return 0;
                                }
                                return this._color < other._color ? -1 : 1;
                            }
                            return this._darkColor < other._darkColor ? -1 : 1;
                        }
                        return this.position - other.position;
                    }
                    return this.forceHeight - other.forceHeight;
                }
                return this.endLineNumber - other.endLineNumber;
            }
            return this.startLineNumber - other.startLineNumber;
        };
        OverviewRulerZone.prototype.setColorZones = function (colorZones) {
            this._colorZones = colorZones;
        };
        OverviewRulerZone.prototype.getColorZones = function () {
            return this._colorZones;
        };
        return OverviewRulerZone;
    }());
    exports.OverviewRulerZone = OverviewRulerZone;
});

define(__m[95], __M([1,0]), function (require, exports) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var __space = ' '.charCodeAt(0);
    var __tab = '\t'.charCodeAt(0);
    /**
     * Compute the diff in spaces between two line's indentation.
     */
    function spacesDiff(a, aLength, b, bLength) {
        // This can go both ways (e.g.):
        //  - a: "\t"
        //  - b: "\t    "
        //  => This should count 1 tab and 4 spaces
        var i;
        for (i = 0; i < aLength && i < bLength; i++) {
            var aCharCode = a.charCodeAt(i);
            var bCharCode = b.charCodeAt(i);
            if (aCharCode !== bCharCode) {
                break;
            }
        }
        var aSpacesCnt = 0, aTabsCount = 0;
        for (var j = i; j < aLength; j++) {
            var aCharCode = a.charCodeAt(j);
            if (aCharCode === __space) {
                aSpacesCnt++;
            }
            else {
                aTabsCount++;
            }
        }
        var bSpacesCnt = 0, bTabsCount = 0;
        for (var j = i; j < bLength; j++) {
            var bCharCode = b.charCodeAt(j);
            if (bCharCode === __space) {
                bSpacesCnt++;
            }
            else {
                bTabsCount++;
            }
        }
        if (aSpacesCnt > 0 && aTabsCount > 0) {
            return 0;
        }
        if (bSpacesCnt > 0 && bTabsCount > 0) {
            return 0;
        }
        var tabsDiff = Math.abs(aTabsCount - bTabsCount);
        var spacesDiff = Math.abs(aSpacesCnt - bSpacesCnt);
        if (tabsDiff === 0) {
            return spacesDiff;
        }
        if (spacesDiff % tabsDiff === 0) {
            return spacesDiff / tabsDiff;
        }
        return 0;
    }
    function guessIndentation(lines, defaultTabSize, defaultInsertSpaces) {
        var linesIndentedWithTabsCount = 0; // number of lines that contain at least one tab in indentation
        var linesIndentedWithSpacesCount = 0; // number of lines that contain only spaces in indentation
        var previousLineText = ''; // content of latest line that contained non-whitespace chars
        var previousLineIndentation = 0; // index at which latest line contained the first non-whitespace char
        var ALLOWED_TAB_SIZE_GUESSES = [2, 4, 6, 8]; // limit guesses for `tabSize` to 2, 4, 6 or 8.
        var MAX_ALLOWED_TAB_SIZE_GUESS = 8; // max(2,4,6,8) = 8
        var spacesDiffCount = [0, 0, 0, 0, 0, 0, 0, 0, 0]; // `tabSize` scores
        for (var i = 0, len = lines.length; i < len; i++) {
            var currentLineText = lines[i];
            var currentLineHasContent = false; // does `currentLineText` contain non-whitespace chars
            var currentLineIndentation = 0; // index at which `currentLineText` contains the first non-whitespace char
            var currentLineSpacesCount = 0; // count of spaces found in `currentLineText` indentation
            var currentLineTabsCount = 0; // count of tabs found in `currentLineText` indentation
            for (var j = 0, lenJ = currentLineText.length; j < lenJ; j++) {
                var charCode = currentLineText.charCodeAt(j);
                if (charCode === __tab) {
                    currentLineTabsCount++;
                }
                else if (charCode === __space) {
                    currentLineSpacesCount++;
                }
                else {
                    // Hit non whitespace character on this line
                    currentLineHasContent = true;
                    currentLineIndentation = j;
                    break;
                }
            }
            // Ignore empty or only whitespace lines
            if (!currentLineHasContent) {
                continue;
            }
            if (currentLineTabsCount > 0) {
                linesIndentedWithTabsCount++;
            }
            else if (currentLineSpacesCount > 1) {
                linesIndentedWithSpacesCount++;
            }
            var currentSpacesDiff = spacesDiff(previousLineText, previousLineIndentation, currentLineText, currentLineIndentation);
            if (currentSpacesDiff <= MAX_ALLOWED_TAB_SIZE_GUESS) {
                spacesDiffCount[currentSpacesDiff]++;
            }
            previousLineText = currentLineText;
            previousLineIndentation = currentLineIndentation;
        }
        // Take into account the last line as well
        var deltaSpacesCount = spacesDiff(previousLineText, previousLineIndentation, '', 0);
        if (deltaSpacesCount <= MAX_ALLOWED_TAB_SIZE_GUESS) {
            spacesDiffCount[deltaSpacesCount]++;
        }
        var insertSpaces = defaultInsertSpaces;
        if (linesIndentedWithTabsCount !== linesIndentedWithSpacesCount) {
            insertSpaces = (linesIndentedWithTabsCount < linesIndentedWithSpacesCount);
        }
        var tabSize = defaultTabSize;
        var tabSizeScore = (insertSpaces ? 0 : 0.1 * lines.length);
        // console.log("score threshold: " + tabSizeScore);
        ALLOWED_TAB_SIZE_GUESSES.forEach(function (possibleTabSize) {
            var possibleTabSizeScore = spacesDiffCount[possibleTabSize];
            if (possibleTabSizeScore > tabSizeScore) {
                tabSizeScore = possibleTabSizeScore;
                tabSize = possibleTabSize;
            }
        });
        // console.log('--------------------------');
        // console.log('linesIndentedWithTabsCount: ' + linesIndentedWithTabsCount + ', linesIndentedWithSpacesCount: ' + linesIndentedWithSpacesCount);
        // console.log('spacesDiffCount: ' + spacesDiffCount);
        // console.log('tabSize: ' + tabSize + ', tabSizeScore: ' + tabSizeScore);
        return {
            insertSpaces: insertSpaces,
            tabSize: tabSize
        };
    }
    exports.guessIndentation = guessIndentation;
});

define(__m[101], __M([1,0,30]), function (require, exports, arrays_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    /**
     * A token on a line.
     */
    var LineToken = (function () {
        function LineToken(startIndex, type) {
            this.startIndex = startIndex | 0; // @perf
            this.type = type;
        }
        LineToken.prototype.equals = function (other) {
            return (this.startIndex === other.startIndex
                && this.type === other.type);
        };
        LineToken.findIndexInSegmentsArray = function (arr, desiredIndex) {
            return arrays_1.Arrays.findIndexInSegmentsArray(arr, desiredIndex);
        };
        LineToken.equalsArray = function (a, b) {
            var aLen = a.length;
            var bLen = b.length;
            if (aLen !== bLen) {
                return false;
            }
            for (var i = 0; i < aLen; i++) {
                if (!a[i].equals(b[i])) {
                    return false;
                }
            }
            return true;
        };
        return LineToken;
    }());
    exports.LineToken = LineToken;
});

define(__m[69], __M([1,0]), function (require, exports) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var TokenIterator = (function () {
        function TokenIterator(model, position) {
            this._model = model;
            this._currentLineNumber = position.lineNumber;
            this._currentTokenIndex = 0;
            this._readLineTokens(this._currentLineNumber);
            this._next = null;
            this._prev = null;
            // start with a position to next/prev run
            var columnIndex = position.column - 1, tokenEndIndex = Number.MAX_VALUE;
            for (var i = this._currentLineTokens.getTokenCount() - 1; i >= 0; i--) {
                var tokenStartIndex = this._currentLineTokens.getTokenStartIndex(i);
                if (tokenStartIndex <= columnIndex && columnIndex <= tokenEndIndex) {
                    this._currentTokenIndex = i;
                    this._next = this._current();
                    this._prev = this._current();
                    break;
                }
                tokenEndIndex = tokenStartIndex;
            }
        }
        TokenIterator.prototype._readLineTokens = function (lineNumber) {
            this._currentLineTokens = this._model.getLineTokens(lineNumber, false);
        };
        TokenIterator.prototype._advanceNext = function () {
            this._prev = this._next;
            this._next = null;
            if (this._currentTokenIndex + 1 < this._currentLineTokens.getTokenCount()) {
                // There are still tokens on current line
                this._currentTokenIndex++;
                this._next = this._current();
            }
            else {
                // find the next line with tokens
                while (this._currentLineNumber + 1 <= this._model.getLineCount()) {
                    this._currentLineNumber++;
                    this._readLineTokens(this._currentLineNumber);
                    if (this._currentLineTokens.getTokenCount() > 0) {
                        this._currentTokenIndex = 0;
                        this._next = this._current();
                        break;
                    }
                }
                if (this._next === null) {
                    // prepare of a previous run
                    this._readLineTokens(this._currentLineNumber);
                    this._currentTokenIndex = this._currentLineTokens.getTokenCount();
                    this._advancePrev();
                    this._next = null;
                }
            }
        };
        TokenIterator.prototype._advancePrev = function () {
            this._next = this._prev;
            this._prev = null;
            if (this._currentTokenIndex > 0) {
                // There are still tokens on current line
                this._currentTokenIndex--;
                this._prev = this._current();
            }
            else {
                // find previous line with tokens
                while (this._currentLineNumber > 1) {
                    this._currentLineNumber--;
                    this._readLineTokens(this._currentLineNumber);
                    if (this._currentLineTokens.getTokenCount() > 0) {
                        this._currentTokenIndex = this._currentLineTokens.getTokenCount() - 1;
                        this._prev = this._current();
                        break;
                    }
                }
            }
        };
        TokenIterator.prototype._current = function () {
            var startIndex = this._currentLineTokens.getTokenStartIndex(this._currentTokenIndex);
            var type = this._currentLineTokens.getTokenType(this._currentTokenIndex);
            var endIndex = this._currentLineTokens.getTokenEndIndex(this._currentTokenIndex, this._model.getLineContent(this._currentLineNumber).length);
            return {
                token: {
                    startIndex: startIndex,
                    type: type
                },
                lineNumber: this._currentLineNumber,
                startColumn: startIndex + 1,
                endColumn: endIndex + 1
            };
        };
        TokenIterator.prototype.hasNext = function () {
            return this._next !== null;
        };
        TokenIterator.prototype.next = function () {
            var result = this._next;
            this._advanceNext();
            return result;
        };
        TokenIterator.prototype.hasPrev = function () {
            return this._prev !== null;
        };
        TokenIterator.prototype.prev = function () {
            var result = this._prev;
            this._advancePrev();
            return result;
        };
        TokenIterator.prototype._invalidate = function () {
            // replace all public functions with errors
            var errorFn = function () {
                throw new Error('iteration isn\'t valid anymore');
            };
            this.hasNext = errorFn;
            this.next = errorFn;
            this.hasPrev = errorFn;
            this.prev = errorFn;
        };
        return TokenIterator;
    }());
    exports.TokenIterator = TokenIterator;
});

define(__m[17], __M([1,0]), function (require, exports) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    exports.USUAL_WORD_SEPARATORS = '`~!@#$%^&*()-=+[{]}\\|;:\'",.<>/?';
    /**
     * Create a word definition regular expression based on default word separators.
     * Optionally provide allowed separators that should be included in words.
     *
     * The default would look like this:
     * /(-?\d*\.\d\w*)|([^\`\~\!\@\#\$\%\^\&\*\(\)\-\=\+\[\{\]\}\\\|\;\:\'\"\,\.\<\>\/\?\s]+)/g
     */
    function createWordRegExp(allowInWords) {
        if (allowInWords === void 0) { allowInWords = ''; }
        var usualSeparators = exports.USUAL_WORD_SEPARATORS;
        var source = '(-?\\d*\\.\\d\\w*)|([^';
        for (var i = 0; i < usualSeparators.length; i++) {
            if (allowInWords.indexOf(usualSeparators[i]) >= 0) {
                continue;
            }
            source += '\\' + usualSeparators[i];
        }
        source += '\\s]+)';
        return new RegExp(source, 'g');
    }
    exports.createWordRegExp = createWordRegExp;
    // catches numbers (including floating numbers) in the first group, and alphanum in the second
    exports.DEFAULT_WORD_REGEXP = createWordRegExp();
    function ensureValidWordDefinition(wordDefinition) {
        var result = exports.DEFAULT_WORD_REGEXP;
        if (wordDefinition && (wordDefinition instanceof RegExp)) {
            if (!wordDefinition.global) {
                var flags = 'g';
                if (wordDefinition.ignoreCase) {
                    flags += 'i';
                }
                if (wordDefinition.multiline) {
                    flags += 'm';
                }
                result = new RegExp(wordDefinition.source, flags);
            }
            else {
                result = wordDefinition;
            }
        }
        result.lastIndex = 0;
        return result;
    }
    exports.ensureValidWordDefinition = ensureValidWordDefinition;
    function getWordAtText(column, wordDefinition, text, textOffset) {
        // console.log('_getWordAtText: ', column, text, textOffset);
        var words = text.match(wordDefinition), k, startWord, endWord, startColumn, endColumn, word;
        if (words) {
            for (k = 0; k < words.length; k++) {
                word = words[k].trim();
                if (word.length > 0) {
                    startWord = text.indexOf(word, endWord);
                    endWord = startWord + word.length;
                    startColumn = textOffset + startWord + 1;
                    endColumn = textOffset + endWord + 1;
                    if (startColumn <= column && column <= endColumn) {
                        return {
                            word: word,
                            startColumn: startColumn,
                            endColumn: endColumn
                        };
                    }
                }
            }
        }
        return null;
    }
    exports.getWordAtText = getWordAtText;
});

define(__m[39], __M([1,0]), function (require, exports) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var AbstractState = (function () {
        function AbstractState(mode, stateData) {
            if (stateData === void 0) { stateData = null; }
            this.mode = mode;
            this.stateData = stateData;
        }
        AbstractState.prototype.getMode = function () {
            return this.mode;
        };
        AbstractState.prototype.clone = function () {
            var result = this.makeClone();
            result.initializeFrom(this);
            return result;
        };
        AbstractState.prototype.makeClone = function () {
            throw new Error('Abstract Method');
        };
        AbstractState.prototype.initializeFrom = function (other) {
            this.stateData = other.stateData !== null ? other.stateData.clone() : null;
        };
        AbstractState.prototype.getStateData = function () {
            return this.stateData;
        };
        AbstractState.prototype.setStateData = function (state) {
            this.stateData = state;
        };
        AbstractState.prototype.equals = function (other) {
            if (other === null || this.mode !== other.getMode()) {
                return false;
            }
            if (other instanceof AbstractState) {
                return AbstractState.safeEquals(this.stateData, other.stateData);
            }
            return false;
        };
        AbstractState.prototype.tokenize = function (stream) {
            throw new Error('Abstract Method');
        };
        AbstractState.safeEquals = function (a, b) {
            if (a === null && b === null) {
                return true;
            }
            if (a === null || b === null) {
                return false;
            }
            return a.equals(b);
        };
        AbstractState.safeClone = function (state) {
            if (state) {
                return state.clone();
            }
            return null;
        };
        return AbstractState;
    }());
    exports.AbstractState = AbstractState;
});

/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Microsoft Corporation. All rights reserved.
 *  Licensed under the MIT License. See License.txt in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
define(__m[82], __M([1,0,55]), function (require, exports, glob_1) {
    'use strict';
    function matches(selection, uri, language) {
        return score(selection, uri, language) > 0;
    }
    Object.defineProperty(exports, "__esModule", { value: true });
    exports.default = matches;
    function score(selector, uri, language) {
        if (Array.isArray(selector)) {
            // for each
            var values = selector.map(function (item) { return score(item, uri, language); });
            return Math.max.apply(Math, values);
        }
        else if (typeof selector === 'string') {
            // compare language id
            if (selector === language) {
                return 10;
            }
            else if (selector === '*') {
                return 5;
            }
            else {
                return 0;
            }
        }
        else if (selector) {
            var filter = selector;
            var value = 0;
            // language id
            if (filter.language) {
                if (filter.language === language) {
                    value += 10;
                }
                else if (filter.language === '*') {
                    value += 5;
                }
                else {
                    return 0;
                }
            }
            // scheme
            if (filter.scheme) {
                if (filter.scheme === uri.scheme) {
                    value += 10;
                }
                else {
                    return 0;
                }
            }
            // match fsPath with pattern
            if (filter.pattern) {
                if (filter.pattern === uri.fsPath) {
                    value += 10;
                }
                else if (glob_1.match(filter.pattern, uri.fsPath)) {
                    value += 5;
                }
                else {
                    return 0;
                }
            }
            return value;
        }
    }
    exports.score = score;
});

define(__m[62], __M([1,0]), function (require, exports) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var LineStream = (function () {
        function LineStream(source) {
            this._source = source;
            this.sourceLength = source.length;
            this._pos = 0;
            this.whitespace = '\t \u00a0';
            this.whitespaceArr = this.stringToArray(this.whitespace);
            this.separators = '';
            this.separatorsArr = this.stringToArray(this.separators);
            this.tokenStart = -1;
            this.tokenEnd = -1;
        }
        LineStream.prototype.stringToArray = function (str) {
            if (!LineStream.STRING_TO_ARRAY_CACHE.hasOwnProperty(str)) {
                LineStream.STRING_TO_ARRAY_CACHE[str] = this.actualStringToArray(str);
            }
            return LineStream.STRING_TO_ARRAY_CACHE[str];
        };
        LineStream.prototype.actualStringToArray = function (str) {
            var maxCharCode = 0;
            for (var i = 0; i < str.length; i++) {
                maxCharCode = Math.max(maxCharCode, str.charCodeAt(i));
            }
            var r = [];
            for (var i = 0; i <= maxCharCode; i++) {
                r[i] = false;
            }
            for (var i = 0; i < str.length; i++) {
                r[str.charCodeAt(i)] = true;
            }
            return r;
        };
        LineStream.prototype.pos = function () {
            return this._pos;
        };
        LineStream.prototype.eos = function () {
            return this._pos >= this.sourceLength;
        };
        LineStream.prototype.peek = function () {
            // Check EOS
            if (this._pos >= this.sourceLength) {
                throw new Error('Stream is at the end');
            }
            return this._source[this._pos];
        };
        LineStream.prototype.next = function () {
            // Check EOS
            if (this._pos >= this.sourceLength) {
                throw new Error('Stream is at the end');
            }
            // Reset peeked token
            this.tokenStart = -1;
            this.tokenEnd = -1;
            return this._source[this._pos++];
        };
        LineStream.prototype.next2 = function () {
            // Check EOS
            if (this._pos >= this.sourceLength) {
                throw new Error('Stream is at the end');
            }
            // Reset peeked token
            this.tokenStart = -1;
            this.tokenEnd = -1;
            this._pos++;
        };
        LineStream.prototype.advance = function (n) {
            if (n === 0) {
                return '';
            }
            var oldPos = this._pos;
            this._pos += n;
            // Reset peeked token
            this.tokenStart = -1;
            this.tokenEnd = -1;
            return this._source.substring(oldPos, this._pos);
        };
        LineStream.prototype._advance2 = function (n) {
            if (n === 0) {
                return n;
            }
            this._pos += n;
            // Reset peeked token
            this.tokenStart = -1;
            this.tokenEnd = -1;
            return n;
        };
        LineStream.prototype.advanceToEOS = function () {
            var oldPos = this._pos;
            this._pos = this.sourceLength;
            this.resetPeekedToken();
            return this._source.substring(oldPos, this._pos);
        };
        LineStream.prototype.goBack = function (n) {
            this._pos -= n;
            this.resetPeekedToken();
        };
        LineStream.prototype.createPeeker = function (condition) {
            var _this = this;
            if (condition instanceof RegExp) {
                return function () {
                    var result = condition.exec(_this._source.substr(_this._pos));
                    if (result === null) {
                        return 0;
                    }
                    else if (result.index !== 0) {
                        throw new Error('Regular expression must begin with the character "^"');
                    }
                    return result[0].length;
                };
            }
            else if ((condition instanceof String || (typeof condition) === 'string') && condition) {
                return function () {
                    var len = condition.length, match = _this._pos + len <= _this.sourceLength;
                    for (var i = 0; match && i < len; i++) {
                        match = _this._source.charCodeAt(_this._pos + i) === condition.charCodeAt(i);
                    }
                    return match ? len : 0;
                };
            }
            throw new Error('Condition must be either a regular expression, function or a non-empty string');
        };
        // --- BEGIN `_advanceIfStringCaseInsensitive`
        LineStream.prototype._advanceIfStringCaseInsensitive = function (condition) {
            var oldPos = this._pos, source = this._source, len = condition.length, i;
            if (len < 1 || oldPos + len > this.sourceLength) {
                return 0;
            }
            for (i = 0; i < len; i++) {
                if (source.charAt(oldPos + i).toLowerCase() !== condition.charAt(i).toLowerCase()) {
                    return 0;
                }
            }
            return len;
        };
        LineStream.prototype.advanceIfStringCaseInsensitive = function (condition) {
            return this.advance(this._advanceIfStringCaseInsensitive(condition));
        };
        LineStream.prototype.advanceIfStringCaseInsensitive2 = function (condition) {
            return this._advance2(this._advanceIfStringCaseInsensitive(condition));
        };
        // --- END
        // --- BEGIN `advanceIfString`
        LineStream.prototype._advanceIfString = function (condition) {
            var oldPos = this._pos, source = this._source, len = condition.length, i;
            if (len < 1 || oldPos + len > this.sourceLength) {
                return 0;
            }
            for (i = 0; i < len; i++) {
                if (source.charCodeAt(oldPos + i) !== condition.charCodeAt(i)) {
                    return 0;
                }
            }
            return len;
        };
        LineStream.prototype.advanceIfString = function (condition) {
            return this.advance(this._advanceIfString(condition));
        };
        LineStream.prototype.advanceIfString2 = function (condition) {
            return this._advance2(this._advanceIfString(condition));
        };
        // --- END
        // --- BEGIN `advanceIfString`
        LineStream.prototype._advanceIfCharCode = function (charCode) {
            if (this._pos < this.sourceLength && this._source.charCodeAt(this._pos) === charCode) {
                return 1;
            }
            return 0;
        };
        LineStream.prototype.advanceIfCharCode = function (charCode) {
            return this.advance(this._advanceIfCharCode(charCode));
        };
        LineStream.prototype.advanceIfCharCode2 = function (charCode) {
            return this._advance2(this._advanceIfCharCode(charCode));
        };
        // --- END
        // --- BEGIN `advanceIfRegExp`
        LineStream.prototype._advanceIfRegExp = function (condition) {
            if (this._pos >= this.sourceLength) {
                return 0;
            }
            if (!condition.test(this._source.substr(this._pos))) {
                return 0;
            }
            return RegExp.lastMatch.length;
        };
        LineStream.prototype.advanceIfRegExp = function (condition) {
            return this.advance(this._advanceIfRegExp(condition));
        };
        LineStream.prototype.advanceIfRegExp2 = function (condition) {
            return this._advance2(this._advanceIfRegExp(condition));
        };
        // --- END
        LineStream.prototype.advanceLoop = function (condition, isWhile, including) {
            if (this.eos()) {
                return '';
            }
            var peeker = this.createPeeker(condition);
            var oldPos = this._pos;
            var n = 0;
            var f = null;
            if (isWhile) {
                f = function (n) {
                    return n > 0;
                };
            }
            else {
                f = function (n) {
                    return n === 0;
                };
            }
            while (!this.eos() && f(n = peeker())) {
                if (n > 0) {
                    this.advance(n);
                }
                else {
                    this.next();
                }
            }
            if (including && !this.eos()) {
                this.advance(n);
            }
            return this._source.substring(oldPos, this._pos);
        };
        LineStream.prototype.advanceWhile = function (condition) {
            return this.advanceLoop(condition, true, false);
        };
        LineStream.prototype.advanceUntil = function (condition, including) {
            return this.advanceLoop(condition, false, including);
        };
        // --- BEGIN `advanceUntilString`
        LineStream.prototype._advanceUntilString = function (condition, including) {
            if (this.eos() || condition.length === 0) {
                return 0;
            }
            var oldPos = this._pos;
            var index = this._source.indexOf(condition, oldPos);
            if (index === -1) {
                // String was not found => advanced to `eos`
                return (this.sourceLength - oldPos);
            }
            if (including) {
                // String was found => advance to include `condition`
                return (index + condition.length - oldPos);
            }
            // String was found => advance right before `condition`
            return (index - oldPos);
        };
        LineStream.prototype.advanceUntilString = function (condition, including) {
            return this.advance(this._advanceUntilString(condition, including));
        };
        LineStream.prototype.advanceUntilString2 = function (condition, including) {
            return this._advance2(this._advanceUntilString(condition, including));
        };
        // --- END
        LineStream.prototype.resetPeekedToken = function () {
            this.tokenStart = -1;
            this.tokenEnd = -1;
        };
        LineStream.prototype.setTokenRules = function (separators, whitespace) {
            if (this.separators !== separators || this.whitespace !== whitespace) {
                this.separators = separators;
                this.separatorsArr = this.stringToArray(this.separators);
                this.whitespace = whitespace;
                this.whitespaceArr = this.stringToArray(this.whitespace);
                this.resetPeekedToken();
            }
        };
        // --- tokens
        LineStream.prototype.peekToken = function () {
            if (this.tokenStart !== -1) {
                return this._source.substring(this.tokenStart, this.tokenEnd);
            }
            var source = this._source, sourceLength = this.sourceLength, whitespaceArr = this.whitespaceArr, separatorsArr = this.separatorsArr, tokenStart = this._pos;
            // Check EOS
            if (tokenStart >= sourceLength) {
                throw new Error('Stream is at the end');
            }
            // Skip whitespace
            while (whitespaceArr[source.charCodeAt(tokenStart)] && tokenStart < sourceLength) {
                tokenStart++;
            }
            var tokenEnd = tokenStart;
            // If a separator is hit, it is a token
            if (separatorsArr[source.charCodeAt(tokenEnd)] && tokenEnd < sourceLength) {
                tokenEnd++;
            }
            else {
                // Advance until a separator or a whitespace is hit
                while (!separatorsArr[source.charCodeAt(tokenEnd)] && !whitespaceArr[source.charCodeAt(tokenEnd)] && tokenEnd < sourceLength) {
                    tokenEnd++;
                }
            }
            // Cache peeked token
            this.tokenStart = tokenStart;
            this.tokenEnd = tokenEnd;
            return source.substring(tokenStart, tokenEnd);
        };
        LineStream.prototype.nextToken = function () {
            // Check EOS
            if (this._pos >= this.sourceLength) {
                throw new Error('Stream is at the end');
            }
            // Peek token if necessary
            var result;
            if (this.tokenStart === -1) {
                result = this.peekToken();
            }
            else {
                result = this._source.substring(this.tokenStart, this.tokenEnd);
            }
            // Advance to tokenEnd
            this._pos = this.tokenEnd;
            // Reset peeked token
            this.tokenStart = -1;
            this.tokenEnd = -1;
            return result;
        };
        // -- whitespace
        LineStream.prototype.peekWhitespace = function () {
            var source = this._source, sourceLength = this.sourceLength, whitespaceArr = this.whitespaceArr, peek = this._pos;
            while (whitespaceArr[source.charCodeAt(peek)] && peek < sourceLength) {
                peek++;
            }
            return source.substring(this._pos, peek);
        };
        // --- BEGIN `advanceIfRegExp`
        LineStream.prototype._skipWhitespace = function () {
            var source = this._source, sourceLength = this.sourceLength, whitespaceArr = this.whitespaceArr, oldPos = this._pos, peek = this._pos;
            while (whitespaceArr[source.charCodeAt(peek)] && peek < sourceLength) {
                peek++;
            }
            return (peek - oldPos);
        };
        LineStream.prototype.skipWhitespace = function () {
            return this.advance(this._skipWhitespace());
        };
        LineStream.prototype.skipWhitespace2 = function () {
            return this._advance2(this._skipWhitespace());
        };
        LineStream.STRING_TO_ARRAY_CACHE = {};
        return LineStream;
    }());
    exports.LineStream = LineStream;
});

define(__m[25], __M([1,0]), function (require, exports) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    /*
     * This module exports common types and functionality shared between
     * the Monarch compiler that compiles JSON to ILexer, and the Monarch
     * Tokenizer (that highlights at runtime)
     */
    /*
     * Type definitions to be used internally to Monarch.
     * Inside monarch we use fully typed definitions and compiled versions of the more abstract JSON descriptions.
     */
    (function (MonarchBracket) {
        MonarchBracket[MonarchBracket["None"] = 0] = "None";
        MonarchBracket[MonarchBracket["Open"] = 1] = "Open";
        MonarchBracket[MonarchBracket["Close"] = -1] = "Close";
    })(exports.MonarchBracket || (exports.MonarchBracket = {}));
    var MonarchBracket = exports.MonarchBracket;
    // Small helper functions
    /**
     * Is a string null, undefined, or empty?
     */
    function empty(s) {
        return (s ? false : true);
    }
    exports.empty = empty;
    /**
     * Puts a string to lower case if 'ignoreCase' is set.
     */
    function fixCase(lexer, str) {
        return (lexer.ignoreCase && str ? str.toLowerCase() : str);
    }
    exports.fixCase = fixCase;
    /**
     * Ensures there are no bad characters in a CSS token class.
     */
    function sanitize(s) {
        return s.replace(/[&<>'"_]/g, '-'); // used on all output token CSS classes
    }
    exports.sanitize = sanitize;
    // Logging
    /**
     * Logs a message.
     */
    function log(lexer, msg) {
        console.log(lexer.languageId + ": " + msg);
    }
    exports.log = log;
    // Throwing errors
    /**
     * Throws error. May actually just log the error and continue.
     */
    function throwError(lexer, msg) {
        throw new Error(lexer.languageId + ": " + msg);
    }
    exports.throwError = throwError;
    // Helper functions for rule finding and substitution
    /**
     * substituteMatches is used on lexer strings and can substitutes predefined patterns:
     * 		$$  => $
     * 		$#  => id
     * 		$n  => matched entry n
     * 		@attr => contents of lexer[attr]
     *
     * See documentation for more info
     */
    function substituteMatches(lexer, str, id, matches, state) {
        var re = /\$((\$)|(#)|(\d\d?)|[sS](\d\d?)|@(\w+))/g;
        var stateMatches = null;
        return str.replace(re, function (full, sub, dollar, hash, n, s, attr, ofs, total) {
            if (!empty(dollar)) {
                return '$'; // $$
            }
            if (!empty(hash)) {
                return fixCase(lexer, id); // default $#
            }
            if (!empty(n) && n < matches.length) {
                return fixCase(lexer, matches[n]); // $n
            }
            if (!empty(attr) && lexer && typeof (lexer[attr]) === 'string') {
                return lexer[attr]; //@attribute
            }
            if (stateMatches === null) {
                stateMatches = state.split('.');
                stateMatches.unshift(state);
            }
            if (!empty(s) && s < stateMatches.length) {
                return fixCase(lexer, stateMatches[s]); //$Sn
            }
            return '';
        });
    }
    exports.substituteMatches = substituteMatches;
    /**
     * Find the tokenizer rules for a specific state (i.e. next action)
     */
    function findRules(lexer, state) {
        while (state && state.length > 0) {
            var rules = lexer.tokenizer[state];
            if (rules) {
                return rules;
            }
            var idx = state.lastIndexOf('.');
            if (idx < 0) {
                state = null; // no further parent
            }
            else {
                state = state.substr(0, idx);
            }
        }
        return null;
    }
    exports.findRules = findRules;
    /**
     * Is a certain state defined? In contrast to 'findRules' this works on a ILexerMin.
     * This is used during compilation where we may know the defined states
     * but not yet whether the corresponding rules are correct.
     */
    function stateExists(lexer, state) {
        while (state && state.length > 0) {
            var exist = lexer.stateNames[state];
            if (exist) {
                return true;
            }
            var idx = state.lastIndexOf('.');
            if (idx < 0) {
                state = null; // no further parent
            }
            else {
                state = state.substr(0, idx);
            }
        }
        return false;
    }
    exports.stateExists = stateExists;
});

define(__m[88], __M([1,0,12,25]), function (require, exports, objects, monarchCommon) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    /*
     * Type helpers
     *
     * Note: this is just for sanity checks on the JSON description which is
     * helpful for the programmer. No checks are done anymore once the lexer is
     * already 'compiled and checked'.
     *
     */
    function isArrayOf(elemType, obj) {
        if (!obj) {
            return false;
        }
        if (!(Array.isArray(obj))) {
            return false;
        }
        var idx;
        for (idx in obj) {
            if (obj.hasOwnProperty(idx)) {
                if (!(elemType(obj[idx]))) {
                    return false;
                }
            }
        }
        return true;
    }
    function bool(prop, def, onerr) {
        if (typeof (prop) === 'boolean') {
            return prop;
        }
        if (onerr && (prop || def === undefined)) {
            onerr(); // type is wrong, or there is no default
        }
        return (def === undefined ? null : def);
    }
    function string(prop, def, onerr) {
        if (typeof (prop) === 'string') {
            return prop;
        }
        if (onerr && (prop || def === undefined)) {
            onerr(); // type is wrong, or there is no default
        }
        return (def === undefined ? null : def);
    }
    // Lexer helpers
    /**
     * Compiles a regular expression string, adding the 'i' flag if 'ignoreCase' is set.
     * Also replaces @\w+ or sequences with the content of the specified attribute
     */
    function compileRegExp(lexer, str) {
        if (typeof (str) !== 'string') {
            return null;
        }
        var n = 0;
        while (str.indexOf('@') >= 0 && n < 5) {
            n++;
            str = str.replace(/@(\w+)/g, function (s, attr) {
                var sub = '';
                if (typeof (lexer[attr]) === 'string') {
                    sub = lexer[attr];
                }
                else if (lexer[attr] && lexer[attr] instanceof RegExp) {
                    sub = lexer[attr].source;
                }
                else {
                    if (lexer[attr] === undefined) {
                        monarchCommon.throwError(lexer, 'language definition does not contain attribute \'' + attr + '\', used at: ' + str);
                    }
                    else {
                        monarchCommon.throwError(lexer, 'attribute reference \'' + attr + '\' must be a string, used at: ' + str);
                    }
                }
                return (monarchCommon.empty(sub) ? '' : '(?:' + sub + ')');
            });
        }
        return new RegExp(str, (lexer.ignoreCase ? 'i' : ''));
    }
    /**
     * Compiles guard functions for case matches.
     * This compiles 'cases' attributes into efficient match functions.
     *
     */
    function selectScrutinee(id, matches, state, num) {
        if (num < 0) {
            return id;
        }
        if (num < matches.length) {
            return matches[num];
        }
        if (num >= 100) {
            num = num - 100;
            var parts = state.split('.');
            parts.unshift(state);
            if (num < parts.length) {
                return parts[num];
            }
        }
        return null;
    }
    function createGuard(lexer, ruleName, tkey, val) {
        // get the scrutinee and pattern
        var scrut = -1; // -1: $!, 0-99: $n, 100+n: $Sn
        var oppat = tkey;
        var matches = tkey.match(/^\$(([sS]?)(\d\d?)|#)(.*)$/);
        if (matches) {
            if (matches[3]) {
                scrut = parseInt(matches[3]);
                if (matches[2]) {
                    scrut = scrut + 100; // if [sS] present
                }
            }
            oppat = matches[4];
        }
        // get operator
        var op = '~';
        var pat = oppat;
        if (!oppat || oppat.length === 0) {
            op = '!=';
            pat = '';
        }
        else if (/^\w*$/.test(pat)) {
            op = '==';
        }
        else {
            matches = oppat.match(/^(@|!@|~|!~|==|!=)(.*)$/);
            if (matches) {
                op = matches[1];
                pat = matches[2];
            }
        }
        // set the tester function
        var tester;
        // special case a regexp that matches just words
        if ((op === '~' || op === '!~') && /^(\w|\|)*$/.test(pat)) {
            var inWords = objects.createKeywordMatcher(pat.split('|'), lexer.ignoreCase);
            tester = function (s) { return (op === '~' ? inWords(s) : !inWords(s)); };
        }
        else if (op === '@' || op === '!@') {
            var words = lexer[pat];
            if (!words) {
                monarchCommon.throwError(lexer, 'the @ match target \'' + pat + '\' is not defined, in rule: ' + ruleName);
            }
            if (!(isArrayOf(function (elem) { return (typeof (elem) === 'string'); }, words))) {
                monarchCommon.throwError(lexer, 'the @ match target \'' + pat + '\' must be an array of strings, in rule: ' + ruleName);
            }
            var inWords = objects.createKeywordMatcher(words, lexer.ignoreCase);
            tester = function (s) { return (op === '@' ? inWords(s) : !inWords(s)); };
        }
        else if (op === '~' || op === '!~') {
            if (pat.indexOf('$') < 0) {
                // precompile regular expression
                var re = compileRegExp(lexer, '^' + pat + '$');
                tester = function (s) { return (op === '~' ? re.test(s) : !re.test(s)); };
            }
            else {
                tester = function (s, id, matches, state) {
                    var re = compileRegExp(lexer, '^' + monarchCommon.substituteMatches(lexer, pat, id, matches, state) + '$');
                    return re.test(s);
                };
            }
        }
        else {
            if (pat.indexOf('$') < 0) {
                var patx = monarchCommon.fixCase(lexer, pat);
                tester = function (s) { return (op === '==' ? s === patx : s !== patx); };
            }
            else {
                var patx = monarchCommon.fixCase(lexer, pat);
                tester = function (s, id, matches, state, eos) {
                    var patexp = monarchCommon.substituteMatches(lexer, patx, id, matches, state);
                    return (op === '==' ? s === patexp : s !== patexp);
                };
            }
        }
        // return the branch object
        if (scrut === -1) {
            return {
                name: tkey, value: val, test: function (id, matches, state, eos) {
                    return tester(id, id, matches, state, eos);
                }
            };
        }
        else {
            return {
                name: tkey, value: val, test: function (id, matches, state, eos) {
                    var scrutinee = selectScrutinee(id, matches, state, scrut);
                    return tester(!scrutinee ? '' : scrutinee, id, matches, state, eos);
                }
            };
        }
    }
    /**
     * Compiles an action: i.e. optimize regular expressions and case matches
     * and do many sanity checks.
     *
     * This is called only during compilation but if the lexer definition
     * contains user functions as actions (which is usually not allowed), then this
     * may be called during lexing. It is important therefore to compile common cases efficiently
     */
    function compileAction(lexer, ruleName, action) {
        if (!action) {
            return { token: '' };
        }
        else if (typeof (action) === 'string') {
            return action; // { token: action };
        }
        else if (action.token || action.token === '') {
            if (typeof (action.token) !== 'string') {
                monarchCommon.throwError(lexer, 'a \'token\' attribute must be of type string, in rule: ' + ruleName);
                return { token: '' };
            }
            else {
                // only copy specific typed fields (only happens once during compile Lexer)
                var newAction = { token: action.token };
                if (action.token.indexOf('$') >= 0) {
                    newAction.tokenSubst = true;
                }
                if (typeof (action.bracket) === 'string') {
                    if (action.bracket === '@open') {
                        newAction.bracket = monarchCommon.MonarchBracket.Open;
                    }
                    else if (action.bracket === '@close') {
                        newAction.bracket = monarchCommon.MonarchBracket.Close;
                    }
                    else {
                        monarchCommon.throwError(lexer, 'a \'bracket\' attribute must be either \'@open\' or \'@close\', in rule: ' + ruleName);
                    }
                }
                if (action.next) {
                    if (typeof (action.next) !== 'string') {
                        monarchCommon.throwError(lexer, 'the next state must be a string value in rule: ' + ruleName);
                    }
                    else {
                        var next = action.next;
                        if (!/^(@pop|@push|@popall)$/.test(next)) {
                            if (next[0] === '@') {
                                next = next.substr(1); // peel off starting @ sign
                            }
                            if (next.indexOf('$') < 0) {
                                if (!monarchCommon.stateExists(lexer, monarchCommon.substituteMatches(lexer, next, '', [], ''))) {
                                    monarchCommon.throwError(lexer, 'the next state \'' + action.next + '\' is not defined in rule: ' + ruleName);
                                }
                            }
                        }
                        newAction.next = next;
                    }
                }
                if (typeof (action.goBack) === 'number') {
                    newAction.goBack = action.goBack;
                }
                if (typeof (action.switchTo) === 'string') {
                    newAction.switchTo = action.switchTo;
                }
                if (typeof (action.log) === 'string') {
                    newAction.log = action.log;
                }
                if (typeof (action.nextEmbedded) === 'string') {
                    newAction.nextEmbedded = action.nextEmbedded;
                    lexer.usesEmbedded = true;
                }
                return newAction;
            }
        }
        else if (Array.isArray(action)) {
            var results = [];
            var idx;
            for (idx in action) {
                if (action.hasOwnProperty(idx)) {
                    results[idx] = compileAction(lexer, ruleName, action[idx]);
                }
            }
            return { group: results };
        }
        else if (action.cases) {
            // build an array of test cases
            var cases = [];
            // for each case, push a test function and result value
            var tkey;
            for (tkey in action.cases) {
                if (action.cases.hasOwnProperty(tkey)) {
                    var val = compileAction(lexer, ruleName, action.cases[tkey]);
                    // what kind of case
                    if (tkey === '@default' || tkey === '@' || tkey === '') {
                        cases.push({ test: null, value: val, name: tkey });
                    }
                    else if (tkey === '@eos') {
                        cases.push({ test: function (id, matches, state, eos) { return eos; }, value: val, name: tkey });
                    }
                    else {
                        cases.push(createGuard(lexer, ruleName, tkey, val)); // call separate function to avoid local variable capture
                    }
                }
            }
            // create a matching function
            var def = lexer.defaultToken;
            return {
                test: function (id, matches, state, eos) {
                    var idx;
                    for (idx in cases) {
                        if (cases.hasOwnProperty(idx)) {
                            var didmatch = (!cases[idx].test || cases[idx].test(id, matches, state, eos));
                            if (didmatch) {
                                return cases[idx].value;
                            }
                        }
                    }
                    return def;
                }
            };
        }
        else {
            monarchCommon.throwError(lexer, 'an action must be a string, an object with a \'token\' or \'cases\' attribute, or an array of actions; in rule: ' + ruleName);
            return '';
        }
    }
    /**
     * Helper class for creating matching rules
     */
    var Rule = (function () {
        function Rule(name) {
            this.regex = new RegExp('');
            this.action = { token: '' };
            this.matchOnlyAtLineStart = false;
            this.name = '';
            this.name = name;
        }
        Rule.prototype.setRegex = function (lexer, re) {
            var sregex;
            if (typeof (re) === 'string') {
                sregex = re;
            }
            else if (re instanceof RegExp) {
                sregex = re.source;
            }
            else {
                monarchCommon.throwError(lexer, 'rules must start with a match string or regular expression: ' + this.name);
            }
            this.matchOnlyAtLineStart = (sregex.length > 0 && sregex[0] === '^');
            this.name = this.name + ': ' + sregex;
            this.regex = compileRegExp(lexer, '^(?:' + (this.matchOnlyAtLineStart ? sregex.substr(1) : sregex) + ')');
        };
        Rule.prototype.setAction = function (lexer, act) {
            this.action = compileAction(lexer, this.name, act);
        };
        return Rule;
    }());
    /**
     * Compiles a json description function into json where all regular expressions,
     * case matches etc, are compiled and all include rules are expanded.
     * We also compile the bracket definitions, supply defaults, and do many sanity checks.
     * If the 'jsonStrict' parameter is 'false', we allow at certain locations
     * regular expression objects and functions that get called during lexing.
     * (Currently we have no samples that need this so perhaps we should always have
     * jsonStrict to true).
     */
    function compile(languageId, json) {
        if (!json || typeof (json) !== 'object') {
            throw new Error('Monarch: expecting a language definition object');
        }
        // Create our lexer
        var lexer = {};
        lexer.languageId = languageId;
        lexer.noThrow = false; // raise exceptions during compilation
        lexer.maxStack = 100;
        // Set standard fields: be defensive about types
        lexer.start = string(json.start);
        lexer.ignoreCase = bool(json.ignoreCase, false);
        lexer.tokenPostfix = string(json.tokenPostfix, '.' + lexer.languageId);
        lexer.defaultToken = string(json.defaultToken, 'source', function () { monarchCommon.throwError(lexer, 'the \'defaultToken\' must be a string'); });
        lexer.usesEmbedded = false; // becomes true if we find a nextEmbedded action
        // For calling compileAction later on
        var lexerMin = json;
        lexerMin.languageId = languageId;
        lexerMin.ignoreCase = lexer.ignoreCase;
        lexerMin.noThrow = lexer.noThrow;
        lexerMin.usesEmbedded = lexer.usesEmbedded;
        lexerMin.stateNames = json.tokenizer;
        lexerMin.defaultToken = lexer.defaultToken;
        // Compile an array of rules into newrules where RegExp objects are created.
        function addRules(state, newrules, rules) {
            var idx;
            for (idx in rules) {
                if (rules.hasOwnProperty(idx)) {
                    var rule = rules[idx];
                    var include = rule.include;
                    if (include) {
                        if (typeof (include) !== 'string') {
                            monarchCommon.throwError(lexer, 'an \'include\' attribute must be a string at: ' + state);
                        }
                        if (include[0] === '@') {
                            include = include.substr(1); // peel off starting @
                        }
                        if (!json.tokenizer[include]) {
                            monarchCommon.throwError(lexer, 'include target \'' + include + '\' is not defined at: ' + state);
                        }
                        addRules(state + '.' + include, newrules, json.tokenizer[include]);
                    }
                    else {
                        var newrule = new Rule(state);
                        // Set up new rule attributes
                        if (Array.isArray(rule) && rule.length >= 1 && rule.length <= 3) {
                            newrule.setRegex(lexerMin, rule[0]);
                            if (rule.length >= 3) {
                                if (typeof (rule[1]) === 'string') {
                                    newrule.setAction(lexerMin, { token: rule[1], next: rule[2] });
                                }
                                else if (typeof (rule[1]) === 'object') {
                                    var rule1 = rule[1];
                                    rule1.next = rule[2];
                                    newrule.setAction(lexerMin, rule1);
                                }
                                else {
                                    monarchCommon.throwError(lexer, 'a next state as the last element of a rule can only be given if the action is either an object or a string, at: ' + state);
                                }
                            }
                            else {
                                newrule.setAction(lexerMin, rule[1]);
                            }
                        }
                        else {
                            if (!rule.regex) {
                                monarchCommon.throwError(lexer, 'a rule must either be an array, or an object with a \'regex\' or \'include\' field at: ' + state);
                            }
                            if (rule.name) {
                                newrule.name = string(rule.name);
                            }
                            if (rule.matchOnlyAtStart) {
                                newrule.matchOnlyAtLineStart = bool(rule.matchOnlyAtLineStart);
                            }
                            newrule.setRegex(lexerMin, rule.regex);
                            newrule.setAction(lexerMin, rule.action);
                        }
                        newrules.push(newrule);
                    }
                }
            }
        }
        // compile the tokenizer rules
        if (!json.tokenizer || typeof (json.tokenizer) !== 'object') {
            monarchCommon.throwError(lexer, 'a language definition must define the \'tokenizer\' attribute as an object');
        }
        lexer.tokenizer = [];
        var key;
        for (key in json.tokenizer) {
            if (json.tokenizer.hasOwnProperty(key)) {
                if (!lexer.start) {
                    lexer.start = key;
                }
                var rules = json.tokenizer[key];
                lexer.tokenizer[key] = new Array();
                addRules('tokenizer.' + key, lexer.tokenizer[key], rules);
            }
        }
        lexer.usesEmbedded = lexerMin.usesEmbedded; // can be set during compileAction
        // Set simple brackets
        if (json.brackets) {
            if (!(Array.isArray(json.brackets))) {
                monarchCommon.throwError(lexer, 'the \'brackets\' attribute must be defined as an array');
            }
        }
        else {
            json.brackets = [
                { open: '{', close: '}', token: 'delimiter.curly' },
                { open: '[', close: ']', token: 'delimiter.square' },
                { open: '(', close: ')', token: 'delimiter.parenthesis' },
                { open: '<', close: '>', token: 'delimiter.angle' }];
        }
        var brackets = [];
        for (var bracketIdx in json.brackets) {
            if (json.brackets.hasOwnProperty(bracketIdx)) {
                var desc = json.brackets[bracketIdx];
                if (desc && Array.isArray(desc) && desc.length === 3) {
                    desc = { token: desc[2], open: desc[0], close: desc[1] };
                }
                if (desc.open === desc.close) {
                    monarchCommon.throwError(lexer, 'open and close brackets in a \'brackets\' attribute must be different: ' + desc.open +
                        '\n hint: use the \'bracket\' attribute if matching on equal brackets is required.');
                }
                if (typeof (desc.open) === 'string' && typeof (desc.token) === 'string') {
                    brackets.push({
                        token: string(desc.token) + lexer.tokenPostfix,
                        open: monarchCommon.fixCase(lexer, string(desc.open)),
                        close: monarchCommon.fixCase(lexer, string(desc.close))
                    });
                }
                else {
                    monarchCommon.throwError(lexer, 'every element in the \'brackets\' array must be a \'{open,close,token}\' object or array');
                }
            }
        }
        lexer.brackets = brackets;
        // Disable throw so the syntax highlighter goes, no matter what
        lexer.noThrow = true;
        return lexer;
    }
    exports.compile = compile;
});

define(__m[48], __M([1,0]), function (require, exports) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var NullState = (function () {
        function NullState(mode, stateData) {
            this.mode = mode;
            this.stateData = stateData;
        }
        NullState.prototype.clone = function () {
            var stateDataClone = (this.stateData ? this.stateData.clone() : null);
            return new NullState(this.mode, stateDataClone);
        };
        NullState.prototype.equals = function (other) {
            if (this.mode !== other.getMode()) {
                return false;
            }
            var otherStateData = other.getStateData();
            if (!this.stateData && !otherStateData) {
                return true;
            }
            if (this.stateData && otherStateData) {
                return this.stateData.equals(otherStateData);
            }
            return false;
        };
        NullState.prototype.getMode = function () {
            return this.mode;
        };
        NullState.prototype.tokenize = function (stream) {
            stream.advanceToEOS();
            return { type: '' };
        };
        NullState.prototype.getStateData = function () {
            return this.stateData;
        };
        NullState.prototype.setStateData = function (stateData) {
            this.stateData = stateData;
        };
        return NullState;
    }());
    exports.NullState = NullState;
    var NullMode = (function () {
        function NullMode() {
        }
        NullMode.prototype.getId = function () {
            return NullMode.ID;
        };
        NullMode.prototype.toSimplifiedMode = function () {
            return this;
        };
        NullMode.ID = 'vs.editor.modes.nullMode';
        return NullMode;
    }());
    exports.NullMode = NullMode;
    function nullTokenize(mode, buffer, state, deltaOffset, stopAtOffset) {
        if (deltaOffset === void 0) { deltaOffset = 0; }
        var tokens = [
            {
                startIndex: deltaOffset,
                type: ''
            }
        ];
        var modeTransitions = [
            {
                startIndex: deltaOffset,
                mode: mode
            }
        ];
        return {
            tokens: tokens,
            actualStopOffset: deltaOffset + buffer.length,
            endState: state,
            modeTransitions: modeTransitions
        };
    }
    exports.nullTokenize = nullTokenize;
});

define(__m[14], __M([1,0,3,12,18]), function (require, exports, strings, objects, modeTransition_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var Token = (function () {
        function Token(startIndex, type) {
            this.startIndex = startIndex;
            this.type = type;
        }
        Token.prototype.toString = function () {
            return '(' + this.startIndex + ', ' + this.type + ')';
        };
        return Token;
    }());
    exports.Token = Token;
    var LineTokens = (function () {
        function LineTokens(tokens, modeTransitions, actualStopOffset, endState) {
            this.tokens = tokens;
            this.modeTransitions = modeTransitions;
            this.actualStopOffset = actualStopOffset;
            this.endState = endState;
            this.retokenize = null;
        }
        return LineTokens;
    }());
    exports.LineTokens = LineTokens;
    function handleEvent(context, offset, runner) {
        var modeTransitions = context.modeTransitions;
        if (modeTransitions.length === 1) {
            return runner(modeTransitions[0].modeId, context, offset);
        }
        var modeIndex = modeTransition_1.ModeTransition.findIndexInSegmentsArray(modeTransitions, offset);
        var nestedMode = modeTransitions[modeIndex].mode;
        var modeStartIndex = modeTransitions[modeIndex].startIndex;
        var firstTokenInModeIndex = context.findIndexOfOffset(modeStartIndex);
        var nextCharacterAfterModeIndex = -1;
        var nextTokenAfterMode = -1;
        if (modeIndex + 1 < modeTransitions.length) {
            nextTokenAfterMode = context.findIndexOfOffset(modeTransitions[modeIndex + 1].startIndex);
            nextCharacterAfterModeIndex = context.getTokenStartIndex(nextTokenAfterMode);
        }
        else {
            nextTokenAfterMode = context.getTokenCount();
            nextCharacterAfterModeIndex = context.getLineContent().length;
        }
        var firstTokenCharacterOffset = context.getTokenStartIndex(firstTokenInModeIndex);
        var newCtx = new FilteredLineContext(context, nestedMode, firstTokenInModeIndex, nextTokenAfterMode, firstTokenCharacterOffset, nextCharacterAfterModeIndex);
        return runner(nestedMode.getId(), newCtx, offset - firstTokenCharacterOffset);
    }
    exports.handleEvent = handleEvent;
    var FilteredLineContext = (function () {
        function FilteredLineContext(actual, mode, firstTokenInModeIndex, nextTokenAfterMode, firstTokenCharacterOffset, nextCharacterAfterModeIndex) {
            this.modeTransitions = [new modeTransition_1.ModeTransition(0, mode)];
            this._actual = actual;
            this._firstTokenInModeIndex = firstTokenInModeIndex;
            this._nextTokenAfterMode = nextTokenAfterMode;
            this._firstTokenCharacterOffset = firstTokenCharacterOffset;
            this._nextCharacterAfterModeIndex = nextCharacterAfterModeIndex;
        }
        FilteredLineContext.prototype.getLineContent = function () {
            var actualLineContent = this._actual.getLineContent();
            return actualLineContent.substring(this._firstTokenCharacterOffset, this._nextCharacterAfterModeIndex);
        };
        FilteredLineContext.prototype.getTokenCount = function () {
            return this._nextTokenAfterMode - this._firstTokenInModeIndex;
        };
        FilteredLineContext.prototype.findIndexOfOffset = function (offset) {
            return this._actual.findIndexOfOffset(offset + this._firstTokenCharacterOffset) - this._firstTokenInModeIndex;
        };
        FilteredLineContext.prototype.getTokenStartIndex = function (tokenIndex) {
            return this._actual.getTokenStartIndex(tokenIndex + this._firstTokenInModeIndex) - this._firstTokenCharacterOffset;
        };
        FilteredLineContext.prototype.getTokenEndIndex = function (tokenIndex) {
            return this._actual.getTokenEndIndex(tokenIndex + this._firstTokenInModeIndex) - this._firstTokenCharacterOffset;
        };
        FilteredLineContext.prototype.getTokenType = function (tokenIndex) {
            return this._actual.getTokenType(tokenIndex + this._firstTokenInModeIndex);
        };
        FilteredLineContext.prototype.getTokenText = function (tokenIndex) {
            return this._actual.getTokenText(tokenIndex + this._firstTokenInModeIndex);
        };
        return FilteredLineContext;
    }());
    exports.FilteredLineContext = FilteredLineContext;
    var IGNORE_IN_TOKENS = /\b(comment|string|regex)\b/;
    function ignoreBracketsInToken(tokenType) {
        return IGNORE_IN_TOKENS.test(tokenType);
    }
    exports.ignoreBracketsInToken = ignoreBracketsInToken;
    // TODO@Martin: find a better home for this code:
    // TODO@Martin: modify suggestSupport to return a boolean if snippets should be presented or not
    //       and turn this into a real registry
    var SnippetsRegistry = (function () {
        function SnippetsRegistry() {
        }
        SnippetsRegistry.registerDefaultSnippets = function (modeId, snippets) {
            this._defaultSnippets[modeId] = (this._defaultSnippets[modeId] || []).concat(snippets);
        };
        SnippetsRegistry.registerSnippets = function (modeId, path, snippets) {
            var snippetsByMode = this._snippets[modeId];
            if (!snippetsByMode) {
                this._snippets[modeId] = snippetsByMode = {};
            }
            snippetsByMode[path] = snippets;
        };
        // the previous
        SnippetsRegistry.getNonWhitespacePrefix = function (model, position) {
            var line = model.getLineContent(position.lineNumber);
            var match = line.match(/[^\s]+$/);
            if (match) {
                return match[0];
            }
            return '';
        };
        SnippetsRegistry.getSnippets = function (model, position) {
            var word = model.getWordAtPosition(position);
            var currentWord = word ? word.word.substring(0, position.column - word.startColumn).toLowerCase() : '';
            var currentFullWord = SnippetsRegistry.getNonWhitespacePrefix(model, position).toLowerCase();
            var result = {
                currentWord: currentWord,
                incomplete: currentWord.length === 0,
                suggestions: []
            };
            var modeId = model.getModeId();
            var snippets = [];
            var snipppetsByMode = this._snippets[modeId];
            if (snipppetsByMode) {
                for (var s in snipppetsByMode) {
                    snippets = snippets.concat(snipppetsByMode[s]);
                }
            }
            var defaultSnippets = this._defaultSnippets[modeId];
            if (defaultSnippets) {
                snippets = snippets.concat(defaultSnippets);
            }
            // to avoid that snippets are too prominent in the intellisense proposals:
            // enforce that current word is matched or the position is after a whitespace
            snippets.forEach(function (p) {
                if (currentWord.length === 0 && currentFullWord.length === 0) {
                }
                else {
                    var label = p.label.toLowerCase();
                    // force that the current word or full word matches with the snippet prefix
                    if (currentWord.length > 0 && strings.startsWith(label, currentWord)) {
                    }
                    else if (currentFullWord.length > currentWord.length && strings.startsWith(label, currentFullWord)) {
                        p = objects.clone(p);
                        p.overwriteBefore = currentFullWord.length;
                    }
                    else {
                        return;
                    }
                }
                result.suggestions.push(p);
            });
            // if (result.suggestions.length > 0) {
            // 	if (word) {
            // 		// Push also the current word as first suggestion, to avoid unexpected snippet acceptance on Enter.
            // 		result.suggestions = result.suggestions.slice(0);
            // 		result.suggestions.unshift({
            // 			codeSnippet: word.word,
            // 			label: word.word,
            // 			type: 'text'
            // 		});
            // 	}
            // 	result.incomplete = true;
            // }
            return result;
        };
        SnippetsRegistry._defaultSnippets = Object.create(null);
        SnippetsRegistry._snippets = Object.create(null);
        return SnippetsRegistry;
    }());
    exports.SnippetsRegistry = SnippetsRegistry;
});

define(__m[94], __M([1,0,14]), function (require, exports, supports_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var CharacterPairSupport = (function () {
        function CharacterPairSupport(registry, modeId, config) {
            this._registry = registry;
            this._modeId = modeId;
            this._autoClosingPairs = config.autoClosingPairs;
            if (!this._autoClosingPairs) {
                this._autoClosingPairs = config.brackets ? config.brackets.map(function (b) { return ({ open: b[0], close: b[1] }); }) : [];
            }
            this._surroundingPairs = config.surroundingPairs || this._autoClosingPairs;
        }
        CharacterPairSupport.prototype.getAutoClosingPairs = function () {
            return this._autoClosingPairs;
        };
        CharacterPairSupport.prototype.shouldAutoClosePair = function (character, context, offset) {
            var _this = this;
            return supports_1.handleEvent(context, offset, function (nestedModeId, context, offset) {
                if (_this._modeId === nestedModeId) {
                    // Always complete on empty line
                    if (context.getTokenCount() === 0) {
                        return true;
                    }
                    var tokenIndex = context.findIndexOfOffset(offset - 1);
                    var tokenType = context.getTokenType(tokenIndex);
                    for (var i = 0; i < _this._autoClosingPairs.length; ++i) {
                        if (_this._autoClosingPairs[i].open === character) {
                            if (_this._autoClosingPairs[i].notIn) {
                                for (var notInIndex = 0; notInIndex < _this._autoClosingPairs[i].notIn.length; ++notInIndex) {
                                    if (tokenType.indexOf(_this._autoClosingPairs[i].notIn[notInIndex]) > -1) {
                                        return false;
                                    }
                                }
                            }
                            break;
                        }
                    }
                    return true;
                }
                var characterPairSupport = _this._registry.getCharacterPairSupport(nestedModeId);
                if (characterPairSupport) {
                    return characterPairSupport.shouldAutoClosePair(character, context, offset);
                }
                return null;
            });
        };
        CharacterPairSupport.prototype.getSurroundingPairs = function () {
            return this._surroundingPairs;
        };
        return CharacterPairSupport;
    }());
    exports.CharacterPairSupport = CharacterPairSupport;
});

define(__m[28], __M([1,0,3,29]), function (require, exports, strings, range_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var RichEditBrackets = (function () {
        function RichEditBrackets(modeId, brackets) {
            var _this = this;
            this.brackets = brackets.map(function (b) {
                return {
                    modeId: modeId,
                    open: b[0],
                    close: b[1],
                    forwardRegex: getRegexForBracketPair({ open: b[0], close: b[1] }),
                    reversedRegex: getReversedRegexForBracketPair({ open: b[0], close: b[1] })
                };
            });
            this.forwardRegex = getRegexForBrackets(this.brackets);
            this.reversedRegex = getReversedRegexForBrackets(this.brackets);
            this.textIsBracket = {};
            this.textIsOpenBracket = {};
            this.maxBracketLength = 0;
            this.brackets.forEach(function (b) {
                _this.textIsBracket[b.open] = b;
                _this.textIsBracket[b.close] = b;
                _this.textIsOpenBracket[b.open] = true;
                _this.textIsOpenBracket[b.close] = false;
                _this.maxBracketLength = Math.max(_this.maxBracketLength, b.open.length);
                _this.maxBracketLength = Math.max(_this.maxBracketLength, b.close.length);
            });
        }
        return RichEditBrackets;
    }());
    exports.RichEditBrackets = RichEditBrackets;
    function once(keyFn, computeFn) {
        var cache = {};
        return function (input) {
            var key = keyFn(input);
            if (!cache.hasOwnProperty(key)) {
                cache[key] = computeFn(input);
            }
            return cache[key];
        };
    }
    var getRegexForBracketPair = once(function (input) { return (input.open + ";" + input.close); }, function (input) {
        return createOrRegex([input.open, input.close]);
    });
    var getReversedRegexForBracketPair = once(function (input) { return (input.open + ";" + input.close); }, function (input) {
        return createOrRegex([toReversedString(input.open), toReversedString(input.close)]);
    });
    var getRegexForBrackets = once(function (input) { return input.map(function (b) { return (b.open + ";" + b.close); }).join(';'); }, function (input) {
        var pieces = [];
        input.forEach(function (b) {
            pieces.push(b.open);
            pieces.push(b.close);
        });
        return createOrRegex(pieces);
    });
    var getReversedRegexForBrackets = once(function (input) { return input.map(function (b) { return (b.open + ";" + b.close); }).join(';'); }, function (input) {
        var pieces = [];
        input.forEach(function (b) {
            pieces.push(toReversedString(b.open));
            pieces.push(toReversedString(b.close));
        });
        return createOrRegex(pieces);
    });
    function createOrRegex(pieces) {
        var regexStr = "(" + pieces.map(strings.escapeRegExpCharacters).join(')|(') + ")";
        return strings.createRegExp(regexStr, true, false, false, false);
    }
    function toReversedString(str) {
        var reversedStr = '';
        for (var i = str.length - 1; i >= 0; i--) {
            reversedStr += str.charAt(i);
        }
        return reversedStr;
    }
    var BracketsUtils = (function () {
        function BracketsUtils() {
        }
        BracketsUtils._findPrevBracketInText = function (reversedBracketRegex, lineNumber, reversedText, offset) {
            var m = reversedText.match(reversedBracketRegex);
            if (!m) {
                return null;
            }
            var matchOffset = reversedText.length - m.index;
            var matchLength = m[0].length;
            var absoluteMatchOffset = offset + matchOffset;
            return new range_1.Range(lineNumber, absoluteMatchOffset - matchLength + 1, lineNumber, absoluteMatchOffset + 1);
        };
        BracketsUtils.findPrevBracketInToken = function (reversedBracketRegex, lineNumber, lineText, currentTokenStart, currentTokenEnd) {
            // Because JS does not support backwards regex search, we search forwards in a reversed string with a reversed regex ;)
            var currentTokenReversedText = '';
            for (var index = currentTokenEnd - 1; index >= currentTokenStart; index--) {
                currentTokenReversedText += lineText.charAt(index);
            }
            return this._findPrevBracketInText(reversedBracketRegex, lineNumber, currentTokenReversedText, currentTokenStart);
        };
        BracketsUtils.findNextBracketInText = function (bracketRegex, lineNumber, text, offset) {
            var m = text.match(bracketRegex);
            if (!m) {
                return null;
            }
            var matchOffset = m.index;
            var matchLength = m[0].length;
            var absoluteMatchOffset = offset + matchOffset;
            return new range_1.Range(lineNumber, absoluteMatchOffset + 1, lineNumber, absoluteMatchOffset + 1 + matchLength);
        };
        BracketsUtils.findNextBracketInToken = function (bracketRegex, lineNumber, lineText, currentTokenStart, currentTokenEnd) {
            var currentTokenText = lineText.substring(currentTokenStart, currentTokenEnd);
            return this.findNextBracketInText(bracketRegex, lineNumber, currentTokenText, currentTokenStart);
        };
        return BracketsUtils;
    }());
    exports.BracketsUtils = BracketsUtils;
});

define(__m[99], __M([1,0,3,14,28]), function (require, exports, strings, supports_1, richEditBrackets_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var BracketElectricCharacterSupport = (function () {
        function BracketElectricCharacterSupport(registry, modeId, brackets, contribution) {
            this._registry = registry;
            this._modeId = modeId;
            this.contribution = contribution || {};
            this.brackets = new Brackets(modeId, brackets, this.contribution.docComment);
        }
        BracketElectricCharacterSupport.prototype.getElectricCharacters = function () {
            if (Array.isArray(this.contribution.embeddedElectricCharacters)) {
                return this.contribution.embeddedElectricCharacters.concat(this.brackets.getElectricCharacters());
            }
            return this.brackets.getElectricCharacters();
        };
        BracketElectricCharacterSupport.prototype.onElectricCharacter = function (context, offset) {
            var _this = this;
            return supports_1.handleEvent(context, offset, function (nestedModeId, context, offset) {
                if (_this._modeId === nestedModeId) {
                    return _this.brackets.onElectricCharacter(context, offset);
                }
                var electricCharacterSupport = _this._registry.getElectricCharacterSupport(nestedModeId);
                if (electricCharacterSupport) {
                    return electricCharacterSupport.onElectricCharacter(context, offset);
                }
                return null;
            });
        };
        return BracketElectricCharacterSupport;
    }());
    exports.BracketElectricCharacterSupport = BracketElectricCharacterSupport;
    var Brackets = (function () {
        function Brackets(modeId, richEditBrackets, docComment) {
            if (docComment === void 0) { docComment = null; }
            this._modeId = modeId;
            this._richEditBrackets = richEditBrackets;
            this._docComment = docComment ? docComment : null;
        }
        Brackets.prototype.getElectricCharacters = function () {
            var result = [];
            if (this._richEditBrackets) {
                for (var i = 0, len = this._richEditBrackets.brackets.length; i < len; i++) {
                    var bracketPair = this._richEditBrackets.brackets[i];
                    var lastChar = bracketPair.close.charAt(bracketPair.close.length - 1);
                    result.push(lastChar);
                }
            }
            // Doc comments
            if (this._docComment) {
                result.push(this._docComment.open.charAt(this._docComment.open.length - 1));
            }
            // Filter duplicate entries
            result = result.filter(function (item, pos, array) {
                return array.indexOf(item) === pos;
            });
            return result;
        };
        Brackets.prototype.onElectricCharacter = function (context, offset) {
            if (context.getTokenCount() === 0) {
                return null;
            }
            return (this._onElectricCharacterDocComment(context, offset) ||
                this._onElectricCharacterStandardBrackets(context, offset));
        };
        Brackets.prototype.containsTokenTypes = function (fullTokenSpec, tokensToLookFor) {
            var array = tokensToLookFor.split('.');
            for (var i = 0; i < array.length; ++i) {
                if (fullTokenSpec.indexOf(array[i]) < 0) {
                    return false;
                }
            }
            return true;
        };
        Brackets.prototype._onElectricCharacterStandardBrackets = function (context, offset) {
            if (!this._richEditBrackets || this._richEditBrackets.brackets.length === 0) {
                return null;
            }
            var reversedBracketRegex = this._richEditBrackets.reversedRegex;
            var lineText = context.getLineContent();
            var tokenIndex = context.findIndexOfOffset(offset);
            var tokenStart = context.getTokenStartIndex(tokenIndex);
            var tokenEnd = offset + 1;
            var firstNonWhitespaceIndex = strings.firstNonWhitespaceIndex(context.getLineContent());
            if (firstNonWhitespaceIndex !== -1 && firstNonWhitespaceIndex < tokenStart) {
                return null;
            }
            if (!supports_1.ignoreBracketsInToken(context.getTokenType(tokenIndex))) {
                var r = richEditBrackets_1.BracketsUtils.findPrevBracketInToken(reversedBracketRegex, 1, lineText, tokenStart, tokenEnd);
                if (r) {
                    var text = lineText.substring(r.startColumn - 1, r.endColumn - 1);
                    var isOpen = this._richEditBrackets.textIsOpenBracket[text];
                    if (!isOpen) {
                        return {
                            matchOpenBracket: text
                        };
                    }
                }
            }
            return null;
        };
        Brackets.prototype._onElectricCharacterDocComment = function (context, offset) {
            // We only auto-close, so do nothing if there is no closing part.
            if (!this._docComment || !this._docComment.close) {
                return null;
            }
            var line = context.getLineContent();
            var char = line[offset];
            // See if the right electric character was pressed
            if (char !== this._docComment.open.charAt(this._docComment.open.length - 1)) {
                return null;
            }
            // If this line already contains the closing tag, do nothing.
            if (line.indexOf(this._docComment.close, offset) >= 0) {
                return null;
            }
            // If we're not in a documentation comment, do nothing.
            var lastTokenIndex = context.findIndexOfOffset(offset);
            if (!this.containsTokenTypes(context.getTokenType(lastTokenIndex), this._docComment.scope)) {
                return null;
            }
            if (line.substring(context.getTokenStartIndex(lastTokenIndex), offset + 1 /* include electric char*/) !== this._docComment.open) {
                return null;
            }
            return { appendText: this._docComment.close };
        };
        return Brackets;
    }());
    exports.Brackets = Brackets;
});

define(__m[53], __M([1,0,62,48,14]), function (require, exports, lineStream_1, nullMode_1, supports_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    function isFunction(something) {
        return typeof something === 'function';
    }
    var TokenizationSupport = (function () {
        function TokenizationSupport(mode, customization, supportsNestedModes) {
            this._mode = mode;
            this.customization = customization;
            this.supportsNestedModes = supportsNestedModes;
            this._embeddedModesListeners = {};
            if (this.supportsNestedModes) {
                if (!this._mode.setTokenizationSupport) {
                    throw new Error('Cannot be a mode with nested modes unless I can emit a tokenizationSupport changed event!');
                }
            }
            this.defaults = {
                enterNestedMode: !isFunction(customization.enterNestedMode),
                getNestedMode: !isFunction(customization.getNestedMode),
                getNestedModeInitialState: !isFunction(customization.getNestedModeInitialState),
                getLeavingNestedModeData: !isFunction(customization.getLeavingNestedModeData),
                onReturningFromNestedMode: !isFunction(customization.onReturningFromNestedMode)
            };
        }
        TokenizationSupport.prototype.dispose = function () {
            for (var listener in this._embeddedModesListeners) {
                this._embeddedModesListeners[listener].dispose();
                delete this._embeddedModesListeners[listener];
            }
        };
        TokenizationSupport.prototype.getInitialState = function () {
            return this.customization.getInitialState();
        };
        TokenizationSupport.prototype.tokenize = function (line, state, deltaOffset, stopAtOffset) {
            if (deltaOffset === void 0) { deltaOffset = 0; }
            if (stopAtOffset === void 0) { stopAtOffset = deltaOffset + line.length; }
            if (state.getMode() !== this._mode) {
                return this._nestedTokenize(line, state, deltaOffset, stopAtOffset, [], []);
            }
            else {
                return this._myTokenize(line, state, deltaOffset, stopAtOffset, [], []);
            }
        };
        /**
         * Precondition is: nestedModeState.getMode() !== this
         * This means we are in a nested mode when parsing starts on this line.
         */
        TokenizationSupport.prototype._nestedTokenize = function (buffer, nestedModeState, deltaOffset, stopAtOffset, prependTokens, prependModeTransitions) {
            var myStateBeforeNestedMode = nestedModeState.getStateData();
            var leavingNestedModeData = this.getLeavingNestedModeData(buffer, myStateBeforeNestedMode);
            // Be sure to give every embedded mode the
            // opportunity to leave nested mode.
            // i.e. Don't go straight to the most nested mode
            var stepOnceNestedState = nestedModeState;
            while (stepOnceNestedState.getStateData() && stepOnceNestedState.getStateData().getMode() !== this._mode) {
                stepOnceNestedState = stepOnceNestedState.getStateData();
            }
            var nestedMode = stepOnceNestedState.getMode();
            if (!leavingNestedModeData) {
                // tokenization will not leave nested mode
                var result;
                if (nestedMode.tokenizationSupport) {
                    result = nestedMode.tokenizationSupport.tokenize(buffer, nestedModeState, deltaOffset, stopAtOffset);
                }
                else {
                    // The nested mode doesn't have tokenization support,
                    // unfortunatelly this means we have to fake it
                    result = nullMode_1.nullTokenize(nestedMode, buffer, nestedModeState, deltaOffset);
                }
                result.tokens = prependTokens.concat(result.tokens);
                result.modeTransitions = prependModeTransitions.concat(result.modeTransitions);
                return result;
            }
            var nestedModeBuffer = leavingNestedModeData.nestedModeBuffer;
            if (nestedModeBuffer.length > 0) {
                // Tokenize with the nested mode
                var nestedModeLineTokens;
                if (nestedMode.tokenizationSupport) {
                    nestedModeLineTokens = nestedMode.tokenizationSupport.tokenize(nestedModeBuffer, nestedModeState, deltaOffset, stopAtOffset);
                }
                else {
                    // The nested mode doesn't have tokenization support,
                    // unfortunatelly this means we have to fake it
                    nestedModeLineTokens = nullMode_1.nullTokenize(nestedMode, nestedModeBuffer, nestedModeState, deltaOffset);
                }
                // Save last state of nested mode
                nestedModeState = nestedModeLineTokens.endState;
                // Prepend nested mode's result to our result
                prependTokens = prependTokens.concat(nestedModeLineTokens.tokens);
                prependModeTransitions = prependModeTransitions.concat(nestedModeLineTokens.modeTransitions);
            }
            var bufferAfterNestedMode = leavingNestedModeData.bufferAfterNestedMode;
            var myStateAfterNestedMode = leavingNestedModeData.stateAfterNestedMode;
            myStateAfterNestedMode.setStateData(myStateBeforeNestedMode.getStateData());
            this.onReturningFromNestedMode(myStateAfterNestedMode, nestedModeState);
            return this._myTokenize(bufferAfterNestedMode, myStateAfterNestedMode, deltaOffset + nestedModeBuffer.length, stopAtOffset, prependTokens, prependModeTransitions);
        };
        /**
         * Precondition is: state.getMode() === this
         * This means we are in the current mode when parsing starts on this line.
         */
        TokenizationSupport.prototype._myTokenize = function (buffer, myState, deltaOffset, stopAtOffset, prependTokens, prependModeTransitions) {
            var _this = this;
            var lineStream = new lineStream_1.LineStream(buffer);
            var tokenResult, beforeTokenizeStreamPos;
            var previousType = null;
            var retokenize = null;
            myState = myState.clone();
            if (prependModeTransitions.length <= 0 || prependModeTransitions[prependModeTransitions.length - 1].mode !== this._mode) {
                // Avoid transitioning to the same mode (this can happen in case of empty embedded modes)
                prependModeTransitions.push({
                    startIndex: deltaOffset,
                    mode: this._mode
                });
            }
            var maxPos = Math.min(stopAtOffset - deltaOffset, buffer.length);
            while (lineStream.pos() < maxPos) {
                beforeTokenizeStreamPos = lineStream.pos();
                do {
                    tokenResult = myState.tokenize(lineStream);
                    if (tokenResult === null || tokenResult === undefined ||
                        ((tokenResult.type === undefined || tokenResult.type === null) &&
                            (tokenResult.nextState === undefined || tokenResult.nextState === null))) {
                        throw new Error('Tokenizer must return a valid state');
                    }
                    if (tokenResult.nextState) {
                        tokenResult.nextState.setStateData(myState.getStateData());
                        myState = tokenResult.nextState;
                    }
                    if (lineStream.pos() <= beforeTokenizeStreamPos) {
                        throw new Error('Stream did not advance while tokenizing. Mode id is ' + this._mode.getId() + ' (stuck at token type: "' + tokenResult.type + '", prepend tokens: "' + (prependTokens.map(function (t) { return t.type; }).join(',')) + '").');
                    }
                } while (!tokenResult.type && tokenResult.type !== '');
                if (previousType !== tokenResult.type || tokenResult.dontMergeWithPrev || previousType === null) {
                    prependTokens.push(new supports_1.Token(beforeTokenizeStreamPos + deltaOffset, tokenResult.type));
                }
                previousType = tokenResult.type;
                if (this.supportsNestedModes && this.enterNestedMode(myState)) {
                    var currentEmbeddedLevels = this._getEmbeddedLevel(myState);
                    if (currentEmbeddedLevels < TokenizationSupport.MAX_EMBEDDED_LEVELS) {
                        var nestedModeState = this.getNestedModeInitialState(myState);
                        // Re-emit tokenizationSupport change events from all modes that I ever embedded
                        var embeddedMode = nestedModeState.state.getMode();
                        if (typeof embeddedMode.addSupportChangedListener === 'function' && !this._embeddedModesListeners.hasOwnProperty(embeddedMode.getId())) {
                            var emitting = false;
                            this._embeddedModesListeners[embeddedMode.getId()] = embeddedMode.addSupportChangedListener(function (e) {
                                if (emitting) {
                                    return;
                                }
                                if (e.tokenizationSupport) {
                                    emitting = true;
                                    _this._mode.setTokenizationSupport(function (mode) {
                                        return mode.tokenizationSupport;
                                    });
                                    emitting = false;
                                }
                            });
                        }
                        if (!lineStream.eos()) {
                            // There is content from the embedded mode
                            var restOfBuffer = buffer.substr(lineStream.pos());
                            var result = this._nestedTokenize(restOfBuffer, nestedModeState.state, deltaOffset + lineStream.pos(), stopAtOffset, prependTokens, prependModeTransitions);
                            result.retokenize = result.retokenize || nestedModeState.missingModePromise;
                            return result;
                        }
                        else {
                            // Transition to the nested mode state
                            myState = nestedModeState.state;
                            retokenize = nestedModeState.missingModePromise;
                        }
                    }
                }
            }
            return {
                tokens: prependTokens,
                actualStopOffset: lineStream.pos() + deltaOffset,
                modeTransitions: prependModeTransitions,
                endState: myState,
                retokenize: retokenize
            };
        };
        TokenizationSupport.prototype._getEmbeddedLevel = function (state) {
            var result = -1;
            while (state) {
                result++;
                state = state.getStateData();
            }
            return result;
        };
        TokenizationSupport.prototype.enterNestedMode = function (state) {
            if (this.defaults.enterNestedMode) {
                return false;
            }
            return this.customization.enterNestedMode(state);
        };
        TokenizationSupport.prototype.getNestedMode = function (state) {
            if (this.defaults.getNestedMode) {
                return null;
            }
            return this.customization.getNestedMode(state);
        };
        TokenizationSupport._validatedNestedMode = function (input) {
            var mode = new nullMode_1.NullMode(), missingModePromise = null;
            if (input && input.mode) {
                mode = input.mode;
            }
            if (input && input.missingModePromise) {
                missingModePromise = input.missingModePromise;
            }
            return {
                mode: mode,
                missingModePromise: missingModePromise
            };
        };
        TokenizationSupport.prototype.getNestedModeInitialState = function (state) {
            if (this.defaults.getNestedModeInitialState) {
                var nestedMode = TokenizationSupport._validatedNestedMode(this.getNestedMode(state));
                var missingModePromise = nestedMode.missingModePromise;
                var nestedModeState;
                if (nestedMode.mode.tokenizationSupport) {
                    nestedModeState = nestedMode.mode.tokenizationSupport.getInitialState();
                }
                else {
                    nestedModeState = new nullMode_1.NullState(nestedMode.mode, null);
                }
                nestedModeState.setStateData(state);
                return {
                    state: nestedModeState,
                    missingModePromise: missingModePromise
                };
            }
            return this.customization.getNestedModeInitialState(state);
        };
        TokenizationSupport.prototype.getLeavingNestedModeData = function (line, state) {
            if (this.defaults.getLeavingNestedModeData) {
                return null;
            }
            return this.customization.getLeavingNestedModeData(line, state);
        };
        TokenizationSupport.prototype.onReturningFromNestedMode = function (myStateAfterNestedMode, lastNestedModeState) {
            if (this.defaults.onReturningFromNestedMode) {
                return null;
            }
            return this.customization.onReturningFromNestedMode(myStateAfterNestedMode, lastNestedModeState);
        };
        TokenizationSupport.MAX_EMBEDDED_LEVELS = 5;
        return TokenizationSupport;
    }());
    exports.TokenizationSupport = TokenizationSupport;
});






define(__m[105], __M([1,0,39,62,25,53]), function (require, exports, abstractState_1, lineStream_1, monarchCommon, tokenizationSupport_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    /**
     * The MonarchLexer class implements a monaco lexer that highlights source code.
     * It takes a compiled lexer to guide the tokenizer and maintains a stack of
     * lexer states.
     */
    var MonarchLexer = (function (_super) {
        __extends(MonarchLexer, _super);
        function MonarchLexer(mode, modeService, lexer, stack, embeddedMode) {
            _super.call(this, mode);
            this.id = MonarchLexer.ID++; // for debugging, assigns unique id to each instance
            this.modeService = modeService;
            this.lexer = lexer; // (compiled) lexer description
            this.stack = (stack ? stack : [lexer.start]); // stack of states
            this.embeddedMode = (embeddedMode ? embeddedMode : null); // are we scanning an embedded section?
            // did we encounter an embedded start on this line?
            // no need for cloning or equality since it is used only within a line
            this.embeddedEntered = false;
            // regular expression group matching
            // these never need cloning or equality since they are only used within a line match
            this.groupActions = null;
            this.groupMatches = null;
            this.groupMatched = null;
            this.groupRule = null;
        }
        MonarchLexer.prototype.makeClone = function () {
            return new MonarchLexer(this.getMode(), this.modeService, this.lexer, this.stack.slice(0), this.embeddedMode);
        };
        MonarchLexer.prototype.equals = function (other) {
            if (!_super.prototype.equals.call(this, other)) {
                return false;
            }
            if (!(other instanceof MonarchLexer)) {
                return false;
            }
            var otherm = other;
            if ((this.stack.length !== otherm.stack.length) || (this.lexer.languageId !== otherm.lexer.languageId) ||
                (this.embeddedMode !== otherm.embeddedMode)) {
                return false;
            }
            var idx;
            for (idx in this.stack) {
                if (this.stack.hasOwnProperty(idx)) {
                    if (this.stack[idx] !== otherm.stack[idx]) {
                        return false;
                    }
                }
            }
            return true;
        };
        /**
         * The main tokenizer: this function gets called by monaco to tokenize lines
         * Note: we don't want to raise exceptions here and always keep going..
         *
         * TODO: there are many optimizations possible here for the common cases
         * but for now I concentrated on functionality and correctness.
         */
        MonarchLexer.prototype.tokenize = function (stream, noConsumeIsOk) {
            var stackLen0 = this.stack.length; // these are saved to check progress
            var groupLen0 = 0;
            var state = this.stack[0]; // the current state
            this.embeddedEntered = false;
            var matches = null;
            var matched = null;
            var action = null;
            var next = null;
            var rule = null;
            // check if we need to process group matches first
            if (this.groupActions) {
                groupLen0 = this.groupActions.length;
                matches = this.groupMatches;
                matched = this.groupMatched.shift();
                action = this.groupActions.shift();
                rule = this.groupRule;
                // cleanup if necessary
                if (this.groupActions.length === 0) {
                    this.groupActions = null;
                    this.groupMatches = null;
                    this.groupMatched = null;
                    this.groupRule = null;
                }
            }
            else {
                // nothing to do
                if (stream.eos()) {
                    return { type: '' };
                }
                // get the entire line
                var line = stream.advanceToEOS();
                stream.goBack(line.length);
                // get the rules for this state
                var rules = this.lexer.tokenizer[state];
                if (!rules) {
                    rules = monarchCommon.findRules(this.lexer, state); // do parent matching
                }
                if (!rules) {
                    monarchCommon.throwError(this.lexer, 'tokenizer state is not defined: ' + state);
                }
                else {
                    // try each rule until we match
                    rule = null;
                    var pos = stream.pos();
                    var idx;
                    for (idx in rules) {
                        if (rules.hasOwnProperty(idx)) {
                            rule = rules[idx];
                            if (pos === 0 || !rule.matchOnlyAtLineStart) {
                                matches = line.match(rule.regex);
                                if (matches) {
                                    matched = matches[0];
                                    action = rule.action;
                                    break;
                                }
                            }
                        }
                    }
                }
            }
            // We matched 'rule' with 'matches' and 'action'
            if (!matches) {
                matches = [''];
                matched = '';
            }
            if (!action) {
                // bad: we didn't match anything, and there is no action to take
                // we need to advance the stream or we get progress trouble
                if (!stream.eos()) {
                    matches = [stream.peek()];
                    matched = matches[0];
                }
                action = this.lexer.defaultToken;
            }
            // advance stream
            stream.advance(matched.length);
            // maybe call action function (used for 'cases')
            while (action.test) {
                var callres = action.test(matched, matches, state, stream.eos());
                action = callres;
            }
            // set the result: either a string or an array of actions
            var result = null;
            if (typeof (action) === 'string' || Array.isArray(action)) {
                result = action;
            }
            else if (action.group) {
                result = action.group;
            }
            else if (action.token !== null && action.token !== undefined) {
                result = action.token;
                // do $n replacements?
                if (action.tokenSubst) {
                    result = monarchCommon.substituteMatches(this.lexer, result, matched, matches, state);
                }
                // enter embedded mode?
                if (action.nextEmbedded) {
                    if (action.nextEmbedded === '@pop') {
                        if (!this.embeddedMode) {
                            monarchCommon.throwError(this.lexer, 'cannot pop embedded mode if not inside one');
                        }
                        this.embeddedMode = null;
                    }
                    else if (this.embeddedMode) {
                        monarchCommon.throwError(this.lexer, 'cannot enter embedded mode from within an embedded mode');
                    }
                    else {
                        this.embeddedMode = monarchCommon.substituteMatches(this.lexer, action.nextEmbedded, matched, matches, state);
                        // substitute language alias to known modes to support syntax highlighting
                        var embeddedMode = this.modeService.getModeIdForLanguageName(this.embeddedMode);
                        if (this.embeddedMode && embeddedMode) {
                            this.embeddedMode = embeddedMode;
                        }
                        this.embeddedEntered = true;
                    }
                }
                // state transformations
                if (action.goBack) {
                    stream.goBack(action.goBack);
                }
                if (action.switchTo && typeof action.switchTo === 'string') {
                    var nextState = monarchCommon.substituteMatches(this.lexer, action.switchTo, matched, matches, state); // switch state without a push...
                    if (nextState[0] === '@') {
                        nextState = nextState.substr(1); // peel off starting '@'
                    }
                    if (!monarchCommon.findRules(this.lexer, nextState)) {
                        monarchCommon.throwError(this.lexer, 'trying to switch to a state \'' + nextState + '\' that is undefined in rule: ' + rule.name);
                    }
                    else {
                        this.stack[0] = nextState;
                    }
                    next = null;
                }
                else if (action.transform && typeof action.transform === 'function') {
                    this.stack = action.transform(this.stack); // if you need to do really funky stuff...
                    next = null;
                }
                else if (action.next) {
                    if (action.next === '@push') {
                        if (this.stack.length >= this.lexer.maxStack) {
                            monarchCommon.throwError(this.lexer, 'maximum tokenizer stack size reached: [' +
                                this.stack[0] + ',' + this.stack[1] + ',...,' +
                                this.stack[this.stack.length - 2] + ',' + this.stack[this.stack.length - 1] + ']');
                        }
                        else {
                            this.stack.unshift(state);
                        }
                    }
                    else if (action.next === '@pop') {
                        if (this.stack.length <= 1) {
                            monarchCommon.throwError(this.lexer, 'trying to pop an empty stack in rule: ' + rule.name);
                        }
                        else {
                            this.stack.shift();
                        }
                    }
                    else if (action.next === '@popall') {
                        if (this.stack.length > 1) {
                            this.stack = [this.stack[this.stack.length - 1]];
                        }
                    }
                    else {
                        var nextState = monarchCommon.substituteMatches(this.lexer, action.next, matched, matches, state);
                        if (nextState[0] === '@') {
                            nextState = nextState.substr(1); // peel off starting '@'
                        }
                        if (!monarchCommon.findRules(this.lexer, nextState)) {
                            monarchCommon.throwError(this.lexer, 'trying to set a next state \'' + nextState + '\' that is undefined in rule: ' + rule.name);
                        }
                        else {
                            this.stack.unshift(nextState);
                        }
                    }
                }
                if (action.log && typeof (action.log) === 'string') {
                    monarchCommon.log(this.lexer, this.lexer.languageId + ': ' + monarchCommon.substituteMatches(this.lexer, action.log, matched, matches, state));
                }
            }
            // check result
            if (result === null) {
                monarchCommon.throwError(this.lexer, 'lexer rule has no well-defined action in rule: ' + rule.name);
                result = this.lexer.defaultToken;
            }
            // is the result a group match?
            if (Array.isArray(result)) {
                if (this.groupActions && this.groupActions.length > 0) {
                    monarchCommon.throwError(this.lexer, 'groups cannot be nested: ' + rule.name);
                }
                if (matches.length !== result.length + 1) {
                    monarchCommon.throwError(this.lexer, 'matched number of groups does not match the number of actions in rule: ' + rule.name);
                }
                var totalLen = 0;
                for (var i = 1; i < matches.length; i++) {
                    totalLen += matches[i].length;
                }
                if (totalLen !== matched.length) {
                    monarchCommon.throwError(this.lexer, 'with groups, all characters should be matched in consecutive groups in rule: ' + rule.name);
                }
                this.groupMatches = matches;
                this.groupMatched = matches.slice(1);
                this.groupActions = result.slice(0);
                this.groupRule = rule;
                stream.goBack(matched.length);
                return this.tokenize(stream); // call recursively to initiate first result match
            }
            else {
                // check for '@rematch'
                if (result === '@rematch') {
                    stream.goBack(matched.length);
                    matched = ''; // better set the next state too..
                    matches = null;
                    result = '';
                }
                // check progress
                if (matched.length === 0) {
                    if (stackLen0 !== this.stack.length || state !== this.stack[0]
                        || (!this.groupActions ? 0 : this.groupActions.length) !== groupLen0) {
                        if (!noConsumeIsOk) {
                            return this.tokenize(stream); // tokenize again in the new state
                        }
                    }
                    else {
                        monarchCommon.throwError(this.lexer, 'no progress in tokenizer in rule: ' + rule.name);
                        stream.advanceToEOS(); // must make progress or editor loops
                    }
                }
                // return the result (and check for brace matching)
                // todo: for efficiency we could pre-sanitize tokenPostfix and substitutions
                if (result.indexOf('@brackets') === 0) {
                    var rest = result.substr('@brackets'.length);
                    var bracket = findBracket(this.lexer, matched);
                    if (!bracket) {
                        monarchCommon.throwError(this.lexer, '@brackets token returned but no bracket defined as: ' + matched);
                        bracket = { token: '', bracketType: monarchCommon.MonarchBracket.None };
                    }
                    return { type: monarchCommon.sanitize(bracket.token + rest) };
                }
                else {
                    var token = (result === '' ? '' : result + this.lexer.tokenPostfix);
                    return { type: monarchCommon.sanitize(token) };
                }
            }
        };
        MonarchLexer.ID = 0;
        return MonarchLexer;
    }(abstractState_1.AbstractState));
    exports.MonarchLexer = MonarchLexer;
    /**
     * Searches for a bracket in the 'brackets' attribute that matches the input.
     */
    function findBracket(lexer, matched) {
        if (!matched) {
            return null;
        }
        matched = monarchCommon.fixCase(lexer, matched);
        var brackets = lexer.brackets;
        for (var i = 0; i < brackets.length; i++) {
            var bracket = brackets[i];
            if (bracket.open === matched) {
                return { token: bracket.token, bracketType: monarchCommon.MonarchBracket.Open };
            }
            else if (bracket.close === matched) {
                return { token: bracket.token, bracketType: monarchCommon.MonarchBracket.Close };
            }
        }
        return null;
    }
    function createTokenizationSupport(modeService, mode, lexer) {
        return new tokenizationSupport_1.TokenizationSupport(mode, {
            getInitialState: function () {
                return new MonarchLexer(mode, modeService, lexer);
            },
            enterNestedMode: function (state) {
                if (state instanceof MonarchLexer) {
                    return state.embeddedEntered;
                }
                return false;
            },
            getNestedMode: function (rawState) {
                var mime = rawState.embeddedMode;
                if (!modeService.isRegisteredMode(mime)) {
                    // unknown mode
                    return {
                        mode: modeService.getMode('text/plain'),
                        missingModePromise: null
                    };
                }
                var mode = modeService.getMode(mime);
                if (mode) {
                    // mode is available
                    return {
                        mode: mode,
                        missingModePromise: null
                    };
                }
                // mode is not yet loaded
                return {
                    mode: modeService.getMode('text/plain'),
                    missingModePromise: modeService.getOrCreateMode(mime).then(function () { return null; })
                };
            },
            getLeavingNestedModeData: function (line, state) {
                // state = state.clone();
                var mstate = state.clone();
                var stream = new lineStream_1.LineStream(line);
                while (!stream.eos() && mstate.embeddedMode) {
                    mstate.tokenize(stream, true); // allow no consumption for @rematch
                }
                if (mstate.embeddedMode) {
                    return null; // don't leave yet
                }
                var end = stream.pos();
                return {
                    nestedModeBuffer: line.substring(0, end),
                    bufferAfterNestedMode: line.substring(end),
                    stateAfterNestedMode: mstate
                };
            }
        }, lexer.usesEmbedded);
    }
    exports.createTokenizationSupport = createTokenizationSupport;
});

define(__m[108], __M([1,0]), function (require, exports) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var PrefixSumIndexOfResult = (function () {
        function PrefixSumIndexOfResult(index, remainder) {
            this.index = index;
            this.remainder = remainder;
        }
        return PrefixSumIndexOfResult;
    }());
    exports.PrefixSumIndexOfResult = PrefixSumIndexOfResult;
    var PrefixSumComputer = (function () {
        function PrefixSumComputer(values) {
            this.values = values;
            this.prefixSum = [];
            for (var i = 0, len = this.values.length; i < len; i++) {
                this.prefixSum[i] = 0;
            }
            this.prefixSumValidIndex = -1;
        }
        PrefixSumComputer.prototype.getCount = function () {
            return this.values.length;
        };
        PrefixSumComputer.prototype.insertValue = function (insertIndex, value) {
            insertIndex = Math.floor(insertIndex); //@perf
            value = Math.floor(value); //@perf
            this.values.splice(insertIndex, 0, value);
            this.prefixSum.splice(insertIndex, 0, 0);
            if (insertIndex - 1 < this.prefixSumValidIndex) {
                this.prefixSumValidIndex = insertIndex - 1;
            }
        };
        PrefixSumComputer.prototype.insertValues = function (insertIndex, values) {
            insertIndex = Math.floor(insertIndex); //@perf
            if (values.length === 0) {
                return;
            }
            this.values = this.values.slice(0, insertIndex).concat(values).concat(this.values.slice(insertIndex));
            this.prefixSum = this.prefixSum.slice(0, insertIndex).concat(PrefixSumComputer._zeroArray(values.length)).concat(this.prefixSum.slice(insertIndex));
            if (insertIndex - 1 < this.prefixSumValidIndex) {
                this.prefixSumValidIndex = insertIndex - 1;
            }
        };
        PrefixSumComputer._zeroArray = function (count) {
            count = Math.floor(count); //@perf
            var r = [];
            for (var i = 0; i < count; i++) {
                r[i] = 0;
            }
            return r;
        };
        PrefixSumComputer.prototype.changeValue = function (index, value) {
            index = Math.floor(index); //@perf
            value = Math.floor(value); //@perf
            if (this.values[index] === value) {
                return;
            }
            this.values[index] = value;
            if (index - 1 < this.prefixSumValidIndex) {
                this.prefixSumValidIndex = index - 1;
            }
        };
        PrefixSumComputer.prototype.removeValues = function (startIndex, cnt) {
            startIndex = Math.floor(startIndex); //@perf
            cnt = Math.floor(cnt); //@perf
            this.values.splice(startIndex, cnt);
            this.prefixSum.splice(startIndex, cnt);
            if (startIndex - 1 < this.prefixSumValidIndex) {
                this.prefixSumValidIndex = startIndex - 1;
            }
        };
        PrefixSumComputer.prototype.getTotalValue = function () {
            if (this.values.length === 0) {
                return 0;
            }
            return this.getAccumulatedValue(this.values.length - 1);
        };
        PrefixSumComputer.prototype.getAccumulatedValue = function (index) {
            index = Math.floor(index); //@perf
            if (index < 0) {
                return 0;
            }
            if (index <= this.prefixSumValidIndex) {
                return this.prefixSum[index];
            }
            var startIndex = this.prefixSumValidIndex + 1;
            if (startIndex === 0) {
                this.prefixSum[0] = this.values[0];
                startIndex++;
            }
            if (index >= this.values.length) {
                index = this.values.length - 1;
            }
            for (var i = startIndex; i <= index; i++) {
                this.prefixSum[i] = this.prefixSum[i - 1] + this.values[i];
            }
            this.prefixSumValidIndex = Math.max(this.prefixSumValidIndex, index);
            return this.prefixSum[index];
        };
        PrefixSumComputer.prototype.getIndexOf = function (accumulatedValue) {
            accumulatedValue = Math.floor(accumulatedValue); //@perf
            var low = 0;
            var high = this.values.length - 1;
            var mid;
            var midStop;
            var midStart;
            while (low <= high) {
                mid = low + ((high - low) / 2) | 0;
                midStop = this.getAccumulatedValue(mid);
                midStart = midStop - this.values[mid];
                if (accumulatedValue < midStart) {
                    high = mid - 1;
                }
                else if (accumulatedValue >= midStop) {
                    low = mid + 1;
                }
                else {
                    break;
                }
            }
            return new PrefixSumIndexOfResult(mid, accumulatedValue - midStart);
        };
        return PrefixSumComputer;
    }());
    exports.PrefixSumComputer = PrefixSumComputer;
});

define(__m[66], __M([7,6]), function(nls, data) { return nls.create("vs/base/common/errors", data); });
define(__m[2], __M([1,0,66,12,10,9,58,3]), function (require, exports, nls, objects, platform, types, arrays, strings) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    // Avoid circular dependency on EventEmitter by implementing a subset of the interface.
    var ErrorHandler = (function () {
        function ErrorHandler() {
            this.listeners = [];
            this.unexpectedErrorHandler = function (e) {
                platform.setTimeout(function () {
                    if (e.stack) {
                        throw new Error(e.message + '\n\n' + e.stack);
                    }
                    throw e;
                }, 0);
            };
        }
        ErrorHandler.prototype.addListener = function (listener) {
            var _this = this;
            this.listeners.push(listener);
            return function () {
                _this._removeListener(listener);
            };
        };
        ErrorHandler.prototype.emit = function (e) {
            this.listeners.forEach(function (listener) {
                listener(e);
            });
        };
        ErrorHandler.prototype._removeListener = function (listener) {
            this.listeners.splice(this.listeners.indexOf(listener), 1);
        };
        ErrorHandler.prototype.setUnexpectedErrorHandler = function (newUnexpectedErrorHandler) {
            this.unexpectedErrorHandler = newUnexpectedErrorHandler;
        };
        ErrorHandler.prototype.getUnexpectedErrorHandler = function () {
            return this.unexpectedErrorHandler;
        };
        ErrorHandler.prototype.onUnexpectedError = function (e) {
            this.unexpectedErrorHandler(e);
            this.emit(e);
        };
        return ErrorHandler;
    }());
    exports.ErrorHandler = ErrorHandler;
    exports.errorHandler = new ErrorHandler();
    function setUnexpectedErrorHandler(newUnexpectedErrorHandler) {
        exports.errorHandler.setUnexpectedErrorHandler(newUnexpectedErrorHandler);
    }
    exports.setUnexpectedErrorHandler = setUnexpectedErrorHandler;
    function onUnexpectedError(e) {
        // ignore errors from cancelled promises
        if (!isPromiseCanceledError(e)) {
            exports.errorHandler.onUnexpectedError(e);
        }
    }
    exports.onUnexpectedError = onUnexpectedError;
    function onUnexpectedPromiseError(promise) {
        return promise.then(null, onUnexpectedError);
    }
    exports.onUnexpectedPromiseError = onUnexpectedPromiseError;
    function transformErrorForSerialization(error) {
        if (error instanceof Error) {
            var name_1 = error.name, message = error.message;
            var stack = error.stacktrace || error.stack;
            return {
                $isError: true,
                name: name_1,
                message: message,
                stack: stack
            };
        }
        // return as is
        return error;
    }
    exports.transformErrorForSerialization = transformErrorForSerialization;
    /**
     * The base class for all connection errors originating from XHR requests.
     */
    var ConnectionError = (function () {
        function ConnectionError(arg) {
            this.status = arg.status;
            this.statusText = arg.statusText;
            this.name = 'ConnectionError';
            try {
                this.responseText = arg.responseText;
            }
            catch (e) {
                this.responseText = '';
            }
            this.errorMessage = null;
            this.errorCode = null;
            this.errorObject = null;
            if (this.responseText) {
                try {
                    var errorObj = JSON.parse(this.responseText);
                    this.errorMessage = errorObj.message;
                    this.errorCode = errorObj.code;
                    this.errorObject = errorObj;
                }
                catch (error) {
                }
            }
        }
        Object.defineProperty(ConnectionError.prototype, "message", {
            get: function () {
                return this.connectionErrorToMessage(this, false);
            },
            enumerable: true,
            configurable: true
        });
        Object.defineProperty(ConnectionError.prototype, "verboseMessage", {
            get: function () {
                return this.connectionErrorToMessage(this, true);
            },
            enumerable: true,
            configurable: true
        });
        ConnectionError.prototype.connectionErrorDetailsToMessage = function (error, verbose) {
            var errorCode = error.errorCode;
            var errorMessage = error.errorMessage;
            if (errorCode !== null && errorMessage !== null) {
                return nls.localize(0, null, strings.rtrim(errorMessage, '.'), errorCode);






            }
            if (errorMessage !== null) {
                return errorMessage;
            }
            if (verbose && error.responseText !== null) {
                return error.responseText;
            }
            return null;
        };
        ConnectionError.prototype.connectionErrorToMessage = function (error, verbose) {
            var details = this.connectionErrorDetailsToMessage(error, verbose);
            // Status Code based Error
            if (error.status === 401) {
                if (details !== null) {
                    return nls.localize(1, null, details);





                }
                return nls.localize(2, null);
            }
            // Return error details if present
            if (details) {
                return details;
            }
            // Fallback to HTTP Status and Code
            if (error.status > 0 && error.statusText !== null) {
                if (verbose && error.responseText !== null && error.responseText.length > 0) {
                    return nls.localize(3, null, error.statusText, error.status, error.responseText);
                }
                return nls.localize(4, null, error.statusText, error.status);
            }
            // Finally its an Unknown Connection Error
            if (verbose && error.responseText !== null && error.responseText.length > 0) {
                return nls.localize(5, null, error.responseText);
            }
            return nls.localize(6, null);
        };
        return ConnectionError;
    }());
    exports.ConnectionError = ConnectionError;
    // Bug: Can not subclass a JS Type. Do it manually (as done in WinJS.Class.derive)
    objects.derive(Error, ConnectionError);
    function xhrToErrorMessage(xhr, verbose) {
        var ce = new ConnectionError(xhr);
        if (verbose) {
            return ce.verboseMessage;
        }
        else {
            return ce.message;
        }
    }
    function exceptionToErrorMessage(exception, verbose) {
        if (exception.message) {
            if (verbose && (exception.stack || exception.stacktrace)) {
                return nls.localize(7, null, detectSystemErrorMessage(exception), exception.stack || exception.stacktrace);
            }
            return detectSystemErrorMessage(exception);
        }
        return nls.localize(8, null);
    }
    function detectSystemErrorMessage(exception) {
        // See https://nodejs.org/api/errors.html#errors_class_system_error
        if (typeof exception.code === 'string' && typeof exception.errno === 'number' && typeof exception.syscall === 'string') {
            return nls.localize(9, null, exception.message);
        }
        return exception.message;
    }
    /**
     * Tries to generate a human readable error message out of the error. If the verbose parameter
     * is set to true, the error message will include stacktrace details if provided.
     * @returns A string containing the error message.
     */
    function toErrorMessage(error, verbose) {
        if (error === void 0) { error = null; }
        if (verbose === void 0) { verbose = false; }
        if (!error) {
            return nls.localize(10, null);
        }
        if (Array.isArray(error)) {
            var errors = arrays.coalesce(error);
            var msg = toErrorMessage(errors[0], verbose);
            if (errors.length > 1) {
                return nls.localize(11, null, msg, errors.length);
            }
            return msg;
        }
        if (types.isString(error)) {
            return error;
        }
        if (!types.isUndefinedOrNull(error.status)) {
            return xhrToErrorMessage(error, verbose);
        }
        if (error.detail) {
            var detail = error.detail;
            if (detail.error) {
                if (detail.error && !types.isUndefinedOrNull(detail.error.status)) {
                    return xhrToErrorMessage(detail.error, verbose);
                }
                if (types.isArray(detail.error)) {
                    for (var i = 0; i < detail.error.length; i++) {
                        if (detail.error[i] && !types.isUndefinedOrNull(detail.error[i].status)) {
                            return xhrToErrorMessage(detail.error[i], verbose);
                        }
                    }
                }
                else {
                    return exceptionToErrorMessage(detail.error, verbose);
                }
            }
            if (detail.exception) {
                if (!types.isUndefinedOrNull(detail.exception.status)) {
                    return xhrToErrorMessage(detail.exception, verbose);
                }
                return exceptionToErrorMessage(detail.exception, verbose);
            }
        }
        if (error.stack) {
            return exceptionToErrorMessage(error, verbose);
        }
        if (error.message) {
            return error.message;
        }
        return nls.localize(12, null);
    }
    exports.toErrorMessage = toErrorMessage;
    var canceledName = 'Canceled';
    /**
     * Checks if the given error is a promise in canceled state
     */
    function isPromiseCanceledError(error) {
        return error instanceof Error && error.name === canceledName && error.message === canceledName;
    }
    exports.isPromiseCanceledError = isPromiseCanceledError;
    /**
     * Returns an error that signals cancellation.
     */
    function canceled() {
        var error = new Error(canceledName);
        error.name = error.message;
        return error;
    }
    exports.canceled = canceled;
    /**
     * Returns an error that signals something is not implemented.
     */
    function notImplemented() {
        return new Error(nls.localize(13, null));
    }
    exports.notImplemented = notImplemented;
    function illegalArgument(name) {
        if (name) {
            return new Error(nls.localize(14, null, name));
        }
        else {
            return new Error(nls.localize(15, null));
        }
    }
    exports.illegalArgument = illegalArgument;
    function illegalState(name) {
        if (name) {
            return new Error(nls.localize(16, null, name));
        }
        else {
            return new Error(nls.localize(17, null));
        }
    }
    exports.illegalState = illegalState;
    function readonly(name) {
        return name
            ? new Error("readonly property '" + name + " cannot be changed'")
            : new Error('readonly property cannot be changed');
    }
    exports.readonly = readonly;
    function loaderError(err) {
        if (platform.isWeb) {
            return new Error(nls.localize(18, null));
        }
        return new Error(nls.localize(19, null, JSON.stringify(err)));
    }
    exports.loaderError = loaderError;
    function create(message, options) {
        if (options === void 0) { options = {}; }
        var result = new Error(message);
        if (types.isNumber(options.severity)) {
            result.severity = options.severity;
        }
        if (options.actions) {
            result.actions = options.actions;
        }
        return result;
    }
    exports.create = create;
});

define(__m[63], __M([1,0,2]), function (require, exports, errors_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var CallbackList = (function () {
        function CallbackList() {
        }
        CallbackList.prototype.add = function (callback, context, bucket) {
            var _this = this;
            if (context === void 0) { context = null; }
            if (!this._callbacks) {
                this._callbacks = [];
                this._contexts = [];
            }
            this._callbacks.push(callback);
            this._contexts.push(context);
            if (Array.isArray(bucket)) {
                bucket.push({ dispose: function () { return _this.remove(callback, context); } });
            }
        };
        CallbackList.prototype.remove = function (callback, context) {
            if (context === void 0) { context = null; }
            if (!this._callbacks) {
                return;
            }
            var foundCallbackWithDifferentContext = false;
            for (var i = 0, len = this._callbacks.length; i < len; i++) {
                if (this._callbacks[i] === callback) {
                    if (this._contexts[i] === context) {
                        // callback & context match => remove it
                        this._callbacks.splice(i, 1);
                        this._contexts.splice(i, 1);
                        return;
                    }
                    else {
                        foundCallbackWithDifferentContext = true;
                    }
                }
            }
            if (foundCallbackWithDifferentContext) {
                throw new Error('When adding a listener with a context, you should remove it with the same context');
            }
        };
        CallbackList.prototype.invoke = function () {
            var args = [];
            for (var _i = 0; _i < arguments.length; _i++) {
                args[_i - 0] = arguments[_i];
            }
            if (!this._callbacks) {
                return;
            }
            var ret = [], callbacks = this._callbacks.slice(0), contexts = this._contexts.slice(0);
            for (var i = 0, len = callbacks.length; i < len; i++) {
                try {
                    ret.push(callbacks[i].apply(contexts[i], args));
                }
                catch (e) {
                    errors_1.onUnexpectedError(e);
                }
            }
            return ret;
        };
        CallbackList.prototype.isEmpty = function () {
            return !this._callbacks || this._callbacks.length === 0;
        };
        CallbackList.prototype.dispose = function () {
            this._callbacks = undefined;
            this._contexts = undefined;
        };
        return CallbackList;
    }());
    Object.defineProperty(exports, "__esModule", { value: true });
    exports.default = CallbackList;
});

define(__m[8], __M([1,0,63]), function (require, exports, callbackList_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var Event;
    (function (Event) {
        var _disposable = { dispose: function () { } };
        Event.None = function () { return _disposable; };
    })(Event || (Event = {}));
    Object.defineProperty(exports, "__esModule", { value: true });
    exports.default = Event;
    /**
     * The Emitter can be used to expose an Event to the public
     * to fire it from the insides.
     * Sample:
        class Document {
    
            private _onDidChange = new Emitter<(value:string)=>any>();
    
            public onDidChange = this._onDidChange.event;
    
            // getter-style
            // get onDidChange(): Event<(value:string)=>any> {
            // 	return this._onDidChange.event;
            // }
    
            private _doIt() {
                //...
                this._onDidChange.fire(value);
            }
        }
     */
    var Emitter = (function () {
        function Emitter(_options) {
            this._options = _options;
        }
        Object.defineProperty(Emitter.prototype, "event", {
            /**
             * For the public to allow to subscribe
             * to events from this Emitter
             */
            get: function () {
                var _this = this;
                if (!this._event) {
                    this._event = function (listener, thisArgs, disposables) {
                        if (!_this._callbacks) {
                            _this._callbacks = new callbackList_1.default();
                        }
                        if (_this._options && _this._options.onFirstListenerAdd && _this._callbacks.isEmpty()) {
                            _this._options.onFirstListenerAdd(_this);
                        }
                        _this._callbacks.add(listener, thisArgs);
                        var result;
                        result = {
                            dispose: function () {
                                result.dispose = Emitter._noop;
                                if (!_this._disposed) {
                                    _this._callbacks.remove(listener, thisArgs);
                                    if (_this._options && _this._options.onLastListenerRemove && _this._callbacks.isEmpty()) {
                                        _this._options.onLastListenerRemove(_this);
                                    }
                                }
                            }
                        };
                        if (Array.isArray(disposables)) {
                            disposables.push(result);
                        }
                        return result;
                    };
                }
                return this._event;
            },
            enumerable: true,
            configurable: true
        });
        /**
         * To be kept private to fire an event to
         * subscribers
         */
        Emitter.prototype.fire = function (event) {
            if (this._callbacks) {
                this._callbacks.invoke.call(this._callbacks, event);
            }
        };
        Emitter.prototype.dispose = function () {
            if (this._callbacks) {
                this._callbacks.dispose();
                this._callbacks = undefined;
                this._disposed = true;
            }
        };
        Emitter._noop = function () { };
        return Emitter;
    }());
    exports.Emitter = Emitter;
    /**
     * Creates an Event which is backed-up by the event emitter. This allows
     * to use the existing eventing pattern and is likely using less memory.
     * Sample:
     *
     * 	class Document {
     *
     *		private _eventbus = new EventEmitter();
     *
     *		public onDidChange = fromEventEmitter(this._eventbus, 'changed');
     *
     *		// getter-style
     *		// get onDidChange(): Event<(value:string)=>any> {
     *		// 	cache fromEventEmitter result and return
     *		// }
     *
     *		private _doIt() {
     *			// ...
     *			this._eventbus.emit('changed', value)
     *		}
     *	}
     */
    function fromEventEmitter(emitter, eventType) {
        return function (listener, thisArgs, disposables) {
            var result = emitter.addListener2(eventType, function () {
                listener.apply(thisArgs, arguments);
            });
            if (Array.isArray(disposables)) {
                disposables.push(result);
            }
            return result;
        };
    }
    exports.fromEventEmitter = fromEventEmitter;
    function mapEvent(event, map) {
        return function (listener, thisArgs, disposables) {
            if (thisArgs === void 0) { thisArgs = null; }
            return event(function (i) { return listener.call(thisArgs, map(i)); }, null, disposables);
        };
    }
    exports.mapEvent = mapEvent;
    function filterEvent(event, filter) {
        return function (listener, thisArgs, disposables) {
            if (thisArgs === void 0) { thisArgs = null; }
            return event(function (e) { return filter(e) && listener.call(thisArgs, e); }, null, disposables);
        };
    }
    exports.filterEvent = filterEvent;
    function debounceEvent(event, merger, delay) {
        if (delay === void 0) { delay = 100; }
        var subscription;
        var output;
        var handle;
        var emitter = new Emitter({
            onFirstListenerAdd: function () {
                subscription = event(function (cur) {
                    output = merger(output, cur);
                    clearTimeout(handle);
                    handle = setTimeout(function () {
                        emitter.fire(output);
                        output = undefined;
                    }, delay);
                });
            },
            onLastListenerRemove: function () {
                subscription.dispose();
            }
        });
        return emitter.event;
    }
    exports.debounceEvent = debounceEvent;
    var EventDelayerState;
    (function (EventDelayerState) {
        EventDelayerState[EventDelayerState["Idle"] = 0] = "Idle";
        EventDelayerState[EventDelayerState["Running"] = 1] = "Running";
    })(EventDelayerState || (EventDelayerState = {}));
    /**
     * The EventDelayer is useful in situations in which you want
     * to delay firing your events during some code.
     * You can wrap that code and be sure that the event will not
     * be fired during that wrap.
     *
     * ```
     * const emitter: Emitter;
     * const delayer = new EventDelayer();
     * const delayedEvent = delayer.delay(emitter.event);
     *
     * delayedEvent(console.log);
     *
     * delayer.wrap(() => {
     *   emitter.fire(); // event will not be fired yet
     * });
     *
     * // event will only be fired at this point
     * ```
     */
    var EventBufferer = (function () {
        function EventBufferer() {
            this.buffers = [];
        }
        EventBufferer.prototype.wrapEvent = function (event) {
            var _this = this;
            return function (listener, thisArgs, disposables) {
                return event(function (i) {
                    var buffer = _this.buffers[_this.buffers.length - 1];
                    if (buffer) {
                        buffer.push(function () { return listener.call(thisArgs, i); });
                    }
                    else {
                        listener.call(thisArgs, i);
                    }
                }, void 0, disposables);
            };
        };
        EventBufferer.prototype.bufferEvents = function (fn) {
            var buffer = [];
            this.buffers.push(buffer);
            fn();
            this.buffers.pop();
            buffer.forEach(function (flush) { return flush(); });
        };
        return EventBufferer;
    }());
    exports.EventBufferer = EventBufferer;
});

/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Microsoft Corporation. All rights reserved.
 *  Licensed under the MIT License. See License.txt in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
define(__m[47], __M([1,0,8]), function (require, exports, event_1) {
    'use strict';
    var CancellationToken;
    (function (CancellationToken) {
        CancellationToken.None = Object.freeze({
            isCancellationRequested: false,
            onCancellationRequested: event_1.default.None
        });
        CancellationToken.Cancelled = Object.freeze({
            isCancellationRequested: true,
            onCancellationRequested: event_1.default.None
        });
    })(CancellationToken = exports.CancellationToken || (exports.CancellationToken = {}));
    var shortcutEvent = Object.freeze(function (callback, context) {
        var handle = setTimeout(callback.bind(context), 0);
        return { dispose: function () { clearTimeout(handle); } };
    });
    var MutableToken = (function () {
        function MutableToken() {
            this._isCancelled = false;
        }
        MutableToken.prototype.cancel = function () {
            if (!this._isCancelled) {
                this._isCancelled = true;
                if (this._emitter) {
                    this._emitter.fire(undefined);
                    this._emitter = undefined;
                }
            }
        };
        Object.defineProperty(MutableToken.prototype, "isCancellationRequested", {
            get: function () {
                return this._isCancelled;
            },
            enumerable: true,
            configurable: true
        });
        Object.defineProperty(MutableToken.prototype, "onCancellationRequested", {
            get: function () {
                if (this._isCancelled) {
                    return shortcutEvent;
                }
                if (!this._emitter) {
                    this._emitter = new event_1.Emitter();
                }
                return this._emitter.event;
            },
            enumerable: true,
            configurable: true
        });
        return MutableToken;
    }());
    var CancellationTokenSource = (function () {
        function CancellationTokenSource() {
        }
        Object.defineProperty(CancellationTokenSource.prototype, "token", {
            get: function () {
                if (!this._token) {
                    // be lazy and create the token only when
                    // actually needed
                    this._token = new MutableToken();
                }
                return this._token;
            },
            enumerable: true,
            configurable: true
        });
        CancellationTokenSource.prototype.cancel = function () {
            if (!this._token) {
                // save an object by returning the default
                // cancelled token when cancellation happens
                // before someone asks for the token
                this._token = CancellationToken.Cancelled;
            }
            else {
                this._token.cancel();
            }
        };
        CancellationTokenSource.prototype.dispose = function () {
            this.cancel();
        };
        return CancellationTokenSource;
    }());
    exports.CancellationTokenSource = CancellationTokenSource;
});






define(__m[15], __M([1,0,2]), function (require, exports, Errors) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var EmitterEvent = (function () {
        function EmitterEvent(eventType, data) {
            if (eventType === void 0) { eventType = null; }
            if (data === void 0) { data = null; }
            this._type = eventType;
            this._data = data;
        }
        EmitterEvent.prototype.getType = function () {
            return this._type;
        };
        EmitterEvent.prototype.getData = function () {
            return this._data;
        };
        return EmitterEvent;
    }());
    exports.EmitterEvent = EmitterEvent;
    var EventEmitter = (function () {
        function EventEmitter(allowedEventTypes) {
            if (allowedEventTypes === void 0) { allowedEventTypes = null; }
            this._listeners = {};
            this._bulkListeners = [];
            this._collectedEvents = [];
            this._deferredCnt = 0;
            if (allowedEventTypes) {
                this._allowedEventTypes = {};
                for (var i = 0; i < allowedEventTypes.length; i++) {
                    this._allowedEventTypes[allowedEventTypes[i]] = true;
                }
            }
            else {
                this._allowedEventTypes = null;
            }
        }
        EventEmitter.prototype.dispose = function () {
            this._listeners = {};
            this._bulkListeners = [];
            this._collectedEvents = [];
            this._deferredCnt = 0;
            this._allowedEventTypes = null;
        };
        EventEmitter.prototype.addListener = function (eventType, listener) {
            if (eventType === '*') {
                throw new Error('Use addBulkListener(listener) to register your listener!');
            }
            if (this._allowedEventTypes && !this._allowedEventTypes.hasOwnProperty(eventType)) {
                throw new Error('This object will never emit this event type!');
            }
            if (this._listeners.hasOwnProperty(eventType)) {
                this._listeners[eventType].push(listener);
            }
            else {
                this._listeners[eventType] = [listener];
            }
            var bound = this;
            return {
                dispose: function () {
                    if (!bound) {
                        // Already called
                        return;
                    }
                    bound._removeListener(eventType, listener);
                    // Prevent leakers from holding on to the event emitter
                    bound = null;
                    listener = null;
                }
            };
        };
        EventEmitter.prototype.addListener2 = function (eventType, listener) {
            return this.addListener(eventType, listener);
        };
        EventEmitter.prototype.addOneTimeDisposableListener = function (eventType, listener) {
            var disposable = this.addListener(eventType, function (value) {
                disposable.dispose();
                listener(value);
            });
            return disposable;
        };
        EventEmitter.prototype.addBulkListener = function (listener) {
            var _this = this;
            this._bulkListeners.push(listener);
            return {
                dispose: function () {
                    _this._removeBulkListener(listener);
                }
            };
        };
        EventEmitter.prototype.addBulkListener2 = function (listener) {
            return this.addBulkListener(listener);
        };
        EventEmitter.prototype.addEmitter = function (eventEmitter) {
            var _this = this;
            return eventEmitter.addBulkListener2(function (events) {
                var newEvents = events;
                if (_this._deferredCnt === 0) {
                    _this._emitEvents(newEvents);
                }
                else {
                    // Collect for later
                    _this._collectedEvents.push.apply(_this._collectedEvents, newEvents);
                }
            });
        };
        EventEmitter.prototype.addEmitter2 = function (eventEmitter) {
            return this.addEmitter(eventEmitter);
        };
        EventEmitter.prototype._removeListener = function (eventType, listener) {
            if (this._listeners.hasOwnProperty(eventType)) {
                var listeners = this._listeners[eventType];
                for (var i = 0, len = listeners.length; i < len; i++) {
                    if (listeners[i] === listener) {
                        listeners.splice(i, 1);
                        break;
                    }
                }
            }
        };
        EventEmitter.prototype._removeBulkListener = function (listener) {
            for (var i = 0, len = this._bulkListeners.length; i < len; i++) {
                if (this._bulkListeners[i] === listener) {
                    this._bulkListeners.splice(i, 1);
                    break;
                }
            }
        };
        EventEmitter.prototype._emitToSpecificTypeListeners = function (eventType, data) {
            if (this._listeners.hasOwnProperty(eventType)) {
                var listeners = this._listeners[eventType].slice(0);
                for (var i = 0, len = listeners.length; i < len; i++) {
                    safeInvoke1Arg(listeners[i], data);
                }
            }
        };
        EventEmitter.prototype._emitToBulkListeners = function (events) {
            var bulkListeners = this._bulkListeners.slice(0);
            for (var i = 0, len = bulkListeners.length; i < len; i++) {
                safeInvoke1Arg(bulkListeners[i], events);
            }
        };
        EventEmitter.prototype._emitEvents = function (events) {
            if (this._bulkListeners.length > 0) {
                this._emitToBulkListeners(events);
            }
            for (var i = 0, len = events.length; i < len; i++) {
                var e = events[i];
                this._emitToSpecificTypeListeners(e.getType(), e.getData());
            }
        };
        EventEmitter.prototype.emit = function (eventType, data) {
            if (data === void 0) { data = {}; }
            if (this._allowedEventTypes && !this._allowedEventTypes.hasOwnProperty(eventType)) {
                throw new Error('Cannot emit this event type because it wasn\'t white-listed!');
            }
            // Early return if no listeners would get this
            if (!this._listeners.hasOwnProperty(eventType) && this._bulkListeners.length === 0) {
                return;
            }
            var emitterEvent = new EmitterEvent(eventType, data);
            if (this._deferredCnt === 0) {
                this._emitEvents([emitterEvent]);
            }
            else {
                // Collect for later
                this._collectedEvents.push(emitterEvent);
            }
        };
        EventEmitter.prototype.deferredEmit = function (callback) {
            this._deferredCnt = this._deferredCnt + 1;
            var result = safeInvokeNoArg(callback);
            this._deferredCnt = this._deferredCnt - 1;
            if (this._deferredCnt === 0) {
                this._emitCollected();
            }
            return result;
        };
        EventEmitter.prototype._emitCollected = function () {
            // Flush collected events
            var events = this._collectedEvents;
            this._collectedEvents = [];
            if (events.length > 0) {
                this._emitEvents(events);
            }
        };
        return EventEmitter;
    }());
    exports.EventEmitter = EventEmitter;
    var EmitQueueElement = (function () {
        function EmitQueueElement(target, arg) {
            this.target = target;
            this.arg = arg;
        }
        return EmitQueueElement;
    }());
    /**
     * Same as EventEmitter, but guarantees events are delivered in order to each listener
     */
    var OrderGuaranteeEventEmitter = (function (_super) {
        __extends(OrderGuaranteeEventEmitter, _super);
        function OrderGuaranteeEventEmitter(allowedEventTypes) {
            if (allowedEventTypes === void 0) { allowedEventTypes = null; }
            _super.call(this, allowedEventTypes);
            this._emitQueue = [];
        }
        OrderGuaranteeEventEmitter.prototype._emitToSpecificTypeListeners = function (eventType, data) {
            if (this._listeners.hasOwnProperty(eventType)) {
                var listeners = this._listeners[eventType];
                for (var i = 0, len = listeners.length; i < len; i++) {
                    this._emitQueue.push(new EmitQueueElement(listeners[i], data));
                }
            }
        };
        OrderGuaranteeEventEmitter.prototype._emitToBulkListeners = function (events) {
            var bulkListeners = this._bulkListeners;
            for (var i = 0, len = bulkListeners.length; i < len; i++) {
                this._emitQueue.push(new EmitQueueElement(bulkListeners[i], events));
            }
        };
        OrderGuaranteeEventEmitter.prototype._emitEvents = function (events) {
            _super.prototype._emitEvents.call(this, events);
            while (this._emitQueue.length > 0) {
                var queueElement = this._emitQueue.shift();
                safeInvoke1Arg(queueElement.target, queueElement.arg);
            }
        };
        return OrderGuaranteeEventEmitter;
    }(EventEmitter));
    exports.OrderGuaranteeEventEmitter = OrderGuaranteeEventEmitter;
    function safeInvokeNoArg(func) {
        try {
            return func();
        }
        catch (e) {
            Errors.onUnexpectedError(e);
        }
    }
    function safeInvoke1Arg(func, arg1) {
        try {
            return func(arg1);
        }
        catch (e) {
            Errors.onUnexpectedError(e);
        }
    }
});

define(__m[26], __M([1,0,10,2,61]), function (require, exports, Platform, errors, precision) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    exports.ENABLE_TIMER = false;
    var msWriteProfilerMark = Platform.globals['msWriteProfilerMark'];
    (function (Topic) {
        Topic[Topic["EDITOR"] = 0] = "EDITOR";
        Topic[Topic["LANGUAGES"] = 1] = "LANGUAGES";
        Topic[Topic["WORKER"] = 2] = "WORKER";
        Topic[Topic["WORKBENCH"] = 3] = "WORKBENCH";
        Topic[Topic["STARTUP"] = 4] = "STARTUP";
    })(exports.Topic || (exports.Topic = {}));
    var Topic = exports.Topic;
    var NullTimerEvent = (function () {
        function NullTimerEvent() {
        }
        NullTimerEvent.prototype.stop = function () {
            return;
        };
        NullTimerEvent.prototype.timeTaken = function () {
            return -1;
        };
        return NullTimerEvent;
    }());
    var TimerEvent = (function () {
        function TimerEvent(timeKeeper, name, topic, startTime, description) {
            this.timeKeeper = timeKeeper;
            this.name = name;
            this.description = description;
            this.topic = topic;
            this.stopTime = null;
            if (startTime) {
                this.startTime = startTime;
                return;
            }
            this.startTime = new Date();
            this.sw = precision.StopWatch.create();
            if (msWriteProfilerMark) {
                var profilerName = ['Monaco', this.topic, this.name, 'start'];
                msWriteProfilerMark(profilerName.join('|'));
            }
        }
        TimerEvent.prototype.stop = function (stopTime) {
            // already stopped
            if (this.stopTime !== null) {
                return;
            }
            if (stopTime) {
                this.stopTime = stopTime;
                this.sw = null;
                this.timeKeeper._onEventStopped(this);
                return;
            }
            this.stopTime = new Date();
            if (this.sw) {
                this.sw.stop();
            }
            this.timeKeeper._onEventStopped(this);
            if (msWriteProfilerMark) {
                var profilerName = ['Monaco', this.topic, this.name, 'stop'];
                msWriteProfilerMark(profilerName.join('|'));
            }
        };
        TimerEvent.prototype.timeTaken = function () {
            if (this.sw) {
                return this.sw.elapsed();
            }
            if (this.stopTime) {
                return this.stopTime.getTime() - this.startTime.getTime();
            }
            return -1;
        };
        return TimerEvent;
    }());
    var TimeKeeper = (function () {
        function TimeKeeper() {
            this.cleaningIntervalId = -1;
            this.collectedEvents = [];
            this.listeners = [];
        }
        TimeKeeper.prototype.isEnabled = function () {
            return exports.ENABLE_TIMER;
        };
        TimeKeeper.prototype.start = function (topic, name, start, description) {
            if (!this.isEnabled()) {
                return exports.nullEvent;
            }
            var strTopic;
            if (typeof topic === 'string') {
                strTopic = topic;
            }
            else if (topic === Topic.EDITOR) {
                strTopic = 'Editor';
            }
            else if (topic === Topic.LANGUAGES) {
                strTopic = 'Languages';
            }
            else if (topic === Topic.WORKER) {
                strTopic = 'Worker';
            }
            else if (topic === Topic.WORKBENCH) {
                strTopic = 'Workbench';
            }
            else if (topic === Topic.STARTUP) {
                strTopic = 'Startup';
            }
            this.initAutoCleaning();
            var event = new TimerEvent(this, name, strTopic, start, description);
            this.addEvent(event);
            return event;
        };
        TimeKeeper.prototype.dispose = function () {
            if (this.cleaningIntervalId !== -1) {
                Platform.clearInterval(this.cleaningIntervalId);
                this.cleaningIntervalId = -1;
            }
        };
        TimeKeeper.prototype.addListener = function (listener) {
            var _this = this;
            this.listeners.push(listener);
            return {
                dispose: function () {
                    for (var i = 0; i < _this.listeners.length; i++) {
                        if (_this.listeners[i] === listener) {
                            _this.listeners.splice(i, 1);
                            return;
                        }
                    }
                }
            };
        };
        TimeKeeper.prototype.addEvent = function (event) {
            event.id = TimeKeeper.EVENT_ID;
            TimeKeeper.EVENT_ID++;
            this.collectedEvents.push(event);
            // expire items from the front of the cache
            if (this.collectedEvents.length > TimeKeeper._EVENT_CACHE_LIMIT) {
                this.collectedEvents.shift();
            }
        };
        TimeKeeper.prototype.initAutoCleaning = function () {
            var _this = this;
            if (this.cleaningIntervalId === -1) {
                this.cleaningIntervalId = Platform.setInterval(function () {
                    var now = Date.now();
                    _this.collectedEvents.forEach(function (event) {
                        if (!event.stopTime && (now - event.startTime.getTime()) >= TimeKeeper._MAX_TIMER_LENGTH) {
                            event.stop();
                        }
                    });
                }, TimeKeeper._CLEAN_UP_INTERVAL);
            }
        };
        TimeKeeper.prototype.getCollectedEvents = function () {
            return this.collectedEvents.slice(0);
        };
        TimeKeeper.prototype.clearCollectedEvents = function () {
            this.collectedEvents = [];
        };
        TimeKeeper.prototype._onEventStopped = function (event) {
            var emitEvents = [event];
            var listeners = this.listeners.slice(0);
            for (var i = 0; i < listeners.length; i++) {
                try {
                    listeners[i](emitEvents);
                }
                catch (e) {
                    errors.onUnexpectedError(e);
                }
            }
        };
        TimeKeeper.prototype.setInitialCollectedEvents = function (events, startTime) {
            var _this = this;
            if (!this.isEnabled()) {
                return;
            }
            if (startTime) {
                TimeKeeper.PARSE_TIME = startTime;
            }
            events.forEach(function (event) {
                var e = new TimerEvent(_this, event.name, event.topic, event.startTime, event.description);
                e.stop(event.stopTime);
                _this.addEvent(e);
            });
        };
        /**
         * After being started for 1 minute, all timers are automatically stopped.
         */
        TimeKeeper._MAX_TIMER_LENGTH = 60000; // 1 minute
        /**
         * Every 2 minutes, a sweep of current started timers is done.
         */
        TimeKeeper._CLEAN_UP_INTERVAL = 120000; // 2 minutes
        /**
         * Collect at most 1000 events.
         */
        TimeKeeper._EVENT_CACHE_LIMIT = 1000;
        TimeKeeper.EVENT_ID = 1;
        TimeKeeper.PARSE_TIME = new Date();
        return TimeKeeper;
    }());
    exports.TimeKeeper = TimeKeeper;
    var timeKeeper = new TimeKeeper();
    exports.nullEvent = new NullTimerEvent();
    function start(topic, name, start, description) {
        return timeKeeper.start(topic, name, start, description);
    }
    exports.start = start;
    function getTimeKeeper() {
        return timeKeeper;
    }
    exports.getTimeKeeper = getTimeKeeper;
});

/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Microsoft Corporation. All rights reserved.
 *  Licensed under the MIT License. See License.txt in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

define(__m[5], __M([110,2]), function (winjs, __Errors__) {
	'use strict';

	var outstandingPromiseErrors = {};
	function promiseErrorHandler(e) {

		//
		// e.detail looks like: { exception, error, promise, handler, id, parent }
		//
		var details = e.detail;
		var id = details.id;

		// If the error has a parent promise then this is not the origination of the
		//  error so we check if it has a handler, and if so we mark that the error
		//  was handled by removing it from outstandingPromiseErrors
		//
		if (details.parent) {
			if (details.handler && outstandingPromiseErrors) {
				delete outstandingPromiseErrors[id];
			}
			return;
		}

		// Indicate that this error was originated and needs to be handled
		outstandingPromiseErrors[id] = details;

		// The first time the queue fills up this iteration, schedule a timeout to
		// check if any errors are still unhandled.
		if (Object.keys(outstandingPromiseErrors).length === 1) {
			setTimeout(function () {
				var errors = outstandingPromiseErrors;
				outstandingPromiseErrors = {};
				Object.keys(errors).forEach(function (errorId) {
					var error = errors[errorId];
					if(error.exception) {
						__Errors__.onUnexpectedError(error.exception);
					} else if(error.error) {
						__Errors__.onUnexpectedError(error.error);
					}
					console.log("WARNING: Promise with no error callback:" + error.id);
					console.log(error);
					if(error.exception) {
						console.log(error.exception.stack);
					}
				});
			}, 0);
		}
	}

	winjs.Promise.addEventListener("error", promiseErrorHandler);

	return {
		Promise: winjs.Promise,
		TPromise: winjs.Promise,
		PPromise: winjs.Promise
	};
});
/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Microsoft Corporation. All rights reserved.
 *  Licensed under the MIT License. See License.txt in the project root for license information.
 *--------------------------------------------------------------------------------------------*/





define(__m[24], __M([1,0,2,10,5,47,16]), function (require, exports, errors, platform, winjs_base_1, cancellation_1, lifecycle_1) {
    'use strict';
    function isThenable(obj) {
        return obj && typeof obj.then === 'function';
    }
    function toThenable(arg) {
        if (isThenable(arg)) {
            return arg;
        }
        else {
            return winjs_base_1.TPromise.as(arg);
        }
    }
    exports.toThenable = toThenable;
    function asWinJsPromise(callback) {
        var source = new cancellation_1.CancellationTokenSource();
        return new winjs_base_1.TPromise(function (resolve, reject) {
            var item = callback(source.token);
            if (isThenable(item)) {
                item.then(resolve, reject);
            }
            else {
                resolve(item);
            }
        }, function () {
            source.cancel();
        });
    }
    exports.asWinJsPromise = asWinJsPromise;
    /**
     * Hook a cancellation token to a WinJS Promise
     */
    function wireCancellationToken(token, promise) {
        token.onCancellationRequested(function () { return promise.cancel(); });
        return promise;
    }
    exports.wireCancellationToken = wireCancellationToken;
    /**
     * A helper to prevent accumulation of sequential async tasks.
     *
     * Imagine a mail man with the sole task of delivering letters. As soon as
     * a letter submitted for delivery, he drives to the destination, delivers it
     * and returns to his base. Imagine that during the trip, N more letters were submitted.
     * When the mail man returns, he picks those N letters and delivers them all in a
     * single trip. Even though N+1 submissions occurred, only 2 deliveries were made.
     *
     * The throttler implements this via the queue() method, by providing it a task
     * factory. Following the example:
     *
     * 		var throttler = new Throttler();
     * 		var letters = [];
     *
     * 		function deliver() {
     * 			const lettersToDeliver = letters;
     * 			letters = [];
     * 			return makeTheTrip(lettersToDeliver);
     * 		}
     *
     * 		function onLetterReceived(l) {
     * 			letters.push(l);
     * 			throttler.queue(deliver);
     * 		}
     */
    var Throttler = (function () {
        function Throttler() {
            this.activePromise = null;
            this.queuedPromise = null;
            this.queuedPromiseFactory = null;
        }
        Throttler.prototype.queue = function (promiseFactory) {
            var _this = this;
            if (this.activePromise) {
                this.queuedPromiseFactory = promiseFactory;
                if (!this.queuedPromise) {
                    var onComplete_1 = function () {
                        _this.queuedPromise = null;
                        var result = _this.queue(_this.queuedPromiseFactory);
                        _this.queuedPromiseFactory = null;
                        return result;
                    };
                    this.queuedPromise = new winjs_base_1.Promise(function (c, e, p) {
                        _this.activePromise.then(onComplete_1, onComplete_1, p).done(c);
                    }, function () {
                        _this.activePromise.cancel();
                    });
                }
                return new winjs_base_1.Promise(function (c, e, p) {
                    _this.queuedPromise.then(c, e, p);
                }, function () {
                    // no-op
                });
            }
            this.activePromise = promiseFactory();
            return new winjs_base_1.Promise(function (c, e, p) {
                _this.activePromise.done(function (result) {
                    _this.activePromise = null;
                    c(result);
                }, function (err) {
                    _this.activePromise = null;
                    e(err);
                }, p);
            }, function () {
                _this.activePromise.cancel();
            });
        };
        return Throttler;
    }());
    exports.Throttler = Throttler;
    // TODO@Joao: can the previous throttler be replaced with this?
    var SimpleThrottler = (function () {
        function SimpleThrottler() {
            this.current = winjs_base_1.TPromise.as(null);
        }
        SimpleThrottler.prototype.queue = function (promiseTask) {
            return this.current = this.current.then(function () { return promiseTask(); });
        };
        return SimpleThrottler;
    }());
    exports.SimpleThrottler = SimpleThrottler;
    /**
     * A helper to delay execution of a task that is being requested often.
     *
     * Following the throttler, now imagine the mail man wants to optimize the number of
     * trips proactively. The trip itself can be long, so the he decides not to make the trip
     * as soon as a letter is submitted. Instead he waits a while, in case more
     * letters are submitted. After said waiting period, if no letters were submitted, he
     * decides to make the trip. Imagine that N more letters were submitted after the first
     * one, all within a short period of time between each other. Even though N+1
     * submissions occurred, only 1 delivery was made.
     *
     * The delayer offers this behavior via the trigger() method, into which both the task
     * to be executed and the waiting period (delay) must be passed in as arguments. Following
     * the example:
     *
     * 		var delayer = new Delayer(WAITING_PERIOD);
     * 		var letters = [];
     *
     * 		function letterReceived(l) {
     * 			letters.push(l);
     * 			delayer.trigger(() => { return makeTheTrip(); });
     * 		}
     */
    var Delayer = (function () {
        function Delayer(defaultDelay) {
            this.defaultDelay = defaultDelay;
            this.timeout = null;
            this.completionPromise = null;
            this.onSuccess = null;
            this.task = null;
        }
        Delayer.prototype.trigger = function (task, delay) {
            var _this = this;
            if (delay === void 0) { delay = this.defaultDelay; }
            this.task = task;
            this.cancelTimeout();
            if (!this.completionPromise) {
                this.completionPromise = new winjs_base_1.Promise(function (c) {
                    _this.onSuccess = c;
                }, function () {
                    // no-op
                }).then(function () {
                    _this.completionPromise = null;
                    _this.onSuccess = null;
                    var task = _this.task;
                    _this.task = null;
                    return task();
                });
            }
            this.timeout = setTimeout(function () {
                _this.timeout = null;
                _this.onSuccess(null);
            }, delay);
            return this.completionPromise;
        };
        Delayer.prototype.isTriggered = function () {
            return this.timeout !== null;
        };
        Delayer.prototype.cancel = function () {
            this.cancelTimeout();
            if (this.completionPromise) {
                this.completionPromise.cancel();
                this.completionPromise = null;
            }
        };
        Delayer.prototype.cancelTimeout = function () {
            if (this.timeout !== null) {
                clearTimeout(this.timeout);
                this.timeout = null;
            }
        };
        return Delayer;
    }());
    exports.Delayer = Delayer;
    /**
     * A helper to delay execution of a task that is being requested often, while
     * preventing accumulation of consecutive executions, while the task runs.
     *
     * Simply combine the two mail man strategies from the Throttler and Delayer
     * helpers, for an analogy.
     */
    var ThrottledDelayer = (function (_super) {
        __extends(ThrottledDelayer, _super);
        function ThrottledDelayer(defaultDelay) {
            _super.call(this, defaultDelay);
            this.throttler = new Throttler();
        }
        ThrottledDelayer.prototype.trigger = function (promiseFactory, delay) {
            var _this = this;
            return _super.prototype.trigger.call(this, function () { return _this.throttler.queue(promiseFactory); }, delay);
        };
        return ThrottledDelayer;
    }(Delayer));
    exports.ThrottledDelayer = ThrottledDelayer;
    /**
     * Similar to the ThrottledDelayer, except it also guarantees that the promise
     * factory doesn't get called more often than every `minimumPeriod` milliseconds.
     */
    var PeriodThrottledDelayer = (function (_super) {
        __extends(PeriodThrottledDelayer, _super);
        function PeriodThrottledDelayer(defaultDelay, minimumPeriod) {
            if (minimumPeriod === void 0) { minimumPeriod = 0; }
            _super.call(this, defaultDelay);
            this.minimumPeriod = minimumPeriod;
            this.periodThrottler = new Throttler();
        }
        PeriodThrottledDelayer.prototype.trigger = function (promiseFactory, delay) {
            var _this = this;
            return _super.prototype.trigger.call(this, function () {
                return _this.periodThrottler.queue(function () {
                    return winjs_base_1.Promise.join([
                        winjs_base_1.TPromise.timeout(_this.minimumPeriod),
                        promiseFactory()
                    ]).then(function (r) { return r[1]; });
                });
            }, delay);
        };
        return PeriodThrottledDelayer;
    }(ThrottledDelayer));
    exports.PeriodThrottledDelayer = PeriodThrottledDelayer;
    var PromiseSource = (function () {
        function PromiseSource() {
            var _this = this;
            this._value = new winjs_base_1.TPromise(function (c, e) {
                _this._completeCallback = c;
                _this._errorCallback = e;
            });
        }
        Object.defineProperty(PromiseSource.prototype, "value", {
            get: function () {
                return this._value;
            },
            enumerable: true,
            configurable: true
        });
        PromiseSource.prototype.complete = function (value) {
            this._completeCallback(value);
        };
        PromiseSource.prototype.error = function (err) {
            this._errorCallback(err);
        };
        return PromiseSource;
    }());
    exports.PromiseSource = PromiseSource;
    var ShallowCancelThenPromise = (function (_super) {
        __extends(ShallowCancelThenPromise, _super);
        function ShallowCancelThenPromise(outer) {
            var completeCallback, errorCallback, progressCallback;
            _super.call(this, function (c, e, p) {
                completeCallback = c;
                errorCallback = e;
                progressCallback = p;
            }, function () {
                // cancel this promise but not the
                // outer promise
                errorCallback(errors.canceled());
            });
            outer.then(completeCallback, errorCallback, progressCallback);
        }
        return ShallowCancelThenPromise;
    }(winjs_base_1.TPromise));
    exports.ShallowCancelThenPromise = ShallowCancelThenPromise;
    /**
     * Returns a new promise that joins the provided promise. Upon completion of
     * the provided promise the provided function will always be called. This
     * method is comparable to a try-finally code block.
     * @param promise a promise
     * @param f a function that will be call in the success and error case.
     */
    function always(promise, f) {
        return new winjs_base_1.TPromise(function (c, e, p) {
            promise.done(function (result) {
                try {
                    f(result);
                }
                catch (e1) {
                    errors.onUnexpectedError(e1);
                }
                c(result);
            }, function (err) {
                try {
                    f(err);
                }
                catch (e1) {
                    errors.onUnexpectedError(e1);
                }
                e(err);
            }, function (progress) {
                p(progress);
            });
        }, function () {
            promise.cancel();
        });
    }
    exports.always = always;
    /**
     * Runs the provided list of promise factories in sequential order. The returned
     * promise will complete to an array of results from each promise.
     */
    function sequence(promiseFactory) {
        var results = [];
        // reverse since we start with last element using pop()
        promiseFactory = promiseFactory.reverse();
        function next() {
            if (promiseFactory.length) {
                return promiseFactory.pop()();
            }
            return null;
        }
        function thenHandler(result) {
            if (result) {
                results.push(result);
            }
            var n = next();
            if (n) {
                return n.then(thenHandler);
            }
            return winjs_base_1.TPromise.as(results);
        }
        return winjs_base_1.TPromise.as(null).then(thenHandler);
    }
    exports.sequence = sequence;
    function once(fn) {
        var _this = this;
        var didCall = false;
        var result;
        return function () {
            if (didCall) {
                return result;
            }
            didCall = true;
            result = fn.apply(_this, arguments);
            return result;
        };
    }
    exports.once = once;
    /**
     * A helper to queue N promises and run them all with a max degree of parallelism. The helper
     * ensures that at any time no more than M promises are running at the same time.
     */
    var Limiter = (function () {
        function Limiter(maxDegreeOfParalellism) {
            this.maxDegreeOfParalellism = maxDegreeOfParalellism;
            this.outstandingPromises = [];
            this.runningPromises = 0;
        }
        Limiter.prototype.queue = function (promiseFactory) {
            var _this = this;
            return new winjs_base_1.TPromise(function (c, e, p) {
                _this.outstandingPromises.push({
                    factory: promiseFactory,
                    c: c,
                    e: e,
                    p: p
                });
                _this.consume();
            });
        };
        Limiter.prototype.consume = function () {
            var _this = this;
            while (this.outstandingPromises.length && this.runningPromises < this.maxDegreeOfParalellism) {
                var iLimitedTask = this.outstandingPromises.shift();
                this.runningPromises++;
                var promise = iLimitedTask.factory();
                promise.done(iLimitedTask.c, iLimitedTask.e, iLimitedTask.p);
                promise.done(function () { return _this.consumed(); }, function () { return _this.consumed(); });
            }
        };
        Limiter.prototype.consumed = function () {
            this.runningPromises--;
            this.consume();
        };
        return Limiter;
    }());
    exports.Limiter = Limiter;
    var TimeoutTimer = (function (_super) {
        __extends(TimeoutTimer, _super);
        function TimeoutTimer() {
            _super.call(this);
            this._token = -1;
        }
        TimeoutTimer.prototype.dispose = function () {
            this.cancel();
            _super.prototype.dispose.call(this);
        };
        TimeoutTimer.prototype.cancel = function () {
            if (this._token !== -1) {
                platform.clearTimeout(this._token);
                this._token = -1;
            }
        };
        TimeoutTimer.prototype.cancelAndSet = function (runner, timeout) {
            var _this = this;
            this.cancel();
            this._token = platform.setTimeout(function () {
                _this._token = -1;
                runner();
            }, timeout);
        };
        TimeoutTimer.prototype.setIfNotSet = function (runner, timeout) {
            var _this = this;
            if (this._token !== -1) {
                // timer is already set
                return;
            }
            this._token = platform.setTimeout(function () {
                _this._token = -1;
                runner();
            }, timeout);
        };
        return TimeoutTimer;
    }(lifecycle_1.Disposable));
    exports.TimeoutTimer = TimeoutTimer;
    var IntervalTimer = (function (_super) {
        __extends(IntervalTimer, _super);
        function IntervalTimer() {
            _super.call(this);
            this._token = -1;
        }
        IntervalTimer.prototype.dispose = function () {
            this.cancel();
            _super.prototype.dispose.call(this);
        };
        IntervalTimer.prototype.cancel = function () {
            if (this._token !== -1) {
                platform.clearInterval(this._token);
                this._token = -1;
            }
        };
        IntervalTimer.prototype.cancelAndSet = function (runner, interval) {
            this.cancel();
            this._token = platform.setInterval(function () {
                runner();
            }, interval);
        };
        return IntervalTimer;
    }(lifecycle_1.Disposable));
    exports.IntervalTimer = IntervalTimer;
    var RunOnceScheduler = (function () {
        function RunOnceScheduler(runner, timeout) {
            this.timeoutToken = -1;
            this.runner = runner;
            this.timeout = timeout;
            this.timeoutHandler = this.onTimeout.bind(this);
        }
        /**
         * Dispose RunOnceScheduler
         */
        RunOnceScheduler.prototype.dispose = function () {
            this.cancel();
            this.runner = null;
        };
        /**
         * Cancel current scheduled runner (if any).
         */
        RunOnceScheduler.prototype.cancel = function () {
            if (this.isScheduled()) {
                platform.clearTimeout(this.timeoutToken);
                this.timeoutToken = -1;
            }
        };
        /**
         * Replace runner. If there is a runner already scheduled, the new runner will be called.
         */
        RunOnceScheduler.prototype.setRunner = function (runner) {
            this.runner = runner;
        };
        /**
         * Cancel previous runner (if any) & schedule a new runner.
         */
        RunOnceScheduler.prototype.schedule = function (delay) {
            if (delay === void 0) { delay = this.timeout; }
            this.cancel();
            this.timeoutToken = platform.setTimeout(this.timeoutHandler, delay);
        };
        /**
         * Returns true if scheduled.
         */
        RunOnceScheduler.prototype.isScheduled = function () {
            return this.timeoutToken !== -1;
        };
        RunOnceScheduler.prototype.onTimeout = function () {
            this.timeoutToken = -1;
            if (this.runner) {
                this.runner();
            }
        };
        return RunOnceScheduler;
    }());
    exports.RunOnceScheduler = RunOnceScheduler;
    function nfcall(fn) {
        var args = [];
        for (var _i = 1; _i < arguments.length; _i++) {
            args[_i - 1] = arguments[_i];
        }
        return new winjs_base_1.Promise(function (c, e) { return fn.apply(void 0, args.concat([function (err, result) { return err ? e(err) : c(result); }])); });
    }
    exports.nfcall = nfcall;
    function ninvoke(thisArg, fn) {
        var args = [];
        for (var _i = 2; _i < arguments.length; _i++) {
            args[_i - 2] = arguments[_i];
        }
        return new winjs_base_1.Promise(function (c, e) { return fn.call.apply(fn, [thisArg].concat(args, [function (err, result) { return err ? e(err) : c(result); }])); });
    }
    exports.ninvoke = ninvoke;
});

define(__m[84], __M([1,0,5]), function (require, exports, winjs_base_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var Schemas;
    (function (Schemas) {
        /**
         * A schema that is used for models that exist in memory
         * only and that have no correspondence on a server or such.
         */
        Schemas.inMemory = 'inmemory';
        /**
         * A schema that is used for setting files
         */
        Schemas.vscode = 'vscode';
        /**
         * A schema that is used for internal private files
         */
        Schemas.internal = 'private';
        Schemas.http = 'http';
        Schemas.https = 'https';
        Schemas.file = 'file';
    })(Schemas = exports.Schemas || (exports.Schemas = {}));
    function xhr(options) {
        var req = null;
        var canceled = false;
        return new winjs_base_1.TPromise(function (c, e, p) {
            req = new XMLHttpRequest();
            req.onreadystatechange = function () {
                if (canceled) {
                    return;
                }
                if (req.readyState === 4) {
                    // Handle 1223: http://bugs.jquery.com/ticket/1450
                    if ((req.status >= 200 && req.status < 300) || req.status === 1223) {
                        c(req);
                    }
                    else {
                        e(req);
                    }
                    req.onreadystatechange = function () { };
                }
                else {
                    p(req);
                }
            };
            req.open(options.type || 'GET', options.url, 
            // Promise based XHR does not support sync.
            //
            true, options.user, options.password);
            req.responseType = options.responseType || '';
            Object.keys(options.headers || {}).forEach(function (k) {
                req.setRequestHeader(k, options.headers[k]);
            });
            if (options.customRequestInitializer) {
                options.customRequestInitializer(req);
            }
            req.send(options.data);
        }, function () {
            canceled = true;
            req.abort();
        });
    }
    exports.xhr = xhr;
});

define(__m[109], __M([1,0,2,89,91]), function (require, exports, errors_1, marshalling_1, workerProtocol) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var WorkerServer = (function () {
        function WorkerServer(postSerializedMessage) {
            this._postSerializedMessage = postSerializedMessage;
            this._workerId = 0;
            this._requestHandler = null;
            this._bindConsole();
        }
        WorkerServer.prototype._bindConsole = function () {
            self.console = {
                log: this._sendPrintMessage.bind(this, workerProtocol.PrintType.LOG),
                debug: this._sendPrintMessage.bind(this, workerProtocol.PrintType.DEBUG),
                info: this._sendPrintMessage.bind(this, workerProtocol.PrintType.INFO),
                warn: this._sendPrintMessage.bind(this, workerProtocol.PrintType.WARN),
                error: this._sendPrintMessage.bind(this, workerProtocol.PrintType.ERROR)
            };
            errors_1.setUnexpectedErrorHandler(function (e) {
                self.console.error(e);
            });
        };
        WorkerServer.prototype._sendPrintMessage = function (level) {
            var objects = [];
            for (var _i = 1; _i < arguments.length; _i++) {
                objects[_i - 1] = arguments[_i];
            }
            var transformedObjects = objects.map(function (obj) { return (obj instanceof Error) ? errors_1.transformErrorForSerialization(obj) : obj; });
            var msg = {
                monacoWorker: true,
                from: this._workerId,
                req: '0',
                type: workerProtocol.MessageType.PRINT,
                level: level,
                payload: (transformedObjects.length === 1 ? transformedObjects[0] : transformedObjects)
            };
            this._postMessage(msg);
        };
        WorkerServer.prototype._sendReply = function (msgId, action, payload) {
            var msg = {
                monacoWorker: true,
                from: this._workerId,
                req: '0',
                id: msgId,
                type: workerProtocol.MessageType.REPLY,
                action: action,
                payload: (payload instanceof Error) ? errors_1.transformErrorForSerialization(payload) : payload
            };
            this._postMessage(msg);
        };
        WorkerServer.prototype.loadModule = function (moduleId, callback, errorback) {
            // Use the global require to be sure to get the global config
            self.require([moduleId], function () {
                var result = [];
                for (var _i = 0; _i < arguments.length; _i++) {
                    result[_i - 0] = arguments[_i];
                }
                callback(result[0]);
            }, errorback);
        };
        WorkerServer.prototype.onmessage = function (msg) {
            this._onmessage(marshalling_1.parse(msg));
        };
        WorkerServer.prototype._postMessage = function (msg) {
            this._postSerializedMessage(marshalling_1.stringify(msg));
        };
        WorkerServer.prototype._onmessage = function (msg) {
            var _this = this;
            var c = this._sendReply.bind(this, msg.id, workerProtocol.ReplyType.COMPLETE);
            var e = this._sendReply.bind(this, msg.id, workerProtocol.ReplyType.ERROR);
            var p = this._sendReply.bind(this, msg.id, workerProtocol.ReplyType.PROGRESS);
            switch (msg.type) {
                case workerProtocol.MessageType.INITIALIZE:
                    this._workerId = msg.payload.id;
                    var loaderConfig = msg.payload.loaderConfiguration;
                    // TODO@Alex: share this code with simpleWorker
                    if (loaderConfig) {
                        // Remove 'baseUrl', handling it is beyond scope for now
                        if (typeof loaderConfig.baseUrl !== 'undefined') {
                            delete loaderConfig['baseUrl'];
                        }
                        if (typeof loaderConfig.paths !== 'undefined') {
                            if (typeof loaderConfig.paths.vs !== 'undefined') {
                                delete loaderConfig.paths['vs'];
                            }
                        }
                        var nlsConfig_1 = loaderConfig['vs/nls'];
                        // We need to have pseudo translation
                        if (nlsConfig_1 && nlsConfig_1.pseudo) {
                            require(['vs/nls'], function (nlsPlugin) {
                                nlsPlugin.setPseudoTranslation(nlsConfig_1.pseudo);
                            });
                        }
                        // Since this is in a web worker, enable catching errors
                        loaderConfig.catchError = true;
                        self.require.config(loaderConfig);
                    }
                    this.loadModule(msg.payload.moduleId, function (handlerModule) {
                        _this._requestHandler = handlerModule.value;
                        c();
                    }, e);
                    break;
                default:
                    this._handleMessage(msg, c, e, p);
                    break;
            }
        };
        WorkerServer.prototype._handleMessage = function (msg, c, e, p) {
            if (!this._requestHandler) {
                e('Request handler not loaded');
                return;
            }
            var handlerMethod = this._requestHandler[msg.type];
            if (typeof handlerMethod !== 'function') {
                e('Handler does not have method ' + msg.type);
                return;
            }
            try {
                handlerMethod.call(this._requestHandler, this, c, e, p, msg.payload);
            }
            catch (handlerError) {
                e(errors_1.transformErrorForSerialization(handlerError));
            }
        };
        return WorkerServer;
    }());
    exports.WorkerServer = WorkerServer;
    function create(postMessage) {
        return new WorkerServer(postMessage);
    }
    exports.create = create;
});

define(__m[56], __M([1,0,2,3,50]), function (require, exports, errors_1, strings, viewLineToken_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var START_INDEX_MASK = 0xffffffff;
    var TYPE_MASK = 0xffff;
    var START_INDEX_OFFSET = 1;
    var TYPE_OFFSET = Math.pow(2, 32);
    var DEFAULT_VIEW_TOKEN = new viewLineToken_1.ViewLineToken(0, '');
    var INFLATED_TOKENS_EMPTY_TEXT = [];
    var DEFLATED_TOKENS_EMPTY_TEXT = [];
    var INFLATED_TOKENS_NON_EMPTY_TEXT = [DEFAULT_VIEW_TOKEN];
    var DEFLATED_TOKENS_NON_EMPTY_TEXT = [0];
    var TokensInflatorMap = (function () {
        function TokensInflatorMap() {
            this._inflate = [''];
            this._deflate = { '': 0 };
        }
        return TokensInflatorMap;
    }());
    exports.TokensInflatorMap = TokensInflatorMap;
    var TokensBinaryEncoding = (function () {
        function TokensBinaryEncoding() {
        }
        TokensBinaryEncoding.deflateArr = function (map, tokens) {
            if (tokens.length === 0) {
                return DEFLATED_TOKENS_EMPTY_TEXT;
            }
            if (tokens.length === 1 && tokens[0].startIndex === 0 && !tokens[0].type) {
                return DEFLATED_TOKENS_NON_EMPTY_TEXT;
            }
            var i, len, deflatedToken, deflated, token, inflateMap = map._inflate, deflateMap = map._deflate, prevStartIndex = -1, result = new Array(tokens.length);
            for (i = 0, len = tokens.length; i < len; i++) {
                token = tokens[i];
                if (token.startIndex <= prevStartIndex) {
                    token.startIndex = prevStartIndex + 1;
                    errors_1.onUnexpectedError({
                        message: 'Invalid tokens detected',
                        tokens: tokens
                    });
                }
                if (deflateMap.hasOwnProperty(token.type)) {
                    deflatedToken = deflateMap[token.type];
                }
                else {
                    deflatedToken = inflateMap.length;
                    deflateMap[token.type] = deflatedToken;
                    inflateMap.push(token.type);
                }
                // http://stackoverflow.com/a/2803010
                // All numbers in JavaScript are actually IEEE-754 compliant floating-point doubles.
                // These have a 53-bit mantissa which should mean that any integer value with a magnitude
                // of approximately 9 quadrillion or less -- more specifically, 9,007,199,254,740,991 --
                // will be represented accurately.
                // http://stackoverflow.com/a/6729252
                // Bitwise operations cast numbers to 32bit representation in JS
                // 32 bits for startIndex => up to 2^32 = 4,294,967,296
                // 16 bits for token => up to 2^16 = 65,536
                // [token][startIndex]
                deflated = deflatedToken * TYPE_OFFSET + token.startIndex * START_INDEX_OFFSET;
                result[i] = deflated;
                prevStartIndex = token.startIndex;
            }
            return result;
        };
        TokensBinaryEncoding.getStartIndex = function (binaryEncodedToken) {
            return (binaryEncodedToken / START_INDEX_OFFSET) & START_INDEX_MASK;
        };
        TokensBinaryEncoding.getType = function (map, binaryEncodedToken) {
            var deflatedType = (binaryEncodedToken / TYPE_OFFSET) & TYPE_MASK;
            if (deflatedType === 0) {
                return strings.empty;
            }
            return map._inflate[deflatedType];
        };
        TokensBinaryEncoding.inflateArr = function (map, binaryEncodedTokens) {
            if (binaryEncodedTokens.length === 0) {
                return INFLATED_TOKENS_EMPTY_TEXT;
            }
            if (binaryEncodedTokens.length === 1 && binaryEncodedTokens[0] === 0) {
                return INFLATED_TOKENS_NON_EMPTY_TEXT;
            }
            var result = [];
            var inflateMap = map._inflate;
            for (var i = 0, len = binaryEncodedTokens.length; i < len; i++) {
                var deflated = binaryEncodedTokens[i];
                var startIndex = (deflated / START_INDEX_OFFSET) & START_INDEX_MASK;
                var deflatedType = (deflated / TYPE_OFFSET) & TYPE_MASK;
                result.push(new viewLineToken_1.ViewLineToken(startIndex, inflateMap[deflatedType]));
            }
            return result;
        };
        TokensBinaryEncoding.findIndexOfOffset = function (binaryEncodedTokens, offset) {
            return this.findIndexInSegmentsArray(binaryEncodedTokens, offset);
        };
        TokensBinaryEncoding.sliceAndInflate = function (map, binaryEncodedTokens, startOffset, endOffset, deltaStartIndex) {
            if (binaryEncodedTokens.length === 0) {
                return INFLATED_TOKENS_EMPTY_TEXT;
            }
            if (binaryEncodedTokens.length === 1 && binaryEncodedTokens[0] === 0) {
                return INFLATED_TOKENS_NON_EMPTY_TEXT;
            }
            var startIndex = this.findIndexInSegmentsArray(binaryEncodedTokens, startOffset);
            var result = [];
            var inflateMap = map._inflate;
            var originalToken = binaryEncodedTokens[startIndex];
            var deflatedType = (originalToken / TYPE_OFFSET) & TYPE_MASK;
            var newStartIndex = 0;
            result.push(new viewLineToken_1.ViewLineToken(newStartIndex, inflateMap[deflatedType]));
            for (var i = startIndex + 1, len = binaryEncodedTokens.length; i < len; i++) {
                originalToken = binaryEncodedTokens[i];
                var originalStartIndex = (originalToken / START_INDEX_OFFSET) & START_INDEX_MASK;
                if (originalStartIndex >= endOffset) {
                    break;
                }
                deflatedType = (originalToken / TYPE_OFFSET) & TYPE_MASK;
                newStartIndex = originalStartIndex - startOffset + deltaStartIndex;
                result.push(new viewLineToken_1.ViewLineToken(newStartIndex, inflateMap[deflatedType]));
            }
            return result;
        };
        TokensBinaryEncoding.findIndexInSegmentsArray = function (arr, desiredIndex) {
            var low = 0, high = arr.length - 1, mid, value;
            while (low < high) {
                mid = low + Math.ceil((high - low) / 2);
                value = arr[mid] & 0xffffffff;
                if (value > desiredIndex) {
                    high = mid - 1;
                }
                else {
                    low = mid;
                }
            }
            return low;
        };
        TokensBinaryEncoding.START_INDEX_MASK = START_INDEX_MASK;
        TokensBinaryEncoding.TYPE_MASK = TYPE_MASK;
        TokensBinaryEncoding.START_INDEX_OFFSET = START_INDEX_OFFSET;
        TokensBinaryEncoding.TYPE_OFFSET = TYPE_OFFSET;
        return TokensBinaryEncoding;
    }());
    exports.TokensBinaryEncoding = TokensBinaryEncoding;
});

define(__m[57], __M([1,0,3,56,18,50]), function (require, exports, strings, tokensBinaryEncoding_1, modeTransition_1, viewLineToken_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var START_INDEX_MASK = tokensBinaryEncoding_1.TokensBinaryEncoding.START_INDEX_MASK;
    var TYPE_MASK = tokensBinaryEncoding_1.TokensBinaryEncoding.TYPE_MASK;
    var START_INDEX_OFFSET = tokensBinaryEncoding_1.TokensBinaryEncoding.START_INDEX_OFFSET;
    var TYPE_OFFSET = tokensBinaryEncoding_1.TokensBinaryEncoding.TYPE_OFFSET;
    var NO_OP_TOKENS_ADJUSTER = {
        adjust: function () { },
        finish: function () { }
    };
    var NO_OP_MARKERS_ADJUSTER = {
        adjustDelta: function () { },
        adjustSet: function () { },
        finish: function () { }
    };
    var MarkerMoveSemantics;
    (function (MarkerMoveSemantics) {
        MarkerMoveSemantics[MarkerMoveSemantics["MarkerDefined"] = 0] = "MarkerDefined";
        MarkerMoveSemantics[MarkerMoveSemantics["ForceMove"] = 1] = "ForceMove";
        MarkerMoveSemantics[MarkerMoveSemantics["ForceStay"] = 2] = "ForceStay";
    })(MarkerMoveSemantics || (MarkerMoveSemantics = {}));
    var ModelLine = (function () {
        function ModelLine(lineNumber, text) {
            this._lineNumber = lineNumber | 0;
            this._text = text;
            this._isInvalid = false;
            this._state = null;
            this._modeTransitions = null;
            this._lineTokens = null;
            this._markers = null;
        }
        Object.defineProperty(ModelLine.prototype, "lineNumber", {
            get: function () { return this._lineNumber; },
            enumerable: true,
            configurable: true
        });
        Object.defineProperty(ModelLine.prototype, "text", {
            get: function () { return this._text; },
            enumerable: true,
            configurable: true
        });
        Object.defineProperty(ModelLine.prototype, "isInvalid", {
            get: function () { return this._isInvalid; },
            set: function (value) { this._isInvalid = value; },
            enumerable: true,
            configurable: true
        });
        // --- BEGIN STATE
        ModelLine.prototype.setState = function (state) {
            this._state = state;
        };
        ModelLine.prototype.getState = function () {
            return this._state || null;
        };
        // --- END STATE
        // --- BEGIN MODE TRANSITIONS
        ModelLine.prototype.getModeTransitions = function (topLevelMode) {
            if (this._modeTransitions) {
                return this._modeTransitions;
            }
            else {
                return [new modeTransition_1.ModeTransition(0, topLevelMode)];
            }
        };
        // --- END MODE TRANSITIONS
        // --- BEGIN TOKENS
        ModelLine.prototype.setTokens = function (map, tokens, topLevelMode, modeTransitions) {
            this._lineTokens = toLineTokensFromInflated(map, tokens, this._text.length);
            this._modeTransitions = toModeTransitions(topLevelMode, modeTransitions);
        };
        ModelLine.prototype._setLineTokensFromDeflated = function (map, tokens) {
            this._lineTokens = toLineTokensFromDeflated(map, tokens, this._text.length);
        };
        ModelLine.prototype.getTokens = function () {
            if (this._lineTokens) {
                return this._lineTokens;
            }
            if (this._text.length === 0) {
                return EmptyLineTokens.INSTANCE;
            }
            return DefaultLineTokens.INSTANCE;
        };
        // --- END TOKENS
        ModelLine.prototype._createTokensAdjuster = function () {
            if (!this._lineTokens) {
                // This line does not have real tokens, so there is nothing to adjust
                return NO_OP_TOKENS_ADJUSTER;
            }
            var lineTokens = this._lineTokens;
            var tokens = lineTokens.getBinaryEncodedTokens();
            var tokensLength = tokens.length;
            var tokensIndex = 0;
            var currentTokenStartIndex = 0;
            var adjust = function (toColumn, delta, minimumAllowedColumn) {
                // console.log('before call: tokensIndex: ' + tokensIndex + ': ' + String(this.getTokens()));
                // console.log('adjustTokens: ' + toColumn + ' with delta: ' + delta + ' and [' + minimumAllowedColumn + ']');
                // console.log('currentTokenStartIndex: ' + currentTokenStartIndex);
                var minimumAllowedIndex = minimumAllowedColumn - 1;
                while (currentTokenStartIndex < toColumn && tokensIndex < tokensLength) {
                    if (currentTokenStartIndex > 0 && delta !== 0) {
                        // adjust token's `startIndex` by `delta`
                        var deflatedType = (tokens[tokensIndex] / TYPE_OFFSET) & TYPE_MASK;
                        var newStartIndex = Math.max(minimumAllowedIndex, currentTokenStartIndex + delta);
                        var newToken = deflatedType * TYPE_OFFSET + newStartIndex * START_INDEX_OFFSET;
                        if (delta < 0) {
                            // pop all previous tokens that have become `collapsed`
                            while (tokensIndex > 0) {
                                var prevTokenStartIndex = (tokens[tokensIndex - 1] / START_INDEX_OFFSET) & START_INDEX_MASK;
                                if (prevTokenStartIndex >= newStartIndex) {
                                    // Token at `tokensIndex` - 1 is now `collapsed` => pop it
                                    tokens.splice(tokensIndex - 1, 1);
                                    tokensLength--;
                                    tokensIndex--;
                                }
                                else {
                                    break;
                                }
                            }
                        }
                        tokens[tokensIndex] = newToken;
                    }
                    tokensIndex++;
                    if (tokensIndex < tokensLength) {
                        currentTokenStartIndex = (tokens[tokensIndex] / START_INDEX_OFFSET) & START_INDEX_MASK;
                    }
                }
                // console.log('after call: tokensIndex: ' + tokensIndex + ': ' + String(this.getTokens()));
            };
            var finish = function (delta, lineTextLength) {
                adjust(Number.MAX_VALUE, delta, 1);
            };
            return {
                adjust: adjust,
                finish: finish
            };
        };
        ModelLine.prototype._setText = function (text) {
            this._text = text;
            if (this._lineTokens) {
                var map = this._lineTokens.getBinaryEncodedTokensMap(), tokens = this._lineTokens.getBinaryEncodedTokens(), lineTextLength = this._text.length;
                // Remove overflowing tokens
                while (tokens.length > 0) {
                    var lastTokenStartIndex = (tokens[tokens.length - 1] / START_INDEX_OFFSET) & START_INDEX_MASK;
                    if (lastTokenStartIndex < lineTextLength) {
                        // Valid token
                        break;
                    }
                    // This token now overflows the text => remove it
                    tokens.pop();
                }
                this._setLineTokensFromDeflated(map, tokens);
            }
        };
        // private _printMarkers(): string {
        // 	if (!this._markers) {
        // 		return '[]';
        // 	}
        // 	if (this._markers.length === 0) {
        // 		return '[]';
        // 	}
        // 	var markers = this._markers;
        // 	var printMarker = (m:ILineMarker) => {
        // 		if (m.stickToPreviousCharacter) {
        // 			return '|' + m.column;
        // 		}
        // 		return m.column + '|';
        // 	};
        // 	return '[' + markers.map(printMarker).join(', ') + ']';
        // }
        ModelLine.prototype._createMarkersAdjuster = function (changedMarkers) {
            var _this = this;
            if (!this._markers) {
                return NO_OP_MARKERS_ADJUSTER;
            }
            if (this._markers.length === 0) {
                return NO_OP_MARKERS_ADJUSTER;
            }
            this._markers.sort(ModelLine._compareMarkers);
            var markers = this._markers;
            var markersLength = markers.length;
            var markersIndex = 0;
            var marker = markers[markersIndex];
            // console.log('------------- INITIAL MARKERS: ' + this._printMarkers());
            var adjustMarkerBeforeColumn = function (toColumn, moveSemantics) {
                if (marker.column < toColumn) {
                    return true;
                }
                if (marker.column > toColumn) {
                    return false;
                }
                if (moveSemantics === MarkerMoveSemantics.ForceMove) {
                    return false;
                }
                if (moveSemantics === MarkerMoveSemantics.ForceStay) {
                    return true;
                }
                return marker.stickToPreviousCharacter;
            };
            var adjustDelta = function (toColumn, delta, minimumAllowedColumn, moveSemantics) {
                // console.log('------------------------------');
                // console.log('adjustDelta called: toColumn: ' + toColumn + ', delta: ' + delta + ', minimumAllowedColumn: ' + minimumAllowedColumn + ', moveSemantics: ' + MarkerMoveSemantics[moveSemantics]);
                // console.log('BEFORE::: markersIndex: ' + markersIndex + ' : ' + this._printMarkers());
                while (markersIndex < markersLength && adjustMarkerBeforeColumn(toColumn, moveSemantics)) {
                    if (delta !== 0) {
                        var newColumn = Math.max(minimumAllowedColumn, marker.column + delta);
                        if (marker.column !== newColumn) {
                            changedMarkers[marker.id] = true;
                            marker.oldLineNumber = marker.oldLineNumber || _this._lineNumber;
                            marker.oldColumn = marker.oldColumn || marker.column;
                            marker.column = newColumn;
                        }
                    }
                    markersIndex++;
                    if (markersIndex < markersLength) {
                        marker = markers[markersIndex];
                    }
                }
                // console.log('AFTER::: markersIndex: ' + markersIndex + ' : ' + this._printMarkers());
            };
            var adjustSet = function (toColumn, newColumn, moveSemantics) {
                // console.log('------------------------------');
                // console.log('adjustSet called: toColumn: ' + toColumn + ', newColumn: ' + newColumn + ', moveSemantics: ' + MarkerMoveSemantics[moveSemantics]);
                // console.log('BEFORE::: markersIndex: ' + markersIndex + ' : ' + this._printMarkers());
                while (markersIndex < markersLength && adjustMarkerBeforeColumn(toColumn, moveSemantics)) {
                    if (marker.column !== newColumn) {
                        changedMarkers[marker.id] = true;
                        marker.oldLineNumber = marker.oldLineNumber || _this._lineNumber;
                        marker.oldColumn = marker.oldColumn || marker.column;
                        marker.column = newColumn;
                    }
                    markersIndex++;
                    if (markersIndex < markersLength) {
                        marker = markers[markersIndex];
                    }
                }
                // console.log('AFTER::: markersIndex: ' + markersIndex + ' : ' + this._printMarkers());
            };
            var finish = function (delta, lineTextLength) {
                adjustDelta(Number.MAX_VALUE, delta, 1, MarkerMoveSemantics.MarkerDefined);
                // console.log('------------- FINAL MARKERS: ' + this._printMarkers());
            };
            return {
                adjustDelta: adjustDelta,
                adjustSet: adjustSet,
                finish: finish
            };
        };
        ModelLine.prototype.applyEdits = function (changedMarkers, edits) {
            var deltaColumn = 0;
            var resultText = this._text;
            var tokensAdjuster = this._createTokensAdjuster();
            var markersAdjuster = this._createMarkersAdjuster(changedMarkers);
            for (var i = 0, len = edits.length; i < len; i++) {
                var edit = edits[i];
                // console.log();
                // console.log('=============================');
                // console.log('EDIT #' + i + ' [ ' + edit.startColumn + ' -> ' + edit.endColumn + ' ] : <<<' + edit.text + '>>>, forceMoveMarkers: ' + edit.forceMoveMarkers);
                // console.log('deltaColumn: ' + deltaColumn);
                var startColumn = deltaColumn + edit.startColumn;
                var endColumn = deltaColumn + edit.endColumn;
                var deletingCnt = endColumn - startColumn;
                var insertingCnt = edit.text.length;
                // Adjust tokens & markers before this edit
                // console.log('Adjust tokens & markers before this edit');
                tokensAdjuster.adjust(edit.startColumn - 1, deltaColumn, 1);
                markersAdjuster.adjustDelta(edit.startColumn, deltaColumn, 1, edit.forceMoveMarkers ? MarkerMoveSemantics.ForceMove : (deletingCnt > 0 ? MarkerMoveSemantics.ForceStay : MarkerMoveSemantics.MarkerDefined));
                // Adjust tokens & markers for the common part of this edit
                var commonLength = Math.min(deletingCnt, insertingCnt);
                if (commonLength > 0) {
                    // console.log('Adjust tokens & markers for the common part of this edit');
                    tokensAdjuster.adjust(edit.startColumn - 1 + commonLength, deltaColumn, startColumn);
                    if (!edit.forceMoveMarkers) {
                        markersAdjuster.adjustDelta(edit.startColumn + commonLength, deltaColumn, startColumn, edit.forceMoveMarkers ? MarkerMoveSemantics.ForceMove : (deletingCnt > insertingCnt ? MarkerMoveSemantics.ForceStay : MarkerMoveSemantics.MarkerDefined));
                    }
                }
                // Perform the edit & update `deltaColumn`
                resultText = resultText.substring(0, startColumn - 1) + edit.text + resultText.substring(endColumn - 1);
                deltaColumn += insertingCnt - deletingCnt;
                // Adjust tokens & markers inside this edit
                // console.log('Adjust tokens & markers inside this edit');
                tokensAdjuster.adjust(edit.endColumn, deltaColumn, startColumn);
                markersAdjuster.adjustSet(edit.endColumn, startColumn + insertingCnt, edit.forceMoveMarkers ? MarkerMoveSemantics.ForceMove : MarkerMoveSemantics.MarkerDefined);
            }
            // Wrap up tokens & markers; adjust remaining if needed
            tokensAdjuster.finish(deltaColumn, resultText.length);
            markersAdjuster.finish(deltaColumn, resultText.length);
            // Save the resulting text
            this._setText(resultText);
            return deltaColumn;
        };
        ModelLine.prototype.split = function (changedMarkers, splitColumn, forceMoveMarkers) {
            // console.log('--> split @ ' + splitColumn + '::: ' + this._printMarkers());
            var myText = this._text.substring(0, splitColumn - 1);
            var otherText = this._text.substring(splitColumn - 1);
            var otherMarkers = null;
            if (this._markers) {
                this._markers.sort(ModelLine._compareMarkers);
                for (var i = 0, len = this._markers.length; i < len; i++) {
                    var marker = this._markers[i];
                    if (marker.column > splitColumn
                        || (marker.column === splitColumn
                            && (forceMoveMarkers
                                || !marker.stickToPreviousCharacter))) {
                        var myMarkers = this._markers.slice(0, i);
                        otherMarkers = this._markers.slice(i);
                        this._markers = myMarkers;
                        break;
                    }
                }
                if (otherMarkers) {
                    for (var i = 0, len = otherMarkers.length; i < len; i++) {
                        var marker = otherMarkers[i];
                        changedMarkers[marker.id] = true;
                        marker.oldLineNumber = marker.oldLineNumber || this._lineNumber;
                        marker.oldColumn = marker.oldColumn || marker.column;
                        marker.column -= splitColumn - 1;
                    }
                }
            }
            this._setText(myText);
            var otherLine = new ModelLine(this._lineNumber + 1, otherText);
            if (otherMarkers) {
                otherLine.addMarkers(otherMarkers);
            }
            return otherLine;
        };
        ModelLine.prototype.append = function (changedMarkers, other) {
            // console.log('--> append: THIS :: ' + this._printMarkers());
            // console.log('--> append: OTHER :: ' + this._printMarkers());
            var thisTextLength = this._text.length;
            this._setText(this._text + other._text);
            var otherLineTokens = other._lineTokens;
            if (otherLineTokens) {
                // Other has real tokens
                var otherTokens = otherLineTokens.getBinaryEncodedTokens();
                // Adjust other tokens
                if (thisTextLength > 0) {
                    for (var i = 0, len = otherTokens.length; i < len; i++) {
                        var token = otherTokens[i];
                        var deflatedStartIndex = (token / START_INDEX_OFFSET) & START_INDEX_MASK;
                        var deflatedType = (token / TYPE_OFFSET) & TYPE_MASK;
                        var newStartIndex = deflatedStartIndex + thisTextLength;
                        var newToken = deflatedType * TYPE_OFFSET + newStartIndex * START_INDEX_OFFSET;
                        otherTokens[i] = newToken;
                    }
                }
                // Append other tokens
                var myLineTokens = this._lineTokens;
                if (myLineTokens) {
                    // I have real tokens
                    this._setLineTokensFromDeflated(myLineTokens.getBinaryEncodedTokensMap(), myLineTokens.getBinaryEncodedTokens().concat(otherTokens));
                }
                else {
                    // I don't have real tokens
                    this._setLineTokensFromDeflated(otherLineTokens.getBinaryEncodedTokensMap(), otherTokens);
                }
            }
            if (other._markers) {
                // Other has markers
                var otherMarkers = other._markers;
                // Adjust other markers
                for (var i = 0, len = otherMarkers.length; i < len; i++) {
                    var marker = otherMarkers[i];
                    changedMarkers[marker.id] = true;
                    marker.oldLineNumber = marker.oldLineNumber || other.lineNumber;
                    marker.oldColumn = marker.oldColumn || marker.column;
                    marker.column += thisTextLength;
                }
                this.addMarkers(otherMarkers);
            }
        };
        ModelLine.prototype.addMarker = function (marker) {
            marker.line = this;
            if (!this._markers) {
                this._markers = [marker];
            }
            else {
                this._markers.push(marker);
            }
        };
        ModelLine.prototype.addMarkers = function (markers) {
            if (markers.length === 0) {
                return;
            }
            var i, len;
            for (i = 0, len = markers.length; i < len; i++) {
                markers[i].line = this;
            }
            if (!this._markers) {
                this._markers = markers.slice(0);
            }
            else {
                this._markers = this._markers.concat(markers);
            }
        };
        ModelLine._compareMarkers = function (a, b) {
            if (a.column === b.column) {
                return (a.stickToPreviousCharacter ? 0 : 1) - (b.stickToPreviousCharacter ? 0 : 1);
            }
            return a.column - b.column;
        };
        ModelLine.prototype.removeMarker = function (marker) {
            if (!this._markers) {
                return;
            }
            var index = this._indexOfMarkerId(marker.id);
            if (index >= 0) {
                marker.line = null;
                this._markers.splice(index, 1);
            }
            if (this._markers.length === 0) {
                this._markers = null;
            }
        };
        ModelLine.prototype.removeMarkers = function (deleteMarkers) {
            if (!this._markers) {
                return;
            }
            for (var i = 0, len = this._markers.length; i < len; i++) {
                var marker = this._markers[i];
                if (deleteMarkers[marker.id]) {
                    marker.line = null;
                    this._markers.splice(i, 1);
                    len--;
                    i--;
                }
            }
            if (this._markers.length === 0) {
                this._markers = null;
            }
        };
        ModelLine.prototype.getMarkers = function () {
            if (!this._markers) {
                return [];
            }
            return this._markers.slice(0);
        };
        ModelLine.prototype.updateLineNumber = function (changedMarkers, newLineNumber) {
            if (this._markers) {
                var markers = this._markers, i, len, marker;
                for (i = 0, len = markers.length; i < len; i++) {
                    marker = markers[i];
                    changedMarkers[marker.id] = true;
                    marker.oldLineNumber = marker.oldLineNumber || this._lineNumber;
                }
            }
            this._lineNumber = newLineNumber;
        };
        ModelLine.prototype.deleteLine = function (changedMarkers, setMarkersColumn, setMarkersOldLineNumber) {
            // console.log('--> deleteLine: ');
            if (this._markers) {
                var markers = this._markers, i, len, marker;
                // Mark all these markers as changed
                for (i = 0, len = markers.length; i < len; i++) {
                    marker = markers[i];
                    changedMarkers[marker.id] = true;
                    marker.oldColumn = marker.oldColumn || marker.column;
                    marker.oldLineNumber = marker.oldLineNumber || setMarkersOldLineNumber;
                    marker.column = setMarkersColumn;
                }
                return markers;
            }
            return [];
        };
        ModelLine.prototype._indexOfMarkerId = function (markerId) {
            var markers = this._markers;
            for (var i = 0, len = markers.length; i < len; i++) {
                if (markers[i].id === markerId) {
                    return i;
                }
            }
        };
        return ModelLine;
    }());
    exports.ModelLine = ModelLine;
    function toLineTokensFromInflated(map, tokens, textLength) {
        if (textLength === 0) {
            return null;
        }
        if (!tokens || tokens.length === 0) {
            return null;
        }
        if (tokens.length === 1) {
            if (tokens[0].startIndex === 0 && tokens[0].type === '') {
                return null;
            }
        }
        var deflated = tokensBinaryEncoding_1.TokensBinaryEncoding.deflateArr(map, tokens);
        return new LineTokens(map, deflated);
    }
    function toLineTokensFromDeflated(map, tokens, textLength) {
        if (textLength === 0) {
            return null;
        }
        if (!tokens || tokens.length === 0) {
            return null;
        }
        if (tokens.length === 1) {
            if (tokens[0] === 0) {
                return null;
            }
        }
        return new LineTokens(map, tokens);
    }
    var LineTokens = (function () {
        function LineTokens(map, tokens) {
            this.map = map;
            this._tokens = tokens;
        }
        LineTokens.prototype.getBinaryEncodedTokensMap = function () {
            return this.map;
        };
        LineTokens.prototype.getBinaryEncodedTokens = function () {
            return this._tokens;
        };
        LineTokens.prototype.getTokenCount = function () {
            return this._tokens.length;
        };
        LineTokens.prototype.getTokenStartIndex = function (tokenIndex) {
            return tokensBinaryEncoding_1.TokensBinaryEncoding.getStartIndex(this._tokens[tokenIndex]);
        };
        LineTokens.prototype.getTokenType = function (tokenIndex) {
            return tokensBinaryEncoding_1.TokensBinaryEncoding.getType(this.map, this._tokens[tokenIndex]);
        };
        LineTokens.prototype.getTokenEndIndex = function (tokenIndex, textLength) {
            if (tokenIndex + 1 < this._tokens.length) {
                return tokensBinaryEncoding_1.TokensBinaryEncoding.getStartIndex(this._tokens[tokenIndex + 1]);
            }
            return textLength;
        };
        LineTokens.prototype.equals = function (other) {
            if (other instanceof LineTokens) {
                if (this.map !== other.map) {
                    return false;
                }
                if (this._tokens.length !== other._tokens.length) {
                    return false;
                }
                for (var i = 0, len = this._tokens.length; i < len; i++) {
                    if (this._tokens[i] !== other._tokens[i]) {
                        return false;
                    }
                }
                return true;
            }
            if (!(other instanceof LineTokens)) {
                return false;
            }
        };
        LineTokens.prototype.findIndexOfOffset = function (offset) {
            return tokensBinaryEncoding_1.TokensBinaryEncoding.findIndexOfOffset(this._tokens, offset);
        };
        LineTokens.prototype.inflate = function () {
            return tokensBinaryEncoding_1.TokensBinaryEncoding.inflateArr(this.map, this._tokens);
        };
        LineTokens.prototype.sliceAndInflate = function (startOffset, endOffset, deltaStartIndex) {
            return tokensBinaryEncoding_1.TokensBinaryEncoding.sliceAndInflate(this.map, this._tokens, startOffset, endOffset, deltaStartIndex);
        };
        return LineTokens;
    }());
    exports.LineTokens = LineTokens;
    var EmptyLineTokens = (function () {
        function EmptyLineTokens() {
        }
        EmptyLineTokens.prototype.getTokenCount = function () {
            return 0;
        };
        EmptyLineTokens.prototype.getTokenStartIndex = function (tokenIndex) {
            return 0;
        };
        EmptyLineTokens.prototype.getTokenType = function (tokenIndex) {
            return strings.empty;
        };
        EmptyLineTokens.prototype.getTokenEndIndex = function (tokenIndex, textLength) {
            return 0;
        };
        EmptyLineTokens.prototype.equals = function (other) {
            return other === this;
        };
        EmptyLineTokens.prototype.findIndexOfOffset = function (offset) {
            return 0;
        };
        EmptyLineTokens.prototype.inflate = function () {
            return [];
        };
        EmptyLineTokens.prototype.sliceAndInflate = function (startOffset, endOffset, deltaStartIndex) {
            return [];
        };
        EmptyLineTokens.INSTANCE = new EmptyLineTokens();
        return EmptyLineTokens;
    }());
    var DefaultLineTokens = (function () {
        function DefaultLineTokens() {
        }
        DefaultLineTokens.prototype.getTokenCount = function () {
            return 1;
        };
        DefaultLineTokens.prototype.getTokenStartIndex = function (tokenIndex) {
            return 0;
        };
        DefaultLineTokens.prototype.getTokenType = function (tokenIndex) {
            return strings.empty;
        };
        DefaultLineTokens.prototype.getTokenEndIndex = function (tokenIndex, textLength) {
            return textLength;
        };
        DefaultLineTokens.prototype.equals = function (other) {
            return this === other;
        };
        DefaultLineTokens.prototype.findIndexOfOffset = function (offset) {
            return 0;
        };
        DefaultLineTokens.prototype.inflate = function () {
            return [new viewLineToken_1.ViewLineToken(0, '')];
        };
        DefaultLineTokens.prototype.sliceAndInflate = function (startOffset, endOffset, deltaStartIndex) {
            return [new viewLineToken_1.ViewLineToken(0, '')];
        };
        DefaultLineTokens.INSTANCE = new DefaultLineTokens();
        return DefaultLineTokens;
    }());
    exports.DefaultLineTokens = DefaultLineTokens;
    function toModeTransitions(topLevelMode, modeTransitions) {
        if (!modeTransitions || modeTransitions.length === 0) {
            return null;
        }
        else if (modeTransitions.length === 1 && modeTransitions[0].startIndex === 0 && modeTransitions[0].mode === topLevelMode) {
            return null;
        }
        return modeTransitions;
    }
});

/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Microsoft Corporation. All rights reserved.
 *  Licensed under the MIT License. See License.txt in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
define(__m[102], __M([1,0,8,82]), function (require, exports, event_1, languageSelector_1) {
    'use strict';
    var LanguageFeatureRegistry = (function () {
        function LanguageFeatureRegistry() {
            this._clock = 0;
            this._entries = [];
            this._onDidChange = new event_1.Emitter();
        }
        Object.defineProperty(LanguageFeatureRegistry.prototype, "onDidChange", {
            get: function () {
                return this._onDidChange.event;
            },
            enumerable: true,
            configurable: true
        });
        LanguageFeatureRegistry.prototype.register = function (selector, provider, isBuiltin) {
            var _this = this;
            if (isBuiltin === void 0) { isBuiltin = false; }
            var entry = {
                selector: selector,
                provider: provider,
                isBuiltin: isBuiltin,
                _score: -1,
                _time: this._clock++
            };
            this._entries.push(entry);
            this._lastCandidate = undefined;
            this._onDidChange.fire(this._entries.length);
            return {
                dispose: function () {
                    if (entry) {
                        var idx = _this._entries.indexOf(entry);
                        if (idx >= 0) {
                            _this._entries.splice(idx, 1);
                            _this._lastCandidate = undefined;
                            _this._onDidChange.fire(_this._entries.length);
                            entry = undefined;
                        }
                    }
                }
            };
        };
        LanguageFeatureRegistry.prototype.has = function (model) {
            return this.all(model).length > 0;
        };
        LanguageFeatureRegistry.prototype.all = function (model) {
            if (!model || model.isTooLargeForHavingAMode()) {
                return [];
            }
            this._updateScores(model);
            var result = [];
            // from registry
            for (var _i = 0, _a = this._entries; _i < _a.length; _i++) {
                var entry = _a[_i];
                if (entry._score > 0) {
                    result.push(entry.provider);
                }
            }
            return result;
        };
        LanguageFeatureRegistry.prototype.ordered = function (model) {
            var result = [];
            this._orderedForEach(model, function (entry) { return result.push(entry.provider); });
            return result;
        };
        LanguageFeatureRegistry.prototype.orderedGroups = function (model) {
            var result = [];
            var lastBucket;
            var lastBucketScore;
            this._orderedForEach(model, function (entry) {
                if (lastBucket && lastBucketScore === entry._score) {
                    lastBucket.push(entry.provider);
                }
                else {
                    lastBucketScore = entry._score;
                    lastBucket = [entry.provider];
                    result.push(lastBucket);
                }
            });
            return result;
        };
        LanguageFeatureRegistry.prototype._orderedForEach = function (model, callback) {
            if (!model || model.isTooLargeForHavingAMode()) {
                return;
            }
            this._updateScores(model);
            for (var from = 0; from < this._entries.length; from++) {
                var entry = this._entries[from];
                if (entry._score > 0) {
                    callback(entry);
                }
            }
        };
        LanguageFeatureRegistry.prototype._updateScores = function (model) {
            var candidate = {
                uri: model.uri.toString(),
                language: model.getModeId()
            };
            if (this._lastCandidate
                && this._lastCandidate.language === candidate.language
                && this._lastCandidate.uri === candidate.uri) {
                // nothing has changed
                return;
            }
            this._lastCandidate = candidate;
            for (var _i = 0, _a = this._entries; _i < _a.length; _i++) {
                var entry = _a[_i];
                entry._score = languageSelector_1.score(entry.selector, model.uri, model.getModeId());
                if (entry.isBuiltin && entry._score > 0) {
                    entry._score = .5;
                    entry._time = -1;
                }
            }
            // needs sorting
            this._entries.sort(LanguageFeatureRegistry._compareByScoreAndTime);
        };
        LanguageFeatureRegistry._compareByScoreAndTime = function (a, b) {
            if (a._score < b._score) {
                return 1;
            }
            else if (a._score > b._score) {
                return -1;
            }
            else if (a._time < b._time) {
                return 1;
            }
            else if (a._time > b._time) {
                return -1;
            }
            else {
                return 0;
            }
        };
        return LanguageFeatureRegistry;
    }());
    Object.defineProperty(exports, "__esModule", { value: true });
    exports.default = LanguageFeatureRegistry;
});

define(__m[21], __M([1,0,102]), function (require, exports, languageFeatureRegistry_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    /**
     * A document highlight kind.
     */
    (function (DocumentHighlightKind) {
        /**
         * A textual occurrence.
         */
        DocumentHighlightKind[DocumentHighlightKind["Text"] = 0] = "Text";
        /**
         * Read-access of a symbol, like reading a variable.
         */
        DocumentHighlightKind[DocumentHighlightKind["Read"] = 1] = "Read";
        /**
         * Write-access of a symbol, like writing to a variable.
         */
        DocumentHighlightKind[DocumentHighlightKind["Write"] = 2] = "Write";
    })(exports.DocumentHighlightKind || (exports.DocumentHighlightKind = {}));
    var DocumentHighlightKind = exports.DocumentHighlightKind;
    /**
     * A symbol kind.
     */
    (function (SymbolKind) {
        SymbolKind[SymbolKind["File"] = 0] = "File";
        SymbolKind[SymbolKind["Module"] = 1] = "Module";
        SymbolKind[SymbolKind["Namespace"] = 2] = "Namespace";
        SymbolKind[SymbolKind["Package"] = 3] = "Package";
        SymbolKind[SymbolKind["Class"] = 4] = "Class";
        SymbolKind[SymbolKind["Method"] = 5] = "Method";
        SymbolKind[SymbolKind["Property"] = 6] = "Property";
        SymbolKind[SymbolKind["Field"] = 7] = "Field";
        SymbolKind[SymbolKind["Constructor"] = 8] = "Constructor";
        SymbolKind[SymbolKind["Enum"] = 9] = "Enum";
        SymbolKind[SymbolKind["Interface"] = 10] = "Interface";
        SymbolKind[SymbolKind["Function"] = 11] = "Function";
        SymbolKind[SymbolKind["Variable"] = 12] = "Variable";
        SymbolKind[SymbolKind["Constant"] = 13] = "Constant";
        SymbolKind[SymbolKind["String"] = 14] = "String";
        SymbolKind[SymbolKind["Number"] = 15] = "Number";
        SymbolKind[SymbolKind["Boolean"] = 16] = "Boolean";
        SymbolKind[SymbolKind["Array"] = 17] = "Array";
        SymbolKind[SymbolKind["Object"] = 18] = "Object";
        SymbolKind[SymbolKind["Key"] = 19] = "Key";
        SymbolKind[SymbolKind["Null"] = 20] = "Null";
    })(exports.SymbolKind || (exports.SymbolKind = {}));
    var SymbolKind = exports.SymbolKind;
    /**
     * @internal
     */
    var SymbolKind;
    (function (SymbolKind) {
        /**
         * @internal
         */
        function from(kind) {
            switch (kind) {
                case SymbolKind.Method:
                    return 'method';
                case SymbolKind.Function:
                    return 'function';
                case SymbolKind.Constructor:
                    return 'constructor';
                case SymbolKind.Variable:
                    return 'variable';
                case SymbolKind.Class:
                    return 'class';
                case SymbolKind.Interface:
                    return 'interface';
                case SymbolKind.Namespace:
                    return 'namespace';
                case SymbolKind.Package:
                    return 'package';
                case SymbolKind.Module:
                    return 'module';
                case SymbolKind.Property:
                    return 'property';
                case SymbolKind.Enum:
                    return 'enum';
                case SymbolKind.String:
                    return 'string';
                case SymbolKind.File:
                    return 'file';
                case SymbolKind.Array:
                    return 'array';
                case SymbolKind.Number:
                    return 'number';
                case SymbolKind.Boolean:
                    return 'boolean';
                case SymbolKind.Object:
                    return 'object';
                case SymbolKind.Key:
                    return 'key';
                case SymbolKind.Null:
                    return 'null';
            }
            return 'property';
        }
        SymbolKind.from = from;
        /**
         * @internal
         */
        function to(type) {
            switch (type) {
                case 'method':
                    return SymbolKind.Method;
                case 'function':
                    return SymbolKind.Function;
                case 'constructor':
                    return SymbolKind.Constructor;
                case 'variable':
                    return SymbolKind.Variable;
                case 'class':
                    return SymbolKind.Class;
                case 'interface':
                    return SymbolKind.Interface;
                case 'namespace':
                    return SymbolKind.Namespace;
                case 'package':
                    return SymbolKind.Package;
                case 'module':
                    return SymbolKind.Module;
                case 'property':
                    return SymbolKind.Property;
                case 'enum':
                    return SymbolKind.Enum;
                case 'string':
                    return SymbolKind.String;
                case 'file':
                    return SymbolKind.File;
                case 'array':
                    return SymbolKind.Array;
                case 'number':
                    return SymbolKind.Number;
                case 'boolean':
                    return SymbolKind.Boolean;
                case 'object':
                    return SymbolKind.Object;
                case 'key':
                    return SymbolKind.Key;
                case 'null':
                    return SymbolKind.Null;
            }
            return SymbolKind.Property;
        }
        SymbolKind.to = to;
    })(SymbolKind = exports.SymbolKind || (exports.SymbolKind = {}));
    /**
     * Describes what to do with the indentation when pressing Enter.
     */
    (function (IndentAction) {
        /**
         * Insert new line and copy the previous line's indentation.
         */
        IndentAction[IndentAction["None"] = 0] = "None";
        /**
         * Insert new line and indent once (relative to the previous line's indentation).
         */
        IndentAction[IndentAction["Indent"] = 1] = "Indent";
        /**
         * Insert two new lines:
         *  - the first one indented which will hold the cursor
         *  - the second one at the same indentation level
         */
        IndentAction[IndentAction["IndentOutdent"] = 2] = "IndentOutdent";
        /**
         * Insert new line and outdent once (relative to the previous line's indentation).
         */
        IndentAction[IndentAction["Outdent"] = 3] = "Outdent";
    })(exports.IndentAction || (exports.IndentAction = {}));
    var IndentAction = exports.IndentAction;
    // --- feature registries ------
    /**
     * @internal
     */
    exports.ReferenceProviderRegistry = new languageFeatureRegistry_1.default();
    /**
     * @internal
     */
    exports.RenameProviderRegistry = new languageFeatureRegistry_1.default();
    /**
     * @internal
     */
    exports.SuggestRegistry = new languageFeatureRegistry_1.default();
    /**
     * @internal
     */
    exports.SignatureHelpProviderRegistry = new languageFeatureRegistry_1.default();
    /**
     * @internal
     */
    exports.HoverProviderRegistry = new languageFeatureRegistry_1.default();
    /**
     * @internal
     */
    exports.DocumentSymbolProviderRegistry = new languageFeatureRegistry_1.default();
    /**
     * @internal
     */
    exports.DocumentHighlightProviderRegistry = new languageFeatureRegistry_1.default();
    /**
     * @internal
     */
    exports.DefinitionProviderRegistry = new languageFeatureRegistry_1.default();
    /**
     * @internal
     */
    exports.CodeLensProviderRegistry = new languageFeatureRegistry_1.default();
    /**
     * @internal
     */
    exports.CodeActionProviderRegistry = new languageFeatureRegistry_1.default();
    /**
     * @internal
     */
    exports.DocumentFormattingEditProviderRegistry = new languageFeatureRegistry_1.default();
    /**
     * @internal
     */
    exports.DocumentRangeFormattingEditProviderRegistry = new languageFeatureRegistry_1.default();
    /**
     * @internal
     */
    exports.OnTypeFormattingEditProviderRegistry = new languageFeatureRegistry_1.default();
    /**
     * @internal
     */
    exports.LinkProviderRegistry = new languageFeatureRegistry_1.default();
});

define(__m[106], __M([1,0,2,3,21,14]), function (require, exports, errors_1, strings, modes_1, supports_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var OnEnterSupport = (function () {
        function OnEnterSupport(registry, modeId, opts) {
            this._registry = registry;
            opts = opts || {};
            opts.brackets = opts.brackets || [
                ['(', ')'],
                ['{', '}'],
                ['[', ']']
            ];
            this._modeId = modeId;
            this._brackets = opts.brackets.map(function (bracket) {
                return {
                    open: bracket[0],
                    openRegExp: OnEnterSupport._createOpenBracketRegExp(bracket[0]),
                    close: bracket[1],
                    closeRegExp: OnEnterSupport._createCloseBracketRegExp(bracket[1]),
                };
            });
            this._regExpRules = opts.regExpRules || [];
            this._indentationRules = opts.indentationRules;
        }
        OnEnterSupport.prototype.onEnter = function (model, position) {
            var _this = this;
            var context = model.getLineContext(position.lineNumber);
            return supports_1.handleEvent(context, position.column - 1, function (nestedModeId, context, offset) {
                if (_this._modeId === nestedModeId) {
                    return _this._onEnter(model, position);
                }
                var onEnterSupport = _this._registry.getOnEnterSupport(nestedModeId);
                if (onEnterSupport) {
                    return onEnterSupport.onEnter(model, position);
                }
                return null;
            });
        };
        OnEnterSupport.prototype._onEnter = function (model, position) {
            var lineText = model.getLineContent(position.lineNumber);
            var beforeEnterText = lineText.substr(0, position.column - 1);
            var afterEnterText = lineText.substr(position.column - 1);
            var oneLineAboveText = position.lineNumber === 1 ? '' : model.getLineContent(position.lineNumber - 1);
            return this._actualOnEnter(oneLineAboveText, beforeEnterText, afterEnterText);
        };
        OnEnterSupport.prototype._actualOnEnter = function (oneLineAboveText, beforeEnterText, afterEnterText) {
            // (1): `regExpRules`
            for (var i = 0, len = this._regExpRules.length; i < len; i++) {
                var rule = this._regExpRules[i];
                if (rule.beforeText.test(beforeEnterText)) {
                    if (rule.afterText) {
                        if (rule.afterText.test(afterEnterText)) {
                            return rule.action;
                        }
                    }
                    else {
                        return rule.action;
                    }
                }
            }
            // (2): Special indent-outdent
            if (beforeEnterText.length > 0 && afterEnterText.length > 0) {
                for (var i = 0, len = this._brackets.length; i < len; i++) {
                    var bracket = this._brackets[i];
                    if (bracket.openRegExp.test(beforeEnterText) && bracket.closeRegExp.test(afterEnterText)) {
                        return OnEnterSupport._INDENT_OUTDENT;
                    }
                }
            }
            // (3): Indentation Support
            if (this._indentationRules) {
                if (this._indentationRules.increaseIndentPattern && this._indentationRules.increaseIndentPattern.test(beforeEnterText)) {
                    return OnEnterSupport._INDENT;
                }
                if (this._indentationRules.indentNextLinePattern && this._indentationRules.indentNextLinePattern.test(beforeEnterText)) {
                    return OnEnterSupport._INDENT;
                }
                if (/^\s/.test(beforeEnterText)) {
                    // No reason to run regular expressions if there is nothing to outdent from
                    if (this._indentationRules.decreaseIndentPattern && this._indentationRules.decreaseIndentPattern.test(afterEnterText)) {
                        return OnEnterSupport._OUTDENT;
                    }
                    if (this._indentationRules.indentNextLinePattern && this._indentationRules.indentNextLinePattern.test(oneLineAboveText)) {
                        return OnEnterSupport._OUTDENT;
                    }
                }
            }
            // (4): Open bracket based logic
            if (beforeEnterText.length > 0) {
                for (var i = 0, len = this._brackets.length; i < len; i++) {
                    var bracket = this._brackets[i];
                    if (bracket.openRegExp.test(beforeEnterText)) {
                        return OnEnterSupport._INDENT;
                    }
                }
            }
            return null;
        };
        OnEnterSupport._createOpenBracketRegExp = function (bracket) {
            var str = strings.escapeRegExpCharacters(bracket);
            if (!/\B/.test(str.charAt(0))) {
                str = '\\b' + str;
            }
            str += '\\s*$';
            return OnEnterSupport._safeRegExp(str);
        };
        OnEnterSupport._createCloseBracketRegExp = function (bracket) {
            var str = strings.escapeRegExpCharacters(bracket);
            if (!/\B/.test(str.charAt(str.length - 1))) {
                str = str + '\\b';
            }
            str = '^\\s*' + str;
            return OnEnterSupport._safeRegExp(str);
        };
        OnEnterSupport._safeRegExp = function (def) {
            try {
                return new RegExp(def);
            }
            catch (err) {
                errors_1.onUnexpectedError(err);
                return null;
            }
        };
        OnEnterSupport._INDENT = { indentAction: modes_1.IndentAction.Indent };
        OnEnterSupport._INDENT_OUTDENT = { indentAction: modes_1.IndentAction.IndentOutdent };
        OnEnterSupport._OUTDENT = { indentAction: modes_1.IndentAction.Outdent };
        return OnEnterSupport;
    }());
    exports.OnEnterSupport = OnEnterSupport;
});

define(__m[31], __M([1,0,21,94,99,106,28,8,2,22,3,17]), function (require, exports, modes_1, characterPair_1, electricCharacter_1, onEnter_1, richEditBrackets_1, event_1, errors_1, position_1, strings, wordHelper_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var RichEditSupport = (function () {
        function RichEditSupport(modeId, previous, rawConf) {
            var prev = null;
            if (previous) {
                prev = previous._conf;
            }
            this._conf = RichEditSupport._mergeConf(prev, rawConf);
            if (this._conf.brackets) {
                this.brackets = new richEditBrackets_1.RichEditBrackets(modeId, this._conf.brackets);
            }
            this._handleOnEnter(modeId, this._conf);
            this._handleComments(modeId, this._conf);
            if (this._conf.autoClosingPairs) {
                this.characterPair = new characterPair_1.CharacterPairSupport(exports.LanguageConfigurationRegistry, modeId, this._conf);
            }
            if (this._conf.__electricCharacterSupport || this._conf.brackets) {
                this.electricCharacter = new electricCharacter_1.BracketElectricCharacterSupport(exports.LanguageConfigurationRegistry, modeId, this.brackets, this._conf.__electricCharacterSupport);
            }
            this.wordDefinition = this._conf.wordPattern || wordHelper_1.DEFAULT_WORD_REGEXP;
        }
        RichEditSupport._mergeConf = function (prev, current) {
            return {
                comments: (prev ? current.comments || prev.comments : current.comments),
                brackets: (prev ? current.brackets || prev.brackets : current.brackets),
                wordPattern: (prev ? current.wordPattern || prev.wordPattern : current.wordPattern),
                indentationRules: (prev ? current.indentationRules || prev.indentationRules : current.indentationRules),
                onEnterRules: (prev ? current.onEnterRules || prev.onEnterRules : current.onEnterRules),
                autoClosingPairs: (prev ? current.autoClosingPairs || prev.autoClosingPairs : current.autoClosingPairs),
                surroundingPairs: (prev ? current.surroundingPairs || prev.surroundingPairs : current.surroundingPairs),
                __electricCharacterSupport: (prev ? current.__electricCharacterSupport || prev.__electricCharacterSupport : current.__electricCharacterSupport),
            };
        };
        RichEditSupport.prototype._handleOnEnter = function (modeId, conf) {
            // on enter
            var onEnter = {};
            var empty = true;
            if (conf.brackets) {
                empty = false;
                onEnter.brackets = conf.brackets;
            }
            if (conf.indentationRules) {
                empty = false;
                onEnter.indentationRules = conf.indentationRules;
            }
            if (conf.onEnterRules) {
                empty = false;
                onEnter.regExpRules = conf.onEnterRules;
            }
            if (!empty) {
                this.onEnter = new onEnter_1.OnEnterSupport(exports.LanguageConfigurationRegistry, modeId, onEnter);
            }
        };
        RichEditSupport.prototype._handleComments = function (modeId, conf) {
            var commentRule = conf.comments;
            // comment configuration
            if (commentRule) {
                this.comments = {};
                if (commentRule.lineComment) {
                    this.comments.lineCommentToken = commentRule.lineComment;
                }
                if (commentRule.blockComment) {
                    var _a = commentRule.blockComment, blockStart = _a[0], blockEnd = _a[1];
                    this.comments.blockCommentStartToken = blockStart;
                    this.comments.blockCommentEndToken = blockEnd;
                }
            }
        };
        return RichEditSupport;
    }());
    exports.RichEditSupport = RichEditSupport;
    var LanguageConfigurationRegistryImpl = (function () {
        function LanguageConfigurationRegistryImpl() {
            this._onDidChange = new event_1.Emitter();
            this.onDidChange = this._onDidChange.event;
            this._entries = Object.create(null);
        }
        LanguageConfigurationRegistryImpl.prototype.register = function (languageId, configuration) {
            var previous = this._entries[languageId] || null;
            this._entries[languageId] = new RichEditSupport(languageId, previous, configuration);
            this._onDidChange.fire(void 0);
            return {
                dispose: function () { }
            };
        };
        LanguageConfigurationRegistryImpl.prototype._getRichEditSupport = function (modeId) {
            return this._entries[modeId];
        };
        LanguageConfigurationRegistryImpl.prototype.getElectricCharacterSupport = function (modeId) {
            var value = this._getRichEditSupport(modeId);
            if (!value) {
                return null;
            }
            return value.electricCharacter || null;
        };
        LanguageConfigurationRegistryImpl.prototype.getComments = function (modeId) {
            var value = this._getRichEditSupport(modeId);
            if (!value) {
                return null;
            }
            return value.comments || null;
        };
        LanguageConfigurationRegistryImpl.prototype.getCharacterPairSupport = function (modeId) {
            var value = this._getRichEditSupport(modeId);
            if (!value) {
                return null;
            }
            return value.characterPair || null;
        };
        LanguageConfigurationRegistryImpl.prototype.getWordDefinition = function (modeId) {
            var value = this._getRichEditSupport(modeId);
            if (!value) {
                return null;
            }
            return value.wordDefinition || null;
        };
        LanguageConfigurationRegistryImpl.prototype.getOnEnterSupport = function (modeId) {
            var value = this._getRichEditSupport(modeId);
            if (!value) {
                return null;
            }
            return value.onEnter || null;
        };
        LanguageConfigurationRegistryImpl.prototype.getRawEnterActionAtPosition = function (model, lineNumber, column) {
            var result;
            var onEnterSupport = this.getOnEnterSupport(model.getMode().getId());
            if (onEnterSupport) {
                try {
                    result = onEnterSupport.onEnter(model, new position_1.Position(lineNumber, column));
                }
                catch (e) {
                    errors_1.onUnexpectedError(e);
                }
            }
            return result;
        };
        LanguageConfigurationRegistryImpl.prototype.getEnterActionAtPosition = function (model, lineNumber, column) {
            var lineText = model.getLineContent(lineNumber);
            var indentation = strings.getLeadingWhitespace(lineText);
            if (indentation.length > column - 1) {
                indentation = indentation.substring(0, column - 1);
            }
            var enterAction = this.getRawEnterActionAtPosition(model, lineNumber, column);
            if (!enterAction) {
                enterAction = {
                    indentAction: modes_1.IndentAction.None,
                    appendText: '',
                };
            }
            else {
                if (!enterAction.appendText) {
                    if ((enterAction.indentAction === modes_1.IndentAction.Indent) ||
                        (enterAction.indentAction === modes_1.IndentAction.IndentOutdent)) {
                        enterAction.appendText = '\t';
                    }
                    else {
                        enterAction.appendText = '';
                    }
                }
            }
            if (enterAction.removeText) {
                indentation = indentation.substring(0, indentation.length - 1);
            }
            return {
                enterAction: enterAction,
                indentation: indentation
            };
        };
        LanguageConfigurationRegistryImpl.prototype.getBracketsSupport = function (modeId) {
            var value = this._getRichEditSupport(modeId);
            if (!value) {
                return null;
            }
            return value.brackets || null;
        };
        return LanguageConfigurationRegistryImpl;
    }());
    exports.LanguageConfigurationRegistryImpl = LanguageConfigurationRegistryImpl;
    exports.LanguageConfigurationRegistry = new LanguageConfigurationRegistryImpl();
});

define(__m[64], __M([1,0,18,31,17]), function (require, exports, modeTransition_1, languageConfigurationRegistry_1, wordHelper_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var WordHelper = (function () {
        function WordHelper() {
        }
        WordHelper._safeGetWordDefinition = function (mode) {
            return languageConfigurationRegistry_1.LanguageConfigurationRegistry.getWordDefinition(mode.getId());
        };
        WordHelper.massageWordDefinitionOf = function (mode) {
            return wordHelper_1.ensureValidWordDefinition(WordHelper._safeGetWordDefinition(mode));
        };
        WordHelper._getWordAtColumn = function (txt, column, modeIndex, modeTransitions) {
            var modeStartIndex = modeTransitions[modeIndex].startIndex, modeEndIndex = (modeIndex + 1 < modeTransitions.length ? modeTransitions[modeIndex + 1].startIndex : txt.length), mode = modeTransitions[modeIndex].mode;
            return wordHelper_1.getWordAtText(column, WordHelper.massageWordDefinitionOf(mode), txt.substring(modeStartIndex, modeEndIndex), modeStartIndex);
        };
        WordHelper.getWordAtPosition = function (textSource, position) {
            if (!textSource._lineIsTokenized(position.lineNumber)) {
                return wordHelper_1.getWordAtText(position.column, WordHelper.massageWordDefinitionOf(textSource.getMode()), textSource.getLineContent(position.lineNumber), 0);
            }
            var result = null;
            var txt = textSource.getLineContent(position.lineNumber), modeTransitions = textSource._getLineModeTransitions(position.lineNumber), columnIndex = position.column - 1, modeIndex = modeTransition_1.ModeTransition.findIndexInSegmentsArray(modeTransitions, columnIndex);
            result = WordHelper._getWordAtColumn(txt, position.column, modeIndex, modeTransitions);
            if (!result && modeIndex > 0 && modeTransitions[modeIndex].startIndex === columnIndex) {
                // The position is right at the beginning of `modeIndex`, so try looking at `modeIndex` - 1 too
                result = WordHelper._getWordAtColumn(txt, position.column, modeIndex - 1, modeTransitions);
            }
            return result;
        };
        return WordHelper;
    }());
    exports.WordHelper = WordHelper;
});

define(__m[65], __M([7,6]), function(nls, data) { return nls.create("vs/base/common/severity", data); });
define(__m[32], __M([1,0,65,3]), function (require, exports, nls, strings) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var Severity;
    (function (Severity) {
        Severity[Severity["Ignore"] = 0] = "Ignore";
        Severity[Severity["Info"] = 1] = "Info";
        Severity[Severity["Warning"] = 2] = "Warning";
        Severity[Severity["Error"] = 3] = "Error";
    })(Severity || (Severity = {}));
    var Severity;
    (function (Severity) {
        var _error = 'error', _warning = 'warning', _warn = 'warn', _info = 'info';
        var _displayStrings = Object.create(null);
        _displayStrings[Severity.Error] = nls.localize(0, null);
        _displayStrings[Severity.Warning] = nls.localize(1, null);
        _displayStrings[Severity.Info] = nls.localize(2, null);
        /**
         * Parses 'error', 'warning', 'warn', 'info' in call casings
         * and falls back to ignore.
         */
        function fromValue(value) {
            if (!value) {
                return Severity.Ignore;
            }
            if (strings.equalsIgnoreCase(_error, value)) {
                return Severity.Error;
            }
            if (strings.equalsIgnoreCase(_warning, value) || strings.equalsIgnoreCase(_warn, value)) {
                return Severity.Warning;
            }
            if (strings.equalsIgnoreCase(_info, value)) {
                return Severity.Info;
            }
            return Severity.Ignore;
        }
        Severity.fromValue = fromValue;
        function toString(value) {
            return _displayStrings[value] || strings.empty;
        }
        Severity.toString = toString;
        function compare(a, b) {
            return b - a;
        }
        Severity.compare = compare;
    })(Severity || (Severity = {}));
    Object.defineProperty(exports, "__esModule", { value: true });
    exports.default = Severity;
});

define(__m[67], __M([7,6]), function(nls, data) { return nls.create("vs/editor/common/config/defaultConfig", data); });
define(__m[68], __M([1,0,67,10,17]), function (require, exports, nls, platform, wordHelper_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    exports.DEFAULT_INDENTATION = {
        tabSize: 4,
        insertSpaces: true,
        detectIndentation: true
    };
    exports.DEFAULT_TRIM_AUTO_WHITESPACE = true;
    var DEFAULT_WINDOWS_FONT_FAMILY = 'Consolas, \'Courier New\', monospace';
    var DEFAULT_MAC_FONT_FAMILY = 'Menlo, Monaco, \'Courier New\', monospace';
    var DEFAULT_LINUX_FONT_FAMILY = '\'Droid Sans Mono\', \'Courier New\', monospace, \'Droid Sans Fallback\'';
    /**
     * Determined from empirical observations.
     */
    exports.GOLDEN_LINE_HEIGHT_RATIO = platform.isMacintosh ? 1.5 : 1.35;
    var ConfigClass = (function () {
        function ConfigClass() {
            this.editor = {
                experimentalScreenReader: true,
                rulers: [],
                wordSeparators: wordHelper_1.USUAL_WORD_SEPARATORS,
                selectionClipboard: true,
                ariaLabel: nls.localize(0, null),
                lineNumbers: true,
                selectOnLineNumbers: true,
                lineNumbersMinChars: 5,
                glyphMargin: false,
                lineDecorationsWidth: 10,
                revealHorizontalRightPadding: 30,
                roundedSelection: true,
                theme: 'vs',
                readOnly: false,
                scrollbar: {
                    verticalScrollbarSize: 14,
                    horizontal: 'auto',
                    useShadows: true,
                    verticalHasArrows: false,
                    horizontalHasArrows: false
                },
                overviewRulerLanes: 2,
                cursorBlinking: 'blink',
                cursorStyle: 'line',
                fontLigatures: false,
                disableTranslate3d: false,
                hideCursorInOverviewRuler: false,
                scrollBeyondLastLine: true,
                automaticLayout: false,
                wrappingColumn: 300,
                wrappingIndent: 'same',
                wordWrapBreakBeforeCharacters: '([{‘“〈《「『【〔（［｛｢£¥＄￡￥+＋',
                wordWrapBreakAfterCharacters: ' \t})]?|&,;¢°′″‰℃、。｡､￠，．：；？！％・･ゝゞヽヾーァィゥェォッャュョヮヵヶぁぃぅぇぉっゃゅょゎゕゖㇰㇱㇲㇳㇴㇵㇶㇷㇸㇹㇺㇻㇼㇽㇾㇿ々〻ｧｨｩｪｫｬｭｮｯｰ’”〉》」』】〕）］｝｣',
                wordWrapBreakObtrusiveCharacters: '.',
                tabFocusMode: false,
                // Features
                hover: true,
                contextmenu: true,
                mouseWheelScrollSensitivity: 1,
                quickSuggestions: true,
                quickSuggestionsDelay: 10,
                parameterHints: true,
                iconsInSuggestions: true,
                autoClosingBrackets: true,
                formatOnType: false,
                suggestOnTriggerCharacters: true,
                acceptSuggestionOnEnter: true,
                selectionHighlight: true,
                outlineMarkers: false,
                referenceInfos: true,
                folding: true,
                renderWhitespace: false,
                indentGuides: false,
                useTabStops: true,
                fontFamily: (platform.isMacintosh ? DEFAULT_MAC_FONT_FAMILY : (platform.isLinux ? DEFAULT_LINUX_FONT_FAMILY : DEFAULT_WINDOWS_FONT_FAMILY)),
                fontSize: (platform.isMacintosh ? 12 : 14),
                lineHeight: 0
            };
        }
        return ConfigClass;
    }());
    exports.DefaultConfig = new ConfigClass();
});






define(__m[45], __M([1,0,15,3,22,29,19,57,95,68,108]), function (require, exports, eventEmitter_1, strings, position_1, range_1, editorCommon, modelLine_1, indentationGuesser_1, defaultConfig_1, prefixSumComputer_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var LIMIT_FIND_COUNT = 999;
    exports.LONG_LINE_BOUNDARY = 1000;
    var TextModel = (function (_super) {
        __extends(TextModel, _super);
        function TextModel(allowedEventTypes, rawText) {
            allowedEventTypes.push(editorCommon.EventType.ModelRawContentChanged, editorCommon.EventType.ModelOptionsChanged);
            _super.call(this, allowedEventTypes);
            this._shouldSimplifyMode = (rawText.length > TextModel.MODEL_SYNC_LIMIT);
            this._shouldDenyMode = (rawText.length > TextModel.MODEL_TOKENIZATION_LIMIT);
            this._options = rawText.options;
            this._constructLines(rawText);
            this._setVersionId(1);
            this._isDisposed = false;
            this._isDisposing = false;
        }
        TextModel.prototype.isTooLargeForHavingAMode = function () {
            return this._shouldDenyMode;
        };
        TextModel.prototype.isTooLargeForHavingARichMode = function () {
            return this._shouldSimplifyMode;
        };
        TextModel.prototype.getOptions = function () {
            return this._options;
        };
        TextModel.prototype.updateOptions = function (newOpts) {
            var somethingChanged = false;
            var changed = {
                tabSize: false,
                insertSpaces: false,
                trimAutoWhitespace: false
            };
            if (typeof newOpts.insertSpaces !== 'undefined') {
                if (this._options.insertSpaces !== newOpts.insertSpaces) {
                    somethingChanged = true;
                    changed.insertSpaces = true;
                    this._options.insertSpaces = newOpts.insertSpaces;
                }
            }
            if (typeof newOpts.tabSize !== 'undefined') {
                if (this._options.tabSize !== newOpts.tabSize) {
                    somethingChanged = true;
                    changed.tabSize = true;
                    this._options.tabSize = newOpts.tabSize;
                }
            }
            if (typeof newOpts.trimAutoWhitespace !== 'undefined') {
                if (this._options.trimAutoWhitespace !== newOpts.trimAutoWhitespace) {
                    somethingChanged = true;
                    changed.trimAutoWhitespace = true;
                    this._options.trimAutoWhitespace = newOpts.trimAutoWhitespace;
                }
            }
            if (somethingChanged) {
                this.emit(editorCommon.EventType.ModelOptionsChanged, changed);
            }
        };
        TextModel.prototype.detectIndentation = function (defaultInsertSpaces, defaultTabSize) {
            var lines = this._lines.map(function (line) { return line.text; });
            var guessedIndentation = indentationGuesser_1.guessIndentation(lines, defaultTabSize, defaultInsertSpaces);
            this.updateOptions({
                insertSpaces: guessedIndentation.insertSpaces,
                tabSize: guessedIndentation.tabSize
            });
        };
        TextModel.prototype._normalizeIndentationFromWhitespace = function (str) {
            var tabSize = this._options.tabSize;
            var insertSpaces = this._options.insertSpaces;
            var spacesCnt = 0;
            for (var i = 0; i < str.length; i++) {
                if (str.charAt(i) === '\t') {
                    spacesCnt += tabSize;
                }
                else {
                    spacesCnt++;
                }
            }
            var result = '';
            if (!insertSpaces) {
                var tabsCnt = Math.floor(spacesCnt / tabSize);
                spacesCnt = spacesCnt % tabSize;
                for (var i = 0; i < tabsCnt; i++) {
                    result += '\t';
                }
            }
            for (var i = 0; i < spacesCnt; i++) {
                result += ' ';
            }
            return result;
        };
        TextModel.prototype.normalizeIndentation = function (str) {
            var firstNonWhitespaceIndex = strings.firstNonWhitespaceIndex(str);
            if (firstNonWhitespaceIndex === -1) {
                firstNonWhitespaceIndex = str.length;
            }
            return this._normalizeIndentationFromWhitespace(str.substring(0, firstNonWhitespaceIndex)) + str.substring(firstNonWhitespaceIndex);
        };
        TextModel.prototype.getOneIndent = function () {
            var tabSize = this._options.tabSize;
            var insertSpaces = this._options.insertSpaces;
            if (insertSpaces) {
                var result = '';
                for (var i = 0; i < tabSize; i++) {
                    result += ' ';
                }
                return result;
            }
            else {
                return '\t';
            }
        };
        TextModel.prototype.getVersionId = function () {
            return this._versionId;
        };
        TextModel.prototype.getAlternativeVersionId = function () {
            return this._alternativeVersionId;
        };
        TextModel.prototype._ensureLineStarts = function () {
            if (!this._lineStarts) {
                var lineStartValues = [];
                var eolLength = this._EOL.length;
                for (var i = 0, len = this._lines.length; i < len; i++) {
                    lineStartValues.push(this._lines[i].text.length + eolLength);
                }
                this._lineStarts = new prefixSumComputer_1.PrefixSumComputer(lineStartValues);
            }
        };
        TextModel.prototype.getOffsetAt = function (rawPosition) {
            var position = this.validatePosition(rawPosition);
            this._ensureLineStarts();
            return this._lineStarts.getAccumulatedValue(position.lineNumber - 2) + position.column - 1;
        };
        TextModel.prototype.getPositionAt = function (offset) {
            offset = Math.floor(offset);
            offset = Math.max(0, offset);
            this._ensureLineStarts();
            var out = this._lineStarts.getIndexOf(offset);
            var lineLength = this._lines[out.index].text.length;
            // Ensure we return a valid position
            return new position_1.Position(out.index + 1, Math.min(out.remainder + 1, lineLength + 1));
        };
        TextModel.prototype._increaseVersionId = function () {
            this._setVersionId(this._versionId + 1);
        };
        TextModel.prototype._setVersionId = function (newVersionId) {
            this._versionId = newVersionId;
            this._alternativeVersionId = this._versionId;
        };
        TextModel.prototype._overwriteAlternativeVersionId = function (newAlternativeVersionId) {
            this._alternativeVersionId = newAlternativeVersionId;
        };
        TextModel.prototype.isDisposed = function () {
            return this._isDisposed;
        };
        TextModel.prototype.dispose = function () {
            this._isDisposed = true;
            // Null out members, such that any use of a disposed model will throw exceptions sooner rather than later
            this._lines = null;
            this._EOL = null;
            this._BOM = null;
            _super.prototype.dispose.call(this);
        };
        TextModel.prototype._createContentChangedFlushEvent = function () {
            return {
                changeType: editorCommon.EventType.ModelRawContentChangedFlush,
                detail: null,
                // TODO@Alex -> remove these fields from here
                versionId: -1,
                isUndoing: false,
                isRedoing: false
            };
        };
        TextModel.prototype._emitContentChanged2 = function (startLineNumber, startColumn, endLineNumber, endColumn, rangeLength, text, isUndoing, isRedoing) {
            var e = {
                range: new range_1.Range(startLineNumber, startColumn, endLineNumber, endColumn),
                rangeLength: rangeLength,
                text: text,
                eol: this._EOL,
                versionId: this.getVersionId(),
                isUndoing: isUndoing,
                isRedoing: isRedoing
            };
            if (!this._isDisposing) {
                this.emit(editorCommon.EventType.ModelContentChanged2, e);
            }
        };
        TextModel.prototype._resetValue = function (e, newValue) {
            this._constructLines(newValue);
            this._increaseVersionId();
            e.detail = this.toRawText();
            e.versionId = this._versionId;
        };
        TextModel.prototype.toRawText = function () {
            return {
                BOM: this._BOM,
                EOL: this._EOL,
                lines: this.getLinesContent(),
                length: this.getValueLength(),
                options: this._options
            };
        };
        TextModel.prototype.equals = function (other) {
            if (this._BOM !== other.BOM) {
                return false;
            }
            if (this._EOL !== other.EOL) {
                return false;
            }
            if (this._lines.length !== other.lines.length) {
                return false;
            }
            for (var i = 0, len = this._lines.length; i < len; i++) {
                if (this._lines[i].text !== other.lines[i]) {
                    return false;
                }
            }
            return true;
        };
        TextModel.prototype.setValue = function (value) {
            if (value === null) {
                // There's nothing to do
                return;
            }
            var rawText = null;
            rawText = TextModel.toRawText(value, {
                tabSize: this._options.tabSize,
                insertSpaces: this._options.insertSpaces,
                trimAutoWhitespace: this._options.trimAutoWhitespace,
                detectIndentation: false,
                defaultEOL: this._options.defaultEOL
            });
            this.setValueFromRawText(rawText);
        };
        TextModel.prototype.setValueFromRawText = function (newValue) {
            if (newValue === null) {
                // There's nothing to do
                return;
            }
            var oldFullModelRange = this.getFullModelRange();
            var oldModelValueLength = this.getValueLengthInRange(oldFullModelRange);
            var endLineNumber = this.getLineCount();
            var endColumn = this.getLineMaxColumn(endLineNumber);
            var e = this._createContentChangedFlushEvent();
            this._resetValue(e, newValue);
            this._emitModelContentChangedFlushEvent(e);
            this._emitContentChanged2(1, 1, endLineNumber, endColumn, oldModelValueLength, this.getValue(), false, false);
        };
        TextModel.prototype.getValue = function (eol, preserveBOM) {
            if (preserveBOM === void 0) { preserveBOM = false; }
            var fullModelRange = this.getFullModelRange();
            var fullModelValue = this.getValueInRange(fullModelRange, eol);
            if (preserveBOM) {
                return this._BOM + fullModelValue;
            }
            return fullModelValue;
        };
        TextModel.prototype.getValueLength = function (eol, preserveBOM) {
            if (preserveBOM === void 0) { preserveBOM = false; }
            var fullModelRange = this.getFullModelRange();
            var fullModelValue = this.getValueLengthInRange(fullModelRange, eol);
            if (preserveBOM) {
                return this._BOM.length + fullModelValue;
            }
            return fullModelValue;
        };
        TextModel.prototype.getEmptiedValueInRange = function (rawRange, fillCharacter, eol) {
            if (fillCharacter === void 0) { fillCharacter = ''; }
            if (eol === void 0) { eol = editorCommon.EndOfLinePreference.TextDefined; }
            var range = this.validateRange(rawRange);
            if (range.isEmpty()) {
                return '';
            }
            if (range.startLineNumber === range.endLineNumber) {
                return this._repeatCharacter(fillCharacter, range.endColumn - range.startColumn);
            }
            var lineEnding = this._getEndOfLine(eol), startLineIndex = range.startLineNumber - 1, endLineIndex = range.endLineNumber - 1, resultLines = [];
            resultLines.push(this._repeatCharacter(fillCharacter, this._lines[startLineIndex].text.length - range.startColumn + 1));
            for (var i = startLineIndex + 1; i < endLineIndex; i++) {
                resultLines.push(this._repeatCharacter(fillCharacter, this._lines[i].text.length));
            }
            resultLines.push(this._repeatCharacter(fillCharacter, range.endColumn - 1));
            return resultLines.join(lineEnding);
        };
        TextModel.prototype._repeatCharacter = function (fillCharacter, count) {
            var r = '';
            for (var i = 0; i < count; i++) {
                r += fillCharacter;
            }
            return r;
        };
        TextModel.prototype.getValueInRange = function (rawRange, eol) {
            if (eol === void 0) { eol = editorCommon.EndOfLinePreference.TextDefined; }
            var range = this.validateRange(rawRange);
            if (range.isEmpty()) {
                return '';
            }
            if (range.startLineNumber === range.endLineNumber) {
                return this._lines[range.startLineNumber - 1].text.substring(range.startColumn - 1, range.endColumn - 1);
            }
            var lineEnding = this._getEndOfLine(eol), startLineIndex = range.startLineNumber - 1, endLineIndex = range.endLineNumber - 1, resultLines = [];
            resultLines.push(this._lines[startLineIndex].text.substring(range.startColumn - 1));
            for (var i = startLineIndex + 1; i < endLineIndex; i++) {
                resultLines.push(this._lines[i].text);
            }
            resultLines.push(this._lines[endLineIndex].text.substring(0, range.endColumn - 1));
            return resultLines.join(lineEnding);
        };
        TextModel.prototype.getValueLengthInRange = function (rawRange, eol) {
            if (eol === void 0) { eol = editorCommon.EndOfLinePreference.TextDefined; }
            var range = this.validateRange(rawRange);
            if (range.isEmpty()) {
                return 0;
            }
            if (range.startLineNumber === range.endLineNumber) {
                return (range.endColumn - range.startColumn);
            }
            var startOffset = this.getOffsetAt(new position_1.Position(range.startLineNumber, range.startColumn));
            var endOffset = this.getOffsetAt(new position_1.Position(range.endLineNumber, range.endColumn));
            return endOffset - startOffset;
        };
        TextModel.prototype.isDominatedByLongLines = function () {
            var smallLineCharCount = 0, longLineCharCount = 0, i, len, lines = this._lines, lineLength;
            for (i = 0, len = this._lines.length; i < len; i++) {
                lineLength = lines[i].text.length;
                if (lineLength >= exports.LONG_LINE_BOUNDARY) {
                    longLineCharCount += lineLength;
                }
                else {
                    smallLineCharCount += lineLength;
                }
            }
            return (longLineCharCount > smallLineCharCount);
        };
        TextModel.prototype.getLineCount = function () {
            return this._lines.length;
        };
        TextModel.prototype.getLineContent = function (lineNumber) {
            if (lineNumber < 1 || lineNumber > this.getLineCount()) {
                throw new Error('Illegal value ' + lineNumber + ' for `lineNumber`');
            }
            return this._lines[lineNumber - 1].text;
        };
        TextModel.prototype.getLinesContent = function () {
            var r = [];
            for (var i = 0, len = this._lines.length; i < len; i++) {
                r[i] = this._lines[i].text;
            }
            return r;
        };
        TextModel.prototype.getEOL = function () {
            return this._EOL;
        };
        TextModel.prototype.setEOL = function (eol) {
            var newEOL = (eol === editorCommon.EndOfLineSequence.CRLF ? '\r\n' : '\n');
            if (this._EOL === newEOL) {
                // Nothing to do
                return;
            }
            var oldFullModelRange = this.getFullModelRange();
            var oldModelValueLength = this.getValueLengthInRange(oldFullModelRange);
            var endLineNumber = this.getLineCount();
            var endColumn = this.getLineMaxColumn(endLineNumber);
            this._EOL = newEOL;
            this._lineStarts = null;
            this._increaseVersionId();
            var e = this._createContentChangedFlushEvent();
            e.detail = this.toRawText();
            e.versionId = this._versionId;
            this._emitModelContentChangedFlushEvent(e);
            this._emitContentChanged2(1, 1, endLineNumber, endColumn, oldModelValueLength, this.getValue(), false, false);
        };
        TextModel.prototype.getLineMinColumn = function (lineNumber) {
            return 1;
        };
        TextModel.prototype.getLineMaxColumn = function (lineNumber) {
            if (lineNumber < 1 || lineNumber > this.getLineCount()) {
                throw new Error('Illegal value ' + lineNumber + ' for `lineNumber`');
            }
            return this._lines[lineNumber - 1].text.length + 1;
        };
        TextModel.prototype.getLineFirstNonWhitespaceColumn = function (lineNumber) {
            if (lineNumber < 1 || lineNumber > this.getLineCount()) {
                throw new Error('Illegal value ' + lineNumber + ' for `lineNumber`');
            }
            var result = strings.firstNonWhitespaceIndex(this._lines[lineNumber - 1].text);
            if (result === -1) {
                return 0;
            }
            return result + 1;
        };
        TextModel.prototype.getLineLastNonWhitespaceColumn = function (lineNumber) {
            if (lineNumber < 1 || lineNumber > this.getLineCount()) {
                throw new Error('Illegal value ' + lineNumber + ' for `lineNumber`');
            }
            var result = strings.lastNonWhitespaceIndex(this._lines[lineNumber - 1].text);
            if (result === -1) {
                return 0;
            }
            return result + 2;
        };
        TextModel.prototype.validateLineNumber = function (lineNumber) {
            if (lineNumber < 1) {
                lineNumber = 1;
            }
            if (lineNumber > this._lines.length) {
                lineNumber = this._lines.length;
            }
            return lineNumber;
        };
        TextModel.prototype.validatePosition = function (position) {
            var lineNumber = position.lineNumber ? position.lineNumber : 1;
            var column = position.column ? position.column : 1;
            if (lineNumber < 1) {
                lineNumber = 1;
                column = 1;
            }
            else if (lineNumber > this._lines.length) {
                lineNumber = this._lines.length;
                column = this.getLineMaxColumn(lineNumber);
            }
            else {
                var maxColumn = this.getLineMaxColumn(lineNumber);
                if (column < 1) {
                    column = 1;
                }
                else if (column > maxColumn) {
                    column = maxColumn;
                }
            }
            return new position_1.Position(lineNumber, column);
        };
        TextModel.prototype.validateRange = function (range) {
            var start = this.validatePosition(new position_1.Position(range.startLineNumber, range.startColumn));
            var end = this.validatePosition(new position_1.Position(range.endLineNumber, range.endColumn));
            return new range_1.Range(start.lineNumber, start.column, end.lineNumber, end.column);
        };
        TextModel.prototype.modifyPosition = function (rawPosition, offset) {
            return this.getPositionAt(this.getOffsetAt(rawPosition) + offset);
        };
        TextModel.prototype.getFullModelRange = function () {
            var lineCount = this.getLineCount();
            return new range_1.Range(1, 1, lineCount, this.getLineMaxColumn(lineCount));
        };
        TextModel.prototype._emitModelContentChangedFlushEvent = function (e) {
            if (!this._isDisposing) {
                this.emit(editorCommon.EventType.ModelRawContentChanged, e);
            }
        };
        TextModel.toRawText = function (rawText, opts) {
            // Count the number of lines that end with \r\n
            var carriageReturnCnt = 0, lastCarriageReturnIndex = -1;
            while ((lastCarriageReturnIndex = rawText.indexOf('\r', lastCarriageReturnIndex + 1)) !== -1) {
                carriageReturnCnt++;
            }
            // Split the text into lines
            var lines = rawText.split(/\r\n|\r|\n/);
            // Remove the BOM (if present)
            var BOM = '';
            if (strings.startsWithUTF8BOM(lines[0])) {
                BOM = strings.UTF8_BOM_CHARACTER;
                lines[0] = lines[0].substr(1);
            }
            var lineFeedCnt = lines.length - 1;
            var EOL = '';
            if (lineFeedCnt === 0) {
                // This is an empty file or a file with precisely one line
                EOL = (opts.defaultEOL === editorCommon.DefaultEndOfLine.LF ? '\n' : '\r\n');
            }
            else if (carriageReturnCnt > lineFeedCnt / 2) {
                // More than half of the file contains \r\n ending lines
                EOL = '\r\n';
            }
            else {
                // At least one line more ends in \n
                EOL = '\n';
            }
            var resolvedOpts;
            if (opts.detectIndentation) {
                var guessedIndentation = indentationGuesser_1.guessIndentation(lines, opts.tabSize, opts.insertSpaces);
                resolvedOpts = {
                    tabSize: guessedIndentation.tabSize,
                    insertSpaces: guessedIndentation.insertSpaces,
                    trimAutoWhitespace: opts.trimAutoWhitespace,
                    defaultEOL: opts.defaultEOL
                };
            }
            else {
                resolvedOpts = {
                    tabSize: opts.tabSize,
                    insertSpaces: opts.insertSpaces,
                    trimAutoWhitespace: opts.trimAutoWhitespace,
                    defaultEOL: opts.defaultEOL
                };
            }
            return {
                BOM: BOM,
                EOL: EOL,
                lines: lines,
                length: rawText.length,
                options: resolvedOpts
            };
        };
        TextModel.prototype._constructLines = function (rawText) {
            var rawLines = rawText.lines, modelLines = [], i, len;
            for (i = 0, len = rawLines.length; i < len; i++) {
                modelLines.push(new modelLine_1.ModelLine(i + 1, rawLines[i]));
            }
            this._BOM = rawText.BOM;
            this._EOL = rawText.EOL;
            this._lines = modelLines;
            this._lineStarts = null;
        };
        TextModel.prototype._getEndOfLine = function (eol) {
            switch (eol) {
                case editorCommon.EndOfLinePreference.LF:
                    return '\n';
                case editorCommon.EndOfLinePreference.CRLF:
                    return '\r\n';
                case editorCommon.EndOfLinePreference.TextDefined:
                    return this.getEOL();
            }
            throw new Error('Unknown EOL preference');
        };
        TextModel.prototype.findMatches = function (searchString, rawSearchScope, isRegex, matchCase, wholeWord, limitResultCount) {
            if (limitResultCount === void 0) { limitResultCount = LIMIT_FIND_COUNT; }
            var regex = strings.createSafeRegExp(searchString, isRegex, matchCase, wholeWord);
            if (!regex) {
                return [];
            }
            var searchRange;
            if (range_1.Range.isIRange(rawSearchScope)) {
                searchRange = rawSearchScope;
            }
            else {
                searchRange = this.getFullModelRange();
            }
            return this._doFindMatches(searchRange, regex, limitResultCount);
        };
        TextModel.prototype.findNextMatch = function (searchString, rawSearchStart, isRegex, matchCase, wholeWord) {
            var regex = strings.createSafeRegExp(searchString, isRegex, matchCase, wholeWord);
            if (!regex) {
                return null;
            }
            var searchStart = this.validatePosition(rawSearchStart), lineCount = this.getLineCount(), startLineNumber = searchStart.lineNumber, text, r;
            // Look in first line
            text = this._lines[startLineNumber - 1].text.substring(searchStart.column - 1);
            r = this._findMatchInLine(regex, text, startLineNumber, searchStart.column - 1);
            if (r) {
                return r;
            }
            for (var i = 1; i <= lineCount; i++) {
                var lineIndex = (startLineNumber + i - 1) % lineCount;
                text = this._lines[lineIndex].text;
                r = this._findMatchInLine(regex, text, lineIndex + 1, 0);
                if (r) {
                    return r;
                }
            }
            return null;
        };
        TextModel.prototype.findPreviousMatch = function (searchString, rawSearchStart, isRegex, matchCase, wholeWord) {
            var regex = strings.createSafeRegExp(searchString, isRegex, matchCase, wholeWord);
            if (!regex) {
                return null;
            }
            var searchStart = this.validatePosition(rawSearchStart), lineCount = this.getLineCount(), startLineNumber = searchStart.lineNumber, text, r;
            // Look in first line
            text = this._lines[startLineNumber - 1].text.substring(0, searchStart.column - 1);
            r = this._findLastMatchInLine(regex, text, startLineNumber);
            if (r) {
                return r;
            }
            for (var i = 1; i <= lineCount; i++) {
                var lineIndex = (lineCount + startLineNumber - i - 1) % lineCount;
                text = this._lines[lineIndex].text;
                r = this._findLastMatchInLine(regex, text, lineIndex + 1);
                if (r) {
                    return r;
                }
            }
            return null;
        };
        TextModel.prototype._doFindMatches = function (searchRange, searchRegex, limitResultCount) {
            var result = [], text, counter = 0;
            // Early case for a search range that starts & stops on the same line number
            if (searchRange.startLineNumber === searchRange.endLineNumber) {
                text = this._lines[searchRange.startLineNumber - 1].text.substring(searchRange.startColumn - 1, searchRange.endColumn - 1);
                counter = this._findMatchesInLine(searchRegex, text, searchRange.startLineNumber, searchRange.startColumn - 1, counter, result, limitResultCount);
                return result;
            }
            // Collect results from first line
            text = this._lines[searchRange.startLineNumber - 1].text.substring(searchRange.startColumn - 1);
            counter = this._findMatchesInLine(searchRegex, text, searchRange.startLineNumber, searchRange.startColumn - 1, counter, result, limitResultCount);
            // Collect results from middle lines
            for (var lineNumber = searchRange.startLineNumber + 1; lineNumber < searchRange.endLineNumber && counter < limitResultCount; lineNumber++) {
                counter = this._findMatchesInLine(searchRegex, this._lines[lineNumber - 1].text, lineNumber, 0, counter, result, limitResultCount);
            }
            // Collect results from last line
            if (counter < limitResultCount) {
                text = this._lines[searchRange.endLineNumber - 1].text.substring(0, searchRange.endColumn - 1);
                counter = this._findMatchesInLine(searchRegex, text, searchRange.endLineNumber, 0, counter, result, limitResultCount);
            }
            return result;
        };
        TextModel.prototype._findMatchInLine = function (searchRegex, text, lineNumber, deltaOffset) {
            var m = searchRegex.exec(text);
            if (!m) {
                return null;
            }
            return new range_1.Range(lineNumber, m.index + 1 + deltaOffset, lineNumber, m.index + 1 + m[0].length + deltaOffset);
        };
        TextModel.prototype._findLastMatchInLine = function (searchRegex, text, lineNumber) {
            var bestResult = null;
            var m;
            while ((m = searchRegex.exec(text))) {
                var result = new range_1.Range(lineNumber, m.index + 1, lineNumber, m.index + 1 + m[0].length);
                if (result.equalsRange(bestResult)) {
                    break;
                }
                bestResult = result;
            }
            return bestResult;
        };
        TextModel.prototype._findMatchesInLine = function (searchRegex, text, lineNumber, deltaOffset, counter, result, limitResultCount) {
            var m;
            // Reset regex to search from the beginning
            searchRegex.lastIndex = 0;
            do {
                m = searchRegex.exec(text);
                if (m) {
                    var range = new range_1.Range(lineNumber, m.index + 1 + deltaOffset, lineNumber, m.index + 1 + m[0].length + deltaOffset);
                    // Exit early if the regex matches the same range
                    if (range.equalsRange(result[result.length - 1])) {
                        return counter;
                    }
                    result.push(range);
                    counter++;
                    if (counter >= limitResultCount) {
                        return counter;
                    }
                }
            } while (m);
            return counter;
        };
        TextModel.MODEL_SYNC_LIMIT = 5 * 1024 * 1024; // 5 MB
        TextModel.MODEL_TOKENIZATION_LIMIT = 20 * 1024 * 1024; // 20 MB
        TextModel.DEFAULT_CREATION_OPTIONS = {
            tabSize: defaultConfig_1.DEFAULT_INDENTATION.tabSize,
            insertSpaces: defaultConfig_1.DEFAULT_INDENTATION.insertSpaces,
            detectIndentation: false,
            defaultEOL: editorCommon.DefaultEndOfLine.LF,
            trimAutoWhitespace: defaultConfig_1.DEFAULT_TRIM_AUTO_WHITESPACE,
        };
        return TextModel;
    }(eventEmitter_1.OrderGuaranteeEventEmitter));
    exports.TextModel = TextModel;
    var RawText = (function () {
        function RawText() {
        }
        RawText.fromString = function (rawText, opts) {
            return TextModel.toRawText(rawText, opts);
        };
        RawText.fromStringWithModelOptions = function (rawText, model) {
            var opts = model.getOptions();
            return TextModel.toRawText(rawText, {
                tabSize: opts.tabSize,
                insertSpaces: opts.insertSpaces,
                trimAutoWhitespace: opts.trimAutoWhitespace,
                detectIndentation: false,
                defaultEOL: opts.defaultEOL
            });
        };
        return RawText;
    }());
    exports.RawText = RawText;
});

define(__m[70], __M([7,6]), function(nls, data) { return nls.create("vs/editor/common/model/textModelWithTokens", data); });





define(__m[71], __M([1,0,70,24,2,16,61,26,5,19,45,64,69,48,14,28,18,101,56,31]), function (require, exports, nls, async_1, errors_1, lifecycle_1, stopwatch_1, timer, winjs_base_1, editorCommon, textModel_1, textModelWithTokensHelpers_1, tokenIterator_1, nullMode_1, supports_1, richEditBrackets_1, modeTransition_1, lineToken_1, tokensBinaryEncoding_1, languageConfigurationRegistry_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var ModeToModelBinder = (function () {
        function ModeToModelBinder(modePromise, model) {
            var _this = this;
            this._modePromise = modePromise;
            // Create an external mode promise that fires after the mode is set to the model
            this._externalModePromise = new winjs_base_1.TPromise(function (c, e, p) {
                _this._externalModePromise_c = c;
                _this._externalModePromise_e = e;
            }, function () {
                // this promise cannot be canceled
            });
            this._model = model;
            this._isDisposed = false;
            // Ensure asynchronicity
            winjs_base_1.TPromise.timeout(0).then(function () {
                return _this._modePromise;
            }).then(function (mode) {
                if (_this._isDisposed) {
                    _this._externalModePromise_c(false);
                    return;
                }
                var model = _this._model;
                _this.dispose();
                model.setMode(mode);
                model._warmUpTokens();
                _this._externalModePromise_c(true);
            }).done(null, function (err) {
                _this._externalModePromise_e(err);
                errors_1.onUnexpectedError(err);
            });
        }
        ModeToModelBinder.prototype.getModePromise = function () {
            return this._externalModePromise;
        };
        ModeToModelBinder.prototype.dispose = function () {
            this._modePromise = null;
            this._model = null;
            this._isDisposed = true;
        };
        return ModeToModelBinder;
    }());
    var FullModelRetokenizer = (function () {
        function FullModelRetokenizer(retokenizePromise, model) {
            var _this = this;
            this._retokenizePromise = retokenizePromise;
            this._model = model;
            this._isDisposed = false;
            this.isFulfilled = false;
            // Ensure asynchronicity
            winjs_base_1.TPromise.timeout(0).then(function () {
                return _this._retokenizePromise;
            }).then(function () {
                if (_this._isDisposed) {
                    return;
                }
                _this.isFulfilled = true;
                _this._model.onRetokenizerFulfilled();
            }).done(null, errors_1.onUnexpectedError);
        }
        FullModelRetokenizer.prototype.getRange = function () {
            return null;
        };
        FullModelRetokenizer.prototype.dispose = function () {
            this._retokenizePromise = null;
            this._model = null;
            this._isDisposed = true;
        };
        return FullModelRetokenizer;
    }());
    exports.FullModelRetokenizer = FullModelRetokenizer;
    var LineContext = (function () {
        function LineContext(topLevelMode, line) {
            this.modeTransitions = line.getModeTransitions(topLevelMode);
            this._text = line.text;
            this._lineTokens = line.getTokens();
        }
        LineContext.prototype.getLineContent = function () {
            return this._text;
        };
        LineContext.prototype.getTokenCount = function () {
            return this._lineTokens.getTokenCount();
        };
        LineContext.prototype.getTokenStartIndex = function (tokenIndex) {
            return this._lineTokens.getTokenStartIndex(tokenIndex);
        };
        LineContext.prototype.getTokenEndIndex = function (tokenIndex) {
            return this._lineTokens.getTokenEndIndex(tokenIndex, this._text.length);
        };
        LineContext.prototype.getTokenType = function (tokenIndex) {
            return this._lineTokens.getTokenType(tokenIndex);
        };
        LineContext.prototype.getTokenText = function (tokenIndex) {
            var startIndex = this._lineTokens.getTokenStartIndex(tokenIndex);
            var endIndex = this._lineTokens.getTokenEndIndex(tokenIndex, this._text.length);
            return this._text.substring(startIndex, endIndex);
        };
        LineContext.prototype.findIndexOfOffset = function (offset) {
            return this._lineTokens.findIndexOfOffset(offset);
        };
        return LineContext;
    }());
    var TextModelWithTokens = (function (_super) {
        __extends(TextModelWithTokens, _super);
        function TextModelWithTokens(allowedEventTypes, rawText, shouldAutoTokenize, modeOrPromise) {
            var _this = this;
            allowedEventTypes.push(editorCommon.EventType.ModelTokensChanged);
            allowedEventTypes.push(editorCommon.EventType.ModelModeChanged);
            allowedEventTypes.push(editorCommon.EventType.ModelModeSupportChanged);
            _super.call(this, allowedEventTypes, rawText);
            this._shouldAutoTokenize = shouldAutoTokenize;
            this._mode = null;
            this._modeListener = null;
            this._modeToModelBinder = null;
            this._tokensInflatorMap = null;
            this._invalidLineStartIndex = 0;
            this._lastState = null;
            this._revalidateTokensTimeout = -1;
            this._scheduleRetokenizeNow = null;
            this._retokenizers = null;
            if (!modeOrPromise) {
                this._mode = new nullMode_1.NullMode();
            }
            else if (winjs_base_1.TPromise.is(modeOrPromise)) {
                // TODO@Alex: To avoid mode id changes, we check if this promise is resolved
                var promiseValue = modeOrPromise._value;
                if (promiseValue && typeof promiseValue.getId === 'function') {
                    // The promise is already resolved
                    this._mode = this._massageMode(promiseValue);
                    this._resetModeListener(this._mode);
                }
                else {
                    var modePromise = modeOrPromise;
                    this._modeToModelBinder = new ModeToModelBinder(modePromise, this);
                    this._mode = new nullMode_1.NullMode();
                }
            }
            else {
                this._mode = this._massageMode(modeOrPromise);
                this._resetModeListener(this._mode);
            }
            this._revalidateTokensTimeout = -1;
            this._scheduleRetokenizeNow = new async_1.RunOnceScheduler(function () { return _this._retokenizeNow(); }, 200);
            this._retokenizers = [];
            this._resetTokenizationState();
        }
        TextModelWithTokens.prototype.dispose = function () {
            if (this._modeToModelBinder) {
                this._modeToModelBinder.dispose();
                this._modeToModelBinder = null;
            }
            this._resetModeListener(null);
            this._clearTimers();
            this._mode = null;
            this._lastState = null;
            this._tokensInflatorMap = null;
            this._retokenizers = lifecycle_1.dispose(this._retokenizers);
            this._scheduleRetokenizeNow.dispose();
            _super.prototype.dispose.call(this);
        };
        TextModelWithTokens.prototype._massageMode = function (mode) {
            if (this.isTooLargeForHavingAMode()) {
                return new nullMode_1.NullMode();
            }
            if (this.isTooLargeForHavingARichMode()) {
                return mode.toSimplifiedMode();
            }
            return mode;
        };
        TextModelWithTokens.prototype.whenModeIsReady = function () {
            var _this = this;
            if (this._modeToModelBinder) {
                // Still waiting for some mode to load
                return this._modeToModelBinder.getModePromise().then(function () { return _this._mode; });
            }
            return winjs_base_1.TPromise.as(this._mode);
        };
        TextModelWithTokens.prototype.onRetokenizerFulfilled = function () {
            this._scheduleRetokenizeNow.schedule();
        };
        TextModelWithTokens.prototype._retokenizeNow = function () {
            var fulfilled = this._retokenizers.filter(function (r) { return r.isFulfilled; });
            this._retokenizers = this._retokenizers.filter(function (r) { return !r.isFulfilled; });
            var hasFullModel = false;
            for (var i = 0; i < fulfilled.length; i++) {
                if (!fulfilled[i].getRange()) {
                    hasFullModel = true;
                }
            }
            if (hasFullModel) {
                // Just invalidate all the lines
                for (var i = 0, len = this._lines.length; i < len; i++) {
                    this._lines[i].isInvalid = true;
                }
                this._invalidLineStartIndex = 0;
            }
            else {
                var minLineNumber = Number.MAX_VALUE;
                for (var i = 0; i < fulfilled.length; i++) {
                    var range = fulfilled[i].getRange();
                    minLineNumber = Math.min(minLineNumber, range.startLineNumber);
                    for (var lineNumber = range.startLineNumber; lineNumber <= range.endLineNumber; lineNumber++) {
                        this._lines[lineNumber - 1].isInvalid = true;
                    }
                }
                if (minLineNumber - 1 < this._invalidLineStartIndex) {
                    if (this._invalidLineStartIndex < this._lines.length) {
                        this._lines[this._invalidLineStartIndex].isInvalid = true;
                    }
                    this._invalidLineStartIndex = minLineNumber - 1;
                }
            }
            this._beginBackgroundTokenization();
            for (var i = 0; i < fulfilled.length; i++) {
                fulfilled[i].dispose();
            }
        };
        TextModelWithTokens.prototype._createRetokenizer = function (retokenizePromise, lineNumber) {
            return new FullModelRetokenizer(retokenizePromise, this);
        };
        TextModelWithTokens.prototype._resetValue = function (e, newValue) {
            _super.prototype._resetValue.call(this, e, newValue);
            // Cancel tokenization, clear all tokens and begin tokenizing
            this._resetTokenizationState();
        };
        TextModelWithTokens.prototype._resetMode = function (e, newMode) {
            // Cancel tokenization, clear all tokens and begin tokenizing
            this._mode = newMode;
            this._resetModeListener(newMode);
            this._resetTokenizationState();
            this.emitModelTokensChangedEvent(1, this.getLineCount());
        };
        TextModelWithTokens.prototype._resetModeListener = function (newMode) {
            var _this = this;
            if (this._modeListener) {
                this._modeListener.dispose();
                this._modeListener = null;
            }
            if (newMode && typeof newMode.addSupportChangedListener === 'function') {
                this._modeListener = newMode.addSupportChangedListener(function (e) { return _this._onModeSupportChanged(e); });
            }
        };
        TextModelWithTokens.prototype._onModeSupportChanged = function (e) {
            this._emitModelModeSupportChangedEvent(e);
            if (e.tokenizationSupport) {
                this._resetTokenizationState();
                this.emitModelTokensChangedEvent(1, this.getLineCount());
            }
        };
        TextModelWithTokens.prototype._resetTokenizationState = function () {
            this._retokenizers = lifecycle_1.dispose(this._retokenizers);
            this._scheduleRetokenizeNow.cancel();
            this._clearTimers();
            for (var i = 0; i < this._lines.length; i++) {
                this._lines[i].setState(null);
            }
            this._initializeTokenizationState();
        };
        TextModelWithTokens.prototype._clearTimers = function () {
            if (this._revalidateTokensTimeout !== -1) {
                clearTimeout(this._revalidateTokensTimeout);
                this._revalidateTokensTimeout = -1;
            }
        };
        TextModelWithTokens.prototype._initializeTokenizationState = function () {
            // Initialize tokenization states
            var initialState = null;
            if (this._mode.tokenizationSupport) {
                try {
                    initialState = this._mode.tokenizationSupport.getInitialState();
                }
                catch (e) {
                    e.friendlyMessage = TextModelWithTokens.MODE_TOKENIZATION_FAILED_MSG;
                    errors_1.onUnexpectedError(e);
                    this._mode = new nullMode_1.NullMode();
                }
            }
            if (!initialState) {
                initialState = new nullMode_1.NullState(this._mode, null);
            }
            this._lines[0].setState(initialState);
            this._lastState = null;
            this._tokensInflatorMap = new tokensBinaryEncoding_1.TokensInflatorMap();
            this._invalidLineStartIndex = 0;
            this._beginBackgroundTokenization();
        };
        TextModelWithTokens.prototype.getLineTokens = function (lineNumber, inaccurateTokensAcceptable) {
            if (inaccurateTokensAcceptable === void 0) { inaccurateTokensAcceptable = false; }
            if (lineNumber < 1 || lineNumber > this.getLineCount()) {
                throw new Error('Illegal value ' + lineNumber + ' for `lineNumber`');
            }
            if (!inaccurateTokensAcceptable) {
                this._updateTokensUntilLine(lineNumber, true);
            }
            return this._lines[lineNumber - 1].getTokens();
        };
        TextModelWithTokens.prototype.getLineContext = function (lineNumber) {
            if (lineNumber < 1 || lineNumber > this.getLineCount()) {
                throw new Error('Illegal value ' + lineNumber + ' for `lineNumber`');
            }
            this._updateTokensUntilLine(lineNumber, true);
            return new LineContext(this._mode, this._lines[lineNumber - 1]);
        };
        TextModelWithTokens.prototype._getInternalTokens = function (lineNumber) {
            this._updateTokensUntilLine(lineNumber, true);
            return this._lines[lineNumber - 1].getTokens();
        };
        TextModelWithTokens.prototype.getMode = function () {
            return this._mode;
        };
        TextModelWithTokens.prototype.setMode = function (newModeOrPromise) {
            if (!newModeOrPromise) {
                // There's nothing to do
                return;
            }
            if (this._modeToModelBinder) {
                this._modeToModelBinder.dispose();
                this._modeToModelBinder = null;
            }
            if (winjs_base_1.TPromise.is(newModeOrPromise)) {
                this._modeToModelBinder = new ModeToModelBinder(newModeOrPromise, this);
            }
            else {
                var actualNewMode = this._massageMode(newModeOrPromise);
                if (this._mode !== actualNewMode) {
                    var e2 = {
                        oldMode: this._mode,
                        newMode: actualNewMode
                    };
                    this._resetMode(e2, actualNewMode);
                    this._emitModelModeChangedEvent(e2);
                }
            }
        };
        TextModelWithTokens.prototype.getModeIdAtPosition = function (_lineNumber, _column) {
            var validPosition = this.validatePosition({
                lineNumber: _lineNumber,
                column: _column
            });
            var lineNumber = validPosition.lineNumber;
            var column = validPosition.column;
            if (column === 1) {
                return this.getStateBeforeLine(lineNumber).getMode().getId();
            }
            else if (column === this.getLineMaxColumn(lineNumber)) {
                return this.getStateAfterLine(lineNumber).getMode().getId();
            }
            else {
                var modeTransitions = this._getLineModeTransitions(lineNumber);
                var modeTransitionIndex = modeTransition_1.ModeTransition.findIndexInSegmentsArray(modeTransitions, column - 1);
                return modeTransitions[modeTransitionIndex].modeId;
            }
        };
        TextModelWithTokens.prototype._invalidateLine = function (lineIndex) {
            this._lines[lineIndex].isInvalid = true;
            if (lineIndex < this._invalidLineStartIndex) {
                if (this._invalidLineStartIndex < this._lines.length) {
                    this._lines[this._invalidLineStartIndex].isInvalid = true;
                }
                this._invalidLineStartIndex = lineIndex;
                this._beginBackgroundTokenization();
            }
        };
        TextModelWithTokens._toLineTokens = function (tokens) {
            if (!tokens || tokens.length === 0) {
                return [];
            }
            if (tokens[0] instanceof lineToken_1.LineToken) {
                return tokens;
            }
            var result = [];
            for (var i = 0, len = tokens.length; i < len; i++) {
                result[i] = new lineToken_1.LineToken(tokens[i].startIndex, tokens[i].type);
            }
            return result;
        };
        TextModelWithTokens._toModeTransitions = function (modeTransitions) {
            if (!modeTransitions || modeTransitions.length === 0) {
                return [];
            }
            if (modeTransitions[0] instanceof modeTransition_1.ModeTransition) {
                return modeTransitions;
            }
            var result = [];
            for (var i = 0, len = modeTransitions.length; i < len; i++) {
                result[i] = new modeTransition_1.ModeTransition(modeTransitions[i].startIndex, modeTransitions[i].mode);
            }
            return result;
        };
        TextModelWithTokens.prototype._updateLineTokens = function (lineIndex, map, topLevelMode, r) {
            this._lines[lineIndex].setTokens(map, TextModelWithTokens._toLineTokens(r.tokens), topLevelMode, TextModelWithTokens._toModeTransitions(r.modeTransitions));
        };
        TextModelWithTokens.prototype._beginBackgroundTokenization = function () {
            var _this = this;
            if (this._shouldAutoTokenize && this._revalidateTokensTimeout === -1) {
                this._revalidateTokensTimeout = setTimeout(function () {
                    _this._revalidateTokensTimeout = -1;
                    _this._revalidateTokensNow();
                }, 0);
            }
        };
        TextModelWithTokens.prototype._warmUpTokens = function () {
            // Warm up first 100 lines (if it takes less than 50ms)
            var maxLineNumber = Math.min(100, this.getLineCount());
            var toLineNumber = maxLineNumber;
            for (var lineNumber = 1; lineNumber <= maxLineNumber; lineNumber++) {
                var text = this._lines[lineNumber - 1].text;
                if (text.length >= 200) {
                    // This line is over 200 chars long, so warm up without it
                    toLineNumber = lineNumber - 1;
                    break;
                }
            }
            this._revalidateTokensNow(toLineNumber);
        };
        TextModelWithTokens.prototype._revalidateTokensNow = function (toLineNumber) {
            if (toLineNumber === void 0) { toLineNumber = this._invalidLineStartIndex + 1000000; }
            var t1 = timer.start(timer.Topic.EDITOR, 'backgroundTokenization');
            toLineNumber = Math.min(this._lines.length, toLineNumber);
            var MAX_ALLOWED_TIME = 20, fromLineNumber = this._invalidLineStartIndex + 1, tokenizedChars = 0, currentCharsToTokenize = 0, currentEstimatedTimeToTokenize = 0, sw = stopwatch_1.StopWatch.create(false), elapsedTime;
            // Tokenize at most 1000 lines. Estimate the tokenization speed per character and stop when:
            // - MAX_ALLOWED_TIME is reached
            // - tokenizing the next line would go above MAX_ALLOWED_TIME
            for (var lineNumber = fromLineNumber; lineNumber <= toLineNumber; lineNumber++) {
                elapsedTime = sw.elapsed();
                if (elapsedTime > MAX_ALLOWED_TIME) {
                    // Stop if MAX_ALLOWED_TIME is reached
                    toLineNumber = lineNumber - 1;
                    break;
                }
                // Compute how many characters will be tokenized for this line
                currentCharsToTokenize = this._lines[lineNumber - 1].text.length;
                if (tokenizedChars > 0) {
                    // If we have enough history, estimate how long tokenizing this line would take
                    currentEstimatedTimeToTokenize = (elapsedTime / tokenizedChars) * currentCharsToTokenize;
                    if (elapsedTime + currentEstimatedTimeToTokenize > MAX_ALLOWED_TIME) {
                        // Tokenizing this line will go above MAX_ALLOWED_TIME
                        toLineNumber = lineNumber - 1;
                        break;
                    }
                }
                this._updateTokensUntilLine(lineNumber, false);
                tokenizedChars += currentCharsToTokenize;
            }
            elapsedTime = sw.elapsed();
            if (fromLineNumber <= toLineNumber) {
                this.emitModelTokensChangedEvent(fromLineNumber, toLineNumber);
            }
            if (this._invalidLineStartIndex < this._lines.length) {
                this._beginBackgroundTokenization();
            }
            t1.stop();
        };
        TextModelWithTokens.prototype.getStateBeforeLine = function (lineNumber) {
            this._updateTokensUntilLine(lineNumber - 1, true);
            return this._lines[lineNumber - 1].getState();
        };
        TextModelWithTokens.prototype.getStateAfterLine = function (lineNumber) {
            this._updateTokensUntilLine(lineNumber, true);
            return lineNumber < this._lines.length ? this._lines[lineNumber].getState() : this._lastState;
        };
        TextModelWithTokens.prototype._getLineModeTransitions = function (lineNumber) {
            if (lineNumber < 1 || lineNumber > this.getLineCount()) {
                throw new Error('Illegal value ' + lineNumber + ' for `lineNumber`');
            }
            this._updateTokensUntilLine(lineNumber, true);
            return this._lines[lineNumber - 1].getModeTransitions(this._mode);
        };
        TextModelWithTokens.prototype._updateTokensUntilLine = function (lineNumber, emitEvents) {
            var linesLength = this._lines.length;
            var endLineIndex = lineNumber - 1;
            var stopLineTokenizationAfter = 1000000000; // 1 billion, if a line is so long, you have other trouble :).
            var fromLineNumber = this._invalidLineStartIndex + 1, toLineNumber = lineNumber;
            // Validate all states up to and including endLineIndex
            for (var lineIndex = this._invalidLineStartIndex; lineIndex <= endLineIndex; lineIndex++) {
                var endStateIndex = lineIndex + 1;
                var r = null;
                var text = this._lines[lineIndex].text;
                if (this._mode.tokenizationSupport) {
                    try {
                        // Tokenize only the first X characters
                        r = this._mode.tokenizationSupport.tokenize(this._lines[lineIndex].text, this._lines[lineIndex].getState(), 0, stopLineTokenizationAfter);
                    }
                    catch (e) {
                        e.friendlyMessage = TextModelWithTokens.MODE_TOKENIZATION_FAILED_MSG;
                        errors_1.onUnexpectedError(e);
                    }
                    if (r && r.retokenize) {
                        this._retokenizers.push(this._createRetokenizer(r.retokenize, lineIndex + 1));
                    }
                    if (r && r.tokens && r.tokens.length > 0) {
                        // Cannot have a stop offset before the last token
                        r.actualStopOffset = Math.max(r.actualStopOffset, r.tokens[r.tokens.length - 1].startIndex + 1);
                    }
                    if (r && r.actualStopOffset < text.length) {
                        // Treat the rest of the line (if above limit) as one default token
                        r.tokens.push({
                            startIndex: r.actualStopOffset,
                            type: ''
                        });
                        // Use as end state the starting state
                        r.endState = this._lines[lineIndex].getState();
                    }
                }
                if (!r) {
                    r = nullMode_1.nullTokenize(this._mode, text, this._lines[lineIndex].getState());
                }
                if (!r.modeTransitions) {
                    r.modeTransitions = [];
                }
                if (r.modeTransitions.length === 0) {
                    // Make sure there is at least the transition to the top-most mode
                    r.modeTransitions.push({
                        startIndex: 0,
                        mode: this._mode
                    });
                }
                this._updateLineTokens(lineIndex, this._tokensInflatorMap, this._mode, r);
                if (this._lines[lineIndex].isInvalid) {
                    this._lines[lineIndex].isInvalid = false;
                }
                if (endStateIndex < linesLength) {
                    if (this._lines[endStateIndex].getState() !== null && r.endState.equals(this._lines[endStateIndex].getState())) {
                        // The end state of this line remains the same
                        var nextInvalidLineIndex = lineIndex + 1;
                        while (nextInvalidLineIndex < linesLength) {
                            if (this._lines[nextInvalidLineIndex].isInvalid) {
                                break;
                            }
                            if (nextInvalidLineIndex + 1 < linesLength) {
                                if (this._lines[nextInvalidLineIndex + 1].getState() === null) {
                                    break;
                                }
                            }
                            else {
                                if (this._lastState === null) {
                                    break;
                                }
                            }
                            nextInvalidLineIndex++;
                        }
                        this._invalidLineStartIndex = Math.max(this._invalidLineStartIndex, nextInvalidLineIndex);
                        lineIndex = nextInvalidLineIndex - 1; // -1 because the outer loop increments it
                    }
                    else {
                        this._lines[endStateIndex].setState(r.endState);
                    }
                }
                else {
                    this._lastState = r.endState;
                }
            }
            this._invalidLineStartIndex = Math.max(this._invalidLineStartIndex, endLineIndex + 1);
            if (emitEvents && fromLineNumber <= toLineNumber) {
                this.emitModelTokensChangedEvent(fromLineNumber, toLineNumber);
            }
        };
        TextModelWithTokens.prototype.emitModelTokensChangedEvent = function (fromLineNumber, toLineNumber) {
            var e = {
                fromLineNumber: fromLineNumber,
                toLineNumber: toLineNumber
            };
            if (!this._isDisposing) {
                this.emit(editorCommon.EventType.ModelTokensChanged, e);
            }
        };
        TextModelWithTokens.prototype._emitModelModeChangedEvent = function (e) {
            if (!this._isDisposing) {
                this.emit(editorCommon.EventType.ModelModeChanged, e);
            }
        };
        TextModelWithTokens.prototype._emitModelModeSupportChangedEvent = function (e) {
            if (!this._isDisposing) {
                this.emit(editorCommon.EventType.ModelModeSupportChanged, e);
            }
        };
        // Having tokens allows implementing additional helper methods
        TextModelWithTokens.prototype._lineIsTokenized = function (lineNumber) {
            return this._invalidLineStartIndex > lineNumber - 1;
        };
        TextModelWithTokens.prototype._getWordDefinition = function () {
            return textModelWithTokensHelpers_1.WordHelper.massageWordDefinitionOf(this._mode);
        };
        TextModelWithTokens.prototype.getWordAtPosition = function (position) {
            return textModelWithTokensHelpers_1.WordHelper.getWordAtPosition(this, this.validatePosition(position));
        };
        TextModelWithTokens.prototype.getWordUntilPosition = function (position) {
            var wordAtPosition = this.getWordAtPosition(position);
            if (!wordAtPosition) {
                return {
                    word: '',
                    startColumn: position.column,
                    endColumn: position.column
                };
            }
            return {
                word: wordAtPosition.word.substr(0, position.column - wordAtPosition.startColumn),
                startColumn: wordAtPosition.startColumn,
                endColumn: position.column
            };
        };
        TextModelWithTokens.prototype.tokenIterator = function (position, callback) {
            var iter = new tokenIterator_1.TokenIterator(this, this.validatePosition(position));
            var result = callback(iter);
            iter._invalidate();
            return result;
        };
        TextModelWithTokens.prototype.findMatchingBracketUp = function (bracket, _position) {
            var position = this.validatePosition(_position);
            var modeTransitions = this._lines[position.lineNumber - 1].getModeTransitions(this._mode);
            var currentModeIndex = modeTransition_1.ModeTransition.findIndexInSegmentsArray(modeTransitions, position.column - 1);
            var currentMode = modeTransitions[currentModeIndex];
            var currentModeBrackets = languageConfigurationRegistry_1.LanguageConfigurationRegistry.getBracketsSupport(currentMode.modeId);
            if (!currentModeBrackets) {
                return null;
            }
            var data = currentModeBrackets.textIsBracket[bracket];
            if (!data) {
                return null;
            }
            return this._findMatchingBracketUp(data, position);
        };
        TextModelWithTokens.prototype.matchBracket = function (position) {
            return this._matchBracket(this.validatePosition(position));
        };
        TextModelWithTokens.prototype._matchBracket = function (position) {
            var lineNumber = position.lineNumber;
            var lineText = this._lines[lineNumber - 1].text;
            var lineTokens = this._lines[lineNumber - 1].getTokens();
            var currentTokenIndex = lineTokens.findIndexOfOffset(position.column - 1);
            var currentTokenStart = lineTokens.getTokenStartIndex(currentTokenIndex);
            var modeTransitions = this._lines[lineNumber - 1].getModeTransitions(this._mode);
            var currentModeIndex = modeTransition_1.ModeTransition.findIndexInSegmentsArray(modeTransitions, position.column - 1);
            var currentMode = modeTransitions[currentModeIndex];
            var currentModeBrackets = languageConfigurationRegistry_1.LanguageConfigurationRegistry.getBracketsSupport(currentMode.modeId);
            // If position is in between two tokens, try first looking in the previous token
            if (currentTokenIndex > 0 && currentTokenStart === position.column - 1) {
                var prevTokenIndex = currentTokenIndex - 1;
                var prevTokenType = lineTokens.getTokenType(prevTokenIndex);
                // check that previous token is not to be ignored
                if (!supports_1.ignoreBracketsInToken(prevTokenType)) {
                    var prevTokenStart = lineTokens.getTokenStartIndex(prevTokenIndex);
                    var prevMode = currentMode;
                    var prevModeBrackets = currentModeBrackets;
                    // check if previous token is in a different mode
                    if (currentModeIndex > 0 && currentMode.startIndex === position.column - 1) {
                        prevMode = modeTransitions[currentModeIndex - 1];
                        prevModeBrackets = languageConfigurationRegistry_1.LanguageConfigurationRegistry.getBracketsSupport(prevMode.modeId);
                    }
                    if (prevModeBrackets) {
                        // limit search in case previous token is very large, there's no need to go beyond `maxBracketLength`
                        prevTokenStart = Math.max(prevTokenStart, position.column - 1 - prevModeBrackets.maxBracketLength);
                        var foundBracket = richEditBrackets_1.BracketsUtils.findPrevBracketInToken(prevModeBrackets.reversedRegex, lineNumber, lineText, prevTokenStart, currentTokenStart);
                        // check that we didn't hit a bracket too far away from position
                        if (foundBracket && foundBracket.startColumn <= position.column && position.column <= foundBracket.endColumn) {
                            var foundBracketText = lineText.substring(foundBracket.startColumn - 1, foundBracket.endColumn - 1);
                            var r = this._matchFoundBracket(foundBracket, prevModeBrackets.textIsBracket[foundBracketText], prevModeBrackets.textIsOpenBracket[foundBracketText]);
                            // check that we can actually match this bracket
                            if (r) {
                                return r;
                            }
                        }
                    }
                }
            }
            // check that the token is not to be ignored
            if (!supports_1.ignoreBracketsInToken(lineTokens.getTokenType(currentTokenIndex))) {
                if (currentModeBrackets) {
                    // limit search to not go before `maxBracketLength`
                    currentTokenStart = Math.max(currentTokenStart, position.column - 1 - currentModeBrackets.maxBracketLength);
                    // limit search to not go after `maxBracketLength`
                    var currentTokenEnd = lineTokens.getTokenEndIndex(currentTokenIndex, lineText.length);
                    currentTokenEnd = Math.min(currentTokenEnd, position.column - 1 + currentModeBrackets.maxBracketLength);
                    // it might still be the case that [currentTokenStart -> currentTokenEnd] contains multiple brackets
                    while (true) {
                        var foundBracket = richEditBrackets_1.BracketsUtils.findNextBracketInText(currentModeBrackets.forwardRegex, lineNumber, lineText.substring(currentTokenStart, currentTokenEnd), currentTokenStart);
                        if (!foundBracket) {
                            // there are no brackets in this text
                            break;
                        }
                        // check that we didn't hit a bracket too far away from position
                        if (foundBracket.startColumn <= position.column && position.column <= foundBracket.endColumn) {
                            var foundBracketText = lineText.substring(foundBracket.startColumn - 1, foundBracket.endColumn - 1);
                            var r = this._matchFoundBracket(foundBracket, currentModeBrackets.textIsBracket[foundBracketText], currentModeBrackets.textIsOpenBracket[foundBracketText]);
                            // check that we can actually match this bracket
                            if (r) {
                                return r;
                            }
                        }
                        currentTokenStart = foundBracket.endColumn - 1;
                    }
                }
            }
            return null;
        };
        TextModelWithTokens.prototype._matchFoundBracket = function (foundBracket, data, isOpen) {
            if (isOpen) {
                var matched = this._findMatchingBracketDown(data, foundBracket.getEndPosition());
                if (matched) {
                    return [foundBracket, matched];
                }
            }
            else {
                var matched = this._findMatchingBracketUp(data, foundBracket.getStartPosition());
                if (matched) {
                    return [foundBracket, matched];
                }
            }
            return null;
        };
        TextModelWithTokens.prototype._findMatchingBracketUp = function (bracket, position) {
            // console.log('_findMatchingBracketUp: ', 'bracket: ', JSON.stringify(bracket), 'startPosition: ', String(position));
            var modeId = bracket.modeId;
            var reversedBracketRegex = bracket.reversedRegex;
            var count = -1;
            for (var lineNumber = position.lineNumber; lineNumber >= 1; lineNumber--) {
                var lineTokens = this._lines[lineNumber - 1].getTokens();
                var lineText = this._lines[lineNumber - 1].text;
                var modeTransitions = this._lines[lineNumber - 1].getModeTransitions(this._mode);
                var currentModeIndex = modeTransitions.length - 1;
                var currentModeStart = modeTransitions[currentModeIndex].startIndex;
                var currentModeId = modeTransitions[currentModeIndex].modeId;
                var tokensLength = lineTokens.getTokenCount() - 1;
                var currentTokenEnd = lineText.length;
                if (lineNumber === position.lineNumber) {
                    tokensLength = lineTokens.findIndexOfOffset(position.column - 1);
                    currentTokenEnd = position.column - 1;
                    currentModeIndex = modeTransition_1.ModeTransition.findIndexInSegmentsArray(modeTransitions, position.column - 1);
                    currentModeStart = modeTransitions[currentModeIndex].startIndex;
                    currentModeId = modeTransitions[currentModeIndex].modeId;
                }
                for (var tokenIndex = tokensLength; tokenIndex >= 0; tokenIndex--) {
                    var currentTokenType = lineTokens.getTokenType(tokenIndex);
                    var currentTokenStart = lineTokens.getTokenStartIndex(tokenIndex);
                    if (currentTokenStart < currentModeStart) {
                        currentModeIndex--;
                        currentModeStart = modeTransitions[currentModeIndex].startIndex;
                        currentModeId = modeTransitions[currentModeIndex].modeId;
                    }
                    if (currentModeId === modeId && !supports_1.ignoreBracketsInToken(currentTokenType)) {
                        while (true) {
                            var r = richEditBrackets_1.BracketsUtils.findPrevBracketInToken(reversedBracketRegex, lineNumber, lineText, currentTokenStart, currentTokenEnd);
                            if (!r) {
                                break;
                            }
                            var hitText = lineText.substring(r.startColumn - 1, r.endColumn - 1);
                            if (hitText === bracket.open) {
                                count++;
                            }
                            else if (hitText === bracket.close) {
                                count--;
                            }
                            if (count === 0) {
                                return r;
                            }
                            currentTokenEnd = r.startColumn - 1;
                        }
                    }
                    currentTokenEnd = currentTokenStart;
                }
            }
            return null;
        };
        TextModelWithTokens.prototype._findMatchingBracketDown = function (bracket, position) {
            // console.log('_findMatchingBracketDown: ', 'bracket: ', JSON.stringify(bracket), 'startPosition: ', String(position));
            var modeId = bracket.modeId;
            var bracketRegex = bracket.forwardRegex;
            var count = 1;
            for (var lineNumber = position.lineNumber, lineCount = this.getLineCount(); lineNumber <= lineCount; lineNumber++) {
                var lineTokens = this._lines[lineNumber - 1].getTokens();
                var lineText = this._lines[lineNumber - 1].text;
                var modeTransitions = this._lines[lineNumber - 1].getModeTransitions(this._mode);
                var currentModeIndex = 0;
                var nextModeStart = (currentModeIndex + 1 < modeTransitions.length ? modeTransitions[currentModeIndex + 1].startIndex : lineText.length + 1);
                var currentModeId = modeTransitions[currentModeIndex].modeId;
                var startTokenIndex = 0;
                var currentTokenStart = lineTokens.getTokenStartIndex(startTokenIndex);
                if (lineNumber === position.lineNumber) {
                    startTokenIndex = lineTokens.findIndexOfOffset(position.column - 1);
                    currentTokenStart = Math.max(currentTokenStart, position.column - 1);
                    currentModeIndex = modeTransition_1.ModeTransition.findIndexInSegmentsArray(modeTransitions, position.column - 1);
                    nextModeStart = (currentModeIndex + 1 < modeTransitions.length ? modeTransitions[currentModeIndex + 1].startIndex : lineText.length + 1);
                    currentModeId = modeTransitions[currentModeIndex].modeId;
                }
                for (var tokenIndex = startTokenIndex, tokensLength = lineTokens.getTokenCount(); tokenIndex < tokensLength; tokenIndex++) {
                    var currentTokenType = lineTokens.getTokenType(tokenIndex);
                    var currentTokenEnd = lineTokens.getTokenEndIndex(tokenIndex, lineText.length);
                    if (currentTokenStart >= nextModeStart) {
                        currentModeIndex++;
                        nextModeStart = (currentModeIndex + 1 < modeTransitions.length ? modeTransitions[currentModeIndex + 1].startIndex : lineText.length + 1);
                        currentModeId = modeTransitions[currentModeIndex].modeId;
                    }
                    if (currentModeId === modeId && !supports_1.ignoreBracketsInToken(currentTokenType)) {
                        while (true) {
                            var r = richEditBrackets_1.BracketsUtils.findNextBracketInToken(bracketRegex, lineNumber, lineText, currentTokenStart, currentTokenEnd);
                            if (!r) {
                                break;
                            }
                            var hitText = lineText.substring(r.startColumn - 1, r.endColumn - 1);
                            if (hitText === bracket.open) {
                                count++;
                            }
                            else if (hitText === bracket.close) {
                                count--;
                            }
                            if (count === 0) {
                                return r;
                            }
                            currentTokenStart = r.endColumn - 1;
                        }
                    }
                    currentTokenStart = currentTokenEnd;
                }
            }
            return null;
        };
        TextModelWithTokens.prototype.findPrevBracket = function (_position) {
            var position = this.validatePosition(_position);
            var reversedBracketRegex = /[\(\)\[\]\{\}]/; // TODO@Alex: use mode's brackets
            for (var lineNumber = position.lineNumber; lineNumber >= 1; lineNumber--) {
                var lineTokens = this._lines[lineNumber - 1].getTokens();
                var lineText = this._lines[lineNumber - 1].text;
                var tokensLength = lineTokens.getTokenCount() - 1;
                var currentTokenEnd = lineText.length;
                if (lineNumber === position.lineNumber) {
                    tokensLength = lineTokens.findIndexOfOffset(position.column - 1);
                    currentTokenEnd = position.column - 1;
                }
                for (var tokenIndex = tokensLength; tokenIndex >= 0; tokenIndex--) {
                    var currentTokenType = lineTokens.getTokenType(tokenIndex);
                    var currentTokenStart = lineTokens.getTokenStartIndex(tokenIndex);
                    if (!supports_1.ignoreBracketsInToken(currentTokenType)) {
                        var r = richEditBrackets_1.BracketsUtils.findPrevBracketInToken(reversedBracketRegex, lineNumber, lineText, currentTokenStart, currentTokenEnd);
                        if (r) {
                            return this._toFoundBracket(r);
                        }
                    }
                    currentTokenEnd = currentTokenStart;
                }
            }
            return null;
        };
        TextModelWithTokens.prototype.findNextBracket = function (_position) {
            var position = this.validatePosition(_position);
            var bracketRegex = /[\(\)\[\]\{\}]/; // TODO@Alex: use mode's brackets
            for (var lineNumber = position.lineNumber, lineCount = this.getLineCount(); lineNumber <= lineCount; lineNumber++) {
                var lineTokens = this._lines[lineNumber - 1].getTokens();
                var lineText = this._lines[lineNumber - 1].text;
                var startTokenIndex = 0;
                var currentTokenStart = lineTokens.getTokenStartIndex(startTokenIndex);
                if (lineNumber === position.lineNumber) {
                    startTokenIndex = lineTokens.findIndexOfOffset(position.column - 1);
                    currentTokenStart = Math.max(currentTokenStart, position.column - 1);
                }
                for (var tokenIndex = startTokenIndex, tokensLength = lineTokens.getTokenCount(); tokenIndex < tokensLength; tokenIndex++) {
                    var currentTokenType = lineTokens.getTokenType(tokenIndex);
                    var currentTokenEnd = lineTokens.getTokenEndIndex(tokenIndex, lineText.length);
                    if (!supports_1.ignoreBracketsInToken(currentTokenType)) {
                        var r = richEditBrackets_1.BracketsUtils.findNextBracketInToken(bracketRegex, lineNumber, lineText, currentTokenStart, currentTokenEnd);
                        if (r) {
                            return this._toFoundBracket(r);
                        }
                    }
                    currentTokenStart = currentTokenEnd;
                }
            }
            return null;
        };
        TextModelWithTokens.prototype._toFoundBracket = function (r) {
            if (!r) {
                return null;
            }
            var text = this.getValueInRange(r);
            // TODO@Alex: use mode's brackets
            switch (text) {
                case '(': return { range: r, open: '(', close: ')', isOpen: true };
                case ')': return { range: r, open: '(', close: ')', isOpen: false };
                case '[': return { range: r, open: '[', close: ']', isOpen: true };
                case ']': return { range: r, open: '[', close: ']', isOpen: false };
                case '{': return { range: r, open: '{', close: '}', isOpen: true };
                case '}': return { range: r, open: '{', close: '}', isOpen: false };
            }
            return null;
        };
        TextModelWithTokens.MODE_TOKENIZATION_FAILED_MSG = nls.localize(0, null);
        return TextModelWithTokens;
    }(textModel_1.TextModel));
    exports.TextModelWithTokens = TextModelWithTokens;
});






define(__m[72], __M([1,0,19,57,45,71,29,22]), function (require, exports, editorCommon, modelLine_1, textModel_1, textModelWithTokens_1, range_1, position_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var AbstractMirrorModel = (function (_super) {
        __extends(AbstractMirrorModel, _super);
        function AbstractMirrorModel(allowedEventTypes, versionId, value, mode, associatedResource) {
            _super.call(this, allowedEventTypes.concat([editorCommon.EventType.ModelDispose]), value, false, mode);
            this._setVersionId(versionId);
            this._associatedResource = associatedResource;
        }
        AbstractMirrorModel.prototype.getModeId = function () {
            return this.getMode().getId();
        };
        AbstractMirrorModel.prototype._constructLines = function (rawText) {
            _super.prototype._constructLines.call(this, rawText);
            // Force EOL to be \n
            this._EOL = '\n';
        };
        AbstractMirrorModel.prototype.destroy = function () {
            this.dispose();
        };
        AbstractMirrorModel.prototype.dispose = function () {
            this.emit(editorCommon.EventType.ModelDispose);
            _super.prototype.dispose.call(this);
        };
        Object.defineProperty(AbstractMirrorModel.prototype, "uri", {
            get: function () {
                return this._associatedResource;
            },
            enumerable: true,
            configurable: true
        });
        AbstractMirrorModel.prototype.getRangeFromOffsetAndLength = function (offset, length) {
            var startPosition = this.getPositionAt(offset);
            var endPosition = this.getPositionAt(offset + length);
            return new range_1.Range(startPosition.lineNumber, startPosition.column, endPosition.lineNumber, endPosition.column);
        };
        AbstractMirrorModel.prototype.getOffsetAndLengthFromRange = function (range) {
            var startOffset = this.getOffsetAt(new position_1.Position(range.startLineNumber, range.startColumn));
            var endOffset = this.getOffsetAt(new position_1.Position(range.endLineNumber, range.endColumn));
            return {
                offset: startOffset,
                length: endOffset - startOffset
            };
        };
        AbstractMirrorModel.prototype.getPositionFromOffset = function (offset) {
            return this.getPositionAt(offset);
        };
        AbstractMirrorModel.prototype.getOffsetFromPosition = function (position) {
            return this.getOffsetAt(position);
        };
        AbstractMirrorModel.prototype.getLineStart = function (lineNumber) {
            if (lineNumber < 1) {
                lineNumber = 1;
            }
            if (lineNumber > this.getLineCount()) {
                lineNumber = this.getLineCount();
            }
            return this.getOffsetAt(new position_1.Position(lineNumber, 1));
        };
        AbstractMirrorModel.prototype.getAllWordsWithRange = function () {
            if (this._lines.length > 10000) {
                // This is a very heavy method, unavailable for very heavy models
                return [];
            }
            var result = [], i;
            var toTextRange = function (info) {
                var s = line.text.substring(info.start, info.end);
                var r = { startLineNumber: i + 1, startColumn: info.start + 1, endLineNumber: i + 1, endColumn: info.end + 1 };
                result.push({ text: s, range: r });
            };
            for (i = 0; i < this._lines.length; i++) {
                var line = this._lines[i];
                this.wordenize(line.text).forEach(toTextRange);
            }
            return result;
        };
        AbstractMirrorModel.prototype.getAllWords = function () {
            var _this = this;
            var result = [];
            this._lines.forEach(function (line) {
                _this.wordenize(line.text).forEach(function (info) {
                    result.push(line.text.substring(info.start, info.end));
                });
            });
            return result;
        };
        AbstractMirrorModel.prototype.getAllUniqueWords = function (skipWordOnce) {
            var foundSkipWord = false;
            var uniqueWords = {};
            return this.getAllWords().filter(function (word) {
                if (skipWordOnce && !foundSkipWord && skipWordOnce === word) {
                    foundSkipWord = true;
                    return false;
                }
                else if (uniqueWords[word]) {
                    return false;
                }
                else {
                    uniqueWords[word] = true;
                    return true;
                }
            });
        };
        //	// TODO@Joh, TODO@Alex - remove these and make sure the super-things work
        AbstractMirrorModel.prototype.wordenize = function (content) {
            var result = [];
            var match;
            var wordsRegexp = this._getWordDefinition();
            while (match = wordsRegexp.exec(content)) {
                result.push({ start: match.index, end: match.index + match[0].length });
            }
            return result;
        };
        return AbstractMirrorModel;
    }(textModelWithTokens_1.TextModelWithTokens));
    exports.AbstractMirrorModel = AbstractMirrorModel;
    function createTestMirrorModelFromString(value, mode, associatedResource) {
        if (mode === void 0) { mode = null; }
        return new MirrorModel(0, textModel_1.TextModel.toRawText(value, textModel_1.TextModel.DEFAULT_CREATION_OPTIONS), mode, associatedResource);
    }
    exports.createTestMirrorModelFromString = createTestMirrorModelFromString;
    var MirrorModel = (function (_super) {
        __extends(MirrorModel, _super);
        function MirrorModel(versionId, value, mode, associatedResource) {
            _super.call(this, ['changed'], versionId, value, mode, associatedResource);
        }
        MirrorModel.prototype.onEvents = function (events) {
            var changed = false;
            for (var i = 0, len = events.contentChanged.length; i < len; i++) {
                var contentChangedEvent = events.contentChanged[i];
                this._setVersionId(contentChangedEvent.versionId);
                switch (contentChangedEvent.changeType) {
                    case editorCommon.EventType.ModelRawContentChangedFlush:
                        this._onLinesFlushed(contentChangedEvent);
                        changed = true;
                        break;
                    case editorCommon.EventType.ModelRawContentChangedLinesDeleted:
                        this._onLinesDeleted(contentChangedEvent);
                        changed = true;
                        break;
                    case editorCommon.EventType.ModelRawContentChangedLinesInserted:
                        this._onLinesInserted(contentChangedEvent);
                        changed = true;
                        break;
                    case editorCommon.EventType.ModelRawContentChangedLineChanged:
                        this._onLineChanged(contentChangedEvent);
                        changed = true;
                        break;
                }
            }
            if (changed) {
                this.emit('changed', {});
            }
        };
        MirrorModel.prototype._onLinesFlushed = function (e) {
            // Flush my lines
            this._constructLines(e.detail);
            this._resetTokenizationState();
        };
        MirrorModel.prototype._onLineChanged = function (e) {
            this._lines[e.lineNumber - 1].applyEdits({}, [{
                    startColumn: 1,
                    endColumn: Number.MAX_VALUE,
                    text: e.detail,
                    forceMoveMarkers: false
                }]);
            if (this._lineStarts) {
                // update prefix sum
                this._lineStarts.changeValue(e.lineNumber - 1, this._lines[e.lineNumber - 1].text.length + this._EOL.length);
            }
            this._invalidateLine(e.lineNumber - 1);
        };
        MirrorModel.prototype._onLinesDeleted = function (e) {
            var fromLineIndex = e.fromLineNumber - 1, toLineIndex = e.toLineNumber - 1;
            // Save first line's state
            var firstLineState = this._lines[fromLineIndex].getState();
            this._lines.splice(fromLineIndex, toLineIndex - fromLineIndex + 1);
            if (this._lineStarts) {
                // update prefix sum
                this._lineStarts.removeValues(fromLineIndex, toLineIndex - fromLineIndex + 1);
            }
            if (fromLineIndex < this._lines.length) {
                // This check is always true in real world, but the tests forced this
                // Restore first line's state
                this._lines[fromLineIndex].setState(firstLineState);
                // Invalidate line
                this._invalidateLine(fromLineIndex);
            }
        };
        MirrorModel.prototype._onLinesInserted = function (e) {
            var lineIndex, i, splitLines = e.detail.split('\n');
            var newLengths = [];
            for (lineIndex = e.fromLineNumber - 1, i = 0; lineIndex < e.toLineNumber; lineIndex++, i++) {
                this._lines.splice(lineIndex, 0, new modelLine_1.ModelLine(0, splitLines[i]));
                newLengths.push(splitLines[i].length + this._EOL.length);
            }
            if (this._lineStarts) {
                // update prefix sum
                this._lineStarts.insertValues(e.fromLineNumber - 1, newLengths);
            }
            if (e.fromLineNumber >= 2) {
                // This check is always true in real world, but the tests forced this
                this._invalidateLine(e.fromLineNumber - 2);
            }
        };
        return MirrorModel;
    }(AbstractMirrorModel));
    exports.MirrorModel = MirrorModel;
});

define(__m[73], __M([7,6]), function(nls, data) { return nls.create("vs/editor/common/modes/modesRegistry", data); });
define(__m[74], __M([7,6]), function(nls, data) { return nls.create("vs/editor/common/modes/supports/suggestSupport", data); });
define(__m[75], __M([7,6]), function(nls, data) { return nls.create("vs/editor/common/services/modeServiceImpl", data); });
define(__m[76], __M([7,6]), function(nls, data) { return nls.create("vs/platform/configuration/common/configurationRegistry", data); });
define(__m[77], __M([7,6]), function(nls, data) { return nls.create("vs/platform/extensions/common/abstractExtensionService", data); });
define(__m[78], __M([7,6]), function(nls, data) { return nls.create("vs/platform/extensions/common/extensionsRegistry", data); });
define(__m[79], __M([7,6]), function(nls, data) { return nls.create("vs/platform/jsonschemas/common/jsonContributionRegistry", data); });





define(__m[33], __M([1,0,2]), function (require, exports, errors_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var AbstractDescriptor = (function () {
        function AbstractDescriptor(_staticArguments) {
            this._staticArguments = _staticArguments;
            // empty
        }
        AbstractDescriptor.prototype.appendStaticArguments = function (more) {
            this._staticArguments.push.apply(this._staticArguments, more);
        };
        AbstractDescriptor.prototype.staticArguments = function (nth) {
            if (isNaN(nth)) {
                return this._staticArguments.slice(0);
            }
            else {
                return this._staticArguments[nth];
            }
        };
        AbstractDescriptor.prototype._validate = function (type) {
            if (!type) {
                throw errors_1.illegalArgument('can not be falsy');
            }
        };
        return AbstractDescriptor;
    }());
    exports.AbstractDescriptor = AbstractDescriptor;
    var SyncDescriptor = (function (_super) {
        __extends(SyncDescriptor, _super);
        function SyncDescriptor(_ctor) {
            var staticArguments = [];
            for (var _i = 1; _i < arguments.length; _i++) {
                staticArguments[_i - 1] = arguments[_i];
            }
            _super.call(this, staticArguments);
            this._ctor = _ctor;
        }
        Object.defineProperty(SyncDescriptor.prototype, "ctor", {
            get: function () {
                return this._ctor;
            },
            enumerable: true,
            configurable: true
        });
        SyncDescriptor.prototype.bind = function () {
            var moreStaticArguments = [];
            for (var _i = 0; _i < arguments.length; _i++) {
                moreStaticArguments[_i - 0] = arguments[_i];
            }
            var allArgs = [];
            allArgs = allArgs.concat(this.staticArguments());
            allArgs = allArgs.concat(moreStaticArguments);
            return new (SyncDescriptor.bind.apply(SyncDescriptor, [void 0].concat([this._ctor], allArgs)))();
        };
        return SyncDescriptor;
    }(AbstractDescriptor));
    exports.SyncDescriptor = SyncDescriptor;
    exports.createSyncDescriptor = function (ctor) {
        var staticArguments = [];
        for (var _i = 1; _i < arguments.length; _i++) {
            staticArguments[_i - 1] = arguments[_i];
        }
        return new (SyncDescriptor.bind.apply(SyncDescriptor, [void 0].concat([ctor], staticArguments)))();
    };
    var AsyncDescriptor = (function (_super) {
        __extends(AsyncDescriptor, _super);
        function AsyncDescriptor(_moduleName, _ctorName) {
            var staticArguments = [];
            for (var _i = 2; _i < arguments.length; _i++) {
                staticArguments[_i - 2] = arguments[_i];
            }
            _super.call(this, staticArguments);
            this._moduleName = _moduleName;
            this._ctorName = _ctorName;
            if (typeof _moduleName !== 'string') {
                throw new Error('Invalid AsyncDescriptor arguments, expected `moduleName` to be a string!');
            }
        }
        AsyncDescriptor.create = function (moduleName, ctorName) {
            return new AsyncDescriptor(moduleName, ctorName);
        };
        Object.defineProperty(AsyncDescriptor.prototype, "moduleName", {
            get: function () {
                return this._moduleName;
            },
            enumerable: true,
            configurable: true
        });
        Object.defineProperty(AsyncDescriptor.prototype, "ctorName", {
            get: function () {
                return this._ctorName;
            },
            enumerable: true,
            configurable: true
        });
        AsyncDescriptor.prototype.bind = function () {
            var moreStaticArguments = [];
            for (var _i = 0; _i < arguments.length; _i++) {
                moreStaticArguments[_i - 0] = arguments[_i];
            }
            var allArgs = [];
            allArgs = allArgs.concat(this.staticArguments());
            allArgs = allArgs.concat(moreStaticArguments);
            return new (AsyncDescriptor.bind.apply(AsyncDescriptor, [void 0].concat([this.moduleName, this.ctorName], allArgs)))();
        };
        return AsyncDescriptor;
    }(AbstractDescriptor));
    exports.AsyncDescriptor = AsyncDescriptor;
    var _createAsyncDescriptor = function (moduleName, ctorName) {
        var staticArguments = [];
        for (var _i = 2; _i < arguments.length; _i++) {
            staticArguments[_i - 2] = arguments[_i];
        }
        return new (AsyncDescriptor.bind.apply(AsyncDescriptor, [void 0].concat([moduleName, ctorName], staticArguments)))();
    };
    exports.createAsyncDescriptor0 = _createAsyncDescriptor;
    exports.createAsyncDescriptor1 = _createAsyncDescriptor;
    exports.createAsyncDescriptor2 = _createAsyncDescriptor;
    exports.createAsyncDescriptor3 = _createAsyncDescriptor;
    exports.createAsyncDescriptor4 = _createAsyncDescriptor;
    exports.createAsyncDescriptor5 = _createAsyncDescriptor;
    exports.createAsyncDescriptor6 = _createAsyncDescriptor;
    exports.createAsyncDescriptor7 = _createAsyncDescriptor;
});

define(__m[4], __M([1,0]), function (require, exports) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    // ------ internal util
    var _util;
    (function (_util) {
        _util.DI_TARGET = '$di$target';
        _util.DI_DEPENDENCIES = '$di$dependencies';
        function getServiceDependencies(ctor) {
            return ctor[_util.DI_DEPENDENCIES] || [];
        }
        _util.getServiceDependencies = getServiceDependencies;
    })(_util = exports._util || (exports._util = {}));
    exports.IInstantiationService = createDecorator('instantiationService');
    function storeServiceDependency(id, target, index, optional) {
        if (target[_util.DI_TARGET] === target) {
            target[_util.DI_DEPENDENCIES].push({ id: id, index: index, optional: optional });
        }
        else {
            target[_util.DI_DEPENDENCIES] = [{ id: id, index: index, optional: optional }];
            target[_util.DI_TARGET] = target;
        }
    }
    /**
     * A *only* valid way to create a {{ServiceIdentifier}}.
     */
    function createDecorator(serviceId) {
        var id = function (target, key, index) {
            if (arguments.length !== 3) {
                throw new Error('@IServiceName-decorator can only be used to decorate a parameter');
            }
            storeServiceDependency(id, target, index, false);
        };
        id.toString = function () { return serviceId; };
        return id;
    }
    exports.createDecorator = createDecorator;
    /**
     * Mark a service dependency as optional.
     */
    function optional(serviceIdentifier) {
        return function (target, key, index) {
            if (arguments.length !== 3) {
                throw new Error('@optional-decorator can only be used to decorate a parameter');
            }
            storeServiceDependency(serviceIdentifier, target, index, true);
        };
    }
    exports.optional = optional;
});

define(__m[35], __M([1,0,4]), function (require, exports, instantiation_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    exports.ICompatWorkerService = instantiation_1.createDecorator('compatWorkerService');
    function findMember(proto, target) {
        for (var i in proto) {
            if (proto[i] === target) {
                return i;
            }
        }
        throw new Error('Member not found in prototype');
    }
    function CompatWorkerAttr(type, target) {
        var methodName = findMember(type.prototype, target);
        type.prototype[methodName] = function () {
            var param = [];
            for (var _i = 0; _i < arguments.length; _i++) {
                param[_i - 0] = arguments[_i];
            }
            var obj = this;
            return obj.compatWorkerService.CompatWorker(obj, methodName, target, param);
        };
    }
    exports.CompatWorkerAttr = CompatWorkerAttr;
});

define(__m[83], __M([1,0,4]), function (require, exports, instantiation_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    exports.ID_EDITOR_WORKER_SERVICE = 'editorWorkerService';
    exports.IEditorWorkerService = instantiation_1.createDecorator(exports.ID_EDITOR_WORKER_SERVICE);
});

define(__m[23], __M([1,0,4]), function (require, exports, instantiation_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    exports.IModeService = instantiation_1.createDecorator('modeService');
});

define(__m[85], __M([1,0,4]), function (require, exports, instantiation_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    exports.IModelService = instantiation_1.createDecorator('modelService');
});

define(__m[37], __M([1,0,4]), function (require, exports, instantiation_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    // Resource Service
    exports.ResourceEvents = {
        ADDED: 'resource.added',
        REMOVED: 'resource.removed',
        CHANGED: 'resource.changed'
    };
    exports.IResourceService = instantiation_1.createDecorator('resourceService');
});






define(__m[87], __M([1,0,15,16,37]), function (require, exports, eventEmitter_1, lifecycle_1, resourceService_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var ResourceService = (function (_super) {
        __extends(ResourceService, _super);
        function ResourceService() {
            _super.call(this);
            this.serviceId = resourceService_1.IResourceService;
            this.data = {};
            this.unbinds = {};
        }
        ResourceService.prototype.addListener2_ = function (eventType, listener) {
            return _super.prototype.addListener2.call(this, eventType, listener);
        };
        ResourceService.prototype._anonymousModelId = function (input) {
            var r = '';
            for (var i = 0; i < input.length; i++) {
                var ch = input[i];
                if (ch >= '0' && ch <= '9') {
                    r += '0';
                    continue;
                }
                if (ch >= 'a' && ch <= 'z') {
                    r += 'a';
                    continue;
                }
                if (ch >= 'A' && ch <= 'Z') {
                    r += 'A';
                    continue;
                }
                r += ch;
            }
            return r;
        };
        ResourceService.prototype.insert = function (url, element) {
            var _this = this;
            // console.log('INSERT: ' + url.toString());
            if (this.contains(url)) {
                // There already exists a model with this id => this is a programmer error
                throw new Error('ResourceService: Cannot add model ' + this._anonymousModelId(url.toString()) + ' because it already exists!');
            }
            // add resource
            var key = url.toString();
            this.data[key] = element;
            this.unbinds[key] = [];
            this.unbinds[key].push(element.addBulkListener2(function (value) {
                _this.emit(resourceService_1.ResourceEvents.CHANGED, { url: url, originalEvents: value });
            }));
            // event
            this.emit(resourceService_1.ResourceEvents.ADDED, { url: url, addedElement: element });
        };
        ResourceService.prototype.get = function (url) {
            if (!this.data[url.toString()]) {
                return null;
            }
            return this.data[url.toString()];
        };
        ResourceService.prototype.all = function () {
            var _this = this;
            return Object.keys(this.data).map(function (key) {
                return _this.data[key];
            });
        };
        ResourceService.prototype.contains = function (url) {
            return !!this.data[url.toString()];
        };
        ResourceService.prototype.remove = function (url) {
            // console.log('REMOVE: ' + url.toString());
            if (!this.contains(url)) {
                return;
            }
            var key = url.toString(), element = this.data[key];
            // stop listen
            this.unbinds[key] = lifecycle_1.dispose(this.unbinds[key]);
            // removal
            delete this.unbinds[key];
            delete this.data[key];
            // event
            this.emit(resourceService_1.ResourceEvents.REMOVED, { url: url, removedElement: element });
        };
        return ResourceService;
    }(eventEmitter_1.EventEmitter));
    exports.ResourceService = ResourceService;
});

/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Microsoft Corporation. All rights reserved.
 *  Licensed under the MIT License. See License.txt in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
define(__m[51], __M([1,0,4]), function (require, exports, instantiation_1) {
    "use strict";
    exports.IConfigurationService = instantiation_1.createDecorator('configurationService');
    function getConfigurationValue(config, settingPath, defaultValue) {
        function accessSetting(config, path) {
            var current = config;
            for (var i = 0; i < path.length; i++) {
                current = current[path[i]];
                if (!current) {
                    return undefined;
                }
            }
            return current;
        }
        var path = settingPath.split('.');
        var result = accessSetting(config, path);
        return typeof result === 'undefined'
            ? defaultValue
            : result;
    }
    exports.getConfigurationValue = getConfigurationValue;
});

define(__m[52], __M([1,0,4]), function (require, exports, instantiation_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    exports.IEventService = instantiation_1.createDecorator('eventService');
});






define(__m[90], __M([1,0,15,52]), function (require, exports, eventEmitter_1, event_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    // --- implementation ------------------------------------------
    var EventService = (function (_super) {
        __extends(EventService, _super);
        function EventService() {
            _super.call(this);
            this.serviceId = event_1.IEventService;
        }
        return EventService;
    }(eventEmitter_1.EventEmitter));
    exports.EventService = EventService;
});

define(__m[38], __M([1,0,4]), function (require, exports, instantiation_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    exports.IExtensionService = instantiation_1.createDecorator('extensionService');
});

define(__m[54], __M([1,0,58]), function (require, exports, arrays_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var ServiceCollection = (function () {
        function ServiceCollection() {
            var entries = [];
            for (var _i = 0; _i < arguments.length; _i++) {
                entries[_i - 0] = arguments[_i];
            }
            this._entries = [];
            for (var _c = 0, entries_1 = entries; _c < entries_1.length; _c++) {
                var entry = entries_1[_c];
                this.set(entry[0], entry[1]);
            }
        }
        ServiceCollection.prototype.set = function (id, instanceOrDescriptor) {
            var entry = [id, instanceOrDescriptor];
            var idx = arrays_1.binarySearch(this._entries, entry, ServiceCollection._entryCompare);
            if (idx < 0) {
                // new element
                this._entries.splice(~idx, 0, entry);
            }
            else {
                var old = this._entries[idx];
                this._entries[idx] = entry;
                return old[1];
            }
        };
        ServiceCollection.prototype.forEach = function (callback) {
            for (var _i = 0, _c = this._entries; _i < _c.length; _i++) {
                var entry = _c[_i];
                var id = entry[0], instanceOrDescriptor = entry[1];
                callback(id, instanceOrDescriptor);
            }
        };
        ServiceCollection.prototype.has = function (id) {
            return arrays_1.binarySearch(this._entries, ServiceCollection._searchEntry(id), ServiceCollection._entryCompare) >= 0;
        };
        ServiceCollection.prototype.get = function (id) {
            var idx = arrays_1.binarySearch(this._entries, ServiceCollection._searchEntry(id), ServiceCollection._entryCompare);
            if (idx >= 0) {
                return this._entries[idx][1];
            }
        };
        ServiceCollection._searchEntry = function (id) {
            ServiceCollection._dummy[0] = id;
            return ServiceCollection._dummy;
        };
        ServiceCollection._entryCompare = function (a, b) {
            var _a = a[0].toString();
            var _b = b[0].toString();
            if (_a < _b) {
                return -1;
            }
            else if (_a > _b) {
                return 1;
            }
            else {
                return 0;
            }
        };
        ServiceCollection._dummy = [undefined, undefined];
        return ServiceCollection;
    }());
    exports.ServiceCollection = ServiceCollection;
});

define(__m[93], __M([1,0,5,2,9,34,86,33,4,54]), function (require, exports, winjs_base_1, errors_1, types_1, assert, graph_1, descriptors_1, instantiation_1, serviceCollection_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var InstantiationService = (function () {
        function InstantiationService(services, strict) {
            if (services === void 0) { services = new serviceCollection_1.ServiceCollection(); }
            if (strict === void 0) { strict = false; }
            this._services = services;
            this._strict = strict;
            this._services.set(instantiation_1.IInstantiationService, this);
        }
        InstantiationService.prototype.createChild = function (services) {
            var _this = this;
            this._services.forEach(function (id, thing) {
                if (services.has(id)) {
                    return;
                }
                // If we copy descriptors we might end up with
                // multiple instances of the same service
                if (thing instanceof descriptors_1.SyncDescriptor) {
                    thing = _this._createAndCacheServiceInstance(id, thing);
                }
                services.set(id, thing);
            });
            return new InstantiationService(services, this._strict);
        };
        InstantiationService.prototype.invokeFunction = function (signature) {
            var _this = this;
            var args = [];
            for (var _i = 1; _i < arguments.length; _i++) {
                args[_i - 1] = arguments[_i];
            }
            var accessor;
            try {
                accessor = {
                    get: function (id, isOptional) {
                        var result = _this._getOrCreateServiceInstance(id);
                        if (!result && isOptional !== instantiation_1.optional) {
                            throw new Error("[invokeFunction] unkown service '" + id + "'");
                        }
                        return result;
                    }
                };
                return signature.apply(undefined, [accessor].concat(args));
            }
            finally {
                accessor.get = function () {
                    throw errors_1.illegalState('service accessor is only valid during the invocation of its target method');
                };
            }
        };
        InstantiationService.prototype.createInstance = function (param) {
            var rest = [];
            for (var _i = 1; _i < arguments.length; _i++) {
                rest[_i - 1] = arguments[_i];
            }
            if (param instanceof descriptors_1.AsyncDescriptor) {
                // async
                return this._createInstanceAsync(param, rest);
            }
            else if (param instanceof descriptors_1.SyncDescriptor) {
                // sync
                return this._createInstance(param, rest);
            }
            else {
                // sync, just ctor
                return this._createInstance(new descriptors_1.SyncDescriptor(param), rest);
            }
        };
        InstantiationService.prototype._createInstanceAsync = function (descriptor, args) {
            var _this = this;
            var canceledError;
            return new winjs_base_1.TPromise(function (c, e, p) {
                require([descriptor.moduleName], function (_module) {
                    if (canceledError) {
                        e(canceledError);
                    }
                    if (!_module) {
                        return e(errors_1.illegalArgument('module not found: ' + descriptor.moduleName));
                    }
                    var ctor;
                    if (!descriptor.ctorName) {
                        ctor = _module;
                    }
                    else {
                        ctor = _module[descriptor.ctorName];
                    }
                    if (typeof ctor !== 'function') {
                        return e(errors_1.illegalArgument('not a function: ' + descriptor.ctorName || descriptor.moduleName));
                    }
                    try {
                        args.unshift.apply(args, descriptor.staticArguments()); // instead of spread in ctor call
                        c(_this._createInstance(new descriptors_1.SyncDescriptor(ctor), args));
                    }
                    catch (error) {
                        return e(error);
                    }
                }, e);
            }, function () {
                canceledError = errors_1.canceled();
            });
        };
        InstantiationService.prototype._createInstance = function (desc, args) {
            var _this = this;
            // arguments given by createInstance-call and/or the descriptor
            var staticArgs = desc.staticArguments().concat(args);
            // arguments defined by service decorators
            var serviceDependencies = instantiation_1._util.getServiceDependencies(desc.ctor).sort(function (a, b) { return a.index - b.index; });
            var serviceArgs = serviceDependencies.map(function (dependency) {
                var service = _this._getOrCreateServiceInstance(dependency.id);
                if (!service && _this._strict && !dependency.optional) {
                    throw new Error("[createInstance] " + desc.ctor.name + " depends on UNKNOWN service " + dependency.id + ".");
                }
                return service;
            });
            var firstServiceArgPos = serviceDependencies.length > 0 ? serviceDependencies[0].index : staticArgs.length;
            // check for argument mismatches, adjust static args if needed
            if (staticArgs.length !== firstServiceArgPos) {
                console.warn("[createInstance] First service dependency of " + desc.ctor.name + " at position " + (firstServiceArgPos + 1) + " conflicts with " + staticArgs.length + " static arguments");
                var delta = firstServiceArgPos - staticArgs.length;
                if (delta > 0) {
                    staticArgs = staticArgs.concat(new Array(delta));
                }
                else {
                    staticArgs = staticArgs.slice(0, firstServiceArgPos);
                }
            }
            // // check for missing args
            // for (let i = 0; i < serviceArgs.length; i++) {
            // 	if (!serviceArgs[i]) {
            // 		console.warn(`${desc.ctor.name} MISSES service dependency ${serviceDependencies[i].id}`, new Error().stack);
            // 	}
            // }
            // now create the instance
            var argArray = [desc.ctor];
            argArray.push.apply(argArray, staticArgs);
            argArray.push.apply(argArray, serviceArgs);
            var instance = types_1.create.apply(null, argArray);
            desc._validate(instance);
            return instance;
        };
        InstantiationService.prototype._getOrCreateServiceInstance = function (id) {
            var thing = this._services.get(id);
            if (thing instanceof descriptors_1.SyncDescriptor) {
                return this._createAndCacheServiceInstance(id, thing);
            }
            else {
                return thing;
            }
        };
        InstantiationService.prototype._createAndCacheServiceInstance = function (id, desc) {
            assert.ok(this._services.get(id) instanceof descriptors_1.SyncDescriptor);
            var graph = new graph_1.Graph(function (data) { return data.id.toString(); });
            function throwCycleError() {
                var err = new Error('[createInstance] cyclic dependency between services');
                err.message = graph.toString();
                throw err;
            }
            var count = 0;
            var stack = [{ id: id, desc: desc }];
            while (stack.length) {
                var item = stack.pop();
                graph.lookupOrInsertNode(item);
                // TODO@joh use the graph to find a cycle
                // a weak heuristic for cycle checks
                if (count++ > 100) {
                    throwCycleError();
                }
                // check all dependencies for existence and if the need to be created first
                var dependencies = instantiation_1._util.getServiceDependencies(item.desc.ctor);
                for (var _i = 0, dependencies_1 = dependencies; _i < dependencies_1.length; _i++) {
                    var dependency = dependencies_1[_i];
                    var instanceOrDesc = this._services.get(dependency.id);
                    if (!instanceOrDesc) {
                        console.warn("[createInstance] " + id + " depends on " + dependency.id + " which is NOT registered.");
                    }
                    if (instanceOrDesc instanceof descriptors_1.SyncDescriptor) {
                        var d = { id: dependency.id, desc: instanceOrDesc };
                        graph.insertEdge(item, d);
                        stack.push(d);
                    }
                }
            }
            while (true) {
                var roots = graph.roots();
                // if there is no more roots but still
                // nodes in the graph we have a cycle
                if (roots.length === 0) {
                    if (graph.length !== 0) {
                        throwCycleError();
                    }
                    break;
                }
                for (var _a = 0, roots_1 = roots; _a < roots_1.length; _a++) {
                    var root = roots_1[_a];
                    // create instance and overwrite the service collections
                    var instance = this._createInstance(root.data.desc, []);
                    this._services.set(root.data.id, instance);
                    graph.removeNode(root.data);
                }
            }
            return this._services.get(id);
        };
        return InstantiationService;
    }());
    exports.InstantiationService = InstantiationService;
});

define(__m[11], __M([1,0,9,34]), function (require, exports, Types, Assert) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var RegistryImpl = (function () {
        function RegistryImpl() {
            this.data = {};
        }
        RegistryImpl.prototype.add = function (id, data) {
            Assert.ok(Types.isString(id));
            Assert.ok(Types.isObject(data));
            Assert.ok(!this.data.hasOwnProperty(id), 'There is already an extension with this id');
            this.data[id] = data;
        };
        RegistryImpl.prototype.knows = function (id) {
            return this.data.hasOwnProperty(id);
        };
        RegistryImpl.prototype.as = function (id) {
            return this.data[id] || null;
        };
        return RegistryImpl;
    }());
    exports.Registry = new RegistryImpl();
    /**
     * A base class for registries that leverage the instantiation service to create instances.
     */
    var BaseRegistry = (function () {
        function BaseRegistry() {
            this.toBeInstantiated = [];
            this.instances = [];
        }
        BaseRegistry.prototype.setInstantiationService = function (service) {
            this.instantiationService = service;
            while (this.toBeInstantiated.length > 0) {
                var entry = this.toBeInstantiated.shift();
                this.instantiate(entry);
            }
        };
        BaseRegistry.prototype.instantiate = function (ctor) {
            var instance = this.instantiationService.createInstance(ctor);
            this.instances.push(instance);
        };
        BaseRegistry.prototype._register = function (ctor) {
            if (this.instantiationService) {
                this.instantiate(ctor);
            }
            else {
                this.toBeInstantiated.push(ctor);
            }
        };
        BaseRegistry.prototype._getInstances = function () {
            return this.instances.slice(0);
        };
        BaseRegistry.prototype._setInstances = function (instances) {
            this.instances = instances;
        };
        return BaseRegistry;
    }());
    exports.BaseRegistry = BaseRegistry;
});

define(__m[40], __M([1,0,73,8,11]), function (require, exports, nls, event_1, platform_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    // Define extension point ids
    exports.Extensions = {
        ModesRegistry: 'editor.modesRegistry'
    };
    var EditorModesRegistry = (function () {
        function EditorModesRegistry() {
            this._onDidAddCompatModes = new event_1.Emitter();
            this.onDidAddCompatModes = this._onDidAddCompatModes.event;
            this._onDidAddLanguages = new event_1.Emitter();
            this.onDidAddLanguages = this._onDidAddLanguages.event;
            this._compatModes = [];
            this._languages = [];
        }
        // --- compat modes
        EditorModesRegistry.prototype.registerCompatModes = function (def) {
            this._compatModes = this._compatModes.concat(def);
            this._onDidAddCompatModes.fire(def);
        };
        EditorModesRegistry.prototype.registerCompatMode = function (def) {
            this._compatModes.push(def);
            this._onDidAddCompatModes.fire([def]);
        };
        EditorModesRegistry.prototype.getCompatModes = function () {
            return this._compatModes.slice(0);
        };
        // --- languages
        EditorModesRegistry.prototype.registerLanguage = function (def) {
            this._languages.push(def);
            this._onDidAddLanguages.fire([def]);
        };
        EditorModesRegistry.prototype.registerLanguages = function (def) {
            this._languages = this._languages.concat(def);
            this._onDidAddLanguages.fire(def);
        };
        EditorModesRegistry.prototype.getLanguages = function () {
            return this._languages.slice(0);
        };
        return EditorModesRegistry;
    }());
    exports.EditorModesRegistry = EditorModesRegistry;
    exports.ModesRegistry = new EditorModesRegistry();
    platform_1.Registry.add(exports.Extensions.ModesRegistry, exports.ModesRegistry);
    exports.ModesRegistry.registerLanguage({
        id: 'plaintext',
        extensions: ['.txt', '.gitignore'],
        aliases: [nls.localize(0, null), 'text'],
        mimetypes: ['text/plain']
    });
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
define(__m[96], __M([1,0,5,35,37,23,72,2,40]), function (require, exports, winjs_base_1, compatWorkerService_1, resourceService_1, modeService_1, mirrorModel_1, errors_1, modesRegistry_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var CompatWorkerServiceWorker = (function () {
        function CompatWorkerServiceWorker(resourceService, modeService, modesRegistryData) {
            this.resourceService = resourceService;
            this.modeService = modeService;
            this.serviceId = compatWorkerService_1.ICompatWorkerService;
            this.isInMainThread = false;
            modesRegistry_1.ModesRegistry.registerCompatModes(modesRegistryData.compatModes);
            modesRegistry_1.ModesRegistry.registerLanguages(modesRegistryData.languages);
            this._compatModes = Object.create(null);
        }
        CompatWorkerServiceWorker.prototype.registerCompatMode = function (compatMode) {
            this._compatModes[compatMode.getId()] = compatMode;
        };
        CompatWorkerServiceWorker.prototype.handleMainRequest = function (rpcId, methodName, args) {
            if (rpcId === '$') {
                switch (methodName) {
                    case 'acceptNewModel':
                        return this._acceptNewModel(args[0]);
                    case 'acceptDidDisposeModel':
                        return this._acceptDidDisposeModel(args[0]);
                    case 'acceptModelEvents':
                        return this._acceptModelEvents(args[0], args[1]);
                    case 'acceptCompatModes':
                        return this._acceptCompatModes(args[0]);
                    case 'acceptLanguages':
                        return this._acceptLanguages(args[0]);
                    case 'instantiateCompatMode':
                        return this._instantiateCompatMode(args[0]);
                }
            }
            var obj = this._compatModes[rpcId];
            return winjs_base_1.TPromise.as(obj[methodName].apply(obj, args));
        };
        CompatWorkerServiceWorker.prototype.CompatWorker = function (obj, methodName, target, param) {
            return target.apply(obj, param);
        };
        CompatWorkerServiceWorker.prototype._acceptNewModel = function (data) {
            var _this = this;
            // Create & insert the mirror model eagerly in the resource service
            var mirrorModel = new mirrorModel_1.MirrorModel(data.versionId, data.value, null, data.url);
            this.resourceService.insert(mirrorModel.uri, mirrorModel);
            // Block worker execution until the mode is instantiated
            return this.modeService.getOrCreateMode(data.modeId).then(function (mode) {
                // Changing mode should trigger a remove & an add, therefore:
                // (1) Remove from resource service
                _this.resourceService.remove(mirrorModel.uri);
                // (2) Change mode
                mirrorModel.setMode(mode);
                // (3) Insert again to resource service (it will have the new mode)
                _this.resourceService.insert(mirrorModel.uri, mirrorModel);
            });
        };
        CompatWorkerServiceWorker.prototype._acceptDidDisposeModel = function (uri) {
            var model = this.resourceService.get(uri);
            this.resourceService.remove(uri);
            model.dispose();
        };
        CompatWorkerServiceWorker.prototype._acceptModelEvents = function (uri, events) {
            var model = this.resourceService.get(uri);
            try {
                model.onEvents(events);
            }
            catch (err) {
                errors_1.onUnexpectedError(err);
            }
        };
        CompatWorkerServiceWorker.prototype._acceptCompatModes = function (modes) {
            modesRegistry_1.ModesRegistry.registerCompatModes(modes);
        };
        CompatWorkerServiceWorker.prototype._acceptLanguages = function (languages) {
            modesRegistry_1.ModesRegistry.registerLanguages(languages);
        };
        CompatWorkerServiceWorker.prototype._instantiateCompatMode = function (modeId) {
            return this.modeService.getOrCreateMode(modeId).then(function (_) { return void 0; });
        };
        CompatWorkerServiceWorker = __decorate([
            __param(0, resourceService_1.IResourceService),
            __param(1, modeService_1.IModeService)
        ], CompatWorkerServiceWorker);
        return CompatWorkerServiceWorker;
    }());
    exports.CompatWorkerServiceWorker = CompatWorkerServiceWorker;
});

define(__m[97], __M([1,0,2,8,46,3,40]), function (require, exports, errors_1, event_1, mime, strings, modesRegistry_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var hasOwnProperty = Object.prototype.hasOwnProperty;
    var LanguagesRegistry = (function () {
        function LanguagesRegistry(useModesRegistry) {
            var _this = this;
            if (useModesRegistry === void 0) { useModesRegistry = true; }
            this._onDidAddModes = new event_1.Emitter();
            this.onDidAddModes = this._onDidAddModes.event;
            this.knownModeIds = {};
            this.mime2LanguageId = {};
            this.name2LanguageId = {};
            this.id2Name = {};
            this.id2Extensions = {};
            this.compatModes = {};
            this.lowerName2Id = {};
            this.id2ConfigurationFiles = {};
            if (useModesRegistry) {
                this._registerCompatModes(modesRegistry_1.ModesRegistry.getCompatModes());
                modesRegistry_1.ModesRegistry.onDidAddCompatModes(function (m) { return _this._registerCompatModes(m); });
                this._registerLanguages(modesRegistry_1.ModesRegistry.getLanguages());
                modesRegistry_1.ModesRegistry.onDidAddLanguages(function (m) { return _this._registerLanguages(m); });
            }
        }
        LanguagesRegistry.prototype._registerCompatModes = function (defs) {
            var addedModes = [];
            for (var i = 0; i < defs.length; i++) {
                var def = defs[i];
                this._registerLanguage({
                    id: def.id,
                    extensions: def.extensions,
                    filenames: def.filenames,
                    firstLine: def.firstLine,
                    aliases: def.aliases,
                    mimetypes: def.mimetypes
                });
                this.compatModes[def.id] = {
                    moduleId: def.moduleId,
                    ctorName: def.ctorName,
                    deps: def.deps
                };
                addedModes.push(def.id);
            }
            this._onDidAddModes.fire(addedModes);
        };
        LanguagesRegistry.prototype._registerLanguages = function (desc) {
            var addedModes = [];
            for (var i = 0; i < desc.length; i++) {
                this._registerLanguage(desc[i]);
                addedModes.push(desc[i].id);
            }
            this._onDidAddModes.fire(addedModes);
        };
        LanguagesRegistry.prototype._setLanguageName = function (languageId, languageName, force) {
            var prevName = this.id2Name[languageId];
            if (prevName) {
                if (!force) {
                    return;
                }
                delete this.name2LanguageId[prevName];
            }
            this.name2LanguageId[languageName] = languageId;
            this.id2Name[languageId] = languageName;
        };
        LanguagesRegistry.prototype._registerLanguage = function (lang) {
            this.knownModeIds[lang.id] = true;
            var primaryMime = null;
            if (typeof lang.mimetypes !== 'undefined' && Array.isArray(lang.mimetypes)) {
                for (var i = 0; i < lang.mimetypes.length; i++) {
                    if (!primaryMime) {
                        primaryMime = lang.mimetypes[i];
                    }
                    this.mime2LanguageId[lang.mimetypes[i]] = lang.id;
                }
            }
            if (!primaryMime) {
                primaryMime = 'text/x-' + lang.id;
                this.mime2LanguageId[primaryMime] = lang.id;
            }
            if (Array.isArray(lang.extensions)) {
                this.id2Extensions[lang.id] = this.id2Extensions[lang.id] || [];
                for (var _i = 0, _a = lang.extensions; _i < _a.length; _i++) {
                    var extension = _a[_i];
                    mime.registerTextMime({ mime: primaryMime, extension: extension });
                    this.id2Extensions[lang.id].push(extension);
                }
            }
            if (Array.isArray(lang.filenames)) {
                for (var _b = 0, _c = lang.filenames; _b < _c.length; _b++) {
                    var filename = _c[_b];
                    mime.registerTextMime({ mime: primaryMime, filename: filename });
                }
            }
            if (Array.isArray(lang.filenamePatterns)) {
                for (var _d = 0, _e = lang.filenamePatterns; _d < _e.length; _d++) {
                    var filenamePattern = _e[_d];
                    mime.registerTextMime({ mime: primaryMime, filepattern: filenamePattern });
                }
            }
            if (typeof lang.firstLine === 'string' && lang.firstLine.length > 0) {
                var firstLineRegexStr = lang.firstLine;
                if (firstLineRegexStr.charAt(0) !== '^') {
                    firstLineRegexStr = '^' + firstLineRegexStr;
                }
                try {
                    var firstLineRegex = new RegExp(firstLineRegexStr);
                    if (!strings.regExpLeadsToEndlessLoop(firstLineRegex)) {
                        mime.registerTextMime({ mime: primaryMime, firstline: firstLineRegex });
                    }
                }
                catch (err) {
                    // Most likely, the regex was bad
                    errors_1.onUnexpectedError(err);
                }
            }
            this.lowerName2Id[lang.id.toLowerCase()] = lang.id;
            if (typeof lang.aliases !== 'undefined' && Array.isArray(lang.aliases)) {
                for (var i = 0; i < lang.aliases.length; i++) {
                    if (!lang.aliases[i] || lang.aliases[i].length === 0) {
                        continue;
                    }
                    this.lowerName2Id[lang.aliases[i].toLowerCase()] = lang.id;
                }
            }
            var containsAliases = (typeof lang.aliases !== 'undefined' && Array.isArray(lang.aliases) && lang.aliases.length > 0);
            if (containsAliases && lang.aliases[0] === null) {
            }
            else {
                var bestName = (containsAliases ? lang.aliases[0] : null) || lang.id;
                this._setLanguageName(lang.id, bestName, containsAliases);
            }
            if (typeof lang.configuration === 'string') {
                this.id2ConfigurationFiles[lang.id] = this.id2ConfigurationFiles[lang.id] || [];
                this.id2ConfigurationFiles[lang.id].push(lang.configuration);
            }
        };
        LanguagesRegistry.prototype.isRegisteredMode = function (mimetypeOrModeId) {
            // Is this a known mime type ?
            if (hasOwnProperty.call(this.mime2LanguageId, mimetypeOrModeId)) {
                return true;
            }
            // Is this a known mode id ?
            return hasOwnProperty.call(this.knownModeIds, mimetypeOrModeId);
        };
        LanguagesRegistry.prototype.getRegisteredModes = function () {
            return Object.keys(this.knownModeIds);
        };
        LanguagesRegistry.prototype.getRegisteredLanguageNames = function () {
            return Object.keys(this.name2LanguageId);
        };
        LanguagesRegistry.prototype.getLanguageName = function (modeId) {
            return this.id2Name[modeId] || null;
        };
        LanguagesRegistry.prototype.getModeIdForLanguageNameLowercase = function (languageNameLower) {
            return this.lowerName2Id[languageNameLower] || null;
        };
        LanguagesRegistry.prototype.getConfigurationFiles = function (modeId) {
            return this.id2ConfigurationFiles[modeId] || [];
        };
        LanguagesRegistry.prototype.getMimeForMode = function (theModeId) {
            var keys = Object.keys(this.mime2LanguageId);
            for (var i = 0, len = keys.length; i < len; i++) {
                var _mime = keys[i];
                var modeId = this.mime2LanguageId[_mime];
                if (modeId === theModeId) {
                    return _mime;
                }
            }
            return null;
        };
        LanguagesRegistry.prototype.extractModeIds = function (commaSeparatedMimetypesOrCommaSeparatedIdsOrName) {
            var _this = this;
            if (!commaSeparatedMimetypesOrCommaSeparatedIdsOrName) {
                return [];
            }
            return (commaSeparatedMimetypesOrCommaSeparatedIdsOrName.
                split(',').
                map(function (mimeTypeOrIdOrName) { return mimeTypeOrIdOrName.trim(); }).
                map(function (mimeTypeOrIdOrName) {
                return _this.mime2LanguageId[mimeTypeOrIdOrName] || mimeTypeOrIdOrName;
            }).
                filter(function (modeId) {
                return _this.knownModeIds[modeId];
            }));
        };
        LanguagesRegistry.prototype.getModeIdsFromLanguageName = function (languageName) {
            if (!languageName) {
                return [];
            }
            if (hasOwnProperty.call(this.name2LanguageId, languageName)) {
                return [this.name2LanguageId[languageName]];
            }
            return [];
        };
        LanguagesRegistry.prototype.getModeIdsFromFilenameOrFirstLine = function (filename, firstLine) {
            if (!filename && !firstLine) {
                return [];
            }
            var mimeTypes = mime.guessMimeTypes(filename, firstLine);
            return this.extractModeIds(mimeTypes.join(','));
        };
        LanguagesRegistry.prototype.getCompatMode = function (modeId) {
            return this.compatModes[modeId] || null;
        };
        LanguagesRegistry.prototype.getExtensions = function (languageName) {
            var languageId = this.name2LanguageId[languageName];
            if (!languageId) {
                return [];
            }
            return this.id2Extensions[languageId];
        };
        return LanguagesRegistry;
    }());
    exports.LanguagesRegistry = LanguagesRegistry;
});

define(__m[41], __M([1,0,79,11,15]), function (require, exports, nls, platform, eventEmitter_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    exports.Extensions = {
        JSONContribution: 'base.contributions.json'
    };
    function normalizeId(id) {
        if (id.length > 0 && id.charAt(id.length - 1) === '#') {
            return id.substring(0, id.length - 1);
        }
        return id;
    }
    var JSONContributionRegistry = (function () {
        function JSONContributionRegistry() {
            this.schemasById = {};
            this.eventEmitter = new eventEmitter_1.EventEmitter();
        }
        JSONContributionRegistry.prototype.addRegistryChangedListener = function (callback) {
            return this.eventEmitter.addListener2('registryChanged', callback);
        };
        JSONContributionRegistry.prototype.registerSchema = function (uri, unresolvedSchemaContent) {
            this.schemasById[normalizeId(uri)] = unresolvedSchemaContent;
            this.eventEmitter.emit('registryChanged', {});
        };
        JSONContributionRegistry.prototype.getSchemaContributions = function () {
            return {
                schemas: this.schemasById,
            };
        };
        return JSONContributionRegistry;
    }());
    var jsonContributionRegistry = new JSONContributionRegistry();
    platform.Registry.add(exports.Extensions.JSONContribution, jsonContributionRegistry);
    // preload the schema-schema with a version that contains descriptions.
    jsonContributionRegistry.registerSchema('http://json-schema.org/draft-04/schema#', {
        'id': 'http://json-schema.org/draft-04/schema#',
        'title': nls.localize(0, null),
        '$schema': 'http://json-schema.org/draft-04/schema#',
        'definitions': {
            'schemaArray': {
                'type': 'array',
                'minItems': 1,
                'items': { '$ref': '#' }
            },
            'positiveInteger': {
                'type': 'integer',
                'minimum': 0
            },
            'positiveIntegerDefault0': {
                'allOf': [{ '$ref': '#/definitions/positiveInteger' }, { 'default': 0 }]
            },
            'simpleTypes': {
                'type': 'string',
                'enum': ['array', 'boolean', 'integer', 'null', 'number', 'object', 'string']
            },
            'stringArray': {
                'type': 'array',
                'items': { 'type': 'string' },
                'minItems': 1,
                'uniqueItems': true
            }
        },
        'type': 'object',
        'properties': {
            'id': {
                'type': 'string',
                'format': 'uri',
                'description': nls.localize(1, null)
            },
            '$schema': {
                'type': 'string',
                'format': 'uri',
                'description': nls.localize(2, null)
            },
            'title': {
                'type': 'string',
                'description': nls.localize(3, null)
            },
            'description': {
                'type': 'string',
                'description': nls.localize(4, null)
            },
            'default': {
                'description': nls.localize(5, null)
            },
            'multipleOf': {
                'type': 'number',
                'minimum': 0,
                'exclusiveMinimum': true,
                'description': nls.localize(6, null)
            },
            'maximum': {
                'type': 'number',
                'description': nls.localize(7, null)
            },
            'exclusiveMaximum': {
                'type': 'boolean',
                'default': false,
                'description': nls.localize(8, null)
            },
            'minimum': {
                'type': 'number',
                'description': nls.localize(9, null)
            },
            'exclusiveMinimum': {
                'type': 'boolean',
                'default': false,
                'description': nls.localize(10, null)
            },
            'maxLength': {
                'allOf': [
                    { '$ref': '#/definitions/positiveInteger' }
                ],
                'description': nls.localize(11, null)
            },
            'minLength': {
                'allOf': [
                    { '$ref': '#/definitions/positiveIntegerDefault0' }
                ],
                'description': nls.localize(12, null)
            },
            'pattern': {
                'type': 'string',
                'format': 'regex',
                'description': nls.localize(13, null)
            },
            'additionalItems': {
                'anyOf': [
                    { 'type': 'boolean' },
                    { '$ref': '#' }
                ],
                'default': {},
                'description': nls.localize(14, null)
            },
            'items': {
                'anyOf': [
                    { '$ref': '#' },
                    { '$ref': '#/definitions/schemaArray' }
                ],
                'default': {},
                'description': nls.localize(15, null)
            },
            'maxItems': {
                'allOf': [
                    { '$ref': '#/definitions/positiveInteger' }
                ],
                'description': nls.localize(16, null)
            },
            'minItems': {
                'allOf': [
                    { '$ref': '#/definitions/positiveIntegerDefault0' }
                ],
                'description': nls.localize(17, null)
            },
            'uniqueItems': {
                'type': 'boolean',
                'default': false,
                'description': nls.localize(18, null)
            },
            'maxProperties': {
                'allOf': [
                    { '$ref': '#/definitions/positiveInteger' }
                ],
                'description': nls.localize(19, null)
            },
            'minProperties': {
                'allOf': [
                    { '$ref': '#/definitions/positiveIntegerDefault0' },
                ],
                'description': nls.localize(20, null)
            },
            'required': {
                'allOf': [
                    { '$ref': '#/definitions/stringArray' }
                ],
                'description': nls.localize(21, null)
            },
            'additionalProperties': {
                'anyOf': [
                    { 'type': 'boolean' },
                    { '$ref': '#' }
                ],
                'default': {},
                'description': nls.localize(22, null)
            },
            'definitions': {
                'type': 'object',
                'additionalProperties': { '$ref': '#' },
                'default': {},
                'description': nls.localize(23, null)
            },
            'properties': {
                'type': 'object',
                'additionalProperties': { '$ref': '#' },
                'default': {},
                'description': nls.localize(24, null)
            },
            'patternProperties': {
                'type': 'object',
                'additionalProperties': { '$ref': '#' },
                'default': {},
                'description': nls.localize(25, null)
            },
            'dependencies': {
                'type': 'object',
                'additionalProperties': {
                    'anyOf': [
                        { '$ref': '#' },
                        { '$ref': '#/definitions/stringArray' }
                    ]
                },
                'description': nls.localize(26, null)
            },
            'enum': {
                'type': 'array',
                'minItems': 1,
                'uniqueItems': true,
                'description': nls.localize(27, null)
            },
            'type': {
                'anyOf': [
                    { '$ref': '#/definitions/simpleTypes' },
                    {
                        'type': 'array',
                        'items': { '$ref': '#/definitions/simpleTypes' },
                        'minItems': 1,
                        'uniqueItems': true
                    }
                ],
                'description': nls.localize(28, null)
            },
            'format': {
                'anyOf': [
                    {
                        'type': 'string',
                        'description': nls.localize(29, null),
                        'enum': ['date-time', 'uri', 'email', 'hostname', 'ipv4', 'ipv6', 'regex']
                    }, {
                        'type': 'string'
                    }
                ]
            },
            'allOf': {
                'allOf': [
                    { '$ref': '#/definitions/schemaArray' }
                ],
                'description': nls.localize(30, null)
            },
            'anyOf': {
                'allOf': [
                    { '$ref': '#/definitions/schemaArray' }
                ],
                'description': nls.localize(31, null)
            },
            'oneOf': {
                'allOf': [
                    { '$ref': '#/definitions/schemaArray' }
                ],
                'description': nls.localize(32, null)
            },
            'not': {
                'allOf': [
                    { '$ref': '#' }
                ],
                'description': nls.localize(33, null)
            }
        },
        'dependencies': {
            'exclusiveMaximum': ['maximum'],
            'exclusiveMinimum': ['minimum']
        },
        'default': {}
    });
});

define(__m[42], __M([1,0,78,2,13,32,41,11]), function (require, exports, nls, errors_1, paths, severity_1, jsonContributionRegistry_1, platform_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var ExtensionMessageCollector = (function () {
        function ExtensionMessageCollector(messageHandler, source) {
            this._messageHandler = messageHandler;
            this._source = source;
        }
        ExtensionMessageCollector.prototype._msg = function (type, message) {
            this._messageHandler({
                type: type,
                message: message,
                source: this._source
            });
        };
        ExtensionMessageCollector.prototype.error = function (message) {
            this._msg(severity_1.default.Error, message);
        };
        ExtensionMessageCollector.prototype.warn = function (message) {
            this._msg(severity_1.default.Warning, message);
        };
        ExtensionMessageCollector.prototype.info = function (message) {
            this._msg(severity_1.default.Info, message);
        };
        return ExtensionMessageCollector;
    }());
    function isValidExtensionDescription(extensionFolderPath, extensionDescription, notices) {
        if (!extensionDescription) {
            notices.push(nls.localize(0, null));
            return false;
        }
        if (typeof extensionDescription.publisher !== 'string') {
            notices.push(nls.localize(1, null, 'publisher'));
            return false;
        }
        if (typeof extensionDescription.name !== 'string') {
            notices.push(nls.localize(2, null, 'name'));
            return false;
        }
        if (typeof extensionDescription.version !== 'string') {
            notices.push(nls.localize(3, null, 'version'));
            return false;
        }
        if (!extensionDescription.engines) {
            notices.push(nls.localize(4, null, 'engines'));
            return false;
        }
        if (typeof extensionDescription.engines.vscode !== 'string') {
            notices.push(nls.localize(5, null, 'engines.vscode'));
            return false;
        }
        if (typeof extensionDescription.extensionDependencies !== 'undefined') {
            if (!_isStringArray(extensionDescription.extensionDependencies)) {
                notices.push(nls.localize(6, null, 'extensionDependencies'));
                return false;
            }
        }
        if (typeof extensionDescription.activationEvents !== 'undefined') {
            if (!_isStringArray(extensionDescription.activationEvents)) {
                notices.push(nls.localize(7, null, 'activationEvents'));
                return false;
            }
            if (typeof extensionDescription.main === 'undefined') {
                notices.push(nls.localize(8, null, 'activationEvents', 'main'));
                return false;
            }
        }
        if (typeof extensionDescription.main !== 'undefined') {
            if (typeof extensionDescription.main !== 'string') {
                notices.push(nls.localize(9, null, 'main'));
                return false;
            }
            else {
                var normalizedAbsolutePath = paths.normalize(paths.join(extensionFolderPath, extensionDescription.main));
                if (normalizedAbsolutePath.indexOf(extensionFolderPath)) {
                    notices.push(nls.localize(10, null, normalizedAbsolutePath, extensionFolderPath));
                }
            }
            if (typeof extensionDescription.activationEvents === 'undefined') {
                notices.push(nls.localize(11, null, 'activationEvents', 'main'));
                return false;
            }
        }
        return true;
    }
    exports.isValidExtensionDescription = isValidExtensionDescription;
    var hasOwnProperty = Object.hasOwnProperty;
    var schemaRegistry = platform_1.Registry.as(jsonContributionRegistry_1.Extensions.JSONContribution);
    var ExtensionPoint = (function () {
        function ExtensionPoint(name, registry) {
            this.name = name;
            this._registry = registry;
            this._handler = null;
            this._messageHandler = null;
        }
        ExtensionPoint.prototype.setHandler = function (handler) {
            if (this._handler) {
                throw new Error('Handler already set!');
            }
            this._handler = handler;
            this._handle();
        };
        ExtensionPoint.prototype.handle = function (messageHandler) {
            this._messageHandler = messageHandler;
            this._handle();
        };
        ExtensionPoint.prototype._handle = function () {
            var _this = this;
            if (!this._handler || !this._messageHandler) {
                return;
            }
            this._registry.registerPointListener(this.name, function (descriptions) {
                var users = descriptions.map(function (desc) {
                    return {
                        description: desc,
                        value: desc.contributes[_this.name],
                        collector: new ExtensionMessageCollector(_this._messageHandler, desc.extensionFolderPath)
                    };
                });
                _this._handler(users);
            });
        };
        return ExtensionPoint;
    }());
    var schemaId = 'vscode://schemas/vscode-extensions';
    var schema = {
        default: {
            'name': '{{name}}',
            'description': '{{description}}',
            'author': '{{author}}',
            'version': '{{1.0.0}}',
            'main': '{{pathToMain}}',
            'dependencies': {}
        },
        properties: {
            // engines: {
            // 	required: [ 'vscode' ],
            // 	properties: {
            // 		'vscode': {
            // 			type: 'string',
            // 			description: nls.localize('vscode.extension.engines.vscode', 'Specifies that this package only runs inside VSCode of the given version.'),
            // 		}
            // 	}
            // },
            displayName: {
                description: nls.localize(12, null),
                type: 'string'
            },
            categories: {
                description: nls.localize(13, null),
                type: 'array',
                items: {
                    type: 'string',
                    enum: ['Languages', 'Snippets', 'Linters', 'Themes', 'Debuggers', 'Productivity', 'Other']
                }
            },
            galleryBanner: {
                type: 'object',
                description: nls.localize(14, null),
                properties: {
                    color: {
                        description: nls.localize(15, null),
                        type: 'string'
                    },
                    theme: {
                        description: nls.localize(16, null),
                        type: 'string',
                        enum: ['dark', 'light']
                    }
                }
            },
            publisher: {
                description: nls.localize(17, null),
                type: 'string'
            },
            activationEvents: {
                description: nls.localize(18, null),
                type: 'array',
                items: {
                    type: 'string'
                }
            },
            extensionDependencies: {
                description: nls.localize(19, null),
                type: 'array',
                items: {
                    type: 'string'
                }
            },
            scripts: {
                type: 'object',
                properties: {
                    'vscode:prepublish': {
                        description: nls.localize(20, null),
                        type: 'string'
                    }
                }
            },
            contributes: {
                description: nls.localize(21, null),
                type: 'object',
                properties: {},
                default: {}
            }
        }
    };
    var ExtensionsRegistryImpl = (function () {
        function ExtensionsRegistryImpl() {
            this._extensionsMap = {};
            this._extensionsArr = [];
            this._activationMap = {};
            this._pointListeners = [];
            this._extensionPoints = {};
            this._oneTimeActivationEventListeners = {};
        }
        ExtensionsRegistryImpl.prototype.registerPointListener = function (point, handler) {
            var entry = {
                extensionPoint: point,
                listener: handler
            };
            this._pointListeners.push(entry);
            this._triggerPointListener(entry, ExtensionsRegistryImpl._filterWithExtPoint(this.getAllExtensionDescriptions(), point));
        };
        ExtensionsRegistryImpl.prototype.registerExtensionPoint = function (extensionPoint, jsonSchema) {
            if (hasOwnProperty.call(this._extensionPoints, extensionPoint)) {
                throw new Error('Duplicate extension point: ' + extensionPoint);
            }
            var result = new ExtensionPoint(extensionPoint, this);
            this._extensionPoints[extensionPoint] = result;
            schema.properties['contributes'].properties[extensionPoint] = jsonSchema;
            schemaRegistry.registerSchema(schemaId, schema);
            return result;
        };
        ExtensionsRegistryImpl.prototype.handleExtensionPoints = function (messageHandler) {
            var _this = this;
            Object.keys(this._extensionPoints).forEach(function (extensionPointName) {
                _this._extensionPoints[extensionPointName].handle(messageHandler);
            });
        };
        ExtensionsRegistryImpl.prototype._triggerPointListener = function (handler, desc) {
            // console.log('_triggerPointListeners: ' + desc.length + ' OF ' + handler.extensionPoint);
            if (!desc || desc.length === 0) {
                return;
            }
            try {
                handler.listener(desc);
            }
            catch (e) {
                errors_1.onUnexpectedError(e);
            }
        };
        ExtensionsRegistryImpl.prototype.registerExtensions = function (extensionDescriptions) {
            for (var i = 0, len = extensionDescriptions.length; i < len; i++) {
                var extensionDescription = extensionDescriptions[i];
                if (hasOwnProperty.call(this._extensionsMap, extensionDescription.id)) {
                    // No overwriting allowed!
                    console.error('Extension `' + extensionDescription.id + '` is already registered');
                    continue;
                }
                this._extensionsMap[extensionDescription.id] = extensionDescription;
                this._extensionsArr.push(extensionDescription);
                if (Array.isArray(extensionDescription.activationEvents)) {
                    for (var j = 0, lenJ = extensionDescription.activationEvents.length; j < lenJ; j++) {
                        var activationEvent = extensionDescription.activationEvents[j];
                        this._activationMap[activationEvent] = this._activationMap[activationEvent] || [];
                        this._activationMap[activationEvent].push(extensionDescription);
                    }
                }
            }
            for (var i = 0, len = this._pointListeners.length; i < len; i++) {
                var listenerEntry = this._pointListeners[i];
                var descriptions = ExtensionsRegistryImpl._filterWithExtPoint(extensionDescriptions, listenerEntry.extensionPoint);
                this._triggerPointListener(listenerEntry, descriptions);
            }
        };
        ExtensionsRegistryImpl._filterWithExtPoint = function (input, point) {
            return input.filter(function (desc) {
                return (desc.contributes && hasOwnProperty.call(desc.contributes, point));
            });
        };
        ExtensionsRegistryImpl.prototype.getExtensionDescriptionsForActivationEvent = function (activationEvent) {
            if (!hasOwnProperty.call(this._activationMap, activationEvent)) {
                return [];
            }
            return this._activationMap[activationEvent].slice(0);
        };
        ExtensionsRegistryImpl.prototype.getAllExtensionDescriptions = function () {
            return this._extensionsArr.slice(0);
        };
        ExtensionsRegistryImpl.prototype.getExtensionDescription = function (extensionId) {
            if (!hasOwnProperty.call(this._extensionsMap, extensionId)) {
                return null;
            }
            return this._extensionsMap[extensionId];
        };
        ExtensionsRegistryImpl.prototype.registerOneTimeActivationEventListener = function (activationEvent, listener) {
            if (!hasOwnProperty.call(this._oneTimeActivationEventListeners, activationEvent)) {
                this._oneTimeActivationEventListeners[activationEvent] = [];
            }
            this._oneTimeActivationEventListeners[activationEvent].push(listener);
        };
        ExtensionsRegistryImpl.prototype.triggerActivationEventListeners = function (activationEvent) {
            if (hasOwnProperty.call(this._oneTimeActivationEventListeners, activationEvent)) {
                var listeners = this._oneTimeActivationEventListeners[activationEvent];
                delete this._oneTimeActivationEventListeners[activationEvent];
                for (var i = 0, len = listeners.length; i < len; i++) {
                    var listener = listeners[i];
                    try {
                        listener();
                    }
                    catch (e) {
                        errors_1.onUnexpectedError(e);
                    }
                }
            }
        };
        return ExtensionsRegistryImpl;
    }());
    function _isStringArray(arr) {
        if (!Array.isArray(arr)) {
            return false;
        }
        for (var i = 0, len = arr.length; i < len; i++) {
            if (typeof arr[i] !== 'string') {
                return false;
            }
        }
        return true;
    }
    var PRExtensions = {
        ExtensionsRegistry: 'ExtensionsRegistry'
    };
    platform_1.Registry.add(PRExtensions.ExtensionsRegistry, new ExtensionsRegistryImpl());
    exports.ExtensionsRegistry = platform_1.Registry.as(PRExtensions.ExtensionsRegistry);
    schemaRegistry.registerSchema(schemaId, schema);
});

define(__m[100], __M([1,0,76,8,11,12,42,41]), function (require, exports, nls, event_1, platform, objects, extensionsRegistry_1, JSONContributionRegistry) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    exports.Extensions = {
        Configuration: 'base.contributions.configuration'
    };
    var schemaId = 'vscode://schemas/settings';
    var contributionRegistry = platform.Registry.as(JSONContributionRegistry.Extensions.JSONContribution);
    var ConfigurationRegistry = (function () {
        function ConfigurationRegistry() {
            this.configurationContributors = [];
            this.configurationSchema = { allOf: [] };
            this._onDidRegisterConfiguration = new event_1.Emitter();
            contributionRegistry.registerSchema(schemaId, this.configurationSchema);
        }
        Object.defineProperty(ConfigurationRegistry.prototype, "onDidRegisterConfiguration", {
            get: function () {
                return this._onDidRegisterConfiguration.event;
            },
            enumerable: true,
            configurable: true
        });
        ConfigurationRegistry.prototype.registerConfiguration = function (configuration) {
            this.configurationContributors.push(configuration);
            this.registerJSONConfiguration(configuration);
            this._onDidRegisterConfiguration.fire(this);
        };
        ConfigurationRegistry.prototype.getConfigurations = function () {
            return this.configurationContributors.slice(0);
        };
        ConfigurationRegistry.prototype.registerJSONConfiguration = function (configuration) {
            var schema = objects.clone(configuration);
            this.configurationSchema.allOf.push(schema);
            contributionRegistry.registerSchema(schemaId, this.configurationSchema);
        };
        return ConfigurationRegistry;
    }());
    var configurationRegistry = new ConfigurationRegistry();
    platform.Registry.add(exports.Extensions.Configuration, configurationRegistry);
    var configurationExtPoint = extensionsRegistry_1.ExtensionsRegistry.registerExtensionPoint('configuration', {
        description: nls.localize(0, null),
        type: 'object',
        defaultSnippets: [{ body: { title: '', properties: {} } }],
        properties: {
            title: {
                description: nls.localize(1, null),
                type: 'string'
            },
            properties: {
                description: nls.localize(2, null),
                type: 'object',
                additionalProperties: {
                    $ref: 'http://json-schema.org/draft-04/schema#'
                }
            }
        }
    });
    configurationExtPoint.setHandler(function (extensions) {
        for (var i = 0; i < extensions.length; i++) {
            var configuration = extensions[i].value;
            var collector = extensions[i].collector;
            if (configuration.type && configuration.type !== 'object') {
                collector.warn(nls.localize(3, null));
            }
            else {
                configuration.type = 'object';
            }
            if (configuration.title && (typeof configuration.title !== 'string')) {
                collector.error(nls.localize(4, null));
            }
            if (configuration.properties && (typeof configuration.properties !== 'object')) {
                collector.error(nls.localize(5, null));
                return;
            }
            var clonedConfiguration = objects.clone(configuration);
            clonedConfiguration.id = extensions[i].description.id;
            configurationRegistry.registerConfiguration(clonedConfiguration);
        }
    });
});

define(__m[59], __M([1,0,80,100,11,74,24]), function (require, exports, filters_1, configurationRegistry_1, platform_1, nls_1, async_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var TextualSuggestSupport = (function () {
        function TextualSuggestSupport(editorWorkerService, configurationService) {
            this._editorWorkerService = editorWorkerService;
            this._configurationService = configurationService;
        }
        Object.defineProperty(TextualSuggestSupport.prototype, "triggerCharacters", {
            /* tslint:enable */
            get: function () {
                return [];
            },
            enumerable: true,
            configurable: true
        });
        Object.defineProperty(TextualSuggestSupport.prototype, "shouldAutotriggerSuggest", {
            get: function () {
                return true;
            },
            enumerable: true,
            configurable: true
        });
        Object.defineProperty(TextualSuggestSupport.prototype, "filter", {
            get: function () {
                return filters_1.matchesStrictPrefix;
            },
            enumerable: true,
            configurable: true
        });
        TextualSuggestSupport.prototype.provideCompletionItems = function (model, position, token) {
            var config = this._configurationService.getConfiguration('editor');
            if (!config || config.wordBasedSuggestions) {
                return async_1.wireCancellationToken(token, this._editorWorkerService.textualSuggest(model.uri, position));
            }
            return [];
        };
        /* tslint:disable */
        TextualSuggestSupport._c = platform_1.Registry.as(configurationRegistry_1.Extensions.Configuration).registerConfiguration({
            type: 'object',
            order: 5.1,
            properties: {
                'editor.wordBasedSuggestions': {
                    'type': 'boolean',
                    'description': nls_1.localize(0, null),
                    'default': true
                }
            }
        });
        return TextualSuggestSupport;
    }());
    exports.TextualSuggestSupport = TextualSuggestSupport;
    function filterSuggestions(value) {
        if (!value) {
            return;
        }
        // filter suggestions
        var accept = filters_1.fuzzyContiguousFilter, result = [];
        result.push({
            currentWord: value.currentWord,
            suggestions: value.suggestions.filter(function (element) { return !!accept(value.currentWord, element.label); }),
            incomplete: value.incomplete
        });
        return result;
    }
    exports.filterSuggestions = filterSuggestions;
});















define(__m[60], __M([1,0,15,5,33,4,51,21,59,83,17]), function (require, exports, eventEmitter_1, winjs_base_1, descriptors_1, instantiation_1, configuration_1, modes, suggestSupport_1, editorWorkerService_1, wordHelper) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    function createWordRegExp(allowInWords) {
        if (allowInWords === void 0) { allowInWords = ''; }
        return wordHelper.createWordRegExp(allowInWords);
    }
    exports.createWordRegExp = createWordRegExp;
    var ModeWorkerManager = (function () {
        function ModeWorkerManager(descriptor, workerModuleId, workerClassName, superWorkerModuleId, instantiationService) {
            this._descriptor = descriptor;
            this._workerDescriptor = descriptors_1.createAsyncDescriptor1(workerModuleId, workerClassName);
            this._superWorkerModuleId = superWorkerModuleId;
            this._instantiationService = instantiationService;
            this._workerPiecePromise = null;
        }
        ModeWorkerManager.prototype.worker = function (runner) {
            return this._getOrCreateWorker().then(runner);
        };
        ModeWorkerManager.prototype._getOrCreateWorker = function () {
            var _this = this;
            if (!this._workerPiecePromise) {
                // TODO@Alex: workaround for missing `bundles` config
                // First, load the code of the worker super class
                var superWorkerCodePromise = (this._superWorkerModuleId ? ModeWorkerManager._loadModule(this._superWorkerModuleId) : winjs_base_1.TPromise.as(null));
                this._workerPiecePromise = superWorkerCodePromise.then(function () {
                    // Second, load the code of the worker (without instantiating it)
                    return ModeWorkerManager._loadModule(_this._workerDescriptor.moduleName);
                }).then(function () {
                    // Finally, create the mode worker instance
                    return _this._instantiationService.createInstance(_this._workerDescriptor, _this._descriptor.id);
                });
            }
            return this._workerPiecePromise;
        };
        ModeWorkerManager._loadModule = function (moduleName) {
            return new winjs_base_1.TPromise(function (c, e, p) {
                // Use the global require to be sure to get the global config
                self.require([moduleName], c, e);
            }, function () {
                // Cannot cancel loading code
            });
        };
        return ModeWorkerManager;
    }());
    exports.ModeWorkerManager = ModeWorkerManager;
    var AbstractMode = (function () {
        function AbstractMode(modeId) {
            this._modeId = modeId;
            this._eventEmitter = new eventEmitter_1.EventEmitter();
            this._simplifiedMode = null;
        }
        AbstractMode.prototype.getId = function () {
            return this._modeId;
        };
        AbstractMode.prototype.toSimplifiedMode = function () {
            if (!this._simplifiedMode) {
                this._simplifiedMode = new SimplifiedMode(this);
            }
            return this._simplifiedMode;
        };
        AbstractMode.prototype.addSupportChangedListener = function (callback) {
            return this._eventEmitter.addListener2('modeSupportChanged', callback);
        };
        AbstractMode.prototype.setTokenizationSupport = function (callback) {
            var _this = this;
            var supportImpl = callback(this);
            this['tokenizationSupport'] = supportImpl;
            this._eventEmitter.emit('modeSupportChanged', _createModeSupportChangedEvent());
            return {
                dispose: function () {
                    if (_this['tokenizationSupport'] === supportImpl) {
                        delete _this['tokenizationSupport'];
                        _this._eventEmitter.emit('modeSupportChanged', _createModeSupportChangedEvent());
                    }
                }
            };
        };
        return AbstractMode;
    }());
    exports.AbstractMode = AbstractMode;
    var CompatMode = (function (_super) {
        __extends(CompatMode, _super);
        function CompatMode(modeId, compatWorkerService) {
            _super.call(this, modeId);
            this.compatWorkerService = compatWorkerService;
            if (this.compatWorkerService) {
                this.compatWorkerService.registerCompatMode(this);
            }
        }
        return CompatMode;
    }(AbstractMode));
    exports.CompatMode = CompatMode;
    var SimplifiedMode = (function () {
        function SimplifiedMode(sourceMode) {
            var _this = this;
            this._sourceMode = sourceMode;
            this._eventEmitter = new eventEmitter_1.EventEmitter();
            this._id = 'vs.editor.modes.simplifiedMode:' + sourceMode.getId();
            this._assignSupports();
            if (this._sourceMode.addSupportChangedListener) {
                this._sourceMode.addSupportChangedListener(function (e) {
                    _this._assignSupports();
                    _this._eventEmitter.emit('modeSupportChanged', e);
                });
            }
        }
        SimplifiedMode.prototype.getId = function () {
            return this._id;
        };
        SimplifiedMode.prototype.toSimplifiedMode = function () {
            return this;
        };
        SimplifiedMode.prototype._assignSupports = function () {
            this.tokenizationSupport = this._sourceMode.tokenizationSupport;
        };
        return SimplifiedMode;
    }());
    exports.isDigit = (function () {
        var _0 = '0'.charCodeAt(0), _1 = '1'.charCodeAt(0), _2 = '2'.charCodeAt(0), _3 = '3'.charCodeAt(0), _4 = '4'.charCodeAt(0), _5 = '5'.charCodeAt(0), _6 = '6'.charCodeAt(0), _7 = '7'.charCodeAt(0), _8 = '8'.charCodeAt(0), _9 = '9'.charCodeAt(0), _a = 'a'.charCodeAt(0), _b = 'b'.charCodeAt(0), _c = 'c'.charCodeAt(0), _d = 'd'.charCodeAt(0), _e = 'e'.charCodeAt(0), _f = 'f'.charCodeAt(0), _A = 'A'.charCodeAt(0), _B = 'B'.charCodeAt(0), _C = 'C'.charCodeAt(0), _D = 'D'.charCodeAt(0), _E = 'E'.charCodeAt(0), _F = 'F'.charCodeAt(0);
        return function isDigit(character, base) {
            var c = character.charCodeAt(0);
            switch (base) {
                case 1:
                    return c === _0;
                case 2:
                    return c >= _0 && c <= _1;
                case 3:
                    return c >= _0 && c <= _2;
                case 4:
                    return c >= _0 && c <= _3;
                case 5:
                    return c >= _0 && c <= _4;
                case 6:
                    return c >= _0 && c <= _5;
                case 7:
                    return c >= _0 && c <= _6;
                case 8:
                    return c >= _0 && c <= _7;
                case 9:
                    return c >= _0 && c <= _8;
                case 10:
                    return c >= _0 && c <= _9;
                case 11:
                    return (c >= _0 && c <= _9) || (c === _a) || (c === _A);
                case 12:
                    return (c >= _0 && c <= _9) || (c >= _a && c <= _b) || (c >= _A && c <= _B);
                case 13:
                    return (c >= _0 && c <= _9) || (c >= _a && c <= _c) || (c >= _A && c <= _C);
                case 14:
                    return (c >= _0 && c <= _9) || (c >= _a && c <= _d) || (c >= _A && c <= _D);
                case 15:
                    return (c >= _0 && c <= _9) || (c >= _a && c <= _e) || (c >= _A && c <= _E);
                default:
                    return (c >= _0 && c <= _9) || (c >= _a && c <= _f) || (c >= _A && c <= _F);
            }
        };
    })();
    var FrankensteinMode = (function (_super) {
        __extends(FrankensteinMode, _super);
        function FrankensteinMode(descriptor, configurationService, editorWorkerService) {
            _super.call(this, descriptor.id);
            if (editorWorkerService) {
                modes.SuggestRegistry.register(this.getId(), new suggestSupport_1.TextualSuggestSupport(editorWorkerService, configurationService), true);
            }
        }
        FrankensteinMode = __decorate([
            __param(1, configuration_1.IConfigurationService),
            __param(2, instantiation_1.optional(editorWorkerService_1.IEditorWorkerService))
        ], FrankensteinMode);
        return FrankensteinMode;
    }(AbstractMode));
    exports.FrankensteinMode = FrankensteinMode;
    function _createModeSupportChangedEvent() {
        return {
            tokenizationSupport: true
        };
    }
});















define(__m[103], __M([1,0,75,2,8,16,12,13,5,46,33,38,42,4,60,40,97,23,51,39,14]), function (require, exports, nls, errors_1, event_1, lifecycle_1, objects, paths, winjs_base_1, mime, descriptors_1, extensions_1, extensionsRegistry_1, instantiation_1, abstractMode_1, modesRegistry_1, languagesRegistry_1, modeService_1, configuration_1, abstractState_1, supports_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var languagesExtPoint = extensionsRegistry_1.ExtensionsRegistry.registerExtensionPoint('languages', {
        description: nls.localize(0, null),
        type: 'array',
        defaultSnippets: [{ body: [{ id: '', aliases: [], extensions: [] }] }],
        items: {
            type: 'object',
            defaultSnippets: [{ body: { id: '', extensions: [] } }],
            properties: {
                id: {
                    description: nls.localize(1, null),
                    type: 'string'
                },
                aliases: {
                    description: nls.localize(2, null),
                    type: 'array',
                    items: {
                        type: 'string'
                    }
                },
                extensions: {
                    description: nls.localize(3, null),
                    default: ['.foo'],
                    type: 'array',
                    items: {
                        type: 'string'
                    }
                },
                filenames: {
                    description: nls.localize(4, null),
                    type: 'array',
                    items: {
                        type: 'string'
                    }
                },
                filenamePatterns: {
                    description: nls.localize(5, null),
                    type: 'array',
                    items: {
                        type: 'string'
                    }
                },
                mimetypes: {
                    description: nls.localize(6, null),
                    type: 'array',
                    items: {
                        type: 'string'
                    }
                },
                firstLine: {
                    description: nls.localize(7, null),
                    type: 'string'
                },
                configuration: {
                    description: nls.localize(8, null),
                    type: 'string'
                }
            }
        }
    });
    function isUndefinedOrStringArray(value) {
        if (typeof value === 'undefined') {
            return true;
        }
        if (!Array.isArray(value)) {
            return false;
        }
        return value.every(function (item) { return typeof item === 'string'; });
    }
    function isValidLanguageExtensionPoint(value, collector) {
        if (!value) {
            collector.error(nls.localize(9, null, languagesExtPoint.name));
            return false;
        }
        if (typeof value.id !== 'string') {
            collector.error(nls.localize(10, null, 'id'));
            return false;
        }
        if (!isUndefinedOrStringArray(value.extensions)) {
            collector.error(nls.localize(11, null, 'extensions'));
            return false;
        }
        if (!isUndefinedOrStringArray(value.filenames)) {
            collector.error(nls.localize(12, null, 'filenames'));
            return false;
        }
        if (typeof value.firstLine !== 'undefined' && typeof value.firstLine !== 'string') {
            collector.error(nls.localize(13, null, 'firstLine'));
            return false;
        }
        if (typeof value.configuration !== 'undefined' && typeof value.configuration !== 'string') {
            collector.error(nls.localize(14, null, 'configuration'));
            return false;
        }
        if (!isUndefinedOrStringArray(value.aliases)) {
            collector.error(nls.localize(15, null, 'aliases'));
            return false;
        }
        if (!isUndefinedOrStringArray(value.mimetypes)) {
            collector.error(nls.localize(16, null, 'mimetypes'));
            return false;
        }
        return true;
    }
    var ModeServiceImpl = (function () {
        function ModeServiceImpl(instantiationService, extensionService) {
            var _this = this;
            this.serviceId = modeService_1.IModeService;
            this._onDidAddModes = new event_1.Emitter();
            this.onDidAddModes = this._onDidAddModes.event;
            this._onDidCreateMode = new event_1.Emitter();
            this.onDidCreateMode = this._onDidCreateMode.event;
            this._instantiationService = instantiationService;
            this._extensionService = extensionService;
            this._activationPromises = {};
            this._instantiatedModes = {};
            this._config = {};
            this._registry = new languagesRegistry_1.LanguagesRegistry();
            this._registry.onDidAddModes(function (modes) { return _this._onDidAddModes.fire(modes); });
        }
        ModeServiceImpl.prototype.getConfigurationForMode = function (modeId) {
            return this._config[modeId] || {};
        };
        ModeServiceImpl.prototype.configureMode = function (mimetype, options) {
            var modeId = this.getModeId(mimetype);
            if (modeId) {
                this.configureModeById(modeId, options);
            }
        };
        ModeServiceImpl.prototype.configureModeById = function (modeId, options) {
            var previousOptions = this._config[modeId] || {};
            var newOptions = objects.mixin(objects.clone(previousOptions), options);
            if (objects.equals(previousOptions, newOptions)) {
                // This configure call is a no-op
                return;
            }
            this._config[modeId] = newOptions;
            var mode = this.getMode(modeId);
            if (mode && mode.configSupport) {
                mode.configSupport.configure(this.getConfigurationForMode(modeId));
            }
        };
        ModeServiceImpl.prototype.configureAllModes = function (config) {
            var _this = this;
            if (!config) {
                return;
            }
            var modes = this._registry.getRegisteredModes();
            modes.forEach(function (modeIdentifier) {
                var configuration = config[modeIdentifier];
                _this.configureModeById(modeIdentifier, configuration);
            });
        };
        ModeServiceImpl.prototype.isRegisteredMode = function (mimetypeOrModeId) {
            return this._registry.isRegisteredMode(mimetypeOrModeId);
        };
        ModeServiceImpl.prototype.isCompatMode = function (modeId) {
            var compatModeData = this._registry.getCompatMode(modeId);
            return (compatModeData ? true : false);
        };
        ModeServiceImpl.prototype.getRegisteredModes = function () {
            return this._registry.getRegisteredModes();
        };
        ModeServiceImpl.prototype.getRegisteredLanguageNames = function () {
            return this._registry.getRegisteredLanguageNames();
        };
        ModeServiceImpl.prototype.getExtensions = function (alias) {
            return this._registry.getExtensions(alias);
        };
        ModeServiceImpl.prototype.getMimeForMode = function (modeId) {
            return this._registry.getMimeForMode(modeId);
        };
        ModeServiceImpl.prototype.getLanguageName = function (modeId) {
            return this._registry.getLanguageName(modeId);
        };
        ModeServiceImpl.prototype.getModeIdForLanguageName = function (alias) {
            return this._registry.getModeIdForLanguageNameLowercase(alias);
        };
        ModeServiceImpl.prototype.getModeId = function (commaSeparatedMimetypesOrCommaSeparatedIds) {
            var modeIds = this._registry.extractModeIds(commaSeparatedMimetypesOrCommaSeparatedIds);
            if (modeIds.length > 0) {
                return modeIds[0];
            }
            return null;
        };
        ModeServiceImpl.prototype.getConfigurationFiles = function (modeId) {
            return this._registry.getConfigurationFiles(modeId);
        };
        // --- instantiation
        ModeServiceImpl.prototype.lookup = function (commaSeparatedMimetypesOrCommaSeparatedIds) {
            var r = [];
            var modeIds = this._registry.extractModeIds(commaSeparatedMimetypesOrCommaSeparatedIds);
            for (var i = 0; i < modeIds.length; i++) {
                var modeId = modeIds[i];
                r.push({
                    modeId: modeId,
                    isInstantiated: this._instantiatedModes.hasOwnProperty(modeId)
                });
            }
            return r;
        };
        ModeServiceImpl.prototype.getMode = function (commaSeparatedMimetypesOrCommaSeparatedIds) {
            var modeIds = this._registry.extractModeIds(commaSeparatedMimetypesOrCommaSeparatedIds);
            var isPlainText = false;
            for (var i = 0; i < modeIds.length; i++) {
                if (this._instantiatedModes.hasOwnProperty(modeIds[i])) {
                    return this._instantiatedModes[modeIds[i]];
                }
                isPlainText = isPlainText || (modeIds[i] === 'plaintext');
            }
            if (isPlainText) {
                // Try to do it synchronously
                var r = null;
                this.getOrCreateMode(commaSeparatedMimetypesOrCommaSeparatedIds).then(function (mode) {
                    r = mode;
                }).done(null, errors_1.onUnexpectedError);
                return r;
            }
        };
        ModeServiceImpl.prototype.getModeIdByLanguageName = function (languageName) {
            var modeIds = this._registry.getModeIdsFromLanguageName(languageName);
            if (modeIds.length > 0) {
                return modeIds[0];
            }
            return null;
        };
        ModeServiceImpl.prototype.getModeIdByFilenameOrFirstLine = function (filename, firstLine) {
            var modeIds = this._registry.getModeIdsFromFilenameOrFirstLine(filename, firstLine);
            if (modeIds.length > 0) {
                return modeIds[0];
            }
            return null;
        };
        ModeServiceImpl.prototype.onReady = function () {
            return this._extensionService.onReady();
        };
        ModeServiceImpl.prototype.getOrCreateMode = function (commaSeparatedMimetypesOrCommaSeparatedIds) {
            var _this = this;
            return this.onReady().then(function () {
                var modeId = _this.getModeId(commaSeparatedMimetypesOrCommaSeparatedIds);
                // Fall back to plain text if no mode was found
                return _this._getOrCreateMode(modeId || 'plaintext');
            });
        };
        ModeServiceImpl.prototype.getOrCreateModeByLanguageName = function (languageName) {
            var _this = this;
            return this.onReady().then(function () {
                var modeId = _this.getModeIdByLanguageName(languageName);
                // Fall back to plain text if no mode was found
                return _this._getOrCreateMode(modeId || 'plaintext');
            });
        };
        ModeServiceImpl.prototype.getOrCreateModeByFilenameOrFirstLine = function (filename, firstLine) {
            var _this = this;
            return this.onReady().then(function () {
                var modeId = _this.getModeIdByFilenameOrFirstLine(filename, firstLine);
                // Fall back to plain text if no mode was found
                return _this._getOrCreateMode(modeId || 'plaintext');
            });
        };
        ModeServiceImpl.prototype._getOrCreateMode = function (modeId) {
            var _this = this;
            if (this._instantiatedModes.hasOwnProperty(modeId)) {
                return winjs_base_1.TPromise.as(this._instantiatedModes[modeId]);
            }
            if (this._activationPromises.hasOwnProperty(modeId)) {
                return this._activationPromises[modeId];
            }
            var c, e;
            var promise = new winjs_base_1.TPromise(function (cc, ee, pp) { c = cc; e = ee; });
            this._activationPromises[modeId] = promise;
            this._createMode(modeId).then(function (mode) {
                _this._instantiatedModes[modeId] = mode;
                delete _this._activationPromises[modeId];
                _this._onDidCreateMode.fire(mode);
                _this._extensionService.activateByEvent("onLanguage:" + modeId).done(null, errors_1.onUnexpectedError);
                return _this._instantiatedModes[modeId];
            }).then(c, e);
            return promise;
        };
        ModeServiceImpl.prototype._createMode = function (modeId) {
            var _this = this;
            var modeDescriptor = this._createModeDescriptor(modeId);
            var compatModeData = this._registry.getCompatMode(modeId);
            if (compatModeData) {
                // This is a compatibility mode
                var resolvedDeps = null;
                if (Array.isArray(compatModeData.deps)) {
                    resolvedDeps = winjs_base_1.TPromise.join(compatModeData.deps.map(function (dep) { return _this.getOrCreateMode(dep); }));
                }
                else {
                    resolvedDeps = winjs_base_1.TPromise.as(null);
                }
                return resolvedDeps.then(function (_) {
                    var compatModeAsyncDescriptor = descriptors_1.createAsyncDescriptor1(compatModeData.moduleId, compatModeData.ctorName);
                    return _this._instantiationService.createInstance(compatModeAsyncDescriptor, modeDescriptor).then(function (compatMode) {
                        if (compatMode.configSupport) {
                            compatMode.configSupport.configure(_this.getConfigurationForMode(modeId));
                        }
                        return compatMode;
                    });
                });
            }
            return winjs_base_1.TPromise.as(this._instantiationService.createInstance(abstractMode_1.FrankensteinMode, modeDescriptor));
        };
        ModeServiceImpl.prototype._createModeDescriptor = function (modeId) {
            return {
                id: modeId
            };
        };
        ModeServiceImpl.prototype._registerTokenizationSupport = function (mode, callback) {
            if (mode.setTokenizationSupport) {
                return mode.setTokenizationSupport(callback);
            }
            else {
                console.warn('Cannot register tokenizationSupport on mode ' + mode.getId() + ' because it does not support it.');
                return lifecycle_1.empty;
            }
        };
        ModeServiceImpl.prototype.registerModeSupport = function (modeId, callback) {
            var _this = this;
            if (this._instantiatedModes.hasOwnProperty(modeId)) {
                return this._registerTokenizationSupport(this._instantiatedModes[modeId], callback);
            }
            var cc;
            var promise = new winjs_base_1.TPromise(function (c, e) { cc = c; });
            var disposable = this.onDidCreateMode(function (mode) {
                if (mode.getId() !== modeId) {
                    return;
                }
                cc(_this._registerTokenizationSupport(mode, callback));
                disposable.dispose();
            });
            return {
                dispose: function () {
                    promise.done(function (disposable) { return disposable.dispose(); }, null);
                }
            };
        };
        ModeServiceImpl.prototype.registerTokenizationSupport = function (modeId, callback) {
            return this.registerModeSupport(modeId, callback);
        };
        ModeServiceImpl.prototype.registerTokenizationSupport2 = function (modeId, support) {
            return this.registerModeSupport(modeId, function (mode) {
                return new TokenizationSupport2Adapter(mode, support);
            });
        };
        return ModeServiceImpl;
    }());
    exports.ModeServiceImpl = ModeServiceImpl;
    var TokenizationState2Adapter = (function () {
        function TokenizationState2Adapter(mode, actual, stateData) {
            this._mode = mode;
            this._actual = actual;
            this._stateData = stateData;
        }
        Object.defineProperty(TokenizationState2Adapter.prototype, "actual", {
            get: function () { return this._actual; },
            enumerable: true,
            configurable: true
        });
        TokenizationState2Adapter.prototype.clone = function () {
            return new TokenizationState2Adapter(this._mode, this._actual.clone(), abstractState_1.AbstractState.safeClone(this._stateData));
        };
        TokenizationState2Adapter.prototype.equals = function (other) {
            if (other instanceof TokenizationState2Adapter) {
                if (!this._actual.equals(other._actual)) {
                    return false;
                }
                return abstractState_1.AbstractState.safeEquals(this._stateData, other._stateData);
            }
            return false;
        };
        TokenizationState2Adapter.prototype.getMode = function () {
            return this._mode;
        };
        TokenizationState2Adapter.prototype.tokenize = function (stream) {
            throw new Error('Unexpected tokenize call!');
        };
        TokenizationState2Adapter.prototype.getStateData = function () {
            return this._stateData;
        };
        TokenizationState2Adapter.prototype.setStateData = function (stateData) {
            this._stateData = stateData;
        };
        return TokenizationState2Adapter;
    }());
    exports.TokenizationState2Adapter = TokenizationState2Adapter;
    var TokenizationSupport2Adapter = (function () {
        function TokenizationSupport2Adapter(mode, actual) {
            this._mode = mode;
            this._actual = actual;
        }
        TokenizationSupport2Adapter.prototype.getInitialState = function () {
            return new TokenizationState2Adapter(this._mode, this._actual.getInitialState(), null);
        };
        TokenizationSupport2Adapter.prototype.tokenize = function (line, state, offsetDelta, stopAtOffset) {
            if (offsetDelta === void 0) { offsetDelta = 0; }
            if (state instanceof TokenizationState2Adapter) {
                var actualResult = this._actual.tokenize(line, state.actual);
                var tokens_1 = [];
                actualResult.tokens.forEach(function (t) {
                    if (typeof t.scopes === 'string') {
                        tokens_1.push(new supports_1.Token(t.startIndex + offsetDelta, t.scopes));
                    }
                    else if (Array.isArray(t.scopes) && t.scopes.length === 1) {
                        tokens_1.push(new supports_1.Token(t.startIndex + offsetDelta, t.scopes[0]));
                    }
                    else {
                        throw new Error('Only token scopes as strings or of precisely 1 length are supported at this time!');
                    }
                });
                return {
                    tokens: tokens_1,
                    actualStopOffset: offsetDelta + line.length,
                    endState: new TokenizationState2Adapter(state.getMode(), actualResult.endState, state.getStateData()),
                    modeTransitions: [{ startIndex: offsetDelta, mode: state.getMode() }],
                };
            }
            throw new Error('Unexpected state to tokenize with!');
        };
        return TokenizationSupport2Adapter;
    }());
    exports.TokenizationSupport2Adapter = TokenizationSupport2Adapter;
    var MainThreadModeServiceImpl = (function (_super) {
        __extends(MainThreadModeServiceImpl, _super);
        function MainThreadModeServiceImpl(instantiationService, extensionService, configurationService) {
            var _this = this;
            _super.call(this, instantiationService, extensionService);
            this._configurationService = configurationService;
            languagesExtPoint.setHandler(function (extensions) {
                var allValidLanguages = [];
                for (var i = 0, len = extensions.length; i < len; i++) {
                    var extension = extensions[i];
                    if (!Array.isArray(extension.value)) {
                        extension.collector.error(nls.localize(17, null, languagesExtPoint.name));
                        continue;
                    }
                    for (var j = 0, lenJ = extension.value.length; j < lenJ; j++) {
                        var ext = extension.value[j];
                        if (isValidLanguageExtensionPoint(ext, extension.collector)) {
                            var configuration = (ext.configuration ? paths.join(extension.description.extensionFolderPath, ext.configuration) : ext.configuration);
                            allValidLanguages.push({
                                id: ext.id,
                                extensions: ext.extensions,
                                filenames: ext.filenames,
                                filenamePatterns: ext.filenamePatterns,
                                firstLine: ext.firstLine,
                                aliases: ext.aliases,
                                mimetypes: ext.mimetypes,
                                configuration: configuration
                            });
                        }
                    }
                }
                modesRegistry_1.ModesRegistry.registerLanguages(allValidLanguages);
            });
            this._configurationService.onDidUpdateConfiguration(function (e) { return _this.onConfigurationChange(e.config); });
        }
        MainThreadModeServiceImpl.prototype.onReady = function () {
            var _this = this;
            if (!this._onReadyPromise) {
                var configuration_2 = this._configurationService.getConfiguration();
                this._onReadyPromise = this._extensionService.onReady().then(function () {
                    _this.onConfigurationChange(configuration_2);
                    return true;
                });
            }
            return this._onReadyPromise;
        };
        MainThreadModeServiceImpl.prototype.onConfigurationChange = function (configuration) {
            var _this = this;
            // Clear user configured mime associations
            mime.clearTextMimes(true /* user configured */);
            // Register based on settings
            if (configuration.files && configuration.files.associations) {
                Object.keys(configuration.files.associations).forEach(function (pattern) {
                    mime.registerTextMime({ mime: _this.getMimeForMode(configuration.files.associations[pattern]), filepattern: pattern, userConfigured: true });
                });
            }
        };
        MainThreadModeServiceImpl = __decorate([
            __param(0, instantiation_1.IInstantiationService),
            __param(1, extensions_1.IExtensionService),
            __param(2, configuration_1.IConfigurationService)
        ], MainThreadModeServiceImpl);
        return MainThreadModeServiceImpl;
    }(ModeServiceImpl));
    exports.MainThreadModeServiceImpl = MainThreadModeServiceImpl;
});

define(__m[104], __M([1,0,77,32,5,38,42]), function (require, exports, nls, severity_1, winjs_base_1, extensions_1, extensionsRegistry_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var hasOwnProperty = Object.hasOwnProperty;
    var ActivatedExtension = (function () {
        function ActivatedExtension(activationFailed) {
            this.activationFailed = activationFailed;
        }
        return ActivatedExtension;
    }());
    exports.ActivatedExtension = ActivatedExtension;
    var AbstractExtensionService = (function () {
        function AbstractExtensionService(isReadyByDefault) {
            var _this = this;
            this.serviceId = extensions_1.IExtensionService;
            if (isReadyByDefault) {
                this._onReady = winjs_base_1.TPromise.as(true);
                this._onReadyC = function (v) { };
            }
            else {
                this._onReady = new winjs_base_1.TPromise(function (c, e, p) {
                    _this._onReadyC = c;
                }, function () {
                    console.warn('You should really not try to cancel this ready promise!');
                });
            }
            this._activatingExtensions = {};
            this._activatedExtensions = {};
        }
        AbstractExtensionService.prototype._triggerOnReady = function () {
            this._onReadyC(true);
        };
        AbstractExtensionService.prototype.onReady = function () {
            return this._onReady;
        };
        AbstractExtensionService.prototype.getExtensionsStatus = function () {
            return null;
        };
        AbstractExtensionService.prototype.isActivated = function (extensionId) {
            return hasOwnProperty.call(this._activatedExtensions, extensionId);
        };
        AbstractExtensionService.prototype.activateByEvent = function (activationEvent) {
            var _this = this;
            return this._onReady.then(function () {
                extensionsRegistry_1.ExtensionsRegistry.triggerActivationEventListeners(activationEvent);
                var activateExtensions = extensionsRegistry_1.ExtensionsRegistry.getExtensionDescriptionsForActivationEvent(activationEvent);
                return _this._activateExtensions(activateExtensions, 0);
            });
        };
        AbstractExtensionService.prototype.activateById = function (extensionId) {
            var _this = this;
            return this._onReady.then(function () {
                var desc = extensionsRegistry_1.ExtensionsRegistry.getExtensionDescription(extensionId);
                if (!desc) {
                    throw new Error('Extension `' + extensionId + '` is not known');
                }
                return _this._activateExtensions([desc], 0);
            });
        };
        /**
         * Handle semantics related to dependencies for `currentExtension`.
         * semantics: `redExtensions` must wait for `greenExtensions`.
         */
        AbstractExtensionService.prototype._handleActivateRequest = function (currentExtension, greenExtensions, redExtensions) {
            var depIds = (typeof currentExtension.extensionDependencies === 'undefined' ? [] : currentExtension.extensionDependencies);
            var currentExtensionGetsGreenLight = true;
            for (var j = 0, lenJ = depIds.length; j < lenJ; j++) {
                var depId = depIds[j];
                var depDesc = extensionsRegistry_1.ExtensionsRegistry.getExtensionDescription(depId);
                if (!depDesc) {
                    // Error condition 1: unknown dependency
                    this._showMessage(severity_1.default.Error, nls.localize(0, null, depId, currentExtension.id));
                    this._activatedExtensions[currentExtension.id] = this._createFailedExtension();
                    return;
                }
                if (hasOwnProperty.call(this._activatedExtensions, depId)) {
                    var dep = this._activatedExtensions[depId];
                    if (dep.activationFailed) {
                        // Error condition 2: a dependency has already failed activation
                        this._showMessage(severity_1.default.Error, nls.localize(1, null, depId, currentExtension.id));
                        this._activatedExtensions[currentExtension.id] = this._createFailedExtension();
                        return;
                    }
                }
                else {
                    // must first wait for the dependency to activate
                    currentExtensionGetsGreenLight = false;
                    greenExtensions[depId] = depDesc;
                }
            }
            if (currentExtensionGetsGreenLight) {
                greenExtensions[currentExtension.id] = currentExtension;
            }
            else {
                redExtensions.push(currentExtension);
            }
        };
        AbstractExtensionService.prototype._activateExtensions = function (extensionDescriptions, recursionLevel) {
            var _this = this;
            // console.log(recursionLevel, '_activateExtensions: ', extensionDescriptions.map(p => p.id));
            if (extensionDescriptions.length === 0) {
                return winjs_base_1.TPromise.as(void 0);
            }
            extensionDescriptions = extensionDescriptions.filter(function (p) { return !hasOwnProperty.call(_this._activatedExtensions, p.id); });
            if (extensionDescriptions.length === 0) {
                return winjs_base_1.TPromise.as(void 0);
            }
            if (recursionLevel > 10) {
                // More than 10 dependencies deep => most likely a dependency loop
                for (var i = 0, len = extensionDescriptions.length; i < len; i++) {
                    // Error condition 3: dependency loop
                    this._showMessage(severity_1.default.Error, nls.localize(2, null, extensionDescriptions[i].id));
                    this._activatedExtensions[extensionDescriptions[i].id] = this._createFailedExtension();
                }
                return winjs_base_1.TPromise.as(void 0);
            }
            var greenMap = Object.create(null), red = [];
            for (var i = 0, len = extensionDescriptions.length; i < len; i++) {
                this._handleActivateRequest(extensionDescriptions[i], greenMap, red);
            }
            // Make sure no red is also green
            for (var i = 0, len = red.length; i < len; i++) {
                if (greenMap[red[i].id]) {
                    delete greenMap[red[i].id];
                }
            }
            var green = Object.keys(greenMap).map(function (id) { return greenMap[id]; });
            // console.log('greenExtensions: ', green.map(p => p.id));
            // console.log('redExtensions: ', red.map(p => p.id));
            if (red.length === 0) {
                // Finally reached only leafs!
                return winjs_base_1.TPromise.join(green.map(function (p) { return _this._activateExtension(p); })).then(function (_) { return void 0; });
            }
            return this._activateExtensions(green, recursionLevel + 1).then(function (_) {
                return _this._activateExtensions(red, recursionLevel + 1);
            });
        };
        AbstractExtensionService.prototype._activateExtension = function (extensionDescription) {
            var _this = this;
            if (hasOwnProperty.call(this._activatedExtensions, extensionDescription.id)) {
                return winjs_base_1.TPromise.as(void 0);
            }
            if (hasOwnProperty.call(this._activatingExtensions, extensionDescription.id)) {
                return this._activatingExtensions[extensionDescription.id];
            }
            this._activatingExtensions[extensionDescription.id] = this._actualActivateExtension(extensionDescription).then(null, function (err) {
                _this._showMessage(severity_1.default.Error, nls.localize(3, null, extensionDescription.id, err.message));
                console.error('Activating extension `' + extensionDescription.id + '` failed: ', err.message);
                console.log('Here is the error stack: ', err.stack);
                // Treat the extension as being empty
                return _this._createFailedExtension();
            }).then(function (x) {
                _this._activatedExtensions[extensionDescription.id] = x;
                delete _this._activatingExtensions[extensionDescription.id];
            });
            return this._activatingExtensions[extensionDescription.id];
        };
        return AbstractExtensionService;
    }());
    exports.AbstractExtensionService = AbstractExtensionService;
});

define(__m[43], __M([1,0,4]), function (require, exports, instantiation_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    exports.IRequestService = instantiation_1.createDecorator('requestService');
});

define(__m[44], __M([1,0,5,26,4]), function (require, exports, winjs_base_1, timer_1, instantiation_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    exports.ITelemetryService = instantiation_1.createDecorator('telemetryService');
    exports.NullTelemetryService = {
        serviceId: undefined,
        timedPublicLog: function (name, data) { return timer_1.nullEvent; },
        publicLog: function (eventName, data) { return winjs_base_1.TPromise.as(null); },
        isOptedIn: true,
        getTelemetryInfo: function () {
            return winjs_base_1.TPromise.as({
                instanceId: 'someValue.instanceId',
                sessionId: 'someValue.sessionId',
                machineId: 'someValue.machineId'
            });
        }
    };
    function combinedAppender() {
        var appenders = [];
        for (var _i = 0; _i < arguments.length; _i++) {
            appenders[_i - 0] = arguments[_i];
        }
        return { log: function (e, d) { return appenders.forEach(function (a) { return a.log(e, d); }); } };
    }
    exports.combinedAppender = combinedAppender;
    exports.NullAppender = { log: function () { return null; } };
    // --- util
    function anonymize(input) {
        if (!input) {
            return input;
        }
        var r = '';
        for (var i = 0; i < input.length; i++) {
            var ch = input[i];
            if (ch >= '0' && ch <= '9') {
                r += '0';
                continue;
            }
            if (ch >= 'a' && ch <= 'z') {
                r += 'a';
                continue;
            }
            if (ch >= 'A' && ch <= 'Z') {
                r += 'A';
                continue;
            }
            r += ch;
        }
        return r;
    }
    exports.anonymize = anonymize;
});

define(__m[107], __M([1,0,20,5,84,3,26,24,12,43,44]), function (require, exports, uri_1, winjs_base_1, network_1, strings, Timer, Async, objects, request_1, telemetry_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    /**
     * Simple IRequestService implementation to allow sharing of this service implementation
     * between different layers of the platform.
     */
    var BaseRequestService = (function () {
        function BaseRequestService(contextService, telemetryService) {
            if (telemetryService === void 0) { telemetryService = telemetry_1.NullTelemetryService; }
            this.serviceId = request_1.IRequestService;
            var workspaceUri = null;
            var workspace = contextService.getWorkspace();
            this._serviceMap = workspace || Object.create(null);
            this._telemetryService = telemetryService;
            if (workspace) {
                workspaceUri = strings.rtrim(workspace.resource.toString(), '/') + '/';
            }
            this.computeOrigin(workspaceUri);
        }
        BaseRequestService.prototype.computeOrigin = function (workspaceUri) {
            if (workspaceUri) {
                // Find root server URL from configuration
                this._origin = workspaceUri;
                var urlPath = uri_1.default.parse(this._origin).path;
                if (urlPath && urlPath.length > 0) {
                    this._origin = this._origin.substring(0, this._origin.length - urlPath.length + 1);
                }
                if (!strings.endsWith(this._origin, '/')) {
                    this._origin += '/';
                }
            }
            else {
                this._origin = '/'; // Configuration not provided, fallback to default
            }
        };
        BaseRequestService.prototype.makeCrossOriginRequest = function (options) {
            return null;
        };
        BaseRequestService.prototype.makeRequest = function (options) {
            var timer = Timer.nullEvent;
            var isXhrRequestCORS = false;
            var url = options.url;
            if (!url) {
                throw new Error('IRequestService.makeRequest: Url is required');
            }
            if ((strings.startsWith(url, 'http://') || strings.startsWith(url, 'https://')) && this._origin && !strings.startsWith(url, this._origin)) {
                var coPromise = this.makeCrossOriginRequest(options);
                if (coPromise) {
                    return coPromise;
                }
                isXhrRequestCORS = true;
            }
            var xhrOptions = options;
            var xhrOptionsPromise = winjs_base_1.TPromise.as(undefined);
            if (!isXhrRequestCORS) {
                xhrOptions = this._telemetryService.getTelemetryInfo().then(function (info) {
                    var additionalHeaders = {};
                    additionalHeaders['X-TelemetrySession'] = info.sessionId;
                    additionalHeaders['X-Requested-With'] = 'XMLHttpRequest';
                    xhrOptions.headers = objects.mixin(xhrOptions.headers, additionalHeaders);
                });
            }
            if (options.timeout) {
                xhrOptions.customRequestInitializer = function (xhrRequest) {
                    xhrRequest.timeout = options.timeout;
                };
            }
            return xhrOptionsPromise.then(function () {
                return Async.always(network_1.xhr(xhrOptions), (function (xhr) {
                    if (timer.data) {
                        timer.data.status = xhr.status;
                    }
                    timer.stop();
                }));
            });
        };
        return BaseRequestService;
    }());
    exports.BaseRequestService = BaseRequestService;
});

define(__m[27], __M([1,0,4]), function (require, exports, instantiation_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    exports.IWorkspaceContextService = instantiation_1.createDecorator('contextService');
});

/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Microsoft Corporation. All rights reserved.
 *  Licensed under the MIT License. See License.txt in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
define(__m[98], __M([1,0,34,24,63,47,49,8,81,16,13,20,11,41,43,27,44,19,21,60,39,25,105,88,31,59,53,85,23,35]), function (require, exports) {
    'use strict';
});

define(__m[92], __M([1,0,20,13,27]), function (require, exports, uri_1, paths, workspace_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    /**
     * Simple IWorkspaceContextService implementation to allow sharing of this service implementation
     * between different layers of the platform.
     */
    var BaseWorkspaceContextService = (function () {
        function BaseWorkspaceContextService(workspace, configuration, options) {
            if (options === void 0) { options = {}; }
            this.serviceId = workspace_1.IWorkspaceContextService;
            this.workspace = workspace;
            this.configuration = configuration;
            this.options = options;
        }
        BaseWorkspaceContextService.prototype.getWorkspace = function () {
            return this.workspace;
        };
        BaseWorkspaceContextService.prototype.getConfiguration = function () {
            return this.configuration;
        };
        BaseWorkspaceContextService.prototype.getOptions = function () {
            return this.options;
        };
        BaseWorkspaceContextService.prototype.isInsideWorkspace = function (resource) {
            if (resource && this.workspace) {
                return paths.isEqualOrParent(resource.fsPath, this.workspace.resource.fsPath);
            }
            return false;
        };
        BaseWorkspaceContextService.prototype.toWorkspaceRelativePath = function (resource) {
            if (this.isInsideWorkspace(resource)) {
                return paths.normalize(paths.relative(this.workspace.resource.fsPath, resource.fsPath));
            }
            return null;
        };
        BaseWorkspaceContextService.prototype.toResource = function (workspaceRelativePath) {
            if (typeof workspaceRelativePath === 'string' && this.workspace) {
                return uri_1.default.file(paths.join(this.workspace.resource.fsPath, workspaceRelativePath));
            }
            return null;
        };
        return BaseWorkspaceContextService;
    }());
    exports.BaseWorkspaceContextService = BaseWorkspaceContextService;
});

/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Microsoft Corporation. All rights reserved.
 *  Licensed under the MIT License. See License.txt in the project root for license information.
 *--------------------------------------------------------------------------------------------*/





define(__m[111], __M([1,0,32,5,90,52,104,38,54,93,107,43,44,92,27,103,23,87,37,96,35,98]), function (require, exports, severity_1, winjs_base_1, eventService_1, event_1, abstractExtensionService_1, extensions_1, serviceCollection_1, instantiationService_1, baseRequestService_1, request_1, telemetry_1, baseWorkspaceContextService_1, workspace_1, modeServiceImpl_1, modeService_1, resourceServiceImpl_1, resourceService_1, compatWorkerServiceWorker_1, compatWorkerService_1) {
    'use strict';
    var WorkerExtensionService = (function (_super) {
        __extends(WorkerExtensionService, _super);
        function WorkerExtensionService() {
            _super.call(this, true);
        }
        WorkerExtensionService.prototype._showMessage = function (severity, msg) {
            switch (severity) {
                case severity_1.default.Error:
                    console.error(msg);
                    break;
                case severity_1.default.Warning:
                    console.warn(msg);
                    break;
                case severity_1.default.Info:
                    console.info(msg);
                    break;
                default:
                    console.log(msg);
            }
        };
        WorkerExtensionService.prototype._createFailedExtension = function () {
            throw new Error('unexpected');
        };
        WorkerExtensionService.prototype._actualActivateExtension = function (extensionDescription) {
            throw new Error('unexpected');
        };
        return WorkerExtensionService;
    }(abstractExtensionService_1.AbstractExtensionService));
    var EditorWorkerServer = (function () {
        function EditorWorkerServer() {
        }
        EditorWorkerServer.prototype.initialize = function (mainThread, complete, error, progress, initData) {
            var services = new serviceCollection_1.ServiceCollection();
            var instantiationService = new instantiationService_1.InstantiationService(services);
            var extensionService = new WorkerExtensionService();
            services.set(extensions_1.IExtensionService, extensionService);
            var contextService = new baseWorkspaceContextService_1.BaseWorkspaceContextService(initData.contextService.workspace, initData.contextService.configuration, initData.contextService.options);
            services.set(workspace_1.IWorkspaceContextService, contextService);
            var resourceService = new resourceServiceImpl_1.ResourceService();
            services.set(resourceService_1.IResourceService, resourceService);
            var requestService = new baseRequestService_1.BaseRequestService(contextService, telemetry_1.NullTelemetryService);
            services.set(request_1.IRequestService, requestService);
            services.set(event_1.IEventService, new eventService_1.EventService());
            var modeService = new modeServiceImpl_1.ModeServiceImpl(instantiationService, extensionService);
            services.set(modeService_1.IModeService, modeService);
            this.compatWorkerService = new compatWorkerServiceWorker_1.CompatWorkerServiceWorker(resourceService, modeService, initData.modesRegistryData);
            services.set(compatWorkerService_1.ICompatWorkerService, this.compatWorkerService);
            complete(undefined);
        };
        EditorWorkerServer.prototype.request = function (mainThread, complete, error, progress, data) {
            try {
                winjs_base_1.TPromise.as(this.compatWorkerService.handleMainRequest(data.target, data.methodName, data.args)).then(complete, error);
            }
            catch (err) {
                error(err);
            }
        };
        return EditorWorkerServer;
    }());
    exports.EditorWorkerServer = EditorWorkerServer;
    exports.value = new EditorWorkerServer();
});

}).call(this);
//# sourceMappingURL=workerServer.js.map
