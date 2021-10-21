import classnames from 'classnames'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useContext, useMemo } from 'react'
import { Link } from 'react-router-dom'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { AuthenticatedUser } from '../../../../../auth'
import { HeroPage } from '../../../../../components/HeroPage'
import { Page } from '../../../../../components/Page'
import { PageTitle } from '../../../../../components/PageTitle'
import { CodeInsightsBackendContext } from '../../../core/backend/code-insights-backend-context'
import { isLangStatsInsight, isSearchBasedInsight } from '../../../core/types'

import { EditLangStatsInsight } from './components/EditLangStatsInsight'
import { EditSearchBasedInsight } from './components/EditSearchInsight'
import styles from './EditInsightPage.module.scss'
import { useEditPageHandlers } from './hooks/use-edit-page-handlers'

export interface EditInsightPageProps {
    /** Normalized insight id <type insight>.insight.<name of insight> */
    insightID: string

    /**
     * Authenticated user info, Used to decide where code insight will appears
     * in personal dashboard (private) or in organisation dashboard (public)
     */
    authenticatedUser: Pick<AuthenticatedUser, 'id' | 'organizations' | 'username'>
}

export const EditInsightPage: React.FunctionComponent<EditInsightPageProps> = props => {
    const { insightID, authenticatedUser } = props

    const { getInsightSubjects, getInsightById } = useContext(CodeInsightsBackendContext)

    const subjects = useObservable(useMemo(() => getInsightSubjects(), [getInsightSubjects]))
    const insight = useObservable(useMemo(() => getInsightById(insightID), [getInsightById, insightID]))

    const { handleSubmit, handleCancel } = useEditPageHandlers({ originalInsight: insight })

    if (insight === undefined || subjects === undefined) {
        return <LoadingSpinner />
    }

    if (!insight) {
        return (
            <HeroPage
                icon={MapSearchIcon}
                title="Oops, we couldn't find that insight"
                subtitle={
                    <span>
                        We couldn't find that insight. Try to find the insight with ID:{' '}
                        <code className="badge badge-secondary">{insightID}</code> in your{' '}
                        <Link to={`/users/${authenticatedUser?.username}/settings`}>user or org settings</Link>
                    </span>
                }
            />
        )
    }

    return (
        <Page className={classnames('col-10', styles.creationPage)}>
            <PageTitle title="Edit code insight" />

            <div className="mb-5">
                <h2>Edit insight</h2>

                <p className="text-muted">
                    Insights analyze your code based on any search query.{' '}
                    <a href="https://docs.sourcegraph.com/code_insights" target="_blank" rel="noopener">
                        Learn more.
                    </a>
                </p>
            </div>

            {isSearchBasedInsight(insight) && (
                <EditSearchBasedInsight
                    insight={insight}
                    subjects={subjects}
                    onSubmit={handleSubmit}
                    onCancel={handleCancel}
                />
            )}

            {isLangStatsInsight(insight) && (
                <EditLangStatsInsight
                    insight={insight}
                    subjects={subjects}
                    onSubmit={handleSubmit}
                    onCancel={handleCancel}
                />
            )}
        </Page>
    )
}
