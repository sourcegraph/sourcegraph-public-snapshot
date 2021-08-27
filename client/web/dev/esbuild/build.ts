import esbuild from 'esbuild'

import { sassPlugin } from './sassPlugin'

export const BUILD_OPTIONS: esbuild.BuildOptions = {
    entryPoints: ['client/web/src/enterprise/main.tsx', 'client/shared/src/api/extension/main.worker.ts'],
    bundle: true,
    format: 'esm',
    outdir: 'ui/assets/esbuild',
    logLevel: 'error',
    splitting: false, // TODO(sqs): need to have splitting:false for main.worker.ts entrypoint
    plugins: [sassPlugin],
    define: {
        'process.env.NODE_ENV': '"development"',
        global: 'window',
        'process.env.SOURCEGRAPH_API_URL': JSON.stringify(process.env.SOURCEGRAPH_API_URL),
    },
    loader: {
        '.yaml': 'text',
        '.ttf': 'file',
        '.png': 'file',
    },
    target: 'es2020',
    sourcemap: true,
    incremental: true,
}
