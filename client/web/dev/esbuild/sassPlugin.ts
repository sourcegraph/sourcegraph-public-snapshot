import fs from 'fs'
import os from 'os'
import path from 'path'

import esbuild from 'esbuild'
import postcss from 'postcss'
import postcssModules from 'postcss-modules'
import sass from 'sass'

import postcssConfig from '../../../../postcss.config'

const resolveFile = (modulePath: string, directory: string): string => {
    if (modulePath.startsWith('.')) {
        return path.resolve(directory, modulePath)
    }

    if (modulePath.startsWith('wildcard/') || modulePath.startsWith('shared/')) {
        return path.resolve(`client/${modulePath}`)
    }

    let p = path.resolve(`node_modules/${modulePath}`)
    try {
        p = fs.realpathSync(p)
    } catch {}
    return p
}
const resolveCache = new Map()
const cachedResolveFile = (modulePath: string, directory: string) => {
    const key = `${modulePath}:${directory}`
    const existing = resolveCache.get(key)
    if (existing) {
        return existing
    }

    const resolvedPath = resolveFile(modulePath, directory)
    resolveCache.set(key, resolvedPath)
    return resolvedPath
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
            generateScopedName: '[name]__[local]___[hash:base64:5]',
            localsConvention: 'camelCase',
            modules: true,
            getJSON: (cssPath: string, json: any) => modulesMap.set(cssPath, json),
        })

        const CWD = process.cwd()
        const cssRender = async (sourceFullPath: string, fileContent: string) => {
            const sourceExtension = path.extname(sourceFullPath)
            const sourceBaseName = path.basename(sourceFullPath, sourceExtension)
            const sourceDirectory = path.dirname(sourceFullPath)
            const sourceRelDir = path.relative(CWD, sourceDirectory)
            const isModule = sourceBaseName.endsWith('.module')
            const temporaryDirectory = path.resolve(temporaryDirectoryPath, sourceRelDir)
            await fs.promises.mkdir(temporaryDirectory, { recursive: true })

            const temporaryFilePath = path.join(temporaryDirectory, `${sourceBaseName}.css`)

            let css: string
            switch (sourceExtension) {
                case '.css':
                    css = fileContent
                    break

                case '.scss':
                    css = sass
                        .renderSync({
                            file: sourceFullPath,
                            data: fileContent,
                            importer: url => ({ file: cachedResolveFile(url) }),
                            quiet: true,
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
            if (false) {
                console.log(
                    'CACHE',
                    existing ? (existing.originalContent === fileContent ? 'HIT' : 'STALE') : 'MISS',
                    sourceFullPath
                )
            }
            if (existing && existing.originalContent === fileContent) {
                if (sourceFullPath.includes('UsagePage')) {
                    if (false) {
                        console.log('CACHE HIT', sourceFullPath, fileContent)
                    }
                }
                return existing.outPath
            }

            const outPath = await cssRender(sourceFullPath, fileContent)
            cssRenderCache.set(key, { path: sourceFullPath, originalContent: fileContent, outPath })
            return outPath
        }

        build.onResolve({ filter: /\.s?css$/, namespace: 'file' }, async args => {
            // Namespace is empty when using CSS as an entrypoint
            if (args.namespace !== 'file' && args.namespace !== '') {
                return
            }

            const sourceFullPath = cachedResolveFile(args.path, args.resolveDir)
            const fileContent = await fs.promises.readFile(sourceFullPath, 'utf8')
            const temporaryFilePath = await cachedCSSRender(sourceFullPath, fileContent)

            const isModule = sourceFullPath.endsWith('.module.css') || sourceFullPath.endsWith('.module.scss')

            return {
                namespace: isModule ? 'postcss-module' : 'file',
                path: temporaryFilePath,
                watchFiles: [sourceFullPath],
                pluginData: {
                    originalPath: sourceFullPath,
                },
            }
        })

        build.onResolve({ filter: /\.ttf$/, namespace: 'file' }, args => {
            // TODO(sqs): hack, need to resolve this from the original path
            if (args.path === './codicon.ttf') {
                return {
                    path: path.resolve('node_modules/monaco-editor/esm/vs/base/browser/ui/codicons/codicon', args.path),
                }
            }
        })
        build.onResolve({ filter: /\.png$/, namespace: 'file' }, args => {
            // TODO(sqs): hack, need to resolve this from the original path
            if (args.path === 'img/bg-sprinkles-2x.png') {
                return {
                    path: path.resolve('ui/assets', args.path),
                }
            }
        })

        build.onLoad({ filter: /./, namespace: 'postcss-module' }, async args => {
            const module_ = modulesMap.get(args.pluginData.originalPath)
            const resolveDirectory = path.dirname(args.path)

            const contents = `import ${JSON.stringify(args.path)}
            export default ${JSON.stringify(module_ || {})}`

            return {
                resolveDir: resolveDirectory,
                contents,
            }
        })

        // Handle the `import`ed CSS files from the previous onLoad filter.
        build.onResolve({ filter: /./, namespace: 'postcss-module' }, args => ({
            path: args.path,
            namespace: 'file',
        }))
    },
}
