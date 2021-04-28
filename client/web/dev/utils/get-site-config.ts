import fs from 'fs'
import path from 'path'

import { parse } from '@sqs/jsonc-parser'
import lodash from 'lodash'

import { SourcegraphContext } from '../../src/jscontext'

import { ROOT_PATH } from './constants'

const SITE_CONFIG_PATH = path.resolve(ROOT_PATH, '../dev-private/enterprise/dev/site-config.json')

// Get site-config from `SITE_CONFIG_PATH` as an object with camel cased keys.
export const getSiteConfig = (): Partial<SourcegraphContext> => {
    try {
        // eslint-disable-next-line no-sync
        const siteConfig = parse(fs.readFileSync(SITE_CONFIG_PATH, 'utf-8'))

        return lodash.mapKeys(siteConfig, (_value, key) => lodash.camelCase(key))
    } catch (error) {
        console.log('Site config not found!', SITE_CONFIG_PATH)
        console.error(error)

        return {}
    }
}
