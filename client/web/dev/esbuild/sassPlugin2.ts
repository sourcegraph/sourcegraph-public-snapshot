import fs from 'fs'
import path from 'path'

import * as esbuild from 'esbuild'
import { sassPlugin as createSassPlugin, postcssModules } from 'esbuild-sass-plugin'
import { createSassImporter } from 'esbuild-sass-plugin/lib/importer'

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
const cachedResolveFile = (modulePath: string, directory: string): string => {
    const key = `${modulePath}:${directory}`
    const existing = resolveCache.get(key)
    if (existing) {
        return existing
    }

    const resolvedPath = resolveFile(modulePath, directory)
    resolveCache.set(key, resolvedPath)
    return resolvedPath
}

const sassImporter = createSassImporter({ basedir: path.join(rootPath, 'client') })
const importer2 = url => {
    try {
        return sassImporter(url, null)
    } catch {
        return sassImporter('./' + url, null)
    }
}

const pm = postcssModules({ localsConvention: 'camelCase' })

export const sassPlugin: esbuild.Plugin = createSassPlugin({
    cache: true,
    transform: (css, resolveDirectory, filePath) => {
        const isModule = filePath.endsWith('.module.css') || filePath.endsWith('.module.scss')
        return isModule ? pm(css, resolveDirectory, filePath) : css
    },
    includePaths: [path.join(rootPath, 'client'), path.join(rootPath, 'node_modules')],
    basedir: path.join(rootPath, 'client'),
    importer: (url, previous) => ({ file: cachedResolveFile(url, previous) }),
})
