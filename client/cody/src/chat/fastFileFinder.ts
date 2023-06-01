import { execFile } from 'child_process'
import * as path from 'path'

/**
 * Checks whether the specified file paths exist within a root path using the 'rg' tool.
 *
 * @param rgPath - The path to the 'rg' tool.
 * @param rootPath - The path of the directory to be searched.
 * @param filePaths - The file paths to search for.
 * @returns An object that maps each file path to a boolean indicating whether the file was found.
 */
export async function fastFilesExist(
    rgPath: string,
    rootPath: string,
    filePaths: string[]
): Promise<{ [filePath: string]: boolean }> {
    const searchPattern = constructSearchPattern(filePaths)
    const rgOutput = await executeRg(rgPath, rootPath, searchPattern)
    return processRgOutput(rgOutput, filePaths)
}

export function makeTrimRegex(sep: string): RegExp {
    if (sep === '\\') {
        sep = '\\\\' // Regex escape this character
    }
    return new RegExp(`(^[*${sep}]+)|([*${sep}]+$)`, 'g')
}

// Regex to match '**', '*' or path.sep at the start (^) or end ($) of the string.
const trimRegex = makeTrimRegex(path.sep)
/**
 * Constructs a search pattern for the 'rg' tool.
 *
 * @param filePaths - The file paths to include in the pattern.
 * @returns The search pattern.
 */
function constructSearchPattern(filePaths: string[]): string {
    const searchPatternParts = filePaths.map(filePath => {
        const pathChunk = filePath.replace(trimRegex, '')
        // Create a pattern that matches any file that ends with the specified filePath
        return `**${path.sep}${pathChunk}${path.sep}**,**${path.sep}${pathChunk}`
    })
    return `{${searchPatternParts.join(',')}}`
}
/**
 * Executes the 'rg' tool and returns the output.
 *
 * @param rgPath - The path to the 'rg' tool.
 * @param rootPath - The path of the directory to be searched.
 * @param searchPattern - The search pattern to use.
 * @returns The output from the 'rg' tool.
 */
async function executeRg(rgPath: string, rootPath: string, searchPattern: string): Promise<string> {
    return new Promise((resolve, reject) => {
        execFile(
            rgPath,
            ['--files', '-g', searchPattern, '--crlf', '--fixed-strings', '--no-config', '--no-ignore-global'],
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
}

/**
 * Processes the output from the 'rg' tool to find matching file paths.
 *
 * @param rgOutput - The output from the 'rg' tool.
 * @param filePaths - The file paths to search for.
 * @returns An object that maps each file path to a boolean indicating whether the file was found.
 */
function processRgOutput(rgOutput: string, filePaths: string[]): { [filePath: string]: boolean } {
    const unvalidatedPaths = new Set<string>(filePaths)
    const filePathsExist: { [filePath: string]: boolean } = {}
    for (const line of rgOutput.split('\n')) {
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
    for (const filePath of filePaths) {
        filePathsExist[filePath] = !unvalidatedPaths.has(filePath)
    }
    return filePathsExist
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
