var sandbox = require("../testSandbox");
var expect = require("expect.js");

var $ = require("jquery");
global.$ = global.jQuery = global.window.$ = $;
var SearchBar = require("./SearchBar");
var Tokens = require("./Tokens");
var Engine = require("./Engine");

describe("SearchBar", function() {
	beforeEach(function() {
		createSearchForm();
	});

	function createSearchForm() {
		document.body.innerHTML =
			"<form id='search-form' method='get' data-pjax='#results' action='/search'>" +
			"  <input name=q />" +
			"</form>" +
			"<div id='results'></div>";
	}
	function newSearchBar(o) {
		o = o || {};
		var defaults = {
			form: "#search-form",
			input: "input[name=q]",
		};
		return new SearchBar($.extend(defaults, o));
	}
	function newSearchBarWithResolvedInitialTokens(o) {
		o = o || {};
		var defaults = {
			initialResolvedTokens: [
				{Type: "x"},
				{Type: "Term", String: "t0"},
				{Type: "Term", String: "t1"},
			],
		};
		$("input[name=q]").val("myval t0 t1");
		return newSearchBar($.extend(defaults, o));
	}

	var mockEngineImpl = function() { return function() {}; };
	var mockSetValAndLabelImpl = function(tok) {
		if (tok.Type === "Term") {
			tok.label = tok.val = tok.String;
		} else {
			tok.val = "myval";
			tok.label = "mylabel";
		}
	};

	describe("initialization", function() {
		beforeEach(function() {
			sandbox.stub(Engine, "create", mockEngineImpl);
		});

		it("focuses the typeahead input if the original input was autofocused", function() {
			document.querySelector("input").autofocus = true;
			expect(newSearchBar().$input.is(":focus")).to.be(true);
		});

		describe("with empty initial value", function() {
			it("treats empty initial value as no scope tokens and no input value", function() {
				var sb = newSearchBar();
				expect(sb.$input.val()).to.eql("");
				expect(sb.getQueryScopePrefix()).to.eql("");
				expect(sb.getQuery()).to.eql({Text: "", InsertionPoint: 0});
			});
		});

		describe("with initial value", function() {
			it("strips non-term tokens from the input field", function() {
				sandbox.stub(Tokens, "setValAndLabel", mockSetValAndLabelImpl);
				var sb = newSearchBarWithResolvedInitialTokens();
				expect(sb.$input.val()).to.eql("t0 t1");
				expect(sb.getQuery()).to.eql({Text: "myval t0 t1", InsertionPoint: 11});
				expect(Tokens.setValAndLabel.called).to.be(true);
			});

			it("adds non-term tokens to the token field", function() {
				sandbox.stub(Tokens, "setValAndLabel", mockSetValAndLabelImpl);
				var sb = newSearchBarWithResolvedInitialTokens();
				expect(sb.getQueryScopePrefix()).to.eql("myval ");
				expect(Tokens.setValAndLabel.called).to.be(true);
			});
		});
	});

	describe("UI interaction", function() {
		beforeEach(function() {
			sandbox.stub(Engine, "create", mockEngineImpl);
		});

		it("toggles .focused on its parent form when it is focused", function() {
			var sb = newSearchBar();
			expect($("form").hasClass("focused")).to.be(false);
			sb.focus();
			expect($("form").hasClass("focused")).to.be(true);
			sb.blur();
			expect($("form").hasClass("focused")).to.be(false);
		});
	});

	describe("deleting tokens", function() {
		beforeEach(function() {
			sandbox.stub(Engine, "create", mockEngineImpl);
			$.pjax = sandbox.stub();
		});

		beforeEach(function() {
			sandbox.stub(Tokens, "setValAndLabel", mockSetValAndLabelImpl);
		});

		it("is prevented if the token has _preventDelete true", function() {
			var sb = newSearchBarWithResolvedInitialTokens({
				initialResolvedTokens: [{_preventDelete: true, Type: "Foo", String: "x"}],
			});
			expect(sb.getQuery().Text).to.be("myval ");
			expect(sb.$form.find(".token .close").length).to.be(0);
			sb.$origInput.tokenfield("remove", $.Event("click", {target: sb.$form.find(".token")}));
			expect(sb.getQuery().Text).to.be("myval ");
		});

		it("removes the token if no _preventDelete", function() {
			var sb = newSearchBarWithResolvedInitialTokens({
				initialResolvedTokens: [{Type: "Foo", String: "x"}],
			});
			expect(sb.getQuery().Text).to.be("myval ");
			sb.$form.find(".token .close").click();
			expect(sb.getQuery().Text).to.be("");
		});
	});

	describe("searching", function() {
		beforeEach(function() {
			sandbox.stub(Tokens, "setValAndLabel", mockSetValAndLabelImpl);
		});

		beforeEach(function() {
			sandbox.stub(Engine, "create", mockEngineImpl);
		});

		it("should retain tokens and input value", function() {
			// Typeahead.js clears the input value when you hit
			// <ENTER> unless we set the value with
			// $myElem.typeahead("val", "xyz...").
			$.pjax = sandbox.stub();
			var sb = newSearchBarWithResolvedInitialTokens();
			sb.$input.focus();
			sb.$input.trigger("focus");
			sb.$input.trigger($.Event("keyup", {keyCode: 13})); // <ENTER> key
			expect($.pjax.callCount).to.be(1);
			expect($.pjax.firstCall.args[0].data).to.eql({q: "myval t0 t1"});
			expect(sb.$input.val()).to.be("t0 t1");
		});
	});

	describe("suggestions", function() {
		beforeEach(function() {
			sandbox.stub(Tokens, "setValAndLabel", mockSetValAndLabelImpl);
			sandbox.stub(Tokens, "renderTokenSuggestion", function(tok) {
				return "<p>" + tok.label + "</p>";
			});
		});

		var suggestions;
		var mockDataFn;
		beforeEach(function() {
			suggestions = [];
			mockDataFn = sandbox.spy(function(q, sync, async) { sync(suggestions); });
			sandbox.stub(Engine, "create", function(sb) {
				return mockDataFn;
			});
		});

		describe("with empty initial value", function() {
			it("gets suggestions on focus", function() {
				suggestions = [{val: "x", label: "x"}, {val: "y", label: "y"}];
				var sb = newSearchBar();
				sb.focus();
				expect(Engine.create.callCount).to.be(1);
				expect(mockDataFn.callCount).to.be(1);
			});

			it("after selecting, it adds the token and searches", function() {
				suggestions = [{val: "xx", label: "xx", foo: "bar"}];
				var sb = newSearchBar();
				sb.focus();
				sb._$typeaheadInput.typeahead("val", "x");

				sb.search = sandbox.stub();

				var $selectable = $(".tt-selectable").first();

				// All of these should work:
				sb._$typeaheadInput.typeahead("autocomplete", $selectable);
				// sb._$typeaheadInput.typeahead("select", $selectable);
				// sb.$input.trigger($.Event("keydown", {keyCode: 9 /* <TAB> */}));

				expect(sb.getQueryScopePrefix()).to.be("xx ");
				expect(sb.search.callCount).to.be(1);
			});

			it("when scope token has been added, suggestions are shown", function() {
				suggestions = [{val: "x", label: "x"}];
				var sb = newSearchBar();
				sb.focus();
				sb.search = sandbox.stub();

				sb._$typeaheadInput.typeahead("autocomplete", $(".tt-selectable").first());
				expect(sb._$typeaheadInput.typeahead("isOpen")).to.be(true);
			});
		});

		describe("when completing", function() {
			var keys = {
				"<TAB>": 9,
				"<SPACE>": 32,
			};
			Object.keys(keys).forEach(function(keyName) {
				var keyCode = keys[keyName];

				it("if a prefix of the val is typed and " + keyName + " is entered, a rich token is created", function() {
					suggestions = [{val: "xx", label: "xx", foo: "bar"}];
					var sb = newSearchBar();
					sb.focus();
					sb._$typeaheadInput.typeahead("val", "x");

					sb.search = sandbox.stub();

					sb.$input.trigger($.Event("keydown", {keyCode: keyCode}));

					expect(sb.getQueryScopePrefix()).to.be("xx ");
					expect(sb._$typeaheadInput.typeahead("val")).to.be("");
					expect(sb.$origInput.tokenfield("getTokens")).to.eql([{val: "xx", label: "xx", foo: "bar"}]);
				});

				it("if the exact val and " + keyName + " is entered, a rich token is created", function() {
					suggestions = [{val: "xx", label: "xx", foo: "bar"}];
					var sb = newSearchBar();
					sb.focus();
					sb._$typeaheadInput.typeahead("val", "xx");

					sb.search = sandbox.stub();

					sb.$input.trigger($.Event("keydown", {keyCode: keyCode}));

					expect(sb.getQueryScopePrefix()).to.be("xx ");
					expect(sb._$typeaheadInput.typeahead("val")).to.be("");
					expect(sb.$origInput.tokenfield("getTokens")).to.eql([{val: "xx", label: "xx", foo: "bar"}]);
				});
			});
		});

		describe("the autocomplete text varies depending on what the user has typed", function() {
			function typeaheadHint() { return $(".tt-hint").val(); }

			var sb;
			beforeEach(function() {
				suggestions = [{val: "aaaa/bbbb/cccc", hints: ["cccc", "bbbb/cccc", "aaaa/bbbb/cccc"]}];
				sb = newSearchBar();
				sb.focus();
				sb.search = sandbox.stub();
			});

			it("displays the rest of the token", function() {
				// This is normal typeahead completion display behavior.
				sb._$typeaheadInput.typeahead("val", "aa");

				expect(typeaheadHint()).to.be("aaaa/bbbb/cccc");
			});

			it("displays the rest of the remaining nonword tokens", function() {
				// This is *special* typeahead completion display
				// behavior so that you can type, e.g., a repo
				// basename and still see what hitting <TAB> will
				// complete.
				sb._$typeaheadInput.typeahead("val", "bb");

				expect(typeaheadHint()).to.be("bbbb/cccc");
			});

			it("displays the rest of the last nonword token when the user has just typed that last portion", function() {
				// This is *special* typeahead completion display
				// behavior so that you can type, e.g., a repo
				// basename and still see what hitting <TAB> will
				// complete.
				sb._$typeaheadInput.typeahead("val", "cc");

				expect(typeaheadHint()).to.be("cccc");
			});
		});

		describe("the cursor-selection displays a shortened val while selected", function() {
			it("displays a shortened val", function() {
				suggestions = [{val: "aaaa/bbbb/cccc", cursorVal: "foo"}];
				var sb = newSearchBar();
				sb.focus();
				sb.search = sandbox.stub();

				// This is normal typeahead completion display behavior.
				sb._$typeaheadInput.typeahead("moveCursor", 1);

				expect(sb.$input.val()).to.be("foo");
			});
		});
	});
});
