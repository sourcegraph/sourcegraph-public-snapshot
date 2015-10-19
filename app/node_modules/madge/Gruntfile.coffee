module.exports = (grunt) ->
	'use strict'

	# Grunt project configuration.
	# ----------------------
	@initConfig
		pkg: grunt.file.readJSON("package.json")

		# Make a release.
		# https://github.com/geddski/grunt-release
		# ----------------------------------------------
		release:
			options:
				commit: true
				tag: true
				push: true
				pushTags: true
				npm: true
				commitMessage: "Bump to version <%= version %>"
				tagMessage: "Tagged <%= version %>"

		# JSHint task.
		# https://github.com/gruntjs/grunt-contrib-jshint
		# -----------------------------------------------
		jshint:
			options:
				jshintrc: ".jshintrc"
			all: ['test/*.js', 'bin', 'lib']

		# Mocha test task.
		# https://github.com/pghalliday/grunt-mocha-test
		# -----------------------------------------
		mochaTest:
			test:
				options:
					reporter: 'spec'
				src: ['test/*.js']

	# Load tasks.
	# -----------
	@loadNpmTasks "grunt-contrib-jshint"
	@loadNpmTasks "grunt-mocha-test"
	@loadNpmTasks "grunt-release"

	# Register tasks.
	# ---------------
	@registerTask "default", [
		"jshint"
		"mochaTest"
	]
