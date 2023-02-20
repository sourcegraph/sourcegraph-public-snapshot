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

  kit: {
    adapter: adapter({
      fallback: 'index.html',
    }),
    alias: {
      // Makes it easier to refer to files outside packages (such as images)
      $root: '../../',
      // Somehow these aliases are necessary to make CSS imports work. Otherwise
      // Vite/postcss/whatever tries to the import these relative to the
      // importing file.
      wildcard: '../wildcard/',
      'open-color': '../../node_modules/open-color',
      // Map node-module to browser version
      path: '../../node_modules/path-browserify',
      // These are directories and cannot be imported from directly in
      // production build. Need to import from _esm5, otherwise there will
      // be runtime compatibility issues.
      'rxjs/operators': '../../node_modules/rxjs/_esm5/operators/index',
      'rxjs/fetch': '../../node_modules/rxjs/_esm5/fetch/index',
      // Without it prod build doesnt work
      '@apollo/client': '../../node_modules/@apollo/client/index.js',
      lodash: './node_modules/lodash-es',
    },
  },
}

export default config
