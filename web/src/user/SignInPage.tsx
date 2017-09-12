import * as React from 'react'
import { events } from 'sourcegraph/tracking/events'
import { sourcegraphContext } from 'sourcegraph/util/sourcegraphContext'

interface Props {
    showEditorFlow: boolean
}

/**
 * A landing page for the user to sign in or register, if not authed
 */
export class SignInPage extends React.Component<Props> {
    public render(): JSX.Element | null {
        // TODO(Dan): don't just use '/' on non-editor sign ins
        const returnTo = this.props.showEditorFlow ? '/editor-auth' : '/'
        const url = `/-/github-oauth/initiate?return-to=${returnTo}`
        return (
            <form method='POST' action={url} onSubmit={this.logTelemetryOnSubmit}>
                <input type='hidden' name='gorilla.csrf.Token' value={sourcegraphContext.csrfToken} />
                <input type='submit' value='Sign in with GitHub' />
            </form>
        )
    }

    private logTelemetryOnSubmit(): void {
        events.InitiateGitHubOAuth2Flow.log()
    }
}
