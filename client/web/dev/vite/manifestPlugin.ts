import { writeFileSync } from 'node:fs'
import path from 'node:path'
import os from 'os'

import { Plugin, normalizePath } from 'vite'

import { WebBuildManifest } from '../utils/webBuildManifest'

export interface ManifestPluginConfig {
    fileName: string
}

function createSimplifyPath(root: string, base: string): (path: string) => string {
    return path => {
        path = normalizePath(path)

        if (root !== '/' && path.startsWith(root)) {
            path = path.slice(root.length)
        }

        if (path.startsWith(base)) {
            path = path.slice(base.length)
        }

        if (path[0] === '/') {
            path = path.slice(1)
        }

        return path
    }
}

/**
 * A Vite plugin that writes a JSON manifest with the JavaScript and CSS output filenames for each
 * entrypoint. This JSON manifest is read by our backend to inject the right scripts and styles into
 * the page.
 */
export function manifestPlugin(pluginConfig: ManifestPluginConfig): Plugin {
    let root: string | undefined

    return {
        name: 'manifest',
        enforce: 'post',
        configResolved(config): void {
            root = config.root
        },

        // Run in dev mode.
        configureServer({ config, httpServer }): void {
            httpServer?.once('listening', () => {
                // Resolve URL.
                const { root: _root, base } = config
                const root = normalizePath(_root)
                const protocol = config.server.https ? 'https' : 'http'
                const host = resolveHost(config.server.host)
                const port = config.server.port
                const url = `${protocol}://${host}:${port}${base}`
                config.server.origin = `${protocol}://${host}:${port}`

                // Resolve inputs.
                const simplifyPath = createSimplifyPath(root, base)
                const inputOptions = config.build.rollupOptions?.input ?? {}
                const inputs =
                    typeof inputOptions === 'string'
                        ? { [inputOptions]: inputOptions }
                        : Array.isArray(inputOptions)
                        ? Object.fromEntries(inputOptions.map(path => [path, path]))
                        : inputOptions

                const manifest: WebBuildManifest = {
                    url: url,
                    assets: {},
                    devInjectHTML: `
                        <script type="module">
                            import RefreshRuntime from "${url}@react-refresh"
                            RefreshRuntime.injectIntoGlobalHook(window)
                            window.$RefreshReg$ = () => {}
                            window.$RefreshSig$ = () => (type) => type
                            window.__vite_plugin_react_preamble_installed__ = true
                        </script>
                        <script type="module" src="${url}@vite/client"></script>`,
                }
                for (const [entryAlias, entryPath] of Object.entries(inputs)) {
                    const relativeEntryAlias = normalizePath(path.relative(root, entryAlias))
                    manifest.assets[noExt(relativeEntryAlias)] = {
                        js: simplifyPath(entryPath),
                    }
                }

                const outputDir = path.resolve(config.root, config.build.outDir)
                writeFileSync(path.resolve(outputDir, pluginConfig.fileName), JSON.stringify(manifest, null, 2))
            })
        },

        // Run when generating the production bundle.
        generateBundle(_options, bundle): void {
            if (root === undefined) {
                throw new Error('no config')
            }

            const manifest: WebBuildManifest = { assets: {} }
            for (const chunk of Object.values(bundle)) {
                if (chunk.type === 'chunk' && chunk.isEntry && chunk.facadeModuleId) {
                    let entryAlias = normalizePath(path.relative(root, chunk.facadeModuleId))
                    const css = chunk.viteMetadata ? Array.from(chunk.viteMetadata?.importedCss.values()) : []
                    if (css.length >= 2) {
                        throw new Error('multiple CSS asset files not supported')
                    }
                    manifest.assets[noExt(entryAlias)] = {
                        js: chunk.fileName,
                        css: css.length === 1 ? css[0] : undefined,
                    }
                }
            }
            this.emitFile({ fileName: pluginConfig.fileName, type: 'asset', source: JSON.stringify(manifest, null, 2) })
        },
    }
}

function noExt(path: string): string {
    return path.replace(/\.(tsx|js)$/, '')
}

/**
 * Resolve host if is passed as `true`
 *
 * Copied from https://github.com/vitejs/vite/blob/d4dcdd1ffaea79ecf8a9fc78cdbe311f0d801fb5/packages/vite/src/node/logger.ts#L197
 */
function resolveHost(host?: string | boolean): string {
    if (!host) return 'localhost'

    if (host === true) {
        const nInterface = Object.values(os.networkInterfaces())
            .flatMap(nInterface => nInterface ?? [])
            .filter(
                detail =>
                    detail &&
                    detail.address &&
                    // Node < v18
                    ((typeof detail.family === 'string' && detail.family === 'IPv4') ||
                        // Node >= v18
                        (typeof detail.family === 'number' && (detail as any).family === 4))
            )
            .filter(detail => {
                return detail.address !== '127.0.0.1'
            })[0]

        if (!nInterface) return 'localhost'

        return nInterface.address
    }

    return host
}
