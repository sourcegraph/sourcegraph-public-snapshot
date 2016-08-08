## Usage

To run the lintspaces validator on one or multiple files take a look at the
following example:

```javascript

	var Validator = require('lintspaces');

	var validator = new Validator({/* options */});
	validator.validate('/path/to/file.ext');
	validator.validate('/path/to/other/file.ext');

	var results = validator.getInvalidFiles();
```

The response of ```getInvalidFiles()``` contains an object. Each key of this
object is a filepath which contains validation errors.

Under each filepath there is an other object with at least one key. Those key(s)
are the specific linenumbers of the file containing an array with errors.

The following lines shows the structure of the validation result in JSON
notation:

```json

	{
		"/path/to/file.ext": {
			"3": [
				{
					"line": 3,
					"code": "INDENTATION_TABS",
					"type": "warning",
					"message": "Unexpected spaces found."
				},
				{
					"line": 3,
					"code": "TRAILINGSPACES",
					"type": "warning",
					"message": "Unexpected trailing spaces found."
				}
			],
			"12": [
				{
					"line": 12,
					"code": "NEWLINE",
					"type": "warning",
					"message": "Expected a newline at the end of the file."
				}
			]
		},
		"/path/to/other/file.ext": {
			"5": [
				{
					"line": 5,
					"code": "NEWLINE_AMOUNT",
					"type": "warning",
					"message": "Unexpected additional newlines at the end of the file."
				}
			]
		}
	}
```
