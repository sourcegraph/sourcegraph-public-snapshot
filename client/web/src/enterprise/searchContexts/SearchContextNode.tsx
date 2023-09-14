import React, { useCallback } from 'react'

import { mdiDotsHorizontal } from '@mdi/js'
import classNames from 'classnames'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { isErrorLike, pluralize } from '@sourcegraph/common'
import type { SearchContextMinimalFields } from '@sourcegraph/shared/src/graphql-operations'
import { Badge, Icon, Link, Menu, MenuButton, MenuItem, MenuLink, MenuList, Tooltip } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'

import { useToggleSearchContextStar } from './hooks/useToggleSearchContextStar'
import { SearchContextStarButton } from './SearchContextStarButton'

import styles from './SearchContextNode.module.scss'

export interface SearchContextNodeProps {
    node: SearchContextMinimalFields
    authenticatedUser: Pick<AuthenticatedUser, 'id'> | null
    setAlert: (message: string) => void
    defaultContext: string | undefined
    setAsDefault: (searchContextId: string, userId?: string) => Promise<void>
}

export const SearchContextNode: React.FunctionComponent<React.PropsWithChildren<SearchContextNodeProps>> = ({
    node,
    authenticatedUser,
    setAlert,
    defaultContext,
    setAsDefault,
}: SearchContextNodeProps) => {
    const { starred, toggleStar } = useToggleSearchContextStar(node.viewerHasStarred, node.id, authenticatedUser?.id)
    const toggleStarWithErrorHandling = useCallback(() => {
        setAlert('') // Clear previous alerts
        toggleStar().catch(error => {
            if (isErrorLike(error)) {
                setAlert(error.message)
            }
        })
    }, [setAlert, toggleStar])

    // Auto-defined search contexts cannot be starred
    const showStarButton = !node.autoDefined && authenticatedUser

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

    const isDefault = defaultContext === node.id

    return (
        <tr className={styles.row}>
            <td className={styles.star}>
                <SearchContextStarButton
                    starred={starred}
                    onClick={toggleStarWithErrorHandling}
                    className={classNames(!showStarButton && 'invisible')} // Render invisible button to keep table layout consistent
                />
            </td>
            <td className={styles.name}>
                <Link to={`/contexts/${encodeURIComponent(node.spec)}`} className={styles.spec}>
                    {node.spec}
                </Link>{' '}
                <span className="d-none d-md-inline-block">{tags}</span>
            </td>
            <td className={classNames(styles.description, 'text-muted')}>{node.description}</td>
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
                {isDefault ? (
                    <Badge variant="secondary" className="text-uppercase" data-testid="search-context-default-badge">
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
                                isDefault
                                    ? 'This is already your default context.'
                                    : !authenticatedUser
                                    ? 'You must be signed in to use a context as default.'
                                    : undefined
                            }
                        >
                            <MenuItem
                                disabled={isDefault || !authenticatedUser}
                                onSelect={() => setAsDefault(node.id, authenticatedUser?.id)}
                            >
                                Use as default
                            </MenuItem>
                        </Tooltip>
                        <Tooltip
                            content={
                                node.autoDefined
                                    ? "Auto-defined contexts can't be edited."
                                    : !node.viewerCanManage
                                    ? "You don't have permissions to edit this context."
                                    : undefined
                            }
                        >
                            <MenuLink
                                as={Link}
                                to={`/contexts/${encodeURIComponent(node.spec)}/edit`}
                                disabled={!node.viewerCanManage}
                            >
                                Edit...
                            </MenuLink>
                        </Tooltip>
                    </MenuList>
                </Menu>
            </td>
        </tr>
    )
}
