import { gql, useLazyQuery, useMutation } from '@apollo/client'
import classNames from 'classnames'
import { debounce } from 'lodash'
import React, { useCallback, useEffect, useRef, useState } from 'react'
import { RouteComponentProps } from 'react-router'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Form } from '@sourcegraph/branded/src/components/Form'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { Page } from '@sourcegraph/web/src/components/Page'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { Alert, Button, Checkbox, Container, Input, Link, LoadingSpinner, PageHeader } from '@sourcegraph/wildcard'

import { ORG_NAME_MAX_LENGTH, VALID_ORG_NAME_REGEXP } from '..'
import {
    CreateOrganizationForOpenBetaResult,
    CreateOrganizationForOpenBetaVariables,
    TryGetOrganizationIDByNameResult,
    TryGetOrganizationIDByNameVariables,
} from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

import styles from './NewOrganization.module.scss'

interface Props extends RouteComponentProps<{ openBetaId: string }> {
    authenticatedUser: AuthenticatedUser
}

const CREATE_ORG_MUTATION = gql`
    mutation CreateOrganizationForOpenBeta($name: String!, $displayName: String) {
        createOrganization(name: $name, displayName: $displayName) {
            id
            name
            settingsURL
        }
    }
`

const TRY_GET_ORG_ID_BY_NAME = gql`
    query TryGetOrganizationIDByName($name: String!) {
        organization(name: $name) {
            id
        }
    }
`

const isValidOpenBetaId = (openBetaId: string): boolean => {
    try {
        if (openBetaId === 'testdev') {
            return true
        }
        const waitingOpenBetaId = localStorage.getItem('sgopenBetaId')
        const isValid = openBetaId === waitingOpenBetaId
        eventLogger.log('OpenBetaIdCheck', { valid: isValid }, { valid: isValid })
        return isValid
    } catch (error: unknown) {
        eventLogger.log('OpenBetaIdCheck', { error }, { error })
        return false
    }
}

const normalizeOrgId = (id: string): string => id.toLowerCase().replace(/[\W_]+/g, '-')

