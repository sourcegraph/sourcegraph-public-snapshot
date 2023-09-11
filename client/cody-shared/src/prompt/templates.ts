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
