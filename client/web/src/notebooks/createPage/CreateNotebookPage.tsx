import React, { useMemo } from 'react'
import { Redirect } from 'react-router'
import { catchError, startWith } from 'rxjs/operators'
import * as uuid from 'uuid'

import { asError, isErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, useObservable, Alert } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { Page } from '../../components/Page'
import { CreateNotebookBlockInput } from '../../graphql-operations'
import { PageRoutes } from '../../routes.constants'
import { createNotebook } from '../backend'
import { blockToGQLInput, deserializeBlockInput } from '../serialize'

const LOADING = 'loading' as const

function deserializeBlocks(serializedBlocks: string): CreateNotebookBlockInput[] {
    return serializedBlocks.split(',').flatMap(serializedBlock => {
        const [type, encodedInput] = serializedBlock.split(':')
        if (type !== 'md' && type !== 'query' && type !== 'file' && type !== 'compute') {
            throw new Error(`Unknown block type: ${type}`)
        }
        if (type === 'compute') {
            return [] // compute blocks do not support deserialization yet.
        }
        const block = deserializeBlockInput(type, decodeURIComponent(encodedInput))
        return [blockToGQLInput({ id: uuid.v4(), ...block })]
    })
}

export const CreateNotebookPage: React.FunctionComponent<TelemetryProps & { authenticatedUser: AuthenticatedUser }> = ({
    telemetryService,
    authenticatedUser,
}) => {
    const notebookOrError = useObservable(
        useMemo(() => {
            const serializedBlocks = location.hash.trim().slice(1)
            const blocks = serializedBlocks.length > 0 ? deserializeBlocks(serializedBlocks) : []
            return createNotebook({
                notebook: { title: 'New Notebook', blocks, public: false, namespace: authenticatedUser.id },
            }).pipe(
                startWith(LOADING),
                catchError(error => [asError(error)])
            )
        }, [authenticatedUser])
    )

    if (notebookOrError && !isErrorLike(notebookOrError) && notebookOrError !== LOADING) {
        telemetryService.log('SearchNotebookCreated')
        return <Redirect to={PageRoutes.Notebook.replace(':id', notebookOrError.id)} />
    }

    return (
        <Page>
            {notebookOrError === LOADING && (
                <div className="d-flex justify-content-center">
                    <LoadingSpinner />
                </div>
            )}
            {isErrorLike(notebookOrError) && (
                <Alert variant="danger">
                    Error while creating the notebook: <strong>{notebookOrError.message}</strong>
                </Alert>
            )}
        </Page>
    )
}
