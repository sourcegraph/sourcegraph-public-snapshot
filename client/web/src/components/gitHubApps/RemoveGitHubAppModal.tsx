import React, { useCallback } from 'react'

import { useMutation } from '@sourcegraph/http-client'
import { Button, Modal, Text, ErrorAlert, H3, AnchorLink, Alert } from '@sourcegraph/wildcard'

import type { DeleteGitHubAppResult, DeleteGitHubAppVariables, GitHubAppByIDFields } from '../../graphql-operations'
import { LoaderButton } from '../LoaderButton'

import { DELETE_GITHUB_APP_BY_ID_QUERY } from './backend'

export interface RemoveGitHubAppModalProps {
    app: Pick<GitHubAppByIDFields, 'id' | 'name' | 'appURL'>
    onCancel: () => void
    afterDelete: () => void
}

export const RemoveGitHubAppModal: React.FunctionComponent<React.PropsWithChildren<RemoveGitHubAppModalProps>> = ({
    app,
    onCancel,
    afterDelete,
}) => {
    const labelId = 'removeGitHubApp'
    const [deleteGitHubApp, { loading, error }] = useMutation<DeleteGitHubAppResult, DeleteGitHubAppVariables>(
        DELETE_GITHUB_APP_BY_ID_QUERY
    )

    const onDelete = useCallback<React.MouseEventHandler>(async () => {
        await deleteGitHubApp({ variables: { gitHubApp: app.id } })
        afterDelete()
    }, [afterDelete, app.id, deleteGitHubApp])

    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelId}>
            <H3 className="mb-3">Remove the GitHub App "{app.name}"?</H3>
            {error && <ErrorAlert error={error} />}
            <Alert variant="warning">
                This will remove the App from Sourcegraph, but it will still exist on GitHub.
            </Alert>
            <Text>While not necessary, if you wish to completely remove the App on GitHub, you must:</Text>
            <ul>
                <li>Uninstall it from the individual user(s) or organization(s) where it is installed, and/or</li>
                <li>
                    {/* TODO: We could route this directly to the Advanced settings page once we can distinguish organization apps from user apps. */}
                    Delete the App entirely.
                </li>
            </ul>

            <Text>
                <AnchorLink to={app.appURL} target="_blank" rel="noopener noreferrer">
                    View the App on GitHub
                </AnchorLink>{' '}
                to uninstall or delete it. You must be an owner of the App on GitHub in order to perform these actions.
            </Text>
            <div className="d-flex justify-content-end pt-1">
                <Button disabled={loading} className="mr-2" onClick={onCancel} outline={true} variant="secondary">
                    Cancel
                </Button>
                <LoaderButton
                    disabled={loading}
                    onClick={onDelete}
                    variant="danger"
                    loading={loading}
                    alwaysShowLabel={true}
                    label="Remove GitHub App"
                />
            </div>
        </Modal>
    )
}
