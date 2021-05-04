import * as jsonc from '@sqs/jsonc-parser'
import { camelCase } from 'lodash'
import React, { useCallback, useContext } from 'react'
import { Redirect } from 'react-router'
import { RouteComponentProps } from 'react-router-dom'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { asError } from '@sourcegraph/shared/src/util/errors'

import { AuthenticatedUser } from '../../../../auth'
import { Page } from '../../../../components/Page'
import { PageTitle } from '../../../../components/PageTitle'
import { FORM_ERROR } from '../../../components/form/hooks/useForm'
import { InsightsApiContext } from '../../../core/backend/api-provider'

import {
    LangStatsInsightCreationForm,
    LangStatsInsightCreationFormProps,
} from './components/lang-stats-insight-creation-form/LangStatsInsightCreationForm'

const DEFAULT_FINAL_SETTINGS = {}

const defaultFormattingOptions: jsonc.FormattingOptions = {
    eol: '\n',
    insertSpaces: true,
    tabSize: 2,
}

export interface CreationLangStatsInsightPageProps
    extends PlatformContextProps<'updateSettings'>,
        Pick<RouteComponentProps, 'history'>,
        SettingsCascadeProps {
    /**
     * Authenticated user info, Used to decide where code insight will appears
     * in personal dashboard (private) or in organisation dashboard (public)
     * */
    authenticatedUser: Pick<AuthenticatedUser, 'id' | 'organizations'> | null
}

export const CreationLangStatsInsightPage: React.FunctionComponent<CreationLangStatsInsightPageProps> = props => {
    const { history, authenticatedUser, settingsCascade, platformContext } = props
    const { getSubjectSettings, updateSubjectSettings } = useContext(InsightsApiContext)

    const handleSubmit = useCallback<LangStatsInsightCreationFormProps['onSubmit']>(
        async values => {
            if (!authenticatedUser) {
                return
            }

            const {
                id: userID,
                organizations: { nodes: orgs },
            } = authenticatedUser
            const subjectID =
                values.visibility === 'personal'
                    ? userID
                    : // TODO [VK] Add org picker in creation UI and not just pick first organization
                      orgs[0].id

            try {
                const settings = await getSubjectSettings(subjectID).toPromise()

                // TODO [VK] Change these settings when multi code insights stats
                // will be supported in code stats insight extension
                const newSettingsString = {
                    title: values.title,
                    repository: values.repository.trim(),
                    threshold: values.threshold,
                }

                const edits = jsonc.modify(
                    settings.contents,
                    // According to our naming convention <type>.insight.<name>
                    [`codeStatsInsights.insight.${camelCase(values.title)}`],
                    newSettingsString,
                    { formattingOptions: defaultFormattingOptions }
                )

                const editedSettings = jsonc.applyEdits(settings.contents, edits)

                await updateSubjectSettings(platformContext, subjectID, editedSettings).toPromise()

                history.push('/insights')
            } catch (error) {
                return { [FORM_ERROR]: asError(error) }
            }

            return
        },
        [history, updateSubjectSettings, getSubjectSettings, platformContext, authenticatedUser]
    )

    const handleCancel = useCallback(() => {
        history.push('/insights')
    }, [history])

    if (authenticatedUser === null) {
        return <Redirect to="/" />
    }

    return (
        <Page className="col-8">
            <PageTitle title="Create new code insight" />

            <div className="mb-5">
                <h2>Set up new language usage insight</h2>

                <p className="text-muted">
                    Shows usage of languages in your repository based on number of lines of code.{' '}
                    <a
                        href="https://docs.sourcegraph.com/dev/background-information/insights"
                        target="_blank"
                        rel="noopener"
                    >
                        Learn more.
                    </a>
                </p>
            </div>

            <LangStatsInsightCreationForm
                className="pb-5"
                settings={settingsCascade.final ?? DEFAULT_FINAL_SETTINGS}
                onSubmit={handleSubmit}
                onCancel={handleCancel}
            />
        </Page>
    )
}
