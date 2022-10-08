import path from 'path'

import * as esbuild from 'esbuild'

import { stylePlugin, buildTimerPlugin } from '@sourcegraph/build-config'

const rootPath = path.resolve(__dirname, '../../../')
const sourceboxRootPath = path.join(rootPath, 'client', 'sourcebox')
const sandboxPath = path.join(sourceboxRootPath, 'sandbox')

export async function serve(): Promise<void> {
    const { host, port, wait } = await esbuild.serve(
        { host: 'localhost', port: 3888, servedir: sandboxPath },
        {
            entryPoints: {
                index: path.join(__dirname, 'index.tsx'),
            },
            bundle: true,
            format: 'esm',
            platform: 'browser',
            splitting: true,
            plugins: [stylePlugin, buildTimerPlugin],
            outdir: path.join(sandboxPath, 'dist'),
            assetNames: '[name]',
            sourcemap: true,
        }
    )
    console.log(`Sourcebox sandbox started on http://${host}:${port}`)
    await wait
}

serve().catch(error => {
    console.error('Error:', error)
    process.exit(1)
})
