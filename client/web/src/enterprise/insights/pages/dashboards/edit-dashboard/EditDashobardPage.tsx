import classNames from 'classnames'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useContext, useMemo } from 'react'
import { useHistory } from 'react-router'
import { Link } from 'react-router-dom'

import { asError } from '@sourcegraph/common'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Badge, Button, Container, LoadingSpinner, PageHeader } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../../auth'
import { HeroPage } from '../../../../../components/HeroPage'
import { LoaderButton } from '../../../../../components/LoaderButton'
import { Page } from '../../../../../components/Page'
import { PageTitle } from '../../../../../components/PageTitle'
import { CodeInsightsIcon } from '../../../components'
import { FORM_ERROR, SubmissionErrors } from '../../../components/form/hooks/useForm'
import { CodeInsightsBackendContext } from '../../../core/backend/code-insights-backend-context'
import { CustomInsightDashboard, isVirtualDashboard } from '../../../core/types'
import { isBuiltInInsightDashboard } from '../../../core/types/dashboard/real-dashboard'
import { isGlobalSubject, SupportedInsightSubject } from '../../../core/types/subjects'
import {
    DashboardCreationFields,
    InsightsDashboardCreationContent,
} from '../creation/components/insights-dashboard-creation-content/InsightsDashboardCreationContent'

import styles from './EditDashboardPage.module.scss'

interface EditDashboardPageProps {
    dashboardId: string
    authenticatedUser: Pick<AuthenticatedUser, 'id' | 'organizations' | 'username'>
}

/**
 * Displays the edit (configure) dashboard page.
 */
export const EditDashboardPage: React.FunctionComponent<EditDashboardPageProps> = props => {
    const { dashboardId, authenticatedUser } = props
    const history = useHistory()

    const { getDashboardById, getDashboardSubjects, updateDashboard } = useContext(CodeInsightsBackendContext)

    // Load edit dashboard information
    const subjects = useObservable(useMemo(() => getDashboardSubjects(), [getDashboardSubjects]))

    const dashboard = useObservable(
        useMemo(
            () => getDashboardById({ dashboardId }),
            // Load only on first render to avoid UI flashing after settings update
            // eslint-disable-next-line react-hooks/exhaustive-deps
            [dashboardId]
        )
    )

    // Loading state
    if (subjects === undefined || dashboard === undefined) {
        return <LoadingSpinner />
    }

    // In case if we got null that means we couldn't find this dashboard
    if (dashboard === null || isVirtualDashboard(dashboard) || isBuiltInInsightDashboard(dashboard)) {
        return (
            <HeroPage
                icon={MapSearchIcon}
                title="Oops, we couldn't find the dashboard"
                subtitle={
                    <span>
                        We couldn't find that dashboard. Try to find the dashboard with ID:
                        <Badge variant="secondary" as="code">
                            {dashboardId}
                        </Badge>{' '}
                        in your <Link to={`/users/${authenticatedUser?.username}/settings`}>user or org settings</Link>
                    </span>
                }
            />
        )
    }

    const handleSubmit = async (dashboardValues: DashboardCreationFields): Promise<SubmissionErrors> => {
        if (!dashboard) {
            return
        }

        const { name, visibility, type } = dashboardValues

        try {
            const updatedDashboard = await updateDashboard({
                previousDashboard: dashboard,
                nextDashboardInput: {
                    name,
                    visibility,
                    type,
                },
            }).toPromise()

            history.push(`/insights/dashboards/${updatedDashboard.id}`)
        } catch (error) {
            return { [FORM_ERROR]: asError(error) }
        }

        return
    }
    const handleCancel = (): void => history.goBack()

    return (
        <Page className={classNames('col-8', styles.page)}>
            <PageTitle title="Configure dashboard" />

            <PageHeader path={[{ icon: CodeInsightsIcon }, { text: 'Configure dashboard' }]} />

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
                <InsightsDashboardCreationContent
                    initialValues={getDashboardInitialValues(dashboard, subjects)}
                    subjects={subjects}
                    onSubmit={handleSubmit}
                >
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
                                label={formAPI.submitting ? 'Saving' : 'Save changes'}
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

function getDashboardInitialValues(
    dashboard: CustomInsightDashboard,
    subjects: SupportedInsightSubject[]
): DashboardCreationFields | undefined {
    if (dashboard.owner) {
        return { name: dashboard.title, visibility: dashboard.owner.id }
    }

    if (dashboard.grants) {
        const { users, organizations, global } = dashboard.grants
        const globalSubject = subjects.find(isGlobalSubject)

        if (global && globalSubject) {
            return { name: dashboard.title, visibility: globalSubject.id }
        }

        return {
            name: dashboard.title,
            visibility: users[0] ?? organizations[0] ?? 'unkown',
        }
    }

    return
}
