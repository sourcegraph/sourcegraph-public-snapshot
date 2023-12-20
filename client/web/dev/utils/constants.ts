import path from 'path'

import { ROOT_PATH } from '@sourcegraph/build-config'

export const DEFAULT_SITE_CONFIG_PATH = path.resolve(ROOT_PATH, '../dev-private/enterprise/dev/site-config.json')
export const assetPathPrefix = '/.assets'

export const WEB_BUILD_MANIFEST_FILENAME = 'vite-manifest.json'
