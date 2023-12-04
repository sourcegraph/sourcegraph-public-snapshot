import React from 'react'

import classNames from 'classnames'

import { Heading, Icon } from '@sourcegraph/wildcard'

import styles from './SiteAdminPageTitle.module.scss'

interface Props {
    icon: string
}
export const SiteAdminPageTitle: React.FunctionComponent<React.PropsWithChildren<Props>> = ({ icon, children }) => {
    const labels = React.Children.toArray(children)
    return (
        <div className="d-flex flex-column justify-content-between align-items-start">
            <Heading as="h3" styleAs="h2" className="mb-4 mt-2 d-flex align-items-center">
                <Icon className="mr-1" color="var(--link-color)" svgPath={icon} size="sm" aria-hidden="true" />
                {labels.map((label, index) => (
                    <React.Fragment key={index}>
                        {label}
                        {index < labels.length - 1 && <span className={classNames(styles.iconColor, 'mx-2')}>/</span>}
                    </React.Fragment>
                ))}
            </Heading>
        </div>
    )
}
