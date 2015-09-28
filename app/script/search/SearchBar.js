/*
Here we create and manipulate two global objects.

    1) window.SB which contains some state and methods for
       manipulating the search bar's DOM.

    2) The DOM element with id=search-form.
        This is expected to have a data-set that's pre-populated with
            {initialResolvedTokens:         [Token],
             initialResolveErrors:          [ResolutionError],
             nonDeletableLeadingTokenCount: Int,
             actionImplicitQueryString:     String
            }

        Once we extract the information from the dataset, we

            - Clean up the data from the dataset.
            - Resize the search bar.
            - Give control of the search bar's <input> field to Bootstrap
              Tokenfield.
            - Add a bunch of event handlers to handle annoying edge-cases with
              typeahead and tokenfield.
            - Handle a TON of other edge-cases related to our approach
              a little bit than what these libraries expect.
            - Handle edge-cases resulting from the interaction of global,
              mutable state with pushstate/popstate events.

*/

var $ = require("jquery");

require("typeahead.js/dist/typeahead.bundle.js");
require("bootstrap-tokenfield/js/bootstrap-tokenfield");

require("../pjax");
var debounce = require("../debounce");

var Dataset = require("./Dataset");
var Engine = require("./Engine");
var Tokens = require("./Tokens");
var Tokenize = require("./Tokenize");

document.addEventListener("DOMContentLoaded", function() {
	var $form = document.getElementById("search-form");
	var $input = document.querySelector("#search-form input[name=q]");
	if (!$input) {
		return;
	}

	var tokens = JSON.parse($input.dataset.initialResolvedTokens || "[]");

	var resolveErrors = JSON.parse($input.dataset.initialResolveErrors || "[]");

	var nonDeletableLeadingTokenCount = parseInt($input.dataset.nonDeletableLeadingTokenCount, 10) || 0;
	for (var i = 0; i < nonDeletableLeadingTokenCount; i++) {
		if (!tokens[i]) throw new Error("nonDeletableLeadingTokenCount (" + nonDeletableLeadingTokenCount + ") exceeds tokens.length (" + tokens.length + ")");
		tokens[i]._preventDelete = true;
	}

	window.SB = new SearchBar({
		form: $form,
		input: $input,
		navContainer: "#nav > div",
		initialResolvedTokens: tokens,
		initialResolveErrors: resolveErrors,
		implicitQueryString: $form.dataset.actionImplicitQueryString || "",
	});
});

function SearchBar(o) {
	o = o || {};

	this.$form = $(o.form);
	if (!this.$form.length) throw new Error("no form: " + o.form);
	this.$origInput = $(o.input);
	if (!this.$origInput.length) throw new Error("no input: " + o.input);

	this.$navContainer = o.navContainer ? $(o.navContainer) : null;

	this.implicitQueryString = (o.implicitQueryString ? o.implicitQueryString + " " : "");

	this.engine = Engine.create(this);
	if (this.engine.initialize) {
		this.engine.initialize()
		.fail(function(err) { console.error("Error preloading search completions:", err); });
	}

	this.initialize(o.initialResolvedTokens || [], o.initialResolveErrors || []);
}

