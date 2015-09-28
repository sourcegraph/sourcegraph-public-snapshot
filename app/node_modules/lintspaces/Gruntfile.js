var
	testableFiles = [
		'tests/**/*.js',
		'!tests/**/fixures/*.js',
	],
	validateableFiles = [
		'Gruntfile.js',
		'lib/**/*.js',
		'tests/**/*.js',
		'!tests/**/fixures/*.js',
		'example/**/*.js'
	],
	docFiles = [
		'./docs/intro.md',
		'./docs/installation.md',
		'./docs/usage.md',
		'./docs/options.md',
		'./docs/functions.md',
		'./docs/contribution.md',
		'./docs/license.md'
	]
;


module.exports = function(grunt) {

	grunt.initConfig({
		pkg: grunt.file.readJSON('package.json'),
	});

	// Node Unit:
	// -------------------------------------------------------------------------
	grunt.config('nodeunit', {
		all: testableFiles
	});
	grunt.loadNpmTasks('grunt-contrib-nodeunit');

	// JS Hint:
	// -------------------------------------------------------------------------
	grunt.config('jshint', {
		all: validateableFiles,
		options: {jshintrc: true}
	});
	grunt.loadNpmTasks('grunt-contrib-jshint');

	// JSCS:
	// -------------------------------------------------------------------------
	grunt.config('jscs', {
		all: {
			src: validateableFiles,
			options: {config: '.jscs.json'}
		}
	});

	grunt.loadNpmTasks('grunt-jscs');

	// Concat:
	// -------------------------------------------------------------------------
	grunt.config('concat', {
		all: {
			src: docFiles,
			dest: 'README.md'
		}
	});

	grunt.loadNpmTasks('grunt-contrib-concat');



	// Setup default tasks:
	// -------------------------------------------------------------------------
	grunt.registerTask('validate', [
		'jshint',
		'jscs'
	]);

	grunt.registerTask('test', [
		'nodeunit'
	]);

	grunt.registerTask('docs', [
		'concat'
	]);

	grunt.registerTask('default', [
		'validate',
		'test',
		'docs'
	]);

};
