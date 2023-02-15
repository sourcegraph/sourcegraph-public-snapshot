import { existsSync } from 'fs'
import path from 'path'
import { nodeExternalsPlugin } from 'esbuild-node-externals'

import * as esbuild from 'esbuild'
import { rm } from 'shelljs'

import {
  packageResolutionPlugin,
  stylePlugin,
  workerPlugin,
  RXJS_RESOLUTIONS,
  buildTimerPlugin,
} from '@sourcegraph/build-config'


const distributionPath = path.resolve(__dirname, '..', 'dist')

async function build(): Promise<void> {
  if (existsSync(distributionPath)) {
    rm('-rf', distributionPath)
  }

  await esbuild.build({
    entryPoints: [path.resolve(__dirname, '..', 'src', 'index.ts')],
    bundle: true,
    format: 'esm',
    logLevel: 'error',
    jsx: 'automatic',
    external: [
      '@backstage/*',
      '@material-ui/*',
      'react-use',
      'react',
      'react-dom',
    ],
    plugins: [
      stylePlugin,
      workerPlugin,
      buildTimerPlugin,
      nodeExternalsPlugin()
    ],
    // mainFields: ['browser', 'module', 'main'],
    // platform: 'browser',
    define: {
      'process.env.IS_TEST': 'false',
      global: 'window',
    },
    loader: {
      '.yaml': 'text',
      '.ttf': 'file',
      '.png': 'file',
    },
    treeShaking: true,
    target: 'esnext',
    sourcemap: true,
    outdir: distributionPath,
  })
}

if (require.main == module) {
  build()
    .catch(error => console.error('Error:', error))
    .finally(() => process.exit(0))
}
