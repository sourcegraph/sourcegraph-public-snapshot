import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { useCallback, useEffect, useRef, useState } from 'react'
import { Alert, Button, Input, Modal } from '@sourcegraph/wildcard'
import { eventLogger } from '../../tracking/eventLogger'
import { gql, useMutation } from '@apollo/client'
import { InviteUserToOrganizationResult, InviteUserToOrganizationVariables } from '../../graphql-operations'
import { debounce } from 'lodash'
import EmailOpenOutlineIcon from 'mdi-react/EmailBoxIcon'
import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import styles from './InviteMemberModal.module.scss'
import { CopyableText } from '../../components/CopyableText'

const INVITE_USERNAME_OR_EMAIL_TO_ORG = gql`
    mutation InviteUserToOrganization($organization: ID!, $username: String!) {
        inviteUserToOrganization(organization: $organization, username: $username) {
            ...InviteUserToOrganizationFields
        }
    }

    fragment InviteUserToOrganizationFields on InviteUserToOrganizationResult {
        sentInvitationEmail
        invitationURL
    }
`
export interface IModalInviteResult {
    username: string
    inviteResult: InviteUserToOrganizationResult
}
export interface InviteMemberModalProps {
    orgName: string
    orgId: string
    onInviteSent: (result: IModalInviteResult) => void
}

export const InviteMemberModal: React.FunctionComponent<InviteMemberModalProps> = props => {
    const { orgName, orgId, onInviteSent } = props
    const emailPattern = useRef(new RegExp(/^\w+([\.-]?\w+)*@\w+([\.-]?\w+)*(\.\w{2,3})+$/))
    const [userNameOrEmail, setUsernameOrEmail] = useState('')
    const [modalOpened, setModalOpened] = React.useState<boolean>()
    const [isEmail, setIsEmail] = useState<boolean>(false)
    const title = `Invite teammate to ${orgName}`

    // const onInviteSentMessageDismiss = useCallback(() => {
    //     setInvite(undefined)
    // }, [setInvite])

    useEffect(() => {
        setIsEmail(emailPattern.current.test(userNameOrEmail))
    }, [userNameOrEmail])

    const [inviteUserToOrganization, { data, loading: isInviting, error }] = useMutation<
        InviteUserToOrganizationResult,
        InviteUserToOrganizationVariables
    >(INVITE_USERNAME_OR_EMAIL_TO_ORG)

    useEffect(() => {
        if (data) {
            eventLogger.log('OrgMemberInvited')
            onInviteSent({ username: userNameOrEmail, inviteResult: data })
            setUsernameOrEmail('')
            setModalOpened(false)
        }
    }, [data])

    useEffect(() => {
        if (error) {
            eventLogger.log('OrgMemberInviteFailed')
        }
    }, [error])

    const onInviteClick = useCallback(() => {
        setModalOpened(true)
    }, [setModalOpened])

    const onCloseIviteModal = useCallback(() => {
        setModalOpened(false)
    }, [setModalOpened])

    const onUsernameChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setUsernameOrEmail(event.currentTarget.value)
    }, [])

    const inviteUser = useCallback(async () => {
        if (!userNameOrEmail) {
            return
        }

        eventLogger.log('InviteOrgMemberClicked')
        await inviteUserToOrganization({ variables: { organization: orgId, username: userNameOrEmail } })
    }, [userNameOrEmail, orgId, isEmail, onInviteSent])

    const debounceInviteUser = debounce(inviteUser, 500, { leading: true })

    return (
        <>
            <Button variant="success" onClick={onInviteClick}>
                Invite member
            </Button>
            {modalOpened && (
                <Modal className={styles.modal} onDismiss={onCloseIviteModal} position="center" aria-label={title}>
                    <Button className={classNames('btn-icon', styles.closeButton)} onClick={onCloseIviteModal}>
                        <VisuallyHidden>Close</VisuallyHidden>
                        <CloseIcon />
                    </Button>

                    <h2>{title}</h2>
                    {error && <ErrorAlert className={styles.alert} error={error} />}
                    <div className="d-flex flex-row position-relative">
                        {isEmail && <EmailOpenOutlineIcon className={`icon-inline ${styles.mailIcon}`} />}
                        <Input
                            autoFocus
                            value={userNameOrEmail}
                            label="Email address or username"
                            title="Email address or username"
                            onChange={onUsernameChange}
                            status={isInviting ? 'loading' : error ? 'error' : undefined}
                            placeholder="Email address or username"
                        />
                    </div>
                    <div className="d-flex justify-content-end mt-4">
                        <Button
                            type="button"
                            className="mr-2"
                            variant="primary"
                            onClick={debounceInviteUser}
                            disabled={isInviting}
                        >
                            Send invite
                        </Button>
                    </div>
                </Modal>
            )}
        </>
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
            <strong>{`You invited ${username} to join ${orgName}`}</strong>
            <div>{`They will receive an email shortly. You can also send them this personal invite link:`}</div>
            <CopyableText text={invitationURL} size={40} className="mt-2" />
        </div>
        <Button className="btn-icon" title="Dismiss" onClick={onDismiss}>
            <CloseIcon className="icon-inline" />
        </Button>
    </Alert>
)
