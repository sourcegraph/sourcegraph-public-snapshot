var
	merge = require('merge'),
	Messages = require('./../../lib/constants/messages'),
	Validator = require('./../../lib/Validator'),
	validator,
	report,
	expected,
	file
;

exports.tests = {
	'should report errors when tabs are used': function(test) {
		file = __dirname + '/fixures/tabs-valid.js';
		validator = new Validator({indentation: 'spaces'});
		validator.validate(file);
		report = validator.getInvalidFiles();
		expected = {};
		expected[file] = {
			'2': [merge({}, Messages.INDENTATION_SPACES, {line: 2})],
			'3': [merge({}, Messages.INDENTATION_SPACES, {line: 3})],
			'4': [merge({}, Messages.INDENTATION_SPACES, {line: 4})],
			'5': [merge({}, Messages.INDENTATION_SPACES, {line: 5})],
			'6': [merge({}, Messages.INDENTATION_SPACES, {line: 6})],
			'7': [merge({}, Messages.INDENTATION_SPACES, {line: 7})],
		};

		test.deepEqual(report, expected);
		test.done();
	},

	'should report errors when tabs and spaces are mixed': function(test) {
		file = __dirname + '/fixures/mixed.js';
		validator = new Validator({indentation: 'spaces'});
		validator.validate(file);
		report = validator.getInvalidFiles();
		expected = {};
		expected[file] = {
			'2': [merge({}, Messages.INDENTATION_SPACES, {line: 2})],
			'3': [merge({}, Messages.INDENTATION_SPACES, {line: 3})],
			'4': [merge({}, Messages.INDENTATION_SPACES, {line: 4})],
			'5': [merge({}, Messages.INDENTATION_SPACES, {line: 5})],
			'7': [merge({}, Messages.INDENTATION_SPACES, {line: 7})],
		};

		test.deepEqual(report, expected);
		test.done();
	},

	'should report errors when indentation is incorrect': function(test) {
		file = __dirname + '/fixures/spaces-invalid-indention.js';
		validator = new Validator({
			indentation: 'spaces',
			spaces: 4
		});
		validator.validate(file);
		report = validator.getInvalidFiles();
		expected = {};
		expected[file] = {
			'3': [merge({}, Messages.INDENTATION_SPACES_AMOUNT, {
				message: Messages
					.INDENTATION_SPACES_AMOUNT
					.message
					.replace('{a}', 4)
					.replace('{b}', 5),
				line: 3,
				payload: {
					expected: 4,
					indent: 5
				}
			})],
			'5': [merge({}, Messages.INDENTATION_SPACES_AMOUNT, {
				message: Messages
					.INDENTATION_SPACES_AMOUNT
					.message
					.replace('{a}', 12)
					.replace('{b}', 10),
				line: 5,
				payload: {
					expected: 12,
					indent: 10
				}
			})]
		};

		test.deepEqual(report, expected);
		test.done();
	},

	'should have no reports when file is valid': function(test) {
		file = __dirname + '/fixures/spaces-valid.js';
		validator = new Validator({
			indentation: 'spaces'
		});
		validator.validate(file);
		report = validator.getInvalidFiles();

		test.deepEqual(report, {});
		test.done();
	},

	'should have no reports when file with BOM is valid and BOM is allowed': function(test) {
		file = __dirname + '/fixures/spaces-bom-valid.js';
		validator = new Validator({
			indentation: 'spaces',
			allowsBOM: true
		});
		validator.validate(file);
		report = validator.getInvalidFiles();

		test.deepEqual(report, {});
		test.done();
	},

	'should report an error when file with BOM is not allowed': function(test) {
		file = __dirname + '/fixures/spaces-bom-valid.js';
		validator = new Validator({
			indentation: 'spaces',
			allowsBOM: false
		});
		validator.validate(file);
		report = validator.getInvalidFiles();

		expected = {};
		expected[file] = {
			'1': [merge({}, Messages.INDENTATION_SPACES, {line: 1})]
		};

		test.deepEqual(report, expected);
		test.done();
	}
};
