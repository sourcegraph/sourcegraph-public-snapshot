import adapter from '@sveltejs/adapter-static'
import { vitePreprocess } from '@sveltejs/kit/vite'

/** @type {import('@sveltejs/kit').Config} */
const config = {
  // Consult https://github.com/sveltejs/svelte-preprocess
  // for more information about preprocessors
  //preprocess: preprocess(),
  // Consult https://kit.svelte.dev/docs/integrations#preprocessors
  // for more information about preprocessors
  preprocess: vitePreprocess(),

  vitePlugin: {
    inspector: {
      showToggleButton: 'always',
      toggleButtonPos: 'bottom-right',
    },
  },

  kit: {
    adapter: adapter({
      fallback: 'index.html',
    }),
    alias: {
      // Makes it easier to refer to files outside packages (such as images)
      $root: '../../',
      // Used inside tests for easy access to helpers
      $testdata: 'src/testdata.ts',
      // Makes it easier to refer to files outside packages (such as images)
      $mocks: 'src/testing/mocks.ts',
      // Map node-module to browser version
      path: '../../node_modules/path-browserify',
      // These are directories and cannot be imported from directly in
      // production build. Need to import from _esm5, otherwise there will
      // be runtime compatibility issues.
      'rxjs/operators': '../../node_modules/rxjs/_esm5/operators/index',
      'rxjs/fetch': '../../node_modules/rxjs/_esm5/fetch/index',
      // Without it prod build doesnt work
      '@apollo/client$': '../../node_modules/@apollo/client/index.js',
      lodash: './node_modules/lodash-es',
    },
    typescript: {
      config: config => {
        config.extends = '../../../tsconfig.base.json'
        config.include = [...(config.include ?? []), '../src/**/*.tsx', '../.storybook/*.ts']
      },
    },
  },
}

export default config
