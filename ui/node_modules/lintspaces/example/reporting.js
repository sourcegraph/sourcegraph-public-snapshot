var
	/* This example refers to the local validator file. When you're using
	 * "node-lintspaces" in your project, require('lintspaces') instead of
	 * require('../lib/Validator').
	 */
	Validator = require('../lib/Validator'),
	validator = new Validator({
		newline: true,
		newlineMaximum: 1,
		trailingspaces: true,
		indentation: 'spaces',
		spaces: 4,
		ignores: ['js-comments']
	}),
	files
;

validator.validate('./reporting.js.example');
files = validator.getInvalidFiles();

// For each invalid file:
Object.keys(files).forEach(function(file) {
	var reports = files[file];

	// For each line in file:
	Object.keys(reports).forEach(function(line) {
		var errors = reports[line];

		// A line can contain more than one error, errors are reported in
		// an array. Loop over these errors:
		errors.forEach(function(error) {
			console.error(
				'Error in Line ' + error.line +
				' (' + error.code + ' / ' + error.type + ')' +
				': ' + error.message
			);

			if (error.payload) {
				Object.keys(error.payload).forEach(function(payload) {
					console.error(
						'\t' + payload + ': ' + error.payload[payload]
					);
				});
			}
		});
	});
});
