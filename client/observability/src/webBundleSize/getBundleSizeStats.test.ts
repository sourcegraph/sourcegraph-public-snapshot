import path from 'path'

import { getBundleSizeStats } from './getBundleSizeStats'

const MOCK_ASSETS_PATH = path.join(__dirname, './__mocks__/assets')

jest.mock(
    'bundlesize.config.js',
    () => ({
        files: [
            {
                path: MOCK_ASSETS_PATH + '/scripts/*.br',
                maxSize: '10kb',
            },
            {
                path: MOCK_ASSETS_PATH + '/styles/*.br',
                maxSize: '10kb',
            },
        ],
    }),
    { virtual: true }
)

jest.mock(
    'webpack.manifest.json',
    () => ({
        'app.js': '/.assets/scripts/app.bundle.js',
        'app.css': '/.assets/styles/app.123.bundle.css',
    }),
    { virtual: true }
)

describe('getBundleSizeStats', () => {
    it('collects bundle stats based on bundlesize config', () => {
        const stats = getBundleSizeStats({
            staticAssetsPath: MOCK_ASSETS_PATH,
            bundlesizeConfigPath: 'bundlesize.config.js',
            webpackManifestPath: 'webpack.manifest.json',
        })

        expect(stats).toEqual({
            'scripts/app.bundle.js': {
                raw: 15,
                gzip: 6,
                brotli: 4,
                isInitial: true,
                isDynamicImport: false,
                isDefaultVendors: false,
                isCss: false,
                isJs: true,
            },
            'scripts/sg_home.js': {
                brotli: 4,
                gzip: 6,
                isCss: false,
                isDefaultVendors: false,
                isDynamicImport: true,
                isInitial: false,
                isJs: true,
                raw: 15,
            },
            'styles/app.123.bundle.css': {
                raw: 25,
                gzip: 6,
                brotli: 4,
                isInitial: true,
                isDynamicImport: false,
                isDefaultVendors: false,
                isCss: true,
                isJs: false,
            },
        })
    })
})
