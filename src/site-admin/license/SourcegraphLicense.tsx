import { gql, queryGraphQL } from '@sourcegraph/webapp/dist/backend/graphql'
import * as GQL from '@sourcegraph/webapp/dist/backend/graphqlschema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '@sourcegraph/webapp/dist/util/errors'
import { pluralize } from '@sourcegraph/webapp/dist/util/strings'
import format from 'date-fns/format'
import formatDistanceStrict from 'date-fns/formatDistanceStrict'
import isAfter from 'date-fns/isAfter'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { Observable, Subscription } from 'rxjs'
import { catchError, map } from 'rxjs/operators'

interface Props {
    className?: string
}

interface State {
    /** The Sourcegraph license info, or an error, or undefined while loading. */
    infoOrError?: GQL.ISourcegraphLicenseInfo | ErrorLike
}

/**
 * A component displaying information about and the status of the Sourcegraph license.
 */
export class SourcegraphLicense extends React.Component<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.querySourcegraphLicenseInfo()
                .pipe(
                    catchError(err => [asError(err)]),
                    map(v => ({ infoOrError: v }))
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), err => console.error(err))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (this.state.infoOrError === undefined) {
            return null
        }

        return (
            <>
                <div className={`sourcegraph-license card ${this.props.className || ''}`}>
                    <div className="sourcegraph-license__bg" />
                    <div className="card-body d-flex">
                        <img className="sourcegraph-license__logo mr-1 p-2" src="/.assets/img/sourcegraph-mark.svg" />
                        <div className="mt-2">
                            {isErrorLike(this.state.infoOrError) ? (
                                <div className="alert alert-danger">
                                    Error querying for license information: {upperFirst(this.state.infoOrError.message)}
                                </div>
                            ) : (
                                <>
                                    <h2 className="font-weight-normal">Sourcegraph License</h2>
                                    <h3 className="text-muted font-weight-bold">
                                        {this.state.infoOrError.plan} &mdash;{' '}
                                        {this.state.infoOrError.maxUserCount === null
                                            ? 'Unlimited users'
                                            : `${this.state.infoOrError.maxUserCount} ${pluralize(
                                                  'user',
                                                  this.state.infoOrError.maxUserCount
                                              )}`}
                                    </h3>
                                    {this.state.infoOrError.expiresAt !== null && (
                                        <p className="text-muted">
                                            Valid until {format(this.state.infoOrError.expiresAt, 'MMMM dd, yyyy')}{' '}
                                            {isAfter(this.state.infoOrError.expiresAt, new Date()) && (
                                                <>
                                                    (
                                                    {formatDistanceStrict(this.state.infoOrError.expiresAt, new Date())}{' '}
                                                    remaining)
                                                </>
                                            )}
                                        </p>
                                    )}
                                </>
                            )}
                        </div>
                    </div>
                    {!isErrorLike(this.state.infoOrError) && (
                        <>
                            {this.state.infoOrError.maxUserCount !== null && (
                                <div className="card-footer d-flex align-items-center justify-content-between">
                                    <div>
                                        <strong>User licenses:</strong> {this.state.infoOrError.userCount} used /{' '}
                                        {this.state.infoOrError.maxUserCount - this.state.infoOrError.userCount}{' '}
                                        remaining
                                    </div>
                                    <a href="https://about.sourcegraph.com/pricing" className="btn btn-primary btn-sm">
                                        Upgrade license
                                    </a>
                                </div>
                            )}
                        </>
                    )}
                </div>
            </>
        )
    }

    private querySourcegraphLicenseInfo(): Observable<GQL.ISourcegraphLicenseInfo> {
        return queryGraphQL(gql`
            query SourcegraphLicenseInfo {
                site {
                    sourcegraphLicense {
                        plan
                        userCount
                        maxUserCount
                        expiresAt
                    }
                }
            }
        `).pipe(
            map(({ data, errors }) => {
                if (!data || !data.site || !data.site.sourcegraphLicense) {
                    throw createAggregateError(errors)
                }
                return data.site.sourcegraphLicense
            })
        )
    }
}
