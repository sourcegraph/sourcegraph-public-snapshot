var
	XML_STYLE = /<!--[\s\S]*?-->/g,
	C_STYLE = /\/\*\*?[\s\S\@\{\}]*?\*\//g,
	PYTHON = /'''[\s\S]*?'''/g,
	RUBY = /=begin[\s\S]*?=end/g,
	APPLESCRIPT = /\(\*[\s\S]*?\*\)/g
;

module.exports = {
	'xml-comments': XML_STYLE,
	'html-comments': XML_STYLE,
	'c-comments': C_STYLE,
	'java-comments': C_STYLE,
	'js-comments': C_STYLE,
	'as-comments': C_STYLE,
	'python-comments': PYTHON,
	'ruby-comments': RUBY,
	'applescript-comments': APPLESCRIPT
};
