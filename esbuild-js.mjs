import esbuild from 'esbuild'
//import sassPlugin from 'esbuild-plugin-sass-modules'
import { sassPlugin } from 'esbuild-sass-plugin'
import cssModulesPlugin from 'esbuild-css-modules-plugin'

/** @type esbuild.Plugin */
const examplePlugin = {
    name: 'example',
    setup: build => {
        build.onResolve({ filter: /./, namespace: 'file' }, args => {
            if (args.path.includes('css')) {
                console.log('onResolve', args)
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
            sassPlugin({
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
            }),

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
        },
        target: 'es2020',
        sourcemap: true,
    })
    .catch(e => console.error(e.message))
