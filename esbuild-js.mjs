import esbuild from 'esbuild'
import path from 'path'
//import sassPlugin from 'esbuild-plugin-sass-modules'
// import { sassPlugin } from 'esbuild-sass-plugin'
import cssModulesPlugin from 'esbuild-css-modules-plugin'
import sass from 'sass'
import postcss from 'postcss'
import postcssConfig from './postcss.config.js'
import cssModules from 'postcss-modules'
import fs from 'fs'

/** @type esbuild.Plugin */
const examplePlugin = {
    name: 'example',
    setup: build => {
        build.onResolve({ filter: /./, namespace: 'file' }, args => {
            if (args.path.endsWith('.css')) {
                console.log('onResolve', args)
            }
        })
    },
}

/** @type esbuild.Plugin */
const sassPlugin = {
    name: 'sass',
    setup: build => {
        build.onResolve({ filter: /\.scss$/ }, args => ({
            path: path.resolve(args.resolveDir, args.path),
            namespace: 'sass',
        }))
        build.onLoad({ filter: /./, namespace: 'sass' }, async args => {
            const { css: rawCSS } = sass.renderSync({
                file: args.path,
                includePaths: ['node_modules', 'client'],
                importer: (url, prev, done) => {
                    if (url.startsWith('wildcard/')) {
                        return { file: `client/${url}` }
                    }
                    if (url.startsWith('@reach') || url.startsWith('graphiql') || url.startsWith('@sourcegraph')) {
                        return { file: `node_modules/${url}` }
                    }
                    return { file: url }
                },
            })

            let cssModulesJSON = null
            const { css } = await postcss({
                ...postcssConfig,
                plugins: [
                    ...postcssConfig.plugins,
                    cssModules({
                        localsConvention: 'camelCaseOnly',
                        getJSON: (cssSourceFile, json) => {
                            cssModulesJSON = { ...json }
                            return cssModulesJSON
                        },
                    }),
                ],
            }).process(rawCSS, { from: args.path })

            if (cssModulesJSON) {
                const basename = path.basename(args.path, '.scss')
                const cssModulePath = path.resolve(build.initialOptions.outdir, basename + '.json')
                console.log('WRITE', cssModulePath)
                fs.writeFileSync(cssModulePath, JSON.stringify(cssModulesJSON))
            }

            return {
                contents: css,
                loader: 'css',
                resolveDir: path.dirname(args.path),
            }
        })
    },
}

esbuild
    .build({
        entryPoints: ['client/web/src/enterprise/main.tsx'],
        bundle: true,
        format: 'esm',
        outdir: 'ui/assets/esbuild',
        plugins: [
            /* sassPlugin({
                type: 'style',
                includePaths: ['node_modules', 'client'],

                importer: (url, prev, done) => {
                    if (url.startsWith('wildcard/')) {
                        return { file: `client/${url}` }
                    }
                    if (url.startsWith('@reach') || url.startsWith('graphiql') || url.startsWith('@sourcegraph')) {
                        return { file: `node_modules/${url}` }
                    }
                    return { file: url }
                },
                // resolveMap: { wildcard: '/home/sqs/src/github.com/sourcegraph/sourcegraph/client/wildcard' },
            }), */
            sassPlugin,
            examplePlugin,

            /* cssModulesPlugin({
                inject: false,
                localsConvention: 'camelCaseOnly', // optional. value could be one of 'camelCaseOnly', 'camelCase', 'dashes', 'dashesOnly', default is 'camelCaseOnly'
            }), */
        ],
        define: {
            'process.env.NODE_ENV': '"development"',
            global: 'window',
            'process.env.SOURCEGRAPH_API_URL': '"' + process.env.SOURCEGRAPH_API_URL + '"',
        },
        splitting: true,
        loader: {
            '.yaml': 'text',
            // '.scss': 'text',
            '.ttf': 'dataurl',
            '.png': 'dataurl',
        },
        target: 'es2020',
        sourcemap: true,
    })
    .catch(e => console.error(e.message))
