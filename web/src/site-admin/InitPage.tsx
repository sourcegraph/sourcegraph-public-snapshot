import * as React from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { SignUpForm } from '../auth/SignUpPage'
import { eventLogger } from '../tracking/eventLogger'
import { updateDeploymentConfiguration } from './backend'

/**
 * A page that is shown when the Sourcegraph instance has not yet been initialized.
 * Only the person who first accesses the instance will see this.
 */
export class InitPage extends React.Component<RouteComponentProps<any>, {}> {
    public render(): JSX.Element {
        if (!window.context.onPrem || !window.context.showOnboarding) {
            return <Redirect to="/search" />
        }

        return (
            <div className="init-page theme-light">
                <div className="init-page__content">
                    <img
                        className="init-page__logo"
                        src={`${window.context.assetsRoot}/img/ui2/sourcegraph-light-head-logo.svg`}
                    />
                    <h2 className="init-page__header">Welcome to Sourcegraph Server!</h2>
                    <p>Create an admin account to start using Sourcegraph.</p>
                    <SignUpForm
                        autoFocus={true}
                        buttonLabel="Create admin account & continue"
                        onDidSignUp={this.onDidSignUp}
                        location={this.props.location}
                        history={this.props.history}
                    />
                </div>
            </div>
        )
    }

    private onDidSignUp = (email: string) => {
        updateDeploymentConfiguration(email, true).subscribe(
            () => window.location.replace('/'),
            error => {
                console.error(error)
            }
        )
        eventLogger.log('ServerInstallationComplete', {
            server: {
                email,
                appId: window.context.trackingAppID,
                telemetryEnabled: true,
            },
        })
    }
}
