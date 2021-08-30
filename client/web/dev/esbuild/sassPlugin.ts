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
                    css = sass
                        .renderSync({
                            file: sourceFullPath,
                            data: fileContent,
                            importer: url => ({ file: cachedResolveFile(url) }),
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

        build.onResolve({ filter: /\.s?css$/, namespace: 'file' }, async args => {
            // Namespace is empty when using CSS as an entrypoint
            if (args.namespace !== 'file' && args.namespace !== '') {
                return
            }
            if (args.path.startsWith('/tmp')) {
                return // exclude from monacoPlugin
            }

            const sourceFullPath = cachedResolveFile(args.path, args.resolveDir)
            const fileContent = await fs.promises.readFile(sourceFullPath, 'utf8')
            const temporaryFilePath = await cachedCSSRender(sourceFullPath, fileContent)

            const isModule = sourceFullPath.endsWith('.module.css') || sourceFullPath.endsWith('.module.scss')

            return {
                namespace: isModule ? 'css-module' : 'file',
                path: isModule ? temporaryFilePath + '.js' : temporaryFilePath,
                watchFiles: [sourceFullPath],
                pluginData: {
                    originalPath: sourceFullPath,
                },
            }
        })

        /*         // Load CSS files resolved by the previous onResolve filter with the original resolveDir so
        // that url(...) references are resolved correctly.
        build.onLoad({ filter: /./, namespace: 'css-file' }, async args => {
            const isRelevant = args.path.includes('enterprise.css')
            if (isRelevant) {
                console.log('ON RESOLVE 1111', args)
            }
            return {
                contents: await fs.promises.readFile(args.path, 'utf-8'),
                resolveDir: path.dirname(args.pluginData.originalPath),
                loader: 'css',
            }
        })
 */
        build.onLoad({ filter: /\.js$/, namespace: 'css-module' }, args => {
            const module_ = modulesMap.get(args.pluginData.originalPath)

            const contents = `
                import ${JSON.stringify(args.path)}
                ${args.path.includes('NavBar.module') ? 'console.log("NavBar CSS")' : ''}
                export default ${JSON.stringify(module_ || {})}`

            return {
                resolveDir: path.dirname(args.pluginData.originalPath),
                contents,
                loader: 'js',
            }
        })

        // Handle the `import`ed CSS files from the previous onLoad filter.
        build.onResolve({ filter: /./, namespace: 'css-module' }, args => {
            if (false) {
                console.log('ON RESOLVE 2222')
            }
            return {
                path: args.path,
                namespace: 'file',
            }
        })
    },
}
