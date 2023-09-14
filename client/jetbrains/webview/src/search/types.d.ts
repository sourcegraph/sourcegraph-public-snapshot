import type { SearchPatternType } from '../graphql-operations'

import type { ActionName } from './java-to-js-bridge'
import type { Request } from './js-to-java-bridge'

/* Add global functions to global window object */
declare global {
    interface Window {
        initializeSourcegraph: () => Promise<void>
        callJava: (request: Request) => Promise<object>
        callJS: (action: ActionName, data: string, callback: (result: string) => void) => Promise<void>
    }
}

export interface Theme {
    isDarkTheme: boolean
    intelliJTheme: { [key: string]: string }
}

export interface PluginConfig {
    instanceURL: string
    accessToken: string | null
    customRequestHeadersAsString: string | null
    pluginVersion: string
    anonymousUserId: string
}

export interface Search {
    query: string | null
    caseSensitive: boolean
    patternType: SearchPatternType
    selectedSearchContextSpec: string
}
