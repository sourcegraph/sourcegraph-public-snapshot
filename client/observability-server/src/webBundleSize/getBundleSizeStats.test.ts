import path from 'path'

import { describe, expect, it, jest } from '@jest/globals'

import { getBundleSizeStats } from './getBundleSizeStats'

const MOCK_ASSETS_PATH = path.join(__dirname, './__mocks__/assets')

jest.mock(
    'bundlesize.config.js',
    () => ({
        files: [
            {
                path: MOCK_ASSETS_PATH + '/scripts/*.js',
                maxSize: '10kb',
            },
            {
                path: MOCK_ASSETS_PATH + '/styles/*.css',
                maxSize: '10kb',
            },
        ],
    }),
    { virtual: true }
)

jest.mock(
    'web.manifest.json',
    () => ({
        'main.js': '/.assets/scripts/main.js',
        'main.css': '/.assets/styles/main.css',
    }),
    { virtual: true }
)

describe('getBundleSizeStats', () => {
    it('collects bundle stats based on bundlesize config', () => {
        const stats = getBundleSizeStats({
            staticAssetsPath: MOCK_ASSETS_PATH,
            bundlesizeConfigPath: 'bundlesize.config.js',
            webBuildManifestPath: 'web.manifest.json',
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
