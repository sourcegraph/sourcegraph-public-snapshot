import React, { useState, useCallback, FormEvent } from 'react'
import { Form } from '../../../components/Form'
import { mutateGraphQL, queryGraphQL } from '../../../backend/graphql'
import { gql, dataOrThrowErrors } from '../../../../../shared/src/graphql/graphql'
import { ID } from '../../../../../shared/src/graphql/schema'
import { RepoNotFoundError } from '../../../../../shared/src/backend/errors'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { asError } from '../../../../../shared/src/util/errors'
import { ErrorAlert } from '../../../components/alerts'

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
                mutation CreateChangeSet($repositoryID: ID!, $externalID: String!) {
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
export const AddChangesetForm: React.FunctionComponent<{ campaignID: ID; onAdd: () => void }> = ({
    campaignID,
    onAdd,
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
            } catch (err) {
                setError(asError(err))
            } finally {
                setIsLoading(false)
            }
        },
        [campaignID, externalID, onAdd, repoName, setError]
    )
    const onChangeRepoName = (repoName: string): void => {
        try {
            const url = new URL(repoName)
            // match for github and bitbucket pr urls
            const githubRegex = /\/pull\/([0-9]+).*$/
            const bitbucketRegex = /\/pull-requests\/([0-9]+).*$/
            const githubMatches = githubRegex.exec(url.pathname)
            if (githubMatches) {
                setRepoName(url.hostname + url.pathname.substring(0, url.pathname.indexOf('/pull/')))
                if (githubMatches[1]) {
                    setExternalID(githubMatches[1])
                }
            } else {
                const bitbucketMatches = bitbucketRegex.exec(url.pathname)
                if (bitbucketMatches) {
                    setRepoName(url.hostname + url.pathname.substring(0, url.pathname.indexOf('/pull-requests/')))
                    if (bitbucketMatches[1]) {
                        setExternalID(bitbucketMatches[1])
                    }
                } else {
                    setRepoName(repoName)
                }
            }
        } catch {
            setRepoName(repoName)
        }
    }
    return (
        <>
            <h3>Track changeset</h3>
            <Form onSubmit={submit}>
                <div className="d-flex">
                    <div className="form-group mr-3 mb-0">
                        <label htmlFor="changeset-repo">Repository name</label>
                        <input
                            required={true}
                            name="changeset-repo"
                            type="text"
                            size={35}
                            className="form-control mr-1"
                            placeholder="Repository name"
                            value={repoName}
                            onChange={event => onChangeRepoName(event.target.value)}
                        />
                        <p className="form-text text-muted">e.g., codehost/organization/repository</p>
                    </div>
                    <div className="form-group mr-3 mb-0">
                        <label htmlFor="changeset-number">Changeset number</label>
                        <input
                            required={true}
                            name="changeset-number"
                            type="number"
                            min={1}
                            step={1}
                            size={16}
                            className="form-control mr-1"
                            placeholder="Changeset number"
                            value={externalID}
                            onChange={event => setExternalID(event.target.value + '')}
                        />
                    </div>
                </div>
                <button type="submit" className="btn btn-primary mr-1">
                    Add changeset
                    {isLoading && <LoadingSpinner className="ml-2 icon-inline" />}
                </button>
            </Form>
            {error && <ErrorAlert error={error} className="mt-3" />}
        </>
    )
}
