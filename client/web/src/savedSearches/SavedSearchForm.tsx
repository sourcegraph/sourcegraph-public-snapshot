import React, { useMemo, useState } from 'react'

import classNames from 'classnames'
import { Omit } from 'utility-types'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Form } from '@sourcegraph/branded/src/components/Form'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { Container, PageHeader, ProductStatusBadge, Button, Link, Alert, Checkbox } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { NamespaceProps } from '../namespaces'

import styles from './SavedSearchForm.module.scss'

export interface SavedQueryFields {
    id: Scalars['ID']
    description: string
    query: string
    notify: boolean
    notifySlack: boolean
    slackWebhookURL: string | null
}

export interface SavedSearchFormProps extends NamespaceProps {
    authenticatedUser: AuthenticatedUser | null
    defaultValues?: Partial<SavedQueryFields>
    title?: string
    submitLabel: string
    onSubmit: (fields: Omit<SavedQueryFields, 'id'>) => void
    loading: boolean
    error?: any
}

export const SavedSearchForm: React.FunctionComponent<React.PropsWithChildren<SavedSearchFormProps>> = props => {
    const [values, setValues] = useState<Omit<SavedQueryFields, 'id'>>(() => ({
        description: props.defaultValues?.description || '',
        query: props.defaultValues?.query || '',
        notify: props.defaultValues?.notify || false,
        notifySlack: props.defaultValues?.notifySlack || false,
        slackWebhookURL: props.defaultValues?.slackWebhookURL || '',
    }))

    /**
     * Returns an input change handler that updates the SavedQueryFields in the component's state
     *
     * @param key The key of saved query fields that a change of this input should update
     */
    const createInputChangeHandler = (
        key: keyof SavedQueryFields
    ): React.FormEventHandler<HTMLInputElement> => event => {
        const { value, checked, type } = event.currentTarget
        setValues(values => ({
            ...values,
            [key]: type === 'checkbox' ? checked : value,
        }))
    }

    const handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        props.onSubmit(values)
    }

    /**
     * Tells if the query is unsupported for sending notifications.
     */
    const isUnsupportedNotifyQuery = useMemo((): boolean => {
        const notifying = values.notify || values.notifySlack
        return notifying && !values.query.includes('type:diff') && !values.query.includes('type:commit')
    }, [values])

    const codeMonitoringUrl = useMemo(() => {
        const searchParameters = new URLSearchParams()
        searchParameters.set('trigger-query', values.query)
        searchParameters.set('description', values.description)
        return `/code-monitoring/new?${searchParameters.toString()}`
    }, [values.query, values.description])

    const { query, description, notify, notifySlack, slackWebhookURL } = values

    return (
        <div className="saved-search-form" data-testid="saved-search-form">
            <PageHeader
                path={[{ text: props.title }]}
                headingElement="h2"
                description="Get notifications when there are new results for specific search queries."
                className="mb-3"
            />
            <Form onSubmit={handleSubmit}>
                <Container className="mb-3">
                    <div className="form-group">
                        <label className={styles.label} htmlFor="saved-search-form-input-description">
                            Description
                        </label>
                        <input
                            id="saved-search-form-input-description"
                            type="text"
                            name="description"
                            className="form-control test-saved-search-form-input-description"
                            placeholder="Description"
                            required={true}
                            value={description}
                            onChange={createInputChangeHandler('description')}
                        />
                    </div>
                    <div className="form-group">
                        <label className={styles.label} htmlFor="saved-search-form-input-query">
                            Query
                        </label>
                        <input
                            id="saved-search-form-input-query"
                            type="text"
                            name="query"
                            className="form-control test-saved-search-form-input-query"
                            placeholder="Query"
                            required={true}
                            value={query}
                            onChange={createInputChangeHandler('query')}
                        />
                    </div>

                    {props.defaultValues?.notify && (
                        <div className="form-group mb-0">
                            {/* Label is for visual benefit, input has more specific label attached */}
                            {/* eslint-disable-next-line jsx-a11y/label-has-associated-control */}
                            <label className={styles.label} id="saved-search-form-email-notifications">
                                Email notifications
                            </label>
                            <div aria-labelledby="saved-search-form-email-notifications">
                                <Checkbox
                                    name="Notify owner"
                                    className={classNames(styles.checkbox, 'mr-0')}
                                    defaultChecked={notify}
                                    wrapperClassName="mb-2"
                                    onChange={createInputChangeHandler('notify')}
                                    id="NotifyOrgMembersInput"
                                    label={
                                        <span className="ml-2">
                                            {props.namespace.__typename === 'Org'
                                                ? 'Send email notifications to all members of this organization'
                                                : props.namespace.__typename === 'User'
                                                ? 'Send email notifications to my email'
                                                : 'Email notifications'}
                                        </span>
                                    }
                                />
                            </div>

                            <Alert variant="primary" className={classNames(styles.codeMonitoringAlert, 'p-3 mb-0')}>
                                <div className="mb-2">
                                    <strong>New:</strong> Watch your code for changes with code monitoring to get
                                    notifications.
                                </div>
                                <Button to={codeMonitoringUrl} variant="primary" as={Link}>
                                    Go to code monitoring →
                                </Button>
                            </Alert>
                        </div>
                    )}

                    {notifySlack && slackWebhookURL && (
                        <div className="form-group mt-3 mb-0">
                            <label className={styles.label} htmlFor="saved-search-form-input-slack">
                                Slack notifications
                            </label>
                            <input
                                id="saved-search-form-input-slack"
                                type="text"
                                name="Slack webhook URL"
                                className="form-control"
                                value={slackWebhookURL}
                                disabled={true}
                                onChange={createInputChangeHandler('slackWebhookURL')}
                            />
                            <small>
                                Slack webhooks are deprecated and will be removed in a future Sourcegraph version.
                            </small>
                        </div>
                    )}
                    {isUnsupportedNotifyQuery && (
                        <Alert className="mt-3 mb-0" variant="warning">
                            <strong>Warning:</strong> non-commit searches do not currently support notifications.
                            Consider adding <code>type:diff</code> or <code>type:commit</code> to your query.
                        </Alert>
                    )}
                    {notify && !window.context.emailEnabled && !isUnsupportedNotifyQuery && (
                        <Alert className="mt-3 mb-0" variant="warning">
                            <strong>Warning:</strong> Sending emails is not currently configured on this Sourcegraph
                            server.{' '}
                            {props.authenticatedUser?.siteAdmin
                                ? 'Use the email.smtp site configuration setting to enable sending emails.'
                                : 'Contact your server admin for more information.'}
                        </Alert>
                    )}
                </Container>
                <Button
                    type="submit"
                    disabled={props.loading}
                    className={classNames(styles.submitButton, 'test-saved-search-form-submit-button')}
                    variant="primary"
                >
                    {props.submitLabel}
                </Button>

                {props.error && !props.loading && <ErrorAlert className="mb-3" error={props.error} />}

                {!props.defaultValues?.notify && (
                    <Container className="d-flex p-3 align-items-start">
                        <ProductStatusBadge status="new" className="mr-3" />
                        <span>
                            Watch for changes to your code and trigger email notifications, webhooks, and more with{' '}
                            <Link to="/code-monitoring">code monitoring →</Link>
                        </span>
                    </Container>
                )}
            </Form>
        </div>
    )
}
