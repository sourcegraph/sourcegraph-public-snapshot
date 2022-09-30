import React, { useMemo, useState } from 'react'

import { mdiClose } from '@mdi/js'
import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'
import classNames from 'classnames'
import { map } from 'rxjs/operators'
import { Omit } from 'utility-types'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Form } from '@sourcegraph/branded/src/components/Form'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { Checkbox, Container, Icon, PageHeader, Code, Label, Input } from '@sourcegraph/wildcard'

import { CreateSavedSearchResult, CreateSavedSearchVariables, SavedSearchFields } from '../../../graphql-operations'
import { WebviewPageProps } from '../../platform/context'

import styles from './SavedSearchForm.module.scss'

// Debt: this is a fork of the web <SearchResultsInfobar>.

export interface SavedSearchFormProps extends Pick<WebviewPageProps, 'instanceURL'> {
    authenticatedUser: AuthenticatedUser | null
    defaultValues?: Partial<SavedSearchFields>
    title?: string
    submitLabel: string
    onSubmit: (fields: Omit<SavedSearchFields, 'id' | 'namespace'>) => void
    loading: boolean
    error?: any
    fullQuery: string
    onComplete: () => void
}

export interface SavedSearchCreateFormProps
    extends Omit<SavedSearchFormProps, 'loading' | 'error' | 'onSubmit'>,
        Pick<WebviewPageProps, 'platformContext' | 'instanceURL'> {
    authenticatedUser: AuthenticatedUser
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

export const SavedSearchCreateForm: React.FunctionComponent<
    React.PropsWithChildren<SavedSearchCreateFormProps>
> = props => {
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<any>()
    const onSubmit: SavedSearchFormProps['onSubmit'] = fields => {
        if (!loading) {
            setLoading(true)
            props.platformContext.telemetryService.log('VSCESaveSearchSubmited')
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

const SavedSearchForm: React.FunctionComponent<React.PropsWithChildren<SavedSearchFormProps>> = props => {
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

    const { query, description, notify, notifySlack, slackWebhookURL } = values

    return (
        <div className={classNames(styles.container, 'saved-search-form position-relative')}>
            <Icon
                className="position-absolute cursor-pointer"
                style={{ top: '1rem', right: '1rem' }}
                onClick={props.onComplete}
                aria-label="Close"
                svgPath={mdiClose}
            />
            <Form onSubmit={handleSubmit}>
                <Container className={styles.container}>
                    <PageHeader
                        path={[{ text: props.title }]}
                        headingElement="h2"
                        description="Get notifications when there are new results for specific search queries."
                        className="mb-3"
                    />
                    <Input
                        id="saved-search-form-input-description"
                        name="description"
                        data-testid="test-saved-search-form-input-description"
                        placeholder="Description"
                        required={true}
                        value={description}
                        onChange={createInputChangeHandler('description')}
                        label="Description"
                        className={styles.label}
                    />
                    <Input
                        id="saved-search-form-input-query"
                        name="query"
                        placeholder="Query"
                        required={true}
                        value={query}
                        onChange={createInputChangeHandler('query')}
                        label="Query"
                        className={styles.label}
                    />

                    {props.defaultValues?.notify && (
                        <div>
                            {/* Label is for visual benefit, input has more specific label attached */}
                            <Label className={styles.label} id="saved-search-form-email-notifications">
                                Email notifications
                            </Label>
                            <div aria-labelledby="saved-search-form-email-notifications">
                                <Checkbox
                                    name="Notify owner"
                                    id="SendEmailNotificationsCheck"
                                    wrapperClassName="mb-2"
                                    className={classNames(styles.checkbox, 'mr-0')}
                                    defaultChecked={notify}
                                    onChange={createInputChangeHandler('notify')}
                                    label={<span className="ml-2">Send email notifications to my email</span>}
                                />
                            </div>
                        </div>
                    )}

                    {notifySlack && slackWebhookURL && (
                        <Input
                            id="saved-search-form-input-slack"
                            name="Slack webhook URL"
                            value={slackWebhookURL}
                            disabled={true}
                            onChange={createInputChangeHandler('slackWebhookURL')}
                            label="Slack notifications"
                            message="Slack webhooks are deprecated and will be removed in a future Sourcegraph version."
                            className={classNames('mt-3', styles.label)}
                        />
                    )}
                    {isUnsupportedNotifyQuery && (
                        <div className="alert alert-warning mt-3 mb-0">
                            <strong>Warning:</strong> non-commit searches do not currently support notifications.
                            Consider adding <Code>type:diff</Code> or <Code>type:commit</Code> to your query.
                        </div>
                    )}
                    {notify && !isUnsupportedNotifyQuery && (
                        <div className="alert alert-warning mt-3 mb-0">
                            <strong>Warning:</strong> Sending emails is not currently configured on this Sourcegraph
                            server. {props.authenticatedUser && 'Contact your server admin to enable sending emails.'}
                        </div>
                    )}
                    {props.error && !props.loading && <ErrorAlert className="my-3" error={props.error} />}
                    <div className="text-right">
                        <VSCodeButton type="submit" disabled={props.loading}>
                            {props.submitLabel}
                        </VSCodeButton>
                    </div>
                </Container>
            </Form>
        </div>
    )
}
