import * as React from 'react'
import { Link, LinkProps } from 'react-router-dom'

/**
 * Uses react-router-dom's <Link> for relative URLs, <a> for absolute URLs. This is useful because passing an
 * absolute URL to <Link> will create an (almost certainly invalid) URL where the absolute URL is resolved to the
 * current URL, such as https://sourcegraph.com/a/b/https://example.com/c/d.
 */
export const RouterLinkOrAnchor: React.SFC<LinkProps> = props =>
    typeof props.to === 'string' && /^https?:\/\//.test(props.to) ? (
        <a href={props.to} {...props} />
    ) : (
        <Link {...props} />
    )
