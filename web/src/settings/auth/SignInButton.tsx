import * as React from 'react'
import { events } from '../../tracking/events'

export class SignInButton extends React.Component {
    public render(): JSX.Element | null {
        // Don't use a <Link /> element here — use an anchor that will break
        // the user out of the single-page app to sign in
        return (
            <a href='/-/sign-in' className='ui-button' onClick={this.logTelemetryOnSignIn}>Sign in</a>
        )
    }

    private logTelemetryOnSignIn(): void {
        events.InitiateAuth0SignIn.log()
    }
}
