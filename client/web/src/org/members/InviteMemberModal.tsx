import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { useCallback, useEffect, useRef, useState } from 'react'
import { Button, Input, Modal } from '@sourcegraph/wildcard'
import styles from './InviteMemberModal.module.scss'
import { eventLogger } from '../../tracking/eventLogger'
import { gql, useMutation } from '@apollo/client'
import { InviteUserToOrganizationResult, InviteUserToOrganizationVariables } from '../../graphql-operations'
import { debounce } from 'lodash'
import EmailOpenOutlineIcon from 'mdi-react/EmailBoxIcon'
import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'

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

export interface InviteMemberModalProps {
    orgName: string
    orgId: string
}

const emailIconStyle: React.CSSProperties = {
    position: 'absolute',
    top: '38px',
    left: '-20px',
    zIndex: 1,
}

export const InviteMemberModal: React.FunctionComponent<InviteMemberModalProps> = props => {
    const { orgName, orgId } = props
    const emailPattern = useRef(new RegExp(/^\w+([\.-]?\w+)*@\w+([\.-]?\w+)*(\.\w{2,3})+$/))
    const [userNameOrEmail, setUsernameOrEmail] = useState('')
    const [modalOpened, setModalOpened] = React.useState(false)
    const [isEmail, setIsEmail] = useState<boolean>(false)
    const title = `Invite teammate to ${orgName}`

    const onInviteClick = useCallback(() => {
        setModalOpened(true)
    }, [setModalOpened])

    const onCloseIviteModal = useCallback(() => {
        setModalOpened(false)
    }, [setModalOpened])

    useEffect(() => {
        setIsEmail(emailPattern.current.test(userNameOrEmail))
    }, [userNameOrEmail])

    const [inviteUserToOrganization, { loading: isInviting, error }] = useMutation<
        InviteUserToOrganizationResult,
        InviteUserToOrganizationVariables
    >(INVITE_USERNAME_OR_EMAIL_TO_ORG)

    const onUsernameChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setUsernameOrEmail(event.currentTarget.value)
    }, [])

    const inviteUser = useCallback(async () => {
        if (!userNameOrEmail) {
            return
        }

        eventLogger.log('InviteOrgMemberClicked')
        await inviteUserToOrganization({ variables: { organization: orgId, username: userNameOrEmail } })
        setUsernameOrEmail('')
    }, [userNameOrEmail, orgId, isEmail])

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
                        {isEmail && <EmailOpenOutlineIcon className="icon-inline" style={emailIconStyle} />}
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
