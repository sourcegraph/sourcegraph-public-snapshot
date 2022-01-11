import * as React from 'react'
import { __RouterContext } from 'react-router'
import { Link } from 'react-router-dom'

import { AnchorLink, LinkProps } from '../AnchorLink'

function useInRouterContext(): boolean {
    return Boolean(React.useContext(__RouterContext))
}

export const RouterLink: React.FunctionComponent<LinkProps> = React.forwardRef(
    ({ to, children, ...rest }: LinkProps, reference) => {
        const isInRouter = useInRouterContext()

        if (process.env.NODE_ENV === 'development' && !isInRouter) {
            throw new Error('Please use the `AnchorLink` component outside of `react-router`')
        }

        return (
            <AnchorLink to={to} as={Link} {...rest} ref={reference}>
                {children}
            </AnchorLink>
        )
    }
)
