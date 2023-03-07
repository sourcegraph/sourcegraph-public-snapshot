// @ts-check

/** @type {import('eslint').ESLint.ConfigData} */
const config = {
	root: true,
	extends: ['@sourcegraph/eslint-config'],
	parserOptions: {
		ecmaVersion: 2018,
		sourceType: 'module',
		EXPERIMENTAL_useSourceOfProjectReferenceRedirect: true,
		project: 'tsconfig.json',
	},
	settings: {
		react: { version: '18.0' },
	},
	rules: {
		'@typescript-eslint/naming-convention': 'warn',
		curly: 'warn',
		eqeqeq: 'warn',
		'no-throw-literal': 'warn',

		// TODO(sqs): There were ~484 problems identified when the eslint config was updated to use
		// @sourcegraph/eslint-config. To let us gradually fix these, I've turned most rules with
		// problems to warn:
		'@typescript-eslint/explicit-member-accessibility': 'warn',
		'id-length': 'warn',
		'ban/ban': 'warn',
		'@typescript-eslint/require-await': 'warn',
		'@typescript-eslint/no-misused-promises': 'warn',
		'@typescript-eslint/no-floating-promises': 'warn',
		'@typescript-eslint/explicit-function-return-type': 'warn',
		'prefer-promise-reject-errors': 'warn',
		'etc/throw-error': 'warn',
		'@typescript-eslint/await-thenable': 'warn',
		'no-async-promise-executor': 'warn',
		'@typescript-eslint/no-explicit-any': 'warn',
		'@typescript-eslint/no-unused-vars': 'warn',
		'@typescript-eslint/no-unsafe-member-access': 'warn',
		'@typescript-eslint/no-base-to-string': 'warn',
		'@typescript-eslint/no-empty-function': 'warn',
		'no-unused-expressions': 'warn',
		'no-useless-concat': 'warn',
		'no-useless-escape': 'warn',
		'@typescript-eslint/no-var-requires': 'warn',
		'no-void': ['warn', { allowAsStatement: true }],
		'@typescript-eslint/no-require-imports': 'warn',
		radix: 'warn',
		'@typescript-eslint/restrict-template-expressions': 'warn',
		'@typescript-eslint/restrict-plus-operands': 'warn',
		'@typescript-eslint/prefer-optional-chain': 'warn',
		'unicorn/prefer-string-slice': 'warn',
		'unicorn/prefer-query-selector': 'warn',
	},
	ignorePatterns: ['out', '**/.eslintrc.js', '**/prettier.config.js', '**/*.d.ts'],
}

module.exports = config
