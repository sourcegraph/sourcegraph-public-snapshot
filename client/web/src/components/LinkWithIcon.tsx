import classNames from 'classnames'
import { kebabCase } from 'lodash'
import React from 'react'
import { NavLink, NavLinkProps } from 'react-router-dom'

import { Button, Icon, AccessibleSvg } from '@sourcegraph/wildcard'

import styles from './LinkWithIcon.module.scss'

interface LinkWithIconProps extends NavLinkProps {
    text: string
    icon: AccessibleSvg
}

/**
 * A link displaying an icon along with text.
 */
export const LinkWithIcon: React.FunctionComponent<LinkWithIconProps> = props => {
    const { to, exact, text, className, activeClassName, icon: linkIcon } = props

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
            <Icon className="mr-1" as={linkIcon} aria-hidden="true" />
            <span className="inline-block">{text}</span>
        </Button>
    )
}
