import NoEntryIcon from '@sourcegraph/icons/lib/NoEntry'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import * as GQL from '../backend/graphqlschema'
import { FilteredConnection } from '../components/FilteredConnection'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { Timestamp } from '../components/time/Timestamp'
import { fetchAllSurveyResponses } from '../marketing/backend'
import { eventLogger } from '../tracking/eventLogger'
import { userURL } from '../user'

interface SurveyResponseNodeProps {
    /**
     * The survey response to display in this list item.
     */
    node: GQL.ISurveyResponse
}

interface SurveyResponseNodeState {
    loading: boolean
    errorDescription?: string
}

class SurveyResponseNode extends React.PureComponent<SurveyResponseNodeProps, SurveyResponseNodeState> {
    public state: SurveyResponseNodeState = {
        loading: false,
    }

    private scoreToClassName(score: number): string {
        return score > 8 ? 'badge-success' : score > 6 ? 'badge-info' : 'badge-danger'
    }

    public render(): JSX.Element | null {
        return (
            <li className="list-group-item py-2">
                <div className="d-flex align-items-center justify-content-between">
                    <div>
                        <strong>
                            {this.props.node.user ? (
                                <Link to={userURL(this.props.node.user.username)}>{this.props.node.user.username}</Link>
                            ) : this.props.node.email ? (
                                this.props.node.email
                            ) : (
                                'anonymous user'
                            )}
                        </strong>
                        <div
                            className={`ml-4 badge badge-pill ${this.scoreToClassName(this.props.node.score)}`}
                            data-tooltip={`${this.props.node.score} out of 10`}
                        >
                            Score: {this.props.node.score}
                        </div>
                    </div>
                    <div>
                        <Timestamp date={this.props.node.createdAt} />
                    </div>
                </div>
                {(this.props.node.reason || this.props.node.better) && (
                    <dl className="mt-3">
                        {this.props.node.reason &&
                            this.props.node.reason !== '' && (
                                <>
                                    <dt>What is the most important reason for the score you gave Sourcegraph?</dt>
                                    <dd>{this.props.node.reason}</dd>
                                </>
                            )}
                        {this.props.node.reason && this.props.node.better && <div className="mt-2" />}
                        {this.props.node.better &&
                            this.props.node.better !== '' && (
                                <>
                                    <dt>What could Sourcegraph do to provide a better product?</dt>
                                    <dd>{this.props.node.better}</dd>
                                </>
                            )}
                    </dl>
                )}
            </li>
        )
    }
}
// site-admin-detail-list__header site-admin-detail-list__header--narrow site-admin-detail-list__header-narrow--center

interface Props extends RouteComponentProps<any> {}

export interface State {
    surveyResponses?: GQL.ISurveyResponse[]
    totalCount?: number
}

class FilteredSurveyResponseConnection extends FilteredConnection<GQL.ISurveyResponse, {}> {}

/**
 * A page displaying the survey responses on this site.
 */
export class SiteAdminSurveyResponsesPage extends React.Component<Props, State> {
    public state: State = {}

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminSurveyResponses')
    }

    public render(): JSX.Element | null {
        if (!window.context.hostSurveysLocallyEnabled) {
            return <HeroPage icon={NoEntryIcon} title="Surveys are not enabled" />
        }
        return (
            <div className="site-admin-survey-responses-page">
                <PageTitle title="Survey Responses - Admin" />
                <h2>Survey responses</h2>
                <p>
                    After using Sourcegraph for a few days, users are presented with a request to answer "How likely is
                    it that you would recommend Sourcegraph to a friend?" on a scale from 0–10 and to provide some
                    feedback. Responses are visible below (and are also sent to Sourcegraph).
                </p>

                <FilteredSurveyResponseConnection
                    className="list-group list-group-flush"
                    hideFilter={true}
                    noun="survey response"
                    pluralNoun="survey responses"
                    queryConnection={fetchAllSurveyResponses}
                    nodeComponent={SurveyResponseNode}
                    nodeComponentProps={{}}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }
}
