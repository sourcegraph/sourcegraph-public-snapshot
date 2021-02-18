import React, { useCallback, useState } from 'react'
import * as H from 'history'
import Dialog from '@reach/dialog'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Form } from '../../../../../branded/src/components/Form'
import { asError, isErrorLike } from '../../../../../shared/src/util/errors'
import { ErrorAlert } from '../../../components/alerts'
import { createCampaignsCredential } from './backend'
import { ExternalServiceKind, Scalars } from '../../../graphql-operations'

export interface AddCredentialModalProps {
    onCancel: () => void
    afterCreate: () => void
    history: H.History
    userID: Scalars['ID']
    externalServiceKind: ExternalServiceKind
    externalServiceURL: string
}

const modalTitles: Record<ExternalServiceKind, string> = {
    [ExternalServiceKind.GITHUB]: 'GitHub',
    [ExternalServiceKind.GITLAB]: 'GitLab',
    [ExternalServiceKind.BITBUCKETSERVER]: 'Bitbucket Server',

    // These are just for type completeness and serve as placeholders for a bright future.
    [ExternalServiceKind.BITBUCKETCLOUD]: 'Unsupported',
    [ExternalServiceKind.GITOLITE]: 'Unsupported',
    [ExternalServiceKind.PERFORCE]: 'Unsupported',
    [ExternalServiceKind.PHABRICATOR]: 'Unsupported',
    [ExternalServiceKind.AWSCODECOMMIT]: 'Unsupported',
    [ExternalServiceKind.OTHER]: 'Unsupported',
}

const helpTexts: Record<ExternalServiceKind, JSX.Element> = {
    [ExternalServiceKind.GITHUB]: (
        <>
            <a
                href="https://docs.sourcegraph.com/campaigns/quickstart#configure-code-host-connections"
                rel="noreferrer noopener"
                target="_blank"
            >
                Create a new access token
            </a>{' '}
            with <code>repo</code>, <code>read:org</code>, <code>user:email</code>, and <code>read:discussion</code>{' '}
            scopes.
        </>
    ),
    [ExternalServiceKind.GITLAB]: (
        <>
            <a
                href="https://docs.sourcegraph.com/campaigns/quickstart#configure-code-host-connections"
                rel="noreferrer noopener"
                target="_blank"
            >
                Create a new access token
            </a>{' '}
            with <code>api</code>, <code>read_repository</code>, and <code>write_repository</code> scopes.
        </>
    ),
    [ExternalServiceKind.BITBUCKETSERVER]: (
        <>
            <a
                href="https://docs.sourcegraph.com/campaigns/quickstart#configure-code-host-connections"
                rel="noreferrer noopener"
                target="_blank"
            >
                Create a new access token
            </a>{' '}
            with <code>write</code> permissions on the project and repository level.
        </>
    ),

    // These are just for type completeness and serve as placeholders for a bright future.
    [ExternalServiceKind.BITBUCKETCLOUD]: <span>Unsupported</span>,
    [ExternalServiceKind.GITOLITE]: <span>Unsupported</span>,
    [ExternalServiceKind.PERFORCE]: <span>Unsupported</span>,
    [ExternalServiceKind.PHABRICATOR]: <span>Unsupported</span>,
    [ExternalServiceKind.AWSCODECOMMIT]: <span>Unsupported</span>,
    [ExternalServiceKind.OTHER]: <span>Unsupported</span>,
}

export const AddCredentialModal: React.FunctionComponent<AddCredentialModalProps> = ({
    onCancel,
    afterCreate,
    history,
    userID,
    externalServiceKind,
    externalServiceURL,
}) => {
    const labelId = 'addCredential'
    const [isLoading, setIsLoading] = useState<boolean | Error>(false)
    const [credential, setCredential] = useState<string>('')

    const onChangeCredential = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setCredential(event.target.value)
    }, [])

    const onSubmit = useCallback<React.FormEventHandler>(
        async event => {
            event.preventDefault()
            setIsLoading(true)
            try {
                await createCampaignsCredential({ user: userID, credential, externalServiceKind, externalServiceURL })
                afterCreate()
            } catch (error) {
                setIsLoading(asError(error))
            }
        },
        [afterCreate, userID, credential, externalServiceKind, externalServiceURL]
    )

    return (
        <Dialog
            className="modal-body modal-body--top-third p-4 rounded border"
            onDismiss={onCancel}
            aria-labelledby={labelId}
        >
            <div className="web-content test-add-credential-modal">
                <h3 id={labelId}>
                    {modalTitles[externalServiceKind]} campaigns token for {externalServiceURL}
                </h3>
                {isErrorLike(isLoading) && <ErrorAlert error={isLoading} history={history} />}
                <Form onSubmit={onSubmit}>
                    <div className="form-group">
                        <label htmlFor="token">Personal access token</label>
                        <input
                            id="token"
                            name="token"
                            type="text"
                            className="form-control test-add-credential-modal-input"
                            required={true}
                            minLength={1}
                            value={credential}
                            onChange={onChangeCredential}
                        />
                        <p className="form-text">{helpTexts[externalServiceKind]}</p>
                    </div>
                    <div className="d-flex justify-content-end">
                        <button
                            type="button"
                            disabled={isLoading === true}
                            className="btn btn-outline-secondary mr-2"
                            onClick={onCancel}
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            disabled={isLoading === true || credential.length === 0}
                            className="btn btn-primary test-add-credential-modal-submit"
                        >
                            {isLoading === true && <LoadingSpinner className="icon-inline" />}
                            Add token
                        </button>
                    </div>
                </Form>
            </div>
        </Dialog>
    )
}
