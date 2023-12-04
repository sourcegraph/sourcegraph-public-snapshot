import React from 'react'

import classNames from 'classnames'

import { useWildcardTheme } from '../../../hooks'
import type { Link } from '../Link'

import styles from './AnchorLink.module.scss'

/**
 * Link that doesn't use react-router under the hood.
 * May be used directly and via setLinkComponent outside of Router context.
 *
 * @see setLinkComponent
 */
export const AnchorLink = React.forwardRef(({ to, children, className, ...rest }, reference) => {
    const { isBranded } = useWildcardTheme()

    const commonProps = {
        ref: reference,
        className: classNames(isBranded && styles.anchorLink, className),
    }

    return (
        // eslint-disable-next-line react/forbid-elements
        <a href={to} {...rest} {...commonProps}>
            {children}
        </a>
    )
}) as Link

AnchorLink.displayName = 'AnchorLink'
