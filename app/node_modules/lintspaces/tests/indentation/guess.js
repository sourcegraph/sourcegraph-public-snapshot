var
	merge = require('merge'),
	Messages = require('./../../lib/constants/messages'),
	Validator = require('./../../lib/Validator'),
	validator,
	file,
	report,
	expected
;

exports.tests = {
	'should guess incorrect indentation using tabs': function(test) {
		file = __dirname + '/fixures/guess-tabs.js';
		validator = new Validator({
			indentation: 'tabs',
			indentationGuess: true
		});
		validator.validate(file);
		report = validator.getInvalidFiles();
		expected = {};
		expected[file] = {
			'10': [merge({}, Messages.INDENTATION_GUESS, {
				message: Messages
					.INDENTATION_GUESS
					.message
					.replace('{a}', 2)
					.replace('{b}', 4),
				line: 10,
				payload: {
					indentation: 4,
					expected: 2
				}
			})],
			'15': [merge({}, Messages.INDENTATION_GUESS, {
				message: Messages
					.INDENTATION_GUESS
					.message
					.replace('{a}', 3)
					.replace('{b}', 4),
				line: 15,
				payload: {
					indentation: 4,
					expected: 3
				}
			})]
		};

		test.deepEqual(report, expected);
		test.done();
	},

	'should guess incorrect indentation using spaces': function(test) {
		file = __dirname + '/fixures/guess-spaces.js';
		validator = new Validator({
			indentation: 'spaces',
			spaces: 4,
			indentationGuess: true
		});
		validator.validate(file);
		report = validator.getInvalidFiles();
		expected = {};
		expected[file] = {
			'10': [merge({}, Messages.INDENTATION_GUESS, {
				message: Messages
					.INDENTATION_GUESS
					.message
					.replace('{a}', 2)
					.replace('{b}', 4),
				line: 10,
				payload: {
					indentation: 4,
					expected: 2
				}
			})],
			'15': [merge({}, Messages.INDENTATION_GUESS, {
				message: Messages
					.INDENTATION_GUESS
					.message
					.replace('{a}', 3)
					.replace('{b}', 4),
				line: 15,
				payload: {
					indentation: 4,
					expected: 3
				}
			})]
		};

		test.deepEqual(report, expected);
		test.done();
	},

	'should be silent when not activated': function(test) {
		file = __dirname + '/fixures/guess-tabs.js';
		validator = new Validator({
			indentation: 'spaces'
		});

		validator.validate(file);
		report = validator.getInvalidFiles();
		expected = {};
		expected[file] = {
			'2': [merge({}, Messages.INDENTATION_SPACES, {line: 2})],
			'3': [merge({}, Messages.INDENTATION_SPACES, {line: 3})],
			'5': [merge({}, Messages.INDENTATION_SPACES, {line: 5})],
			'6': [merge({}, Messages.INDENTATION_SPACES, {line: 6})],
			'7': [merge({}, Messages.INDENTATION_SPACES, {line: 7})],
			'9': [merge({}, Messages.INDENTATION_SPACES, {line: 9})],
			'10': [merge({}, Messages.INDENTATION_SPACES, {line: 10})],
			'11': [merge({}, Messages.INDENTATION_SPACES, {line: 11})],
			'13': [merge({}, Messages.INDENTATION_SPACES, {line: 13})],
			'14': [merge({}, Messages.INDENTATION_SPACES, {line: 14})],
			'15': [merge({}, Messages.INDENTATION_SPACES, {line: 15})],
			'16': [merge({}, Messages.INDENTATION_SPACES, {line: 16})],
			'17': [merge({}, Messages.INDENTATION_SPACES, {line: 17})],
			'18': [merge({}, Messages.INDENTATION_SPACES, {line: 18})],
			'19': [merge({}, Messages.INDENTATION_SPACES, {line: 19})],
			'20': [merge({}, Messages.INDENTATION_SPACES, {line: 20})],
		};

		test.deepEqual(report, expected);
		test.done();
	}
};
