import classnames from 'classnames'
import { camelCase } from 'lodash'
import React, { useContext, useMemo } from 'react'
import { useHistory } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { asError } from '@sourcegraph/shared/src/util/errors'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { PageHeader, Container, Button, LoadingSpinner } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../../auth'
import { LoaderButton } from '../../../../../components/LoaderButton'
import { Page } from '../../../../../components/Page'
import { PageTitle } from '../../../../../components/PageTitle'
import { CodeInsightsIcon } from '../../../components'
import { FORM_ERROR, SubmissionErrors } from '../../../components/form/hooks/useForm'
import { CodeInsightsBackendContext } from '../../../core/backend/code-insights-backend-context'

import {
    DashboardCreationFields,
    InsightsDashboardCreationContent,
} from './components/insights-dashboard-creation-content/InsightsDashboardCreationContent'
import styles from './InsightsDashboardCreationPage.module.scss'

interface InsightsDashboardCreationPageProps extends TelemetryProps {
    authenticatedUser: AuthenticatedUser
}

export const InsightsDashboardCreationPage: React.FunctionComponent<InsightsDashboardCreationPageProps> = props => {
    const { telemetryService } = props

    const history = useHistory()
    const { createDashboard, getInsightSubjects } = useContext(CodeInsightsBackendContext)

    const subjects = useObservable(useMemo(() => getInsightSubjects(), [getInsightSubjects]))

    const handleSubmit = async (values: DashboardCreationFields): Promise<void | SubmissionErrors> => {
        try {
            await createDashboard(values).toPromise()

            telemetryService.log('CodeInsightsDashboardCreationPageSubmitClick')

            // Navigate user to the dashboard page with new created dashboard
            history.push(`/insights/dashboards/${camelCase(values.name)}`)
        } catch (error) {
            return { [FORM_ERROR]: asError(error) }
        }

        return
    }

    const handleCancel = (): void => history.goBack()

    // Loading state
    if (subjects === undefined) {
        return <LoadingSpinner />
    }

    return (
        <Page className={classnames('col-8', styles.page)}>
            <PageTitle title="Add new dashboard" />

            <PageHeader path={[{ icon: CodeInsightsIcon }, { text: 'Add new dashboard' }]} />

            <span className="text-muted d-block mt-2">
                Dashboards group your insights and let you share them with others.{' '}
                <a
                    href="https://docs.sourcegraph.com/code_insights/explanations/viewing_code_insights"
                    target="_blank"
                    rel="noopener"
                >
                    Learn more.
                </a>
            </span>

            <Container className="mt-4">
                <InsightsDashboardCreationContent subjects={subjects} onSubmit={handleSubmit}>
                    {formAPI => (
                        <>
                            <Button
                                type="button"
                                variant="secondary"
                                outline={true}
                                className="mb-2"
                                onClick={handleCancel}
                            >
                                Cancel
                            </Button>

                            <LoaderButton
                                alwaysShowLabel={true}
                                data-testid="insight-save-button"
                                loading={formAPI.submitting}
                                label={formAPI.submitting ? 'Creating' : 'Create dashboard'}
                                type="submit"
                                disabled={formAPI.submitting}
                                className="btn btn-primary ml-2 mb-2"
                            />
                        </>
                    )}
                </InsightsDashboardCreationContent>
            </Container>
        </Page>
    )
}
