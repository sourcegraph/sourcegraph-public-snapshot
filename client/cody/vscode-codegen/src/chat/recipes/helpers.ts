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

export function getNormalizedLanguageName(extension: string): string {
    if (!extension) {
        return ''
    }
    const language = EXTENSION_TO_LANGUAGE[extension]
    if (language) {
        return language
    }
    return extension.charAt(0).toUpperCase() + extension.slice(1)
}
