import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React, { useCallback, useMemo, useState } from 'react'
import { map } from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { pluralize } from '../../../../../shared/src/util/strings'
import { queryGraphQL } from '../../../backend/graphql'
import { ProjectAreaContext } from '../ProjectArea'
import { LabelRow } from './LabelRow'
import { NewLabelForm } from './NewLabelForm'

const queryProjectLabels = (project: GQL.ID): Promise<GQL.ILabelConnection> =>
    queryGraphQL(
        gql`
            query ProjectLabels($project: ID!) {
                node(id: $project) {
                    __typename
                    ... on Project {
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
        { project }
    )
        .pipe(
            map(({ data, errors }) => {
                if (
                    !data ||
                    !data.node ||
                    data.node.__typename !== 'Project' ||
                    !data.node.labels ||
                    (errors && errors.length > 0)
                ) {
                    throw createAggregateError(errors)
                }
                return data.node.labels
            })
        )
        .toPromise()

const LOADING: 'loading' = 'loading'

interface Props extends Pick<ProjectAreaContext, 'project'>, ExtensionsControllerProps {}

/**
 * Lists a project's labels.
 */
export const ProjectLabelsPage: React.FunctionComponent<Props> = ({ project, ...props }) => {
    const [labelsOrError, setLabelsOrError] = useState<typeof LOADING | GQL.ILabelConnection | ErrorLike>(LOADING)
    const loadLabels = useCallback(async () => {
        setLabelsOrError(LOADING)
        try {
            setLabelsOrError(await queryProjectLabels(project.id))
        } catch (err) {
            setLabelsOrError(asError(err))
        }
    }, [project])
    // tslint:disable-next-line: no-floating-promises
    useMemo(loadLabels, [project])

    const [isShowingNewLabelForm, setIsShowingNewLabelForm] = useState(false)
    const toggleIsShowingNewLabelForm = useCallback(() => setIsShowingNewLabelForm(!isShowingNewLabelForm), [
        isShowingNewLabelForm,
    ])

    return (
        <div className="project-labels-page container mt-2">
            <div className="d-flex align-items-center justify-content-between mb-3">
                <h2 className="mb-0">Labels</h2>
                <button type="button" className="btn btn-success" onClick={toggleIsShowingNewLabelForm}>
                    New label
                </button>
            </div>
            {isShowingNewLabelForm && (
                <NewLabelForm
                    project={project}
                    onDismiss={toggleIsShowingNewLabelForm}
                    onLabelCreate={loadLabels}
                    className="my-3 p-2 border rounded"
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
                                <li key={label.id} className="list-group-item p-2">
                                    <LabelRow {...props} label={label} onLabelUpdate={loadLabels} />
                                </li>
                            ))}
                        </ul>
                    ) : (
                        <div className="p-2 text-muted">No labels yet.</div>
                    )}
                </div>
            )}
        </div>
    )
}
