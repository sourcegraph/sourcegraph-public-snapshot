"use strict";
var __assign = (this && this.__assign) || function () {
    __assign = Object.assign || function(t) {
        for (var s, i = 1, n = arguments.length; i < n; i++) {
            s = arguments[i];
            for (var p in s) if (Object.prototype.hasOwnProperty.call(s, p))
                t[p] = s[p];
        }
        return t;
    };
    return __assign.apply(this, arguments);
};
var __spreadArray = (this && this.__spreadArray) || function (to, from, pack) {
    if (pack || arguments.length === 2) for (var i = 0, l = from.length, ar; i < l; i++) {
        if (ar || !(i in from)) {
            if (!ar) ar = Array.prototype.slice.call(from, 0, i);
            ar[i] = from[i];
        }
    }
    return to.concat(ar || Array.prototype.slice.call(from));
};
exports.__esModule = true;
exports.fetch = exports.isSearchMatchOfType = exports.getMatchUrl = exports.getCommitMatchUrl = exports.getRepoMatchUrl = exports.getRepoMatchLabel = exports.getFileMatchUrl = exports.getRevision = exports.getRepositoryUrl = exports.aggregateStreamingSearch = exports.search = exports.messageHandlers = exports.observeMessages = exports.switchAggregateSearchResults = exports.emptyAggregateResults = exports.LATEST_VERSION = void 0;
var fetch_event_source_1 = require("@microsoft/fetch-event-source");
/* eslint-disable id-length */
var rxjs_1 = require("rxjs");
var operators_1 = require("rxjs/operators");
var common_1 = require("@sourcegraph/common");
// The latest supported version of our search syntax. Users should never be able to determine the search version.
// The version is set based on the release tag of the instance. Anything before 3.9.0 will not pass a version parameter,
// and will therefore default to V1.
exports.LATEST_VERSION = 'V2';
exports.emptyAggregateResults = {
    state: 'loading',
    results: [],
    filters: [],
    progress: {
        durationMs: 0,
        matchCount: 0,
        skipped: []
    }
};
/**
 * Converts a stream of SearchEvents into AggregateStreamingSearchResults
 */
exports.switchAggregateSearchResults = (0, rxjs_1.pipe)((0, operators_1.materialize)(), (0, operators_1.scan)(function (results, newEvent) {
    var _a;
    switch (newEvent.kind) {
        case 'N': {
            switch ((_a = newEvent.value) === null || _a === void 0 ? void 0 : _a.type) {
                case 'matches':
                    return __assign(__assign({}, results), { 
                        // Matches are additive
                        results: results.results.concat(newEvent.value.data) });
                case 'progress':
                    return __assign(__assign({}, results), { 
                        // Progress updates replace
                        progress: newEvent.value.data });
                case 'filters':
                    return __assign(__assign({}, results), { 
                        // New filter results replace all previous ones
                        filters: newEvent.value.data });
                case 'alert':
                    return __assign(__assign({}, results), { alert: newEvent.value.data });
                default:
                    return results;
            }
        }
        case 'E': {
            // Add the error as an extra skipped item
            var error = (0, common_1.asError)(newEvent.error);
            var errorSkipped = {
                title: 'Error loading results',
                message: error.message,
                reason: 'error',
                severity: 'error'
            };
            return __assign(__assign({}, results), { error: error, progress: __assign(__assign({}, results.progress), { skipped: __spreadArray([errorSkipped], results.progress.skipped, true) }), state: 'error' });
        }
        case 'C':
            return __assign(__assign({}, results), { state: 'complete' });
        default:
            return results;
    }
}, exports.emptyAggregateResults), (0, operators_1.defaultIfEmpty)(exports.emptyAggregateResults));
var observeMessages = function (type, eventSource) {
    return (0, rxjs_1.fromEvent)(eventSource, type).pipe((0, operators_1.map)(function (event) {
        if (!(event instanceof MessageEvent)) {
            throw new TypeError("internal error: expected MessageEvent in streaming search ".concat(type));
        }
        try {
            var parsedData = JSON.parse(event.data);
            return parsedData;
        }
        catch (_a) {
            throw new Error("Could not parse ".concat(type, " message data in streaming search"));
        }
    }), (0, operators_1.map)(function (data) { return ({ type: type, data: data }); }));
};
exports.observeMessages = observeMessages;
var observeMessagesHandler = function (type, eventSource, observer) { return (0, exports.observeMessages)(type, eventSource).subscribe(observer); };
exports.messageHandlers = {
    done: function (type, eventSource, observer) {
        return (0, rxjs_1.fromEvent)(eventSource, type).subscribe(function () {
            observer.complete();
            eventSource.close();
        });
    },
    error: function (type, eventSource, observer) {
        return (0, rxjs_1.fromEvent)(eventSource, type).subscribe(function (event) {
            var error = null;
            if (event instanceof MessageEvent) {
                try {
                    error = JSON.parse(event.data);
                }
                catch (_a) {
                    error = null;
                }
            }
            if ((0, common_1.isErrorLike)(error)) {
                observer.error(error);
            }
            else {
                // The EventSource API can return a DOM event that is not an Error object
                // (e.g. doesn't have the message property), so we need to construct our own here.
                // See https://developer.mozilla.org/en-US/docs/Web/API/EventSource/error_event
                observer.error(new Error('The connection was closed before your search was completed. This may be due to a problem with a firewall, VPN or proxy, or a failure with the Sourcegraph server.'));
            }
            eventSource.close();
        });
    },
    matches: observeMessagesHandler,
    progress: observeMessagesHandler,
    filters: observeMessagesHandler,
    alert: observeMessagesHandler
};
function initiateSearchStream(query, _a, messageHandlers) {
    var version = _a.version, patternType = _a.patternType, caseSensitive = _a.caseSensitive, trace = _a.trace, decorationKinds = _a.decorationKinds, decorationContextLines = _a.decorationContextLines, _b = _a.sourcegraphURL, sourcegraphURL = _b === void 0 ? '' : _b;
    return new rxjs_1.Observable(function (observer) {
        var subscriptions = new rxjs_1.Subscription();
        var parameters = [
            ['q', "".concat(query, " ").concat(caseSensitive ? 'case:yes' : '')],
            ['v', version],
            ['t', patternType],
            ['dl', '0'],
            ['dk', (decorationKinds || ['html']).join('|')],
            ['dc', (decorationContextLines || '1').toString()],
            ['display', '1500'],
        ];
        if (trace) {
            parameters.push(['trace', trace]);
        }
        var parameterEncoded = parameters.map(function (_a) {
            var k = _a[0], v = _a[1];
            return k + '=' + encodeURIComponent(v);
        }).join('&');
        var eventSource = new EventSource("".concat(sourcegraphURL, "/search/stream?").concat(parameterEncoded));
        subscriptions.add(function () { return eventSource.close(); });
        for (var _i = 0, _a = Object.entries(messageHandlers); _i < _a.length; _i++) {
            var _b = _a[_i], eventType = _b[0], handleMessages = _b[1];
            subscriptions.add(handleMessages(eventType, eventSource, observer));
        }
        return function () {
            subscriptions.unsubscribe();
        };
    });
}
/**
 * Initiates a streaming search.
 * This is a type safe wrapper around Sourcegraph's streaming search API (using Server Sent Events). The observable will emit each event returned from the backend.
 *
 * @param queryObservable is an observables that resolves to a query string
 * @param options contains the search query and the necessary context to perform the search (version, patternType, caseSensitive, etc.)
 * @param messageHandlers provide handler functions for each possible `SearchEvent` type
 */
