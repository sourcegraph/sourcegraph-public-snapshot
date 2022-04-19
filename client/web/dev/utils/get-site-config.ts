import fs from 'fs'

import { parse } from '@sqs/jsonc-parser'
import lodash from 'lodash'

import { SourcegraphContext } from '../../src/jscontext'

import { ENVIRONMENT_CONFIG } from './environment-config'

const { SITE_CONFIG_PATH } = ENVIRONMENT_CONFIG

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
