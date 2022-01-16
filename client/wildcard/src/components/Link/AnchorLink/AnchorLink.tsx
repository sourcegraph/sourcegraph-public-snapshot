import classNames from 'classnames'
import * as H from 'history'
import * as React from 'react'

import type { LinkProps } from '../Link'

import styles from './AnchorLink.module.scss'

export type AnchorLinkProps = LinkProps & {
    as?: LinkComponent
}

export type LinkComponent = React.FunctionComponent<LinkProps>

export const AnchorLink: React.FunctionComponent<AnchorLinkProps> = React.forwardRef(
    ({ to, as: Component, children, branded = true, className, ...rest }: AnchorLinkProps, reference) => {
        const brandedClassName = branded && styles.anchorLink

        if (!Component) {
            return (
                <a
                    href={to && typeof to !== 'string' ? H.createPath(to) : to}
                    {...rest}
                    className={classNames(brandedClassName, className)}
                    ref={reference}
                >
                    {children}
                </a>
            )
        }

        return (
            <Component to={to} {...rest} className={classNames(brandedClassName, className)} ref={reference}>
                {children}
            </Component>
        )
    }
)
