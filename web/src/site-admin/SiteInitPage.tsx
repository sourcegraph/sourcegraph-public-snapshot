import * as React from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { SignUpArgs, SignUpForm } from '../auth/SignUpPage'
import { eventLogger } from '../tracking/eventLogger'

/**
 * A page that is shown when the Sourcegraph instance has not yet been initialized.
 * Only the person who first accesses the instance will see this.
 */
export class SiteInitPage extends React.Component<RouteComponentProps<any>, {}> {
    public render(): JSX.Element {
        if (!window.context.showOnboarding) {
            return <Redirect to="/search" />
        }

        return (
            <div className="site-init-page theme-light">
                <div className="site-init-page__content">
                    <img
                        className="site-init-page__logo"
                        src={`${window.context.assetsRoot}/img/sourcegraph-light-head-logo.svg`}
                    />
                    <h2 className="site-init-page__header">Welcome to Sourcegraph Server!</h2>
                    <p>Create an admin account to start using Sourcegraph.</p>
                    <SignUpForm
                        buttonLabel="Create admin account & continue"
                        doSignUp={this.doSignUp}
                        location={this.props.location}
                        history={this.props.history}
                    />
                </div>
            </div>
        )
    }

    private doSignUp = (args: SignUpArgs): Promise<void> =>
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

            eventLogger.log('ServerInstallationComplete', {
                server: {
                    email: args.email,
                    appId: window.context.siteID,
                },
            })

            window.location.replace('/site-admin')
            return Promise.resolve()
        })
}
