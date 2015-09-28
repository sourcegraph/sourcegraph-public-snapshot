var
	fs = require('fs'),
	Validator = require('./../../lib/Validator'),
	validator
;

exports.tests = {
	'should validate its self ;-)': function(test) {
		var
			files = [],
			path
		;

		// Get files in source folder:
		path = __dirname + '/../../lib/';
		fs.readdirSync(path)
			.forEach(function(file) {
				if (fs.statSync(path + file).isFile()) {
					files.push(path + file);
				}
			});

		path = __dirname + '/../../lib/constants/';
		fs.readdirSync(path)
			.forEach(function(file) {
				if (fs.statSync(path + file).isFile()) {
					files.push(path + file);
				}
			});


		// Validate these files:
		validator = new Validator({
			editorconfig: __dirname + '/../../.editorconfig',
			ignores: ['js-comments']
		});

		files.forEach(function(file) {
			validator.validate(file);
		});

		test.deepEqual(validator.getInvalidFiles(), {});
		test.done();
	},

	'should throw an exception if path is not a file': function(test) {
		validator = new Validator({trailingspaces: true});

		test.throws(function() { validator.validate(__dirname + '/fixures/'); }, Error);
		test.throws(function() { validator.validate(__dirname); }, Error);
		test.throws(function() { validator.validate('.'); }, Error);
		test.throws(function() { validator.validate('/this/file/does/not/exists'); }, Error);
		test.done();
	}
};
