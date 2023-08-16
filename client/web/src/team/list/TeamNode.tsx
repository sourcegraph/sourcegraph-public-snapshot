import React, { useCallback, useState } from 'react'

import { mdiAccount, mdiDelete } from '@mdi/js'
import classNames from 'classnames'

import { pluralize } from '@sourcegraph/common'
import { TeamAvatar } from '@sourcegraph/shared/src/components/TeamAvatar'
import { Button, Link, Icon, Tooltip, useDebounce } from '@sourcegraph/wildcard'

import { Collapsible } from '../../components/Collapsible'
import type { ListTeamFields } from '../../graphql-operations'

import { useChildTeams } from './backend'
import { DeleteTeamModal } from './DeleteTeamModal'
import { TeamList } from './TeamListPage'

import styles from './TeamNode.module.scss'

export interface TeamNodeProps {
    node: ListTeamFields
    refetchAll: () => void
}

type OpenModal = 'delete'

export const TeamNode: React.FunctionComponent<React.PropsWithChildren<TeamNodeProps>> = ({ node, refetchAll }) => {
    const [openModal, setOpenModal] = useState<OpenModal | undefined>()

    const onClickDelete = useCallback<React.MouseEventHandler>(event => {
        event.preventDefault()
        setOpenModal('delete')
    }, [])
    const closeModal = useCallback(() => {
        setOpenModal(undefined)
    }, [])
    const afterAction = useCallback(() => {
        setOpenModal(undefined)
        refetchAll()
    }, [refetchAll])

    return (
        <>
            <li className={classNames(styles.node, 'list-group-item')}>
                {node.childTeams.totalCount > 0 && (
                    <Collapsible
                        title={
                            <NodeContent
                                node={node}
                                onClickDelete={onClickDelete}
                                className="d-flex align-items-center justify-content-between"
                            />
                        }
                        titleClassName="flex-1"
                        wholeTitleClickable={false}
                        defaultExpanded={false}
                    >
                        <ChildTeamList parentTeam={node.name} />
                    </Collapsible>
                )}
                {node.childTeams.totalCount === 0 && (
                    <NodeContent
                        node={node}
                        onClickDelete={onClickDelete}
                        className={classNames(styles.noChildNode, 'd-flex align-items-center justify-content-between')}
                    />
                )}
            </li>

            {openModal === 'delete' && <DeleteTeamModal onCancel={closeModal} afterDelete={afterAction} team={node} />}
        </>
    )
}

interface NodeContentProps {
    node: ListTeamFields
    onClickDelete: React.MouseEventHandler
    className?: string
}

const NodeContent: React.FunctionComponent<NodeContentProps> = ({ node, onClickDelete, className }) => (
    <div className={classNames(className)}>
        <div>
            <TeamAvatar team={node} className="mr-2" />
            <Link to={node.url}>
                <strong>{node.name}</strong>
            </Link>
            <br />
            <span className="text-muted">{node.displayName}</span>
        </div>
        <div>
            <Tooltip content="Team members">
                <Button to={`${node.url}/members`} variant="secondary" size="sm" as={Link}>
                    <Icon aria-hidden={true} svgPath={mdiAccount} /> {node.members.totalCount}{' '}
                    {pluralize('member', node.members.totalCount)}
                </Button>
            </Tooltip>{' '}
            <Tooltip content="Delete team">
                <Button
                    aria-label="Delete team"
                    onClick={onClickDelete}
                    disabled={!node.viewerCanAdminister}
                    variant="danger"
                    size="sm"
                >
                    <Icon aria-hidden={true} svgPath={mdiDelete} />
                </Button>
            </Tooltip>
        </div>
    </div>
)

const ChildTeamList: React.FunctionComponent<{ parentTeam: string }> = ({ parentTeam }) => {
    const [searchValue, setSearchValue] = useState('')
    const query = useDebounce(searchValue, 200)

    const connection = useChildTeams(parentTeam, query)

    return (
        <TeamList
            searchValue={searchValue}
            setSearchValue={setSearchValue}
            query={query}
            {...connection}
            className={classNames(styles.childTeams, 'py-2 pl-3 pr-2')}
        />
    )
}
