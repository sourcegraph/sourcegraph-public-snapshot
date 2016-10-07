/*!-----------------------------------------------------------
 * Copyright (c) Microsoft Corporation. All rights reserved.
 * Version: 0.5.3(793ede49d53dba79d39e52205f16321278f5183c)
 * Released under the MIT license
 * https://github.com/Microsoft/vscode/blob/master/LICENSE.txt
 *-----------------------------------------------------------*/

(function() {
var __m = ["exports","require","vs/base/common/winjs.base","vs/base/common/platform","vs/base/common/strings","vs/base/common/errors","vs/base/common/types","vs/editor/common/core/range","vs/base/common/map","vs/editor/common/core/position","vs/base/common/lifecycle","vs/nls","vs/nls!vs/base/common/worker/simpleWorker","vs/base/common/uri","vs/base/common/event","vs/base/common/cancellation","vs/editor/common/modes/supports/inplaceReplaceSupport","vs/editor/common/model/wordHelper","vs/editor/common/modes/linkComputer","vs/base/common/diff/diffChange","vs/editor/common/viewModel/prefixSumComputer","vs/editor/common/model/mirrorModel2","vs/nls!vs/base/common/errors","vs/base/common/objects","vs/base/common/diff/diff","vs/base/common/filters","vs/base/common/callbackList","vs/base/common/arrays","vs/editor/common/core/selection","vs/editor/common/diff/diffComputer","vs/editor/common/standalone/standaloneBase","vs/base/common/async","vs/base/common/severity","vs/nls!vs/base/common/keyCodes","vs/base/common/keyCodes","vs/nls!vs/base/common/severity","vs/base/common/worker/simpleWorker","vs/base/common/winjs.base.raw","vs/editor/common/services/editorSimpleWorker"];
var __M = function(deps) {
  var result = [];
  for (var i = 0, len = deps.length; i < len; i++) {
    result[i] = __m[deps[i]];
  }
  return result;
};
define(__m[27], __M([1,0]), function (require, exports) {
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

define(__m[19], __M([1,0]), function (require, exports) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    exports.DifferenceType = {
        Add: 0,
        Remove: 1,
        Change: 2
    };
    /**
     * Represents information about a specific difference between two sequences.
     */
    var DiffChange = (function () {
        /**
         * Constructs a new DiffChange with the given sequence information
         * and content.
         */
        function DiffChange(originalStart, originalLength, modifiedStart, modifiedLength) {
            //Debug.Assert(originalLength > 0 || modifiedLength > 0, "originalLength and modifiedLength cannot both be <= 0");
            this.originalStart = originalStart;
            this.originalLength = originalLength;
            this.modifiedStart = modifiedStart;
            this.modifiedLength = modifiedLength;
        }
        /**
         * The type of difference.
         */
        DiffChange.prototype.getChangeType = function () {
            if (this.originalLength === 0) {
                return exports.DifferenceType.Add;
            }
            else if (this.modifiedLength === 0) {
                return exports.DifferenceType.Remove;
            }
            else {
                return exports.DifferenceType.Change;
            }
        };
        /**
         * The end point (exclusive) of the change in the original sequence.
         */
        DiffChange.prototype.getOriginalEnd = function () {
            return this.originalStart + this.originalLength;
        };
        /**
         * The end point (exclusive) of the change in the modified sequence.
         */
        DiffChange.prototype.getModifiedEnd = function () {
            return this.modifiedStart + this.modifiedLength;
        };
        return DiffChange;
    }());
    exports.DiffChange = DiffChange;
});

define(__m[24], __M([1,0,19]), function (require, exports, diffChange_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    //
    // The code below has been ported from a C# implementation in VS
    //
    var Debug = (function () {
        function Debug() {
        }
        Debug.Assert = function (condition, message) {
            if (!condition) {
                throw new Error(message);
            }
        };
        return Debug;
    }());
    exports.Debug = Debug;
    var MyArray = (function () {
        function MyArray() {
        }
        /**
         * Copies a range of elements from an Array starting at the specified source index and pastes
         * them to another Array starting at the specified destination index. The length and the indexes
         * are specified as 64-bit integers.
         * sourceArray:
         *		The Array that contains the data to copy.
         * sourceIndex:
         *		A 64-bit integer that represents the index in the sourceArray at which copying begins.
         * destinationArray:
         *		The Array that receives the data.
         * destinationIndex:
         *		A 64-bit integer that represents the index in the destinationArray at which storing begins.
         * length:
         *		A 64-bit integer that represents the number of elements to copy.
         */
        MyArray.Copy = function (sourceArray, sourceIndex, destinationArray, destinationIndex, length) {
            for (var i = 0; i < length; i++) {
                destinationArray[destinationIndex + i] = sourceArray[sourceIndex + i];
            }
        };
        return MyArray;
    }());
    exports.MyArray = MyArray;
    //*****************************************************************************
    // LcsDiff.cs
    //
    // An implementation of the difference algorithm described in
    // "An O(ND) Difference Algorithm and its letiations" by Eugene W. Myers
    //
    // Copyright (C) 2008 Microsoft Corporation @minifier_do_not_preserve
    //*****************************************************************************
    // Our total memory usage for storing history is (worst-case):
    // 2 * [(MaxDifferencesHistory + 1) * (MaxDifferencesHistory + 1) - 1] * sizeof(int)
    // 2 * [1448*1448 - 1] * 4 = 16773624 = 16MB
    var MaxDifferencesHistory = 1447;
    //let MaxDifferencesHistory = 100;
    /**
     * A utility class which helps to create the set of DiffChanges from
     * a difference operation. This class accepts original DiffElements and
     * modified DiffElements that are involved in a particular change. The
     * MarktNextChange() method can be called to mark the separation between
     * distinct changes. At the end, the Changes property can be called to retrieve
     * the constructed changes.
     */
    var DiffChangeHelper = (function () {
        /**
         * Constructs a new DiffChangeHelper for the given DiffSequences.
         */
        function DiffChangeHelper() {
            this.m_changes = [];
            this.m_originalStart = Number.MAX_VALUE;
            this.m_modifiedStart = Number.MAX_VALUE;
            this.m_originalCount = 0;
            this.m_modifiedCount = 0;
        }
        /**
         * Marks the beginning of the next change in the set of differences.
         */
        DiffChangeHelper.prototype.MarkNextChange = function () {
            // Only add to the list if there is something to add
            if (this.m_originalCount > 0 || this.m_modifiedCount > 0) {
                // Add the new change to our list
                this.m_changes.push(new diffChange_1.DiffChange(this.m_originalStart, this.m_originalCount, this.m_modifiedStart, this.m_modifiedCount));
            }
            // Reset for the next change
            this.m_originalCount = 0;
            this.m_modifiedCount = 0;
            this.m_originalStart = Number.MAX_VALUE;
            this.m_modifiedStart = Number.MAX_VALUE;
        };
        /**
         * Adds the original element at the given position to the elements
         * affected by the current change. The modified index gives context
         * to the change position with respect to the original sequence.
         * @param originalIndex The index of the original element to add.
         * @param modifiedIndex The index of the modified element that provides corresponding position in the modified sequence.
         */
        DiffChangeHelper.prototype.AddOriginalElement = function (originalIndex, modifiedIndex) {
            // The 'true' start index is the smallest of the ones we've seen
            this.m_originalStart = Math.min(this.m_originalStart, originalIndex);
            this.m_modifiedStart = Math.min(this.m_modifiedStart, modifiedIndex);
            this.m_originalCount++;
        };
        /**
         * Adds the modified element at the given position to the elements
         * affected by the current change. The original index gives context
         * to the change position with respect to the modified sequence.
         * @param originalIndex The index of the original element that provides corresponding position in the original sequence.
         * @param modifiedIndex The index of the modified element to add.
         */
        DiffChangeHelper.prototype.AddModifiedElement = function (originalIndex, modifiedIndex) {
            // The 'true' start index is the smallest of the ones we've seen
            this.m_originalStart = Math.min(this.m_originalStart, originalIndex);
            this.m_modifiedStart = Math.min(this.m_modifiedStart, modifiedIndex);
            this.m_modifiedCount++;
        };
        /**
         * Retrieves all of the changes marked by the class.
         */
        DiffChangeHelper.prototype.getChanges = function () {
            if (this.m_originalCount > 0 || this.m_modifiedCount > 0) {
                // Finish up on whatever is left
                this.MarkNextChange();
            }
            return this.m_changes;
        };
        DiffChangeHelper.prototype.getReverseChanges = function () {
            /// <summary>
            /// Retrieves all of the changes marked by the class in the reverse order
            /// </summary>
            if (this.m_originalCount > 0 || this.m_modifiedCount > 0) {
                // Finish up on whatever is left
                this.MarkNextChange();
            }
            this.m_changes.reverse();
            return this.m_changes;
        };
        return DiffChangeHelper;
    }());
    /**
     * An implementation of the difference algorithm described in
     * "An O(ND) Difference Algorithm and its letiations" by Eugene W. Myers
     */
    var LcsDiff = (function () {
        /**
         * Constructs the DiffFinder
         */
        function LcsDiff(originalSequence, newSequence, continueProcessingPredicate) {
            if (continueProcessingPredicate === void 0) { continueProcessingPredicate = null; }
            this.OriginalSequence = originalSequence;
            this.ModifiedSequence = newSequence;
            this.ContinueProcessingPredicate = continueProcessingPredicate;
            this.m_originalIds = [];
            this.m_modifiedIds = [];
            this.m_forwardHistory = [];
            this.m_reverseHistory = [];
            this.ComputeUniqueIdentifiers();
        }
        LcsDiff.prototype.ComputeUniqueIdentifiers = function () {
            var originalSequenceLength = this.OriginalSequence.getLength();
            var modifiedSequenceLength = this.ModifiedSequence.getLength();
            this.m_originalIds = new Array(originalSequenceLength);
            this.m_modifiedIds = new Array(modifiedSequenceLength);
            // Create a new hash table for unique elements from the original
            // sequence.
            var hashTable = {};
            var currentUniqueId = 1;
            var i;
            // Fill up the hash table for unique elements
            for (i = 0; i < originalSequenceLength; i++) {
                var originalElementHash = this.OriginalSequence.getElementHash(i);
                if (!hashTable.hasOwnProperty(originalElementHash)) {
                    // No entry in the hashtable so this is a new unique element.
                    // Assign the element a new unique identifier and add it to the
                    // hash table
                    this.m_originalIds[i] = currentUniqueId++;
                    hashTable[originalElementHash] = this.m_originalIds[i];
                }
                else {
                    this.m_originalIds[i] = hashTable[originalElementHash];
                }
            }
            // Now match up modified elements
            for (i = 0; i < modifiedSequenceLength; i++) {
                var modifiedElementHash = this.ModifiedSequence.getElementHash(i);
                if (!hashTable.hasOwnProperty(modifiedElementHash)) {
                    this.m_modifiedIds[i] = currentUniqueId++;
                    hashTable[modifiedElementHash] = this.m_modifiedIds[i];
                }
                else {
                    this.m_modifiedIds[i] = hashTable[modifiedElementHash];
                }
            }
        };
        LcsDiff.prototype.ElementsAreEqual = function (originalIndex, newIndex) {
            return this.m_originalIds[originalIndex] === this.m_modifiedIds[newIndex];
        };
        LcsDiff.prototype.ComputeDiff = function () {
            return this._ComputeDiff(0, this.OriginalSequence.getLength() - 1, 0, this.ModifiedSequence.getLength() - 1);
        };
        /**
         * Computes the differences between the original and modified input
         * sequences on the bounded range.
         * @returns An array of the differences between the two input sequences.
         */
        LcsDiff.prototype._ComputeDiff = function (originalStart, originalEnd, modifiedStart, modifiedEnd) {
            var quitEarlyArr = [false];
            return this.ComputeDiffRecursive(originalStart, originalEnd, modifiedStart, modifiedEnd, quitEarlyArr);
        };
        /**
         * Private helper method which computes the differences on the bounded range
         * recursively.
         * @returns An array of the differences between the two input sequences.
         */
        LcsDiff.prototype.ComputeDiffRecursive = function (originalStart, originalEnd, modifiedStart, modifiedEnd, quitEarlyArr) {
            quitEarlyArr[0] = false;
            // Find the start of the differences
            while (originalStart <= originalEnd && modifiedStart <= modifiedEnd && this.ElementsAreEqual(originalStart, modifiedStart)) {
                originalStart++;
                modifiedStart++;
            }
            // Find the end of the differences
            while (originalEnd >= originalStart && modifiedEnd >= modifiedStart && this.ElementsAreEqual(originalEnd, modifiedEnd)) {
                originalEnd--;
                modifiedEnd--;
            }
            // In the special case where we either have all insertions or all deletions or the sequences are identical
            if (originalStart > originalEnd || modifiedStart > modifiedEnd) {
                var changes = void 0;
                if (modifiedStart <= modifiedEnd) {
                    Debug.Assert(originalStart === originalEnd + 1, 'originalStart should only be one more than originalEnd');
                    // All insertions
                    changes = [
                        new diffChange_1.DiffChange(originalStart, 0, modifiedStart, modifiedEnd - modifiedStart + 1)
                    ];
                }
                else if (originalStart <= originalEnd) {
                    Debug.Assert(modifiedStart === modifiedEnd + 1, 'modifiedStart should only be one more than modifiedEnd');
                    // All deletions
                    changes = [
                        new diffChange_1.DiffChange(originalStart, originalEnd - originalStart + 1, modifiedStart, 0)
                    ];
                }
                else {
                    Debug.Assert(originalStart === originalEnd + 1, 'originalStart should only be one more than originalEnd');
                    Debug.Assert(modifiedStart === modifiedEnd + 1, 'modifiedStart should only be one more than modifiedEnd');
                    // Identical sequences - No differences
                    changes = [];
                }
                return changes;
            }
            // This problem can be solved using the Divide-And-Conquer technique.
            var midOriginalArr = [0], midModifiedArr = [0];
            var result = this.ComputeRecursionPoint(originalStart, originalEnd, modifiedStart, modifiedEnd, midOriginalArr, midModifiedArr, quitEarlyArr);
            var midOriginal = midOriginalArr[0];
            var midModified = midModifiedArr[0];
            if (result !== null) {
                // Result is not-null when there was enough memory to compute the changes while
                // searching for the recursion point
                return result;
            }
            else if (!quitEarlyArr[0]) {
                // We can break the problem down recursively by finding the changes in the
                // First Half:   (originalStart, modifiedStart) to (midOriginal, midModified)
                // Second Half:  (midOriginal + 1, minModified + 1) to (originalEnd, modifiedEnd)
                // NOTE: ComputeDiff() is inclusive, therefore the second range starts on the next point
                var leftChanges = this.ComputeDiffRecursive(originalStart, midOriginal, modifiedStart, midModified, quitEarlyArr);
                var rightChanges = [];
                if (!quitEarlyArr[0]) {
                    rightChanges = this.ComputeDiffRecursive(midOriginal + 1, originalEnd, midModified + 1, modifiedEnd, quitEarlyArr);
                }
                else {
                    // We did't have time to finish the first half, so we don't have time to compute this half.
                    // Consider the entire rest of the sequence different.
                    rightChanges = [
                        new diffChange_1.DiffChange(midOriginal + 1, originalEnd - (midOriginal + 1) + 1, midModified + 1, modifiedEnd - (midModified + 1) + 1)
                    ];
                }
                return this.ConcatenateChanges(leftChanges, rightChanges);
            }
            // If we hit here, we quit early, and so can't return anything meaningful
            return [
                new diffChange_1.DiffChange(originalStart, originalEnd - originalStart + 1, modifiedStart, modifiedEnd - modifiedStart + 1)
            ];
        };
        LcsDiff.prototype.WALKTRACE = function (diagonalForwardBase, diagonalForwardStart, diagonalForwardEnd, diagonalForwardOffset, diagonalReverseBase, diagonalReverseStart, diagonalReverseEnd, diagonalReverseOffset, forwardPoints, reversePoints, originalIndex, originalEnd, midOriginalArr, modifiedIndex, modifiedEnd, midModifiedArr, deltaIsEven, quitEarlyArr) {
            var forwardChanges = null, reverseChanges = null;
            // First, walk backward through the forward diagonals history
            var changeHelper = new DiffChangeHelper();
            var diagonalMin = diagonalForwardStart;
            var diagonalMax = diagonalForwardEnd;
            var diagonalRelative = (midOriginalArr[0] - midModifiedArr[0]) - diagonalForwardOffset;
            var lastOriginalIndex = Number.MIN_VALUE;
            var historyIndex = this.m_forwardHistory.length - 1;
            var diagonal;
            do {
                // Get the diagonal index from the relative diagonal number
                diagonal = diagonalRelative + diagonalForwardBase;
                // Figure out where we came from
                if (diagonal === diagonalMin || (diagonal < diagonalMax && forwardPoints[diagonal - 1] < forwardPoints[diagonal + 1])) {
                    // Vertical line (the element is an insert)
                    originalIndex = forwardPoints[diagonal + 1];
                    modifiedIndex = originalIndex - diagonalRelative - diagonalForwardOffset;
                    if (originalIndex < lastOriginalIndex) {
                        changeHelper.MarkNextChange();
                    }
                    lastOriginalIndex = originalIndex;
                    changeHelper.AddModifiedElement(originalIndex + 1, modifiedIndex);
                    diagonalRelative = (diagonal + 1) - diagonalForwardBase; //Setup for the next iteration
                }
                else {
                    // Horizontal line (the element is a deletion)
                    originalIndex = forwardPoints[diagonal - 1] + 1;
                    modifiedIndex = originalIndex - diagonalRelative - diagonalForwardOffset;
                    if (originalIndex < lastOriginalIndex) {
                        changeHelper.MarkNextChange();
                    }
                    lastOriginalIndex = originalIndex - 1;
                    changeHelper.AddOriginalElement(originalIndex, modifiedIndex + 1);
                    diagonalRelative = (diagonal - 1) - diagonalForwardBase; //Setup for the next iteration
                }
                if (historyIndex >= 0) {
                    forwardPoints = this.m_forwardHistory[historyIndex];
                    diagonalForwardBase = forwardPoints[0]; //We stored this in the first spot
                    diagonalMin = 1;
                    diagonalMax = forwardPoints.length - 1;
                }
            } while (--historyIndex >= -1);
            // Ironically, we get the forward changes as the reverse of the
            // order we added them since we technically added them backwards
            forwardChanges = changeHelper.getReverseChanges();
            if (quitEarlyArr[0]) {
                // TODO: Calculate a partial from the reverse diagonals.
                //       For now, just assume everything after the midOriginal/midModified point is a diff
                var originalStartPoint = midOriginalArr[0] + 1;
                var modifiedStartPoint = midModifiedArr[0] + 1;
                if (forwardChanges !== null && forwardChanges.length > 0) {
                    var lastForwardChange = forwardChanges[forwardChanges.length - 1];
                    originalStartPoint = Math.max(originalStartPoint, lastForwardChange.getOriginalEnd());
                    modifiedStartPoint = Math.max(modifiedStartPoint, lastForwardChange.getModifiedEnd());
                }
                reverseChanges = [
                    new diffChange_1.DiffChange(originalStartPoint, originalEnd - originalStartPoint + 1, modifiedStartPoint, modifiedEnd - modifiedStartPoint + 1)
                ];
            }
            else {
                // Now walk backward through the reverse diagonals history
                changeHelper = new DiffChangeHelper();
                diagonalMin = diagonalReverseStart;
                diagonalMax = diagonalReverseEnd;
                diagonalRelative = (midOriginalArr[0] - midModifiedArr[0]) - diagonalReverseOffset;
                lastOriginalIndex = Number.MAX_VALUE;
                historyIndex = (deltaIsEven) ? this.m_reverseHistory.length - 1 : this.m_reverseHistory.length - 2;
                do {
                    // Get the diagonal index from the relative diagonal number
                    diagonal = diagonalRelative + diagonalReverseBase;
                    // Figure out where we came from
                    if (diagonal === diagonalMin || (diagonal < diagonalMax && reversePoints[diagonal - 1] >= reversePoints[diagonal + 1])) {
                        // Horizontal line (the element is a deletion))
                        originalIndex = reversePoints[diagonal + 1] - 1;
                        modifiedIndex = originalIndex - diagonalRelative - diagonalReverseOffset;
                        if (originalIndex > lastOriginalIndex) {
                            changeHelper.MarkNextChange();
                        }
                        lastOriginalIndex = originalIndex + 1;
                        changeHelper.AddOriginalElement(originalIndex + 1, modifiedIndex + 1);
                        diagonalRelative = (diagonal + 1) - diagonalReverseBase; //Setup for the next iteration
                    }
                    else {
                        // Vertical line (the element is an insertion)
                        originalIndex = reversePoints[diagonal - 1];
                        modifiedIndex = originalIndex - diagonalRelative - diagonalReverseOffset;
                        if (originalIndex > lastOriginalIndex) {
                            changeHelper.MarkNextChange();
                        }
                        lastOriginalIndex = originalIndex;
                        changeHelper.AddModifiedElement(originalIndex + 1, modifiedIndex + 1);
                        diagonalRelative = (diagonal - 1) - diagonalReverseBase; //Setup for the next iteration
                    }
                    if (historyIndex >= 0) {
                        reversePoints = this.m_reverseHistory[historyIndex];
                        diagonalReverseBase = reversePoints[0]; //We stored this in the first spot
                        diagonalMin = 1;
                        diagonalMax = reversePoints.length - 1;
                    }
                } while (--historyIndex >= -1);
                // There are cases where the reverse history will find diffs that
                // are correct, but not intuitive, so we need shift them.
                reverseChanges = changeHelper.getChanges();
            }
            return this.ConcatenateChanges(forwardChanges, reverseChanges);
        };
        /**
         * Given the range to compute the diff on, this method finds the point:
         * (midOriginal, midModified)
         * that exists in the middle of the LCS of the two sequences and
         * is the point at which the LCS problem may be broken down recursively.
         * This method will try to keep the LCS trace in memory. If the LCS recursion
         * point is calculated and the full trace is available in memory, then this method
         * will return the change list.
         * @param originalStart The start bound of the original sequence range
         * @param originalEnd The end bound of the original sequence range
         * @param modifiedStart The start bound of the modified sequence range
         * @param modifiedEnd The end bound of the modified sequence range
         * @param midOriginal The middle point of the original sequence range
         * @param midModified The middle point of the modified sequence range
         * @returns The diff changes, if available, otherwise null
         */
        LcsDiff.prototype.ComputeRecursionPoint = function (originalStart, originalEnd, modifiedStart, modifiedEnd, midOriginalArr, midModifiedArr, quitEarlyArr) {
            var originalIndex, modifiedIndex;
            var diagonalForwardStart = 0, diagonalForwardEnd = 0;
            var diagonalReverseStart = 0, diagonalReverseEnd = 0;
            var numDifferences;
            // To traverse the edit graph and produce the proper LCS, our actual
            // start position is just outside the given boundary
            originalStart--;
            modifiedStart--;
            // We set these up to make the compiler happy, but they will
            // be replaced before we return with the actual recursion point
            midOriginalArr[0] = 0;
            midModifiedArr[0] = 0;
            // Clear out the history
            this.m_forwardHistory = [];
            this.m_reverseHistory = [];
            // Each cell in the two arrays corresponds to a diagonal in the edit graph.
            // The integer value in the cell represents the originalIndex of the furthest
            // reaching point found so far that ends in that diagonal.
            // The modifiedIndex can be computed mathematically from the originalIndex and the diagonal number.
            var maxDifferences = (originalEnd - originalStart) + (modifiedEnd - modifiedStart);
            var numDiagonals = maxDifferences + 1;
            var forwardPoints = new Array(numDiagonals);
            var reversePoints = new Array(numDiagonals);
            // diagonalForwardBase: Index into forwardPoints of the diagonal which passes through (originalStart, modifiedStart)
            // diagonalReverseBase: Index into reversePoints of the diagonal which passes through (originalEnd, modifiedEnd)
            var diagonalForwardBase = (modifiedEnd - modifiedStart);
            var diagonalReverseBase = (originalEnd - originalStart);
            // diagonalForwardOffset: Geometric offset which allows modifiedIndex to be computed from originalIndex and the
            //    diagonal number (relative to diagonalForwardBase)
            // diagonalReverseOffset: Geometric offset which allows modifiedIndex to be computed from originalIndex and the
            //    diagonal number (relative to diagonalReverseBase)
            var diagonalForwardOffset = (originalStart - modifiedStart);
            var diagonalReverseOffset = (originalEnd - modifiedEnd);
            // delta: The difference between the end diagonal and the start diagonal. This is used to relate diagonal numbers
            //   relative to the start diagonal with diagonal numbers relative to the end diagonal.
            // The Even/Oddn-ness of this delta is important for determining when we should check for overlap
            var delta = diagonalReverseBase - diagonalForwardBase;
            var deltaIsEven = (delta % 2 === 0);
            // Here we set up the start and end points as the furthest points found so far
            // in both the forward and reverse directions, respectively
            forwardPoints[diagonalForwardBase] = originalStart;
            reversePoints[diagonalReverseBase] = originalEnd;
            // Remember if we quit early, and thus need to do a best-effort result instead of a real result.
            quitEarlyArr[0] = false;
            // A couple of points:
            // --With this method, we iterate on the number of differences between the two sequences.
            //   The more differences there actually are, the longer this will take.
            // --Also, as the number of differences increases, we have to search on diagonals further
            //   away from the reference diagonal (which is diagonalForwardBase for forward, diagonalReverseBase for reverse).
            // --We extend on even diagonals (relative to the reference diagonal) only when numDifferences
            //   is even and odd diagonals only when numDifferences is odd.
            var diagonal, tempOriginalIndex;
            for (numDifferences = 1; numDifferences <= (maxDifferences / 2) + 1; numDifferences++) {
                var furthestOriginalIndex = 0;
                var furthestModifiedIndex = 0;
                // Run the algorithm in the forward direction
                diagonalForwardStart = this.ClipDiagonalBound(diagonalForwardBase - numDifferences, numDifferences, diagonalForwardBase, numDiagonals);
                diagonalForwardEnd = this.ClipDiagonalBound(diagonalForwardBase + numDifferences, numDifferences, diagonalForwardBase, numDiagonals);
                for (diagonal = diagonalForwardStart; diagonal <= diagonalForwardEnd; diagonal += 2) {
                    // STEP 1: We extend the furthest reaching point in the present diagonal
                    // by looking at the diagonals above and below and picking the one whose point
                    // is further away from the start point (originalStart, modifiedStart)
                    if (diagonal === diagonalForwardStart || (diagonal < diagonalForwardEnd && forwardPoints[diagonal - 1] < forwardPoints[diagonal + 1])) {
                        originalIndex = forwardPoints[diagonal + 1];
                    }
                    else {
                        originalIndex = forwardPoints[diagonal - 1] + 1;
                    }
                    modifiedIndex = originalIndex - (diagonal - diagonalForwardBase) - diagonalForwardOffset;
                    // Save the current originalIndex so we can test for false overlap in step 3
                    tempOriginalIndex = originalIndex;
                    // STEP 2: We can continue to extend the furthest reaching point in the present diagonal
                    // so long as the elements are equal.
                    while (originalIndex < originalEnd && modifiedIndex < modifiedEnd && this.ElementsAreEqual(originalIndex + 1, modifiedIndex + 1)) {
                        originalIndex++;
                        modifiedIndex++;
                    }
                    forwardPoints[diagonal] = originalIndex;
                    if (originalIndex + modifiedIndex > furthestOriginalIndex + furthestModifiedIndex) {
                        furthestOriginalIndex = originalIndex;
                        furthestModifiedIndex = modifiedIndex;
                    }
                    // STEP 3: If delta is odd (overlap first happens on forward when delta is odd)
                    // and diagonal is in the range of reverse diagonals computed for numDifferences-1
                    // (the previous iteration; we haven't computed reverse diagonals for numDifferences yet)
                    // then check for overlap.
                    if (!deltaIsEven && Math.abs(diagonal - diagonalReverseBase) <= (numDifferences - 1)) {
                        if (originalIndex >= reversePoints[diagonal]) {
                            midOriginalArr[0] = originalIndex;
                            midModifiedArr[0] = modifiedIndex;
                            if (tempOriginalIndex <= reversePoints[diagonal] && MaxDifferencesHistory > 0 && numDifferences <= (MaxDifferencesHistory + 1)) {
                                // BINGO! We overlapped, and we have the full trace in memory!
                                return this.WALKTRACE(diagonalForwardBase, diagonalForwardStart, diagonalForwardEnd, diagonalForwardOffset, diagonalReverseBase, diagonalReverseStart, diagonalReverseEnd, diagonalReverseOffset, forwardPoints, reversePoints, originalIndex, originalEnd, midOriginalArr, modifiedIndex, modifiedEnd, midModifiedArr, deltaIsEven, quitEarlyArr);
                            }
                            else {
                                // Either false overlap, or we didn't have enough memory for the full trace
                                // Just return the recursion point
                                return null;
                            }
                        }
                    }
                }
                // Check to see if we should be quitting early, before moving on to the next iteration.
                var matchLengthOfLongest = ((furthestOriginalIndex - originalStart) + (furthestModifiedIndex - modifiedStart) - numDifferences) / 2;
                if (this.ContinueProcessingPredicate !== null && !this.ContinueProcessingPredicate(furthestOriginalIndex, this.OriginalSequence, matchLengthOfLongest)) {
                    // We can't finish, so skip ahead to generating a result from what we have.
                    quitEarlyArr[0] = true;
                    // Use the furthest distance we got in the forward direction.
                    midOriginalArr[0] = furthestOriginalIndex;
                    midModifiedArr[0] = furthestModifiedIndex;
                    if (matchLengthOfLongest > 0 && MaxDifferencesHistory > 0 && numDifferences <= (MaxDifferencesHistory + 1)) {
                        // Enough of the history is in memory to walk it backwards
                        return this.WALKTRACE(diagonalForwardBase, diagonalForwardStart, diagonalForwardEnd, diagonalForwardOffset, diagonalReverseBase, diagonalReverseStart, diagonalReverseEnd, diagonalReverseOffset, forwardPoints, reversePoints, originalIndex, originalEnd, midOriginalArr, modifiedIndex, modifiedEnd, midModifiedArr, deltaIsEven, quitEarlyArr);
                    }
                    else {
                        // We didn't actually remember enough of the history.
                        //Since we are quiting the diff early, we need to shift back the originalStart and modified start
                        //back into the boundary limits since we decremented their value above beyond the boundary limit.
                        originalStart++;
                        modifiedStart++;
                        return [
                            new diffChange_1.DiffChange(originalStart, originalEnd - originalStart + 1, modifiedStart, modifiedEnd - modifiedStart + 1)
                        ];
                    }
                }
                // Run the algorithm in the reverse direction
                diagonalReverseStart = this.ClipDiagonalBound(diagonalReverseBase - numDifferences, numDifferences, diagonalReverseBase, numDiagonals);
                diagonalReverseEnd = this.ClipDiagonalBound(diagonalReverseBase + numDifferences, numDifferences, diagonalReverseBase, numDiagonals);
                for (diagonal = diagonalReverseStart; diagonal <= diagonalReverseEnd; diagonal += 2) {
                    // STEP 1: We extend the furthest reaching point in the present diagonal
                    // by looking at the diagonals above and below and picking the one whose point
                    // is further away from the start point (originalEnd, modifiedEnd)
                    if (diagonal === diagonalReverseStart || (diagonal < diagonalReverseEnd && reversePoints[diagonal - 1] >= reversePoints[diagonal + 1])) {
                        originalIndex = reversePoints[diagonal + 1] - 1;
                    }
                    else {
                        originalIndex = reversePoints[diagonal - 1];
                    }
                    modifiedIndex = originalIndex - (diagonal - diagonalReverseBase) - diagonalReverseOffset;
                    // Save the current originalIndex so we can test for false overlap
                    tempOriginalIndex = originalIndex;
                    // STEP 2: We can continue to extend the furthest reaching point in the present diagonal
                    // as long as the elements are equal.
                    while (originalIndex > originalStart && modifiedIndex > modifiedStart && this.ElementsAreEqual(originalIndex, modifiedIndex)) {
                        originalIndex--;
                        modifiedIndex--;
                    }
                    reversePoints[diagonal] = originalIndex;
                    // STEP 4: If delta is even (overlap first happens on reverse when delta is even)
                    // and diagonal is in the range of forward diagonals computed for numDifferences
                    // then check for overlap.
                    if (deltaIsEven && Math.abs(diagonal - diagonalForwardBase) <= numDifferences) {
                        if (originalIndex <= forwardPoints[diagonal]) {
                            midOriginalArr[0] = originalIndex;
                            midModifiedArr[0] = modifiedIndex;
                            if (tempOriginalIndex >= forwardPoints[diagonal] && MaxDifferencesHistory > 0 && numDifferences <= (MaxDifferencesHistory + 1)) {
                                // BINGO! We overlapped, and we have the full trace in memory!
                                return this.WALKTRACE(diagonalForwardBase, diagonalForwardStart, diagonalForwardEnd, diagonalForwardOffset, diagonalReverseBase, diagonalReverseStart, diagonalReverseEnd, diagonalReverseOffset, forwardPoints, reversePoints, originalIndex, originalEnd, midOriginalArr, modifiedIndex, modifiedEnd, midModifiedArr, deltaIsEven, quitEarlyArr);
                            }
                            else {
                                // Either false overlap, or we didn't have enough memory for the full trace
                                // Just return the recursion point
                                return null;
                            }
                        }
                    }
                }
                // Save current vectors to history before the next iteration
                if (numDifferences <= MaxDifferencesHistory) {
                    // We are allocating space for one extra int, which we fill with
                    // the index of the diagonal base index
                    var temp = new Array(diagonalForwardEnd - diagonalForwardStart + 2);
                    temp[0] = diagonalForwardBase - diagonalForwardStart + 1;
                    MyArray.Copy(forwardPoints, diagonalForwardStart, temp, 1, diagonalForwardEnd - diagonalForwardStart + 1);
                    this.m_forwardHistory.push(temp);
                    temp = new Array(diagonalReverseEnd - diagonalReverseStart + 2);
                    temp[0] = diagonalReverseBase - diagonalReverseStart + 1;
                    MyArray.Copy(reversePoints, diagonalReverseStart, temp, 1, diagonalReverseEnd - diagonalReverseStart + 1);
                    this.m_reverseHistory.push(temp);
                }
            }
            // If we got here, then we have the full trace in history. We just have to convert it to a change list
            // NOTE: This part is a bit messy
            return this.WALKTRACE(diagonalForwardBase, diagonalForwardStart, diagonalForwardEnd, diagonalForwardOffset, diagonalReverseBase, diagonalReverseStart, diagonalReverseEnd, diagonalReverseOffset, forwardPoints, reversePoints, originalIndex, originalEnd, midOriginalArr, modifiedIndex, modifiedEnd, midModifiedArr, deltaIsEven, quitEarlyArr);
        };
        /**
         * Concatenates the two input DiffChange lists and returns the resulting
         * list.
         * @param The left changes
         * @param The right changes
         * @returns The concatenated list
         */
        LcsDiff.prototype.ConcatenateChanges = function (left, right) {
            var mergedChangeArr = [];
            var result = null;
            if (left.length === 0 || right.length === 0) {
                return (right.length > 0) ? right : left;
            }
            else if (this.ChangesOverlap(left[left.length - 1], right[0], mergedChangeArr)) {
                // Since we break the problem down recursively, it is possible that we
                // might recurse in the middle of a change thereby splitting it into
                // two changes. Here in the combining stage, we detect and fuse those
                // changes back together
                result = new Array(left.length + right.length - 1);
                MyArray.Copy(left, 0, result, 0, left.length - 1);
                result[left.length - 1] = mergedChangeArr[0];
                MyArray.Copy(right, 1, result, left.length, right.length - 1);
                return result;
            }
            else {
                result = new Array(left.length + right.length);
                MyArray.Copy(left, 0, result, 0, left.length);
                MyArray.Copy(right, 0, result, left.length, right.length);
                return result;
            }
        };
        /**
         * Returns true if the two changes overlap and can be merged into a single
         * change
         * @param left The left change
         * @param right The right change
         * @param mergedChange The merged change if the two overlap, null otherwise
         * @returns True if the two changes overlap
         */
        LcsDiff.prototype.ChangesOverlap = function (left, right, mergedChangeArr) {
            Debug.Assert(left.originalStart <= right.originalStart, 'Left change is not less than or equal to right change');
            Debug.Assert(left.modifiedStart <= right.modifiedStart, 'Left change is not less than or equal to right change');
            if (left.originalStart + left.originalLength >= right.originalStart || left.modifiedStart + left.modifiedLength >= right.modifiedStart) {
                var originalStart = left.originalStart;
                var originalLength = left.originalLength;
                var modifiedStart = left.modifiedStart;
                var modifiedLength = left.modifiedLength;
                if (left.originalStart + left.originalLength >= right.originalStart) {
                    originalLength = right.originalStart + right.originalLength - left.originalStart;
                }
                if (left.modifiedStart + left.modifiedLength >= right.modifiedStart) {
                    modifiedLength = right.modifiedStart + right.modifiedLength - left.modifiedStart;
                }
                mergedChangeArr[0] = new diffChange_1.DiffChange(originalStart, originalLength, modifiedStart, modifiedLength);
                return true;
            }
            else {
                mergedChangeArr[0] = null;
                return false;
            }
        };
        /**
         * Helper method used to clip a diagonal index to the range of valid
         * diagonals. This also decides whether or not the diagonal index,
         * if it exceeds the boundary, should be clipped to the boundary or clipped
         * one inside the boundary depending on the Even/Odd status of the boundary
         * and numDifferences.
         * @param diagonal The index of the diagonal to clip.
         * @param numDifferences The current number of differences being iterated upon.
         * @param diagonalBaseIndex The base reference diagonal.
         * @param numDiagonals The total number of diagonals.
         * @returns The clipped diagonal index.
         */
        LcsDiff.prototype.ClipDiagonalBound = function (diagonal, numDifferences, diagonalBaseIndex, numDiagonals) {
            if (diagonal >= 0 && diagonal < numDiagonals) {
                // Nothing to clip, its in range
                return diagonal;
            }
            // diagonalsBelow: The number of diagonals below the reference diagonal
            // diagonalsAbove: The number of diagonals above the reference diagonal
            var diagonalsBelow = diagonalBaseIndex;
            var diagonalsAbove = numDiagonals - diagonalBaseIndex - 1;
            var diffEven = (numDifferences % 2 === 0);
            if (diagonal < 0) {
                var lowerBoundEven = (diagonalsBelow % 2 === 0);
                return (diffEven === lowerBoundEven) ? 0 : 1;
            }
            else {
                var upperBoundEven = (diagonalsAbove % 2 === 0);
                return (diffEven === upperBoundEven) ? numDiagonals - 1 : numDiagonals - 2;
            }
        };
        return LcsDiff;
    }());
    exports.LcsDiff = LcsDiff;
});

/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Microsoft Corporation. All rights reserved.
 *  Licensed under the MIT License. See License.txt in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
define(__m[8], __M([1,0]), function (require, exports) {
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

define(__m[3], __M([1,0]), function (require, exports) {
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

define(__m[4], __M([1,0,8]), function (require, exports, map_1) {
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

define(__m[25], __M([1,0,4,8]), function (require, exports, strings, map_1) {
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

define(__m[6], __M([1,0]), function (require, exports) {
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






define(__m[10], __M([1,0,6]), function (require, exports, types_1) {
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

define(__m[23], __M([1,0,6]), function (require, exports, Types) {
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

define(__m[13], __M([1,0,3]), function (require, exports, platform) {
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
define(__m[9], __M([1,0]), function (require, exports) {
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
define(__m[7], __M([1,0,9]), function (require, exports, position_1) {
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






define(__m[28], __M([1,0,7]), function (require, exports, range_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    /**
     * The direction of a selection.
     */
    (function (SelectionDirection) {
        /**
         * The selection starts above where it ends.
         */
        SelectionDirection[SelectionDirection["LTR"] = 0] = "LTR";
        /**
         * The selection starts below where it ends.
         */
        SelectionDirection[SelectionDirection["RTL"] = 1] = "RTL";
    })(exports.SelectionDirection || (exports.SelectionDirection = {}));
    var SelectionDirection = exports.SelectionDirection;
    /**
     * A selection in the editor.
     * The selection is a range that has an orientation.
     */
    var Selection = (function (_super) {
        __extends(Selection, _super);
        function Selection(selectionStartLineNumber, selectionStartColumn, positionLineNumber, positionColumn) {
            _super.call(this, selectionStartLineNumber, selectionStartColumn, positionLineNumber, positionColumn);
            this.selectionStartLineNumber = selectionStartLineNumber;
            this.selectionStartColumn = selectionStartColumn;
            this.positionLineNumber = positionLineNumber;
            this.positionColumn = positionColumn;
        }
        /**
         * Clone this selection.
         */
        Selection.prototype.clone = function () {
            return new Selection(this.selectionStartLineNumber, this.selectionStartColumn, this.positionLineNumber, this.positionColumn);
        };
        /**
         * Transform to a human-readable representation.
         */
        Selection.prototype.toString = function () {
            return '[' + this.selectionStartLineNumber + ',' + this.selectionStartColumn + ' -> ' + this.positionLineNumber + ',' + this.positionColumn + ']';
        };
        /**
         * Test if equals other selection.
         */
        Selection.prototype.equalsSelection = function (other) {
            return (Selection.selectionsEqual(this, other));
        };
        /**
         * Test if the two selections are equal.
         */
        Selection.selectionsEqual = function (a, b) {
            return (a.selectionStartLineNumber === b.selectionStartLineNumber &&
                a.selectionStartColumn === b.selectionStartColumn &&
                a.positionLineNumber === b.positionLineNumber &&
                a.positionColumn === b.positionColumn);
        };
        /**
         * Get directions (LTR or RTL).
         */
        Selection.prototype.getDirection = function () {
            if (this.selectionStartLineNumber === this.startLineNumber && this.selectionStartColumn === this.startColumn) {
                return SelectionDirection.LTR;
            }
            return SelectionDirection.RTL;
        };
        /**
         * Create a new selection with a different `positionLineNumber` and `positionColumn`.
         */
        Selection.prototype.setEndPosition = function (endLineNumber, endColumn) {
            if (this.getDirection() === SelectionDirection.LTR) {
                return new Selection(this.startLineNumber, this.startColumn, endLineNumber, endColumn);
            }
            return new Selection(endLineNumber, endColumn, this.startLineNumber, this.startColumn);
        };
        /**
         * Create a new selection with a different `selectionStartLineNumber` and `selectionStartColumn`.
         */
        Selection.prototype.setStartPosition = function (startLineNumber, startColumn) {
            if (this.getDirection() === SelectionDirection.LTR) {
                return new Selection(startLineNumber, startColumn, this.endLineNumber, this.endColumn);
            }
            return new Selection(this.endLineNumber, this.endColumn, startLineNumber, startColumn);
        };
        // ----
        /**
         * Create a `Selection` from an `ISelection`.
         */
        Selection.liftSelection = function (sel) {
            return new Selection(sel.selectionStartLineNumber, sel.selectionStartColumn, sel.positionLineNumber, sel.positionColumn);
        };
        /**
         * `a` equals `b`.
         */
        Selection.selectionsArrEqual = function (a, b) {
            if (a && !b || !a && b) {
                return false;
            }
            if (!a && !b) {
                return true;
            }
            if (a.length !== b.length) {
                return false;
            }
            for (var i = 0, len = a.length; i < len; i++) {
                if (!this.selectionsEqual(a[i], b[i])) {
                    return false;
                }
            }
            return true;
        };
        /**
         * Test if `obj` is an `ISelection`.
         */
        Selection.isISelection = function (obj) {
            return (obj
                && (typeof obj.selectionStartLineNumber === 'number')
                && (typeof obj.selectionStartColumn === 'number')
                && (typeof obj.positionLineNumber === 'number')
                && (typeof obj.positionColumn === 'number'));
        };
        /**
         * Create with a direction.
         */
        Selection.createWithDirection = function (startLineNumber, startColumn, endLineNumber, endColumn, direction) {
            if (direction === SelectionDirection.LTR) {
                return new Selection(startLineNumber, startColumn, endLineNumber, endColumn);
            }
            return new Selection(endLineNumber, endColumn, startLineNumber, startColumn);
        };
        return Selection;
    }(range_1.Range));
    exports.Selection = Selection;
});






define(__m[29], __M([1,0,24,4]), function (require, exports, diff_1, strings) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var MAXIMUM_RUN_TIME = 5000; // 5 seconds
    var MINIMUM_MATCHING_CHARACTER_LENGTH = 3;
    function computeDiff(originalSequence, modifiedSequence, continueProcessingPredicate) {
        var diffAlgo = new diff_1.LcsDiff(originalSequence, modifiedSequence, continueProcessingPredicate);
        return diffAlgo.ComputeDiff();
    }
    var MarkerSequence = (function () {
        function MarkerSequence(buffer, startMarkers, endMarkers) {
            this.buffer = buffer;
            this.startMarkers = startMarkers;
            this.endMarkers = endMarkers;
        }
        MarkerSequence.prototype.equals = function (other) {
            if (!(other instanceof MarkerSequence)) {
                return false;
            }
            var otherMarkerSequence = other;
            if (this.getLength() !== otherMarkerSequence.getLength()) {
                return false;
            }
            for (var i = 0, len = this.getLength(); i < len; i++) {
                var myElement = this.getElementHash(i);
                var otherElement = otherMarkerSequence.getElementHash(i);
                if (myElement !== otherElement) {
                    return false;
                }
            }
            return true;
        };
        MarkerSequence.prototype.getLength = function () {
            return this.startMarkers.length;
        };
        MarkerSequence.prototype.getElementHash = function (i) {
            return this.buffer.substring(this.startMarkers[i].offset, this.endMarkers[i].offset);
        };
        MarkerSequence.prototype.getStartLineNumber = function (i) {
            if (i === this.startMarkers.length) {
                // This is the special case where a change happened after the last marker
                return this.startMarkers[i - 1].lineNumber + 1;
            }
            return this.startMarkers[i].lineNumber;
        };
        MarkerSequence.prototype.getStartColumn = function (i) {
            return this.startMarkers[i].column;
        };
        MarkerSequence.prototype.getEndLineNumber = function (i) {
            return this.endMarkers[i].lineNumber;
        };
        MarkerSequence.prototype.getEndColumn = function (i) {
            return this.endMarkers[i].column;
        };
        return MarkerSequence;
    }());
    var LineMarkerSequence = (function (_super) {
        __extends(LineMarkerSequence, _super);
        function LineMarkerSequence(lines, shouldIgnoreTrimWhitespace) {
            var i, length, pos;
            var buffer = '';
            var startMarkers = [], endMarkers = [], startColumn, endColumn;
            for (pos = 0, i = 0, length = lines.length; i < length; i++) {
                buffer += lines[i];
                startColumn = 1;
                endColumn = lines[i].length + 1;
                if (shouldIgnoreTrimWhitespace) {
                    startColumn = LineMarkerSequence._getFirstNonBlankColumn(lines[i], 1);
                    endColumn = LineMarkerSequence._getLastNonBlankColumn(lines[i], 1);
                }
                startMarkers.push({
                    offset: pos + startColumn - 1,
                    lineNumber: i + 1,
                    column: startColumn
                });
                endMarkers.push({
                    offset: pos + endColumn - 1,
                    lineNumber: i + 1,
                    column: endColumn
                });
                pos += lines[i].length;
            }
            _super.call(this, buffer, startMarkers, endMarkers);
        }
        LineMarkerSequence._getFirstNonBlankColumn = function (txt, defaultValue) {
            var r = strings.firstNonWhitespaceIndex(txt);
            if (r === -1) {
                return defaultValue;
            }
            return r + 1;
        };
        LineMarkerSequence._getLastNonBlankColumn = function (txt, defaultValue) {
            var r = strings.lastNonWhitespaceIndex(txt);
            if (r === -1) {
                return defaultValue;
            }
            return r + 2;
        };
        LineMarkerSequence.prototype.getCharSequence = function (startIndex, endIndex) {
            var startMarkers = [], endMarkers = [], index, i, startMarker, endMarker;
            for (index = startIndex; index <= endIndex; index++) {
                startMarker = this.startMarkers[index];
                endMarker = this.endMarkers[index];
                for (i = startMarker.offset; i < endMarker.offset; i++) {
                    startMarkers.push({
                        offset: i,
                        lineNumber: startMarker.lineNumber,
                        column: startMarker.column + (i - startMarker.offset)
                    });
                    endMarkers.push({
                        offset: i + 1,
                        lineNumber: startMarker.lineNumber,
                        column: startMarker.column + (i - startMarker.offset) + 1
                    });
                }
            }
            return new MarkerSequence(this.buffer, startMarkers, endMarkers);
        };
        return LineMarkerSequence;
    }(MarkerSequence));
    var CharChange = (function () {
        function CharChange(diffChange, originalCharSequence, modifiedCharSequence) {
            if (diffChange.originalLength === 0) {
                this.originalStartLineNumber = 0;
                this.originalStartColumn = 0;
                this.originalEndLineNumber = 0;
                this.originalEndColumn = 0;
            }
            else {
                this.originalStartLineNumber = originalCharSequence.getStartLineNumber(diffChange.originalStart);
                this.originalStartColumn = originalCharSequence.getStartColumn(diffChange.originalStart);
                this.originalEndLineNumber = originalCharSequence.getEndLineNumber(diffChange.originalStart + diffChange.originalLength - 1);
                this.originalEndColumn = originalCharSequence.getEndColumn(diffChange.originalStart + diffChange.originalLength - 1);
            }
            if (diffChange.modifiedLength === 0) {
                this.modifiedStartLineNumber = 0;
                this.modifiedStartColumn = 0;
                this.modifiedEndLineNumber = 0;
                this.modifiedEndColumn = 0;
            }
            else {
                this.modifiedStartLineNumber = modifiedCharSequence.getStartLineNumber(diffChange.modifiedStart);
                this.modifiedStartColumn = modifiedCharSequence.getStartColumn(diffChange.modifiedStart);
                this.modifiedEndLineNumber = modifiedCharSequence.getEndLineNumber(diffChange.modifiedStart + diffChange.modifiedLength - 1);
                this.modifiedEndColumn = modifiedCharSequence.getEndColumn(diffChange.modifiedStart + diffChange.modifiedLength - 1);
            }
        }
        return CharChange;
    }());
    function postProcessCharChanges(rawChanges) {
        if (rawChanges.length <= 1) {
            return rawChanges;
        }
        var result = [rawChanges[0]];
        var i, len, originalMatchingLength, modifiedMatchingLength, matchingLength, prevChange = result[0], currChange;
        for (i = 1, len = rawChanges.length; i < len; i++) {
            currChange = rawChanges[i];
            originalMatchingLength = currChange.originalStart - (prevChange.originalStart + prevChange.originalLength);
            modifiedMatchingLength = currChange.modifiedStart - (prevChange.modifiedStart + prevChange.modifiedLength);
            // Both of the above should be equal, but the continueProcessingPredicate may prevent this from being true
            matchingLength = Math.min(originalMatchingLength, modifiedMatchingLength);
            if (matchingLength < MINIMUM_MATCHING_CHARACTER_LENGTH) {
                // Merge the current change into the previous one
                prevChange.originalLength = (currChange.originalStart + currChange.originalLength) - prevChange.originalStart;
                prevChange.modifiedLength = (currChange.modifiedStart + currChange.modifiedLength) - prevChange.modifiedStart;
            }
            else {
                // Add the current change
                result.push(currChange);
                prevChange = currChange;
            }
        }
        return result;
    }
    var LineChange = (function () {
        function LineChange(diffChange, originalLineSequence, modifiedLineSequence, continueProcessingPredicate, shouldPostProcessCharChanges) {
            if (diffChange.originalLength === 0) {
                this.originalStartLineNumber = originalLineSequence.getStartLineNumber(diffChange.originalStart) - 1;
                this.originalEndLineNumber = 0;
            }
            else {
                this.originalStartLineNumber = originalLineSequence.getStartLineNumber(diffChange.originalStart);
                this.originalEndLineNumber = originalLineSequence.getEndLineNumber(diffChange.originalStart + diffChange.originalLength - 1);
            }
            if (diffChange.modifiedLength === 0) {
                this.modifiedStartLineNumber = modifiedLineSequence.getStartLineNumber(diffChange.modifiedStart) - 1;
                this.modifiedEndLineNumber = 0;
            }
            else {
                this.modifiedStartLineNumber = modifiedLineSequence.getStartLineNumber(diffChange.modifiedStart);
                this.modifiedEndLineNumber = modifiedLineSequence.getEndLineNumber(diffChange.modifiedStart + diffChange.modifiedLength - 1);
            }
            if (diffChange.originalLength !== 0 && diffChange.modifiedLength !== 0 && continueProcessingPredicate()) {
                var originalCharSequence = originalLineSequence.getCharSequence(diffChange.originalStart, diffChange.originalStart + diffChange.originalLength - 1);
                var modifiedCharSequence = modifiedLineSequence.getCharSequence(diffChange.modifiedStart, diffChange.modifiedStart + diffChange.modifiedLength - 1);
                var rawChanges = computeDiff(originalCharSequence, modifiedCharSequence, continueProcessingPredicate);
                if (shouldPostProcessCharChanges) {
                    rawChanges = postProcessCharChanges(rawChanges);
                }
                this.charChanges = [];
                for (var i = 0, length = rawChanges.length; i < length; i++) {
                    this.charChanges.push(new CharChange(rawChanges[i], originalCharSequence, modifiedCharSequence));
                }
            }
        }
        return LineChange;
    }());
    var DiffComputer = (function () {
        function DiffComputer(originalLines, modifiedLines, opts) {
            this.shouldPostProcessCharChanges = opts.shouldPostProcessCharChanges;
            this.shouldIgnoreTrimWhitespace = opts.shouldIgnoreTrimWhitespace;
            this.maximumRunTimeMs = MAXIMUM_RUN_TIME;
            this.original = new LineMarkerSequence(originalLines, this.shouldIgnoreTrimWhitespace);
            this.modified = new LineMarkerSequence(modifiedLines, this.shouldIgnoreTrimWhitespace);
            if (opts.shouldConsiderTrimWhitespaceInEmptyCase && this.shouldIgnoreTrimWhitespace && this.original.equals(this.modified)) {
                // Diff would be empty with `shouldIgnoreTrimWhitespace`
                this.shouldIgnoreTrimWhitespace = false;
                this.original = new LineMarkerSequence(originalLines, this.shouldIgnoreTrimWhitespace);
                this.modified = new LineMarkerSequence(modifiedLines, this.shouldIgnoreTrimWhitespace);
            }
        }
        DiffComputer.prototype.computeDiff = function () {
            this.computationStartTime = (new Date()).getTime();
            var rawChanges = computeDiff(this.original, this.modified, this._continueProcessingPredicate.bind(this));
            var lineChanges = [];
            for (var i = 0, length = rawChanges.length; i < length; i++) {
                lineChanges.push(new LineChange(rawChanges[i], this.original, this.modified, this._continueProcessingPredicate.bind(this), this.shouldPostProcessCharChanges));
            }
            return lineChanges;
        };
        DiffComputer.prototype._continueProcessingPredicate = function () {
            if (this.maximumRunTimeMs === 0) {
                return true;
            }
            var now = (new Date()).getTime();
            return now - this.computationStartTime < this.maximumRunTimeMs;
        };
        return DiffComputer;
    }());
    exports.DiffComputer = DiffComputer;
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

define(__m[18], __M([1,0]), function (require, exports) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    // State machine for http:// or https://
    var STATE_MAP = [], START_STATE = 1, END_STATE = 9, ACCEPT_STATE = 10;
    STATE_MAP[1] = { 'h': 2, 'H': 2, 'f': 11, 'F': 11 };
    STATE_MAP[2] = { 't': 3, 'T': 3 };
    STATE_MAP[3] = { 't': 4, 'T': 4 };
    STATE_MAP[4] = { 'p': 5, 'P': 5 };
    STATE_MAP[5] = { 's': 6, 'S': 6, ':': 7 };
    STATE_MAP[6] = { ':': 7 };
    STATE_MAP[7] = { '/': 8 };
    STATE_MAP[8] = { '/': 9 };
    STATE_MAP[11] = { 'i': 12, 'I': 12 };
    STATE_MAP[12] = { 'l': 13, 'L': 13 };
    STATE_MAP[13] = { 'e': 6, 'E': 6 };
    var CharacterClass;
    (function (CharacterClass) {
        CharacterClass[CharacterClass["None"] = 0] = "None";
        CharacterClass[CharacterClass["ForceTermination"] = 1] = "ForceTermination";
        CharacterClass[CharacterClass["CannotEndIn"] = 2] = "CannotEndIn";
    })(CharacterClass || (CharacterClass = {}));
    var _openParens = '('.charCodeAt(0);
    var _closeParens = ')'.charCodeAt(0);
    var _openSquareBracket = '['.charCodeAt(0);
    var _closeSquareBracket = ']'.charCodeAt(0);
    var _openCurlyBracket = '{'.charCodeAt(0);
    var _closeCurlyBracket = '}'.charCodeAt(0);
    var CharacterClassifier = (function () {
        function CharacterClassifier() {
            var FORCE_TERMINATION_CHARACTERS = ' \t<>\'\"、。｡､，．：；？！＠＃＄％＆＊‘“〈《「『【〔（［｛｢｣｝］）〕】』」》〉”’｀～…';
            var CANNOT_END_WITH_CHARACTERS = '.,;';
            this._asciiMap = [];
            for (var i = 0; i < 256; i++) {
                this._asciiMap[i] = CharacterClass.None;
            }
            this._map = [];
            for (var i = 0; i < FORCE_TERMINATION_CHARACTERS.length; i++) {
                this._set(FORCE_TERMINATION_CHARACTERS.charCodeAt(i), CharacterClass.ForceTermination);
            }
            for (var i = 0; i < CANNOT_END_WITH_CHARACTERS.length; i++) {
                this._set(CANNOT_END_WITH_CHARACTERS.charCodeAt(i), CharacterClass.CannotEndIn);
            }
        }
        CharacterClassifier.prototype._set = function (charCode, charClass) {
            if (charCode < 256) {
                this._asciiMap[charCode] = charClass;
            }
            this._map[charCode] = charClass;
        };
        CharacterClassifier.prototype.classify = function (charCode) {
            if (charCode < 256) {
                return this._asciiMap[charCode];
            }
            var charClass = this._map[charCode];
            if (charClass) {
                return charClass;
            }
            return CharacterClass.None;
        };
        return CharacterClassifier;
    }());
    var LinkComputer = (function () {
        function LinkComputer() {
        }
        LinkComputer._createLink = function (line, lineNumber, linkBeginIndex, linkEndIndex) {
            return {
                range: {
                    startLineNumber: lineNumber,
                    startColumn: linkBeginIndex + 1,
                    endLineNumber: lineNumber,
                    endColumn: linkEndIndex + 1
                },
                url: line.substring(linkBeginIndex, linkEndIndex)
            };
        };
        LinkComputer.computeLinks = function (model) {
            var i, lineCount, result = [];
            var line, j, lastIncludedCharIndex, len, linkBeginIndex, state, ch, chCode, chClass, resetStateMachine, hasOpenParens, hasOpenSquareBracket, hasOpenCurlyBracket, characterClassifier = LinkComputer._characterClassifier;
            for (i = 1, lineCount = model.getLineCount(); i <= lineCount; i++) {
                line = model.getLineContent(i);
                j = 0;
                len = line.length;
                linkBeginIndex = 0;
                state = START_STATE;
                hasOpenParens = false;
                hasOpenSquareBracket = false;
                hasOpenCurlyBracket = false;
                while (j < len) {
                    ch = line.charAt(j);
                    chCode = line.charCodeAt(j);
                    resetStateMachine = false;
                    if (state === ACCEPT_STATE) {
                        switch (chCode) {
                            case _openParens:
                                hasOpenParens = true;
                                chClass = CharacterClass.None;
                                break;
                            case _closeParens:
                                chClass = (hasOpenParens ? CharacterClass.None : CharacterClass.ForceTermination);
                                break;
                            case _openSquareBracket:
                                hasOpenSquareBracket = true;
                                chClass = CharacterClass.None;
                                break;
                            case _closeSquareBracket:
                                chClass = (hasOpenSquareBracket ? CharacterClass.None : CharacterClass.ForceTermination);
                                break;
                            case _openCurlyBracket:
                                hasOpenCurlyBracket = true;
                                chClass = CharacterClass.None;
                                break;
                            case _closeCurlyBracket:
                                chClass = (hasOpenCurlyBracket ? CharacterClass.None : CharacterClass.ForceTermination);
                                break;
                            default:
                                chClass = characterClassifier.classify(chCode);
                        }
                        // Check if character terminates link
                        if (chClass === CharacterClass.ForceTermination) {
                            // Do not allow to end link in certain characters...
                            lastIncludedCharIndex = j - 1;
                            do {
                                chCode = line.charCodeAt(lastIncludedCharIndex);
                                chClass = characterClassifier.classify(chCode);
                                if (chClass !== CharacterClass.CannotEndIn) {
                                    break;
                                }
                                lastIncludedCharIndex--;
                            } while (lastIncludedCharIndex > linkBeginIndex);
                            result.push(LinkComputer._createLink(line, i, linkBeginIndex, lastIncludedCharIndex + 1));
                            resetStateMachine = true;
                        }
                    }
                    else if (state === END_STATE) {
                        chClass = characterClassifier.classify(chCode);
                        // Check if character terminates link
                        if (chClass === CharacterClass.ForceTermination) {
                            resetStateMachine = true;
                        }
                        else {
                            state = ACCEPT_STATE;
                        }
                    }
                    else {
                        if (STATE_MAP[state].hasOwnProperty(ch)) {
                            state = STATE_MAP[state][ch];
                        }
                        else {
                            resetStateMachine = true;
                        }
                    }
                    if (resetStateMachine) {
                        state = START_STATE;
                        hasOpenParens = false;
                        hasOpenSquareBracket = false;
                        hasOpenCurlyBracket = false;
                        // Record where the link started
                        linkBeginIndex = j + 1;
                    }
                    j++;
                }
                if (state === ACCEPT_STATE) {
                    result.push(LinkComputer._createLink(line, i, linkBeginIndex, len));
                }
            }
            return result;
        };
        LinkComputer._characterClassifier = new CharacterClassifier();
        return LinkComputer;
    }());
    /**
     * Returns an array of all links contains in the provided
     * document. *Note* that this operation is computational
     * expensive and should not run in the UI thread.
     */
    function computeLinks(model) {
        if (!model || typeof model.getLineCount !== 'function' || typeof model.getLineContent !== 'function') {
            // Unknown caller!
            return [];
        }
        return LinkComputer.computeLinks(model);
    }
    exports.computeLinks = computeLinks;
});

define(__m[16], __M([1,0]), function (require, exports) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var BasicInplaceReplace = (function () {
        function BasicInplaceReplace() {
            this._defaultValueSet = [
                ['true', 'false'],
                ['True', 'False'],
                ['Private', 'Public', 'Friend', 'ReadOnly', 'Partial', 'Protected', 'WriteOnly'],
                ['public', 'protected', 'private'],
            ];
        }
        BasicInplaceReplace.prototype.navigateValueSet = function (range1, text1, range2, text2, up) {
            if (range1 && text1) {
                var result = this.doNavigateValueSet(text1, up);
                if (result) {
                    return {
                        range: range1,
                        value: result
                    };
                }
            }
            if (range2 && text2) {
                var result = this.doNavigateValueSet(text2, up);
                if (result) {
                    return {
                        range: range2,
                        value: result
                    };
                }
            }
            return null;
        };
        BasicInplaceReplace.prototype.doNavigateValueSet = function (text, up) {
            var numberResult = this.numberReplace(text, up);
            if (numberResult !== null) {
                return numberResult;
            }
            return this.textReplace(text, up);
        };
        BasicInplaceReplace.prototype.numberReplace = function (value, up) {
            var precision = Math.pow(10, value.length - (value.lastIndexOf('.') + 1)), n1 = Number(value), n2 = parseFloat(value);
            if (!isNaN(n1) && !isNaN(n2) && n1 === n2) {
                if (n1 === 0 && !up) {
                    return null; // don't do negative
                }
                else {
                    n1 = Math.floor(n1 * precision);
                    n1 += up ? precision : -precision;
                    return String(n1 / precision);
                }
            }
            return null;
        };
        BasicInplaceReplace.prototype.textReplace = function (value, up) {
            return this.valueSetsReplace(this._defaultValueSet, value, up);
        };
        BasicInplaceReplace.prototype.valueSetsReplace = function (valueSets, value, up) {
            var result = null;
            for (var i = 0, len = valueSets.length; result === null && i < len; i++) {
                result = this.valueSetReplace(valueSets[i], value, up);
            }
            return result;
        };
        BasicInplaceReplace.prototype.valueSetReplace = function (valueSet, value, up) {
            var idx = valueSet.indexOf(value);
            if (idx >= 0) {
                idx += up ? +1 : -1;
                if (idx < 0) {
                    idx = valueSet.length - 1;
                }
                else {
                    idx %= valueSet.length;
                }
                return valueSet[idx];
            }
            return null;
        };
        BasicInplaceReplace.INSTANCE = new BasicInplaceReplace();
        return BasicInplaceReplace;
    }());
    exports.BasicInplaceReplace = BasicInplaceReplace;
});

define(__m[20], __M([1,0]), function (require, exports) {
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

define(__m[21], __M([1,0,20]), function (require, exports, prefixSumComputer_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var MirrorModel2 = (function () {
        function MirrorModel2(uri, lines, eol, versionId) {
            this._uri = uri;
            this._lines = lines;
            this._eol = eol;
            this._versionId = versionId;
        }
        MirrorModel2.prototype.dispose = function () {
            this._lines.length = 0;
        };
        Object.defineProperty(MirrorModel2.prototype, "version", {
            get: function () {
                return this._versionId;
            },
            enumerable: true,
            configurable: true
        });
        MirrorModel2.prototype.getText = function () {
            return this._lines.join(this._eol);
        };
        MirrorModel2.prototype.onEvents = function (events) {
            var newEOL = null;
            for (var i = 0, len = events.length; i < len; i++) {
                var e = events[i];
                if (e.eol) {
                    newEOL = e.eol;
                }
            }
            if (newEOL && newEOL !== this._eol) {
                this._eol = newEOL;
                this._lineStarts = null;
            }
            // Update my lines
            var lastVersionId = -1;
            for (var i = 0, len = events.length; i < len; i++) {
                var e = events[i];
                this._acceptDeleteRange(e.range);
                this._acceptInsertText({
                    lineNumber: e.range.startLineNumber,
                    column: e.range.startColumn
                }, e.text);
                lastVersionId = Math.max(lastVersionId, e.versionId);
            }
            if (lastVersionId !== -1) {
                this._versionId = lastVersionId;
            }
        };
        MirrorModel2.prototype._ensureLineStarts = function () {
            if (!this._lineStarts) {
                var lineStartValues = [];
                var eolLength = this._eol.length;
                for (var i = 0, len = this._lines.length; i < len; i++) {
                    lineStartValues.push(this._lines[i].length + eolLength);
                }
                this._lineStarts = new prefixSumComputer_1.PrefixSumComputer(lineStartValues);
            }
        };
        /**
         * All changes to a line's text go through this method
         */
        MirrorModel2.prototype._setLineText = function (lineIndex, newValue) {
            this._lines[lineIndex] = newValue;
            if (this._lineStarts) {
                // update prefix sum
                this._lineStarts.changeValue(lineIndex, this._lines[lineIndex].length + this._eol.length);
            }
        };
        MirrorModel2.prototype._acceptDeleteRange = function (range) {
            if (range.startLineNumber === range.endLineNumber) {
                if (range.startColumn === range.endColumn) {
                    // Nothing to delete
                    return;
                }
                // Delete text on the affected line
                this._setLineText(range.startLineNumber - 1, this._lines[range.startLineNumber - 1].substring(0, range.startColumn - 1)
                    + this._lines[range.startLineNumber - 1].substring(range.endColumn - 1));
                return;
            }
            // Take remaining text on last line and append it to remaining text on first line
            this._setLineText(range.startLineNumber - 1, this._lines[range.startLineNumber - 1].substring(0, range.startColumn - 1)
                + this._lines[range.endLineNumber - 1].substring(range.endColumn - 1));
            // Delete middle lines
            this._lines.splice(range.startLineNumber, range.endLineNumber - range.startLineNumber);
            if (this._lineStarts) {
                // update prefix sum
                this._lineStarts.removeValues(range.startLineNumber, range.endLineNumber - range.startLineNumber);
            }
        };
        MirrorModel2.prototype._acceptInsertText = function (position, insertText) {
            if (insertText.length === 0) {
                // Nothing to insert
                return;
            }
            var insertLines = insertText.split(/\r\n|\r|\n/);
            if (insertLines.length === 1) {
                // Inserting text on one line
                this._setLineText(position.lineNumber - 1, this._lines[position.lineNumber - 1].substring(0, position.column - 1)
                    + insertLines[0]
                    + this._lines[position.lineNumber - 1].substring(position.column - 1));
                return;
            }
            // Append overflowing text from first line to the end of text to insert
            insertLines[insertLines.length - 1] += this._lines[position.lineNumber - 1].substring(position.column - 1);
            // Delete overflowing text from first line and insert text on first line
            this._setLineText(position.lineNumber - 1, this._lines[position.lineNumber - 1].substring(0, position.column - 1)
                + insertLines[0]);
            // Insert new lines & store lengths
            var newLengths = new Array(insertLines.length - 1);
            for (var i = 1; i < insertLines.length; i++) {
                this._lines.splice(position.lineNumber + i - 1, 0, insertLines[i]);
                newLengths[i - 1] = insertLines[i].length + this._eol.length;
            }
            if (this._lineStarts) {
                // update prefix sum
                this._lineStarts.insertValues(position.lineNumber, newLengths);
            }
        };
        return MirrorModel2;
    }());
    exports.MirrorModel2 = MirrorModel2;
});

define(__m[22], __M([11,12]), function(nls, data) { return nls.create("vs/base/common/errors", data); });
define(__m[5], __M([1,0,22,23,3,6,27,4]), function (require, exports, nls, objects, platform, types, arrays, strings) {
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

define(__m[26], __M([1,0,5]), function (require, exports, errors_1) {
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

define(__m[14], __M([1,0,26]), function (require, exports, callbackList_1) {
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
define(__m[15], __M([1,0,14]), function (require, exports, event_1) {
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

/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Microsoft Corporation. All rights reserved.
 *  Licensed under the MIT License. See License.txt in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

define(__m[2], __M([37,5]), function (winjs, __Errors__) {
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





define(__m[31], __M([1,0,5,3,2,15,10]), function (require, exports, errors, platform, winjs_base_1, cancellation_1, lifecycle_1) {
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






define(__m[36], __M([1,0,5,10,2,31]), function (require, exports, errors_1, lifecycle_1, winjs_base_1, async_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    var INITIALIZE = '$initialize';
    var SimpleWorkerProtocol = (function () {
        function SimpleWorkerProtocol(handler) {
            this._workerId = -1;
            this._handler = handler;
            this._lastSentReq = 0;
            this._pendingReplies = Object.create(null);
        }
        SimpleWorkerProtocol.prototype.setWorkerId = function (workerId) {
            this._workerId = workerId;
        };
        SimpleWorkerProtocol.prototype.sendMessage = function (method, args) {
            var req = String(++this._lastSentReq);
            var reply = {
                c: null,
                e: null
            };
            var result = new winjs_base_1.TPromise(function (c, e, p) {
                reply.c = c;
                reply.e = e;
            }, function () {
                // Cancel not supported
            });
            this._pendingReplies[req] = reply;
            this._send({
                vsWorker: this._workerId,
                req: req,
                method: method,
                args: args
            });
            return result;
        };
        SimpleWorkerProtocol.prototype.handleMessage = function (serializedMessage) {
            var message;
            try {
                message = JSON.parse(serializedMessage);
            }
            catch (e) {
            }
            if (!message.vsWorker) {
                return;
            }
            if (this._workerId !== -1 && message.vsWorker !== this._workerId) {
                return;
            }
            this._handleMessage(message);
        };
        SimpleWorkerProtocol.prototype._handleMessage = function (msg) {
            var _this = this;
            if (msg.seq) {
                var replyMessage = msg;
                if (!this._pendingReplies[replyMessage.seq]) {
                    console.warn('Got reply to unknown seq');
                    return;
                }
                var reply = this._pendingReplies[replyMessage.seq];
                delete this._pendingReplies[replyMessage.seq];
                if (replyMessage.err) {
                    var err = replyMessage.err;
                    if (replyMessage.err.$isError) {
                        err = new Error();
                        err.name = replyMessage.err.name;
                        err.message = replyMessage.err.message;
                        err.stack = replyMessage.err.stack;
                    }
                    reply.e(err);
                    return;
                }
                reply.c(replyMessage.res);
                return;
            }
            var requestMessage = msg;
            var req = requestMessage.req;
            var result = this._handler.handleMessage(requestMessage.method, requestMessage.args);
            result.then(function (r) {
                _this._send({
                    vsWorker: _this._workerId,
                    seq: req,
                    res: r,
                    err: undefined
                });
            }, function (e) {
                _this._send({
                    vsWorker: _this._workerId,
                    seq: req,
                    res: undefined,
                    err: errors_1.transformErrorForSerialization(e)
                });
            });
        };
        SimpleWorkerProtocol.prototype._send = function (msg) {
            var strMsg = JSON.stringify(msg);
            // console.log('SENDING: ' + strMsg);
            this._handler.sendMessage(strMsg);
        };
        return SimpleWorkerProtocol;
    }());
    /**
     * Main thread side
     */
    var SimpleWorkerClient = (function (_super) {
        __extends(SimpleWorkerClient, _super);
        function SimpleWorkerClient(workerFactory, moduleId) {
            var _this = this;
            _super.call(this);
            this._lastRequestTimestamp = -1;
            var lazyProxyFulfill = null;
            var lazyProxyReject = null;
            this._worker = this._register(workerFactory.create('vs/base/common/worker/simpleWorker', function (msg) {
                _this._protocol.handleMessage(msg);
            }, function (err) {
                // in Firefox, web workers fail lazily :(
                // we will reject the proxy
                lazyProxyReject(err);
            }));
            this._protocol = new SimpleWorkerProtocol({
                sendMessage: function (msg) {
                    _this._worker.postMessage(msg);
                },
                handleMessage: function (method, args) {
                    // Intentionally not supporting worker -> main requests
                    return winjs_base_1.TPromise.as(null);
                }
            });
            this._protocol.setWorkerId(this._worker.getId());
            // Gather loader configuration
            var loaderConfiguration = null;
            var globalRequire = window.require;
            if (typeof globalRequire.getConfig === 'function') {
                // Get the configuration from the Monaco AMD Loader
                loaderConfiguration = globalRequire.getConfig();
            }
            else if (typeof window.requirejs !== 'undefined') {
                // Get the configuration from requirejs
                loaderConfiguration = window.requirejs.s.contexts._.config;
            }
            this._lazyProxy = new winjs_base_1.TPromise(function (c, e, p) {
                lazyProxyFulfill = c;
                lazyProxyReject = e;
            }, function () { });
            // Send initialize message
            this._onModuleLoaded = this._protocol.sendMessage(INITIALIZE, [
                this._worker.getId(),
                moduleId,
                loaderConfiguration
            ]);
            this._onModuleLoaded.then(function (availableMethods) {
                var proxy = {};
                for (var i = 0; i < availableMethods.length; i++) {
                    proxy[availableMethods[i]] = createProxyMethod(availableMethods[i], proxyMethodRequest);
                }
                lazyProxyFulfill(proxy);
            }, function (e) {
                lazyProxyReject(e);
                _this._onError('Worker failed to load ' + moduleId, e);
            });
            // Create proxy to loaded code
            var proxyMethodRequest = function (method, args) {
                return _this._request(method, args);
            };
            var createProxyMethod = function (method, proxyMethodRequest) {
                return function () {
                    var args = Array.prototype.slice.call(arguments, 0);
                    return proxyMethodRequest(method, args);
                };
            };
        }
        SimpleWorkerClient.prototype.getProxyObject = function () {
            // Do not allow chaining promises to cancel the proxy creation
            return new async_1.ShallowCancelThenPromise(this._lazyProxy);
        };
        SimpleWorkerClient.prototype.getLastRequestTimestamp = function () {
            return this._lastRequestTimestamp;
        };
        SimpleWorkerClient.prototype._request = function (method, args) {
            var _this = this;
            return new winjs_base_1.TPromise(function (c, e, p) {
                _this._onModuleLoaded.then(function () {
                    _this._lastRequestTimestamp = Date.now();
                    _this._protocol.sendMessage(method, args).then(c, e);
                }, e);
            }, function () {
                // Cancel intentionally not supported
            });
        };
        SimpleWorkerClient.prototype._onError = function (message, error) {
            console.error(message);
            console.info(error);
        };
        return SimpleWorkerClient;
    }(lifecycle_1.Disposable));
    exports.SimpleWorkerClient = SimpleWorkerClient;
    /**
     * Worker side
     */
    var SimpleWorkerServer = (function () {
        function SimpleWorkerServer(postSerializedMessage) {
            var _this = this;
            this._protocol = new SimpleWorkerProtocol({
                sendMessage: function (msg) {
                    postSerializedMessage(msg);
                },
                handleMessage: function (method, args) { return _this._handleMessage(method, args); }
            });
        }
        SimpleWorkerServer.prototype.onmessage = function (msg) {
            this._protocol.handleMessage(msg);
        };
        SimpleWorkerServer.prototype._handleMessage = function (method, args) {
            if (method === INITIALIZE) {
                return this.initialize(args[0], args[1], args[2]);
            }
            if (!this._requestHandler || typeof this._requestHandler[method] !== 'function') {
                return winjs_base_1.TPromise.wrapError(new Error('Missing requestHandler or method: ' + method));
            }
            try {
                return winjs_base_1.TPromise.as(this._requestHandler[method].apply(this._requestHandler, args));
            }
            catch (e) {
                return winjs_base_1.TPromise.wrapError(e);
            }
        };
        SimpleWorkerServer.prototype.initialize = function (workerId, moduleId, loaderConfig) {
            var _this = this;
            this._protocol.setWorkerId(workerId);
            // TODO@Alex: share this code with workerServer
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
            var cc;
            var ee;
            var r = new winjs_base_1.TPromise(function (c, e, p) {
                cc = c;
                ee = e;
            });
            // Use the global require to be sure to get the global config
            self.require([moduleId], function () {
                var result = [];
                for (var _i = 0; _i < arguments.length; _i++) {
                    result[_i - 0] = arguments[_i];
                }
                var handlerModule = result[0];
                _this._requestHandler = handlerModule.create();
                var methods = [];
                for (var prop in _this._requestHandler) {
                    if (typeof _this._requestHandler[prop] === 'function') {
                        methods.push(prop);
                    }
                }
                cc(methods);
            }, ee);
            return r;
        };
        return SimpleWorkerServer;
    }());
    exports.SimpleWorkerServer = SimpleWorkerServer;
    /**
     * Called on the worker side
     */
    function create(postMessage) {
        return new SimpleWorkerServer(postMessage);
    }
    exports.create = create;
});

define(__m[33], __M([11,12]), function(nls, data) { return nls.create("vs/base/common/keyCodes", data); });
/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Microsoft Corporation. All rights reserved.
 *  Licensed under the MIT License. See License.txt in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
define(__m[34], __M([1,0,33,3]), function (require, exports, nls, defaultPlatform) {
    'use strict';
    /**
     * Virtual Key Codes, the value does not hold any inherent meaning.
     * Inspired somewhat from https://msdn.microsoft.com/en-us/library/windows/desktop/dd375731(v=vs.85).aspx
     * But these are "more general", as they should work across browsers & OS`s.
     */
    (function (KeyCode) {
        /**
         * Placed first to cover the 0 value of the enum.
         */
        KeyCode[KeyCode["Unknown"] = 0] = "Unknown";
        KeyCode[KeyCode["Backspace"] = 1] = "Backspace";
        KeyCode[KeyCode["Tab"] = 2] = "Tab";
        KeyCode[KeyCode["Enter"] = 3] = "Enter";
        KeyCode[KeyCode["Shift"] = 4] = "Shift";
        KeyCode[KeyCode["Ctrl"] = 5] = "Ctrl";
        KeyCode[KeyCode["Alt"] = 6] = "Alt";
        KeyCode[KeyCode["PauseBreak"] = 7] = "PauseBreak";
        KeyCode[KeyCode["CapsLock"] = 8] = "CapsLock";
        KeyCode[KeyCode["Escape"] = 9] = "Escape";
        KeyCode[KeyCode["Space"] = 10] = "Space";
        KeyCode[KeyCode["PageUp"] = 11] = "PageUp";
        KeyCode[KeyCode["PageDown"] = 12] = "PageDown";
        KeyCode[KeyCode["End"] = 13] = "End";
        KeyCode[KeyCode["Home"] = 14] = "Home";
        KeyCode[KeyCode["LeftArrow"] = 15] = "LeftArrow";
        KeyCode[KeyCode["UpArrow"] = 16] = "UpArrow";
        KeyCode[KeyCode["RightArrow"] = 17] = "RightArrow";
        KeyCode[KeyCode["DownArrow"] = 18] = "DownArrow";
        KeyCode[KeyCode["Insert"] = 19] = "Insert";
        KeyCode[KeyCode["Delete"] = 20] = "Delete";
        KeyCode[KeyCode["KEY_0"] = 21] = "KEY_0";
        KeyCode[KeyCode["KEY_1"] = 22] = "KEY_1";
        KeyCode[KeyCode["KEY_2"] = 23] = "KEY_2";
        KeyCode[KeyCode["KEY_3"] = 24] = "KEY_3";
        KeyCode[KeyCode["KEY_4"] = 25] = "KEY_4";
        KeyCode[KeyCode["KEY_5"] = 26] = "KEY_5";
        KeyCode[KeyCode["KEY_6"] = 27] = "KEY_6";
        KeyCode[KeyCode["KEY_7"] = 28] = "KEY_7";
        KeyCode[KeyCode["KEY_8"] = 29] = "KEY_8";
        KeyCode[KeyCode["KEY_9"] = 30] = "KEY_9";
        KeyCode[KeyCode["KEY_A"] = 31] = "KEY_A";
        KeyCode[KeyCode["KEY_B"] = 32] = "KEY_B";
        KeyCode[KeyCode["KEY_C"] = 33] = "KEY_C";
        KeyCode[KeyCode["KEY_D"] = 34] = "KEY_D";
        KeyCode[KeyCode["KEY_E"] = 35] = "KEY_E";
        KeyCode[KeyCode["KEY_F"] = 36] = "KEY_F";
        KeyCode[KeyCode["KEY_G"] = 37] = "KEY_G";
        KeyCode[KeyCode["KEY_H"] = 38] = "KEY_H";
        KeyCode[KeyCode["KEY_I"] = 39] = "KEY_I";
        KeyCode[KeyCode["KEY_J"] = 40] = "KEY_J";
        KeyCode[KeyCode["KEY_K"] = 41] = "KEY_K";
        KeyCode[KeyCode["KEY_L"] = 42] = "KEY_L";
        KeyCode[KeyCode["KEY_M"] = 43] = "KEY_M";
        KeyCode[KeyCode["KEY_N"] = 44] = "KEY_N";
        KeyCode[KeyCode["KEY_O"] = 45] = "KEY_O";
        KeyCode[KeyCode["KEY_P"] = 46] = "KEY_P";
        KeyCode[KeyCode["KEY_Q"] = 47] = "KEY_Q";
        KeyCode[KeyCode["KEY_R"] = 48] = "KEY_R";
        KeyCode[KeyCode["KEY_S"] = 49] = "KEY_S";
        KeyCode[KeyCode["KEY_T"] = 50] = "KEY_T";
        KeyCode[KeyCode["KEY_U"] = 51] = "KEY_U";
        KeyCode[KeyCode["KEY_V"] = 52] = "KEY_V";
        KeyCode[KeyCode["KEY_W"] = 53] = "KEY_W";
        KeyCode[KeyCode["KEY_X"] = 54] = "KEY_X";
        KeyCode[KeyCode["KEY_Y"] = 55] = "KEY_Y";
        KeyCode[KeyCode["KEY_Z"] = 56] = "KEY_Z";
        KeyCode[KeyCode["Meta"] = 57] = "Meta";
        KeyCode[KeyCode["ContextMenu"] = 58] = "ContextMenu";
        KeyCode[KeyCode["F1"] = 59] = "F1";
        KeyCode[KeyCode["F2"] = 60] = "F2";
        KeyCode[KeyCode["F3"] = 61] = "F3";
        KeyCode[KeyCode["F4"] = 62] = "F4";
        KeyCode[KeyCode["F5"] = 63] = "F5";
        KeyCode[KeyCode["F6"] = 64] = "F6";
        KeyCode[KeyCode["F7"] = 65] = "F7";
        KeyCode[KeyCode["F8"] = 66] = "F8";
        KeyCode[KeyCode["F9"] = 67] = "F9";
        KeyCode[KeyCode["F10"] = 68] = "F10";
        KeyCode[KeyCode["F11"] = 69] = "F11";
        KeyCode[KeyCode["F12"] = 70] = "F12";
        KeyCode[KeyCode["F13"] = 71] = "F13";
        KeyCode[KeyCode["F14"] = 72] = "F14";
        KeyCode[KeyCode["F15"] = 73] = "F15";
        KeyCode[KeyCode["F16"] = 74] = "F16";
        KeyCode[KeyCode["F17"] = 75] = "F17";
        KeyCode[KeyCode["F18"] = 76] = "F18";
        KeyCode[KeyCode["F19"] = 77] = "F19";
        KeyCode[KeyCode["NumLock"] = 78] = "NumLock";
        KeyCode[KeyCode["ScrollLock"] = 79] = "ScrollLock";
        /**
         * Used for miscellaneous characters; it can vary by keyboard.
         * For the US standard keyboard, the ';:' key
         */
        KeyCode[KeyCode["US_SEMICOLON"] = 80] = "US_SEMICOLON";
        /**
         * For any country/region, the '+' key
         * For the US standard keyboard, the '=+' key
         */
        KeyCode[KeyCode["US_EQUAL"] = 81] = "US_EQUAL";
        /**
         * For any country/region, the ',' key
         * For the US standard keyboard, the ',<' key
         */
        KeyCode[KeyCode["US_COMMA"] = 82] = "US_COMMA";
        /**
         * For any country/region, the '-' key
         * For the US standard keyboard, the '-_' key
         */
        KeyCode[KeyCode["US_MINUS"] = 83] = "US_MINUS";
        /**
         * For any country/region, the '.' key
         * For the US standard keyboard, the '.>' key
         */
        KeyCode[KeyCode["US_DOT"] = 84] = "US_DOT";
        /**
         * Used for miscellaneous characters; it can vary by keyboard.
         * For the US standard keyboard, the '/?' key
         */
        KeyCode[KeyCode["US_SLASH"] = 85] = "US_SLASH";
        /**
         * Used for miscellaneous characters; it can vary by keyboard.
         * For the US standard keyboard, the '`~' key
         */
        KeyCode[KeyCode["US_BACKTICK"] = 86] = "US_BACKTICK";
        /**
         * Used for miscellaneous characters; it can vary by keyboard.
         * For the US standard keyboard, the '[{' key
         */
        KeyCode[KeyCode["US_OPEN_SQUARE_BRACKET"] = 87] = "US_OPEN_SQUARE_BRACKET";
        /**
         * Used for miscellaneous characters; it can vary by keyboard.
         * For the US standard keyboard, the '\|' key
         */
        KeyCode[KeyCode["US_BACKSLASH"] = 88] = "US_BACKSLASH";
        /**
         * Used for miscellaneous characters; it can vary by keyboard.
         * For the US standard keyboard, the ']}' key
         */
        KeyCode[KeyCode["US_CLOSE_SQUARE_BRACKET"] = 89] = "US_CLOSE_SQUARE_BRACKET";
        /**
         * Used for miscellaneous characters; it can vary by keyboard.
         * For the US standard keyboard, the ''"' key
         */
        KeyCode[KeyCode["US_QUOTE"] = 90] = "US_QUOTE";
        /**
         * Used for miscellaneous characters; it can vary by keyboard.
         */
        KeyCode[KeyCode["OEM_8"] = 91] = "OEM_8";
        /**
         * Either the angle bracket key or the backslash key on the RT 102-key keyboard.
         */
        KeyCode[KeyCode["OEM_102"] = 92] = "OEM_102";
        KeyCode[KeyCode["NUMPAD_0"] = 93] = "NUMPAD_0";
        KeyCode[KeyCode["NUMPAD_1"] = 94] = "NUMPAD_1";
        KeyCode[KeyCode["NUMPAD_2"] = 95] = "NUMPAD_2";
        KeyCode[KeyCode["NUMPAD_3"] = 96] = "NUMPAD_3";
        KeyCode[KeyCode["NUMPAD_4"] = 97] = "NUMPAD_4";
        KeyCode[KeyCode["NUMPAD_5"] = 98] = "NUMPAD_5";
        KeyCode[KeyCode["NUMPAD_6"] = 99] = "NUMPAD_6";
        KeyCode[KeyCode["NUMPAD_7"] = 100] = "NUMPAD_7";
        KeyCode[KeyCode["NUMPAD_8"] = 101] = "NUMPAD_8";
        KeyCode[KeyCode["NUMPAD_9"] = 102] = "NUMPAD_9";
        KeyCode[KeyCode["NUMPAD_MULTIPLY"] = 103] = "NUMPAD_MULTIPLY";
        KeyCode[KeyCode["NUMPAD_ADD"] = 104] = "NUMPAD_ADD";
        KeyCode[KeyCode["NUMPAD_SEPARATOR"] = 105] = "NUMPAD_SEPARATOR";
        KeyCode[KeyCode["NUMPAD_SUBTRACT"] = 106] = "NUMPAD_SUBTRACT";
        KeyCode[KeyCode["NUMPAD_DECIMAL"] = 107] = "NUMPAD_DECIMAL";
        KeyCode[KeyCode["NUMPAD_DIVIDE"] = 108] = "NUMPAD_DIVIDE";
        /**
         * Placed last to cover the length of the enum.
         */
        KeyCode[KeyCode["MAX_VALUE"] = 109] = "MAX_VALUE";
    })(exports.KeyCode || (exports.KeyCode = {}));
    var KeyCode = exports.KeyCode;
    var Mapping = (function () {
        function Mapping(fromKeyCode, toKeyCode) {
            this._fromKeyCode = fromKeyCode;
            this._toKeyCode = toKeyCode;
        }
        Mapping.prototype.fromKeyCode = function (keyCode) {
            return this._fromKeyCode[keyCode];
        };
        Mapping.prototype.toKeyCode = function (str) {
            if (this._toKeyCode.hasOwnProperty(str)) {
                return this._toKeyCode[str];
            }
            return KeyCode.Unknown;
        };
        return Mapping;
    }());
    function createMapping(fill1, fill2) {
        var MAP = [];
        fill1(MAP);
        var REVERSE_MAP = {};
        for (var i = 0, len = MAP.length; i < len; i++) {
            if (!MAP[i]) {
                continue;
            }
            REVERSE_MAP[MAP[i]] = i;
        }
        fill2(REVERSE_MAP);
        var FINAL_REVERSE_MAP = {};
        for (var entry in REVERSE_MAP) {
            if (REVERSE_MAP.hasOwnProperty(entry)) {
                FINAL_REVERSE_MAP[entry] = REVERSE_MAP[entry];
                FINAL_REVERSE_MAP[entry.toLowerCase()] = REVERSE_MAP[entry];
            }
        }
        return new Mapping(MAP, FINAL_REVERSE_MAP);
    }
    var STRING = createMapping(function (TO_STRING_MAP) {
        TO_STRING_MAP[KeyCode.Unknown] = 'unknown';
        TO_STRING_MAP[KeyCode.Backspace] = 'Backspace';
        TO_STRING_MAP[KeyCode.Tab] = 'Tab';
        TO_STRING_MAP[KeyCode.Enter] = 'Enter';
        TO_STRING_MAP[KeyCode.Shift] = 'Shift';
        TO_STRING_MAP[KeyCode.Ctrl] = 'Ctrl';
        TO_STRING_MAP[KeyCode.Alt] = 'Alt';
        TO_STRING_MAP[KeyCode.PauseBreak] = 'PauseBreak';
        TO_STRING_MAP[KeyCode.CapsLock] = 'CapsLock';
        TO_STRING_MAP[KeyCode.Escape] = 'Escape';
        TO_STRING_MAP[KeyCode.Space] = 'Space';
        TO_STRING_MAP[KeyCode.PageUp] = 'PageUp';
        TO_STRING_MAP[KeyCode.PageDown] = 'PageDown';
        TO_STRING_MAP[KeyCode.End] = 'End';
        TO_STRING_MAP[KeyCode.Home] = 'Home';
        TO_STRING_MAP[KeyCode.LeftArrow] = 'LeftArrow';
        TO_STRING_MAP[KeyCode.UpArrow] = 'UpArrow';
        TO_STRING_MAP[KeyCode.RightArrow] = 'RightArrow';
        TO_STRING_MAP[KeyCode.DownArrow] = 'DownArrow';
        TO_STRING_MAP[KeyCode.Insert] = 'Insert';
        TO_STRING_MAP[KeyCode.Delete] = 'Delete';
        TO_STRING_MAP[KeyCode.KEY_0] = '0';
        TO_STRING_MAP[KeyCode.KEY_1] = '1';
        TO_STRING_MAP[KeyCode.KEY_2] = '2';
        TO_STRING_MAP[KeyCode.KEY_3] = '3';
        TO_STRING_MAP[KeyCode.KEY_4] = '4';
        TO_STRING_MAP[KeyCode.KEY_5] = '5';
        TO_STRING_MAP[KeyCode.KEY_6] = '6';
        TO_STRING_MAP[KeyCode.KEY_7] = '7';
        TO_STRING_MAP[KeyCode.KEY_8] = '8';
        TO_STRING_MAP[KeyCode.KEY_9] = '9';
        TO_STRING_MAP[KeyCode.KEY_A] = 'A';
        TO_STRING_MAP[KeyCode.KEY_B] = 'B';
        TO_STRING_MAP[KeyCode.KEY_C] = 'C';
        TO_STRING_MAP[KeyCode.KEY_D] = 'D';
        TO_STRING_MAP[KeyCode.KEY_E] = 'E';
        TO_STRING_MAP[KeyCode.KEY_F] = 'F';
        TO_STRING_MAP[KeyCode.KEY_G] = 'G';
        TO_STRING_MAP[KeyCode.KEY_H] = 'H';
        TO_STRING_MAP[KeyCode.KEY_I] = 'I';
        TO_STRING_MAP[KeyCode.KEY_J] = 'J';
        TO_STRING_MAP[KeyCode.KEY_K] = 'K';
        TO_STRING_MAP[KeyCode.KEY_L] = 'L';
        TO_STRING_MAP[KeyCode.KEY_M] = 'M';
        TO_STRING_MAP[KeyCode.KEY_N] = 'N';
        TO_STRING_MAP[KeyCode.KEY_O] = 'O';
        TO_STRING_MAP[KeyCode.KEY_P] = 'P';
        TO_STRING_MAP[KeyCode.KEY_Q] = 'Q';
        TO_STRING_MAP[KeyCode.KEY_R] = 'R';
        TO_STRING_MAP[KeyCode.KEY_S] = 'S';
        TO_STRING_MAP[KeyCode.KEY_T] = 'T';
        TO_STRING_MAP[KeyCode.KEY_U] = 'U';
        TO_STRING_MAP[KeyCode.KEY_V] = 'V';
        TO_STRING_MAP[KeyCode.KEY_W] = 'W';
        TO_STRING_MAP[KeyCode.KEY_X] = 'X';
        TO_STRING_MAP[KeyCode.KEY_Y] = 'Y';
        TO_STRING_MAP[KeyCode.KEY_Z] = 'Z';
        TO_STRING_MAP[KeyCode.ContextMenu] = 'ContextMenu';
        TO_STRING_MAP[KeyCode.F1] = 'F1';
        TO_STRING_MAP[KeyCode.F2] = 'F2';
        TO_STRING_MAP[KeyCode.F3] = 'F3';
        TO_STRING_MAP[KeyCode.F4] = 'F4';
        TO_STRING_MAP[KeyCode.F5] = 'F5';
        TO_STRING_MAP[KeyCode.F6] = 'F6';
        TO_STRING_MAP[KeyCode.F7] = 'F7';
        TO_STRING_MAP[KeyCode.F8] = 'F8';
        TO_STRING_MAP[KeyCode.F9] = 'F9';
        TO_STRING_MAP[KeyCode.F10] = 'F10';
        TO_STRING_MAP[KeyCode.F11] = 'F11';
        TO_STRING_MAP[KeyCode.F12] = 'F12';
        TO_STRING_MAP[KeyCode.F13] = 'F13';
        TO_STRING_MAP[KeyCode.F14] = 'F14';
        TO_STRING_MAP[KeyCode.F15] = 'F15';
        TO_STRING_MAP[KeyCode.F16] = 'F16';
        TO_STRING_MAP[KeyCode.F17] = 'F17';
        TO_STRING_MAP[KeyCode.F18] = 'F18';
        TO_STRING_MAP[KeyCode.F19] = 'F19';
        TO_STRING_MAP[KeyCode.NumLock] = 'NumLock';
        TO_STRING_MAP[KeyCode.ScrollLock] = 'ScrollLock';
        TO_STRING_MAP[KeyCode.US_SEMICOLON] = ';';
        TO_STRING_MAP[KeyCode.US_EQUAL] = '=';
        TO_STRING_MAP[KeyCode.US_COMMA] = ',';
        TO_STRING_MAP[KeyCode.US_MINUS] = '-';
        TO_STRING_MAP[KeyCode.US_DOT] = '.';
        TO_STRING_MAP[KeyCode.US_SLASH] = '/';
        TO_STRING_MAP[KeyCode.US_BACKTICK] = '`';
        TO_STRING_MAP[KeyCode.US_OPEN_SQUARE_BRACKET] = '[';
        TO_STRING_MAP[KeyCode.US_BACKSLASH] = '\\';
        TO_STRING_MAP[KeyCode.US_CLOSE_SQUARE_BRACKET] = ']';
        TO_STRING_MAP[KeyCode.US_QUOTE] = '\'';
        TO_STRING_MAP[KeyCode.OEM_8] = 'OEM_8';
        TO_STRING_MAP[KeyCode.OEM_102] = 'OEM_102';
        TO_STRING_MAP[KeyCode.NUMPAD_0] = 'NumPad0';
        TO_STRING_MAP[KeyCode.NUMPAD_1] = 'NumPad1';
        TO_STRING_MAP[KeyCode.NUMPAD_2] = 'NumPad2';
        TO_STRING_MAP[KeyCode.NUMPAD_3] = 'NumPad3';
        TO_STRING_MAP[KeyCode.NUMPAD_4] = 'NumPad4';
        TO_STRING_MAP[KeyCode.NUMPAD_5] = 'NumPad5';
        TO_STRING_MAP[KeyCode.NUMPAD_6] = 'NumPad6';
        TO_STRING_MAP[KeyCode.NUMPAD_7] = 'NumPad7';
        TO_STRING_MAP[KeyCode.NUMPAD_8] = 'NumPad8';
        TO_STRING_MAP[KeyCode.NUMPAD_9] = 'NumPad9';
        TO_STRING_MAP[KeyCode.NUMPAD_MULTIPLY] = 'NumPad_Multiply';
        TO_STRING_MAP[KeyCode.NUMPAD_ADD] = 'NumPad_Add';
        TO_STRING_MAP[KeyCode.NUMPAD_SEPARATOR] = 'NumPad_Separator';
        TO_STRING_MAP[KeyCode.NUMPAD_SUBTRACT] = 'NumPad_Subtract';
        TO_STRING_MAP[KeyCode.NUMPAD_DECIMAL] = 'NumPad_Decimal';
        TO_STRING_MAP[KeyCode.NUMPAD_DIVIDE] = 'NumPad_Divide';
        // for (let i = 0; i < KeyCode.MAX_VALUE; i++) {
        // 	if (!TO_STRING_MAP[i]) {
        // 		console.warn('Missing string representation for ' + KeyCode[i]);
        // 	}
        // }
    }, function (FROM_STRING_MAP) {
        FROM_STRING_MAP['\r'] = KeyCode.Enter;
    });
    var USER_SETTINGS = createMapping(function (TO_USER_SETTINGS_MAP) {
        for (var i = 0, len = STRING._fromKeyCode.length; i < len; i++) {
            TO_USER_SETTINGS_MAP[i] = STRING._fromKeyCode[i];
        }
        TO_USER_SETTINGS_MAP[KeyCode.LeftArrow] = 'Left';
        TO_USER_SETTINGS_MAP[KeyCode.UpArrow] = 'Up';
        TO_USER_SETTINGS_MAP[KeyCode.RightArrow] = 'Right';
        TO_USER_SETTINGS_MAP[KeyCode.DownArrow] = 'Down';
    }, function (FROM_USER_SETTINGS_MAP) {
        FROM_USER_SETTINGS_MAP['OEM_1'] = KeyCode.US_SEMICOLON;
        FROM_USER_SETTINGS_MAP['OEM_PLUS'] = KeyCode.US_EQUAL;
        FROM_USER_SETTINGS_MAP['OEM_COMMA'] = KeyCode.US_COMMA;
        FROM_USER_SETTINGS_MAP['OEM_MINUS'] = KeyCode.US_MINUS;
        FROM_USER_SETTINGS_MAP['OEM_PERIOD'] = KeyCode.US_DOT;
        FROM_USER_SETTINGS_MAP['OEM_2'] = KeyCode.US_SLASH;
        FROM_USER_SETTINGS_MAP['OEM_3'] = KeyCode.US_BACKTICK;
        FROM_USER_SETTINGS_MAP['OEM_4'] = KeyCode.US_OPEN_SQUARE_BRACKET;
        FROM_USER_SETTINGS_MAP['OEM_5'] = KeyCode.US_BACKSLASH;
        FROM_USER_SETTINGS_MAP['OEM_6'] = KeyCode.US_CLOSE_SQUARE_BRACKET;
        FROM_USER_SETTINGS_MAP['OEM_7'] = KeyCode.US_QUOTE;
        FROM_USER_SETTINGS_MAP['OEM_8'] = KeyCode.OEM_8;
        FROM_USER_SETTINGS_MAP['OEM_102'] = KeyCode.OEM_102;
    });
    var KeyCode;
    (function (KeyCode) {
        function toString(key) {
            return STRING.fromKeyCode(key);
        }
        KeyCode.toString = toString;
        function fromString(key) {
            return STRING.toKeyCode(key);
        }
        KeyCode.fromString = fromString;
    })(KeyCode = exports.KeyCode || (exports.KeyCode = {}));
    // Binary encoding strategy:
    // 15:  1 bit for ctrlCmd
    // 14:  1 bit for shift
    // 13:  1 bit for alt
    // 12:  1 bit for winCtrl
    //  0: 12 bits for keyCode (up to a maximum keyCode of 4096. Given we have 83 at this point thats good enough)
    var BIN_CTRLCMD_MASK = 1 << 15;
    var BIN_SHIFT_MASK = 1 << 14;
    var BIN_ALT_MASK = 1 << 13;
    var BIN_WINCTRL_MASK = 1 << 12;
    var BIN_KEYCODE_MASK = 0x00000fff;
    var BinaryKeybindings = (function () {
        function BinaryKeybindings() {
        }
        BinaryKeybindings.extractFirstPart = function (keybinding) {
            return keybinding & 0x0000ffff;
        };
        BinaryKeybindings.extractChordPart = function (keybinding) {
            return (keybinding >> 16) & 0x0000ffff;
        };
        BinaryKeybindings.hasChord = function (keybinding) {
            return (this.extractChordPart(keybinding) !== 0);
        };
        BinaryKeybindings.hasCtrlCmd = function (keybinding) {
            return (keybinding & BIN_CTRLCMD_MASK ? true : false);
        };
        BinaryKeybindings.hasShift = function (keybinding) {
            return (keybinding & BIN_SHIFT_MASK ? true : false);
        };
        BinaryKeybindings.hasAlt = function (keybinding) {
            return (keybinding & BIN_ALT_MASK ? true : false);
        };
        BinaryKeybindings.hasWinCtrl = function (keybinding) {
            return (keybinding & BIN_WINCTRL_MASK ? true : false);
        };
        BinaryKeybindings.extractKeyCode = function (keybinding) {
            return (keybinding & BIN_KEYCODE_MASK);
        };
        return BinaryKeybindings;
    }());
    exports.BinaryKeybindings = BinaryKeybindings;
    var KeyMod = (function () {
        function KeyMod() {
        }
        KeyMod.chord = function (firstPart, secondPart) {
            return firstPart | ((secondPart & 0x0000ffff) << 16);
        };
        KeyMod.CtrlCmd = BIN_CTRLCMD_MASK;
        KeyMod.Shift = BIN_SHIFT_MASK;
        KeyMod.Alt = BIN_ALT_MASK;
        KeyMod.WinCtrl = BIN_WINCTRL_MASK;
        return KeyMod;
    }());
    exports.KeyMod = KeyMod;
    /**
     * A set of usual keybindings that can be reused in code
     */
    var CommonKeybindings = (function () {
        function CommonKeybindings() {
        }
        CommonKeybindings.ENTER = KeyCode.Enter;
        CommonKeybindings.SHIFT_ENTER = KeyMod.Shift | KeyCode.Enter;
        CommonKeybindings.CTRLCMD_ENTER = KeyMod.CtrlCmd | KeyCode.Enter;
        CommonKeybindings.WINCTRL_ENTER = KeyMod.WinCtrl | KeyCode.Enter;
        CommonKeybindings.TAB = KeyCode.Tab;
        CommonKeybindings.SHIFT_TAB = KeyMod.Shift | KeyCode.Tab;
        CommonKeybindings.ESCAPE = KeyCode.Escape;
        CommonKeybindings.SPACE = KeyCode.Space;
        CommonKeybindings.DELETE = KeyCode.Delete;
        CommonKeybindings.SHIFT_DELETE = KeyMod.Shift | KeyCode.Delete;
        CommonKeybindings.CTRLCMD_BACKSPACE = KeyMod.CtrlCmd | KeyCode.Backspace;
        CommonKeybindings.UP_ARROW = KeyCode.UpArrow;
        CommonKeybindings.SHIFT_UP_ARROW = KeyMod.Shift | KeyCode.UpArrow;
        CommonKeybindings.CTRLCMD_UP_ARROW = KeyMod.CtrlCmd | KeyCode.UpArrow;
        CommonKeybindings.DOWN_ARROW = KeyCode.DownArrow;
        CommonKeybindings.SHIFT_DOWN_ARROW = KeyMod.Shift | KeyCode.DownArrow;
        CommonKeybindings.CTRLCMD_DOWN_ARROW = KeyMod.CtrlCmd | KeyCode.DownArrow;
        CommonKeybindings.LEFT_ARROW = KeyCode.LeftArrow;
        CommonKeybindings.RIGHT_ARROW = KeyCode.RightArrow;
        CommonKeybindings.HOME = KeyCode.Home;
        CommonKeybindings.END = KeyCode.End;
        CommonKeybindings.PAGE_UP = KeyCode.PageUp;
        CommonKeybindings.SHIFT_PAGE_UP = KeyMod.Shift | KeyCode.PageUp;
        CommonKeybindings.PAGE_DOWN = KeyCode.PageDown;
        CommonKeybindings.SHIFT_PAGE_DOWN = KeyMod.Shift | KeyCode.PageDown;
        CommonKeybindings.F2 = KeyCode.F2;
        CommonKeybindings.CTRLCMD_S = KeyMod.CtrlCmd | KeyCode.KEY_S;
        CommonKeybindings.CTRLCMD_C = KeyMod.CtrlCmd | KeyCode.KEY_C;
        CommonKeybindings.CTRLCMD_V = KeyMod.CtrlCmd | KeyCode.KEY_V;
        return CommonKeybindings;
    }());
    exports.CommonKeybindings = CommonKeybindings;
    var Keybinding = (function () {
        function Keybinding(keybinding) {
            this.value = keybinding;
        }
        /**
         * Format the binding to a format appropiate for rendering in the UI
         */
        Keybinding._toUSLabel = function (value, Platform) {
            return _asString(value, (Platform.isMacintosh ? MacUIKeyLabelProvider.INSTANCE : ClassicUIKeyLabelProvider.INSTANCE), Platform);
        };
        /**
         * Format the binding to a format appropiate for placing in an aria-label.
         */
        Keybinding._toUSAriaLabel = function (value, Platform) {
            return _asString(value, AriaKeyLabelProvider.INSTANCE, Platform);
        };
        /**
         * Format the binding to a format appropiate for rendering in the UI
         */
        Keybinding._toUSHTMLLabel = function (value, Platform) {
            return _asHTML(value, (Platform.isMacintosh ? MacUIKeyLabelProvider.INSTANCE : ClassicUIKeyLabelProvider.INSTANCE), Platform);
        };
        /**
         * Format the binding to a format appropiate for rendering in the UI
         */
        Keybinding._toCustomLabel = function (value, labelProvider, Platform) {
            return _asString(value, labelProvider, Platform);
        };
        /**
         * Format the binding to a format appropiate for rendering in the UI
         */
        Keybinding._toCustomHTMLLabel = function (value, labelProvider, Platform) {
            return _asHTML(value, labelProvider, Platform);
        };
        /**
         * This prints the binding in a format suitable for electron's accelerators.
         * See https://github.com/electron/electron/blob/master/docs/api/accelerator.md
         */
        Keybinding._toElectronAccelerator = function (value, Platform) {
            if (BinaryKeybindings.hasChord(value)) {
                // Electron cannot handle chords
                return null;
            }
            return _asString(value, ElectronAcceleratorLabelProvider.INSTANCE, Platform);
        };
        Keybinding.getUserSettingsKeybindingRegex = function () {
            if (!this._cachedKeybindingRegex) {
                var numpadKey = 'numpad(0|1|2|3|4|5|6|7|8|9|_multiply|_add|_subtract|_decimal|_divide|_separator)';
                var oemKey = '`|\\-|=|\\[|\\]|\\\\\\\\|;|\'|,|\\.|\\/|oem_8|oem_102';
                var specialKey = 'left|up|right|down|pageup|pagedown|end|home|tab|enter|escape|space|backspace|delete|pausebreak|capslock|insert|contextmenu|numlock|scrolllock';
                var casualKey = '[a-z]|[0-9]|f(1|2|3|4|5|6|7|8|9|10|11|12|13|14|15|16|17|18|19)';
                var key = '((' + [numpadKey, oemKey, specialKey, casualKey].join(')|(') + '))';
                var mod = '((ctrl|shift|alt|cmd|win|meta)\\+)*';
                var keybinding = '(' + mod + key + ')';
                this._cachedKeybindingRegex = '"\\s*(' + keybinding + '(\\s+' + keybinding + ')?' + ')\\s*"';
            }
            return this._cachedKeybindingRegex;
        };
        /**
         * Format the binding to a format appropiate for the user settings file.
         */
        Keybinding.toUserSettingsLabel = function (value, Platform) {
            if (Platform === void 0) { Platform = defaultPlatform; }
            var result = _asString(value, UserSettingsKeyLabelProvider.INSTANCE, Platform);
            result = result.toLowerCase();
            if (Platform.isMacintosh) {
                result = result.replace(/meta/g, 'cmd');
            }
            else if (Platform.isWindows) {
                result = result.replace(/meta/g, 'win');
            }
            return result;
        };
        Keybinding.fromUserSettingsLabel = function (input, Platform) {
            if (Platform === void 0) { Platform = defaultPlatform; }
            if (!input) {
                return null;
            }
            input = input.toLowerCase().trim();
            var ctrlCmd = false, shift = false, alt = false, winCtrl = false, key = '';
            while (/^(ctrl|shift|alt|meta|win|cmd)(\+|\-)/.test(input)) {
                if (/^ctrl(\+|\-)/.test(input)) {
                    if (Platform.isMacintosh) {
                        winCtrl = true;
                    }
                    else {
                        ctrlCmd = true;
                    }
                    input = input.substr('ctrl-'.length);
                }
                if (/^shift(\+|\-)/.test(input)) {
                    shift = true;
                    input = input.substr('shift-'.length);
                }
                if (/^alt(\+|\-)/.test(input)) {
                    alt = true;
                    input = input.substr('alt-'.length);
                }
                if (/^meta(\+|\-)/.test(input)) {
                    if (Platform.isMacintosh) {
                        ctrlCmd = true;
                    }
                    else {
                        winCtrl = true;
                    }
                    input = input.substr('meta-'.length);
                }
                if (/^win(\+|\-)/.test(input)) {
                    if (Platform.isMacintosh) {
                        ctrlCmd = true;
                    }
                    else {
                        winCtrl = true;
                    }
                    input = input.substr('win-'.length);
                }
                if (/^cmd(\+|\-)/.test(input)) {
                    if (Platform.isMacintosh) {
                        ctrlCmd = true;
                    }
                    else {
                        winCtrl = true;
                    }
                    input = input.substr('cmd-'.length);
                }
            }
            var chord = 0;
            var firstSpaceIdx = input.indexOf(' ');
            if (firstSpaceIdx > 0) {
                key = input.substring(0, firstSpaceIdx);
                chord = Keybinding.fromUserSettingsLabel(input.substring(firstSpaceIdx), Platform);
            }
            else {
                key = input;
            }
            var keyCode = USER_SETTINGS.toKeyCode(key);
            var result = 0;
            if (ctrlCmd) {
                result |= KeyMod.CtrlCmd;
            }
            if (shift) {
                result |= KeyMod.Shift;
            }
            if (alt) {
                result |= KeyMod.Alt;
            }
            if (winCtrl) {
                result |= KeyMod.WinCtrl;
            }
            result |= keyCode;
            return KeyMod.chord(result, chord);
        };
        Keybinding.prototype.hasCtrlCmd = function () {
            return BinaryKeybindings.hasCtrlCmd(this.value);
        };
        Keybinding.prototype.hasShift = function () {
            return BinaryKeybindings.hasShift(this.value);
        };
        Keybinding.prototype.hasAlt = function () {
            return BinaryKeybindings.hasAlt(this.value);
        };
        Keybinding.prototype.hasWinCtrl = function () {
            return BinaryKeybindings.hasWinCtrl(this.value);
        };
        Keybinding.prototype.extractKeyCode = function () {
            return BinaryKeybindings.extractKeyCode(this.value);
        };
        /**
         * Format the binding to a format appropiate for rendering in the UI
         */
        Keybinding.prototype._toUSLabel = function (Platform) {
            if (Platform === void 0) { Platform = defaultPlatform; }
            return Keybinding._toUSLabel(this.value, Platform);
        };
        /**
         * Format the binding to a format appropiate for placing in an aria-label.
         */
        Keybinding.prototype._toUSAriaLabel = function (Platform) {
            if (Platform === void 0) { Platform = defaultPlatform; }
            return Keybinding._toUSAriaLabel(this.value, Platform);
        };
        /**
         * Format the binding to a format appropiate for rendering in the UI
         */
        Keybinding.prototype._toUSHTMLLabel = function (Platform) {
            if (Platform === void 0) { Platform = defaultPlatform; }
            return Keybinding._toUSHTMLLabel(this.value, Platform);
        };
        /**
         * Format the binding to a format appropiate for rendering in the UI
         */
        Keybinding.prototype.toCustomLabel = function (labelProvider, Platform) {
            if (Platform === void 0) { Platform = defaultPlatform; }
            return Keybinding._toCustomLabel(this.value, labelProvider, Platform);
        };
        /**
         * Format the binding to a format appropiate for rendering in the UI
         */
        Keybinding.prototype.toCustomHTMLLabel = function (labelProvider, Platform) {
            if (Platform === void 0) { Platform = defaultPlatform; }
            return Keybinding._toCustomHTMLLabel(this.value, labelProvider, Platform);
        };
        /**
         * This prints the binding in a format suitable for electron's accelerators.
         * See https://github.com/electron/electron/blob/master/docs/api/accelerator.md
         */
        Keybinding.prototype._toElectronAccelerator = function (Platform) {
            if (Platform === void 0) { Platform = defaultPlatform; }
            return Keybinding._toElectronAccelerator(this.value, Platform);
        };
        /**
         * Format the binding to a format appropiate for the user settings file.
         */
        Keybinding.prototype.toUserSettingsLabel = function (Platform) {
            if (Platform === void 0) { Platform = defaultPlatform; }
            return Keybinding.toUserSettingsLabel(this.value, Platform);
        };
        Keybinding._cachedKeybindingRegex = null;
        return Keybinding;
    }());
    exports.Keybinding = Keybinding;
    /**
     * Print for Electron
     */
    var ElectronAcceleratorLabelProvider = (function () {
        function ElectronAcceleratorLabelProvider() {
            this.ctrlKeyLabel = 'Ctrl';
            this.shiftKeyLabel = 'Shift';
            this.altKeyLabel = 'Alt';
            this.cmdKeyLabel = 'Cmd';
            this.windowsKeyLabel = 'Super';
            this.modifierSeparator = '+';
        }
        ElectronAcceleratorLabelProvider.prototype.getLabelForKey = function (keyCode) {
            switch (keyCode) {
                case KeyCode.UpArrow:
                    return 'Up';
                case KeyCode.DownArrow:
                    return 'Down';
                case KeyCode.LeftArrow:
                    return 'Left';
                case KeyCode.RightArrow:
                    return 'Right';
            }
            return KeyCode.toString(keyCode);
        };
        ElectronAcceleratorLabelProvider.INSTANCE = new ElectronAcceleratorLabelProvider();
        return ElectronAcceleratorLabelProvider;
    }());
    exports.ElectronAcceleratorLabelProvider = ElectronAcceleratorLabelProvider;
    /**
     * Print for Mac UI
     */
    var MacUIKeyLabelProvider = (function () {
        function MacUIKeyLabelProvider() {
            this.ctrlKeyLabel = '\u2303';
            this.shiftKeyLabel = '\u21E7';
            this.altKeyLabel = '\u2325';
            this.cmdKeyLabel = '\u2318';
            this.windowsKeyLabel = nls.localize(0, null);
            this.modifierSeparator = '';
        }
        MacUIKeyLabelProvider.prototype.getLabelForKey = function (keyCode) {
            switch (keyCode) {
                case KeyCode.LeftArrow:
                    return MacUIKeyLabelProvider.leftArrowUnicodeLabel;
                case KeyCode.UpArrow:
                    return MacUIKeyLabelProvider.upArrowUnicodeLabel;
                case KeyCode.RightArrow:
                    return MacUIKeyLabelProvider.rightArrowUnicodeLabel;
                case KeyCode.DownArrow:
                    return MacUIKeyLabelProvider.downArrowUnicodeLabel;
            }
            return KeyCode.toString(keyCode);
        };
        MacUIKeyLabelProvider.INSTANCE = new MacUIKeyLabelProvider();
        MacUIKeyLabelProvider.leftArrowUnicodeLabel = String.fromCharCode(8592);
        MacUIKeyLabelProvider.upArrowUnicodeLabel = String.fromCharCode(8593);
        MacUIKeyLabelProvider.rightArrowUnicodeLabel = String.fromCharCode(8594);
        MacUIKeyLabelProvider.downArrowUnicodeLabel = String.fromCharCode(8595);
        return MacUIKeyLabelProvider;
    }());
    exports.MacUIKeyLabelProvider = MacUIKeyLabelProvider;
    /**
     * Aria label provider for Mac.
     */
    var AriaKeyLabelProvider = (function () {
        function AriaKeyLabelProvider() {
            this.ctrlKeyLabel = nls.localize(1, null);
            this.shiftKeyLabel = nls.localize(2, null);
            this.altKeyLabel = nls.localize(3, null);
            this.cmdKeyLabel = nls.localize(4, null);
            this.windowsKeyLabel = nls.localize(5, null);
            this.modifierSeparator = '+';
        }
        AriaKeyLabelProvider.prototype.getLabelForKey = function (keyCode) {
            return KeyCode.toString(keyCode);
        };
        AriaKeyLabelProvider.INSTANCE = new MacUIKeyLabelProvider();
        return AriaKeyLabelProvider;
    }());
    exports.AriaKeyLabelProvider = AriaKeyLabelProvider;
    /**
     * Print for Windows, Linux UI
     */
    var ClassicUIKeyLabelProvider = (function () {
        function ClassicUIKeyLabelProvider() {
            this.ctrlKeyLabel = nls.localize(6, null);
            this.shiftKeyLabel = nls.localize(7, null);
            this.altKeyLabel = nls.localize(8, null);
            this.cmdKeyLabel = nls.localize(9, null);
            this.windowsKeyLabel = nls.localize(10, null);
            this.modifierSeparator = '+';
        }
        ClassicUIKeyLabelProvider.prototype.getLabelForKey = function (keyCode) {
            return KeyCode.toString(keyCode);
        };
        ClassicUIKeyLabelProvider.INSTANCE = new ClassicUIKeyLabelProvider();
        return ClassicUIKeyLabelProvider;
    }());
    exports.ClassicUIKeyLabelProvider = ClassicUIKeyLabelProvider;
    /**
     * Print for the user settings file.
     */
    var UserSettingsKeyLabelProvider = (function () {
        function UserSettingsKeyLabelProvider() {
            this.ctrlKeyLabel = 'Ctrl';
            this.shiftKeyLabel = 'Shift';
            this.altKeyLabel = 'Alt';
            this.cmdKeyLabel = 'Meta';
            this.windowsKeyLabel = 'Meta';
            this.modifierSeparator = '+';
        }
        UserSettingsKeyLabelProvider.prototype.getLabelForKey = function (keyCode) {
            return USER_SETTINGS.fromKeyCode(keyCode);
        };
        UserSettingsKeyLabelProvider.INSTANCE = new UserSettingsKeyLabelProvider();
        return UserSettingsKeyLabelProvider;
    }());
    function _asString(keybinding, labelProvider, Platform) {
        var result = [], ctrlCmd = BinaryKeybindings.hasCtrlCmd(keybinding), shift = BinaryKeybindings.hasShift(keybinding), alt = BinaryKeybindings.hasAlt(keybinding), winCtrl = BinaryKeybindings.hasWinCtrl(keybinding), keyCode = BinaryKeybindings.extractKeyCode(keybinding);
        var keyLabel = labelProvider.getLabelForKey(keyCode);
        if (!keyLabel) {
            // cannot trigger this key code under this kb layout
            return '';
        }
        // translate modifier keys: Ctrl-Shift-Alt-Meta
        if ((ctrlCmd && !Platform.isMacintosh) || (winCtrl && Platform.isMacintosh)) {
            result.push(labelProvider.ctrlKeyLabel);
        }
        if (shift) {
            result.push(labelProvider.shiftKeyLabel);
        }
        if (alt) {
            result.push(labelProvider.altKeyLabel);
        }
        if (ctrlCmd && Platform.isMacintosh) {
            result.push(labelProvider.cmdKeyLabel);
        }
        if (winCtrl && !Platform.isMacintosh) {
            result.push(labelProvider.windowsKeyLabel);
        }
        // the actual key
        result.push(keyLabel);
        var actualResult = result.join(labelProvider.modifierSeparator);
        if (BinaryKeybindings.hasChord(keybinding)) {
            return actualResult + ' ' + _asString(BinaryKeybindings.extractChordPart(keybinding), labelProvider, Platform);
        }
        return actualResult;
    }
    function _pushKey(result, str) {
        if (result.length > 0) {
            result.push({
                tagName: 'span',
                text: '+'
            });
        }
        result.push({
            tagName: 'span',
            className: 'monaco-kbkey',
            text: str
        });
    }
    function _asHTML(keybinding, labelProvider, Platform, isChord) {
        if (isChord === void 0) { isChord = false; }
        var result = [], ctrlCmd = BinaryKeybindings.hasCtrlCmd(keybinding), shift = BinaryKeybindings.hasShift(keybinding), alt = BinaryKeybindings.hasAlt(keybinding), winCtrl = BinaryKeybindings.hasWinCtrl(keybinding), keyCode = BinaryKeybindings.extractKeyCode(keybinding);
        var keyLabel = labelProvider.getLabelForKey(keyCode);
        if (!keyLabel) {
            // cannot trigger this key code under this kb layout
            return [];
        }
        // translate modifier keys: Ctrl-Shift-Alt-Meta
        if ((ctrlCmd && !Platform.isMacintosh) || (winCtrl && Platform.isMacintosh)) {
            _pushKey(result, labelProvider.ctrlKeyLabel);
        }
        if (shift) {
            _pushKey(result, labelProvider.shiftKeyLabel);
        }
        if (alt) {
            _pushKey(result, labelProvider.altKeyLabel);
        }
        if (ctrlCmd && Platform.isMacintosh) {
            _pushKey(result, labelProvider.cmdKeyLabel);
        }
        if (winCtrl && !Platform.isMacintosh) {
            _pushKey(result, labelProvider.windowsKeyLabel);
        }
        // the actual key
        _pushKey(result, keyLabel);
        var chordTo = null;
        if (BinaryKeybindings.hasChord(keybinding)) {
            chordTo = _asHTML(BinaryKeybindings.extractChordPart(keybinding), labelProvider, Platform, true);
            result.push({
                tagName: 'span',
                text: ' '
            });
            result = result.concat(chordTo);
        }
        if (isChord) {
            return result;
        }
        return [{
                tagName: 'span',
                className: 'monaco-kb',
                children: result
            }];
    }
});

define(__m[35], __M([11,12]), function(nls, data) { return nls.create("vs/base/common/severity", data); });
define(__m[32], __M([1,0,35,4]), function (require, exports, nls, strings) {
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

define(__m[30], __M([1,0,14,34,9,7,28,2,15,32,13]), function (require, exports, event_1, keyCodes_1, position_1, range_1, selection_1, winjs_base_1, cancellation_1, severity_1, uri_1) {
    /*---------------------------------------------------------------------------------------------
     *  Copyright (c) Microsoft Corporation. All rights reserved.
     *  Licensed under the MIT License. See License.txt in the project root for license information.
     *--------------------------------------------------------------------------------------------*/
    'use strict';
    function createMonacoBaseAPI() {
        return {
            editor: undefined,
            languages: undefined,
            CancellationTokenSource: cancellation_1.CancellationTokenSource,
            Emitter: event_1.Emitter,
            KeyCode: keyCodes_1.KeyCode,
            KeyMod: keyCodes_1.KeyMod,
            Position: position_1.Position,
            Range: range_1.Range,
            Selection: selection_1.Selection,
            SelectionDirection: selection_1.SelectionDirection,
            Severity: severity_1.default,
            Promise: winjs_base_1.TPromise,
            Uri: uri_1.default
        };
    }
    exports.createMonacoBaseAPI = createMonacoBaseAPI;
});

/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Microsoft Corporation. All rights reserved.
 *  Licensed under the MIT License. See License.txt in the project root for license information.
 *--------------------------------------------------------------------------------------------*/





define(__m[38], __M([1,0,13,2,7,25,29,21,18,16,17,30]), function (require, exports, uri_1, winjs_base_1, range_1, filters_1, diffComputer_1, mirrorModel2_1, linkComputer_1, inplaceReplaceSupport_1, wordHelper_1, standaloneBase_1) {
    'use strict';
    /**
     * @internal
     */
    var MirrorModel = (function (_super) {
        __extends(MirrorModel, _super);
        function MirrorModel() {
            _super.apply(this, arguments);
        }
        Object.defineProperty(MirrorModel.prototype, "uri", {
            get: function () {
                return this._uri;
            },
            enumerable: true,
            configurable: true
        });
        Object.defineProperty(MirrorModel.prototype, "version", {
            get: function () {
                return this._versionId;
            },
            enumerable: true,
            configurable: true
        });
        MirrorModel.prototype.getValue = function () {
            return this.getText();
        };
        MirrorModel.prototype.getLinesContent = function () {
            return this._lines.slice(0);
        };
        MirrorModel.prototype.getLineCount = function () {
            return this._lines.length;
        };
        MirrorModel.prototype.getLineContent = function (lineNumber) {
            return this._lines[lineNumber - 1];
        };
        MirrorModel.prototype.getWordAtPosition = function (position, wordDefinition) {
            var wordAtText = wordHelper_1.getWordAtText(position.column, wordHelper_1.ensureValidWordDefinition(wordDefinition), this._lines[position.lineNumber - 1], 0);
            if (wordAtText) {
                return new range_1.Range(position.lineNumber, wordAtText.startColumn, position.lineNumber, wordAtText.endColumn);
            }
            return null;
        };
        MirrorModel.prototype.getWordUntilPosition = function (position, wordDefinition) {
            var wordAtPosition = this.getWordAtPosition(position, wordDefinition);
            if (!wordAtPosition) {
                return {
                    word: '',
                    startColumn: position.column,
                    endColumn: position.column
                };
            }
            return {
                word: this._lines[position.lineNumber - 1].substring(wordAtPosition.startColumn - 1, position.column - 1),
                startColumn: wordAtPosition.startColumn,
                endColumn: position.column
            };
        };
        MirrorModel.prototype._getAllWords = function (wordDefinition) {
            var _this = this;
            var result = [];
            this._lines.forEach(function (line) {
                _this._wordenize(line, wordDefinition).forEach(function (info) {
                    result.push(line.substring(info.start, info.end));
                });
            });
            return result;
        };
        MirrorModel.prototype.getAllUniqueWords = function (wordDefinition, skipWordOnce) {
            var foundSkipWord = false;
            var uniqueWords = {};
            return this._getAllWords(wordDefinition).filter(function (word) {
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
        // TODO@Joh, TODO@Alex - remove these and make sure the super-things work
        MirrorModel.prototype._wordenize = function (content, wordDefinition) {
            var result = [];
            var match;
            while (match = wordDefinition.exec(content)) {
                if (match[0].length === 0) {
                    // it did match the empty string
                    break;
                }
                result.push({ start: match.index, end: match.index + match[0].length });
            }
            return result;
        };
        MirrorModel.prototype.getValueInRange = function (range) {
            if (range.startLineNumber === range.endLineNumber) {
                return this._lines[range.startLineNumber - 1].substring(range.startColumn - 1, range.endColumn - 1);
            }
            var lineEnding = this._eol, startLineIndex = range.startLineNumber - 1, endLineIndex = range.endLineNumber - 1, resultLines = [];
            resultLines.push(this._lines[startLineIndex].substring(range.startColumn - 1));
            for (var i = startLineIndex + 1; i < endLineIndex; i++) {
                resultLines.push(this._lines[i]);
            }
            resultLines.push(this._lines[endLineIndex].substring(0, range.endColumn - 1));
            return resultLines.join(lineEnding);
        };
        return MirrorModel;
    }(mirrorModel2_1.MirrorModel2));
    exports.MirrorModel = MirrorModel;
    /**
     * @internal
     */
    var BaseEditorSimpleWorker = (function () {
        function BaseEditorSimpleWorker() {
            this._foreignModule = null;
        }
        // ---- BEGIN diff --------------------------------------------------------------------------
        BaseEditorSimpleWorker.prototype.computeDiff = function (originalUrl, modifiedUrl, ignoreTrimWhitespace) {
            var original = this._getModel(originalUrl);
            var modified = this._getModel(modifiedUrl);
            if (!original || !modified) {
                return null;
            }
            var originalLines = original.getLinesContent();
            var modifiedLines = modified.getLinesContent();
            var diffComputer = new diffComputer_1.DiffComputer(originalLines, modifiedLines, {
                shouldPostProcessCharChanges: true,
                shouldIgnoreTrimWhitespace: ignoreTrimWhitespace,
                shouldConsiderTrimWhitespaceInEmptyCase: true
            });
            return winjs_base_1.TPromise.as(diffComputer.computeDiff());
        };
        BaseEditorSimpleWorker.prototype.computeDirtyDiff = function (originalUrl, modifiedUrl, ignoreTrimWhitespace) {
            var original = this._getModel(originalUrl);
            var modified = this._getModel(modifiedUrl);
            if (!original || !modified) {
                return null;
            }
            var originalLines = original.getLinesContent();
            var modifiedLines = modified.getLinesContent();
            var diffComputer = new diffComputer_1.DiffComputer(originalLines, modifiedLines, {
                shouldPostProcessCharChanges: false,
                shouldIgnoreTrimWhitespace: ignoreTrimWhitespace,
                shouldConsiderTrimWhitespaceInEmptyCase: false
            });
            return winjs_base_1.TPromise.as(diffComputer.computeDiff());
        };
        // ---- END diff --------------------------------------------------------------------------
        BaseEditorSimpleWorker.prototype.computeLinks = function (modelUrl) {
            var model = this._getModel(modelUrl);
            if (!model) {
                return null;
            }
            return winjs_base_1.TPromise.as(linkComputer_1.computeLinks(model));
        };
        // ---- BEGIN suggest --------------------------------------------------------------------------
        BaseEditorSimpleWorker.prototype.textualSuggest = function (modelUrl, position, wordDef, wordDefFlags) {
            var model = this._getModel(modelUrl);
            if (!model) {
                return null;
            }
            return winjs_base_1.TPromise.as(this._suggestFiltered(model, position, new RegExp(wordDef, wordDefFlags)));
        };
        BaseEditorSimpleWorker.prototype._suggestFiltered = function (model, position, wordDefRegExp) {
            var value = this._suggestUnfiltered(model, position, wordDefRegExp);
            // filter suggestions
            return [{
                    currentWord: value.currentWord,
                    suggestions: value.suggestions.filter(function (element) { return !!filters_1.fuzzyContiguousFilter(value.currentWord, element.label); }),
                    incomplete: value.incomplete
                }];
        };
        BaseEditorSimpleWorker.prototype._suggestUnfiltered = function (model, position, wordDefRegExp) {
            var currentWord = model.getWordUntilPosition(position, wordDefRegExp).word;
            var allWords = model.getAllUniqueWords(wordDefRegExp, currentWord);
            var suggestions = allWords.filter(function (word) {
                return !(/^-?\d*\.?\d/.test(word)); // filter out numbers
            }).map(function (word) {
                return {
                    type: 'text',
                    label: word,
                    codeSnippet: word,
                    noAutoAccept: true
                };
            });
            return {
                currentWord: currentWord,
                suggestions: suggestions
            };
        };
        // ---- END suggest --------------------------------------------------------------------------
        BaseEditorSimpleWorker.prototype.navigateValueSet = function (modelUrl, range, up, wordDef, wordDefFlags) {
            var model = this._getModel(modelUrl);
            if (!model) {
                return null;
            }
            var wordDefRegExp = new RegExp(wordDef, wordDefFlags);
            if (range.startColumn === range.endColumn) {
                range.endColumn += 1;
            }
            var selectionText = model.getValueInRange(range);
            var wordRange = model.getWordAtPosition({ lineNumber: range.startLineNumber, column: range.startColumn }, wordDefRegExp);
            var word = null;
            if (wordRange !== null) {
                word = model.getValueInRange(wordRange);
            }
            var result = inplaceReplaceSupport_1.BasicInplaceReplace.INSTANCE.navigateValueSet(range, selectionText, wordRange, word, up);
            return winjs_base_1.TPromise.as(result);
        };
        // ---- BEGIN foreign module support --------------------------------------------------------------------------
        BaseEditorSimpleWorker.prototype.loadForeignModule = function (moduleId, createData) {
            var _this = this;
            return new winjs_base_1.TPromise(function (c, e) {
                // Use the global require to be sure to get the global config
                self.require([moduleId], function (foreignModule) {
                    var ctx = {
                        getMirrorModels: function () {
                            return _this._getModels();
                        }
                    };
                    _this._foreignModule = foreignModule.create(ctx, createData);
                    var methods = [];
                    for (var prop in _this._foreignModule) {
                        if (typeof _this._foreignModule[prop] === 'function') {
                            methods.push(prop);
                        }
                    }
                    c(methods);
                }, e);
            });
        };
        // foreign method request
        BaseEditorSimpleWorker.prototype.fmr = function (method, args) {
            if (!this._foreignModule || typeof this._foreignModule[method] !== 'function') {
                return winjs_base_1.TPromise.wrapError(new Error('Missing requestHandler or method: ' + method));
            }
            try {
                return winjs_base_1.TPromise.as(this._foreignModule[method].apply(this._foreignModule, args));
            }
            catch (e) {
                return winjs_base_1.TPromise.wrapError(e);
            }
        };
        return BaseEditorSimpleWorker;
    }());
    exports.BaseEditorSimpleWorker = BaseEditorSimpleWorker;
    /**
     * @internal
     */
    var EditorSimpleWorkerImpl = (function (_super) {
        __extends(EditorSimpleWorkerImpl, _super);
        function EditorSimpleWorkerImpl() {
            _super.call(this);
            this._models = Object.create(null);
        }
        EditorSimpleWorkerImpl.prototype.dispose = function () {
            this._models = Object.create(null);
        };
        EditorSimpleWorkerImpl.prototype._getModel = function (uri) {
            return this._models[uri];
        };
        EditorSimpleWorkerImpl.prototype._getModels = function () {
            var _this = this;
            var all = [];
            Object.keys(this._models).forEach(function (key) { return all.push(_this._models[key]); });
            return all;
        };
        EditorSimpleWorkerImpl.prototype.acceptNewModel = function (data) {
            this._models[data.url] = new MirrorModel(uri_1.default.parse(data.url), data.value.lines, data.value.EOL, data.versionId);
        };
        EditorSimpleWorkerImpl.prototype.acceptModelChanged = function (strURL, events) {
            if (!this._models[strURL]) {
                return;
            }
            var model = this._models[strURL];
            model.onEvents(events);
        };
        EditorSimpleWorkerImpl.prototype.acceptRemovedModel = function (strURL) {
            if (!this._models[strURL]) {
                return;
            }
            delete this._models[strURL];
        };
        return EditorSimpleWorkerImpl;
    }(BaseEditorSimpleWorker));
    exports.EditorSimpleWorkerImpl = EditorSimpleWorkerImpl;
    /**
     * Called on the worker side
     * @internal
     */
    function create() {
        return new EditorSimpleWorkerImpl();
    }
    exports.create = create;
    var global = self;
    var isWebWorker = (typeof global.importScripts === 'function');
    if (isWebWorker) {
        global.monaco = standaloneBase_1.createMonacoBaseAPI();
    }
});

}).call(this);
//# sourceMappingURL=simpleWorker.js.map
