import React from 'react'

import classNames from 'classnames'
import * as H from 'history'
import isAbsoluteUrl from 'is-absolute-url'

import { useWildcardTheme } from '@sourcegraph/wildcard'
// eslint-disable-next-line no-restricted-imports
import styles from '@sourcegraph/wildcard/src/components/Link/AnchorLink/AnchorLink.module.scss'

// This is based off the @sourcegraph/wildcard/Link component
// to handle links in VSCE that works differently than in our web app
// because in VSCE history does not work as it is in browser

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
export let Link: React.FunctionComponent<React.PropsWithChildren<LinkProps>> = ({ to, children, ...props }) => (
    // eslint-disable-next-line react/forbid-elements
    <a
        href={checkLink(to && typeof to !== 'string' ? H.createPath(to) : to)}
        id={to && typeof to !== 'string' ? H.createPath(to) : to}
        {...props}
    >
        {children}
    </a>
)

if (process.env.NODE_ENV !== 'production') {
    // Fail with helpful message if setLinkComponent has not been called when the <Link> component is used.
    Link = () => {
        throw new Error('No Link component set. You must call setLinkComponent to set the Link component to use.')
    }
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

export type AnchorLinkProps = LinkProps & {
    as?: LinkComponent
}

export type LinkComponent = React.FunctionComponent<React.PropsWithChildren<LinkProps>>

export const AnchorLink: React.FunctionComponent<React.PropsWithChildren<AnchorLinkProps>> = React.forwardRef(
    ({ to, as: Component, children, className, ...rest }: AnchorLinkProps, reference) => {
        const { isBranded } = useWildcardTheme()

        const commonProps = {
            ref: reference,
            className: classNames(isBranded && styles.anchorLink, className),
        }

        if (!Component) {
            return (
                // eslint-disable-next-line react/forbid-elements
                <a href={checkLink(to && typeof to !== 'string' ? H.createPath(to) : to)} {...rest} {...commonProps}>
                    {children}
                </a>
            )
        }

        return (
            <Component to={to} {...rest} {...commonProps}>
                {children}
            </Component>
        )
    }
)

/**
 * Uses react-router-dom's <Link> for relative URLs, <a> for absolute URLs. This is useful because passing an
 * absolute URL to <Link> will create an (almost certainly invalid) URL where the absolute URL is resolved to the
 * current URL, such as https://example.com/a/b/https://example.com/c/d.
 */
export const RouterLink: React.FunctionComponent<React.PropsWithChildren<AnchorLinkProps>> = React.forwardRef(
    ({ to, children, ...rest }: AnchorLinkProps, reference) => (
        <AnchorLink
            to={checkLink(to && typeof to !== 'string' ? H.createPath(to) : to)}
            as={typeof to === 'string' && isAbsoluteUrl(to) ? undefined : Link}
            {...rest}
            ref={reference}
        >
            {children}
        </AnchorLink>
    )
)

/**
 * Check if link is valid
 * Set invalid links to '#' because VS Code Web opens invalid links in new tabs
 * Invalid links includes links that start with 'sourcegraph://'
 */
function checkLink(uri: string): string {
    // Private instance user are required to provide access token
    // This is for users who has not provide an access token and is using dotcom by default
    if (
        uri.startsWith('/sign-up') ||
        uri.startsWith('/contexts') ||
        uri.startsWith('/code_search/reference/queries') ||
        uri.startsWith('/help')
    ) {
        return `https://sourcegraph.com${uri}?editor=vscode&utm_medium=VSCODE&utm_source=sidebar&utm_campaign=vsce-sign-up&utm_content=sign-up`
    }
    if (uri.startsWith('https://')) {
        return uri
    }
    return '#'
}
