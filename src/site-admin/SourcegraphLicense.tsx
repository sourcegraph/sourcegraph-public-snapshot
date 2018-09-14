import { upperFirst } from 'lodash'
import * as React from 'react'
import { Observable, Subscription } from 'rxjs'
import { catchError, map } from 'rxjs/operators'
import { gql, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../util/errors'
import { pluralize } from '../util/strings'

interface Props {
    className?: string
}

interface State {
    /** The Sourcegraph license info, or an error, or undefined while loading. */
    licenseOrError?: GQL.ISourcegraphLicense | ErrorLike
}

/**
 * A component displaying information about and the status of the Sourcegraph software license.
 */
export class SourcegraphLicense extends React.Component<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.querySourcegraphLicense()
                .pipe(
                    catchError(err => [asError(err)]),
                    map(v => ({ licenseOrError: v }))
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), err => console.error(err))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (this.state.licenseOrError === undefined) {
            return null
        }

        return (
            <>
                <div className={`sourcegraph-license card ${this.props.className}`}>
                    <div className="sourcegraph-license__bg" />
                    <div className="card-header d-flex align-items-center">
                        <img
                            className="sourcegraph-license__logo mr-1"
                            src={`${window.context.assetsRoot}/img/sourcegraph-mark.svg`}
                        />
                        {isErrorLike(this.state.licenseOrError) ? 'Sourcegraph' : this.state.licenseOrError.productName}
                    </div>
                    {isErrorLike(this.state.licenseOrError) ? (
                        <div className="card-body">
                            <div className="alert alert-danger">
                                Error querying for license information: {upperFirst(this.state.licenseOrError.message)}
                            </div>
                        </div>
                    ) : (
                        <>
                            <div className="list-group list-group-flush">
                                <div className="list-group-item py-2">
                                    <strong>Site ID:</strong>{' '}
                                    <span className="sourcegraph-license__site-id">
                                        {this.state.licenseOrError.siteID}
                                    </span>
                                </div>
                                <div className="list-group-item py-2">
                                    <strong>Primary site admin:</strong>{' '}
                                    {this.state.licenseOrError.primarySiteAdminEmail}
                                </div>
                            </div>
                            <div className="card-body mt-3">
                                <p className="card-text mb-0">
                                    {this.state.licenseOrError.premiumFeatures.some(f => f.enabled) ? (
                                        <>
                                            Premium features are in use for {this.state.licenseOrError.userCount}{' '}
                                            {pluralize('user', this.state.licenseOrError.userCount)}:
                                        </>
                                    ) : (
                                        <>Premium features are not in use:</>
                                    )}
                                </p>
                            </div>
                            <div className="list-group list-group-flush">
                                {this.state.licenseOrError.premiumFeatures.map((f, i) => (
                                    <div key={i} className="list-group-item py-2">
                                        <div className="card-title d-flex align-items-center">
                                            <strong className="mr-2">{f.title}</strong>
                                            {f.enabled ? (
                                                <span className="badge badge-success">ON</span>
                                            ) : (
                                                <span className="badge badge-secondary">OFF</span>
                                            )}
                                        </div>
                                        <div className="card-subtitle">
                                            {f.description} &mdash;{' '}
                                            <a href="https://about.sourcegraph.com/pricing" target="_blank">
                                                pricing
                                            </a>,{' '}
                                            <a href="https://about.sourcegraph.com/contact/sales" target="_blank">
                                                sales
                                            </a>, &amp;{' '}
                                            <a href={f.informationURL} target="_blank">
                                                more information
                                            </a>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </>
                    )}
                </div>
            </>
        )
    }

    private querySourcegraphLicense(): Observable<GQL.ISourcegraphLicense> {
        return queryGraphQL(gql`
            query SourcegraphLicense {
                site {
                    sourcegraphLicense {
                        siteID
                        primarySiteAdminEmail
                        userCount
                        productName
                        premiumFeatures {
                            title
                            description
                            enabled
                            informationURL
                        }
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
