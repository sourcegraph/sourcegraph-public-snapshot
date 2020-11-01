import esbuild from 'esbuild'
//import sassPlugin from 'esbuild-plugin-sass-modules'
import { sassPlugin } from 'esbuild-sass-plugin'
import cssModulesPlugin from 'esbuild-css-modules-plugin'

esbuild
    .build({
        entryPoints: ['client/web/src/enterprise/main.tsx'],
        bundle: true,
        format: 'esm',
        outdir: 'ui/assets/esbuild',
        plugins: [
            sassPlugin({
                type: 'style',
                includePaths: [
                    '/home/sqs/src/github.com/sourcegraph/sourcegraph/node_modules',
                    '/home/sqs/src/github.com/sourcegraph/sourcegraph/client',
                ],
                basedir: '/home/sqs/src/github.com/sourcegraph/sourcegraph/client',
                transform: css => {
                    // console.log('FOO', css)
                    return css.replace(/'wildcard/g, "'~wildcard")
                },
                // resolveMap: { wildcard: '/home/sqs/src/github.com/sourcegraph/sourcegraph/client/wildcard' },
            }),

            cssModulesPlugin({
                inject: false,
                localsConvention: 'camelCaseOnly', // optional. value could be one of 'camelCaseOnly', 'camelCase', 'dashes', 'dashesOnly', default is 'camelCaseOnly'
            }),
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
