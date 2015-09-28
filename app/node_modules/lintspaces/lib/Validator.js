var
	fs = require('fs'),
	merge = require('merge'),
	editorconfig = require('editorconfig'),

	DEFAULTS = require('./constants/defaults'),
	MESSAGES = require('./constants/messages'),
	PATTERNS = require('./constants/ignorePatterns'),
	MAPPINGS = require('./constants/editorconfig-mappings'),

	ValidationError = require('./ValidationError'),

	// Regular Expressions:
	eol = '\\r?\\n',
	eolRegExp = new RegExp(eol),
	tabsRegExp = /^\t*(?!\s).*$/, // leading tabs without leading spaces
	tabsRegExpForBOM = /^\t*(?! |\t).*$/, // leading tabs without leading spaces (allows BOM)
	tabsLeadingRegExp = /^(\t*).*$/, // leading tabs
	spacesRegExp = /^ *(?!\s).*$/, // leading spaces without leading tabs
	spacesRegExpForBOM = /^ *(?!\t).*$/, // leading spaces without leading tabs (allows BOM)
	spacesLeadingRegExp = /^( *).*$/ // leading spaces
;

/**
 * @constructor
 * @param {Object} options
 */
function Validator(options) {
	this._options = merge({}, DEFAULTS, options || {});
	this._processedFiles = 0;
	this._invalid = {};
}

// Externalize constants:
Validator.DEFAULTS = DEFAULTS;
Validator.MESSAGES = MESSAGES;
Validator.PATTERNS = PATTERNS;

/**
 * Check if a file is valid based on validation settings
 * @param {String} path Path to file
 */
Validator.prototype.validate = function(path) {
	var self = this, stat;

	try {
		stat = fs.statSync(path);
	} catch(e) {
		this._fail(
			MESSAGES.PATH_INVALID.message
				.replace('{a}', path)
		);
	}

	if (stat.isFile()) {
		this._cleanUp();

		// Load file, settings & ignores:
		this._path = path;
		this._loadSettings();
		this._loadFile();
		this._loadIgnores();

		// Validate total file:
		this._validateNewlineMaximum();
		this._validateNewlinesEOF();

		// Validate single lines:
		this._lines.forEach(function(line, index) {
			self._validateIndentation(line, index);
			self._validateTrailingspaces(line, index);
		});

		// Validation is done:
		this._done();
	} else {
		this._fail(
			MESSAGES.PATH_ISNT_FILE.message
				.replace('{a}', path)
		);
	}
};

/**
 * Get count of processed files
 * @return {Number}
 */
Validator.prototype.getProcessedFiles = function() {
	return this._processedFiles;
};

/**
 * After validation
 * @private
 */
Validator.prototype._done = function() {
	this._processedFiles++;
	this._cleanUp();
};

/**
 * Reset references
 * @private
 */
Validator.prototype._cleanUp = function() {
	this._settings = null;
	this._data = undefined;
	this._lines = null;
	this._ignoredLines = null;
};

/**
 * Load file data
 * @private
 */
Validator.prototype._loadFile = function() {
	this._data = fs.readFileSync(this._path, this._settings.encoding);
	this._lines = this._data.split(eolRegExp);
};

/**
 * Load settings
 * @private
 */
Validator.prototype._loadSettings = function() {
	var config, key;

	// Initially the users options are the current settings:
	this._settings = merge({}, this._options);

	// Overwrite settings by .editorconfig file's settings:
	if (typeof this._settings.editorconfig === 'string') {
		var stat;

		try {
			stat = fs.statSync(this._settings.editorconfig);
		} catch(e) {
			this._fail(
				MESSAGES.EDITORCONFIG_NOTFOUND.message
					.replace('{a}', this._settings.editorconfig)
			);
		}

		if (stat.isFile()) {
			// Load config for current path
			config = editorconfig.parse(
				this._path, {
					config: this._settings.editorconfig
				}
			);

			if (typeof config === 'object') {
				// Merge editorconfig values into the correct settings names:
				for (key in config) {
					if (typeof MAPPINGS[key] === 'string') {
						switch (key) {
							case 'indent_style':
								// The 'indent_style' property value isn't
								// equal to the expected setting value:
								this._settings[MAPPINGS[key]] = config[key] + 's';
								break;
							default:
								this._settings[MAPPINGS[key]] = config[key];
								break;
						}
					}
				}
			}
		} else {
			this._fail(
				MESSAGES.PATH_ISNT_FILE.message
					.replace('{a}', this._settings.editorconfig)
			);
		}
	}
};

