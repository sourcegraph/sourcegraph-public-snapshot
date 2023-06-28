import path from 'path'

const CODE_CONTEXT_TEMPLATE = `Use following code snippet from file \`{filePath}\`:
\`\`\`{language}
{text}
\`\`\``

const CODE_CONTEXT_TEMPLATE_WITH_REPO = `Use following code snippet from file \`{filePath}\` in repository \`{repoName}\`:
\`\`\`{language}
{text}
\`\`\``

export function populateCodeContextTemplate(code: string, filePath: string, repoName?: string): string {
    return (repoName ? CODE_CONTEXT_TEMPLATE_WITH_REPO.replace('{repoName}', repoName) : CODE_CONTEXT_TEMPLATE)
        .replace('{filePath}', filePath)
        .replace('{language}', getExtension(filePath))
        .replace('{text}', code)
}

const MARKDOWN_CONTEXT_TEMPLATE = 'Use the following text from file `{filePath}`:\n{text}'

const MARKDOWN_CONTEXT_TEMPLATE_WITH_REPO =
    'Use the following text from file `{filePath}` in repository `{repoName}`:\n{text}'

export function populateMarkdownContextTemplate(markdown: string, filePath: string, repoName?: string): string {
    return (repoName ? MARKDOWN_CONTEXT_TEMPLATE_WITH_REPO.replace('{repoName}', repoName) : MARKDOWN_CONTEXT_TEMPLATE)
        .replace('{filePath}', filePath)
        .replace('{text}', markdown)
}

const CURRENT_EDITOR_CODE_TEMPLATE = 'I have the `{filePath}` file opened in my editor. '

const CURRENT_EDITOR_CODE_TEMPLATE_WITH_REPO =
    'I have the `{filePath}` file from the repository `{repoName}` opened in my editor. '

export function populateCurrentEditorContextTemplate(code: string, filePath: string, repoName?: string): string {
    const context = isMarkdownFile(filePath)
        ? populateMarkdownContextTemplate(code, filePath, repoName)
        : populateCodeContextTemplate(code, filePath, repoName)
    return (
        (repoName
            ? CURRENT_EDITOR_CODE_TEMPLATE_WITH_REPO.replace('{repoName}', repoName)
            : CURRENT_EDITOR_CODE_TEMPLATE
        ).replace(/{filePath}/g, filePath) + context
    )
}

const CURRENT_EDITOR_SELECTED_CODE_TEMPLATE = 'I am currently looking at this part of the code from `{filePath}`. '

const CURRENT_EDITOR_SELECTED_CODE_TEMPLATE_WITH_REPO =
    'I am currently looking at this part of the code from `{filePath}` in repository {repoName}. '

export function populateCurrentEditorSelectedContextTemplate(
    code: string,
    filePath: string,
    repoName?: string
): string {
    const context = isMarkdownFile(filePath)
        ? populateMarkdownContextTemplate(code, filePath, repoName)
        : populateCodeContextTemplate(code, filePath, repoName)
    return (
        (repoName
            ? CURRENT_EDITOR_SELECTED_CODE_TEMPLATE_WITH_REPO.replace('{repoName}', repoName)
            : CURRENT_EDITOR_SELECTED_CODE_TEMPLATE
        ).replace(/{filePath}/g, filePath) + context
    )
}

const MARKDOWN_EXTENSIONS = new Set(['md', 'markdown'])

export function isMarkdownFile(filePath: string): boolean {
    return MARKDOWN_EXTENSIONS.has(getExtension(filePath))
}

function getExtension(filePath: string): string {
    return path.extname(filePath).slice(1)
}

const OWNER_INFO_TEMPLATE =
    'The owner and the right {subject} to contact for any help regarding the file `{filePath}` is {owner}, as they {reason}.'
const COMMIT_INFO_TEMPLATE =
    'The file `{filePath}` was last edited in commit `{commitId}` by `{author}` on `{date}`. The subject of the commit was `{subject}`.'

export const OWNERSHIP_REASON_MAP: Record<string, string> = {
    CodeOwnersFileEntry: 'are listed in the CODEOWNERS file',
    AssignedOwner: 'are assigned as the owner of the file',
    RecentViewOwnershipSignal: 'recently viewed the file',
    RecentContributorOwnershipSignal: 'recently contributed to the file',
}

export function appendOwnerAndCommitInfo(
    base: string,
    filePath: string,
    owner?: {
        reason?: string
        type: 'Person' | 'Team'
        name: string
    },
    commit?: { oid: string; date: string; subject: string; author: string }
): string {
    let prompt = base

    if (commit) {
        prompt =
            prompt +
            '\n' +
            COMMIT_INFO_TEMPLATE.replace('{filePath}', filePath)
                .replace('{commitId}', commit.oid)
                .replace('{subject}', commit.subject)
                .replace('{author}', commit.author)
                .replace('{date}', commit.date)
    }

    if (owner) {
        prompt =
            prompt +
            '\n' +
            OWNER_INFO_TEMPLATE.replace('{filePath}', filePath)
                .replace('{subject}', owner.type.toLowerCase())
                .replace('{owner}', owner.name)
                .replace('{reason}', OWNERSHIP_REASON_MAP[owner.reason || 'CodeOwnersFileEntry'])
    }

    return prompt
}
