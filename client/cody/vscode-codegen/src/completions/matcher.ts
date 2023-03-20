class DocumentMatcher {
    private lines: string[]
    constructor(private document: string) {
        this.lines = document.split('\n')
    }

    public bestMatch(snippet: string): string {}
}

function toKeywords(text: string): { [word: string]: number } {
    const keywords: { [word: string]: number } = {}
    for (const word of text.split(/\W+/)) {
        if (word.length === 0) {
            continue
        }
        if (!keywords[word]) {
            keywords[word] = 0
        }
        keywords[word]++
    }
    return keywords
}

function tokenizeCamelCase(input: string): string[] {
    const tokens: string[] = []
    let currentWord = ''

    for (let i = 0; i < input.length; i++) {
        const char = input[i]
        const isUpperCase = char === char.toUpperCase()
        const isAlphanumeric = /^[a-zA-Z0-9]+$/.test(char)

        if (isAlphanumeric) {
            if (isUpperCase && currentWord.length > 0) {
                tokens.push(currentWord)
                currentWord = ''
            }
            currentWord += char
        } else {
            if (currentWord.length > 0) {
                tokens.push(currentWord)
                currentWord = ''
            }
            if (char.trim().length > 0) {
                tokens.push(char)
            }
        }
    }

    if (currentWord.length > 0) {
        tokens.push(currentWord)
    }

    return tokens
}
