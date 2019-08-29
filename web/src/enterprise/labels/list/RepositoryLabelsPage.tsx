import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React, { useCallback, useMemo, useState, useEffect } from 'react'
import { map } from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { pluralize } from '../../../../../shared/src/util/strings'
import { queryGraphQL } from '../../../backend/graphql'
import { RepoContainerContext } from '../../../repo/RepoContainer'
import { LabelRow } from './LabelRow'
import { NewLabelForm } from './NewLabelForm'

const queryRepositoryLabels = (repository: GQL.ID): Promise<GQL.ILabelConnection> =>
    queryGraphQL(
        gql`
            query RepositoryLabels($repository: ID!) {
                node(id: $repository) {
                    __typename
                    ... on Repository {
                        labels {
                            nodes {
                                id
                                name
                                description
                                color
                            }
                            totalCount
                        }
                    }
                }
            }
        `,
        { repository }
    )
        .pipe(
            map(dataOrThrowErrors),
            map(data => {
                if (!data.node || data.node.__typename !== 'Repository') {
                    throw new Error('not a repository')
                }
                return data.node.labels
            })
        )
        .toPromise()

const LOADING: 'loading' = 'loading'

interface Props extends Pick<RepoContainerContext, 'repo'>, ExtensionsControllerProps {}

/**
 * Lists a repository's labels.
 */
export const RepositoryLabelsPage: React.FunctionComponent<Props> = ({ repo: repository, ...props }) => {
    const [labelsOrError, setLabelsOrError] = useState<typeof LOADING | GQL.ILabelConnection | ErrorLike>(LOADING)
    const loadLabels = useCallback(async () => {
        setLabelsOrError(LOADING)
        try {
            setLabelsOrError(await queryRepositoryLabels(repository.id))
        } catch (err) {
            setLabelsOrError(asError(err))
        }
    }, [repository])
    useEffect(loadLabels, [repository])

    const [isShowingNewLabelForm, setIsShowingNewLabelForm] = useState(false)
    const toggleIsShowingNewLabelForm = useCallback(() => setIsShowingNewLabelForm(!isShowingNewLabelForm), [
        isShowingNewLabelForm,
    ])

    return (
        <div className="repository-labels-page container mt-4">
            <div className="d-flex align-items-center justify-content-between mb-3">
                <h2 className="mb-0">Labels</h2>
                <button type="button" className="btn btn-success" onClick={toggleIsShowingNewLabelForm}>
                    New label
                </button>
            </div>
            {isShowingNewLabelForm && (
                <NewLabelForm
                    repository={repository}
                    onDismiss={toggleIsShowingNewLabelForm}
                    onLabelCreate={loadLabels}
                    className="my-3 p-3 border"
                />
            )}
            {labelsOrError === LOADING ? (
                <LoadingSpinner className="icon-inline mt-3" />
            ) : isErrorLike(labelsOrError) ? (
                <div className="alert alert-danger mt-3">{labelsOrError.message}</div>
            ) : (
                <div className="card">
                    <div className="card-header">
                        <span className="text-muted">
                            {labelsOrError.totalCount} {pluralize('label', labelsOrError.totalCount)}
                        </span>
                    </div>
                    {labelsOrError.nodes.length > 0 ? (
                        <ul className="list-group list-group-flush">
                            {labelsOrError.nodes.map(label => (
                                <li key={label.id} className="list-group-item p-3">
                                    <LabelRow {...props} label={label} onLabelUpdate={loadLabels} />
                                </li>
                            ))}
                        </ul>
                    ) : (
                        <div className="p-3 text-muted">No labels yet.</div>
                    )}
                </div>
            )}
        </div>
    )
}
