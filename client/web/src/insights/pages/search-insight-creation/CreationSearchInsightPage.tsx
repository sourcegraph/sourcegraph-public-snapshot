import * as jsonc from '@sqs/jsonc-parser'
import { FORM_ERROR } from 'final-form'
import React, { useCallback, useContext } from 'react'
import { Redirect } from 'react-router'
import { RouteComponentProps } from 'react-router-dom'
import * as uuid from 'uuid'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { asError } from '@sourcegraph/shared/src/util/errors'

import { AuthenticatedUser } from '../../../auth'
import { Page } from '../../../components/Page'
import { PageTitle } from '../../../components/PageTitle'
import { InsightsApiContext } from '../../core/backend/api-provider'

import {
    CreationSearchInsightForm,
    CreationSearchInsightFormProps,
} from './components/creation-search-insight-form/CreationSearchInsightForm'
import styles from './CreationSearchInsightPage.module.scss'

const defaultFormattingOptions: jsonc.FormattingOptions = {
    eol: '\n',
    insertSpaces: true,
    tabSize: 2,
}

export interface CreationSearchInsightPageProps extends PlatformContextProps, RouteComponentProps {
    /**
     * Authenticated user info, Used to decide where code insight will appears
     * in personal dashboard (private) or in organisation dashboard (public)
     * */
    authenticatedUser: AuthenticatedUser | null
}

/** Displays create insight page with creation form. */
export const CreationSearchInsightPage: React.FunctionComponent<CreationSearchInsightPageProps> = props => {
    const { platformContext, authenticatedUser, history } = props
    const { updateSubjectSettings, getSubjectSettings } = useContext(InsightsApiContext)

    const handleSubmit = useCallback<CreationSearchInsightFormProps['onSubmit']>(
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
                    : // TODO [VK] Add orgs picker in creation UI and not just pick first organization
                      orgs[0].id

            try {
                const settings = await getSubjectSettings(subjectID).toPromise()
                const insightID = uuid.v4()

                const newSettingsString = JSON.stringify(
                    {
                        title: values.title,
                        repositories: values.repositories.split(','),
                        series: values.series.map(line => ({
                            name: line.name,
                            // Query field is a reg exp field for code insight query setting
                            // Native html input element adds escape symbols by himself
                            // to prevent this behavior below we replace double escaping
                            // with just one series of escape characters e.g. - //
                            query: line.query.replace(/\\\\/g, '\\'),
                            stroke: line.color,
                        })),
                        step: {
                            [values.step]: +values.stepValue,
                        },
                    },
                    null,
                    2,
                )

                const edits = jsonc.modify(
                    settings.contents,
                    [`searchInsights.insight.${insightID}`],
                    newSettingsString,
                    { formattingOptions: defaultFormattingOptions }
                )

                const editedSettings = jsonc.applyEdits(settings.contents, edits);

                await updateSubjectSettings(
                    platformContext,
                    subjectID,
                    editedSettings,
                ).toPromise()

                history.push('/insights')
            } catch (error) {
                return { [FORM_ERROR]: asError(error) }
            }

            return
        },
        [history, updateSubjectSettings, getSubjectSettings, platformContext, authenticatedUser]
    )

    if (authenticatedUser === null) {
        return <Redirect to="/" />
    }

    return (
        <Page className="col-8">
            <PageTitle title="Create new code insight" />

            <div className={styles.createInsightPageSubTitleContainer}>
                <h2>Create new code insight</h2>

                <p className="text-muted">
                    Search-based code insights analyse your code based on any search query.{' '}
                    <a
                        href="https://docs.sourcegraph.com/code_monitoring/how-tos/starting_points"
                        target="_blank"
                        rel="noopener"
                    >
                        Learn more.
                    </a>
                </p>
            </div>

            <CreationSearchInsightForm onSubmit={handleSubmit} />
        </Page>
    )
}
