import classnames from 'classnames'
import { camelCase } from 'lodash'
import React, { useContext } from 'react'
import { useHistory } from 'react-router-dom'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { asError } from '@sourcegraph/shared/src/util/errors'
import { PageHeader, Container } from '@sourcegraph/wildcard/src'

import { AuthenticatedUser } from '../../../../auth'
import { LoaderButton } from '../../../../components/LoaderButton'
import { Page } from '../../../../components/Page'
import { PageTitle } from '../../../../components/PageTitle'
import { Settings } from '../../../../schema/settings.schema'
import { CodeInsightsIcon } from '../../../components'
import { FORM_ERROR, SubmissionErrors } from '../../../components/form/hooks/useForm'
import { InsightsApiContext } from '../../../core/backend/api-provider'
import { addDashboardToSettings } from '../../../core/settings-action/dashboards'

import {
    DashboardCreationFields,
    InsightsDashboardCreationContent,
} from './components/insights-dashboard-creation-content/InsightsDashboardCreationContent'
import { useDashboardSettings } from './hooks/use-dashboard-settings'
import styles from './InsightsDashboardCreationPage.module.scss'
import { createSanitizedDashboard } from './utils/dashboard-sanitizer'

interface InsightsDashboardCreationPageProps
    extends PlatformContextProps<'updateSettings'>,
        TelemetryProps,
        SettingsCascadeProps<Settings> {
    authenticatedUser: AuthenticatedUser
}

export const InsightsDashboardCreationPage: React.FunctionComponent<InsightsDashboardCreationPageProps> = props => {
    const { platformContext, telemetryService, authenticatedUser, settingsCascade } = props

    const history = useHistory()
    const { updateSubjectSettings, getSubjectSettings } = useContext(InsightsApiContext)

    const finalSettings = useDashboardSettings({ settingsCascade })

    const handleSubmit = async (values: DashboardCreationFields): Promise<void | SubmissionErrors> => {
        const { id: userID } = authenticatedUser

        const subjectID =
            values.visibility === 'personal'
                ? userID
                : // If this is not a 'personal' value than we are dealing with an org id
                  values.visibility

        try {
            const settings = await getSubjectSettings(subjectID).toPromise()
            const dashboard = createSanitizedDashboard(values)
            const editedSettings = addDashboardToSettings(settings.contents, dashboard)

            await updateSubjectSettings(platformContext, subjectID, editedSettings).toPromise()

            telemetryService.log('CodeInsightsDashboardCreationPageSubmitClick')

            // Navigate user to the dashboard page with new created dashboard
            history.push(`/insights/dashboards/${camelCase(dashboard.title)}`)
        } catch (error) {
            return { [FORM_ERROR]: asError(error) }
        }

        return
    }

    const handleCancel = (): void => history.goBack()

    return (
        <Page className={classnames('col-8', styles.page)}>
            <PageTitle title="Create new code insight" />

            <PageHeader path={[{ icon: CodeInsightsIcon }, { text: 'Add new dashboard' }]} />

            <Container className="mt-4">
                <InsightsDashboardCreationContent
                    dashboardsSettings={finalSettings}
                    organizations={authenticatedUser.organizations.nodes}
                    onSubmit={handleSubmit}
                >
                    {formAPI => (
                        <>
                            <button type="button" className="btn btn-outline-secondary mb-2" onClick={handleCancel}>
                                Cancel
                            </button>

                            <LoaderButton
                                alwaysShowLabel={true}
                                data-testid="insight-save-button"
                                loading={formAPI.submitting}
                                label={formAPI.submitting ? 'Submitting' : 'Create dashboard'}
                                spinnerClassName="mr-2"
                                type="submit"
                                disabled={formAPI.submitting}
                                className="d-flex btn btn-primary ml-2 mb-2"
                            />
                        </>
                    )}
                </InsightsDashboardCreationContent>
            </Container>
        </Page>
    )
}
