import { createWriteStream } from 'fs'
import { pipeline, Readable } from 'stream'
import { promisify } from 'util'
import { createGzip } from 'zlib'

import fetch from 'node-fetch'

import { COMPRESSED_SCHEMA_PATH, OLDEST_VERSION_TAG } from './constants'

// Downloads, creates, and stores the oldest supported graphql schema:
// - Use SG API to search for all graphql files w/ content
// - Stitch together graphql file(s) content with simple `\n`
// - Compress with gzip
// - Write file in __schemas__ (name TDB) directory in `/tests` directory
async function main(): Promise<void> {
    // Fetch all graphql files from the oldest supported Sourcegraph version.
    const result = await fetch('https://sourcegraph.com/.api/graphql', {
        method: 'post',
        body: JSON.stringify({
            query: `
            query GQLSchemaFiles($query:String!) {
                search(query: $query) {
                  results {
                    results{
                      ... on FileMatch {
                        file {
                          path,
                          content
                        }
                      }
                    }
                  }
                }
              }`,
            variables: {
                query: `repo:github.com/sourcegraph/sourcegraph$ lang:graphql rev:${OLDEST_VERSION_TAG}`,
            },
        }),
    }).then(
        (
            response
        ): Promise<{ data: { search: { results: { results: { file: { path: string; content: string } }[] } } } }> =>
            response.json()
    )

    // Stitch schema files together and store compressed version in `client/vscode/tests`
    const stitchedSchema = result.data.search.results.results.reduce(
        (stitched, { file }) => (stitched += `\n${file.content}`),
        ''
    )

    // ~66kb gzipped for v3.31.2
    await compressStringToFile(stitchedSchema, COMPRESSED_SCHEMA_PATH)
}

const pipe = promisify(pipeline)

async function compressStringToFile(content: string, outputPath: string): Promise<void> {
    const gzip = createGzip()
    const source = new Readable({
        read() {
            this.push(content)
            this.push(null)
        },
    })
    const destination = createWriteStream(outputPath)

    await pipe(source, gzip, destination)
}

main().catch(error => console.error(error))
