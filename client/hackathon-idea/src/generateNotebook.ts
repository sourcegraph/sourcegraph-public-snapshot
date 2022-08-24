import { CreateNotebookVariables } from '../../web/src/graphql-operations'
import type { Package } from './types'
import { NotebookBlockType } from '@sourcegraph/shared/src/schema'

import fetch, { Headers, Response } from 'node-fetch'

// @ts-ignore
process.env['NODE_TLS_REJECT_UNAUTHORIZED'] = 0

const INSTANCE_URL = 'https://sourcegraph.test:3443/'

export async function generateNotebook(packageA: Package, packageB: Package): Promise<void> {
    const notebook: CreateNotebookVariables['notebook'] = {
        title: 'Test page',
        namespace: 'VXNlcjox',
        public: false,
        blocks: [
            {
                id: '1',
                type: NotebookBlockType.MARKDOWN,
                markdownInput: '# Packages that use ' + packageA + ' and ' + packageB,
            },
            {
                id: '2',
                type: NotebookBlockType.QUERY,
                queryInput: `${packageA} AND ${packageB}`,
            },
        ],
    }

    const response = await requestGraphQL(
        `
        mutation createNotebook($notebook: NotebookInput!) {
            createNotebook(notebook: $notebook) {
                id
            }
        }
        `,
        {
            notebook,
        }
    )
    console.log(`Created: ${INSTANCE_URL}notebooks/${response?.data?.createNotebook?.id}`)
}

async function requestGraphQL(query: string, variables: CreateNotebookVariables) {
    const headers = new Headers()
    headers.set('Content-Type', 'application/json')
    headers.set('Authorization', `token ${process.env.TOKEN}`)
    let response: Response | null = null

    try {
        response = await fetch(INSTANCE_URL + '.api/graphql', {
            body: JSON.stringify({
                query,
                variables,
            }),
            method: 'POST',
            headers,
        })
    } catch (error) {
        console.log('Error requesting GraphQL', error, response)
        throw new Error(error as any)
    }

    if (!response || !response.ok) {
        throw new Error(`GraphQL request failed: ${response.status} ${response.statusText}\n${await response.text()}`)
    }

    return response.json()
}
