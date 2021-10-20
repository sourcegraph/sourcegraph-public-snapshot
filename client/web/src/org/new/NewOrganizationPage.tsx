import * as H from 'history'
import React, { useCallback, useEffect, useState } from 'react'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { ORG_NAME_MAX_LENGTH, VALID_ORG_NAME_REGEXP } from '..'
import { ErrorAlert } from '../../components/alerts'
import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { createOrganization } from '../backend'

import styles from './NewOrgPage.module.scss'

interface Props {
    history: H.History
}

export const NewOrganizationPage: React.FunctionComponent<Props> = ({ history }) => {
    useEffect(() => {
        eventLogger.logViewEvent('NewOrg')
    }, [])
    const [loading, setLoading] = useState<boolean | Error>(false)
    const [name, setName] = useState<string>('')
    const [displayName, setDisplayName] = useState<string>('')

    const onNameChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        const hyphenatedName = event.currentTarget.value.replace(/\s/g, '-')
        setName(hyphenatedName)
    }

    const onDisplayNameChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        setDisplayName(event.currentTarget.value)
    }

    const onSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(
        async event => {
            event.preventDefault()
            eventLogger.log('CreateNewOrgClicked')
            if (!event.currentTarget.checkValidity()) {
                return
            }
            setLoading(true)
            try {
                const org = await createOrganization({ name, displayName })
                setLoading(false)
                history.push(org.settingsURL!)
            } catch (error) {
                setLoading(asError(error))
            }
        },
        [displayName, history, name]
    )

    return (
        <Page className={styles.newOrgPage}>
            <PageTitle title="New organization" />
            <PageHeader
                path={[{ text: 'Create a new organization' }]}
                description={
                    <>
                        An organization is a set of users with associated configuration. See{' '}
                        <Link to="/help/admin/organizations">Sourcegraph documentation</Link> for information about
                        configuring organizations.
                    </>
                }
                className="mb-3"
            />
            <Form className="settings-form" onSubmit={onSubmit}>
                <Container className="mb-3">
                    {isErrorLike(loading) && <ErrorAlert className="mb-3" error={loading} />}
                    <div className="form-group">
                        <label htmlFor="new-org-page__form-name">Organization name</label>
                        <input
                            id="new-org-page__form-name"
                            type="text"
                            className="form-control test-new-org-name-input"
                            placeholder="acme-corp"
                            pattern={VALID_ORG_NAME_REGEXP}
                            maxLength={ORG_NAME_MAX_LENGTH}
                            required={true}
                            autoCorrect="off"
                            autoComplete="off"
                            autoFocus={true}
                            value={name}
                            onChange={onNameChange}
                            disabled={loading === true}
                            aria-describedby="new-org-page__form-name-help"
                        />
                        <small id="new-org-page__form-name-help" className="form-text text-muted">
                            An organization name consists of letters, numbers, hyphens (-), dots (.) and may not begin
                            or end with a dot, nor begin with a hyphen.
                        </small>
                    </div>

                    <div className="form-group mb-0">
                        <label htmlFor="new-org-page__form-display-name">Display name</label>
                        <input
                            id="new-org-page__form-display-name"
                            type="text"
                            className="form-control test-new-org-display-name-input"
                            placeholder="ACME Corporation"
                            autoCorrect="off"
                            value={displayName}
                            onChange={onDisplayNameChange}
                            disabled={loading === true}
                        />
                    </div>
                </Container>

                <button
                    type="submit"
                    className="btn btn-primary test-create-org-submit-button"
                    disabled={loading === true}
                >
                    {loading === true && <LoadingSpinner className="icon-inline" />}
                    Create organization
                </button>
            </Form>
        </Page>
    )
}
