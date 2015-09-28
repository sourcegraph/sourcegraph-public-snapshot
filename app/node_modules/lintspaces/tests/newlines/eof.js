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
	'should report an error when newline is missing': function(test) {
		file = __dirname + '/fixures/eof-missing.js';
		validator = new Validator({newline: true});
		validator.validate(file);
		report = validator.getInvalidFiles();
		expected = {};
		expected[file] = {
			'8': [merge({}, Messages.NEWLINE, {line: 8})]
		};

		test.deepEqual(report, expected);
		test.done();
	},

	'should report an error when too much newlines are used': function(test) {
		file = __dirname + '/fixures/eof-toomuch.js';
		validator = new Validator({newline: true});
		validator.validate(file);
		report = validator.getInvalidFiles();
		expected = {};
		expected[file] = {
			'12': [merge({}, Messages.NEWLINE_AMOUNT, {line: 12})]
		};

		test.deepEqual(report, expected);
		test.done();
	},

	'should have no reports when file is valid': function(test) {
		file = __dirname + '/fixures/eof-valid.js';
		validator = new Validator({newline: true});
		validator.validate(file);
		report = validator.getInvalidFiles();

		test.deepEqual(report, {});
		test.done();
	}
};
