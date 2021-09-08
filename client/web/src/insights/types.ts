import React from 'react'

import { ExtensionViewsSectionProps } from './sections/ExtenstionViewsSection'

/**
 * Common props for components (the homepage, the directory page) needing to render
 * Code Insights and extension views for the enterprise version and extension views
 * only for the OSS version.
 */
export interface CodeInsightsProps {
    codeInsightsEnabled?: boolean
    extensionViews: React.FunctionComponent<ExtensionViewsSectionProps>
}
