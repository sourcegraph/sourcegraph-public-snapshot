import classNames from 'classnames'
import * as H from 'history'
import * as React from 'react'

import styles from './AnchorLink.module.scss'

export type LinkProps = {
    to: string | H.LocationDescriptor<any>
    ref?: React.Ref<HTMLAnchorElement>
    as?: LinkComponent
} & Pick<
    React.AnchorHTMLAttributes<HTMLAnchorElement>,
    Exclude<keyof React.AnchorHTMLAttributes<HTMLAnchorElement>, 'href'>
>

export type LinkComponent = React.FunctionComponent<LinkProps>

export const AnchorLink: React.FunctionComponent<LinkProps> = React.forwardRef(
    ({ to, as: Component, children, className, ...rest }: LinkProps, reference) => {
        if (!Component) {
            return (
                <a
                    href={to && typeof to !== 'string' ? H.createPath(to) : to}
                    {...rest}
                    className={classNames(styles.anchorLink, className)}
                    ref={reference}
                >
                    {children}
                </a>
            )
        }

        return (
            <Component to={to} {...rest} className={classNames(styles.anchorLink, className)} ref={reference}>
                {children}
            </Component>
        )
    }
)
