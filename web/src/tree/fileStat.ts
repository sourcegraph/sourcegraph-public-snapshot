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
export interface ICustomResolveFileOptions {
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
 * Converts from IResolveFileOptions to our slightly custom and optimized ICustomResolveFileOptions.
 *
 * It precomputes/transforms the input parameters for toFileStat. For example, instead of passing toFileStat a full
 * URI for each resolveTo entry, it only passes the relative paths. This lets toFileStat avoid URI operations in
 * its tight loop.
 */
export function toICustomResolveFileOptions(
    parentPath?: string,
    options?: ICustomResolveFileOptions
): ICustomResolveFileOptions {
    return {
        parentPath,
        resolveSingleChildDescendants: options ? options.resolveSingleChildDescendants : undefined,
        resolveTo: options && options.resolveTo,
        resolveAllDescendants: options && options.resolveAllDescendants,
    }
}

/**
 * toFileStat returns a tree of IFileStat that represents a tree.
 *
 * @param files A lexicographically sorted list of file paths in the workspace, relative to the root and without a
 *              leading '/'.
 */
export function toFileStat(files: string[], options: ICustomResolveFileOptions, skipFiles = 0): FileStat | null {
    const { parentPath, resolveSingleChildDescendants, resolveTo, resolveAllDescendants } = options

    if (parentPath && parentPath.startsWith('/')) {
        throw new Error('parentPath must not have a leading slash: ' + parentPath)
    }
    if (resolveTo && resolveTo.some(p => p.startsWith('/'))) {
        throw new Error(
            'resolveTo entries must not have a leading slash: ' + resolveTo.filter(p => p.startsWith('/')).join(', ')
        )
    }

    const resolveToPrefixes = Object.create(null)
    if (resolveTo) {
        for (const path of resolveTo) {
            const parts = path.split('/')
            // The path might be a file or dir (or might not exist).
            for (let i = 1; i <= parts.length; i++) {
                const ancestor = parts.slice(0, i).join('/')
                resolveToPrefixes[ancestor] = true
            }
        }
    }

    // Only return files under this rootPath.
    const rootPath = parentPath || '' // no trailing '/'
    const rootPathPrefix = rootPath ? rootPath + '/' : ''

    const rootStat: FileStat = {
        path: rootPath || '.',
        name: path.basename(rootPath),
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
            if (parentPath && !file.startsWith(parentPath + '/')) {
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
    let lastSubdir: string | undefined
    let hasSeenParentPathPrefix = false
    for (let i = skipFiles; i < files.length; i++) {
        const file = files[i]
        if (file === parentPath) {
            rootStat.isDirectory = false
            rootStat.hasChildren = false
            return rootStat
        } else if (!parentPath || file.startsWith(parentPath + '/')) {
            if (parentPath) {
                hasSeenParentPathPrefix = true
            }
            if (!rootStat.hasChildren) {
                rootStat.isDirectory = true
                rootStat.hasChildren = true
                rootStat.children = []
            }

            const pathComponentPos = parentPath ? parentPath.length + 1 : 0
            const slashPos = file.indexOf('/', pathComponentPos)
            if (slashPos === -1) {
                // Is a file directly underneath parentPath.
                rootStat.children!.push({
                    path: file,
                    name: path.basename(file),
                    isDirectory: false,
                    hasChildren: false,
                })
                lastSubdir = undefined
            } else {
                // Is a file that is two or more levels below parentPath.
                const subdir = file.slice(pathComponentPos, slashPos)
                const parent = file.slice(0, slashPos)
                const resolveToThisFile = resolveToPrefixes && resolveToPrefixes[parent]
                if (subdir !== lastSubdir) {
                    lastSubdir = subdir
                    const recurse =
                        resolveToThisFile ||
                        (resolveSingleChildDescendants &&
                            (i === files.length - 1 ||
                                !files[i + 1].startsWith(parentPath ? parentPath + '/' + subdir : subdir)))
                    if (recurse) {
                        const recurseParentPath = parentPath ? parentPath + '/' + subdir : subdir
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
        } else if (hasSeenParentPathPrefix) {
            // Because we assume files is sorted, we know we won't find any more matches.
            break
        }
    }

    if (!rootStat.isDirectory) {
        return null // not found
    }

    if (!rootStat.children) {
        rootStat.hasChildren = false
    }
    return rootStat
}
