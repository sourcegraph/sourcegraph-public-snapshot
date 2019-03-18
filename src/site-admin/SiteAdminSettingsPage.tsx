import { upperFirst } from 'lodash'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable, Subscription } from 'rxjs'
import { catchError, map } from 'rxjs/operators'
import { gql, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { PageTitle } from '../components/PageTitle'
import { ExtensionsProps } from '../extensions/ExtensionsClientCommonContext'
import { SettingsArea } from '../settings/SettingsArea'
import { createAggregateError, ErrorLike, isErrorLike } from '../util/errors'

function querySiteConfigDeprecatedSettings(): Observable<string | null> {
    return queryGraphQL(gql`
        query SiteConfigDeprecatedSettings {
            site {
                deprecatedSiteConfigurationSettings
            }
        }
    `).pipe(
        map(({ data, errors }) => {
            if (!data || !data.site) {
                throw createAggregateError(errors)
            }
            return data.site.deprecatedSiteConfigurationSettings
        })
    )
}

interface Props extends RouteComponentProps<{}>, ExtensionsProps {
    authenticatedUser: GQL.IUser
    isLightTheme: boolean
    site: Pick<GQL.ISite, '__typename' | 'id'>
}

interface State {
    /** The deprecated settings from the site config "settings" field, undefined while loading, or an error. */
    deprecatedSettingsOrError?: string | null | ErrorLike
}

export class SiteAdminSettingsPage extends React.Component<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            querySiteConfigDeprecatedSettings()
                .pipe(
                    catchError(error => [error]),
                    map(c => ({ deprecatedSettingsOrError: c }))
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), err => console.error(err))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <>
                <PageTitle title="Site settings" />
                <SettingsArea
                    {...this.props}
                    subject={this.props.site}
                    authenticatedUser={this.props.authenticatedUser}
                    extraHeader={
                        <>
                            <p>
                                Global settings apply to all organizations and users. Settings for a user or
                                organization override global settings.
                            </p>
                            {this.state.deprecatedSettingsOrError !== undefined &&
                                this.state.deprecatedSettingsOrError !== null &&
                                (isErrorLike(this.state.deprecatedSettingsOrError) ? (
                                    <div className="alert alert-danger my-2">
                                        {upperFirst(this.state.deprecatedSettingsOrError.message)}
                                    </div>
                                ) : (
                                    <div className="alert alert-warning my-2">
                                        <p>
                                            Your <Link to="/site-admin/configuration">site configuration</Link> contains
                                            a <strong>deprecated</strong> <code>settings</code> field (contents below).
                                            Support for providing settings in the site configuration file will be
                                            removed in a future Sourcegraph release.
                                        </p>
                                        <p>
                                            To fix this problem: Add these settings to the editable global settings
                                            below, and remove them from{' '}
                                            <Link to="/site-admin/configuration">site configuration</Link>.
                                        </p>
                                        <pre className="form-control">
                                            <code>{this.state.deprecatedSettingsOrError}</code>
                                        </pre>
                                    </div>
                                ))}
                        </>
                    }
                />
            </>
        )
    }
}
