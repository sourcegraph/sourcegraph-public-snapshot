import { MdiReactIconComponentType } from 'mdi-react'
import DatabaseIcon from 'mdi-react/DatabaseIcon'
import PuzzleIcon from 'mdi-react/PuzzleIcon'
import React, { useState } from 'react'
import { useLocation } from 'react-router-dom'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ErrorAlert } from '../../../../../components/alerts'
import { ErrorBoundary } from '../../../../../components/ErrorBoundary'
import { ViewInsightProviderResult, ViewInsightProviderSourceType } from '../../../../core/backend/types'
import { InsightTypePrefix } from '../../../../core/types'
import { InsightViewContent } from '../../../insight-view-content/InsightViewContent'

import { InsightDescription } from './components/insight-card-description/InsightCardDescription'
import { InsightCardMenu } from './components/insight-card-menu/InsightCardMenu'
import styles from './InsightCard.module.scss'

export interface InsightCardProps extends TelemetryProps {
    /** Insight data (title, chart content) */
    insight: ViewInsightProviderResult

    /** Deleting handler fires when the user clicks delete in the insight menu. */
    onDelete: (id: string) => Promise<void>

    /**
     * Prop for enabling and disabling insight context menu.
     * Now only insight page has insights with context menu.
     * */
    hasContextMenu?: boolean

    /** To get container to track hovers for pings */
    containerClassName?: string
}

/**
 * Renders insight card content. Loading state, error state and insight itself.
 */
export const InsightContentCard: React.FunctionComponent<InsightCardProps> = props => {
    const {
        insight: { id, view, source },
        containerClassName,
        hasContextMenu,
        onDelete,
        telemetryService,
    } = props

    const location = useLocation()

    // We should disable delete and any other actions if we already have started
    // operation over some insight.
    const [isDeleting, setDeletingState] = useState(false)

    // We support actions only over search and lang insights and not able to edit or delete
    // custom insight or backend insight.
    const hasMenu =
        hasContextMenu && (id.startsWith(InsightTypePrefix.search) || id.startsWith(InsightTypePrefix.langStats))

    const handleDelete = async (): Promise<void> => {
        setDeletingState(true)

        await onDelete(id)

        setDeletingState(false)
    }

    return (
        <ErrorBoundary
            location={location}
            extraContext={
                <>
                    <p>ID: {id}</p>
                    <pre>View: {JSON.stringify(view, null, 2)}</pre>
                </>
            }
            className="pt-0"
        >
            {view === undefined || isDeleting ? (
                <>
                    <div className="flex-grow-1 d-flex flex-column align-items-center justify-content-center">
                        <LoadingSpinner /> {isDeleting ? 'Deleting code insight' : 'Loading code insight'}
                    </div>
                    <InsightDescription
                        className={styles.insightCardDescription}
                        title={id}
                        icon={getInsightViewIcon(source)}
                    />
                </>
            ) : isErrorLike(view) ? (
                <>
                    <ErrorAlert data-testid={`${id} insight error`} className="m-0" error={view} />
                    <InsightDescription
                        className={styles.insightCardDescription}
                        title={id}
                        icon={getInsightViewIcon(source)}
                    />
                </>
            ) : (
                <>
                    <header className={styles.insightCardHeader}>
                        <div className={styles.insightCardHeaderContent}>
                            <h4 className={styles.insightCardTitle}>{view.title}</h4>
                            {view.subtitle && <div className={styles.insightCardSubtitle}>{view.subtitle}</div>}
                        </div>

                        {hasMenu && (
                            <InsightCardMenu className="mr-n2 d-inline-flex" insightID={id} onDelete={handleDelete} />
                        )}
                    </header>

                    <InsightViewContent
                        telemetryService={telemetryService}
                        viewContent={view.content}
                        viewID={id}
                        containerClassName={containerClassName}
                    />
                </>
            )}
        </ErrorBoundary>
    )
}

const getInsightViewIcon = (source: ViewInsightProviderSourceType): MdiReactIconComponentType => {
    switch (source) {
        case ViewInsightProviderSourceType.Backend:
            return DatabaseIcon
        case ViewInsightProviderSourceType.Extension:
            return PuzzleIcon
    }
}
