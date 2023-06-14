import path from 'path'

import { cleanEnv, str } from 'envalid'

export const ENVIRONMENT_CONFIG = cleanEnv(process.env, {
    SOURCEGRAPH_ACCESS_TOKEN: str(),
    OUTPUT_PATH: str({ default: path.join(__dirname, '../../data') }),
})
