import ErrorIcon from '@sourcegraph/icons/lib/Error'
import { upperFirst } from 'lodash-es'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs/Observable'
import { catchError } from 'rxjs/operators/catchError'
import { concat } from 'rxjs/operators/concat'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { switchMap } from 'rxjs/operators/switchMap'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { gql, queryGraphQL } from '../../backend/graphql'
import { HeroPage } from '../../components/HeroPage'
import { PageTitle } from '../../components/PageTitle'
import { SettingsFile } from '../../settings/SettingsFile'
import { eventLogger } from '../../tracking/eventLogger'
import { refreshConfiguration } from '../../user/settings/backend'
import { createAggregateError, ErrorLike, isErrorLike } from '../../util/errors'
import { OrgAreaPageProps } from '../area/OrgArea'
import { updateOrgSettings } from '../backend'

function fetchOrgSettings(args: { id: string }): Observable<GQL.ISettings | null> {
    return queryGraphQL(
        gql`
            query OrganizationSettings($id: ID!) {
                node(id: $id) {
                    ... on Org {
                        latestSettings {
                            id
                            configuration {
                                contents
                            }
                        }
                    }
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.node || errors) {
                throw createAggregateError(errors)
            }
            const org = data.node as GQL.IOrg
            return org.latestSettings
        })
    )
}

interface Props extends OrgAreaPageProps, RouteComponentProps<{}> {
    isLightTheme: boolean
}

interface State {
    settingsOrError?: GQL.ISettings | null | ErrorLike
    commitError?: Error
}

export class OrgSettingsConfigurationPage extends React.PureComponent<Props, State> {
    public state: State = {}

    private orgChanges = new Subject<{ id: GQLID /* org ID */ }>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Load settings.
        this.subscriptions.add(
            this.orgChanges
                .pipe(
                    distinctUntilChanged(),
                    switchMap(({ id }) => {
                        type PartialStateUpdate = Pick<State, 'settingsOrError'>
                        return fetchOrgSettings({ id }).pipe(
                            catchError(error => [error]),
                            map(c => ({ settingsOrError: c } as PartialStateUpdate))
                        )
                    })
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), err => console.error(err))
        )

        // Clear settings when org ID changes.
        this.subscriptions.add(
            this.orgChanges
                .pipe(distinctUntilChanged((a, b) => a.id === b.id))
                .subscribe(() => this.setState({ settingsOrError: undefined }))
        )

        // Log view event.
        this.subscriptions.add(
            this.orgChanges
                .pipe(
                    distinctUntilChanged((a, b) => a.id === b.id),
                    tap(() => eventLogger.logViewEvent('OrgSettingsConfiguration'))
                )
                .subscribe()
        )

        this.orgChanges.next(this.props.org)
    }

    public componentWillReceiveProps(props: Props): void {
        if (props.org !== this.props.org) {
            this.orgChanges.next(props.org)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (this.state.settingsOrError === undefined) {
            return null // loading
        }
        if (isErrorLike(this.state.settingsOrError)) {
            // TODO!(sqs): show a 404 if org not found, instead of a generic error
            return <HeroPage icon={ErrorIcon} title="Error" subtitle={upperFirst(this.state.settingsOrError.message)} />
        }

        return (
            <div className="settings-file-container">
                <PageTitle title="Organization configuration" />
                <h2>Configuration</h2>
                <p>View and edit your organization's search scopes and saved queries.</p>
                <SettingsFile
                    settings={this.state.settingsOrError}
                    commitError={this.state.commitError}
                    onDidCommit={this.onDidCommit}
                    onDidDiscard={this.onDidDiscard}
                    history={this.props.history}
                    isLightTheme={this.props.isLightTheme}
                />
                <small className="form-text">
                    Documentation:{' '}
                    <a target="_blank" href="https://about.sourcegraph.com/docs/server/config/search-scopes">
                        Customizing search scopes for org members
                    </a>
                </small>
            </div>
        )
    }

    private onDidCommit = (lastKnownSettingsID: number | null, contents: string) =>
        updateOrgSettings(this.props.org.id, lastKnownSettingsID, contents)
            .pipe(mergeMap(() => refreshConfiguration().pipe(concat([null]))))
            .subscribe(
                () => {
                    this.setState({ commitError: undefined })
                    this.orgChanges.next({ id: this.props.org.id })
                },
                err => {
                    this.setState({ commitError: err })
                    console.error(err)
                }
            )

    private onDidDiscard = (): void => {
        this.setState({ commitError: undefined })
    }
}
