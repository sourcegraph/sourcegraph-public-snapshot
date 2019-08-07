import React, { useCallback, useEffect, useState } from 'react'
import { Redirect } from 'react-router'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { mutateGraphQL } from '../../../../backend/graphql'
import { ModalPage } from '../../../../components/ModalPage'
import { PageTitle } from '../../../../components/PageTitle'
import { ThreadForm, ThreadFormData } from '../../form/ThreadForm'
import { RepositoryThreadsAreaContext } from '../RepositoryThreadsArea'

export const createThread = (input: GQL.ICreateThreadInput): Promise<GQL.IThread> =>
    mutateGraphQL(
        gql`
            mutation CreateThread($input: CreateThreadInput!) {
                createThread(input: $input) {
                    id
                    url
                }
            }
        `,
        { input }
    )
        .pipe(
            map(dataOrThrowErrors),
            map(data => data.createThread)
        )
        .toPromise()

interface Props extends Pick<RepositoryThreadsAreaContext, 'repo' | 'setBreadcrumbItem'> {}

const LOADING = 'loading' as const

/**
 * Shows a form to create a new thread.
 */
export const ThreadsNewPage: React.FunctionComponent<Props> = ({ repo, setBreadcrumbItem }) => {
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
        null | typeof LOADING | Pick<GQL.IThread, 'url'> | ErrorLike
    >(null)
    const onSubmit = useCallback(
        async (data: ThreadFormData) => {
            setCreationOrError(LOADING)
            try {
                setCreationOrError(
                    await createThread({
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
            <PageTitle title="New thread" />
            <ModalPage>
                <h2>New thread</h2>
                <ThreadForm
                    onSubmit={onSubmit}
                    buttonText="Create thread"
                    isLoading={creationOrError === LOADING}
                />
                {isErrorLike(creationOrError) && (
                    <div className="alert alert-danger mt-3">{creationOrError.message}</div>
                )}
            </ModalPage>
        </>
    )
}
