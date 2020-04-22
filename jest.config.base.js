// @ts-check

const path = require('path')

/** @type {jest.InitialOptions} */
const config = {
  // uses latest jsdom and exposes jsdom as a global,
  // for example to change the URL in window.location
  testEnvironment: __dirname + '/shared/dev/jest-environment.js',

  collectCoverage: !!process.env.CI,
  collectCoverageFrom: ['<rootDir>/src/**/*.{ts,tsx}'],
  coverageDirectory: '<rootDir>/coverage',
  coveragePathIgnorePatterns: [/\.(test|story)\.tsx?$/.source, /\.d\.ts$/.source],
  roots: ['<rootDir>/src'],

  // Transform packages that do not distribute CommonJS packages (typically because they only distribute ES6
  // modules). If you get an error from jest like "Jest encountered an unexpected token. ... SyntaxError:
  // unexpected token import/export", then add it here. See
  // https://github.com/facebook/create-react-app/issues/5241#issuecomment-426269242 for more information on why
  // this is necessary.
  transformIgnorePatterns: [
    '/node_modules/(?!abortable-rx|@sourcegraph/react-loading-spinner|@sourcegraph/codeintellify|@sourcegraph/comlink|monaco-editor)',
  ],

  moduleNameMapper: {
    '\\.s?css$': 'identity-obj-proxy',
    '^worker-loader': 'identity-obj-proxy',
    // monaco-editor uses the "module" field in package.json, which isn't supported by Jest
    // https://github.com/facebook/jest/issues/2702
    // https://github.com/Microsoft/monaco-editor/issues/996
    '^monaco-editor': '<rootDir>/../node_modules/monaco-editor/esm/vs/editor/editor.main.js',
  },

  // By default, don't clutter `yarn test --watch` output with the full coverage table. To see it, use the
  // `--coverageReporters text` jest option.
  coverageReporters: ['json', 'lcov', 'text-summary'],

  setupFiles: [
    path.join(__dirname, 'shared/dev/mockDate.js'),
    path.join(__dirname, 'shared/dev/globalThis.js'),
    // Needed for reusing API functions that use fetch
    // Neither NodeJS nor JSDOM have fetch + AbortController yet
    require.resolve('abort-controller/polyfill'),
    path.join(__dirname, 'shared/dev/fetch'),
    path.join(__dirname, 'shared/dev/setLinkComponentForTest.ts'),
  ],
  setupFilesAfterEnv: [require.resolve('core-js/stable'), require.resolve('regenerator-runtime/runtime')],
  globalSetup: path.join(__dirname, 'shared/dev/jestGlobalSetup.js'),
}

module.exports = config
