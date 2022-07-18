import React, { useContext, useMemo } from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'

import { Badge, LoadingSpinner, useObservable, Link, PageHeader, Text } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../../auth'
import { HeroPage } from '../../../../../components/HeroPage'
import { PageTitle } from '../../../../../components/PageTitle'
import { CodeInsightsIcon } from '../../../../../insights/Icons'
import { CodeInsightsPage } from '../../../components'
import {
    CodeInsightsBackendContext,
    isCaptureGroupInsight,
    isComputeInsight,
    isLangStatsInsight,
    isSearchBasedInsight,
} from '../../../core'
import { useUiFeatures } from '../../../hooks'

import { EditCaptureGroupInsight } from './components/EditCaptureGroupInsight'
import { EditComputeInsight } from './components/EditComputeInsight'
import { EditLangStatsInsight } from './components/EditLangStatsInsight'
import { EditSearchBasedInsight } from './components/EditSearchInsight'
import { useEditPageHandlers } from './hooks/use-edit-page-handlers'

export interface EditInsightPageProps {
    /** Normalized insight id <type insight>.insight.<name of insight> */
    insightID: string

    /**
     * Authenticated user info, Used to decide where code insight will appear
     * in personal dashboard (private) or in organisation dashboard (public)
     */
    authenticatedUser: Pick<AuthenticatedUser, 'id' | 'organizations' | 'username'>
}

export const EditInsightPage: React.FunctionComponent<React.PropsWithChildren<EditInsightPageProps>> = props => {
    const { insightID, authenticatedUser } = props

    const { getInsightById } = useContext(CodeInsightsBackendContext)
    const { licensed, insight: insightFeatures } = useUiFeatures()

    const insight = useObservable(useMemo(() => getInsightById(insightID), [getInsightById, insightID]))
    const { handleSubmit, handleCancel } = useEditPageHandlers({ id: insight?.id })

    const editPermission = useObservable(
        useMemo(() => insightFeatures.getEditPermissions(insight), [insightFeatures, insight])
    )

    if (insight === undefined) {
        return <LoadingSpinner inline={false} />
    }

    if (!insight) {
        return (
            <HeroPage
                icon={MapSearchIcon}
                title="Oops, we couldn't find that insight"
                subtitle={
                    <span>
                        We couldn't find that insight. Try to find the insight with ID:{' '}
                        <Badge variant="secondary" as="code">
                            {insightID}
                        </Badge>{' '}
                        in your <Link to={`/users/${authenticatedUser?.username}/settings`}>user or org settings</Link>
                    </span>
                }
            />
        )
    }

    return (
        <CodeInsightsPage>
            <PageTitle title="Edit insight - Code Insights" />

            <PageHeader
                className="mb-3"
                path={[{ icon: CodeInsightsIcon }, { text: 'Edit insight' }]}
                description={
                    <Text className="text-muted">
                        Insights analyze your code based on any search query.{' '}
                        <Link to="/help/code_insights" target="_blank" rel="noopener">
                            Learn more.
                        </Link>
                    </Text>
                }
            />

            {isSearchBasedInsight(insight) && (
                <EditSearchBasedInsight
                    licensed={licensed}
                    isEditAvailable={editPermission?.available}
                    insight={insight}
                    onSubmit={handleSubmit}
                    onCancel={handleCancel}
                />
            )}

            {isCaptureGroupInsight(insight) && (
                <EditCaptureGroupInsight
                    licensed={licensed}
                    isEditAvailable={editPermission?.available}
                    insight={insight}
                    onSubmit={handleSubmit}
                    onCancel={handleCancel}
                />
            )}

            {isLangStatsInsight(insight) && (
                <EditLangStatsInsight
                    licensed={licensed}
                    isEditAvailable={editPermission?.available}
                    insight={insight}
                    onSubmit={handleSubmit}
                    onCancel={handleCancel}
                />
            )}

            {isComputeInsight(insight) && (
                <EditComputeInsight
                    licensed={licensed}
                    isEditAvailable={editPermission?.available}
                    insight={insight}
                    onSubmit={handleSubmit}
                    onCancel={handleCancel}
                />
            )}
        </CodeInsightsPage>
    )
}
