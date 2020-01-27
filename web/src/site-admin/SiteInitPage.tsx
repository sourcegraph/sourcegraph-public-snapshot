import * as React from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import * as GQL from '../../../shared/src/graphql/schema'
import { SignUpArgs, SignUpForm } from '../auth/SignUpForm'
import { submitTrialRequest } from '../marketing/backend'
import { BrandLogo } from '../components/branding/BrandLogo'
import { ThemeProps } from '../../../shared/src/theme'

interface Props extends RouteComponentProps<{}>, ThemeProps {
    authenticatedUser: GQL.IUser | null
}

/**
 * A page that is shown when the Sourcegraph instance has not yet been initialized.
 * Only the person who first accesses the instance will see this.
 */
export class SiteInitPage extends React.Component<Props> {
    public render(): JSX.Element {
        if (!window.context.showOnboarding) {
            return <Redirect to="/search" />
        }

        return (
            <div className="site-init-page">
                <div className="site-init-page__content">
                    <BrandLogo className="site-init-page__logo" isLightTheme={this.props.isLightTheme} />
                    {this.props.authenticatedUser ? (
                        // If there's already a user but the site is not initialized, then the we're in an
                        // unexpected state, likely because of a previous bug or because someone manually modified
                        // the site_config DB table.
                        <p>
                            You're signed in as <strong>{this.props.authenticatedUser.username}</strong>. A site admin
                            must initialize Sourcegraph before you can continue.
                        </p>
                    ) : (
                        <>
                            <h2 className="site-init-page__header">Welcome</h2>
                            <p>Create an admin account to start using Sourcegraph.</p>
                            <SignUpForm
                                buttonLabel="Create admin account & continue"
                                doSignUp={this.doSiteInit}
                                location={this.props.location}
                                history={this.props.history}
                            />
                        </>
                    )}
                </div>
            </div>
        )
    }

    private doSiteInit = (args: SignUpArgs): Promise<void> =>
        fetch('/-/site-init', {
            credentials: 'same-origin',
            method: 'POST',
            headers: {
                ...window.context.xhrHeaders,
                Accept: 'application/json',
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(args),
        }).then(resp => {
            if (resp.status !== 200) {
                return resp.text().then(text => Promise.reject(new Error(text)))
            }
            if (args.requestedTrial) {
                submitTrialRequest(args.email)
            }

            window.location.replace('/site-admin')
            return Promise.resolve()
        })
}