SearchBar.prototype.initialize = function(initialResolvedTokens, initialResolveErrors) {
	$(".search-form").show();
	// Need to expand search bar before calling $.tokenfield()
	// because that stores the original width of $origInput.
	if (this.$navContainer && this.$navContainer.length) {
		this._fillNavWidth(this.$origInput);
		window.addEventListener("resize", function() {
			this._fillNavWidth(this.$form.find(".tokenfield"));
		}.bind(this));
	}

	// Set label and val on the initial tokens.
	initialResolvedTokens.forEach(Tokens.setValAndLabel);

	// Remove scope tokens (i.e., query tokens that are not just
	// terms) from $origInput so we can make them *real*
	// tokenfield tokens.
	var scopeTokens = [];
	var termTokens = [];
	initialResolvedTokens.forEach(function(tok) {
		if (tok.Type === "Term") termTokens.push(tok.String);
		else scopeTokens.push(tok);
	});
	this.$origInput.val(termTokens.join(" "));

	// Set up tokenfield and typeahead.
	this.$origInput.tokenfield({
		html: true,
		delimiter: [],
		tokens: scopeTokens,
		allowEditing: false,
		typeahead: [
			{hint: true, highlight: false, minLength: 0},
			Dataset.newTokenCompletions(this, this.engine),
		],
	});

	this.$input = this.$form.find("input.tt-input");
	if (!this.$input.length) throw new Error("Error loading tokenfield and typeahead");

	// this._$typeaheadInput is the element that has the typeahead
	// func and that you can register "typeahead:*" event
	// listeners on. TODO(sqs): if this is always the same as
	// this.$input, no need for a separate variable for it.
	this._$typeaheadInput = $(this.$input);

	// TODO What does this do?
	this._$typeaheadInput.typeahead("val", termTokens.join(" "));

	this.$origInput.on("tokenfield:createtoken", function(ev) {
		if (this.getQueryScopePrefix() && !Tokenize.lastTermMightBecomeToken(this.getQuery().Text)) {
			Reflect.deleteProperty(ev, "attrs"); // stop event propagation
		}
	}.bind(this));
	this.$origInput.on("tokenfield:removedtoken", function(ev) {
		this.search({replace: true});
		this.focus();
	}.bind(this));

	// If the user has typed the top suggestion's val, then <TAB>
	// does not autocomplete in typeahead.js, so
	// bootstrap-tokenfield creates a Term token (not a rich
	// token, like the RepoToken if the user had typed a repo
	// suggestion's val). This fixes that behavior so that <TAB>
	// always creates a rich token.
	this.$origInput.on("tokenfield:createtoken", function(ev) {
		// Prevent non-rich tokens from being created.
		var isRichToken = ev.attrs && Object.keys(ev.attrs).length > ["val", "label"].length;
		if (!isRichToken) Reflect.deleteProperty(ev, "attrs");
	});
	this.$input.on("keydown", function(ev) {
		// Ensure that <TAB> always triggers
		// typeahead:autocomplete (which is listened to by
		// bootstrap-tokenfield and triggers token creation), even
		// if the input val is already a full match.
		var keyCode = ev.which || ev.keyCode;
		if (keyCode !== 9 /* <TAB> */) return;
		this._autocompleteTopSelectable();
	}.bind(this));

	// Autocomplete the token when <SPACE> is pressed.
	this.$input.on("keydown", function(ev) {
		var keyCode = ev.which || ev.keyCode;
		if (keyCode !== 32 /* <SPACE> */) return;
		if (!this.getQueryScopePrefix() || Tokenize.lastTermMightBecomeToken(this.getQuery().Text)) {
			this._autocompleteTopSelectable();
		}
	}.bind(this));

	// Update search results after autocompletion.
	this._$typeaheadInput.on("typeahead:autocomplete typeahead:select", function(ev, sugg) {
		if (!sugg || !sugg[0]) return;
		this.search({replace: true});
	}.bind(this));

	// HACK: Force an update of the query used to fetch remote
	// suggestions when a suggestion is selected with the
	// mouse. When that occurs, the typeahead input query doesn't
	// change, so the dataset is not updated. This breaks because
	// in the engine we prepend the scope prefix to the query, but
	// typeahead doesn't know about that.
	this._$typeaheadInput.on("typeahead:select", function(ev, sugg) {
		this._$typeaheadInput.data().ttTypeahead.menu.datasets.forEach(function(dataset) {
			dataset.update("");
		});
	}.bind(this));

	// Search when ENTER is pressed. TODO(sqs): try keydown
	this.$input.on("keyup", function(ev) {
		var keyCode = ev.which || ev.keyCode;
		if (keyCode === 13 /* <ENTER> */) {
			this.search({push: true});
			this._$typeaheadInput.typeahead("close");
		}
	}.bind(this));

	// Search when a character is added to or removed from the
	// input field.
	this.$input.on("input", function(ev) {
		this.search({replace: true, debounce: true});
	}.bind(this));

	// Prevent the typeahead menu from opening when we are
	// suggesting def name completions, unless there are source
	// unit completions.
	this._$typeaheadInput.on("typeahead:open", setTypeaheadMenuVisibility.bind(this));
	this.$input.on("keyup", setTypeaheadMenuVisibility.bind(this));
	this._$typeaheadInput.on("typeahead:asyncreceive", setTypeaheadMenuVisibility.bind(this));

	function setTypeaheadMenuVisibility(ev) {
		// HACK: determine if there are source unit completions,
		// in which case we *do* want to show the menu.
		var $selectables = this._$typeaheadInput.data().ttTypeahead.menu.datasets[0].$el.children(".tt-selectable");
		var hasNonTermSuggestions = $selectables.filter(function() {
			return $(this).data("ttSelectableObject").Type !== "Term";
		}).length > 0;

		if (this.getQueryScopePrefix() && !hasNonTermSuggestions && !Tokenize.lastTermMightBecomeToken(this.getQuery().Text)) {
			this.$form.find(".tt-menu").addClass("hide-tt-menu");
		} else {
			this.$form.find(".tt-menu").removeClass("hide-tt-menu");
		}
	}

	// Prevent activation (focus) and deletion of tokens with
	// _preventDelete: true.
	var _$origInput = this.$origInput;
	this.$form.find(".token").each(function() {
		var $tok = $(this);
		var data = _$origInput.tokenfield("getTokenData", $tok);
		if (data._preventDelete) {
			applyTokenPreventDelete($tok);
		}
	});
	this.$origInput.on("tokenfield:createdtoken", function(ev) {
		if (ev.attrs._preventDelete) applyTokenPreventDelete($(ev.relatedTarget));
	});
	function applyTokenPreventDelete($tok) {
		$tok.find(".close").remove();
		$tok.addClass("prevent-delete");
	}
	this.$origInput.on("tokenfield:activatetoken", function(ev) {
		var data = this.$origInput.tokenfield("getTokenData", ev.relatedTarget);
		if (data && data._preventDelete) {
			ev.preventDefault();
			this._$typeaheadInput.typeahead("activate");
		}
	}.bind(this));
	this.$form.on("click", ".token.prevent-delete", function(ev) {
		this.focus();
	}.bind(this));
	this.$origInput.on("tokenfield:removetoken", function(ev) {
		if (ev.attrs && ev.attrs._preventDelete) ev.preventDefault();
	});

	// Set token resolution state initially.
	this.updateTokenResolutionErrors(initialResolveErrors);

	// Add .focused class to the form when the input is focused, so we can style
	// the input overlays and hide other nav links.
	this.$input.on("focus", function(ev) {
		this.$form.addClass("focused");
	}.bind(this));
	this.$input.on("blur", function() {
		this.$form.removeClass("focused");
	}.bind(this));

	// Respect autofocus on original input.
	if (this.$origInput.prop("autofocus")) {
		this.$input.attr("autofocus", true);
		this.$input.focus(); // autofocus search input
	}

	// Clear the search bar on popstate back to a non-search page
	// (otherwise, if you initiate search on a non-search page and
	// then hit back to return to that previous non-search page,
	// the query remains and the input obscures some of the nav
	// links).
	//
	// KNOWN ISSUE: If you popstate back to a non-search page then
	// hit the forward button to return to the search page, the
	// results are displayed but the search field is empty (no
	// query, no tokens). It should contain the query
	// corresponding to the results. This would require tracking a
	// lot more state.
	$(window).on("popstate", function(ev) {
		var toSearchPage = ev && ev.state && /\/\.?search/.test(ev.state.url);
		if (!toSearchPage) {
			this._$typeaheadInput.typeahead("val", "");
			this._$typeaheadInput.typeahead("deactivate");
			var keepTokens = this.$origInput.tokenfield("getTokens").filter(function(tok) {
				return tok._preventDelete;
			});
			this.$origInput.tokenfield("setTokens", keepTokens);
		}
	}.bind(this));
};

