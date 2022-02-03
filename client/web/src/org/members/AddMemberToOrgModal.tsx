import { gql, useMutation } from '@apollo/client'
import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'
import { debounce } from 'lodash'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { useCallback, useState } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Alert, Button, Input, Modal } from '@sourcegraph/wildcard'

import { AddUserToOrganizationResult, AddUserToOrganizationVariables } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

import styles from './InviteMemberModal.module.scss'

const ADD_USERNAME_OR_EMAIL_TO_ORG = gql`
    mutation AddUserToOrganization($organization: ID!, $username: String!) {
        addUserToOrganization(organization: $organization, username: $username) {
            alwaysNil
        }
    }
`

export interface AddMemberToOrgModalProps {
    orgName: string
    orgId: string
    onMemberAdded: (username: string) => void
}

export const AddMemberToOrgModal: React.FunctionComponent<AddMemberToOrgModalProps> = props => {
    const { orgName, orgId, onMemberAdded } = props

    const [username, setUsername] = useState('')
    const [modalOpened, setModalOpened] = React.useState(false)
    const title = `Add teammate to ${orgName}`

    const onAddUserClick = useCallback(() => {
        setModalOpened(true)
    }, [setModalOpened])

    const onCloseAddUserModal = useCallback(() => {
        setModalOpened(false)
    }, [setModalOpened])

    const [addUserToOrganization, { loading, error }] = useMutation<
        AddUserToOrganizationResult,
        AddUserToOrganizationVariables
    >(ADD_USERNAME_OR_EMAIL_TO_ORG)

    const onUsernameChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setUsername(event.currentTarget.value)
    }, [])

    const addUser = useCallback(async () => {
        if (!username) {
            return
        }

        eventLogger.log('AddOrgMemberClicked')
        try {
            await addUserToOrganization({ variables: { organization: orgId, username } })
            onMemberAdded(username)
            setUsername('')
            setModalOpened(false)
            eventLogger.log('OrgMemberAdded')
        } catch {
            eventLogger.log('AddOrgMemberFailed')
        }
    }, [username, orgId, onMemberAdded, addUserToOrganization])

    const debounceAddUser = debounce(addUser, 500, { leading: true })

    return (
        <>
            <Button variant="primary" onClick={onAddUserClick} className="mr-1">
                + Add member
            </Button>
            {modalOpened && (
                <Modal className={styles.modal} onDismiss={onCloseAddUserModal} position="center" aria-label={title}>
                    <Button className={classNames('btn-icon', styles.closeButton)} onClick={onCloseAddUserModal}>
                        <VisuallyHidden>Close</VisuallyHidden>
                        <CloseIcon />
                    </Button>

                    <h2>{title}</h2>
                    {error && <ErrorAlert className={styles.alert} error={error} />}
                    <div className="d-flex flex-row position-relative">
                        <Input
                            autoFocus={true}
                            value={username}
                            label="Add member by username"
                            title="Add member by username"
                            onChange={onUsernameChange}
                            status={loading ? 'loading' : error ? 'error' : undefined}
                            placeholder="Username"
                        />
                    </div>
                    <div className="d-flex justify-content-end mt-4">
                        <Button
                            type="button"
                            className="mr-2"
                            variant="primary"
                            onClick={debounceAddUser}
                            disabled={loading}
                        >
                            Add member
                        </Button>
                    </div>
                </Modal>
            )}
        </>
    )
}

interface AddMemberNotificationProps {
    username: string
    orgName: string
    onDismiss: () => void
    className?: string
}

export const AddMemberNotification: React.FunctionComponent<AddMemberNotificationProps> = ({
    className,
    username,
    orgName,
    onDismiss,
}) => (
    <Alert variant="success" className={classNames(styles.invitedNotification, className)}>
        <div className={styles.message}>
            <strong>{`You succesfully added ${username} to ${orgName}`}</strong>
        </div>
        <Button className="btn-icon" title="Dismiss" onClick={onDismiss}>
            <CloseIcon className="icon-inline" />
        </Button>
    </Alert>
)
