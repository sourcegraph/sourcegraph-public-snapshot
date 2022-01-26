import React, { useCallback, useEffect, useState } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError, isErrorLike } from '@sourcegraph/common'
import { Container, PageHeader, Button, LoadingSpinner } from '@sourcegraph/wildcard'

import { ORG_DISPLAY_NAME_MAX_LENGTH } from '../'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { OrgAreaPageProps } from '../area/OrgArea'
import { updateOrganization } from '../backend'
import { InviteMemberModal } from './InviteMemberModal'

interface Props extends Pick<OrgAreaPageProps, 'org' | 'onOrganizationUpdate'> {}

/**
 * The organization members list page.
 */
export const OrgMembersListPage: React.FunctionComponent<Props> = ({ org, onOrganizationUpdate }) => {
    useEffect(() => {
        eventLogger.logViewEvent('OrgSettingsProfile')
    }, [org.id])

    const [modalInvite, setModalInvite] = React.useState(false)

    // const [displayName, setDisplayName] = useState<string>(org.displayName ?? '')
    // const onDisplayNameFieldChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
    //     setDisplayName(event.target.value)
    // }, [])
    // const [isLoading, setIsLoading] = useState<boolean | Error>(false)
    // const [updated, setIsUpdated] = useState<boolean>(false)
    // const [updateResetTimer, setUpdateResetTimer] = useState<NodeJS.Timer>()

    // useEffect(
    //     () => () => {
    //         if (updateResetTimer) {
    //             clearTimeout(updateResetTimer)
    //         }
    //     },
    //     [updateResetTimer]
    // )

    // const onSubmit = useCallback<React.FormEventHandler>(
    //     async event => {
    //         event.preventDefault()
    //         setIsLoading(true)
    //         try {
    //             await updateOrganization(org.id, displayName)
    //             onOrganizationUpdate()
    //             // Reenable submit button, flash "updated" text
    //             setIsLoading(false)
    //             setIsUpdated(true)
    //             setUpdateResetTimer(
    //                 setTimeout(() => {
    //                     // Hide "updated" text again after 1s
    //                     setIsUpdated(false)
    //                 }, 1000)
    //             )
    //         } catch (error) {
    //             setIsLoading(asError(error))
    //         }
    //     },
    //     [displayName, onOrganizationUpdate, org.id]
    // )
    const onInviteClick = () => {
        setModalInvite(true)
    }

    const onCloseIviteModal = () => {
        setModalInvite(false)
    }
    return (
        <>
            <div className="org-members-page">
                <PageTitle title={`${org.name} Members`} />
                <div className="d-flex flex-0 justify-content-between align-items-start">
                    <PageHeader path={[{ text: 'Organization Members' }]} headingElement="h2" className="mb-3" />
                    <Button variant="success" onClick={onInviteClick}>
                        Invite member
                    </Button>
                </div>

                <Container>list of members here</Container>
            </div>
            {modalInvite && <InviteMemberModal onClose={onCloseIviteModal} orgName={org.name} />}
        </>
    )
}
