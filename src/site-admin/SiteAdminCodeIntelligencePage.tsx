import { upperFirst } from 'lodash'
import * as React from 'react'
import { Observable, Subscription } from 'rxjs'
import { catchError, map } from 'rxjs/operators'
import { gql, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../util/errors'
import { SiteAdminLangServers } from './SiteAdminLangServers'

interface State {
    /** The language server management status, or an error, or undefined while loading. */
    statusOrError?: GQL.ILanguageServerManagementStatus | ErrorLike
}

/**
 * A page displaying information about code intelligence on this site.
 */
export class SiteAdminCodeIntelligencePage extends React.PureComponent<{}, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminCodeIntelligence')

        this.subscriptions.add(
            this.queryLanguageServerManagementStatus()
                .pipe(
                    catchError(err => [asError(err)]),
                    map(v => ({ statusOrError: v }))
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), err => console.error(err))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (this.state.statusOrError === undefined) {
            return null
        }

        return (
            <div>
                <PageTitle title="Code intelligence" />
                <h2>Code intelligence</h2>
                <p>
                    Sourcegraph uses language servers built on the{' '}
                    <a href="https://langserver.org/">Language Server Protocol</a> (LSP) standard to provide code
                    intelligence: hovers, definitions, references, implementations, etc. See &ldquo;<a
                        href="https://about.sourcegraph.com/docs/code-intelligence"
                        target="_blank"
                    >
                        Code intelligence overview
                    </a>&rdquo; for more information.
                </p>
                {this.state.statusOrError &&
                    (isErrorLike(this.state.statusOrError) ? (
                        <div className="alert alert-danger">
                            Error querying language server management capabilities:{' '}
                            {upperFirst(this.state.statusOrError.message)}
                        </div>
                    ) : (
                        !this.state.statusOrError.siteCanManage &&
                        this.state.statusOrError.reason && (
                            <div className="alert alert-info">{upperFirst(this.state.statusOrError.reason)}</div>
                        )
                    ))}
                <SiteAdminLangServers />
            </div>
        )
    }

    private queryLanguageServerManagementStatus(): Observable<GQL.ILanguageServerManagementStatus> {
        return queryGraphQL(gql`
            query LanguageServerManagementStatus {
                site {
                    languageServerManagementStatus {
                        siteCanManage
                        reason
                    }
                }
            }
        `).pipe(
            map(({ data, errors }) => {
                if (!data || !data.site || !data.site.languageServerManagementStatus) {
                    throw createAggregateError(errors)
                }
                return data.site.languageServerManagementStatus
            })
        )
    }
}
