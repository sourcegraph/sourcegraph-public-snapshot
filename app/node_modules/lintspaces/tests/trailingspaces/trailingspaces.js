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
	'should report errors when trailingspaces found': function(test) {
		file = __dirname + '/fixures/invalid.js';
		validator = new Validator({trailingspaces: true});
		validator.validate(file);
		report = validator.getInvalidFiles();
		expected = {};
		expected[file] = {
			'2': [merge({}, Messages.TRAILINGSPACES, {line: 2})],
			'5': [merge({}, Messages.TRAILINGSPACES, {line: 5})],
			'8': [merge({}, Messages.TRAILINGSPACES, {line: 8})]
		};

		test.deepEqual(report, expected);
		test.done();
	},

	'should have no reports when file is valid': function(test) {
		file = __dirname + '/fixures/valid.js';
		validator = new Validator({trailingspaces: true});
		validator.validate(file);
		report = validator.getInvalidFiles();

		test.deepEqual(report, {});
		test.done();
	},

	'should ignore trailingspaces inside comments when option is set': function(test) {
		file = __dirname + '/fixures/ignores.js';
		validator = new Validator({
			trailingspaces: true,
			trailingspacesToIgnores: true,
			ignores: ['js-comments']
		});

		validator.validate(file);
		report = validator.getInvalidFiles();
		expected = {};
		expected[file] = {
			'9': [merge({}, Messages.TRAILINGSPACES, {line: 9})], // singleline
			'11': [merge({}, Messages.TRAILINGSPACES, {line: 11})], // singleline
			'15': [merge({}, Messages.TRAILINGSPACES, {line: 15})], // multiline
			'17': [merge({}, Messages.TRAILINGSPACES, {line: 17})], // singleline
			'21': [merge({}, Messages.TRAILINGSPACES, {line: 21})], // multiline
			'26': [merge({}, Messages.TRAILINGSPACES, {line: 26})] // multiline
		};

		test.deepEqual(expected, report);
		test.done();
	},

	'should ignore trailingspaces on blank lines when option is set': function(test) {
		file = __dirname + '/fixures/invalid.js';
		validator = new Validator({
			trailingspaces: true,
			trailingspacesSkipBlanks: true
		});
		validator.validate(file);
		report = validator.getInvalidFiles();
		expected = {};
		expected[file] = {
			'2': [merge({}, Messages.TRAILINGSPACES, {line: 2})],
			'8': [merge({}, Messages.TRAILINGSPACES, {line: 8})]
		};

		test.deepEqual(expected, report);
		test.done();
	}
};
