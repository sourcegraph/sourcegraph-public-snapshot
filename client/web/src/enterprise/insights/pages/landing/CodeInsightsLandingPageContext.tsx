import { createContext, useContext } from 'react'

export enum CodeInsightsLandingPageType {
    InProduct,
    Cloud,
}

interface CodeInsightsLandingPageContextData {
    /**
     * Most of the components for the in-product landing page are universal and could be
     * (and actually are used) for the cloud versions of landing page as well. To be able
     * to separate context of usage of those component (for example for pings) we should
     * provide context with particular mode type
     */
    mode: CodeInsightsLandingPageType
}

export const CodeInsightsLandingPageContext = createContext<CodeInsightsLandingPageContextData>({
    mode: CodeInsightsLandingPageType.InProduct,
})

/**
 * Hook helper for the ping event's name that is generated based on
 * the landing page mode.
 */
export function useLogEventName(originalName: string): string {
    const { mode } = useContext(CodeInsightsLandingPageContext)

    return mode === CodeInsightsLandingPageType.Cloud ? `Cloud${originalName}` : originalName
}
