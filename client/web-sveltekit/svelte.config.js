import adapter from '@sveltejs/adapter-auto'
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
    adapter: adapter(),
    alias: {
      // Makes it easier to refer to files outside packages (such as images)
      $root: '../../',
      // Somehow these aliases are necessary to make CSS imports work. Otherwise
      // Vite/postcss/whatever tries to the import these relative to the
      // importing file.
      wildcard: '../wildcard/',
      'open-color': '../../node_modules/open-color',
      // These are directories and cannot be imported from directly in
      // production build
      'rxjs/operators': '../../node_modules/rxjs/operators/index',
      'rxjs/fetch': '../../node_modules/rxjs/fetch/index',
      // Map node-module to browser version
      path: 'node_modules/path-browserify',
    },
  },
}

export default config
