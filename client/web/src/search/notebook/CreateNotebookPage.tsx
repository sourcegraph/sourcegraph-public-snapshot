import React, { useMemo } from 'react'
import { Redirect } from 'react-router'
import { catchError, startWith } from 'rxjs/operators'
import * as uuid from 'uuid'

import { asError, isErrorLike } from '@sourcegraph/common'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { Page } from '../../components/Page'
import { CreateNotebookBlockInput, NotebookBlockType } from '../../graphql-operations'

import { createNotebook } from './backend'
import { deserializeBlockInput } from './serialize'

const LOADING = 'loading' as const

function deserializeBlocks(serializedBlocks: string): CreateNotebookBlockInput[] {
    return serializedBlocks.split(',').map(serializedBlock => {
        const [type, encodedInput] = serializedBlock.split(':')
        if (type !== 'md' && type !== 'query' && type !== 'file') {
            throw new Error(`Unknown block type: ${type}`)
        }
        const block = deserializeBlockInput(type, decodeURIComponent(encodedInput))
        switch (block.type) {
            case 'md':
                return { id: uuid.v4(), type: NotebookBlockType.MARKDOWN, markdownInput: block.input }
            case 'query':
                return { id: uuid.v4(), type: NotebookBlockType.QUERY, queryInput: block.input }
            case 'file':
                return { id: uuid.v4(), type: NotebookBlockType.FILE, fileInput: block.input }
        }
    })
}

export const CreateNotebookPage: React.FunctionComponent = () => {
    const notebookOrError = useObservable(
        useMemo(() => {
            const serializedBlocks = location.hash.trim().slice(1)
            const blocks = serializedBlocks.length > 0 ? deserializeBlocks(serializedBlocks) : []
            return createNotebook({ notebook: { title: 'New Notebook', blocks, public: false } }).pipe(
                startWith(LOADING),
                catchError(error => [asError(error)])
            )
        }, [])
    )

    if (notebookOrError && !isErrorLike(notebookOrError) && notebookOrError !== LOADING) {
        return <Redirect to={`/notebooks/${notebookOrError.id}`} />
    }

    return (
        <Page>
            {notebookOrError === LOADING && (
                <div className="d-flex justify-content-center">
                    <LoadingSpinner />
                </div>
            )}
            {isErrorLike(notebookOrError) && (
                <div className="alert alert-danger">
                    Error while creating the notebook: <strong>{notebookOrError.message}</strong>
                </div>
            )}
        </Page>
    )
}
