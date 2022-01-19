import classNames from 'classnames'
import * as H from 'history'
import * as React from 'react'

import { useWildcardTheme } from '../../../hooks/useWildcardTheme'
import type { LinkProps } from '../Link'

import styles from './AnchorLink.module.scss'

export type AnchorLinkProps = LinkProps & {
    as?: LinkComponent
}

export type LinkComponent = React.FunctionComponent<LinkProps>

export const AnchorLink: React.FunctionComponent<AnchorLinkProps> = React.forwardRef(
    ({ to, as: Component, children, className, ...rest }: AnchorLinkProps, reference) => {
        const { isBranded } = useWildcardTheme()

        const commonProps = {
            ref: reference,
            className: classNames(isBranded && styles.anchorLink, className),
        }

        if (!Component) {
            return (
                <a href={to && typeof to !== 'string' ? H.createPath(to) : to} {...rest} {...commonProps}>
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