/**
 * Load ignores
 * @private
 */
Validator.prototype._loadIgnores = function() {
	var
		self = this,
		ignores = []
	;

	this._ignoredLines = {};

	// Load ignore patterns:
	if (Array.isArray(this._settings.ignores)) {
		this._settings.ignores.forEach(function(ignore) {
			if (typeof ignore === 'string' && typeof PATTERNS[ignore] === 'object') {
				ignores.push(PATTERNS[ignore]);
			} else if (typeof ignore === 'object' && typeof ignore.test === 'function') {
				ignores.push(ignore);
			}
		});
	}

	// When no patterns are defined, disable the following search for lines:
	if (ignores.length === 0) {
		ignores = false;
	}

	// Index lines which match patterns, when available:
	if (Array.isArray(ignores)) {

		// Loop all given regular expressions:
		ignores.forEach(function(expression) {

			var matches = self._data.match(expression) || [];

			matches.forEach(function(match) {

				// Only perform an action when match has more
				// than one line:
				if (eolRegExp.test(match)) {

					// Use fake replace cycle to find indices of all
					// lines to be ignored. Return unchanged match:
					self._data = self._data.replace(match, function(matched) {
						var
							index = 1,
							args,
							indexOfMatch,
							indexOfSecondLine,
							totalLines
						;

						// last argument is whole string, remove it:
						args = Array.prototype.slice.call(arguments);
						args.pop();

						// matched string start index:
						indexOfMatch = args.pop();

						// slice source data from beginning to matched
						// string start index to find index of second
						// line to be ignored:
						indexOfSecondLine = self._data.slice(0, indexOfMatch).split(eolRegExp).length;
						totalLines = matched.split(eolRegExp).length;

						//Count and store lines:
						while (index < totalLines) {
							self._ignoredLines[indexOfSecondLine + index - 1] = true;
							index++;
						}

						// Fillup result with linebreaks and overwrite
						// data string in case that the data string contains
						// the current 'matched' more than once:
						return Array(totalLines).join('\n');
					});

				}
			});
		});
	}
};

/**
 * Check the maximum of newlines
 * @private
 */
Validator.prototype._validateNewlineMaximum = function() {
	if (typeof this._settings.newlineMaximum === 'number') {
		if (this._settings.newlineMaximum > 0) {
			var
				self = this,

				// To grep all blocks at the begining of a file
				// which have at least 1 more new line than the defined
				// criteria, match "newlineMaximum + 1" (or more) instances of eol:
				newlinesAtBeginn = '^(?:' + eol + '){' + (this._settings.newlineMaximum + 1) + ',}',

				// Each block inside a file has an extra leading newline
				// from the previous line above. To grep all blocks
				// which have at least 1 more new line than the defined criteria,
				// match "newlineMaximum + 2" (or more) instances of eol:
				newlinesInFile = '(?:' + eol + '){' + (this._settings.newlineMaximum + 2) + ',}',

				// Define function which is used as fake replace cycle to
				// validate matches:
				validate = function(match, offset, original) {
					var
						substring = original.substr(0, offset),
						newlines = substring.match(new RegExp(eol, 'g')),
						amount = match.match(new RegExp(eol, 'g')).length,
						atLine = 0,
						message,
						data,
						line,
						payload
					;

					// When current match is not at the beginning of a file,
					// newlines is defined. In this case update variables:
					if (newlines) {
						atLine = newlines.length + 1;
						amount = amount - 1;
					}

					// Test if found lines are not in ignored lines:
					if (!self._ignoredLines[atLine + 1]) {

						// Build message and report:
						message = MESSAGES.NEWLINE_MAXIMUM.message
							.replace('{a}', amount)
							.replace('{b}', self._settings.newlineMaximum);

						data = {message: message};
						data = merge({}, MESSAGES.NEWLINE_MAXIMUM, data);
						line = atLine + 1;
						payload = {
							amount: amount,
							maximum: self._settings.newlineMaximum
						};

						self._report(data, line, payload);
					}

					return original;
				}
			;

			this._data.replace(new RegExp(newlinesAtBeginn, 'g'), validate);
			this._data.replace(new RegExp(newlinesInFile, 'g'), validate);
		} else {
			this._fail(
				MESSAGES.NEWLINE_MAXIMUM_INVALIDVALUE.message
					.replace('{a}', this._settings.newlineMaximum)
			);
		}
	}
};

