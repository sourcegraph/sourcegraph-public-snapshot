import React from 'react'

import { mdiCogOutline } from '@mdi/js'

import { Button, Icon, Text, Modal, H3, Form } from '@sourcegraph/wildcard'

interface RoleAssignmentModalProps {
    onCancel: () => void
}

export const RoleAssignmentModal:  React.FunctionComponent<React.PropsWithChildren<RoleAssignmentModalProps>> = ({
    onCancel
}) => {
    const labelID = 'RoleAssignment'

    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelID}>
            <div className="d-flex align-items-center mb-2">
                <Icon className="icon mr-1" svgPath={mdiCogOutline} inline={false} aria-hidden={true} />{' '}
                <H3 id={labelID} className="mb-0">
                    Assign roles
                </H3>
            </div>
        </Modal>
    )
}
