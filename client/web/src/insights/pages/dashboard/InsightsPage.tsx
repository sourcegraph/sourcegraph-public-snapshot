import * as jsonc from '@sqs/jsonc-parser'
import GearIcon from 'mdi-react/GearIcon'
import PlusIcon from 'mdi-react/PlusIcon'
import React, { useCallback, useEffect, useMemo, useContext, useState } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { PageHeader } from '@sourcegraph/wildcard'

import { FeedbackBadge } from '../../../components/FeedbackBadge'
import { Page } from '../../../components/Page'
import { InsightsIcon, InsightsViewGrid, InsightsViewGridProps } from '../../components'
import { InsightsApiContext } from '../../core/backend/api-provider'
import { defaultFormattingOptions } from '../../core/jsonc-settings'

export interface InsightsPageProps
    extends ExtensionsControllerProps,
        Omit<InsightsViewGridProps, 'views'>,
        TelemetryProps,
        PlatformContextProps<'updateSettings'> {
    isCreationUIEnabled: boolean
}

/**
 * Renders insight page. (insights grid and navigation for insight)
 */
export const InsightsPage: React.FunctionComponent<InsightsPageProps> = props => {
    const { isCreationUIEnabled, settingsCascade, platformContext } = props
    const { getInsightCombinedViews, getSubjectSettings, updateSubjectSettings } = useContext(InsightsApiContext)

    const views = useObservable(
        useMemo(() => getInsightCombinedViews(props.extensionsController?.extHostAPI), [
            props.extensionsController,
            getInsightCombinedViews,
        ])
    )

    // We should disable delete and any other actions if we already have started operation
    // over some particular insight
    const [processingInsights, setProcessingInsights] = useState({})

    const handleDelete = useCallback(
        async (id: string) => {
            // According to our naming convention of insight
            // <type>.<name>.<render view = insight page | directory | home page>
            const insightID = id.split('.').slice(0, -1).join('.')
            const subjects = settingsCascade.subjects

            const subjectID = subjects?.find(
                ({ settings }) => settings && !isErrorLike(settings) && !!settings[insightID]
            )?.subject?.id

            if (!subjectID) {
                return
            }

            setProcessingInsights(insights => ({ ...insights, [id]: true }))

            try {
                // Fetch the settings of particular subject which the insight belongs to
                const settings = await getSubjectSettings(subjectID).toPromise()

                // Remove insight settings from subject (user/org settings)
                const edits = jsonc.modify(
                    settings.contents,
                    // According to our naming convention <type>.insight.<name>
                    [`${insightID}`],
                    undefined,
                    { formattingOptions: defaultFormattingOptions }
                )

                const editedSettings = jsonc.applyEdits(settings.contents, edits)

                // Update local settings of application with new settings without insight
                await updateSubjectSettings(platformContext, subjectID, editedSettings).toPromise()
            } catch (error) {
                // TODO [VK] Improve error UI for deleting
                console.error(error)
            }

            setProcessingInsights(insights => ({ ...insights, [id]: false }))
        },
        [platformContext, settingsCascade, getSubjectSettings, updateSubjectSettings]
    )

    // Tracking handlers and logic
    useEffect(() => {
        props.telemetryService.logViewEvent('Insights')
    }, [props.telemetryService])

    const logConfigureClick = useCallback(() => {
        props.telemetryService.log('InsightConfigureClick')
    }, [props.telemetryService])

    const logAddMoreClick = useCallback(() => {
        props.telemetryService.log('InsightAddMoreClick')
    }, [props.telemetryService])

    const configureURL = isCreationUIEnabled ? '/insights/create-intro' : '/user/settings'

    return (
        <div className="w-100">
            <Page>
                <PageHeader
                    annotation={<FeedbackBadge status="prototype" feedback={{ mailto: 'support@sourcegraph.com' }} />}
                    path={[{ icon: InsightsIcon, text: 'Code insights' }]}
                    actions={
                        <>
                            <Link
                                to="/extensions?query=category:Insights"
                                onClick={logAddMoreClick}
                                className="btn btn-secondary mr-1"
                            >
                                <PlusIcon className="icon-inline" /> Add more insights
                            </Link>
                            <Link to={configureURL} onClick={logConfigureClick} className="btn btn-secondary">
                                <GearIcon className="icon-inline" /> Configure insights
                            </Link>
                        </>
                    }
                    className="mb-3"
                />
                {views === undefined ? (
                    <div className="d-flex w-100">
                        <LoadingSpinner className="my-4" />
                    </div>
                ) : (
                    <InsightsViewGrid
                        {...props}
                        views={views}
                        processingInsights={processingInsights}
                        onDelete={handleDelete}
                    />
                )}
            </Page>
        </div>
    )
}
