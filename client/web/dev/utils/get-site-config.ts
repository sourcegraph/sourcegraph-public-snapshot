import fs from 'fs'
import path from 'path'

import { camelCase, mapKeys } from 'lodash'
import stripJsonComments from 'strip-json-comments'

import { SourcegraphContext } from '../../src/jscontext'

import { ROOT_PATH } from './constants'

const SITE_CONFIG_PATH = path.resolve(ROOT_PATH, '../dev-private/enterprise/dev/site-config.json')

export const getSiteConfig = (): Partial<SourcegraphContext> => {
    try {
        // eslint-disable-next-line no-sync
        const siteConfig = JSON.parse(stripJsonComments(fs.readFileSync(SITE_CONFIG_PATH, 'utf-8')))

        return mapKeys(siteConfig, (_value, key) => camelCase(key))
    } catch (error) {
        console.log('Site config not found!', SITE_CONFIG_PATH)
        console.error(error)

        return {}
    }
}
