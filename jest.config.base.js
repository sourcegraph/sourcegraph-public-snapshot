// @ts-check

const path = require('path')

// TODO(bazel): drop when non-bazel removed.
const IS_BAZEL = !!(process.env.JS_BINARY__TARGET || process.env.BAZEL_BINDIR || process.env.BAZEL_TEST)
const SRC_EXT = IS_BAZEL ? 'js' : 'ts'
const rootDir = IS_BAZEL ? process.cwd() : __dirname

function toPackagePath(pkgPath) {
  // TODO(bazel): bazel runs tests using the pre-compiled npm packages from
  // the pnpm workspace projects. the legacy non-bazel tests run on local ts
  // source files which are compiled by jest and npm packages are mapped to
  // the compiled or local js files.
  if (IS_BAZEL) {
    return pkgPath.replace('.ts', '').replace('.js', '')
  }

  return pkgPath
}

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

const ESM_NPM_DEPS = [
  'abortable-rx',
  '@sourcegraph/.*',
  'monaco-editor',
  'monaco-yaml',
  '@ampproject/.*',
  'marked',
  'date-fns',
  'react-sticky-box',
  'uuid',
  'vscode-languageserver-types',
].join('|')

/** @type {import('@jest/types').Config.InitialOptions} */
const config = {
  // uses latest jsdom and exposes jsdom as a global,
  // for example to change the URL in window.location
  testEnvironment: toPackagePath(path.join(rootDir, 'client/shared/dev/jest-environment.js')),

  roots: ['<rootDir>/src'],
  snapshotResolver: path.join(rootDir, 'jest.snapshot-resolver.js'),
  snapshotFormat: {
    escapeString: true,
    printBasicPrototype: true,
  },

  injectGlobals: false,

  // Only run JavaScript tests in Bazel; otherwise run TypeScript and JavaScript.
  testMatch: [`**/?(*.)+(spec|test).(js|${SRC_EXT})?(x)`],

  transform: {
    [IS_BAZEL ? '\\.js$' : '\\.[jt]sx?$']: [
      'babel-jest',
      {
        root: rootDir,
        configFile: path.join(rootDir, IS_BAZEL ? 'babel.config.jest.js' : 'babel.config.js'),
      },
    ],
  },

  // Transform packages that do not distribute CommonJS packages (typically because they only distribute ES6
  // modules). If you get an error from jest like "Jest encountered an unexpected token. ... SyntaxError:
  // unexpected token import/export", then add it here. See
  // https://github.com/facebook/create-react-app/issues/5241#issuecomment-426269242 for more information on why
  // this is necessary.
  // Include the pnpm-style rules_js.
  // See pnpm notes at https://jestjs.io/docs/configuration#transformignorepatterns-arraystring
  transformIgnorePatterns: [
    // packages within the root pnpm/rules_js package store
    `<rootDir>/node_modules/.(aspect_rules_js|pnpm)/(?!(${ESM_NPM_DEPS.replace('/', '\\+')})@)`,
    // files under a subdir: eg. '/packages/lib-a/'
    `(../)+node_modules/.(aspect_rules_js|pnpm)/(?!(${ESM_NPM_DEPS.replace('/', '\\+')})@)`,
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

  setupFiles: [
    path.join(rootDir, 'client/shared/dev/mockDate.js'),
    // Needed for reusing API functions that use fetch
    // Neither NodeJS nor JSDOM have fetch + AbortController yet
    require.resolve('abort-controller/polyfill'),
    require.resolve('message-port-polyfill'),
    path.join(rootDir, 'client/shared/dev/fetch'),
    path.join(rootDir, 'client/shared/dev/mockDomRect.ts'),
    path.join(rootDir, 'client/shared/dev/mockResizeObserver.ts'),
    path.join(rootDir, 'client/shared/dev/mockUniqueId.ts'),
    path.join(rootDir, 'client/shared/dev/mockSentryBrowser.ts'),
    path.join(rootDir, 'client/shared/dev/mockMatchMedia.ts'),
  ].map(toPackagePath),

  setupFilesAfterEnv: [
    require.resolve('core-js/stable'),
    require.resolve('regenerator-runtime/runtime'),
    require.resolve('@testing-library/jest-dom/jest-globals'),
    toPackagePath(path.join(rootDir, 'client/shared/dev/reactCleanup.ts')),
  ],
  globalSetup: toPackagePath(path.join(rootDir, 'client/shared/dev/jestGlobalSetup.js')),
  globals: {
    Uint8Array,
  },
}

module.exports = config
