var globals = require("../globals");
var router = require("../routing/router");

var AppDispatcher = require("../dispatchers/AppDispatcher");
var CodeFileRouterActions = require("../actions/CodeFileRouterActions");

var firstFile = true;

// This is an initial version of a "hacked-together router. It's more or less a
// throw-away implementation with a lot of room for improvement.

function registerFile(data) {
	var file = data.EntrySpec;
	var fileUrl = router.fileURL(file.RepoRev.URI, file.RepoRev.Rev||file.RepoRev.CommitID, file.Path);
	var state = {};

	if (data.Definition) {
		var def = data.Definition;
		fileUrl = def.URL;
		state.def = def;
		state.byteRange = [def.ByteStartPosition, def.ByteEndPosition];
	} else {
		state = hashToObject();
	}

	state.file = file;

	history.replaceState(state, "", fileUrl + location.hash);
}

function registerToken(token) {
	var CodeFileStore = require("../stores/CodeFileStore");
	var file = CodeFileStore.get("file");
	var fileUrl = router.fileURL(file.RepoRev.URI, file.RepoRev.Rev||file.RepoRev.CommitID, file.Path);

	history.pushState({
		file: file,
		selected: token.tid,
	}, "", fileUrl+"#selected="+token.tid);
}

function unregisterTokens() {
	var CodeFileStore = require("../stores/CodeFileStore");
	var file = CodeFileStore.get("file");
	var fileUrl = router.fileURL(file.RepoRev.URI, file.RepoRev.Rev||file.RepoRev.CommitID, file.Path);

	history.pushState({file: file}, "", fileUrl);
}

function registerDefinition(def) {
	history.pushState({
		file: def.File,
		def: {URL: def.URL, File: def.File},
		byteRange: [def.ByteStartPosition, def.ByteEndPosition],
	}, "", def.URL);
}

function registerSnippet(snippet) {
	var file = snippet.file;
	var fileUrl = router.fileURL(file.RepoRev.URI, file.RepoRev.Rev||file.RepoRev.CommitID, file.Path);
	var hash = "#startline="+snippet.startLine+"&endline="+snippet.endLine;
	var state = {
		file: file,
		lineRange: [snippet.startLine, snippet.endLine],
	};

	if (snippet.defUrl) {
		hash += "&defUrl="+snippet.defUrl;
		state.defUrl = snippet.defUrl;
	}

	history.pushState(state, "", fileUrl + hash);
}

function registerLineSelection(lineArray) {
	var CodeFileStore = require("../stores/CodeFileStore");
	var file = CodeFileStore.get("file");
	var fileUrl = router.fileURL(file.RepoRev.URI, file.RepoRev.Rev||file.RepoRev.CommitID, file.Path);
	var start = lineArray[0].get("number");
	var end = lineArray[lineArray.length-1].get("number");
	var hash = "#startline="+start+"&endline="+end;
	var state = {
		file: file,
		lineRange: [start, end],
	};
	history.pushState(state, "", fileUrl + hash);
}

module.exports.matchInitialHashState = function() {
	if (!firstFile) return;
	firstFile = false;

	var obj = hashToObject();

	var CodeFileStore = require("../stores/CodeFileStore");

	var cm = CodeFileStore.get("codeModel");
	var def = obj.def || obj.defUrl;
	var storeDef = CodeFileStore.get("activeDef");

	if (typeof def === "string") {
		var token = cm.tokens.get(def).get("refs")[0];
		CodeFileRouterActions.selectToken(token);
	}

	if (storeDef) {
		CodeFileStore._showByteRange(storeDef.ByteStartPosition, storeDef.ByteEndPosition);
	}

	if (obj.selected) {
		var selectedToken = cm.tokens.byId(obj.selected);
		CodeFileStore._scrollIntoView(selectedToken);
		CodeFileRouterActions.selectToken(selectedToken);
	}

	if (obj.lineRange) {
		CodeFileStore._showLineRange(obj.lineRange[0], obj.lineRange[1]);
	}

	if (obj.byteRange) {
		CodeFileStore._showByteRange(obj.byteRange[0], obj.byteRange[1]);
	}
};

function hashToObject() {
	var obj = {}, ret = {};

	location.hash.slice(1).split("&").forEach(kv => {
		var p = kv.split("=");
		obj[p[0]] = p[1];
	});

	Object.keys(obj).forEach(k => {
		if (k === "startline") {
			ret.lineRange = [parseInt(obj[k], 10)];
		} else if (k === "endline") {
			if (!Array.isArray(ret.lineRange)) ret.lineRange = [];
			ret.lineRange[1] = parseInt(obj[k], 10);
		} else if (k === "startbyte") {
			ret.byteRange = [parseInt(obj[k], 10)];
		} else if (k === "endbyte") {
			if (!Array.isArray(ret.byteRange)) ret.byteRange = [];
			ret.byteRange[1] = parseInt(obj[k], 10);
		} else if (k === "def") {
			ret.defUrl = obj[k];
		} else {
			ret[k] = obj[k];
		}
	});

	return ret;
}

module.exports.start = function() {
	module.exports.dispatchToken = AppDispatcher.register(function(payload) {
		if (["ROUTER_ACTION", "DEPENDENT_ACTION"].indexOf(payload.source) > -1) return;
		if (payload.action.opts && payload.action.opts.silent) return;

		var action = payload.action,
			CodeFileStore = require("../stores/CodeFileStore");

		switch (action.type) {
		case globals.Actions.RECEIVED_FILE:
			if (firstFile) {
				registerFile(action.data);
			}
			break;

		case globals.Actions.TOKEN_SELECT:
			registerToken(action.token);
			break;

		case globals.Actions.TOKEN_CLEAR:
			unregisterTokens();
			break;

		case globals.Actions.SHOW_DEFINITION:
			registerDefinition(action.params);
			break;

		case globals.Actions.LINE_SELECT:
			AppDispatcher.waitFor([CodeFileStore.dispatchToken]);
			registerLineSelection(CodeFileStore.getHighlightedLines());
			break;

		case globals.Actions.SHOW_SNIPPET:
			registerSnippet(action.params);
			break;
		}
	});

	window.onpopstate = e => {
		if (!e.state) return;

		var CodeFileStore = require("../stores/CodeFileStore");

		function matchState() {
			var cm = CodeFileStore.get("codeModel");

			cm.clearHighlightedLines();
			cm.clearSelectedTokens();

			if (e.state.selected) {
				var selectedToken = cm.tokens.byId(e.state.selected);
				CodeFileStore._scrollIntoView(selectedToken);
				CodeFileRouterActions.selectToken(selectedToken);
			}

			if (e.state.def) {
				CodeFileRouterActions.navigateToDefinition(e.state.def);
			}

			if (e.state.defUrl) {
				var token = cm.tokens.get(e.state.defUrl).get("refs")[0];
				CodeFileRouterActions.selectToken(token);
			}

			if (e.state.byteRange) {
				CodeFileStore._showByteRange(e.state.byteRange[0], e.state.byteRange[1]);
			}

			if (e.state.lineRange) {
				CodeFileStore._showLineRange(e.state.lineRange[0], e.state.lineRange[1]);
			}
		}

		if (!CodeFileStore.isSameFile(e.state.file)) {
			var file = e.state.file;
			var fileUrl = router.fileURL(file.RepoRev.URI, file.RepoRev.Rev||file.RepoRev.CommitID, file.Path);
			CodeFileRouterActions.loadFile(fileUrl).then(matchState);
		} else {
			matchState();
		}
	};
};
