import fs from 'fs'
import path from 'path'

import esbuild from 'esbuild'
import postcss from 'postcss'
import postcssModules from 'postcss-modules'
import sass from 'sass'

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore
import postcssConfig from '../../../../postcss.config'
import { ROOT_PATH } from '../utils'

/**
 * An esbuild plugin that builds .css and .scss stylesheets (including support for CSS modules).
 */
export const stylePlugin: esbuild.Plugin = {
    name: 'style',
    setup: build => {
        const modulesMap = new Map<string, string>()
        const modulesPlugin = postcssModules({
            generateScopedName: '[name]__[local]', // omit hash for local dev
            localsConvention: 'camelCase',
            getJSON: (cssPath, json) => modulesMap.set(cssPath, JSON.stringify(json)),
        })

        interface TransformArguments {
            inputPath: string
            inputContents: string
        }
        interface TransformResult {
            outputPath: string
            outputContents: string
        }
        const transform = async ({ inputPath, inputContents }: TransformArguments): Promise<TransformResult> => {
            const isSCSS = inputPath.endsWith('.scss')
            const css = isSCSS
                ? // renderSync is ~20% faster than render with an async callback (because it's blocked on CPU, not IO).
                  // eslint-disable-next-line no-sync
                  sass
                      .renderSync({
                          file: inputPath,
                          data: inputContents,
                          includePaths: [path.resolve(ROOT_PATH, 'node_modules'), path.resolve(ROOT_PATH, 'client')],
                      })
                      .css.toString()
                : inputContents

            const outputPath = isSCSS ? inputPath.replace(/\.scss$/, '.css') : inputPath

            const isCSSModule = outputPath.endsWith('.module.css')
            const result = await postcss({
                ...postcssConfig,
                // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
                plugins: isCSSModule ? [...postcssConfig.plugins, modulesPlugin] : postcssConfig.plugins,
            }).process(css, {
                from: outputPath,
            })
            return {
                outputPath,
                outputContents: result.css,
            }
        }
        const transformCache = new Map<
            TransformArguments['inputPath'],
            { inputContents: string; outputPath: string; outputContents: string }
        >()
        const cachedTransform = async ({ inputPath, inputContents }: TransformArguments): Promise<TransformResult> => {
            const cached = transformCache.get(inputPath)
            if (cached && cached.inputContents === inputContents) {
                return cached
            }

            const output = await transform({ inputPath, inputContents })
            transformCache.set(inputPath, { inputContents, ...output })
            return output
        }

        build.onResolve({ filter: /\.s?css$/, namespace: 'file' }, async args => {
            const inputPath = path.join(args.resolveDir, args.path)
            const { outputPath, outputContents } = await cachedTransform({
                inputPath,
                inputContents: await fs.promises.readFile(inputPath, 'utf8'),
            })
            const isCSSModule = outputPath.endsWith('.module.css')

            return {
                path: outputPath,
                namespace: isCSSModule ? 'css-module' : 'css',
                pluginData: { contents: outputContents },
            }
        })

        // Resolve CSS modules imported by the next onLoad callback to the actual stylesheet (not
        // the synthesized JavaScript module that exports the CSS module's class names).
        build.onResolve({ filter: /\.css$/, namespace: 'css-module' }, args => ({
            path: args.path,
            namespace: 'css',
            // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-member-access
            pluginData: { contents: args.pluginData?.contents },
        }))

        // Load a synthesized JavaScript module that exports the CSS module's class names and
        // imports (for side effects) the actual CSS file.
        build.onLoad({ filter: /\.module\.css$/, namespace: 'css-module' }, args => ({
            contents: `
import ${JSON.stringify(args.path)}
export default ${modulesMap.get(args.path) || '{}'}`,
            loader: 'js',
            resolveDir: path.dirname(args.path),
            // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
            pluginData: args.pluginData,
        }))

        // Load the contents of all CSS files. The transformed CSS was passed through `pluginData.contents`.
        build.onLoad({ filter: /\.css$/, namespace: 'css' }, args => ({
            // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-member-access
            contents: args.pluginData?.contents,
            resolveDir: path.dirname(args.path),
            loader: 'css',
        }))
    },
}
