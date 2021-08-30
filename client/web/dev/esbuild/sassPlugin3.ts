import fs from 'fs'
import os from 'os'
import path from 'path'

import esbuild from 'esbuild'
import postcss from 'postcss'
import postcssModules from 'postcss-modules'
import sass from 'sass'

import postcssConfig from '../../../../postcss.config'

const rootPath = path.resolve(__dirname, '..', '..', '..', '..')

const resolveFile = (modulePath: string, directory: string): string => {
    if (modulePath.startsWith('.')) {
        return path.resolve(directory, modulePath)
    }

    if (modulePath.startsWith('wildcard/') || modulePath.startsWith('shared/')) {
        return path.resolve(rootPath, `client/${modulePath}`)
    }

    let p = path.resolve(rootPath, `node_modules/${modulePath}`)
    try {
        p = fs.realpathSync(p)
    } catch {}
    return p
}

export const sassPlugin: esbuild.Plugin = {
    name: 'sass',
    setup: build => {
        const modulesMap = new Map<string, string>()
        const modulesPlugin = postcssModules({
            generateScopedName: '[name]__[local]', // TODO(sqs): omit hash for local dev
            localsConvention: 'camelCase',
            getJSON: (cssPath: string, json: any) => modulesMap.set(cssPath, JSON.stringify(json)),
        })

        const cssRender = async (sourceFullPath: string, fileContent: string): Promise<string> => {
            const css = sourceFullPath.endsWith('.scss') // renderSync is ~20% faster than render (because it's blocked on CPU, not IO).
                ? sass
                      .renderSync({
                          file: sourceFullPath,
                          data: fileContent,
                          importer: (url, directory) => ({ file: resolveFile(url, path.dirname(directory)) }),
                      })
                      .css.toString()
                : fileContent

            const result = await postcss({
                ...postcssConfig,
                plugins:
                    sourceFullPath.endsWith('.module.css') || sourceFullPath.endsWith('.module.scss')
                        ? [...postcssConfig.plugins, modulesPlugin]
                        : postcssConfig.plugins,
            }).process(css, {
                from: sourceFullPath,
            })
            return result.css
        }
        const cssRenderCache = new Map<string, { path: string; originalContent: string; output: string }>()
        const cachedCSSRender = async (sourceFullPath: string, fileContent: string): Promise<string> => {
            // TODO(sqs): invalidate
            const key = sourceFullPath
            const existing = cssRenderCache.get(key)
            if (existing && existing.originalContent === fileContent) {
                return existing.output
            }

            const t0 = Date.now()
            const output = await cssRender(sourceFullPath, fileContent)
            if (Date.now() - t0 > 15) {
                cssRenderCache.set(key, { path: sourceFullPath, originalContent: fileContent, output })
                // console.log('SLOW', Date.now() - t0)
            } else {
                // console.log('FAST', Date.now() - t0)
            }
            return output
        }

        build.onResolve({ filter: /\.scss$/, namespace: 'file' }, async args => {
            const fullPath = resolveFile(args.path, args.resolveDir)
            const contents = await cachedCSSRender(fullPath, await fs.promises.readFile(fullPath, 'utf8'))

            return {
                path: fullPath.replace(/\.scss$/, '.css'),
                namespace: 'css',
                pluginData: {
                    contents,
                },
            }
        })

        // Resolve CSS modules imported by the next onLoad callback.
        build.onResolve({ filter: /\.css$/, namespace: 'css' }, args => {
            if (args.pluginData?.contents !== undefined) {
                return {
                    path: args.path.replace(/\.module/, 'RAWCSS'), // TODO(sqs): hack
                    namespace: 'css',
                    pluginData: {
                        contents: args.pluginData.contents,
                    },
                }
            }
            return
        })

        build.onLoad({ filter: /\.css$/, namespace: 'css' }, args => {
            const isModule = args.path.includes('.module')
            if (isModule) {
                return {
                    contents: `
      import ${JSON.stringify(args.path)}
      export default ${modulesMap.get(args.path.replace(/\.css$/, '.scss')) || '{}'}`,
                    loader: 'js',
                    resolveDir: path.dirname(args.path),
                    pluginData: args.pluginData,
                    // TODO(sqs): set watchFiles
                }
            }
            return {
                contents: args.pluginData.contents,
                resolveDir: path.dirname(args.path),
                loader: 'css',
            }
        })
    },
}
