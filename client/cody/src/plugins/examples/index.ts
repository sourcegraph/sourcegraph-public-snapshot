import { IPlugin } from '../api/types'

import { confluencePlugin } from './confluence'
import { githubIssuesPlugin } from './github-issues'
import { weatherPlugin } from './weather'

export const defaultPlugins: IPlugin[] = [weatherPlugin, confluencePlugin, githubIssuesPlugin]
