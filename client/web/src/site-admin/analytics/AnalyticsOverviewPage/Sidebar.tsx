import React from 'react'

import { mdiArrowRight } from '@mdi/js'
import classNames from 'classnames'

import { Text, H4, Icon, Link, Tooltip } from '@sourcegraph/wildcard'

import { formatNumber } from '../utils'

import styles from './index.module.scss'

interface IProps {
    sections: {
        title: string
        link: string
        icon: string
        items: { label: string; value: number }[]
    }[]
}

export const Sidebar = React.memo(function OverviewSidebar({ sections }: IProps) {
    return (
        <>
            {sections.map(section => (
                <div key={section.title} className="d-flex flex-column mb-4">
                    <Link to={section.link} className="text-decoration-none">
                        <div
                            className={classNames(
                                'd-flex align-items-center justify-content-between pb-2 mb-3 cursor-pointer text-body',
                                styles.border
                            )}
                        >
                            <div className="d-flex align-items-center">
                                <Icon aria-label={section.title} svgPath={section.icon} size="md" className="mr-2" />
                                <H4 className="mb-0">{section.title}</H4>
                            </div>
                            <Icon aria-label="link" svgPath={mdiArrowRight} size="md" className={styles.link} />
                        </div>
                    </Link>
                    <table className={styles.sidebarTable}>
                        <tbody>
                            {section.items.map(item => (
                                <tr key={item.label} className="py-2">
                                    <Text as="td" className="text-muted py-1" weight="bold">
                                        <Tooltip
                                            content={
                                                formatNumber(item.value) !== `${item.value}`
                                                    ? `${item.value}`
                                                    : undefined
                                            }
                                        >
                                            <span>{formatNumber(item.value)}</span>
                                        </Tooltip>
                                    </Text>
                                    <Text as="td" className="py-1">
                                        {item.label}
                                    </Text>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            ))}
        </>
    )
})
