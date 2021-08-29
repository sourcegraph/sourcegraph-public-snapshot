import fs from 'fs'

import findCacheDir from 'find-cache-dir'

type EslintPluginCache = (entryName: string, key: string, func: () => Promise<string>) => Promise<string>

export const newEsbuildPluginCache = (pluginName: string): EslintPluginCache => {
    const cachePath = findCacheDir({ name: pluginName, thunk: true })
    if (!cachePath) {
        throw new Error('unable to find cache directory for esbuild plugin (unexpected behavior from find-cache-dir)')
    }

    /**
     * The entry is stale if the existing key file is missing or invalid, or if it doesn't match the
     * new key.
     */
    const isEntryStale = async (keyPath: string, newKey: string): Promise<boolean> => {
        try {
            const existingKey = await fs.promises.readFile(keyPath, 'utf-8')
            return existingKey !== newKey
        } catch (error) {
            if (error.code !== 'ENOENT') {
                throw error
            }
            return true // missing or invalid key existing key file
        }
    }

    return async (entryName, key, func) => {
        const keyPath = cachePath(`${entryName}.key.json`)
        const resultPath = cachePath(entryName)
        if (await isEntryStale(keyPath, key)) {
            const result = await func()
            await fs.promises.writeFile(resultPath, result)
        }
        return resultPath
    }
}
