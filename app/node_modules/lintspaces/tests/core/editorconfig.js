var
	Validator = require('./../../lib/Validator'),
	validator,
	options
;

exports.tests = {
	'should override the settings by editorconfig': function(test) {
		options = {
			trailingspaces: false,
			newlineMaximum: false,
			indentation: 'spaces',
			spaces: 2,
			newline: false,
			ignores: ['js-comments'],
			editorconfig: '.editorconfig'
		};

		// fake loading:
		validator = new Validator(options);
		validator._path = __dirname + '/editorconfig.js';
		validator._loadSettings();

		// newline:
		test.ok(validator._settings.newline !== options.newline);
		test.equal(validator._settings.newline, true);

		// indentation will be overwritten:
		test.ok(validator._settings.indentation !== options.indentation);
		test.equal(validator._settings.indentation, 'tabs');

		// spaces will be overwritten:
		test.ok(validator._settings.spaces !== options.spaces);
		test.equal(validator._settings.spaces, 'tab');

		// trailingspaces will be overwritten:
		test.ok(validator._settings.trailingspaces !== options.trailingspaces);
		test.equal(validator._settings.trailingspaces, true);

		// newlineMaximum will be unchanged:
		test.equal(validator._settings.newlineMaximum, options.newlineMaximum);

		test.done();
	},

	'should throw an exception if editorconfig has no valid path': function(test) {
		test.throws(function() {
			validator = new Validator({editorconfig: '.'});
			validator.validate(__filename);
		}, Error);

		test.throws(function() {
			validator = new Validator({editorconfig: __dirname});
			validator.validate(__filename);
		}, Error);

		test.throws(function() {
			validator = new Validator({editorconfig: __dirname + '/path/that/doesnt/existis/.editorconfig'});
			validator.validate(__filename);
		}, Error);

		test.done();
	}
};
