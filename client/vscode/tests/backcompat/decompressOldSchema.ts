import { exists } from 'fs'
import { promisify } from 'util'
import { InputType, unzip, ZlibOptions } from 'zlib'

import { readFile, writeFile } from 'mz/fs'

import { COMPRESSED_SCHEMA_PATH, DECOMPRESSED_SCHEMA_PATH } from './constants'

const unzipAsPromise: (buffer: InputType, options?: ZlibOptions) => Promise<Buffer> = promisify(unzip)
const existsAsPromise = promisify(exists)

// Decompresses oldest supported graphql schema for use in tests.
// Writes to file if it doesn't already exist, then returns
// the schema as a string?
export async function decompressSchema(): Promise<string> {
    if (!(await existsAsPromise(DECOMPRESSED_SCHEMA_PATH))) {
        const content = await readFile(COMPRESSED_SCHEMA_PATH, 'base64')
        const result = await unzipAsPromise(Buffer.from(content, 'base64'))
        await writeFile(DECOMPRESSED_SCHEMA_PATH, result)
        return result.toString()
    }
    return readFile(DECOMPRESSED_SCHEMA_PATH, 'utf8')
}
