import { IPlugin } from '../api/types'

import { confluencePlugin } from './confluence'
import { weatherPlugin } from './waether'

export const defaultPlugins: IPlugin[] = [weatherPlugin, confluencePlugin]
