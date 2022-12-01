import React from 'react'

import { mdiDotsHorizontal } from '@mdi/js'
import * as H from 'history'

import { pluralize } from '@sourcegraph/common'
import { SearchContextMinimalFields } from '@sourcegraph/search'
import { SyntaxHighlightedSearchQuery } from '@sourcegraph/search-ui'
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
}: SearchContextNodeProps) => (
    <tr className={styles.row}>
        <td />
        <td>
            {node.spec}
            {!node.public ? (
                <Badge variant="secondary" className="ml-2" pill={true}>
                    Private
                </Badge>
            ) : null}
            {node.autoDefined ? (
                <Badge variant="outlineSecondary" className="ml-2" pill={true}>
                    Auto
                </Badge>
            ) : null}
        </td>
        <td>
            {node.description ? (
                <div className="text-muted">{node.description}</div>
            ) : node.query ? (
                <small>
                    <SyntaxHighlightedSearchQuery query={node.query} key={node.name} />
                </small>
            ) : null}
        </td>
        <td className="text-muted">
            {node.repositories && node.repositories.length > 0 ? (
                <>
                    {node.repositories.length} {pluralize('repository', node.repositories.length, 'repositories')}
                </>
            ) : node.query ? (
                <>Query based</>
            ) : null}
        </td>
        <td className="text-muted">{node.autoDefined ? null : <Timestamp date={node.updatedAt} noAbout={true} />}</td>
        <td>
            {node.viewerHasAsDefault ? (
                <Badge variant="secondary" className="text-uppercase">
                    Default
                </Badge>
            ) : null}
        </td>
        <td>
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
