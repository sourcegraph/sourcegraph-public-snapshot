// @ts-check

const path = require('path')

// TODO(bazel): drop when non-bazel removed.
const IS_BAZEL = !!process.env.BAZEL_TEST
const SRC_EXT = IS_BAZEL ? 'js' : 'ts'
const rootDir = IS_BAZEL ? process.cwd() : __dirname

// Use the same locale for test runs so that snapshots generated using code that
// uses Intl or toLocaleString() are consistent.
//
// We have to do this at this point instead of in the normal Jest setup hooks
// because ICU (and therefore Intl) only reads the environment once:
// specifically, at the point any Intl object is instantiated. Jest indirectly
// uses a locale-sensitive sort while setting up test suites _before_ invoking
// the setup hooks, so we have no opportunity to change $LANG outside of this
// ugly side effect. (This is especially evident when running tests in-band.)
process.env.LANG = 'en_US.UTF-8'

// TODO(bazel): use esm modules and remove transform exclusion.
const ESM_NPM_DEPS = [
  'abortable-rx',
  '@sourcegraph/comlink',
  'monaco-editor',
  'monaco-yaml',
  'marked',
  'date-fns',
  'react-sticky-box',
  'uuid',
  'vscode-languageserver-types',
]
  .join('|')
  .replace(/\//g, '\\+')

/** @type {import('@jest/types').Config.InitialOptions} */
const config = {
  // uses latest jsdom and exposes jsdom as a global,
  // for example to change the URL in window.location
  testEnvironment: path.join(rootDir, 'client/shared/dev/jest-environment.js'),

  collectCoverage: !!process.env.CI,
  collectCoverageFrom: [`<rootDir>/src/**/*.{${SRC_EXT},${SRC_EXT}x}`],
  coverageDirectory: '<rootDir>/coverage',
  coveragePathIgnorePatterns: [/\/node_modules\//.source, /\.(test|story)\.{jsx,tsx}?$/.source, /\.d\.ts$/.source],
  roots: ['<rootDir>/src'],

  transform: { '\\.[jt]sx?$': ['babel-jest', { root: rootDir }] },

  // TODO(bazel): use esm modules and remove transforms.
  // PNPM style rules_js version.
  // See pnpm notes at https://jestjs.io/docs/configuration#transformignorepatterns-arraystring
  transformIgnorePatterns: [
    // packages within the root pnpm/rules_js package store
    `<rootDir>/node_modules/.(aspect_rules_js|pnpm)/(?!(${ESM_NPM_DEPS})@)`,
    // files under a subdir: eg. '/packages/lib-a/'
    `(../)+node_modules/.(aspect_rules_js|pnpm)/(?!(${ESM_NPM_DEPS})@)`,
    // packages nested within another
    `node_modules/(?!.aspect_rules_js|.pnpm|${ESM_NPM_DEPS})`,
  ],

  moduleNameMapper: {
    '\\.s?css$': 'identity-obj-proxy',
    '\\.ya?ml$': 'identity-obj-proxy',
    '\\.svg$': 'identity-obj-proxy',
    '^worker-loader': 'identity-obj-proxy',
    // monaco-editor uses the "module" field in package.json, which isn't supported by Jest
    // https://github.com/facebook/jest/issues/2702
    // https://github.com/Microsoft/monaco-editor/issues/996
    '^monaco-editor': 'monaco-editor/esm/vs/editor/editor.main.js',
  },
  modulePaths: ['node_modules', '<rootDir>/src'],

  // By default, don't clutter `pnpm run test --watch` output with the full coverage table. To see it, use the
  // `--coverageReporters text` jest option.
  coverageReporters: ['json', 'lcov', 'text-summary'],

  setupFiles: [
    path.join(rootDir, 'client/shared/dev/mockDate.js'),
    // Needed for reusing API functions that use fetch
    // Neither NodeJS nor JSDOM have fetch + AbortController yet
    require.resolve('abort-controller/polyfill'),
    path.join(rootDir, 'client/shared/dev/fetch'),
    path.join(rootDir, 'client/shared/dev/setLinkComponentForTest.ts'),
    path.join(rootDir, 'client/shared/dev/mockDomRect.ts'),
    path.join(rootDir, 'client/shared/dev/mockResizeObserver.ts'),
    path.join(rootDir, 'client/shared/dev/mockUniqueId.ts'),
    path.join(rootDir, 'client/shared/dev/mockSentryBrowser.ts'),
    path.join(rootDir, 'client/shared/dev/mockMatchMedia.ts'),
  ].map(file => IS_BAZEL ? file.replace('.ts', '') : file),
  setupFilesAfterEnv: [
    require.resolve('core-js/stable'),
    require.resolve('regenerator-runtime/runtime'),
    require.resolve('@testing-library/jest-dom'),
    path.join(rootDir, 'client/shared/dev/reactCleanup.ts'),
  ].map(file => IS_BAZEL ? file.replace('.ts', '') : file),
  globalSetup: path.join(rootDir, 'client/shared/dev/jestGlobalSetup.js'),
  globals: {
    Uint8Array,
  },
}

module.exports = config
