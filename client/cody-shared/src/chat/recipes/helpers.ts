import path from 'path'

import { CodebaseContext } from '../../codebase-context'
import { ContextMessage, getContextMessageWithResponse } from '../../codebase-context/messages'
import { populateCodeContextTemplate } from '../../prompt/templates'

export const MARKDOWN_FORMAT_PROMPT = 'Enclose code snippets with three backticks like so: ```.'

const EXTENSION_TO_LANGUAGE: { [key: string]: string } = {
    py: 'Python',
    rb: 'Ruby',
    md: 'Markdown',
    php: 'PHP',
    js: 'Javascript',
    ts: 'Typescript',
    jsx: 'JSX',
    tsx: 'TSX',
}

export const commandRegex = {
    chat: new RegExp(/^(?!.*\/n(ew)?\s|.*\/f(ix)?\s)/i), // For now, if the input does not start with /n or /f, it is a chat
    fix: new RegExp(/^\/f(ix)?\s/i),
    touch: new RegExp(/^\/t(ouch)?\s/i),
    touchNeedFileName: new RegExp(/^\/t(ouch)?\s(?!.*test)/i), // Has /touch or /t but no test or tests in the string
    noTest: new RegExp(/^(?!.*test)/i),
    search: new RegExp(/^\/s(earch)?\s/i),
    test: new RegExp(/^\/n(ew)?\s|test(s)?\s/, 'i'),
}

export function getNormalizedLanguageName(extension: string): string {
    return extension ? EXTENSION_TO_LANGUAGE[extension] ?? extension.charAt(0).toUpperCase() + extension.slice(1) : ''
}

export async function getContextMessagesFromSelection(
    selectedText: string,
    precedingText: string,
    followingText: string,
    { fileName, repoName, revision }: { fileName: string; repoName?: string; revision?: string },
    codebaseContext: CodebaseContext
): Promise<ContextMessage[]> {
    const selectedTextContext = await codebaseContext.getContextMessages(selectedText, {
        numCodeResults: 4,
        numTextResults: 0,
    })

    return selectedTextContext.concat(
        [precedingText, followingText].flatMap(text =>
            getContextMessageWithResponse(populateCodeContextTemplate(text, fileName, repoName), {
                fileName,
                repoName,
                revision,
            })
        )
    )
}

export function getFileExtension(fileName: string): string {
    return path.extname(fileName).slice(1).toLowerCase()
}

// This cleans up the code returned by Cody based on current behavior
// ex. Remove  `tags:` that Cody sometimes include in the returned content
// It also removes all spaces before a new line to keep the indentations
export function contentSanitizer(text: string): string {
    let output = text + '\n'
    const tagsIndex = text.indexOf('tags:')
    if (tagsIndex !== -1) {
        // NOTE: 6 is the length of `tags:` + 1 space
        output = output.slice(tagsIndex + 6)
    }
    return output.replace(/^\s*\n/, '')
}

export function convertGitCloneURLToCodebaseName(cloneURL: string): string | null {
    if (!cloneURL) {
        console.error(`Unable to determine the git clone URL for this workspace.\ngit output: ${cloneURL}`)
        return null
    }
    try {
        const uri = new URL(cloneURL.replace('git@', ''))
        // Handle common Git SSH URL format
        const match = cloneURL.match(/git@([^:]+):([\w-]+)\/([\w-]+)(\.git)?/)
        if (cloneURL.startsWith('git@') && match) {
            const host = match[1]
            const owner = match[2]
            const repo = match[3]
            return `${host}/${owner}/${repo}`
        }
        // Handle GitHub URLs
        if (uri.protocol.startsWith('github') || uri.href.startsWith('github')) {
            return `github.com/${uri.pathname.replace('.git', '')}`
        }
        // Handle GitLab URLs
        if (uri.protocol.startsWith('gitlab') || uri.href.startsWith('gitlab')) {
            return `gitlab.com/${uri.pathname.replace('.git', '')}`
        }
        // Handle HTTPS URLs
        if (uri.protocol.startsWith('http') && uri.hostname && uri.pathname) {
            return `${uri.hostname}${uri.pathname.replace('.git', '')}`
        }
        // Generic URL
        if (uri.hostname && uri.pathname) {
            return `${uri.hostname}${uri.pathname.replace('.git', '')}`
        }
        return null
    } catch (error) {
        console.error(`Cody could not extract repo name from clone URL ${cloneURL}:`, error)
        return null
    }
}
