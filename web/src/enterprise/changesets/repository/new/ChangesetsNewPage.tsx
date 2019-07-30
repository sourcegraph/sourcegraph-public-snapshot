import React, { useCallback, useEffect, useState } from 'react'
import { Redirect } from 'react-router'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { mutateGraphQL } from '../../../../backend/graphql'
import { ModalPage } from '../../../../components/ModalPage'
import { PageTitle } from '../../../../components/PageTitle'
import { ChangesetForm, ChangesetFormData } from '../../form/ChangesetForm'
import { RepositoryChangesetsAreaContext } from '../RepositoryChangesetsArea'

const createChangeset = (input: GQL.ICreateChangesetInput): Promise<GQL.IChangeset> =>
    mutateGraphQL(
        gql`
            mutation CreateChangeset($input: CreateChangesetInput!) {
                createChangeset(input: $input) {
                    id
                    url
                }
            }
        `,
        { input }
    )
        .pipe(
            map(dataOrThrowErrors),
            map(data => data.createChangeset)
        )
        .toPromise()

interface Props extends Pick<RepositoryChangesetsAreaContext, 'repo' | 'setBreadcrumbItem'> {}

const LOADING = 'loading' as const

/**
 * Shows a form to create a new changeset.
 */
export const ChangesetsNewPage: React.FunctionComponent<Props> = ({ repo, setBreadcrumbItem }) => {
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
        null | typeof LOADING | Pick<GQL.IChangeset, 'url'> | ErrorLike
    >(null)
    const onSubmit = useCallback(
        async (data: ChangesetFormData) => {
            setCreationOrError(LOADING)
            try {
                setCreationOrError(
                    await createChangeset({
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
            <PageTitle title="New changeset" />
            <ModalPage>
                <h2>New changeset</h2>
                <ChangesetForm
                    onSubmit={onSubmit}
                    buttonText="Create changeset"
                    isLoading={creationOrError === LOADING}
                />
                {isErrorLike(creationOrError) && (
                    <div className="alert alert-danger mt-3">{creationOrError.message}</div>
                )}
            </ModalPage>
        </>
    )
}
