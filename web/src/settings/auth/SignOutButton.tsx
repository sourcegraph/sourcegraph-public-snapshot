import * as React from 'react'
import { events } from '../../tracking/events'

export class SignOutButton extends React.Component {
    public render(): JSX.Element | null {
        // Don't use a <Link /> element here — use an anchor that will break
        // the user out of the single-page app to sign out
        return (
            <a href='/-/sign-out' className='ui-button' onClick={this.logTelemetryOnSignOut}>Sign out</a>
        )
    }

    private logTelemetryOnSignOut(): void {
        events.SignOutClicked.log()
    }
}
