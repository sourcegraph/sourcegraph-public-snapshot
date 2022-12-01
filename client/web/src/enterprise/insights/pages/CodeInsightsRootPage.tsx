import { Suspense, FC } from 'react'

import { mdiPlus } from '@mdi/js'
import { useHistory } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import {
    Button,
    Link,
    PageHeader,
    Tabs,
    TabList,
    Tab,
    Icon,
    TabPanels,
    TabPanel,
    LoadingSpinner,
} from '@sourcegraph/wildcard'

import { CodeInsightsIcon } from '../../../insights/Icons'
import { CodeInsightsPage } from '../components'
import { useQueryParameters } from '../hooks'

import { DashboardsView } from './dashboards/dashboard-view/DashboardsView'

import styles from './CodeInsightsRootPage.module.scss'

const LazyCodeInsightsGettingStartedPage = lazyComponent(
    () => import('./landing/getting-started/CodeInsightsGettingStartedPage'),
    'CodeInsightsGettingStartedPage'
)

export enum CodeInsightsRootPageTab {
    Dashboards,
    AllInsights,
    GettingStarted,
}

interface CodeInsightsRootPageProps extends TelemetryProps {
    dashboardId?: string
    activeTab: CodeInsightsRootPageTab
}

export const CodeInsightsRootPage: FC<CodeInsightsRootPageProps> = props => {
    const { dashboardId, activeTab, telemetryService } = props

    const history = useHistory()
    const { dashboardId: queryParamDashboardId } = useQueryParameters(['dashboardId'])

    const handleTabNavigationChange = (selectedTab: CodeInsightsRootPageTab): void => {
        switch (selectedTab) {
            case CodeInsightsRootPageTab.Dashboards:
                return history.push(`/insights/dashboards/${queryParamDashboardId ?? ''}`)
            case CodeInsightsRootPageTab.AllInsights:
                return history.push(`/insights/dashboards/all?dashboardId=${dashboardId}`)
            case CodeInsightsRootPageTab.GettingStarted:
                return history.push(`/insights/about?dashboardId=${dashboardId}`)
        }
    }

    return (
        <CodeInsightsPage>
            <PageHeader
                path={[{ icon: CodeInsightsIcon, text: 'Insights' }]}
                actions={
                    <CodeInsightHeaderActions
                        dashboardId={dashboardId ?? queryParamDashboardId}
                        telemetryService={telemetryService}
                    />
                }
                className={styles.header}
            />

            <Tabs
                index={activeTab}
                lazy={true}
                size="medium"
                className={styles.tabs}
                onChange={handleTabNavigationChange}
            >
                <TabList>
                    <Tab index={CodeInsightsRootPageTab.Dashboards}>Dashboards</Tab>
                    <Tab index={CodeInsightsRootPageTab.AllInsights}>All insights</Tab>
                    <Tab index={CodeInsightsRootPageTab.GettingStarted}>Getting started</Tab>
                </TabList>
                <TabPanels className={styles.tabPanels}>
                    <TabPanel tabIndex={-1}>
                        <DashboardsView dashboardId={dashboardId} telemetryService={telemetryService} />
                    </TabPanel>
                    <TabPanel tabIndex={-1}>
                        <Suspense fallback={<LoadingSpinner aria-label="Loading Code Insights Getting started page" />}>
                            <LazyCodeInsightsGettingStartedPage telemetryService={telemetryService} />
                        </Suspense>
                    </TabPanel>
                </TabPanels>
            </Tabs>
        </CodeInsightsPage>
    )
}

interface CodeInsightHeaderActionsProps extends TelemetryProps {
    dashboardId?: string
}

const CodeInsightHeaderActions: FC<CodeInsightHeaderActionsProps> = props => {
    const { dashboardId, telemetryService } = props

    return (
        <>
            <Button
                as={Link}
                to="/insights/add-dashboard"
                variant="secondary"
                className="mr-2"
                aria-label="Add dashboard"
            >
                <Icon aria-hidden={true} svgPath={mdiPlus} /> Add dashboard
            </Button>
            <Button
                as={Link}
                to={`/insights/create?dashboardId=${dashboardId}`}
                variant="primary"
                onClick={() => telemetryService.log('InsightAddMoreClick')}
            >
                <Icon aria-hidden={true} svgPath={mdiPlus} /> Create insight
            </Button>
        </>
    )
}
