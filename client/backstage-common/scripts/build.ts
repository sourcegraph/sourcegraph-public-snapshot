import { existsSync } from 'fs'
import path from 'path'

import * as esbuild from 'esbuild'
import { rm } from 'shelljs'

import { buildTimerPlugin } from '@sourcegraph/build-config'

const distributionPath = path.resolve(__dirname, '..', 'dist')

  ; (async function build(): Promise<void> {
    if (existsSync(distributionPath)) {
      rm('-rf', distributionPath)
    }

    await esbuild.build({
      entryPoints: [path.resolve(__dirname, '..', 'src', 'index.ts')],
      bundle: true,
      external: ["@backstage/cli", "@backstage/catalog-model", "@backstage/config", "@backstage/plugin-catalog-backend", "graphql-request", "lodash", "aptllo"],
      // If we don't specify module first, esbuild bundles somethings "incorrectly" and you'll get a error with './impl/format' error
      mainFields: ['module', 'main'],
      format: 'cjs',
      platform: 'node',
      define: {
        'process.env.IS_TEST': 'false',
        global: 'globalThis',
      },
      //splitting: false,
      //inject: ['./scripts/react-shim.js', './scripts/process-shim.js', './scripts/buffer-shim.js'],
      plugins: [
        //stylePlugin,
        //workerPlugin,
        /*packageResolutionPlugin({
          process: require.resolve('process/browser'),
          path: require.resolve('path-browserify'),
          http: require.resolve('stream-http'),
          https: require.resolve('https-browserify'),
          url: require.resolve('url'),
          util: require.resolve('util'),
        }),*/
        buildTimerPlugin,
      ],
      // loader: {
      //   '.ttf': 'file',
      // },
      ignoreAnnotations: true,
      treeShaking: true,
      watch: !!process.env.WATCH,
      sourcemap: true,
      outdir: distributionPath,
      // outExtension: {
      //   ".js": ".cjs"
      // }
    })
  })()
