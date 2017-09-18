import * as React from 'react'
import { events } from 'sourcegraph/tracking/events'
import { sourcegraphContext } from 'sourcegraph/util/sourcegraphContext'

interface Props {
    showEditorFlow: boolean
}

const newAuth = true

/**
 * A landing page for the user to sign in or register, if not authed
 */
export class SignInPage extends React.Component<Props> {
    public render(): JSX.Element | null {
        if (!newAuth) { // TODO remove after Sept 20 auth switchover
            // TODO(Dan): don't just use '/' on non-editor sign ins
            // tslint:disable-next-line
            const returnTo = this.props.showEditorFlow ? '/editor-auth' : '/'
            const url = `/-/github-oauth/initiate?return-to=${returnTo}`
            return (
                <form method='POST' action={url} onSubmit={this.logTelemetryOnSubmit}>
                    <input type='hidden' name='gorilla.csrf.Token' value={sourcegraphContext.csrfToken} />
                    <input type='submit' value='Sign in with GitHub' />
                </form>
            )
        }
        return (
            <div>
                <a href='/-/sign-in' onClick={this.logTelemetryOnSubmit}>
                    <input type='submit' value='Sign in' />
                </a>
                <a href='/-/sign-out'>
                    <input type='submit' value='Sign out' />
                </a>
            </div>
        )
    }

    private logTelemetryOnSubmit(): void {
        events.InitiateGitHubOAuth2Flow.log() // TODO update this
    }
}
