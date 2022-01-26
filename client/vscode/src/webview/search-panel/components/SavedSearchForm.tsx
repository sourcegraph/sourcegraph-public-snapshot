import classNames from 'classnames'
import React, { useMemo, useState } from 'react'
import { map } from 'rxjs/operators'
import { Omit } from 'utility-types'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Form } from '@sourcegraph/branded/src/components/Form'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { Container, Link, PageHeader, ProductStatusBadge } from '@sourcegraph/wildcard'

import { CreateSavedSearchResult, CreateSavedSearchVariables, SavedSearchFields } from '../../../graphql-operations'
import { WebviewPageProps } from '../../platform/context'

import styles from './SavedSearchForm.module.scss'

// Debt: this is a fork of the web <SearchResultsInfobar>.

export interface SavedSearchFormProps {
    authenticatedUser: AuthenticatedUser | null
    defaultValues?: Partial<SavedSearchFields>
    title?: string
    submitLabel: string
    onSubmit: (fields: Omit<SavedSearchFields, 'id' | 'namespace'>) => void
    loading: boolean
    error?: any
    fullQuery: string
}

const savedSearchFragment = gql`
    fragment SavedSearchFields on SavedSearch {
        id
        description
        notify
        notifySlack
        query
        namespace {
            __typename
            id
            namespaceName
        }
        slackWebhookURL
    }
`

const createSavedSearchQuery = gql`
    mutation CreateSavedSearch(
        $description: String!
        $query: String!
        $notifyOwner: Boolean!
        $notifySlack: Boolean!
        $userID: ID
        $orgID: ID
    ) {
        createSavedSearch(
            description: $description
            query: $query
            notifyOwner: $notifyOwner
            notifySlack: $notifySlack
            userID: $userID
            orgID: $orgID
        ) {
            ...SavedSearchFields
        }
    }
    ${savedSearchFragment}
`

export interface SavedSearchCreateFormProps
    extends Omit<SavedSearchFormProps, 'loading' | 'error' | 'onSubmit'>,
        Pick<WebviewPageProps, 'platformContext'> {
    authenticatedUser: AuthenticatedUser
    onComplete: () => void
}

export const SavedSearchCreateForm: React.FunctionComponent<SavedSearchCreateFormProps> = props => {
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<any>()
    const onSubmit: SavedSearchFormProps['onSubmit'] = fields => {
        if (!loading) {
            setLoading(true)

            props.platformContext
                .requestGraphQL<CreateSavedSearchResult, CreateSavedSearchVariables>({
                    request: createSavedSearchQuery,
                    variables: {
                        ...fields,
                        notifyOwner: fields.notify,
                        orgID: null,
                        userID: props.authenticatedUser.id,
                    },
                    mightContainPrivateInfo: true,
                })
                .pipe(map(dataOrThrowErrors))
                .toPromise()
                .then(() => {
                    // Don't need to set loading to false, this form will be closed.
                    props.onComplete()
                })
                .catch(error => {
                    setLoading(false)
                    setError(error)
                })
        }
    }

    const defaultValues: Partial<SavedSearchFields> = {
        id: '',
        description: '',
        query: props.fullQuery,
        notify: false,
        notifySlack: false,
        slackWebhookURL: null,
    }

    return (
        <SavedSearchForm {...props} onSubmit={onSubmit} defaultValues={defaultValues} loading={loading} error={error} />
    )
}

const SavedSearchForm: React.FunctionComponent<SavedSearchFormProps> = props => {
    const [values, setValues] = useState<Omit<SavedSearchFields, 'id' | 'namespace'>>(() => ({
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
        key: keyof SavedSearchFields
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
        <div className="saved-search-form">
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
                                <label>
                                    <input
                                        type="checkbox"
                                        name="Notify owner"
                                        className={styles.checkbox}
                                        defaultChecked={notify}
                                        onChange={createInputChangeHandler('notify')}
                                    />{' '}
                                    <span>Send email notifications to my email</span>
                                </label>
                            </div>

                            <div className={classNames(styles.codeMonitoringAlert, 'alert alert-primary p-3 mb-0')}>
                                <div className="mb-2">
                                    <strong>New:</strong> Watch your code for changes with code monitoring to get
                                    notifications.
                                </div>
                                <Link to={codeMonitoringUrl} className="btn btn-primary">
                                    Go to code monitoring →
                                </Link>
                            </div>
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
                        <div className="alert alert-warning mt-3 mb-0">
                            <strong>Warning:</strong> non-commit searches do not currently support notifications.
                            Consider adding <code>type:diff</code> or <code>type:commit</code> to your query.
                        </div>
                    )}
                    {notify && !isUnsupportedNotifyQuery && (
                        <div className="alert alert-warning mt-3 mb-0">
                            <strong>Warning:</strong> Sending emails is not currently configured on this Sourcegraph
                            server. {props.authenticatedUser && 'Contact your server admin to enable sending emails.'}
                        </div>
                    )}
                </Container>
                <button
                    type="submit"
                    disabled={props.loading}
                    className={classNames(styles.submitButton, 'btn btn-primary test-saved-search-form-submit-button')}
                >
                    {props.submitLabel}
                </button>

                {props.error && !props.loading && <ErrorAlert className="mb-3" error={props.error} />}

                {!props.defaultValues?.notify && (
                    <Container className="d-flex p-3 align-items-start">
                        <ProductStatusBadge status="new" className="mr-3">
                            New
                        </ProductStatusBadge>
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
