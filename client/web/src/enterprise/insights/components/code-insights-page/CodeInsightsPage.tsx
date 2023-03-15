import React from 'react'

import { Page } from '../../../../components/Page'
import { useUiFeatures } from '../../hooks'

import { CodeInsightsLimitedAccessAppBanner } from './limit-access-banner/CodeInsightsLimitAccessAppBanner'
import { CodeInsightsLimitAccessBanner } from './limit-access-banner/CodeInsightsLimitAccessBanner'

interface CodeInsightsPageProps extends React.HTMLAttributes<HTMLDivElement> {
    isSourcegraphApp: boolean
}

/**
 * Shared common component for creation a typical code insights pages. Contains common styles
 * and demo mode banner in order to render it across all pages.
 */
export const CodeInsightsPage: React.FunctionComponent<React.PropsWithChildren<CodeInsightsPageProps>> = props => {
    const { licensed } = useUiFeatures()

    return (
        <Page {...props}>
            {props.isSourcegraphApp && <CodeInsightsLimitedAccessAppBanner authenticatedUser={null} />}
            {!licensed && !props.isSourcegraphApp && <CodeInsightsLimitAccessBanner className="mb-4" />}
            {props.children}
        </Page>
    )
}
