import { IPlugin } from '../api/types'

import { airQualityPlugin } from './air-quality'
import { confluencePlugin } from './confluence'
import { githubIssuesPlugin } from './github-issues'
import { timezonePlugin } from './timezone'
import { urlReaderPlugin } from './url-reader'
import { weatherPlugin } from './weather'

export const defaultPlugins: IPlugin[] = [
    // are included by default
    confluencePlugin,
    githubIssuesPlugin,
    urlReaderPlugin,
    // not included by default
    weatherPlugin,
    timezonePlugin,
    airQualityPlugin,
]