// updateTokenResolutionErrors takes a list of ResolveErrors from
// the completions API response and updates the token states
// accordingly.
SearchBar.prototype.updateTokenResolutionErrors = function(resolveErrors) {
	var $toks = this.$form.find(".token");
	$toks.removeClass("invalid");
	(resolveErrors || []).forEach(function(rserr) {
		if (rserr.Index) {
			var $tok = $toks.eq(rserr.Index - 1);
			$tok.addClass("invalid");
		}
	});
};

SearchBar.prototype._autocompleteTopSelectable = function() {
	var val = this._$typeaheadInput.typeahead("val");
	if (val) {
		var $topSelectable = this.$form.find(".tt-selectable").first();
		if ($topSelectable.length) {
			// HACK: Clear typeahead input's query (private
			// var) so that autocomplete is triggered (since
			// the suggestion will no longer be an exact match
			// for the just-cleared query) and autocomplete to
			// the original top suggestion.
			this._$typeaheadInput.data().ttTypeahead.input.query = "";
			this._$typeaheadInput.typeahead("autocomplete", $topSelectable);
		}
	}
};

// _fillNavWidth expands $e to take up all available
// horizontal space in the navbar.
SearchBar.prototype._fillNavWidth = function($e) {
	if (!$e.length) return;
	if (!this.$navContainer || !this.$navContainer.length) return;
	var availableWidth = this.$navContainer.get(0).clientWidth;
	// Subtract the logo and nav links.
	availableWidth -= this.$navContainer.find(".navbar-header .navbar-brand").get(0).offsetWidth + this.$navContainer.find("ul.navbar-right").get(0).offsetWidth + this.$navContainer.find(".navbar-toggle").get(0).offsetWidth;
	availableWidth -= 35;
	$e.css("width", availableWidth + "px");
};

