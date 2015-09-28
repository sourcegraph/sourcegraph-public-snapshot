/**
 * This file contains contants for reports and exceptions.
 */

var
	types = require('./types')
;

module.exports = {
	// Reports:
	// Reports always should contain
	//  * code (unique error code)
	//  * type (warning or hint)
	//  * message (the message)
	// -------------------------------------------------------------------------
	INDENTATION_TABS: {
		code: 'INDENTATION_TABS',
		type: types.WARNING,
		message: 'Unexpected spaces found.'
	},
	INDENTATION_SPACES: {
		code: 'INDENTATION_SPACES',
		type: types.WARNING,
		message: 'Unexpected tabs found.'
	},
	INDENTATION_SPACES_AMOUNT: {
		code: 'INDENTATION_SPACES_AMOUNT',
		type: types.WARNING,
		message: 'Expected an indentation at {a} instead of at {b}.'
	},
	INDENTATION_GUESS: {
		type: types.HINT,
		code: 'NEWLINE_GUESS',
		message: 'The indentation in this line seems to be incorrect. ' +
			'The expected indention is {a}, but {b} was found.'
	},
	TRAILINGSPACES: {
		code: 'TRAILINGSPACES',
		type: types.WARNING,
		message: 'Unexpected trailing spaces found.'
	},
	NEWLINE: {
		code: 'NEWLINE',
		type: types.WARNING,
		message: 'Expected a newline at the end of the file.'
	},
	NEWLINE_AMOUNT: {
		code: 'NEWLINE_AMOUNT',
		type: types.WARNING,
		message: 'Unexpected additional newlines at the end of the file.'
	},
	NEWLINE_MAXIMUM: {
		code: 'NEWLINE_MAXIMUM',
		type: types.WARNING,
		message: 'Maximum amount of newlines exceeded. Found {a} newlines, ' +
			'expected maximum is {b}.'
	},
	NEWLINE_MAXIMUM_INVALIDVALUE: {
		type: types.WARNING,
		code: 'NEWLINE_MAXIMUM_INVALIDVALUE',
		message: 'The value "{a}" for the maximum of newlines is invalid.'
	},

	// Exceptions:
	// -------------------------------------------------------------------------
	EDITORCONFIG_NOTFOUND: {
		message: 'The config file "{a}" wasn\'t found.'
	},
	PATH_INVALID: {
		message: '"{a}" does not exists.'
	},
	PATH_ISNT_FILE: {
		message: '"{a}" is not a file.'
	}
};
