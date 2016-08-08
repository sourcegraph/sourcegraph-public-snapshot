function ValidationError(data, payload) {
	if (!data.line) {
		throw new Error('Supply a linenumber in the data parameter.');
	}

	if (!data.code) {
		throw new Error('Supply an errorcode in the data parameter.');
	}

	if (!data.type) {
		throw new Error('Supply an errortype in the data parameter.');
	}

	if (!data.message) {
		throw new Error('Supply an errormessage in the data parameter.');
	}

	this.line = data.line;
	this.code = data.code;
	this.type = data.type;
	this.message = data.message;

	if (payload) {
		if (typeof payload !== 'object') {
			throw new Error('The existing payload must be an object.');
		}

		this.payload = payload;
	}
}

// Expose Validator:
module.exports = ValidationError;
