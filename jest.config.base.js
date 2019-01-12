// @ts-check

/** @type {jest.InitialOptions} */
const config = {
  collectCoverage: true,
  coverageDirectory: '<rootDir>/coverage',
  moduleFileExtensions: ['ts', 'tsx', 'js', 'jsx', 'json'],
  preset: 'ts-jest/presets/js-with-ts',
  roots: ['<rootDir>/src'],
  transform: { '^.+\\.[jt]sx?$': 'ts-jest' },

  // Transform packages that do not distribute CommonJS packages (typically because they only distribute ES6
  // modules). If you get an error from jest like "Jest encountered an unexpected token. ... SyntaxError:
  // unexpected token import/export", then add it here. See
  // https://github.com/facebook/create-react-app/issues/5241#issuecomment-426269242 for more information on why
  // this is necessary.
  transformIgnorePatterns: [
    '/node_modules/(?!abortable-rx|@sourcegraph/react-loading-spinner|@sourcegraph/codeintellify)',
  ],

  moduleNameMapper: { '\\.s?css$': 'identity-obj-proxy' },

  // By default, don't clutter `yarn test --watch` output with the full coverage table. To see it, use the
  // `--coverageReporters text` jest option.
  coverageReporters: ['json', 'lcov', 'text-summary'],
}

module.exports = config
