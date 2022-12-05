import React from 'react'

import { mdiDotsHorizontal } from '@mdi/js'
import classNames from 'classnames'
import * as H from 'history'

import { pluralize } from '@sourcegraph/common'
import { SearchContextMinimalFields } from '@sourcegraph/search'
import { Badge, Icon, Link, Menu, MenuButton, MenuLink, MenuList, Tooltip } from '@sourcegraph/wildcard'

import { Timestamp } from '../../components/time/Timestamp'

import styles from './SearchContextNode.module.scss'

export interface SearchContextNodeProps {
    node: SearchContextMinimalFields
    location: H.Location
    history: H.History
}

export const SearchContextNode: React.FunctionComponent<React.PropsWithChildren<SearchContextNodeProps>> = ({
    node,
}: SearchContextNodeProps) => {
    const contents =
        node.repositories && node.repositories.length > 0 ? (
            <>
                {node.repositories.length} {pluralize('repository', node.repositories.length, 'repositories')}
                &nbsp;
            </>
        ) : node.query ? (
            <>Query based&nbsp;</>
        ) : null

    const tags = (
        <>
            {!node.public ? (
                <Badge variant="secondary" pill={true}>
                    Private
                </Badge>
            ) : null}{' '}
            {node.autoDefined ? (
                <Badge variant="outlineSecondary" pill={true}>
                    Auto
                </Badge>
            ) : null}
        </>
    )

    const timestamp = node.autoDefined ? null : (
        <>
            <span className="d-md-none" aria-hidden={true}>
                Updated{' '}
            </span>{' '}
            <Timestamp date={node.updatedAt} noAbout={true} />
        </>
    )

    return (
        <tr className={styles.row}>
            <td className={styles.star} />
            <td className={styles.name}>
                <Link to={`/contexts/${node.spec}`}>{node.spec}</Link>{' '}
                <span className="d-none d-md-inline-block">{tags}</span>
            </td>
            <td className={styles.description}>
                {node.description ? <div className="text-muted">{node.description}</div> : null}
            </td>
            <td className={classNames(styles.contents, 'text-muted')}>{contents}</td>
            <td className={classNames(styles.lastUpdated, 'text-muted')}>
                {contents && timestamp && (
                    <span className="d-md-none" aria-hidden={true}>
                        {' '}
                        •{' '}
                    </span>
                )}
                {timestamp}
            </td>
            <td className={styles.tags}>
                {node.viewerHasAsDefault ? (
                    <Badge variant="secondary" className="text-uppercase">
                        Default
                    </Badge>
                ) : null}
                <span className="d-md-none">{tags}</span>
            </td>
            <td className={styles.actions}>
                <Menu>
                    <MenuButton variant="icon" className={styles.button}>
                        <Icon svgPath={mdiDotsHorizontal} aria-label="Actions" />
                    </MenuButton>
                    <MenuList>
                        <Tooltip
                            content={
                                node.autoDefined
                                    ? "Auto-defined contexts can't be edited."
                                    : !node.viewerCanManage
                                    ? "You don't have permissions to edit this context."
                                    : undefined
                            }
                        >
                            <MenuLink as={Link} to={`/contexts/${node.spec}/edit`} disabled={!node.viewerCanManage}>
                                Edit...
                            </MenuLink>
                        </Tooltip>
                    </MenuList>
                </Menu>
            </td>
        </tr>
    )
}
