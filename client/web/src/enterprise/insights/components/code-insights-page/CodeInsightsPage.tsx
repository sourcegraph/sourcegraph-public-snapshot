import React, { useContext, useMemo } from 'react'

import { Page } from '../../../../components/Page'
import { CodeInsightsBackendContext } from '../../core/backend/code-insights-backend-context'
import { CodeInsightsLimitAccessBanner } from '../limit-access-banner/CodeInsightsLimitAccessBanner'

interface CodeInsightsPageProps extends React.HTMLAttributes<HTMLDivElement> {}

/**
 * Shared common component for creation a typical code insights pages. Contains common styles
 * and demo mode banner in order to render it across all pages.
 */
export const CodeInsightsPage: React.FunctionComponent<CodeInsightsPageProps> = props => {
    const { getUiFeatures } = useContext(CodeInsightsBackendContext)
    const { licensed } = useMemo(() => getUiFeatures(), [getUiFeatures])

    return (
        <Page {...props}>
            {!licensed && <CodeInsightsLimitAccessBanner className="mb-3" />}
            {props.children}
        </Page>
    )
}
