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
	'should fail, when no ignore is set': function(test) {
		file = __dirname + '/fixures/buildin.js';
		validator = new Validator({
			indentation: 'tabs',
			trailingspaces: true
		});
		validator.validate(file);
		report = validator.getInvalidFiles();
		expected = {};
		expected[file] = {
			'2': [merge({}, Messages.INDENTATION_TABS, {line: 2})],
			'3': [
				merge({}, Messages.INDENTATION_TABS, {line: 3}),
				merge({}, Messages.TRAILINGSPACES, {line: 3})
			],
			'4': [merge({}, Messages.INDENTATION_TABS, {line: 4})],
			'10': [merge({}, Messages.INDENTATION_TABS, {line: 10})],
			'11': [
				merge({}, Messages.INDENTATION_TABS, {line: 11}),
				merge({}, Messages.TRAILINGSPACES, {line: 11})
			],
			'12': [merge({}, Messages.INDENTATION_TABS, {line: 12})],
			'16': [merge({}, Messages.INDENTATION_TABS, {line: 16})],
			'17': [merge({}, Messages.INDENTATION_TABS, {line: 17})],
			'18': [merge({}, Messages.INDENTATION_TABS, {line: 18})],
			'19': [
				merge({}, Messages.INDENTATION_TABS, {line: 19}),
				merge({}, Messages.TRAILINGSPACES, {line: 19})
			],
			'20': [merge({}, Messages.INDENTATION_TABS, {line: 20})],
			'27': [merge({}, Messages.INDENTATION_TABS, {line: 27})],
			'28': [
				merge({}, Messages.INDENTATION_TABS, {line: 28}),
				merge({}, Messages.TRAILINGSPACES, {line: 28})
			],
			'29': [merge({}, Messages.INDENTATION_TABS, {line: 29})]
		};

		test.deepEqual(report, expected);
		test.done();
	},

	'should fail on expected lines, when ignore is set': function(test) {
		// this should fail because there is one invalid line and there
		// are also trainingspaces
		file = __dirname + '/fixures/buildin.js';
		validator = new Validator({
			indentation: 'tabs',
			trailingspaces: true,
			ignores: ['js-comments']
		});

		validator.validate(file);
		report = validator.getInvalidFiles();
		expected = {};
		expected[file] = {
			'3': [merge({}, Messages.TRAILINGSPACES, {line: 3})],
			'11': [merge({}, Messages.TRAILINGSPACES, {line: 11})],
			'16': [merge({}, Messages.INDENTATION_TABS, {line: 16})],
			'19': [merge({}, Messages.TRAILINGSPACES, {line: 19})],
			'28': [merge({}, Messages.TRAILINGSPACES, {line: 28})]
		};

		test.deepEqual(report, expected);
		test.done();
	},

	'should fail on expected lines, when ignore and maximum newlines is set': function(test) {
		// this should fail because there is one invalid line and there
		// are also trainingspaces
		file = __dirname + '/fixures/buildin.js';
		validator = new Validator({
			indentation: 'tabs',
			trailingspaces: true,
			newlineMaximum: 2,
			ignores: ['js-comments']
		});

		validator.validate(file);
		report = validator.getInvalidFiles();
		expected = {};
		expected[file] = {
			'3': [merge({}, Messages.TRAILINGSPACES, {line: 3})],
			'11': [merge({}, Messages.TRAILINGSPACES, {line: 11})],
			'16': [merge({}, Messages.INDENTATION_TABS, {line: 16})],
			'19': [merge({}, Messages.TRAILINGSPACES, {line: 19})],
			'28': [merge({}, Messages.TRAILINGSPACES, {line: 28})]
		};

		test.deepEqual(report, expected);
		test.done();
	}
};
