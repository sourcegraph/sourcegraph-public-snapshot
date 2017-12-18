import * as React from 'react'
import { eventLogger } from '../tracking/eventLogger'
import { updateDeploymentConfiguration } from './backend'

/**
 * A page that is shown when the Sourcegraph instance has not yet been initialized.
 * Only the person who first accesses the instance will see this.
 */
export class InitializePage extends React.Component<{}, {}> {
    private emailInput: HTMLInputElement | null = null
    private telemetryInput: HTMLInputElement | null = null

    private onSubmit = () => {
        if (this.emailInput && this.telemetryInput) {
            eventLogger.log('ServerInstallationComplete', {
                server: {
                    email: this.emailInput.value,
                    appId: window.context.trackingAppID,
                    telemetryEnabled: this.telemetryInput.checked,
                },
            })
            updateDeploymentConfiguration(this.emailInput.value, this.telemetryInput.checked).subscribe(
                () => window.location.reload(true),
                error => {
                    console.error(error)
                }
            )
        }
    }

    public render(): JSX.Element {
        return (
            <div className="search-page__onboarding-container">
                <div className="search-page__onboarding-details-container">
                    <div className="search-page__onboarding-details">
                        <div style={{ padding: 25, textAlign: 'left' }}>
                            <img
                                style={{ maxWidth: '90%' }}
                                src={`${window.context.assetsRoot}/img/` + 'ui2/sourcegraph-light-head-logo.svg'}
                            />
                            <form onSubmit={this.onSubmit}>
                                <div style={{ textAlign: 'left' }}>
                                    <h2 style={{ color: 'black', marginBottom: 0, paddingTop: 20 }}>
                                        Welcome to Sourcegraph Server!
                                    </h2>
                                    <div style={{ color: 'black' }}>
                                        Configure your server with an optional admin email address.
                                    </div>
                                </div>
                                <div style={{ paddingTop: '1rem' }}>
                                    <input
                                        ref={e => (this.emailInput = e)}
                                        style={{ width: '100%', padding: 5 }}
                                        placeholder="Admin email (optional)"
                                        type="email"
                                        autoFocus={true}
                                    />
                                </div>
                                <div style={{ margin: '9px 0 15px 0' }}>
                                    <label style={{ color: 'black', paddingLeft: 5 }}>
                                        <input
                                            ref={e => (this.telemetryInput = e)}
                                            defaultChecked={true}
                                            type="checkbox"
                                        />
                                        &nbsp; Send product usage data and check for updates (file contents and names
                                        are never sent)
                                    </label>
                                </div>
                                <div style={{ textAlign: 'right' }}>
                                    <button
                                        style={{ maxWidth: 225 }}
                                        type="submit"
                                        className="btn btn-primary btn-block"
                                    >
                                        Continue
                                    </button>
                                </div>
                            </form>
                        </div>
                    </div>
                </div>
            </div>
        )
    }
}
