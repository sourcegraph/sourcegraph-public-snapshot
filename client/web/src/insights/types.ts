import React from 'react'

import { ExtensionViewsSectionProps } from './sections/ExtensionViewsSection'

/**
 * Common props for code insights consumer components (the homepage, the directory page)
 * These props are needed for condition rendering of Code Insights and extension-like views.
 * For the enterprise version it's code insights + extension views and for the OSS version
 * it's extension views only.
 */
export interface CodeInsightsProps {
    codeInsightsEnabled?: boolean
    extensionViews: React.FunctionComponent<ExtensionViewsSectionProps>
}

/**
 * Props that are needed to tune code insights internal logic. Like switching different
 * code insights api backend (setting-based vs gql based api)
 */
export interface CodeInsightsContextProps {
    isCodeInsightsGqlApiEnabled: boolean
}
