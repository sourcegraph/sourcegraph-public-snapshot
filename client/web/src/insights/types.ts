/**
 * Common props for code insights consumer components (the homepage, the directory page)
 * These props are needed for condition rendering of Code Insights and extension-like views.
 * For the enterprise version it's code insights + extension views and for the OSS version
 * it's extension views only.
 */
export interface CodeInsightsProps {
    codeInsightsEnabled?: boolean
}
