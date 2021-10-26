import fs from 'fs'

import { MANIFEST_PATH } from '../utils/constants'

export interface WebpackManifest {
    /** Main app entry JS bundle */
    'app.js': string
    /** Main app entry CSS bundle, only used in production mode */
    'app.css'?: string
    /** Runtime bundle, only used in development mode */
    'runtime.js'?: string
    /** React entry bundle, only used in production mode */
    'react.js'?: string
    /** TODO */
    isModule?: boolean
}

export const getWebpackManifest = (): WebpackManifest =>
    // eslint-disable-next-line no-sync
    JSON.parse(fs.readFileSync(MANIFEST_PATH, 'utf8')) as WebpackManifest
