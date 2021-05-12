import * as jsonc from '@sqs/jsonc-parser'
import { uniqBy } from 'lodash'
import GearIcon from 'mdi-react/GearIcon'
import PlusIcon from 'mdi-react/PlusIcon'
import React, { useCallback, useEffect, useMemo, useContext, useState } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { PlatformContextProps } from '@sourcegraph/shared/out/src/platform/context'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
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

            if (isErrorLike(subjects) || !subjects) {
                // TODO [VK] Add error UI for case when we user for some reason doesn't have proper settings
                return
            }

            const subjectID = subjects.find(
                ({ settings }) => settings && !isErrorLike(settings) && !!settings[insightID]
            )?.subject?.id

            if (!subjectID) {
                return
            }

            setProcessingInsights(insights => ({ ...insights, [id]: true }))

            try {
                const settings = await getSubjectSettings(subjectID).toPromise()

                const edits = jsonc.modify(
                    settings.contents,
                    // According to our naming convention <type>.insight.<name>
                    [`${insightID}`],
                    undefined,
                    { formattingOptions: defaultFormattingOptions }
                )

                const editedSettings = jsonc.applyEdits(settings.contents, edits)

                await updateSubjectSettings(platformContext, subjectID, editedSettings).toPromise()
            } catch (error) {
                console.error(error)
            }

            setProcessingInsights(insights => ({ ...insights, [id]: false }))
        },
        [platformContext, settingsCascade, getSubjectSettings, updateSubjectSettings]
    )

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

    // Remove uniqBy when this extension api issue will be resolved
    // https://github.com/sourcegraph/sourcegraph/issues/20442
    const filteredViews = useMemo(() => {
        if (!views) {
            return views
        }

        return uniqBy(views, view => view.id)
    }, [views])

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
                {filteredViews === undefined ? (
                    <div className="d-flex w-100">
                        <LoadingSpinner className="my-4" />
                    </div>
                ) : (
                    <InsightsViewGrid
                        {...props}
                        views={filteredViews}
                        processingInsights={processingInsights}
                        onDelete={handleDelete}
                    />
                )}
            </Page>
        </div>
    )
}
