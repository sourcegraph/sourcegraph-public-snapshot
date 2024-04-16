import path from 'path'

import { describe, expect, it } from 'vitest'

import { getBundleSizeStats } from './getBundleSizeStats'

const MOCK_ASSETS_PATH = path.join(__dirname, './__mocks__/assets')

describe('getBundleSizeStats', () => {
    it('collects bundle stats based on bundlesize config', () => {
        const stats = getBundleSizeStats({
            staticAssetsPath: MOCK_ASSETS_PATH,
            bundlesizeConfigPath: __dirname + '/__mocks__/bundlesize.config.js',
            webBuildManifestPath: __dirname + '/__mocks__/web.manifest.json',
        })

        expect(stats).toEqual({
            'scripts/main.js': {
                raw: 15,
                isInitial: true,
                isDynamicImport: false,
                isCss: false,
                isJs: true,
            },
            'styles/main.css': {
                raw: 25,
                isInitial: true,
                isDynamicImport: false,
                isCss: true,
                isJs: false,
            },
        })
    })
})
