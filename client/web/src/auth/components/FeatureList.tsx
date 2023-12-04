import React from 'react'

import classNames from 'classnames'

import { Icon, H4, Text } from '@sourcegraph/wildcard'

import styles from './FeatureList.module.scss'

const Item: React.FC<React.PropsWithChildren<{ title: string; icon: string }>> = ({ title, children, icon }) => (
    <li className="d-flex align-items-start">
        <div className={classNames('p-2 rounded mr-3 d-flex', styles.iconContainer)}>
            <Icon svgPath={icon} aria-label={title} />
        </div>
        <div className="flex-grow-1 d-flex flex-column">
            <H4 className="m-0 p-0">{title}</H4>
            <Text className={classNames('text-muted', styles.description)}>{children}</Text>
        </div>
    </li>
)

export const List: React.FC<React.PropsWithChildren<{}>> = ({ children }) => (
    <ul className="list-unstyled">{children}</ul>
)

export const FeatureList = Object.assign(List, { Item })
