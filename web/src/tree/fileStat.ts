import * as path from '../util/path'

export interface FileStat {
    path: string
    name: string
    isDirectory: boolean
    hasChildren: boolean
    children?: FileStat[]
}

/**
 * ICustomResolveFileOptions is based on vscode.ResolveFileOptions. It is only used when calling toFileStat.
 */
interface ICustomResolveFileOptions {
    /**
     * If specified, return only the subtree rooted at parentPath.
     */
    parentPath?: string

    /**
     * Also resolve nodes that are the only children of an otherwise resolved node.
     */
    resolveSingleChildDescendants?: boolean

    /**
     * Resolve the nodes with the specified paths, including all of their descendents (and the siblings of their
     * descendents). The array elements are file paths relative to the toFileStat root.
     */
    resolveTo?: string[]

    /**
     * Resolve all nodes in the tree.
     */
    resolveAllDescendants?: boolean
}

/**
 * toFileStat returns a tree of IFileStat that represents a tree.
 *
 * @param files A lexicographically sorted list of file paths in the workspace, relative to the root and without a
 *              leading '/'.
 */
export function toFileStat(files: string[], options: ICustomResolveFileOptions, skipFiles = 0): FileStat | null {
    // console.log('toFileStat', files.length, skipFiles, options.parentPath)

    const { parentPath, resolveSingleChildDescendants, resolveTo, resolveAllDescendants } = options

    if (parentPath === '' || parentPath === '.') {
        throw new Error('parentPath must be undefined if empty or "." (not empty string)')
    }
    if (parentPath && parentPath.startsWith('/')) {
        throw new Error('parentPath must not have a leading slash: ' + parentPath)
    }
    if (resolveTo && resolveTo.some(p => p.startsWith('/'))) {
        throw new Error(
            'resolveTo entries must not have a leading slash: ' + resolveTo.filter(p => p.startsWith('/')).join(', ')
        )
    }

    let resolveToPrefixes: string[] | undefined
    if (resolveTo) {
        resolveToPrefixes = []
        for (const path of resolveTo) {
            const parts = path.split('/')
            // The path might be a file or dir (or might not exist).
            for (let i = 1; i <= parts.length; i++) {
                const ancestor = parts.slice(0, i).join('/')
                if (!resolveToPrefixes!.includes(ancestor)) {
                    resolveToPrefixes!.push(ancestor)
                }
            }
        }
    }

    // Only return files under this rootPath.
    const rootPath = parentPath || '' // no trailing '/'
    const rootPathPrefix = rootPath ? rootPath + '/' : ''

    const rootStat: FileStat = {
        path: rootPath || '.',
        name: path.basename(rootPath),
        children: [],
        isDirectory: undefined!,
        hasChildren: undefined!,
    }
    if (!parentPath) {
        // The root is assumed to be a directory that exists.
        rootStat.isDirectory = true
    }

    if (resolveAllDescendants) {
        // Simple flat case.
        rootStat.children = []
        for (const file of files) {
            if (parentPath && !startsWith(file, parentPath + '/')) {
                continue
            }
            rootStat.children.push({
                path: rootPathPrefix + file,
                name: path.basename(file),
                isDirectory: false,
                hasChildren: false,
            })
        }
        rootStat.isDirectory = true
        rootStat.hasChildren = !!rootStat.children
        return rootStat
    }

    // NOTE: This loop is performance sensitive. It runs one iteration per file in the repo. We use shortcuts/hacks
    // because they meaningfully improve performance and we have tested them to ensure they don't cause other
    // problems.
    const parentPathPrefix = parentPath ? parentPath + '/' : '' // empty string value is never used
    const pathComponentPos = parentPath ? parentPath.length + 1 : 0
    let lastSubdir = ''
    let lastParentPathPrefixPlusSubdir: string | undefined
    let hasSeenParentPathPrefix = false
    for (let i = skipFiles; i < files.length; i++) {
        if (parentPath === undefined || startsWith(files[i], parentPathPrefix)) {
            if (parentPath !== undefined && !hasSeenParentPathPrefix) {
                hasSeenParentPathPrefix = true
            }

            const slashPos = files[i].indexOf('/', pathComponentPos)
            if (slashPos === -1) {
                // Is a file directly underneath parentPath.
                rootStat.children!.push({
                    path: files[i],
                    name: path.basename(files[i]),
                    isDirectory: false,
                    hasChildren: false,
                })
                lastSubdir = ''
            } else {
                // Is a file that is two or more levels below parentPath.
                const subdir = files[i].slice(pathComponentPos, slashPos)
                if (subdir !== lastSubdir) {
                    lastSubdir = subdir
                    lastParentPathPrefixPlusSubdir = parentPathPrefix + subdir
                    const resolveToThisFile =
                        resolveToPrefixes && resolveToPrefixes.includes(lastParentPathPrefixPlusSubdir)
                    const recurse =
                        resolveToThisFile ||
                        (resolveSingleChildDescendants &&
                            (i === files.length - 1 ||
                                !startsWith(files[i + 1], parentPath ? lastParentPathPrefixPlusSubdir : subdir)))
                    if (recurse) {
                        const recurseParentPath = parentPath ? lastParentPathPrefixPlusSubdir : subdir
                        rootStat.children!.push(
                            toFileStat(
                                files,
                                {
                                    parentPath: recurseParentPath,
                                    resolveSingleChildDescendants,
                                    resolveTo,
                                },
                                i
                            )!
                        )
                    } else {
                        rootStat.children!.push({
                            path: rootPathPrefix + subdir,
                            name: subdir,
                            isDirectory: true,
                            hasChildren: true,
                        })
                    }
                }
            }
            continue
        }
        if (hasSeenParentPathPrefix) {
            // Because we assume files is sorted, we know we won't find any more matches.
            break
        }
        if (parentPath !== undefined && files[i] === parentPath) {
            rootStat.isDirectory = false
            rootStat.hasChildren = false
            return rootStat
        }
    }

    rootStat.hasChildren = Boolean(rootStat.children && rootStat.children.length > 0)
    rootStat.isDirectory = Boolean(rootStat.hasChildren || (parentPath && files[skipFiles] === parentPath))
    if (!rootStat.isDirectory) {
        return null // not found
    }
    return rootStat
}

function startsWith(s: string, prefix: string): boolean {
    return s.substring(0, prefix.length) === prefix
}
