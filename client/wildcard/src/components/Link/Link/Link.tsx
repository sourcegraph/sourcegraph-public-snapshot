import React from 'react'

import * as H from 'history'

import { AnchorLink } from '../AnchorLink'

export interface LinkProps
    extends Pick<
        React.AnchorHTMLAttributes<HTMLAnchorElement>,
        Exclude<keyof React.AnchorHTMLAttributes<HTMLAnchorElement>, 'href'>
    > {
    to: string | H.LocationDescriptor<any>
    ref?: React.Ref<HTMLAnchorElement>
}

/**
 * The component used to render a link. All shared code must use this component for links—not <a>, <Link>, etc.
 *
 * Different platforms (web app vs. browser extension) require the use of different link components:
 *
 * The web app uses <RouterLinkOrAnchor>, which uses react-router-dom's <Link> for relative URLs (for page
 * navigation using the HTML history API) and <a> for absolute URLs. The react-router-dom <Link> component only
 * works inside a react-router <BrowserRouter> context, so it wouldn't work in the browser extension.
 *
 * The browser extension uses <a> for everything (because code hosts don't generally use react-router). A
 * react-router-dom <Link> wouldn't work in the browser extension, because there is no <BrowserRouter>.
 *
 * This variable must be set at initialization time by calling {@link setLinkComponent}.
 *
 * The `to` property holds the destination URL (do not use `href`). If <a> is used, the `to` property value is
 * given as the `href` property value on the <a> element.
 *
 * @see setLinkComponent
 */
export let Link: typeof AnchorLink

if (process.env.NODE_ENV !== 'production') {
    // Fail with helpful message if setLinkComponent has not been called when the <Link> component is used.
    Link = React.forwardRef(() => {
        throw new Error('No Link component set. You must call setLinkComponent to set the Link component to use.')
    }) as typeof Link
}

/**
 * Sets (globally) the component to use for links. This must be set at initialization time.
 *
 * @see Link
 * @see AnchorLink
 */
export function setLinkComponent(component: typeof Link): void {
    Link = component
}