/**
 * Check newlines of the end of file
 * @private
 */
Validator.prototype._validateNewlinesEOF = function() {
	if (this._settings.newline && this._lines.length > 1) {
		var
			index = this._lines.length - 1
		;


		// check last line:
		if (this._lines[index].length > 0) {
			this._report(MESSAGES.NEWLINE, index + 1);
		}

		// check line before last line:
		if (this._lines[index - 1].length === 0) {
			this._report(MESSAGES.NEWLINE_AMOUNT, index + 1);
		}
	}
};

/**
 * Check trailing spaces
 * @private
 */
Validator.prototype._validateTrailingspaces = function(line, index) {
	if (this._settings.trailingspaces && typeof line === 'string') {
		var matchSpaces = line.match(/\s*$/);

		// Is there a trailing whitespace?
		if (matchSpaces.length > 0 && matchSpaces[0].length > 0) {

			// Check the options if trainlingspaces should be ignored and the
			// current line is inside the ignored lines: stop reporting!
			//
			// NOTE: the '+1' at the _ignoredLines property is necessary because
			// on how the 'ignored lines' options work. The are shifted by +1.
			if (this._options.trailingspacesToIgnores && this._ignoredLines[index + 1]) {
				return;
			}

			// Check is empty lines should not be reported and
			// skipped when empty:
			if (this._options.trailingspacesSkipBlanks && line.trim() === '') {
				return;
			}

			this._report(MESSAGES.TRAILINGSPACES, index + 1);
		}
	}
};

/**
 * Check indentations
 * @private
 */
Validator.prototype._validateIndentation = function(line, index) {
	if (!this._ignoredLines[index] &&
		typeof this._settings.indentation === 'string' &&
		typeof line === 'string') {

		var
			spacesExpected,
			indent,
			message,
			data,
			payload
		;

		switch (this._settings.indentation) {
			case 'tabs':
				var tabsRegExpFinal = (this._settings.allowsBOM ? tabsRegExpForBOM : tabsRegExp);
				if (!tabsRegExpFinal.test(line)) {
					// indentation failed...
					return this._report(MESSAGES.INDENTATION_TABS, index + 1);
				}

				this._guessIndentation(line, index);
				break;

			case 'spaces':
				var spacesRegExpFinal = (this._settings.allowsBOM ? spacesRegExpForBOM : spacesRegExp);
				if (!spacesRegExpFinal.test(line)) {
					// Indentation failed...
					this._report(MESSAGES.INDENTATION_SPACES, index + 1);
				} else {
					// Indentation correct, is amount of spaces correct?
					if (typeof this._settings.spaces === 'number') {
						indent = line.match(spacesLeadingRegExp)[1].length;
						if (indent % this._settings.spaces !== 0) {
							// Indentation incorrect, create message and report:
							spacesExpected = Math.round(indent / this._settings.spaces) * this._settings.spaces;
							message = MESSAGES.INDENTATION_SPACES_AMOUNT.message
								.replace('{a}', spacesExpected)
								.replace('{b}', indent);

							data = {message: message};
							data = merge({}, MESSAGES.INDENTATION_SPACES_AMOUNT, data);
							payload = {
								expected: spacesExpected,
								indent: indent
							};

							this._report(data, index + 1, payload);
						}
					}
				}

				this._guessIndentation(line, index);
				break;
		}
	}
};

