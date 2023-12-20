import path from 'path'

import react from '@vitejs/plugin-react'
import { UserConfig, defineConfig, mergeConfig } from 'vite'

import { ENVIRONMENT_CONFIG } from '../utils/environment-config'

import { manifestPlugin } from './manifestPlugin'

/** Whether we're running in Bazel. */
const BAZEL = !!process.env.BAZEL_BINDIR

const repoRoot = BAZEL ? process.cwd() : path.join(__dirname, '../../../..')
const clientWebRoot = path.join(repoRoot, 'client/web')

export default defineConfig(() => {
    let config: UserConfig = {
        plugins: [react(), manifestPlugin({ fileName: 'vite-manifest.json' })],
        build: {
            rollupOptions: {
                input: ENVIRONMENT_CONFIG.CODY_APP
                    ? ['src/enterprise/app/main.tsx']
                    : ['src/enterprise/main.tsx', 'src/enterprise/embed/embedMain.tsx']
                          .map(BAZEL ? toJSExtension : String)
                          .map(p => path.join(clientWebRoot, p)),
            },
            sourcemap: true,

            // modulepreload is supported widely enough now (https://caniuse.com/link-rel-modulepreload)
            // and is only relevant for local dev.
            modulePreload: { polyfill: false },

            emptyOutDir: false, // client/web/dist has static assets checked in
        },
        base: '/.assets',
        root: clientWebRoot,
        publicDir: 'dist',
        assetsInclude: ['**/*.yaml'],
        define: {
            ...Object.fromEntries(
                Object.entries({ ...ENVIRONMENT_CONFIG, SOURCEGRAPH_API_URL: undefined }).map(([key, value]) => [
                    `process.env.${key}`,
                    JSON.stringify(value === undefined ? null : value),
                ])
            ),
        },
        optimizeDeps: {
            exclude: [
                // Without addings this Vite throws an error
                'linguist-languages',
            ],
        },
        resolve: {
            alias: {
                path: require.resolve('path-browserify'),
            },
            mainFields: ['browser', 'module', 'main'],
        },
        css: {
            devSourcemap: true,
            preprocessorOptions: {
                scss: {
                    includePaths: [
                        // Our scss files and scss files in client/* often import global styles via @import 'wildcard/src/...'
                        // Adding '..' as load path causes scss to look for these imports in the client folder.
                        // (without it scss @import paths are always relative to the importing file)
                        path.join(clientWebRoot, '..'),
                    ],
                },
            },
            modules: {
                localsConvention: 'camelCaseOnly',
            },
        },
    }

    if (BAZEL) {
        // TODO(sqs): dedupe with client/web-sveltekit

        // Merge settings necessary to make the build work with bazel
        config = mergeConfig(config, {
            resolve: {
                alias: [
                    // When using Bazel, @sourcegraph/* dependencies will refer to the built packages.
                    // These do not contain the source *.module.scss files but still contain import statements
                    // that reference *.scss files. Processing them with vite throws an error unless we
                    // update the imports to reference the corresponding *.css files instead.
                    // Additionally our own source files might reference *.module.scss files, which we also want
                    // to rewrite.
                    {
                        find: /^(.+)\.module\.scss$/,
                        replacement: '$1.module.css',
                        customResolver(source, importer, options) {
                            // The this.resolve(...) part is taken from the @rollup/plugin-alias implementation. Without
                            // it it appears the bundler tries to resolve relative module IDs to the current working
                            // directory.
                            return source.includes('@sourcegraph') || importer?.includes('@sourcegraph/')
                                ? this.resolve(source, importer, { skipSelf: true, ...options }).then(
                                      resolved => resolved || { id: source }
                                  )
                                : null
                        },
                    },

                    // Assume all other *.scss files have been built.
                    {
                        find: /^(.+)\.scss(\?.*)?$/,
                        replacement: '$1.css$2',
                    },
                ],
            },
            ssr: {
                // By default vite treats dependencies that are links to other packages in the monorepo as source code
                // and processes them as well.
                // In a bazel sandbox however all @sourcegraph/* dependencies are built packages and thus not processed
                // by vite without this additional setting.
                // We have to process those files to apply certain "fixes", such as aliases defined in svelte.config.js.
                noExternal: [/@sourcegraph\/.*/],
            },
        } satisfies UserConfig)
    }

    return config
})

function toJSExtension(path: string): string {
    return path.replace(/\.ts$/, '.js').replace(/\.tsx$/, '.js')
}
