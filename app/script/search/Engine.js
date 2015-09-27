// Here, we create a Bloodhound Engine that pulls suggestions from
// the ‘/api/search/complete?Text=%s&InsertionPoint=%d’ endpoint.
//
// We also do some miscellaneous bookkeeping here: marking token
// dom notes with errors, and marking tokens with information about where they
// came from.

var Bloodhound = require("typeahead.js/dist/typeahead.bundle.js");
var Tokens = require("./Tokens");
var QueryURL = "/api/search/complete";

var identity = function(x) { return x; };

module.exports.create = function(searchBar) {
	// Sets the css class of the dom elements of all tokens in the
	// search bar to `valid` or `invalid` by matching tokens against
	// resolveErrors.
	var markResolveErrors = function(transformer) {
		transformer = transformer || identity;
		return function(info) {
			searchBar.updateTokenResolutionErrors(info.ResolveErrors);
			return transformer(info);
		};
	};

	var engine = new Bloodhound({
		prefetch: {
			// TODO What is the purpose of this request? Does it actually get
			// any meaningful results?
			url: `${QueryURL}?Text=&InsertionPoint=0`,

			// TODO I don"t see `transform` in the typeahead/bloodhound
			// documentation or source code. Is this a mistake?
			transform: addLabel("prefetch", tokenCompTransformer),
		},
		remote: {
			url: `${QueryURL}?Text=%s&InsertionPoint=%d`,
			replace(url, rawQuery) {
				var q = searchBar.getQuery();
				return url.replace("%s", q.Text).replace("%d", q.InsertionPoint);
			},

			// TODO I don"t see `transform` in the typeahead/bloodhound
			// documentation or source code. Is this a mistake?
			transform: addLabel("remote", markResolveErrors(tokenCompTransformer)),
			rateLimitWait: 100,
		},

		// TODO Why is this hard-coded as "val"? Is this broken?
		// TODO What does Bloodhound.tokenizers.obj.whitespace actually do?
		datumTokenizer: Bloodhound.tokenizers.obj.whitespace("val"),

		// TODO A query is {Text:String, InsertionPoint:int}, is this what
		// bloodhound.tokenizers.whitespace expects?
		queryTokenizer: Bloodhound.tokenizers.whitespace,

		// TODO I don"t see this in the typeahead/bloodhound documentation
		// or source code. Is this a mistake?
		identify(d) { return `${d.Type}:${d.val}`; },
	});

	// TODO What is `async`? It looks like it"s ignored. __ttAdapter()
	// expects to take only two arguments.
	return function(q, sync, async) {
		var dropRepoAndUserToks = function(tokens) {
			return tokens.filter(function(t) {
				return ["RepoToken", "UserToken"].indexOf(t.Type) === -1;
			});
		};

		// If anything has been sucessfully tokenized, then don"t tokenize any
		// further repos or users. This makes sense, because we only allow repos
		// and users as the first token in the token-string.
		if (searchBar.getQueryScopePrefix()) {
			var origSync = sync;
			var origAsync = async;
			sync = function(ts) { return origSync(dropRepoAndUserToks(ts)); };
			async = function(ts) { return origAsync(dropRepoAndUserToks(ts)); };
		}
		return engine.__ttAdapter()(q, sync, async);
	};
};

// addLabel wraps an existing engine result transformer and makes it
// also add a label that can be used for debugging purposes (e.g., to
// record which data source a suggestion came from).
//
// This mutates the object that `transformer` returns, adding or
// updating the _label slot to be the same as the passed label.
function addLabel(label, transformer) {
	transformer = transformer || identity;
	return function(info) {
		return transformer(info).map(function(d) {
			d._label = label;
			return d;
		});
	};
}

// This mutates info.TokenCompletions, filling out missing fields in each token.
function tokenCompTransformer(info) {
	return (info.TokenCompletions || []).map(function(tok) {
		Tokens.setValAndLabel(tok);
		return tok;
	});
}
