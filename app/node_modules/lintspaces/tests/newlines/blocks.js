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
	'should report an error when too many newlines are used between blocks': function(test) {
		file = __dirname + '/fixures/blocks.js';
		validator = new Validator({newlineMaximum: 2});
		validator.validate(file);
		report = validator.getInvalidFiles();
		expected = {};
		expected[file] = {
			'8': [merge({}, Messages.NEWLINE_MAXIMUM, {
				message: Messages.NEWLINE_MAXIMUM
					.message
					.replace('{a}', 3)
					.replace('{b}', 2),
				line: 8,
				payload: {
					amount: 3,
					maximum: 2
				}
			})],
			'16': [merge({}, Messages.NEWLINE_MAXIMUM, {
				message: Messages.NEWLINE_MAXIMUM
					.message
					.replace('{a}', 4)
					.replace('{b}', 2),
				line: 16,
				payload: {
					amount: 4,
					maximum: 2
				}
			})]
		};

		test.deepEqual(report, expected);
		test.done();
	},

	'should have no reports when file is valid': function(test) {
		file = __dirname + '/fixures/blocks.js';
		validator = new Validator({newlineMaximum: 4});
		validator.validate(file);
		report = validator.getInvalidFiles();

		test.deepEqual(report, {});
		test.done();
	},

	'should throw an exception if newlineMaximum is less than 0': function(test) {
		file = __dirname + '/fixures/blocks.js';
		test.throws(function() {
			validator = new Validator({newlineMaximum: -2});
			validator.validate(file);
		}, Error);

		test.done();
	}
};
