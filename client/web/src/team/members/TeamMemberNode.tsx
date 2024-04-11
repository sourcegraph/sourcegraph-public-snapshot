import React, { useCallback, useState } from 'react'

import { mdiDelete } from '@mdi/js'
import classNames from 'classnames'

import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Button, Link, Icon, Tooltip } from '@sourcegraph/wildcard'

import type { ListTeamMemberFields, Scalars } from '../../graphql-operations'

import { RemoveTeamMemberModal } from './RemoveTeamMemberModal'

import styles from './TeamMemberNode.module.scss'

export interface TeamMemberNodeProps extends TelemetryV2Props {
    viewerCanAdminister: boolean
    teamID: Scalars['ID']
    teamName: string
    node: ListTeamMemberFields
    refetchAll: () => void
}

type OpenModal = 'remove'

export const TeamMemberNode: React.FunctionComponent<React.PropsWithChildren<TeamMemberNodeProps>> = ({
    viewerCanAdminister,
    teamID,
    teamName,
    node,
    refetchAll,
    telemetryRecorder,
}) => {
    const [openModal, setOpenModal] = useState<OpenModal | undefined>()

    const onClickRemove = useCallback<React.MouseEventHandler>(event => {
        event.preventDefault()
        setOpenModal('remove')
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
                <div className="d-flex align-items-center justify-content-between">
                    <div>
                        <Link to={node.url}>
                            <strong>{node.username}</strong>
                        </Link>
                        <br />
                        <span className="text-muted">{node.displayName}</span>
                    </div>
                    <div>
                        <Tooltip content="Remove from team">
                            <Button
                                aria-label="Remove from team"
                                onClick={onClickRemove}
                                disabled={!viewerCanAdminister}
                                variant="danger"
                                size="sm"
                            >
                                <Icon aria-hidden={true} svgPath={mdiDelete} />
                            </Button>
                        </Tooltip>
                    </div>
                </div>
            </li>

            {openModal === 'remove' && (
                <RemoveTeamMemberModal
                    onCancel={closeModal}
                    afterRemove={afterAction}
                    teamID={teamID}
                    teamName={teamName}
                    member={node}
                    telemetryRecorder={telemetryRecorder}
                />
            )}
        </>
    )
}
