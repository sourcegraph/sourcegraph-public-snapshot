import React, { useCallback, useState } from 'react'

import { useMutation } from '@apollo/client'
import { mdiClose } from '@mdi/js'
import { VisuallyHidden } from '@reach/visually-hidden'
import { debounce } from 'lodash'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Button, Input, Modal, H3, Icon } from '@sourcegraph/wildcard'

import { AddUserToOrganizationResult, AddUserToOrganizationVariables } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

import { ADD_USERNAME_OR_EMAIL_TO_ORG_MUTATION } from './gqlQueries'

import styles from './InviteMemberModal.module.scss'

export interface AddMemberToOrgModalProps {
    orgName: string
    orgId: string
    onMemberAdded: (username: string) => void
}

export const AddMemberToOrgModal: React.FunctionComponent<
    React.PropsWithChildren<AddMemberToOrgModalProps>
> = props => {
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
    >(ADD_USERNAME_OR_EMAIL_TO_ORG_MUTATION)

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
            <Button variant="primary" onClick={onAddUserClick} size="sm" className="mr-1">
                + Add member
            </Button>
            {modalOpened && (
                <Modal className={styles.modal} onDismiss={onCloseAddUserModal} position="center" aria-label={title}>
                    <div className="d-flex flex-row align-items-end">
                        <H3>{title}</H3>
                        <Button variant="icon" className={styles.closeButton} onClick={onCloseAddUserModal}>
                            <VisuallyHidden>Close</VisuallyHidden>
                            <Icon svgPath={mdiClose} inline={false} aria-hidden={true} />
                        </Button>
                    </div>
                    {error && <ErrorAlert className={styles.alert} error={error} />}
                    <div className="d-flex flex-row position-relative mt-2">
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
                        <Button type="button" variant="primary" onClick={debounceAddUser} disabled={loading}>
                            Add member
                        </Button>
                    </div>
                </Modal>
            )}
        </>
    )
}
