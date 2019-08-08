import split2 from 'split2'
import { fs } from 'mz'
import { validate } from 'jsonschema'
import { readEnvInt } from './env'

/**
 * Limit on the file size accepted by the /upload endpoint. Defaults to 100MB.
 */
export const MAX_FILE_SIZE = readEnvInt({ key: 'LSIF_MAX_FILE_SIZE', defaultValue: 100 * 1024 * 1024 })

/**
 * Limit the size of each line for a JSON-line encoded LSIF dump. Defaults to 1MB.
 */
export const MAX_LINE_SIZE = readEnvInt({ key: 'LSIF_MAX_LINE_SIZE', defaultValue: 1024 * 1024 })

/**
 * Validate each line in the temp file against the expected schema. This reads
 * the temp file line-by-line (with a max size to limit memory growth).
 */
export function validateContent(tempPath: string, schema: any): Promise<void> {
    let lineno = 0
    const validateLine = (line: string, reject: (_: any) => void) => {
        let data: any
        try {
            data = JSON.parse(line)
        } catch (e) {
            reject(Object.assign(new Error(`Malformed JSON on line ${lineno}`), { status: 422 }))
            return
        }

        const result = validate(data, schema)
        if (result.errors.length > 0) {
            reject(
                Object.assign(new Error(`Invalid JSON data on line ${lineno}: ${result.errors.join(', ')}`), {
                    status: 422,
                })
            )
        }

        lineno++
    }

    return new Promise((resolve, reject) => {
        fs.createReadStream(tempPath)
            .pipe(split2({ maxLength: MAX_LINE_SIZE }))
            .on('data', line => validateLine(line, reject))
            .on('close', resolve)
            .on('error', reject)
    })
}

/**
 * Throws an error with status 400 if the repository is invalid.
 */
export function checkRepository(repository: any): void {
    if (typeof repository !== 'string') {
        throw Object.assign(new Error('Must specify the repository (usually of the form github.com/user/repo)'), {
            status: 400,
        })
    }
}

/**
 * Throws an error with status 400 if the commit is invalid.
 */
export function checkCommit(commit: any): void {
    if (typeof commit !== 'string' || commit.length !== 40 || !/^[0-9a-f]+$/.test(commit)) {
        throw Object.assign(new Error('Must specify the commit as a 40 character hash ' + commit), { status: 400 })
    }
}

/**
 * Throws an error with status 413 if the content length is too large.
 */
export function checkContentLength(rawContentLength: string | undefined): void {
    if (rawContentLength && parseInt(rawContentLength || '', 10) > MAX_FILE_SIZE) {
        throw Object.assign(
            new Error(
                `The size of the given LSIF file (${rawContentLength} bytes) exceeds the max of ${MAX_FILE_SIZE}`
            ),
            { status: 413 }
        )
    }
}

/**
 * Throws an error with status 422 if the method is not supported by the
 * current backend.
 */
export function checkMethod(method: string, supportedMethods: string[]): void {
    if (!supportedMethods.includes(method)) {
        throw Object.assign(new Error(`Method must be one of ${Array.from(supportedMethods).join(', ')}`), {
            status: 422,
        })
    }
}
