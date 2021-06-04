import classnames from 'classnames'
import React, { useCallback, useContext, useEffect } from 'react'
import { Redirect } from 'react-router'
import { useHistory } from 'react-router-dom'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { asError } from '@sourcegraph/shared/src/util/errors'
import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'

import { AuthenticatedUser } from '../../../../auth'
import { Page } from '../../../../components/Page'
import { PageTitle } from '../../../../components/PageTitle'
import { FORM_ERROR, FormChangeEvent } from '../../../components/form/hooks/useForm'
import { InsightsApiContext } from '../../../core/backend/api-provider'
import { addInsightToCascadeSetting } from '../../../core/jsonc-operation'

import {
    SearchInsightCreationContent,
    SearchInsightCreationContentProps,
} from './components/search-insight-creation-content/SearchInsightCreationContent'
import styles from './SearchInsightCreationPage.module.scss'
import { CreateInsightFormFields } from './types'
import { getSanitizedSearchInsight } from './utils/insight-sanitizer'

export interface SearchInsightCreationPageProps
    extends PlatformContextProps<'updateSettings'>,
        SettingsCascadeProps,
        TelemetryProps {
    /**
     * Authenticated user info, Used to decide where code insight will appears
     * in personal dashboard (private) or in organisation dashboard (public)
     * */
    authenticatedUser: Pick<AuthenticatedUser, 'id' | 'organizations'> | null
}

/** Displays create insight page with creation form. */
export const SearchInsightCreationPage: React.FunctionComponent<SearchInsightCreationPageProps> = props => {
    const { platformContext, authenticatedUser, settingsCascade, telemetryService } = props
    const { updateSubjectSettings, getSubjectSettings } = useContext(InsightsApiContext)
    const history = useHistory()

    const [initialFormValues, setInitialFormValues] = useLocalStorage<CreateInsightFormFields | undefined>(
        'insights.search-insight-creation',
        undefined
    )

    useEffect(() => {
        telemetryService.logViewEvent('CodeInsightsSearchBasedCreationPage')
    }, [telemetryService])

    const handleSubmit = useCallback<SearchInsightCreationContentProps['onSubmit']>(
        async values => {
            if (!authenticatedUser) {
                return
            }

            const { id: userID } = authenticatedUser

            const subjectID =
                values.visibility === 'personal'
                    ? userID
                    : // If this is not a 'personal' value than we are dealing with org id
                      values.visibility

            try {
                const settings = await getSubjectSettings(subjectID).toPromise()
                const insight = getSanitizedSearchInsight(values)
                const editedSettings = addInsightToCascadeSetting(settings.contents, insight)

                await updateSubjectSettings(platformContext, subjectID, editedSettings).toPromise()

                telemetryService.log('CodeInsightsSearchBasedCreationPageSubmitClick')

                // Clear initial values if user successfully created search insight
                setInitialFormValues(undefined)
                history.push('/insights')
            } catch (error) {
                return { [FORM_ERROR]: asError(error) }
            }

            return
        },
        [
            authenticatedUser,
            getSubjectSettings,
            updateSubjectSettings,
            platformContext,
            telemetryService,
            setInitialFormValues,
            history,
        ]
    )

    const handleChange = (event: FormChangeEvent<CreateInsightFormFields>): void => {
        setInitialFormValues(event.values)
    }

    const handleCancel = useCallback(() => {
        telemetryService.log('CodeInsightsSearchBasedCreationPageCancelClick')
        setInitialFormValues(undefined)
        history.push('/insights')
    }, [history, setInitialFormValues, telemetryService])

    // TODO [VK] Move this logic to high order component to simplify logic here
    if (authenticatedUser === null) {
        return <Redirect to="/" />
    }

    const {
        organizations: { nodes: orgs },
    } = authenticatedUser

    return (
        <Page className={classnames('col-10', styles.creationPage)}>
            <PageTitle title="Create new code insight" />

            <div className="mb-5">
                <h2>Create new code insight</h2>

                <p className="text-muted">
                    Search-based code insights analyze your code based on any search query.{' '}
                    <a href="https://docs.sourcegraph.com/code_insights" target="_blank" rel="noopener">
                        Learn more.
                    </a>
                </p>
            </div>

            <SearchInsightCreationContent
                className="pb-5"
                dataTestId="search-insight-create-page-content"
                settings={settingsCascade.final}
                initialValue={initialFormValues}
                organizations={orgs}
                onSubmit={handleSubmit}
                onCancel={handleCancel}
                onChange={handleChange}
            />
        </Page>
    )
}
