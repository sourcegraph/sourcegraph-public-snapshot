// @ts-check

const { visit } = require('graphql')

/**
 * Custom version of the `typed-document-node` plugin that only generates the type definitions
 * for the documents without generating a parsed version. This is all we need because the
 * `@rollup/plugin-graphql` plugin already parses the documents for us.
 *
 * @param {import('graphql').GraphQLSchema} _schema
 * @param {import('@graphql-codegen/plugin-helpers').Types.DocumentFile[]} documents
 * @param {{operationResultSuffix?: string}} config
 */
const plugin = (_schema, documents, config) => {
    const { operationResultSuffix = '' } = config

    /** @type {{name: string}[]} */
    const allOperations = []

    for (const item of documents) {
        if (item.document) {
            visit(item.document, {
                OperationDefinition: {
                    enter: node => {
                        if (node.name && node.name.value) {
                            allOperations.push({
                                name: node.name.value,
                            })
                        }
                    },
                },
            })
        }
    }

    const documentNodes = allOperations.map(
        ({ name }) =>
            `export declare const ${name}: TypedDocumentNode<${name}${operationResultSuffix}, ${name}Variables>`
    )
    if (documentNodes.length === 0) {
        return ''
    }
    return {
        prepend: ["import type { TypedDocumentNode } from '@graphql-typed-document-node/core'\n"],
        content: documentNodes.join('\n'),
    }
}

module.exports = { plugin }
