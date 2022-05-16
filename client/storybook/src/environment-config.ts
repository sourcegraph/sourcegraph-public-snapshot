import { getEnvironmentBoolean } from '@sourcegraph/build-config'

export const ENVIRONMENT_CONFIG = {
    STORIES_GLOB: process.env.STORIES_GLOB,
    WEBPACK_DLL_PLUGIN: getEnvironmentBoolean('WEBPACK_DLL_PLUGIN'),
    WEBPACK_PROGRESS_PLUGIN: getEnvironmentBoolean('WEBPACK_PROGRESS_PLUGIN'),
    WEBPACK_BUNDLE_ANALYZER: getEnvironmentBoolean('WEBPACK_BUNDLE_ANALYZER'),
    WEBPACK_SPEED_ANALYZER: getEnvironmentBoolean('WEBPACK_SPEED_ANALYZER'),
    MINIFY: getEnvironmentBoolean('MINIFY'),
}
