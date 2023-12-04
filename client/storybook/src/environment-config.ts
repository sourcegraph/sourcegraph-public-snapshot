import { getEnvironmentBoolean } from '@sourcegraph/build-config'

export const ENVIRONMENT_CONFIG = {
    STORIES_GLOB: process.env.STORIES_GLOB,
    CHROMATIC: getEnvironmentBoolean('CHROMATIC'),
}
