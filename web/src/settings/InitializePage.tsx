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
            <div className="initialize-page theme-light">
                <div className="initialize-page__content">
                    <img
                        className="initialize-page__logo"
                        src={`${window.context.assetsRoot}/img/ui2/sourcegraph-light-head-logo.svg`}
                    />
                    <form onSubmit={this.onSubmit}>
                        <h2 className="initialize-page__header">Welcome to Sourcegraph Server!</h2>
                        <input
                            className="form-control initialize-page__input-email initialize-page__control"
                            ref={e => (this.emailInput = e)}
                            placeholder="Admin email (optional)"
                            type="email"
                            autoFocus={true}
                        />
                        <label className="initialize-page__label initialize-page__control">
                            <input
                                className="initialize-page__input-telemetry"
                                ref={e => (this.telemetryInput = e)}
                                defaultChecked={true}
                                type="checkbox"
                            />
                            <small>
                                Send product usage data to Sourcegraph (file contents and names are never sent)
                            </small>
                        </label>
                        <button type="submit" className="btn btn-primary btn-block">
                            Continue
                        </button>
                    </form>
                </div>
            </div>
        )
    }
}
