import path from 'path'

import { mapKeys } from 'lodash'

import { getBundleSizeStats } from './getBundleSizeStats'

const MOCKS_PATH = path.join(__dirname, './__mocks__')
const BUNDLESIZE_CONFIG_PATH = path.join(MOCKS_PATH, 'bundlesize.config.js')

describe('getBundleSizeStats', () => {
    it('collects bundle stats based on bundlesize config', () => {
        const stats = getBundleSizeStats(BUNDLESIZE_CONFIG_PATH)
        const prettyStats = mapKeys(stats, (_value, key) => key.replace(`${MOCKS_PATH}/`, ''))

        expect(prettyStats).toEqual({
            'assets/scripts/app.bundle.js': { raw: 15, gzip: 6, brotli: 4 },
            'assets/styles/app.123.bundle.css': { raw: 25, gzip: 6, brotli: 4 },
        })
    })
})