export const NewOrgOpenBetaPage: React.FunctionComponent<Props> = ({ match, history, authenticatedUser }) => {
    const openBetaId = match.params.openBetaId
    const [orgId, setOrgId] = useState<string>('')
    const [displayName, setDisplayName] = useState<string>('')
    const [termsAccepted, setTermsAccepted] = useState(false)
    const displayBox = useRef(false)
    const isSuggested = useRef(false)

    const [createOpenBetaOrg, { loading, error }] = useMutation<
        CreateOrganizationForOpenBetaResult,
        CreateOrganizationForOpenBetaVariables
    >(CREATE_ORG_MUTATION)

    const [tryGetOrg, { loading: loadingOrg, data }] = useLazyQuery<
        TryGetOrganizationIDByNameResult,
        TryGetOrganizationIDByNameVariables
    >(TRY_GET_ORG_ID_BY_NAME, {
        variables: { name: orgId },
    })
    const debounceTryGetOrg = useRef(debounce(tryGetOrg, 250, { leading: false }))
    const existId = !!data?.organization?.id
    const hasValidId = !existId && orgId

    useEffect(() => {
        if (!openBetaId || !isValidOpenBetaId(openBetaId)) {
            history.push('/organizations/joinopenbeta')
        }
    }, [openBetaId, history])

    useEffect(() => {
        if (existId && isSuggested.current && orgId && !loadingOrg) {
            displayBox.current = true
            const autofixID = `${orgId}-1`
            tryGetOrg({ variables: { name: autofixID } })
            setOrgId(autofixID)
        }
    }, [existId, tryGetOrg, orgId, loadingOrg])

    const onDisplayNameChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        isSuggested.current = true
        displayBox.current = false
        const orgId = normalizeOrgId(event.currentTarget.value)
        setOrgId(orgId)
        setDisplayName(event.currentTarget.value)
        debounceTryGetOrg.current({ variables: { name: orgId } })
    }

    const onDisplayNameFocus: React.ChangeEventHandler<HTMLInputElement> = () => {
        if (displayName && !hasValidId && orgId) {
            displayBox.current = false
            isSuggested.current = true
            debounceTryGetOrg.current({ variables: { name: orgId } })
        }
    }

    const onOrgIdChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        isSuggested.current = false
        displayBox.current = false
        const orgId = normalizeOrgId(event.currentTarget.value)
        setOrgId(orgId)
        debounceTryGetOrg.current({ variables: { name: orgId } })
    }

    const onCancelClick = (): void => {
        history.push(`/users/${authenticatedUser.username}/settings/organizations`)
    }

    const onTermsAcceptedChange: React.ChangeEventHandler<HTMLInputElement> = () => {
        setTermsAccepted(!termsAccepted)
    }

    const onSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(
        async event => {
            event.preventDefault()
            eventLogger.log('CreateNewOrgBetaClicked')
            if (!event.currentTarget.checkValidity()) {
                return
            }
            try {
                const result = await createOpenBetaOrg({ variables: { name: orgId, displayName } })
                eventLogger.log('CreateNewOrgBetaOK')
                if (result?.data?.createOrganization) {
                    history.push(result.data.createOrganization.settingsURL as string)
                }
            } catch {
                eventLogger.log('CreateNewOrgBetaFailed')
            }
        },
        [orgId, displayName, history, createOpenBetaOrg]
    )

    return (
        <Page className={styles.newOrgPage}>
            <PageTitle title="New organization" />
            <PageHeader path={[{ text: 'Set up your organization' }]} className="mb-4" />
            <Form className="mb-3" onSubmit={onSubmit}>
                <Container className={styles.formContainer}>
                    {error && <ErrorAlert className="mb-3" error={error} />}
                    <div className={classNames('form-group', styles.formItem)}>
                        <label htmlFor="new-org-page__form-name">Organization name</label>
                        <input
                            id="new-org-page__form-name"
                            type="text"
                            className="form-control test-new-org-name-input mb-2"
                            placeholder="ACME Corporation"
                            maxLength={ORG_NAME_MAX_LENGTH}
                            required={true}
                            autoCorrect="off"
                            autoComplete="off"
                            autoFocus={true}
                            onFocus={onDisplayNameFocus}
                            value={displayName}
                            onChange={onDisplayNameChange}
                            disabled={loading}
                            aria-describedby="new-org-page__form-name-help"
                        />
                        <small id="new-org-page__form-name-help" className="form-text text-muted">
                            This will be your organization’s name on Sourcegraph. You can change this any time.
                        </small>
                    </div>

                    <div className={classNames('form-group', styles.formItem)}>
                        <Input
                            id="new-org-page__form-organizationid"
                            type="text"
                            placeholder="acme-corp"
                            autoCorrect="off"
                            value={orgId}
                            label="Organization identifier"
                            required={true}
                            pattern={VALID_ORG_NAME_REGEXP}
                            maxLength={ORG_NAME_MAX_LENGTH}
                            onChange={onOrgIdChange}
                            disabled={loading}
                            status={
                                loadingOrg
                                    ? 'loading'
                                    : hasValidId
                                    ? 'valid'
                                    : !isSuggested.current && orgId
                                    ? 'error'
                                    : undefined
                            }
                            title="An organization identifier consists of letters, numbers, hyphens (-), dots (.) and may not begin
                            or end with a dot, nor begin with a hyphen."
                        />
                        {displayBox.current && hasValidId && (
                            <Alert variant="info" className="mb-2">
                                <h4>We’ve suggested an alternative organization ID</h4>
                                <div>{`${normalizeOrgId(
                                    displayName
                                )} is already in use. Use our suggestion or choose a new ID for your organization.`}</div>
                            </Alert>
                        )}
                        {!loadingOrg && !hasValidId && !isSuggested.current && orgId && (
                            <Alert variant="info" className="mb-2">
                                <h4>{`The organization identifier ${orgId} is already in use`}</h4>
                                <div>Please try choosing a different one</div>
                            </Alert>
                        )}
                        <small id="new-org-page__form-orgid-help" className="form-text text-muted">
                            Cannot be changed after creating your organization. This will be used to reference your
                            organization in locations such as URLs or custom search contexts.
                        </small>
                    </div>

                    <div className="form-group">
                        <Checkbox
                            name="userSearchable"
                            id="userSearchable"
                            value="searchable"
                            checked={termsAccepted}
                            required={true}
                            onChange={onTermsAcceptedChange}
                            label={
                                <>
                                    I accept the <Link to="/">terms of service</Link> for participating in the
                                    Sourcegraph Cloud for small teams open beta.
                                </>
                            }
                        />
                    </div>

                    <div className={classNames('form-group d-flex justify-content-end', styles.buttonsRow)}>
                        <Button disabled={loading} variant="secondary" size="sm" onClick={onCancelClick}>
                            Cancel
                        </Button>
                        <Button
                            type="submit"
                            disabled={loading || !termsAccepted || !hasValidId || loadingOrg}
                            variant="primary"
                            size="sm"
                        >
                            {(loading || loadingOrg) && <LoadingSpinner />}
                            Continue
                        </Button>
                    </div>
                </Container>
            </Form>
        </Page>
    )
}
