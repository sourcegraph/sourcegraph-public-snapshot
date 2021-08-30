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

const temporaryDirectoryPath = fs.mkdtempSync(path.join(os.tmpdir(), 'esbuild-'))
const cleanup = () => fs.rmdirSync(temporaryDirectoryPath, { recursive: true })

export const sassPlugin: esbuild.Plugin = {
    name: 'sass',
    setup: build => {
        let buildStarted: number
        build.onStart(() => {
            buildStarted = Date.now()
        })
        build.onEnd(() => console.log(`> ${Date.now() - buildStarted}ms`))

        const modulesMap = new Map<string, any>()
        const modulesPlugin = postcssModules({
            generateScopedName: '[name]__[local]', // TODO(sqs): omit hash for local dev
            localsConvention: 'camelCase',
            getJSON: (cssPath: string, json: any) => modulesMap.set(cssPath, json),
        })

        const cssRender = async (sourceFullPath: string, fileContent: string): Promise<string> => {
            const sourceExtension = path.extname(sourceFullPath)
            const sourceBaseName = path.basename(sourceFullPath, sourceExtension)
            const sourceDirectory = path.dirname(sourceFullPath)
            const sourceRelativeDirectory = path.relative(rootPath, sourceDirectory)
            const isModule = sourceBaseName.endsWith('.module')
            const temporaryDirectory = path.resolve(temporaryDirectoryPath, sourceRelativeDirectory)
            await fs.promises.mkdir(temporaryDirectory, { recursive: true })

            const temporaryFilePath = path.join(temporaryDirectory, `${sourceBaseName}.css`)

            let css: string
            switch (sourceExtension) {
                case '.css':
                    css = fileContent
                    break

                case '.scss':
                    // renderSync is ~20% faster than render (because it's blocked on CPU, not IO).
                    css = sass
                        .renderSync({
                            file: sourceFullPath,
                            data: fileContent,
                            importer: (url, directory) => ({ file: resolveFile(url, path.dirname(directory)) }),
                        })
                        .css.toString()

                    break

                default:
                    throw new Error(`unknown file extension: ${sourceExtension}`)
            }

            const result = await postcss({
                ...postcssConfig,
                plugins: isModule ? [...postcssConfig.plugins, modulesPlugin] : postcssConfig.plugins,
            }).process(css, {
                from: sourceFullPath,
                to: temporaryFilePath,
            })

            await fs.promises.writeFile(temporaryFilePath, result.css)
            return temporaryFilePath
        }
        const cssRenderCache = new Map<string, { path: string; originalContent: string; outPath: string }>()
        const cachedCSSRender = async (sourceFullPath: string, fileContent: string): Promise<string> => {
            // TODO(sqs): invalidate
            const key = sourceFullPath
            const existing = cssRenderCache.get(key)
            if (existing && existing.originalContent === fileContent) {
                return existing.outPath
            }

            const outPath = await cssRender(sourceFullPath, fileContent)
            cssRenderCache.set(key, { path: sourceFullPath, originalContent: fileContent, outPath })
            return outPath
        }

        build.onResolve({ filter: /\.scss$/, namespace: 'file' }, async args => {
            const fullPath = resolveFile(args.path, args.resolveDir)
            const fileContent = await fs.promises.readFile(fullPath, 'utf8')
            const temporaryFilePath = await cachedCSSRender(fullPath, fileContent)
            const contents = await fs.promises.readFile(temporaryFilePath, 'utf8')

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
      import ${JSON.stringify(args.path.replace(/\.js$/, ''))}
      export default ${JSON.stringify(modulesMap.get(args.path.replace(/\.css$/, '.scss')) || {})}`,
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
