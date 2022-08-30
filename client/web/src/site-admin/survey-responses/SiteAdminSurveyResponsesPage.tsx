import React, { useEffect } from 'react'

import { RouteComponentProps } from 'react-router'
import { Subscription } from 'rxjs'

import { useLocalStorage, Tab, TabList, TabPanel, TabPanels, Tabs, H2, H3, Text } from '@sourcegraph/wildcard'

import { FilteredConnection } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import { SingleValueCard } from '../../components/SingleValueCard'
import {
    SurveyResponseAggregateFields,
    SurveyResponseFields,
    UserWithSurveyResponseFields,
} from '../../graphql-operations'
import {
    fetchAllSurveyResponses,
    fetchAllUsersWithSurveyResponses,
    fetchSurveyResponseAggregates,
} from '../../marketing/backend'
import { eventLogger } from '../../tracking/eventLogger'
import { USER_ACTIVITY_FILTERS } from '../SiteAdminUsageStatisticsPage'

import { GenericSurveyResponseNode } from './SiteAdminGenericSurveyResponseNode'
import { scoreToClassSuffix } from './utils'

import styles from './SiteAdminSurveyResponsesPage.module.scss'
import { UserSurveyResponseNode } from './SiteAdminUserSurveyResponseNode'

const UserSurveyResponsesHeader: React.FunctionComponent<
    React.PropsWithChildren<{ nodes: UserWithSurveyResponseFields[] }>
> = () => (
    <thead>
        <tr>
            <th>User</th>
            <th>Last active on Sourcegraph</th>
            <th>Latest survey response</th>
            <th />
        </tr>
    </thead>
)

interface SiteAdminSurveyResponsesSummaryState {
    summary?: SurveyResponseAggregateFields
}

class SiteAdminSurveyResponsesSummary extends React.PureComponent<{}, SiteAdminSurveyResponsesSummaryState> {
    private subscriptions = new Subscription()
    public state: SiteAdminSurveyResponsesSummaryState = {}
    constructor(props: {}) {
        super(props)
        this.subscriptions.add(fetchSurveyResponseAggregates().subscribe(summary => this.setState({ summary })))
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.state.summary) {
            return null
        }
        const anyResults = this.state.summary.last30DaysCount > 0
        let npsText = `${this.state.summary.netPromoterScore}`
        if (this.state.summary.netPromoterScore > 0) {
            npsText = `+${npsText}`
        } else if (this.state.summary.netPromoterScore < 0) {
            npsText = `${npsText}`
        }
        const npsClass =
            this.state.summary.netPromoterScore > 0
                ? 'text-success'
                : this.state.summary.netPromoterScore < 0
                ? 'text-danger'
                : 'text-info'
        const roundAvg = Math.round(this.state.summary.averageScore * 10) / 10
        return (
            <div className="mb-2">
                <H3>Summary</H3>
                <div className={styles.container}>
                    <SingleValueCard
                        className={styles.item}
                        value={this.state.summary.last30DaysCount}
                        title="Number of submissions"
                        subTitle="Last 30 days"
                    />
                    <SingleValueCard
                        className={styles.item}
                        value={anyResults ? roundAvg : '-'}
                        title="Average score"
                        subTitle="Last 30 days"
                        valueTooltip={`${roundAvg} out of 10`}
                        valueClassName={anyResults ? `text-${scoreToClassSuffix(roundAvg)}` : ''}
                    />
                    <SingleValueCard
                        className={styles.item}
                        value={anyResults ? npsText : '-'}
                        title="Net promoter score"
                        subTitle="Last 30 days"
                        valueTooltip={`${npsText} (between -100 and +100)`}
                        valueClassName={anyResults ? npsClass : ''}
                    />
                </div>
            </div>
        )
    }
}

interface Props extends RouteComponentProps<{}> {}

class FilteredSurveyResponseConnection extends FilteredConnection<SurveyResponseFields, {}> {}
class FilteredUserSurveyResponseConnection extends FilteredConnection<UserWithSurveyResponseFields, {}> {}

const LAST_TAB_STORAGE_KEY = 'site-admin-survey-responses-last-tab'
/**
 * A page displaying the survey responses on this site.
 */

export const SiteAdminSurveyResponsesPage: React.FunctionComponent<React.PropsWithChildren<Props>> = props => {
    const [persistedTabIndex, setPersistedTabIndex] = useLocalStorage(LAST_TAB_STORAGE_KEY, 0)

    useEffect(() => {
        eventLogger.logViewEvent('SiteAdminSurveyResponses')
    }, [])

    return (
        <div className="site-admin-survey-responses-page">
            <PageTitle title="User feedback survey - Admin" />
            <H2>User feedback survey</H2>
            <Text>
                After using Sourcegraph for a few days, users are presented with a request to answer "How likely is it
                that you would recommend Sourcegraph to a friend?" on a scale from 0–10 and to provide some feedback.
                Responses are visible below (and are also sent to Sourcegraph).
            </Text>

            <SiteAdminSurveyResponsesSummary />

            <H3>Responses</H3>

            <Tabs defaultIndex={persistedTabIndex} onChange={setPersistedTabIndex}>
                <TabList>
                    <Tab>Chronological feed</Tab>
                    <Tab>Sort by user</Tab>
                </TabList>
                <TabPanels>
                    <TabPanel>
                        <FilteredSurveyResponseConnection
                            key="chronological"
                            className="list-group list-group-flush"
                            hideSearch={true}
                            noun="survey response"
                            pluralNoun="survey responses"
                            queryConnection={fetchAllSurveyResponses}
                            nodeComponent={GenericSurveyResponseNode}
                            history={props.history}
                            location={props.location}
                        />
                    </TabPanel>
                    <TabPanel>
                        <FilteredUserSurveyResponseConnection
                            key="by-user"
                            listComponent="table"
                            headComponent={UserSurveyResponsesHeader}
                            className="table mt-2 site-admin-survey-responses-connection"
                            hideSearch={false}
                            filters={USER_ACTIVITY_FILTERS}
                            noun="user"
                            pluralNoun="users"
                            queryConnection={fetchAllUsersWithSurveyResponses}
                            nodeComponent={UserSurveyResponseNode}
                            history={props.history}
                            location={props.location}
                        />
                    </TabPanel>
                </TabPanels>
            </Tabs>
        </div>
    )
}
