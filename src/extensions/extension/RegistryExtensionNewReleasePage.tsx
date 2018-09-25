import { PageTitle } from '@sourcegraph/webapp/dist/components/PageTitle'
import { ExtensionAreaRouteContext } from '@sourcegraph/webapp/dist/extensions/extension/ExtensionArea'
import * as React from 'react'
import { Redirect, RouteComponentProps } from 'react-router'

interface Props extends ExtensionAreaRouteContext, RouteComponentProps<{}> {}

/** A page for publishing a new release of an extension to the extension registry. */
export const RegistryExtensionNewReleasePage: React.SFC<Props> = props => {
    // If not logged in, redirect to sign in
    if (!props.authenticatedUser) {
        const newUrl = new URL(window.location.href)
        newUrl.pathname = '/sign-in'
        // Return to the current page after sign up/in.
        newUrl.searchParams.set('returnTo', window.location.href)
        return <Redirect to={newUrl.pathname + newUrl.search} />
    }

    return (
        <div className="registry-extension-new-release-page">
            <PageTitle title="Publish new release" />
            <h2>Publish new release</h2>
            <p>
                Use the{' '}
                <a href="https://github.com/sourcegraph/src-cli" target="_blank">
                    <code>src</code> CLI tool
                </a>{' '}
                to publish a new release:
            </p>
            <pre>
                <code>$ src extensions publish</code>
            </pre>
        </div>
    )
}
