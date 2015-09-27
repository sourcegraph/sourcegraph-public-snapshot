/*
  Here, we deal with tokens, which are encoded as a record that
  represents the following tagged union:

  data Token = Term {String:String}
             | Repo {URI:String, Repo:{URI:String,Language:String}}
             | Rev {Rev:String,
                    Commit:Maybe {Author:{Email:String,Date:Date},
                                  Message:String}}
             | Unit {UnitType:String,
                     Name:String,
                     Unit:Maybe{Data:Maybe{Data:Maybe{Data:Maybe{Name:String}},
                                           Repo:Maybe:String}}}
             | File {Path:String, Entry:{ModTime:Date}}
             | User {Login:String,
                     User:{Login:String, Name:String, Company:String,
                           Location:string}}

  Tokens are represented as a record with the following slots:

    Type:      String
    more:      (Token -> [String])
    title:     (Token -> String)
    _label:    Maybe String
    cursorVal: Maybe String
    hints:     Maybe [String]
    icon:      Maybe Icon
    label:     Maybe HtmlString
    labelText: Maybe (Token -> String)
    val:       Maybe (Token -> String)

  the nullable slots val, cursorVal, hints, and label can be explicitly filled
  by `setValAndLabel` based on other data in the Token.
*/

var Repo = require("../Repo");
var escapeHTML = require("../escapeHTML");
var highlightWordRemainingChars = require("./util").highlightWordRemainingChars;
var moment = require("moment");

var suggestionTemplates = {
	repo: {
		val(d) { return d.URI; },
		labelText(d) { return Repo.label(d.URI); },
		hints(d) { return [d.URI, Repo.label(d.URI), Repo.base(d.URI)]; },
		icon: "repo",
		title(d) { return Repo.label(d.Repo.URI); },
		more(d) { return [d.Repo.Language, d.Repo.Description]; },
	},
	rev: {
		val(d) { return `:${d.Rev}`; },
		labelText(d) { return d.Rev.length === 40 ? d.Rev.slice(0, 6) : d.Rev; },
		cursorVal(d) { return `:${d.Rev}`; },
		hints(d) { return [d.Rev]; },
		icon: "git-commit",
		title(d) { return d.Rev; },
		more(d) { return d.Commit ? [d.Commit.Author.Email, moment(d.Commit.Author.Date).fromNow(), d.Commit.Message] : []; },
	},
	unit: {
		val(d) { return `~${d.Name}@${d.UnitType || ""}`; },
		labelText(d) {
			// Strip off repo prefix, if present.
			if (d.Unit) {
				// GoPackage
				if (d.Unit.Data && d.Unit.Data && d.Unit.Data.Data && d.Unit.Data.Data.Name) return d.Unit.Data.Data.Name;
			}
			// Shorten (possibly losing uniqueness).
			return d.Name.split("/").slice(-1);
		},
		cursorVal(d) { return `~${d.Name}@${d.UnitType || ""}`; },
		hints(d) {
			var basename = d.Name.split("/").slice(-1).join("");
			return [d.Name, basename, `~${basename}`];
		},
		icon: "primitive-dot",
		title(d) {
			if (d.Unit && d.Unit.Data && d.Unit.Data.Repo) {
				if (d.Unit.Data.Repo === d.Name) return Repo.base(d.Unit.Data.Repo);
				return d.Name.replace(`${d.Unit.Data.Repo}/`, "");
			}
			return d.Name;
		},
		more(d) { return [d.UnitType]; },
	},
	file: {
		val(d) { return `/${d.Path}`; },
		labelText(d) { return d.Path; },
		cursorVal(d) { return `/${d.Path}`; },
		hints(d) { return [d.Path]; },
		icon: "file-text",
		title(d) { return d.Path; },
		more(d) { return []; },
	},
	dir: {
		val(d) { return `/${d.Path}`; },
		labelText(d) { return d.Path || "/"; },
		cursorVal(d) { return `/${d.Path}`; },
		hints(d) { return [d.Path]; },
		icon: "file-directory",
		title(d) { return d.Path || "/"; },
		more(d) { return []; },
	},
	user: {
		val(d) { return `@${d.Login}`; },
		labelText(d) { return `@${d.Login}`; },
		hints(d) { return [d.Login, `@${d.Login}`]; },
		icon: "person",
		title(d) { return d.User.Login; },
		more(d) { return [d.User.Name, d.User.Company, d.User.Location]; },
	},
	org: {
		val(d) { return `@${d.Login}`; },
		labelText(d) { return `@${d.Login}`; },
		hints(d) { return [d.Login, `@${d.Login}`]; },
		icon: "organization",
		title(d) { return d.User.Login; },
		more(d) { return [d.User.Name, d.User.Location]; },
	},
	term: {
		val(d) { return d.String; },
		labelText(d) { return d.String; },
		hints(d) { return [d.String]; },
		icon: "search",
		title(d) { return d.String; },
		more(d) { return []; },
	},
};

exports.setValAndLabel = function(tok) {
	var typ = suggestionTypeFromToken(tok);
	if (!tok.val) {
		tok.val = suggestionTemplates[typ].val(tok);
	}
	if (!tok.label) {
		tok.label = renderTokenLabel(tok);
	}
	if (!tok.hints) {
		tok.hints = suggestionTemplates[typ].hints(tok);
	}
	if (!tok.cursorVal) {
		tok.cursorVal = (suggestionTemplates[typ].cursorVal || suggestionTemplates[typ].labelText)(tok);
	}
};

exports.renderTokenSuggestion = function(tok) {
	return renderToken(tok, tok._query);
};

function suggestionTypeFromToken(tok) {
	if (tok.Type === "Term" || tok.Type === "AnyToken") return "term";
	if (tok.Type === "RepoToken") return "repo";
	if (tok.Type === "RevToken") return "rev";
	if (tok.Type === "UnitToken") return "unit";
	if (tok.Type === "FileToken") return (tok.Entry && tok.Entry.Type === "file") ? "file" : "dir";
	if (tok.Type === "UserToken") return (tok.User && tok.User.Type === "Organization") ? "org" : "user";
	throw new Error(`unexpected token type ${tok.Type}`);
}

function renderToken(tok, query) {
	var typ = suggestionTypeFromToken(tok);
	var data = suggestionTemplates[typ];

	if (!data) throw new Error(`no suggestion templates for type ${typ} (token type ${tok.Type})`);

	var terms = query ? [query] : null;

	var title = data.title(tok);
	var more = data.more(tok).filter(function(d) { return Boolean(d); }).join(" — ");
	var html = `<p class="${tok._label} ${typ}">
	<span class="main-icon octicon octicon-${data.icon}"></span>
	<span class="title">${escapeHTML(title)}</span>
	<span class="more">${escapeHTML(more)}</span>
</p>`;
	return highlightWordRemainingChars(html, terms, "title");
}

function renderTokenLabel(tok) {
	var typ = suggestionTypeFromToken(tok);
	return `<span class="octicon octicon-${suggestionTemplates[typ].icon}"></span> ${escapeHTML(suggestionTemplates[typ].labelText(tok))}`;
}
