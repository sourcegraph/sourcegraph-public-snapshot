import React, { useCallback, useEffect, useState } from 'react'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { asError, isErrorLike } from '@sourcegraph/common'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Container, PageHeader, Button, LoadingSpinner, Input, Text, ErrorAlert, Form } from '@sourcegraph/wildcard'

import { ORG_DISPLAY_NAME_MAX_LENGTH } from '../..'
import { PageTitle } from '../../../components/PageTitle'
import { eventLogger } from '../../../tracking/eventLogger'
import type { OrgAreaRouteContext } from '../../area/OrgArea'
import { updateOrganization } from '../../backend'

interface Props extends Pick<OrgAreaRouteContext, 'org' | 'onOrganizationUpdate'>, TelemetryV2Props {}

/**
 * The organization profile settings page.
 */
export const OrgSettingsProfilePage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    org,
    onOrganizationUpdate,
    telemetryRecorder,
}) => {
    useEffect(() => {
        telemetryRecorder.recordEvent('orgSettingsProfile', 'viewed')
        eventLogger.logViewEvent('OrgSettingsProfile')
    }, [org.id])

    const [displayName, setDisplayName] = useState<string>(org.displayName ?? '')
    const onDisplayNameFieldChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setDisplayName(event.target.value)
    }, [])
    const [isLoading, setIsLoading] = useState<boolean | Error>(false)
    const [updated, setIsUpdated] = useState<boolean>(false)
    const [updateResetTimer, setUpdateResetTimer] = useState<NodeJS.Timer>()

    useEffect(
        () => () => {
            if (updateResetTimer) {
                clearTimeout(updateResetTimer)
            }
        },
        [updateResetTimer]
    )

    const onSubmit = useCallback<React.FormEventHandler>(
        async event => {
            event.preventDefault()
            setIsLoading(true)
            try {
                await updateOrganization(org.id, displayName)
                onOrganizationUpdate()
                // Reenable submit button, flash "updated" text
                setIsLoading(false)
                setIsUpdated(true)
                setUpdateResetTimer(
                    setTimeout(() => {
                        // Hide "updated" text again after 1s
                        setIsUpdated(false)
                    }, 1000)
                )
            } catch (error) {
                setIsLoading(asError(error))
            }
        },
        [displayName, onOrganizationUpdate, org.id]
    )

    return (
        <div className="org-settings-profile-page">
            <PageTitle title={org.name} />
            <PageHeader
                path={[{ text: 'Organization profile' }]}
                headingElement="h2"
                description={
                    <>
                        {org.displayName ? (
                            <>
                                {org.displayName} ({org.name})
                            </>
                        ) : (
                            org.name
                        )}{' '}
                        was created <Timestamp date={org.createdAt} />.
                    </>
                }
                className="mb-3"
            />
            <Container>
                <Form className="org-settings-profile-page" onSubmit={onSubmit}>
                    <Input
                        id="org-settings-profile-page-display-name"
                        inputClassName="org-settings-profile-page__display-name"
                        placeholder="Organization name"
                        onChange={onDisplayNameFieldChange}
                        value={displayName}
                        spellCheck={false}
                        maxLength={ORG_DISPLAY_NAME_MAX_LENGTH}
                        label="Display name"
                        className="form-group"
                    />

                    <Button type="submit" disabled={isLoading === true} variant="primary">
                        Update
                    </Button>
                    {isLoading === true && <LoadingSpinner />}
                    {updated && (
                        <Text className="mb-0">
                            <small>Updated!</small>
                        </Text>
                    )}
                    {isErrorLike(isLoading) && <ErrorAlert error={isLoading} />}
                </Form>
            </Container>
        </div>
    )
}
