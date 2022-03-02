import { useMutation } from '@apollo/client'
import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'
import { debounce } from 'lodash'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { Component, FunctionComponent, useCallback, useEffect, useState } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Alert, Button, ButtonProps, Modal } from '@sourcegraph/wildcard'

import { CopyableText } from '../../components/CopyableText'
import { InviteUserToOrganizationResult, InviteUserToOrganizationVariables } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

import { INVITE_USERNAME_OR_EMAIL_TO_ORG_MUTATION } from './gqlQueries'
import styles from './InviteMemberModal.module.scss'
import { AutocompleteSearchUsers } from './SearchUserAutocomplete'

export interface IModalInviteResult {
    username: string
    inviteResult: InviteUserToOrganizationResult
}
export interface InviteMemberModalProps {
    orgName: string
    orgId: string
    onInviteSent: (result: IModalInviteResult) => void
    onDismiss: () => void
}

export const InviteMemberModal: React.FunctionComponent<InviteMemberModalProps> = props => {
    const { orgName, orgId, onInviteSent, onDismiss } = props
    const [userNameOrEmail, setUsernameOrEmail] = useState('')
    const [isEmail, setIsEmail] = useState<boolean>(false)
    const title = `Invite a teammate to ${orgName}`

    const [inviteUserToOrganization, { data, loading: isInviting, error }] = useMutation<
        InviteUserToOrganizationResult,
        InviteUserToOrganizationVariables
    >(INVITE_USERNAME_OR_EMAIL_TO_ORG_MUTATION)

    useEffect(() => {
        if (data) {
            onInviteSent({ username: userNameOrEmail, inviteResult: data })
            setUsernameOrEmail('')
            onDismiss()
        }
    }, [data, onDismiss, setUsernameOrEmail, onInviteSent, userNameOrEmail])

    const onValueChanged = useCallback((value: string, isEmail: boolean) => {
        setUsernameOrEmail(value)
        setIsEmail(isEmail)
    }, [])

    const inviteUser = useCallback(async () => {
        if (!userNameOrEmail) {
            return
        }

        eventLogger.log('InviteOrgMemberClicked', isEmail)
        try {
            await inviteUserToOrganization({
                variables: {
                    organization: orgId,
                    username: isEmail ? null : userNameOrEmail,
                    email: isEmail ? userNameOrEmail : null,
                },
            })
            eventLogger.log('OrgMemberInvited')
        } catch {
            eventLogger.log('OrgMemberInviteFailed')
        }
    }, [userNameOrEmail, orgId, inviteUserToOrganization, isEmail])

    const debounceInviteUser = debounce(inviteUser, 500, { leading: true })

    return (
        <Modal className={styles.modal} onDismiss={onDismiss} position="center" aria-label={title}>
            <div className="d-flex flex-row align-items-end">
                <h3>{title}</h3>
                <Button className={classNames('btn-icon', styles.closeButton)} onClick={onDismiss}>
                    <VisuallyHidden>Close</VisuallyHidden>
                    <CloseIcon />
                </Button>
            </div>
            {error && <ErrorAlert className={styles.alert} error={error} />}
            <div className="d-flex flex-row position-relative mt-2">
                <AutocompleteSearchUsers onValueChanged={onValueChanged} disabled={isInviting} orgId={orgId} />
            </div>
            <div className="d-flex justify-content-end mt-4">
                <Button type="button" variant="primary" onClick={debounceInviteUser} disabled={isInviting}>
                    Send invite
                </Button>
            </div>
        </Modal>
    )
}

interface InvitedNotificationProps {
    username: string
    orgName: string
    invitationURL: string
    onDismiss: () => void
    className?: string
}

export const InvitedNotification: React.FunctionComponent<InvitedNotificationProps> = ({
    className,
    username,
    orgName,
    invitationURL,
    onDismiss,
}) => (
    <Alert variant="success" className={classNames(styles.invitedNotification, className)}>
        <div className={styles.message}>
            <div>{`You invited ${username} to join ${orgName}`}</div>
            <div>They will receive an email shortly. You can also send them this personal invite link:</div>
            <CopyableText text={invitationURL} size={40} className="mt-2" />
        </div>
        <Button className="btn-icon" title="Dismiss" onClick={onDismiss}>
            <CloseIcon className="icon-inline" />
        </Button>
    </Alert>
)

export interface InviteMemberModalButtonProps extends ButtonProps {
    orgName: string
    orgId: string
    onInviteSent: (result: IModalInviteResult) => void
    triggerLabel?: string
    as?: keyof JSX.IntrinsicElements | Component | FunctionComponent
}
export const InviteMemberModalHandler: React.FunctionComponent<InviteMemberModalButtonProps> = (
    props: InviteMemberModalButtonProps
) => {
    const { orgName, orgId, onInviteSent, triggerLabel, as, ...rest } = props
    const [modalOpened, setModalOpened] = React.useState<boolean>()

    const onInviteClick = useCallback(() => {
        setModalOpened(true)
    }, [setModalOpened])

    const onCloseIviteModal = useCallback(() => {
        setModalOpened(false)
    }, [setModalOpened])

    return (
        <>
            <Button {...rest} onClick={onInviteClick} as={as as any} size="sm">
                {triggerLabel || 'Invite member'}
            </Button>

            {modalOpened && (
                <InviteMemberModal
                    orgId={orgId}
                    orgName={orgName}
                    onInviteSent={onInviteSent}
                    onDismiss={onCloseIviteModal}
                />
            )}
        </>
    )
}