Validator.prototype._guessIndentation = function(line, index) {
	if (!this._ignoredLines[index] &&
		this._settings.indentationGuess &&
		this._settings.indentation) {

		var
			indentation = this._settings.indentation,
			indentationPrevious,
			regExp = indentation === 'tabs' ? tabsLeadingRegExp : spacesLeadingRegExp,
			match = line.match(regExp),
			matchPrevious = 0,
			diff,
			message,
			data
		;

		// Get amount of whitespaces at the beginnig of a line for the
		// current line and the previous line:
		match = match.length > 1 ? match[1].length : 0;
		if (index > 0) {
			matchPrevious = this._lines[index - 1].match(regExp);
			matchPrevious = matchPrevious.length > 1 ? matchPrevious[1].length : 0;
		}

		// Calculate the indentation for both lines:
		indentation = match;
		indentationPrevious = matchPrevious;
		if (this._settings.indentation === 'spaces') {
			indentation = match / this._settings.spaces;
			indentationPrevious = matchPrevious / this._settings.spaces;
		}

		// Check if indentations has no decimal values which would be an
		// invalid indentation:
		if (indentation % 1 !== 0 || indentationPrevious % 1 !== 0) {
			// Stop indentation guessing for this line...
			return;
		}

		// The indentation of the current line is correct when:
		// * the amount of indentations is equal to the previous or
		// * the amount of indentations is less than the previous line or
		// * the amount of indentations is one more than the previous line
		// * the amount of indentations is zero and the lines length is also
		//   zero which is an empty line without trailing whitespaces
		if (indentation - indentationPrevious <= 1) {
			// Valid, nothing to report...
			return;
		}

		// The line differs from the previous line:
		// Get the expected offset:
		diff = indentation - indentationPrevious;
		diff = diff / Math.abs(diff);

		// report:
		message = MESSAGES.INDENTATION_GUESS.message
			.replace('{a}', indentationPrevious + 1)
			.replace('{b}', indentation);


		data = {message: message};
		data = merge({}, MESSAGES.INDENTATION_GUESS, data);

		this._report(data, index + 1, {
			indentation: indentation,
			expected: indentationPrevious + 1
		});
	}
};

/**
 * Throw an exception
 * @param {String} message Exception message
 * @private
 */
Validator.prototype._fail = function(message) {
	throw new Error(message);
};

/**
 * Add an invalid line
 * @param {Object} data Data message and errocode
 * @param {Number} linenumber where error appeared
 * @param {String} payload Optional data for the validation error
 */
Validator.prototype._report = function(data, linenumber, payload) {
	var
		line,
		file,
		validationError
	;

	// Build dataset, aware to not overwrite the given data:
	data = merge({}, data);
	data.line = linenumber;

	// Lookup for current file:
	if (!this._invalid[this._path]) {
		this._invalid[this._path] = {};
	}
	file = this._invalid[this._path];

	// Lookup for given line:
	if (!file[linenumber]) {
		file[linenumber] = [];
	}
	line = file[linenumber];

	// Build error:
	validationError = new ValidationError(data, payload);

	// Store error:
	line.push(validationError);
};

/**
 * Get invalid lines by path
 * @param {String} path
 * @return {Object} each key in the returned object represents a line from the
 * file of the given path, each value an exeption message of the current line.
 */
Validator.prototype.getInvalidLines = function(path) {
	if (!this._invalid[path]) {
		return {};
	}

	return this._invalid[path];
};

/**
 * Get a all invalid lines and messages from processed files.
 * @return {Object} each key in the returned object represents a path of a
 * processed invalid file. Each value is an other object
 * containing the validation result.
 */
Validator.prototype.getInvalidFiles = function() {
	return this._invalid;
};

// Expose Validator:
module.exports = Validator;
