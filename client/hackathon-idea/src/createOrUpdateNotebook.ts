import { CreateNotebookVariables, UpdateNotebookVariables } from '../../web/src/graphql-operations'
import type { Package } from './types'
import { NotebookBlockType } from '@sourcegraph/shared/src/schema'

import * as packageDescriptions from '../descriptor/packages.json'

import fetch, { Headers, Response } from 'node-fetch'

// @ts-ignore
process.env['NODE_TLS_REJECT_UNAUTHORIZED'] = 0

const INSTANCE_URL = 'https://sourcegraph.test:3443/'

export async function createOrUpdateNotebook(
    notebookId: string | null,
    packageA: Package,
    packageB: Package
): Promise<string> {
    const notebook: CreateNotebookVariables['notebook'] = {
        title: `Mixing ${packageA} with ${packageB}`,
        namespace: 'VXNlcjox',
        public: false,
        blocks: [
            {
                id: '1',
                type: NotebookBlockType.MARKDOWN,
                markdownInput: `## ${packageA}
${getDescriptionForPackage(packageA)}

## ${packageB}
${getDescriptionForPackage(packageB)}

## Find npm packages that mix both:`,
            },
            {
                id: '2',
                type: NotebookBlockType.QUERY,
                queryInput: `\"${packageA}\":\s\"[0-9a-zA-Z-~^*.+><=|\s]+\" AND \"${packageB}\":\s\"[0-9a-zA-Z-~^*.+><=|\s]+\" file:^package\.json`,
            },
        ],
    }

    if (notebookId == null) {
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
            } as CreateNotebookVariables
        )
        console.log(`Created: ${INSTANCE_URL}notebooks/${response?.data?.createNotebook?.id}`)

        return response?.data?.createNotebook?.id
    } else {
        await requestGraphQL(
            `
            mutation createNotebook($notebook: NotebookInput!, $id: ID!) {
                updateNotebook(notebook: $notebook, id: $id) {
                    id
                }
            }
            `,
            {
                notebook,
                id: notebookId,
            } as UpdateNotebookVariables
        )
        console.log(`Updated: ${INSTANCE_URL}notebooks/${notebookId}`)

        return notebookId
    }
}

async function requestGraphQL(query: string, variables: CreateNotebookVariables | UpdateNotebookVariables) {
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

function getDescriptionForPackage(pkg: Package): string {
    // @ts-ignore
    return packageDescriptions.default.find((p: any) => p.name === pkg)?.description
}
