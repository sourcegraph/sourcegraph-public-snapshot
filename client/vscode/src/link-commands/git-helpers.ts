import * as path from 'path'

import execa from 'execa'
import vscode, { TextEditor } from 'vscode'

import { version } from '../../package.json'
import { log } from '../log'

interface RepositoryInfo extends Branch, RemoteName {
    /** Git repository remote URL */
    remoteURL: string

    /** File path relative to the repository root */
    fileRelative: string
}

export type GitHelpers = typeof gitHelpers

export interface RemoteName {
    /**
     * Remote name of the upstream repository,
     * or the first found remote name if no upstream is found
     */
    remoteName: string
}

export interface Branch {
    /**
     * Remote branch name, or 'HEAD' if it isn't found because
     * e.g. detached HEAD state, upstream branch points to a local branch
     */
    branch: string
}

/**
 * Returns the Git repository remote URL, the current branch, and the file path
 * relative to the repository root. Returns undefined if no remote is found
 */
export async function repoInfo(filePath: string): Promise<RepositoryInfo | undefined> {
    try {
        // Determine repository root directory.
        const fileDirectory = path.dirname(filePath)
        const repoRoot = await gitHelpers.rootDirectory(fileDirectory)

        // Determine file path relative to repository root, then replace slashes
        // as \\ does not work in Sourcegraphl links
        const fileRelative = filePath.slice(repoRoot.length + 1).replace(/\\/g, '/')

        let { branch, remoteName } = await gitRemoteNameAndBranch(repoRoot, gitHelpers, log)

        branch = getDefaultBranch() || branch

        const remoteURL = await gitRemoteUrlWithReplacements(repoRoot, remoteName, gitHelpers, log)

        return { remoteURL, branch, fileRelative, remoteName }
    } catch {
        return undefined
    }
}

export async function gitRemoteNameAndBranch(
    repoDirectory: string,
    git: Pick<GitHelpers, 'branch' | 'remotes' | 'upstreamAndBranch'>,
    log?: {
        appendLine: (value: string) => void
    }
): Promise<RemoteName & Branch> {
    let remoteName: string | undefined

    // Used to determine which part of upstreamAndBranch is the remote name, or as fallback if no upstream is set
    const remotes = await git.remotes(repoDirectory)
    const branch = await git.branch(repoDirectory)

    try {
        const upstreamAndBranch = await git.upstreamAndBranch(repoDirectory)
        // Subtract $BRANCH_NAME from $UPSTREAM_REMOTE/$BRANCH_NAME.
        // We can't just split on the delineating `/`, since refnames can include `/`:
        // https://sourcegraph.com/github.com/git/git@454cb6bd52a4de614a3633e4f547af03d5c3b640/-/blob/refs.c#L52-67

        // Example:
        // stdout: remote/two/tj/feature
        // remoteName: remote/two, branch: tj/feature

        const branchPosition = upstreamAndBranch.lastIndexOf(branch)
        const maybeRemote = upstreamAndBranch.slice(0, branchPosition - 1)
        if (branchPosition !== -1 && maybeRemote) {
            remoteName = maybeRemote
        }
    } catch {
        // noop. upstream may not be set
    }

    // If we cannot find the remote name deterministically, we use the first
    // Git remote found.
    if (!remoteName) {
        if (remotes.length > 1) {
            log?.appendLine(`no upstream found, using first git remote: ${remotes[0]}`)
        }
        remoteName = remotes[0]
    }

    // Throw if a remote still isn't found
    if (!remoteName) {
        throw new Error('no configured git remotes')
    }

    return { remoteName, branch }
}

export const gitHelpers = {
    /**
     * Returns the repository root directory for any directory within the
     * repository.
     */
    async rootDirectory(repoDirectory: string): Promise<string> {
        const { stdout } = await execa('git', ['rev-parse', '--show-toplevel'], { cwd: repoDirectory })
        return stdout
    },

    /**
     * Returns the names of all git remotes, e.g. ["origin", "foobar"]
     */
    async remotes(repoDirectory: string): Promise<string[]> {
        const { stdout } = await execa('git', ['remote'], { cwd: repoDirectory })
        return stdout.split('\n')
    },

    /**
     * Returns the remote URL for the given remote name.
     * e.g. `origin` -> `git@github.com:foo/bar`
     */
    async remoteUrl(remoteName: string, repoDirectory: string): Promise<string> {
        const { stdout } = await execa('git', ['remote', 'get-url', remoteName], { cwd: repoDirectory })
        return stdout
    },

    /**
     * Returns either the current branch name of the repository OR in all
     * other cases (e.g. detached HEAD state), it returns "HEAD".
     */
    async branch(repoDirectory: string): Promise<string> {
        const { stdout } = await execa('git', ['rev-parse', '--abbrev-ref', 'HEAD'], { cwd: repoDirectory })
        return stdout
    },

    /**
     * Returns a string in the format $UPSTREAM_REMOTE/$BRANCH_NAME, e.g. "origin/branch-name", throws if not found
     */
    async upstreamAndBranch(repoDirectory: string): Promise<string> {
        const { stdout } = await execa('git', ['rev-parse', '--abbrev-ref', 'HEAD@{upstream}'], { cwd: repoDirectory })
        return stdout
    },
}

/**
 * Returns the remote URL for the given remote name with remote URL replacements.
 * e.g. `origin` -> `git@github.com:foo/bar`
 */
export async function gitRemoteUrlWithReplacements(
    repoDirectory: string,
    remoteName: string,
    gitHelpers: Pick<GitHelpers, 'remoteUrl'>,
    log?: { appendLine: (value: string) => void }
): Promise<string> {
    let stdout = await gitHelpers.remoteUrl(remoteName, repoDirectory)
    const replacementsList = getRemoteUrlReplacements()

    const stdoutBefore = stdout

    for (const replacement in replacementsList) {
        if (typeof replacement === 'string') {
            stdout = stdout.replace(replacement, replacementsList[replacement])
        }
    }

    log?.appendLine(`${stdoutBefore} became ${stdout}`)
    return stdout
}

/**
 * Uses editor endpoint to construct sourcegraph file URL
 */
export function getSourcegraphFileUrl(
    SourcegraphUrl: string,
    remoteURL: string,
    branch: string,
    fileRelative: string,
    editor: TextEditor
): string {
    return (
        `${SourcegraphUrl}/-/editor` +
        `?remote_url=${encodeURIComponent(remoteURL)}` +
        `&branch=${encodeURIComponent(branch)}` +
        `&file=${encodeURIComponent(fileRelative)}` +
        `&editor=${encodeURIComponent('VSCode')}` +
        `&version=${encodeURIComponent(version)}` +
        `&start_row=${encodeURIComponent(String(editor.selection.start.line))}` +
        `&start_col=${encodeURIComponent(String(editor.selection.start.character))}` +
        `&end_row=${encodeURIComponent(String(editor.selection.end.line))}` +
        `&end_col=${encodeURIComponent(String(editor.selection.end.character))}`
    )
}

function getRemoteUrlReplacements(): Record<string, string> {
    // has default value
    // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
    const replacements = vscode.workspace
        .getConfiguration('sourcegraph')
        .get<Record<string, string>>('remoteUrlReplacements')!
    return replacements
}

export function getDefaultBranch(): string {
    // has default value
    // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
    const branch = vscode.workspace.getConfiguration('sourcegraph').get<string>('defaultBranch')!

    return branch
}
