import highlight from 'highlight.js'

export const highlightCode = (code: string, language?: string): string => {
    try {
        if (language === 'plaintext' || language === 'text') {
            return code
        }
        if (language) {
            return highlight.highlight(code, { language, ignoreIllegals: true }).value
        }
        return highlight.highlightAuto(code).value
    } catch {
        return code
    }
}

let registered = false

export function registerHighlightContributions(): void {
    if (registered) {
        // Don't double-register these. (There is no way to unregister them.)
        return
    }
    registered = true

    /* eslint-disable @typescript-eslint/no-require-imports */
    /* eslint-disable @typescript-eslint/no-var-requires */
    highlight.registerLanguage('go', require('highlight.js/lib/languages/go'))
    highlight.registerLanguage('javascript', require('highlight.js/lib/languages/javascript'))
    highlight.registerLanguage('typescript', require('highlight.js/lib/languages/typescript'))
    highlight.registerLanguage('java', require('highlight.js/lib/languages/java'))
    highlight.registerLanguage('python', require('highlight.js/lib/languages/python'))
    highlight.registerLanguage('php', require('highlight.js/lib/languages/php'))
    highlight.registerLanguage('bash', require('highlight.js/lib/languages/bash'))
    highlight.registerLanguage('clojure', require('highlight.js/lib/languages/clojure'))
    highlight.registerLanguage('cpp', require('highlight.js/lib/languages/cpp'))
    highlight.registerLanguage('cs', require('highlight.js/lib/languages/csharp'))
    highlight.registerLanguage('csharp', require('highlight.js/lib/languages/csharp'))
    highlight.registerLanguage('css', require('highlight.js/lib/languages/css'))
    highlight.registerLanguage('dockerfile', require('highlight.js/lib/languages/dockerfile'))
    highlight.registerLanguage('elixir', require('highlight.js/lib/languages/elixir'))
    highlight.registerLanguage('haskell', require('highlight.js/lib/languages/haskell'))
    highlight.registerLanguage('html', require('highlight.js/lib/languages/xml'))
    highlight.registerLanguage('lua', require('highlight.js/lib/languages/lua'))
    highlight.registerLanguage('ocaml', require('highlight.js/lib/languages/ocaml'))
    highlight.registerLanguage('r', require('highlight.js/lib/languages/r'))
    highlight.registerLanguage('ruby', require('highlight.js/lib/languages/ruby'))
    highlight.registerLanguage('rust', require('highlight.js/lib/languages/rust'))
    highlight.registerLanguage('swift', require('highlight.js/lib/languages/swift'))
    highlight.registerLanguage('markdown', require('highlight.js/lib/languages/markdown'))
    highlight.registerLanguage('diff', require('highlight.js/lib/languages/diff'))
    highlight.registerLanguage('json', require('highlight.js/lib/languages/json'))
    highlight.registerLanguage('jsonc', require('highlight.js/lib/languages/json'))
    highlight.registerLanguage('yaml', require('highlight.js/lib/languages/yaml'))
    highlight.registerLanguage('kotlin', require('highlight.js/lib/languages/kotlin'))
    highlight.registerLanguage('dart', require('highlight.js/lib/languages/dart'))
    highlight.registerLanguage('perl', require('highlight.js/lib/languages/perl'))
    highlight.registerLanguage('scala', require('highlight.js/lib/languages/scala'))
    highlight.registerLanguage('apex', require('highlight.js/lib/languages/java'))
    highlight.registerLanguage('http', require('highlight.js/lib/languages/http'))
}
