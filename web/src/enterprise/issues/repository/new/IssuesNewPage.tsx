import React, { useCallback, useEffect, useState } from 'react'
import { Redirect } from 'react-router'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { mutateGraphQL } from '../../../../backend/graphql'
import { ModalPage } from '../../../../components/ModalPage'
import { PageTitle } from '../../../../components/PageTitle'
import { IssueForm, IssueFormData } from '../../form/IssueForm'
import { RepositoryIssuesAreaContext } from '../RepositoryIssuesArea'

export const createIssue = (input: GQL.ICreateIssueInput): Promise<GQL.IIssue> =>
    mutateGraphQL(
        gql`
            mutation CreateIssue($input: CreateIssueInput!) {
                createIssue(input: $input) {
                    id
                    url
                }
            }
        `,
        { input }
    )
        .pipe(
            map(dataOrThrowErrors),
            map(data => data.createIssue)
        )
        .toPromise()

interface Props extends Pick<RepositoryIssuesAreaContext, 'repo' | 'setBreadcrumbItem'> {}

const LOADING = 'loading' as const

/**
 * Shows a form to create a new issue.
 */
export const IssuesNewPage: React.FunctionComponent<Props> = ({ repo, setBreadcrumbItem }) => {
    useEffect(() => {
        if (setBreadcrumbItem) {
            setBreadcrumbItem({ text: 'New' })
        }
        return () => {
            if (setBreadcrumbItem) {
                setBreadcrumbItem(undefined)
            }
        }
    }, [setBreadcrumbItem])

    const [creationOrError, setCreationOrError] = useState<
        null | typeof LOADING | Pick<GQL.IIssue, 'url'> | ErrorLike
    >(null)
    const onSubmit = useCallback(
        async (data: IssueFormData) => {
            setCreationOrError(LOADING)
            try {
                setCreationOrError(
                    await createIssue({
                        ...data,
                        repository: repo.id,
                        // TODO!(sqs): dummy
                        baseRef: 'master~2',
                        headRef: 'master',
                    })
                )
            } catch (err) {
                setCreationOrError(asError(err))
                alert(err.message) // TODO!(sqs)
            }
        },
        [repo.id]
    )

    return (
        <>
            {creationOrError !== null && creationOrError !== LOADING && !isErrorLike(creationOrError) && (
                <Redirect to={creationOrError.url} />
            )}
            <PageTitle title="New issue" />
            <ModalPage>
                <h2>New issue</h2>
                <IssueForm
                    onSubmit={onSubmit}
                    buttonText="Create issue"
                    isLoading={creationOrError === LOADING}
                />
                {isErrorLike(creationOrError) && (
                    <div className="alert alert-danger mt-3">{creationOrError.message}</div>
                )}
            </ModalPage>
        </>
    )
}