// getQueryScopePrefix returns the query string prefix that the
// current scope tokens represent (without the implicit query
// prefix, if any).
SearchBar.prototype.getQueryScopePrefix = function() {
	var prefix = this.$origInput.tokenfield("getTokens").map(function(tok) { return tok.val; }).join(" ");
	return prefix ? prefix + " " : "";
};

// getQuery returns the current query (without the implicit query
// prefix (e.g., for the current repo), if any).
SearchBar.prototype.getQuery = function() {
	var prefix = this.getQueryScopePrefix();
	return {
		Text: prefix + this.$origInput.tokenfield("getInput"),
		InsertionPoint: prefix.length + (this.$input.get(0).selectionStart || this.$input.val().length),
	};
};

// _getExplicitQuery returns the current query with the implicit
// query prefix removed, if any).
SearchBar.prototype._getExplicitQuery = function() {
	var q = this.getQuery();
	q.Text = q.Text.substring(this.implicitQueryString.length);
	q.InsertionPoint -= this.implicitQueryString.length;
	return q;
};

SearchBar.prototype.search = function(opt) {
	if (!this._debouncedSearch) this._debouncedSearch = debounce(this.search.bind(this), 200);

	var defaults = {
		type: "GET",
		url: this.$form.attr("action"),
		data: {q: this._getExplicitQuery().Text},
		container: this.$form.data("pjax"),
		target: this.$form,
		noLoadingIndicator: false,
	};
	opt = $.extend(defaults, opt);
	if (opt.debounce) {
		opt.debounce = false;
		this._debouncedSearch(opt);
	} else {
		// If we're on another (non-search) page and start
		// searching, force a push (not replace) state so the back
		// button returns us to that non-search page.
		if (window.location.pathname !== opt.url) {
			opt.replace = false;
			opt.push = true;
		}

		$.pjax(opt);
	}
};
SearchBar.prototype._debouncedSearch = null;

SearchBar.prototype.focus = function() {
	this.$input.focus();
	this.$input.trigger("focus");
};
SearchBar.prototype.blur = function() {
	this.$input.blur();
	this.$input.trigger("blur");
};

module.exports = SearchBar;
