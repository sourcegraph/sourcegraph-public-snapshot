import { execFile } from 'child_process'
import * as path from 'path'

export async function fastFilesExist(
    rgPath: string,
    rootPath: string,
    filePaths: string[]
): Promise<{ [filePath: string]: boolean }> {
    const searchPath =
        '{' +
        filePaths
            .flatMap(filePath => {
                let pathChunk = filePath
                while (
                    pathChunk.endsWith('*') ||
                    pathChunk.endsWith(path.sep) ||
                    pathChunk.startsWith('*') ||
                    pathChunk.startsWith(path.sep)
                ) {
                    const trimToks = ['**', '*', path.sep]
                    for (const trimTok of trimToks) {
                        if (pathChunk.startsWith(trimTok)) {
                            pathChunk = pathChunk.slice(trimTok.length)
                        }
                        if (pathChunk.endsWith(trimTok)) {
                            pathChunk = pathChunk.slice(0, -trimTok.length)
                        }
                    }
                }
                return [`**${path.sep}${pathChunk}${path.sep}**`, `**${path.sep}${pathChunk}`]
            })
            .join(',') +
        '}'
    const out = await new Promise<string>((resolve, reject) => {
        execFile(
            rgPath,
            ['--files', '-g', searchPath, '--crlf', '--fixed-strings', '--no-config', '--no-ignore-global'],
            {
                cwd: rootPath,
                maxBuffer: 1024 * 1024 * 1024,
            },
            (error, stdout, stderr) => {
                if (error?.code === 2) {
                    reject(new Error(`${error.message}: ${stderr}`))
                } else {
                    resolve(stdout)
                }
            }
        )
    })
    const unvalidatedPaths = new Set<string>(filePaths)
    for (const line of out.split('\n')) {
        const realFile = line.trim()
        for (const filePath of [...unvalidatedPaths]) {
            if (filePathContains(realFile, filePath)) {
                unvalidatedPaths.delete(filePath)
            }
        }
        if (unvalidatedPaths.size === 0) {
            break
        }
    }

    const ret: { [filePath: string]: boolean } = {}
    for (const filePath of filePaths) {
        ret[filePath] = !unvalidatedPaths.has(filePath)
    }

    return ret
}

export function filePathContains(container: string, contained: string): boolean {
    let trimmedContained = contained
    if (trimmedContained.endsWith(path.sep)) {
        trimmedContained = trimmedContained.slice(0, -path.sep.length)
    }
    if (trimmedContained.startsWith(path.sep)) {
        trimmedContained = trimmedContained.slice(path.sep.length)
    }
    if (trimmedContained.startsWith('.' + path.sep)) {
        trimmedContained = trimmedContained.slice(1 + path.sep.length)
    }
    return (
        container === contained || // exact match
        container === path.sep + trimmedContained ||
        container === trimmedContained ||
        container.startsWith(trimmedContained + path.sep) || // relative parent directory
        container.includes(path.sep + trimmedContained + path.sep) || // mid-level directory
        container.endsWith(path.sep + trimmedContained) // child
    )
}
