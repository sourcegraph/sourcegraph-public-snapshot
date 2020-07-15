import React, { useState, useCallback, FormEvent } from 'react'
import { Form } from '../../../components/Form'
import { mutateGraphQL, queryGraphQL } from '../../../backend/graphql'
import { gql, dataOrThrowErrors } from '../../../../../shared/src/graphql/graphql'
import { ID } from '../../../../../shared/src/graphql/schema'
import { RepoNotFoundError } from '../../../../../shared/src/backend/errors'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { asError } from '../../../../../shared/src/util/errors'
import { ErrorAlert } from '../../../components/alerts'
import * as H from 'history'

async function addChangeset({
    campaignID,
    repoName,
    externalID,
}: {
    campaignID: ID
    repoName: string
    externalID: string
}): Promise<void> {
    const repository = dataOrThrowErrors(
        await queryGraphQL(
            gql`
                query RepositoryID($repoName: String!) {
                    repository(name: $repoName) {
                        id
                    }
                }
            `,
            { repoName }
        ).toPromise()
    ).repository
    if (!repository) {
        throw new RepoNotFoundError(repoName)
    }

    const changeset = dataOrThrowErrors(
        await mutateGraphQL(
            gql`
                mutation CreateChangeset($repositoryID: ID!, $externalID: String!) {
                    createChangesets(input: { repository: $repositoryID, externalID: $externalID }) {
                        id
                    }
                }
            `,
            { repositoryID: repository.id, externalID }
        ).toPromise()
    ).createChangesets[0]

    dataOrThrowErrors(
        await mutateGraphQL(
            gql`
                mutation AddChangeSetToCampaign($campaignID: ID!, $changesets: [ID!]!) {
                    addChangesetsToCampaign(campaign: $campaignID, changesets: $changesets) {
                        id
                    }
                }
            `,
            { campaignID, changesets: [changeset.id] }
        ).toPromise()
    )
}

/**
 * Simple, temporary form to add changesets.
 */
export const AddChangesetForm: React.FunctionComponent<{ campaignID: ID; onAdd: () => void; history: H.History }> = ({
    campaignID,
    onAdd,
    history,
}) => {
    const [error, setError] = useState<Error>()
    const [repoName, setRepoName] = useState('')
    const [externalID, setExternalID] = useState('')
    const [isLoading, setIsLoading] = useState(false)
    const submit = useCallback(
        async (event: FormEvent) => {
            event.preventDefault()
            try {
                setIsLoading(true)
                setError(undefined)
                await addChangeset({ campaignID, repoName, externalID })
                setExternalID('')
                onAdd()
            } catch (error) {
                setError(asError(error))
            } finally {
                setIsLoading(false)
            }
        },
        [campaignID, externalID, onAdd, repoName, setError]
    )
    return (
        <>
            <h3 className="mb-2 mt-4">Add changeset</h3>
            <Form onSubmit={submit}>
                <div className="d-flex">
                    <div className="form-group mr-3 mb-0">
                        <label htmlFor="changeset-repo">Repository path</label>
                        <input
                            required={true}
                            id="changeset-repo"
                            type="text"
                            size={35}
                            className="form-control mr-1 test-track-changeset-repo"
                            placeholder="codehost.example.com/example-org/example-repository"
                            value={repoName}
                            onChange={event => setRepoName(event.target.value)}
                        />
                        <p className="form-text text-muted">
                            The location of the repository in Sourcegraph. See the repository's directory URL:{' '}
                            {/* False positive */}
                            {/* eslint-disable-next-line react/jsx-no-comment-textnodes */}
                            {window.location.protocol}//
                            {window.location.host}/<strong>&lt;REPOSITORY_PATH&gt;</strong>
                            <br />
                            (Depending on your instance's configuration the path may or may not contain the code host
                            name)
                        </p>
                    </div>
                    <div className="form-group mr-3 mb-0">
                        <label htmlFor="changeset-number">Changeset number</label>
                        <input
                            required={true}
                            id="changeset-number"
                            type="number"
                            min={1}
                            step={1}
                            size={16}
                            className="form-control mr-1 test-track-changeset-id"
                            placeholder="1234"
                            value={externalID}
                            onChange={event => setExternalID(event.target.value + '')}
                        />
                    </div>
                </div>
                <button type="submit" className="btn btn-primary mr-1 test-track-changeset-btn" disabled={isLoading}>
                    Add changeset
                    {isLoading && <LoadingSpinner className="ml-2 icon-inline" />}
                </button>
            </Form>
            {error && <ErrorAlert error={error} className="mt-2" history={history} />}
        </>
    )
}
