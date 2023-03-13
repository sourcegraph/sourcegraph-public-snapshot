// eslint-disable-next-line import/extensions
const baseConfig = require('../.eslintrc.js')
module.exports = {
	extends: '../.eslintrc.js',
	parserOptions: {
		...baseConfig.parserOptions,
		project: [__dirname + '/tsconfig.json'],
	},
	overrides: baseConfig.overrides,
	'react/react-in-jsx-scope': 'off',
	'react/jsx-filename-extension': [1, { extensions: ['.ts', '.tsx'] }],
}
