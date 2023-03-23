import path from 'path'

const CODE_CONTEXT_TEMPLATE = `Use following code snippet from file \`{filePath}\`:
\`\`\`{language}
{text}
\`\`\``

export function populateCodeContextTemplate(code: string, filePath: string): string {
    const language = path.extname(filePath).slice(1)
    return CODE_CONTEXT_TEMPLATE.replace('{filePath}', filePath).replace('{language}', language).replace('{text}', code)
}

const MARKDOWN_CONTEXT_TEMPLATE = 'Use the following text from file `{filePath}`:\n{text}'

export function populateMarkdownContextTemplate(md: string, filePath: string): string {
    return MARKDOWN_CONTEXT_TEMPLATE.replace('{filePath}', filePath).replace('{text}', md)
}
