import * as React from 'react'
import * as GQL from '../backend/graphqlschema'

// Production extension Url
const extensionUrl = 'chrome-extension://dgjhfomjieaadpoljlnidmbgkdffpack'

// Dev extension url
// const extensionUrl = 'chrome-extension://bmfbcejdknlknpncfpeloejonjoledha'

interface Props {
    authenticatedUser: GQL.IUser
}

/**
 * Returns the URL used to link the browser extension to a Sourcegraph instance.
 *
 * @param email the email address of the authenticated user.
 */
const getExtensionLinkURL = (email: string): string => {
    const url = new URL(`${extensionUrl}/link.html`)
    url.searchParams.set('sourceurl', location.origin)
    url.searchParams.set('userId', email)
    return url.toString()
}

/**
 * Embeds an iframe responsible for passing the current user and Sourcegraph URL to the browser
 * extension.
 */
export const LinkExtension: React.SFC<Props> = props => (
    <iframe className="link-extension" src={getExtensionLinkURL(props.authenticatedUser.email)} />
)