function search(queryObservable, options, messageHandlers) {
    return queryObservable.pipe((0, operators_1.switchMap)(function (query) { return initiateSearchStream(query, options, messageHandlers); }));
}
exports.search = search;
/** Initiates a streaming search with and aggregates the results. */
function aggregateStreamingSearch(queryObservable, options) {
    return search(queryObservable, options, exports.messageHandlers).pipe(exports.switchAggregateSearchResults);
}
exports.aggregateStreamingSearch = aggregateStreamingSearch;
function getRepositoryUrl(repository, branches) {
    var branch = branches === null || branches === void 0 ? void 0 : branches[0];
    var revision = branch ? "@".concat(branch) : '';
    var label = repository + revision;
    return '/' + encodeURI(label);
}
exports.getRepositoryUrl = getRepositoryUrl;
function getRevision(branches, version) {
    var revision = '';
    if (branches) {
        var branch = branches[0];
        if (branch !== '') {
            revision = branch;
        }
    }
    else if (version) {
        revision = version;
    }
    return revision;
}
exports.getRevision = getRevision;
function getFileMatchUrl(fileMatch) {
    var revision = getRevision(fileMatch.branches, fileMatch.commit);
    return "/".concat(fileMatch.repository).concat(revision ? '@' + revision : '', "/-/blob/").concat(fileMatch.path);
}
exports.getFileMatchUrl = getFileMatchUrl;
function getRepoMatchLabel(repoMatch) {
    var _a;
    var branch = (_a = repoMatch === null || repoMatch === void 0 ? void 0 : repoMatch.branches) === null || _a === void 0 ? void 0 : _a[0];
    var revision = branch ? "@".concat(branch) : '';
    return repoMatch.repository + revision;
}
exports.getRepoMatchLabel = getRepoMatchLabel;
function getRepoMatchUrl(repoMatch) {
    var label = getRepoMatchLabel(repoMatch);
    return '/' + encodeURI(label);
}
exports.getRepoMatchUrl = getRepoMatchUrl;
function getCommitMatchUrl(commitMatch) {
    return '/' + encodeURI(commitMatch.repository) + '/-/commit/' + commitMatch.oid;
}
exports.getCommitMatchUrl = getCommitMatchUrl;
function getMatchUrl(match) {
    switch (match.type) {
        case 'path':
        case 'content':
        case 'symbol':
            return getFileMatchUrl(match);
        case 'commit':
            return getCommitMatchUrl(match);
        case 'repo':
            return getRepoMatchUrl(match);
    }
}
exports.getMatchUrl = getMatchUrl;
function isSearchMatchOfType(type) {
    return function (match) { return match.type === type; };
}
exports.isSearchMatchOfType = isSearchMatchOfType;
// Write a TypeScript function to call the compute endpoint using fetchEventSource. The function should take the compute query and return any results into the console.
// Write a function to call the compute endpoint using fetchEventSource
var URL = './api/compute/stream';
function fetch(URL) {
    return (0, fetch_event_source_1.fetchEventSource)(URL, {
        onmessage: function (event) {
            console.log(event.data);
        }
    });
}
exports.fetch = fetch;
// call the fetch function
console.log("function output: " + fetch(URL));
