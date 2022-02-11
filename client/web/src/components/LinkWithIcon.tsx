import classNames from 'classnames'
import { kebabCase } from 'lodash'
import React from 'react'
import { NavLink, NavLinkProps } from 'react-router-dom'

import { Button } from '@sourcegraph/wildcard'

import styles from './LinkWithIcon.module.scss'

interface LinkWithIconProps extends NavLinkProps {
    text: string
    icon: React.ComponentType<{ className?: string }>
}

/**
 * A link displaying an icon along with text.
 */
export const LinkWithIcon: React.FunctionComponent<LinkWithIconProps> = props => {
    const { to, exact, text, className, activeClassName, icon: Icon } = props

    return (
        <Button
            as={NavLink}
            to={to}
            exact={exact}
            className={classNames('d-flex', 'align-items-center', styles.linkWithIcon, className)}
            activeClassName={activeClassName}
            variant="link"
            data-testid={kebabCase(text)}
        >
            <Icon className="icon-inline mr-1" />
            <span className="inline-block">{text}</span>
        </Button>
    )
}
