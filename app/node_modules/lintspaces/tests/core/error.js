var
	ValidatonError = require('./../../lib/ValidationError')
;

exports.tests = {
	'should throw an error when line is missing': function(test) {
		test.throws(function() {
			new ValidatonError({
				code: 'code',
				type: 'warning',
				message: 'message'
			});
		}, Error);
		test.done();
	},

	'should throw an error when code is missing': function(test) {
		test.throws(function() {
			new ValidatonError({
				line: 5,
				type: 'warning',
				message: 'message'
			});
		}, Error);
		test.done();
	},

	'should throw an error when type is missing': function(test) {
		test.throws(function() {
			new ValidatonError({
				line: 5,
				code: 'code',
				message: 'message'
			});
		}, Error);
		test.done();
	},

	'should throw an error when message is missing': function(test) {
		test.throws(function() {
			new ValidatonError({
				line: 5,
				type: 'warning',
				code: 'code'
			});
		}, Error);
		test.done();
	},

	'should throw an error when no data is given': function(test) {
		test.throws(function() {
			new ValidatonError();
		}, Error);
		test.done();
	},

	'should throw an error when payload is not a string': function(test) {
		test.throws(function() {
			new ValidatonError({
				line: 5,
				type: 'warning',
				code: 'code',
				message: 'message'
			}, 'string');
		}, Error);

		test.throws(function() {
			new ValidatonError({
				line: 5,
				type: 'warning',
				code: 'code',
				message: 'message'
			}, 1);
		}, Error);

		test.throws(function() {
			new ValidatonError({
				line: 5,
				type: 'warning',
				code: 'code',
				message: 'message'
			}, true);
		}, Error);

		test.done();
	},

	'should conatin all given data': function(test) {
		var error = new ValidatonError({
			line: 5,
			type: 'warning',
			code: 'code',
			message: 'message'
		}, {
			foo: 'foo',
			bar: 'bar'
		});

		test.ok(error instanceof ValidatonError);
		test.equal(error.line, 5);
		test.equal(error.type, 'warning');
		test.equal(error.code, 'code');
		test.equal(error.message, 'message');
		test.deepEqual(error.payload, {
			foo: 'foo',
			bar: 'bar'
		});

		test.done();
	},
};
