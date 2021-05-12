import { MdiReactIconComponentType } from 'mdi-react'
import DatabaseIcon from 'mdi-react/DatabaseIcon'
import PuzzleIcon from 'mdi-react/PuzzleIcon'
import React, { useMemo } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ErrorAlert } from '../../../../../components/alerts'
import { ErrorBoundary } from '../../../../../components/ErrorBoundary'
import { ViewContent, ViewContentProps } from '../../../../../views/ViewContent'
import { ViewInsightProviderResult, ViewInsightProviderSourceType } from '../../../../core/backend/types'
import { InsightTypeSuffix } from '../../../../core/types'

import { InsightDescription } from './components/description /InsightCardDescription'
import { InsightCardMenu } from './components/menu/InsightCardMenu'
import styles from './InsightCard.module.scss'

export interface InsightCardProps extends Omit<ViewContentProps, 'viewContent' | 'viewID'>, TelemetryProps {
    /** Insight data (title, chart content) */
    insight: ViewInsightProviderResult

    /** Setting for showing deleting state. */
    deleting: boolean

    /** Deleting handler fires when the user clicks delete in the insight menu. */
    onDelete: (id: string) => void
}

/**
 * Renders insight card content. Loading state, error state and insight itself.
 */
export const InsightContentCard: React.FunctionComponent<InsightCardProps> = props => {
    const {
        insight: { id, view, source },
        location,
        onDelete,
        deleting,
    } = props

    // We support actions only over search and lang insights and not able to edit or delete
    // custom insight or backend insight.
    const hasMenu = useMemo(
        () => id.startsWith(InsightTypeSuffix.search) || id.startsWith(InsightTypeSuffix.langStats),
        [id]
    )

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
            {view === undefined || deleting ? (
                <>
                    <div className="flex-grow-1 d-flex flex-column align-items-center justify-content-center">
                        <LoadingSpinner /> {deleting ? 'Deleting code insight' : 'Loading code insight'}
                    </div>
                    <InsightDescription
                        className={styles.insightCardDescription}
                        title={id}
                        icon={getInsightViewIcon(source)}
                    />
                </>
            ) : isErrorLike(view) ? (
                <>
                    <ErrorAlert className="m-0" error={view} />
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
                            <h3 className={styles.insightCardTitle}>{view.title}</h3>
                            {view.subtitle && <div className={styles.insightCardSubtitle}>{view.subtitle}</div>}
                        </div>

                        {hasMenu && (
                            <InsightCardMenu
                                /* eslint-disable-next-line react/jsx-no-bind */
                                onDelete={() => onDelete(id)}
                            />
                        )}
                    </header>

                    <ViewContent {...props} viewContent={view.content} viewID={id} />
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
