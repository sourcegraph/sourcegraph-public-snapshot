import * as monaco from 'monaco-editor'

let lastThemeName: string | undefined

export function adaptMonacoThemeToEditorTheme(): void {
    let editor: monaco.editor.ICodeEditor | undefined

    // Wait for init to set theme.
    monaco.editor.onDidCreateEditor(newEditor => {
        editor = newEditor
        setMonacoTheme({ editor })
    })

    const mutationObserver = new MutationObserver(() => {
        setMonacoTheme({ editor })
    })

    mutationObserver.observe(document.documentElement, { childList: false, attributes: true })
}

function setMonacoTheme({ editor }: { editor?: monaco.editor.ICodeEditor }): void {
    const body = document.querySelector<HTMLBodyElement>('body')
    const themeName = body?.dataset.vscodeThemeName

    if (lastThemeName !== themeName || !lastThemeName) {
        const themeKind = body?.dataset.vscodeThemeKind === 'vscode-light' ? 'theme-light' : 'theme-dark'

        const computedStyle = getComputedStyle(document.documentElement)

        const colors: monaco.editor.IColors = {}

        for (const colorId of Object.keys(monacoColorIdWebviewCustomProperties)) {
            try {
                const customProperty = monacoColorIdWebviewCustomProperties[colorId]
                const style = computedStyle.getPropertyValue(customProperty)

                colors[colorId] = rgbaToHex(style)
            } catch (error) {
                console.error('Error computing style for search box:', error)
            }
        }

        const rules: monaco.editor.ITokenThemeRule[] = []

        for (const { token, foregroundProperty } of tokenRuleCustomProperties) {
            try {
                const style = computedStyle.getPropertyValue(foregroundProperty)
                rules.push({ token, foreground: rgbaToHex(style) })
            } catch (error) {
                console.error('Error computing style for search box:', error)
            }
        }

        // Set font size
        const fontSize = parseInt(computedStyle.getPropertyValue('--vscode-editor-font-size'), 10)
        if (!isNaN(fontSize)) {
            editor?.updateOptions({ fontSize })
        }

        monaco.editor.defineTheme(themeKind === 'theme-light' ? 'sourcegraph-light' : 'sourcegraph-dark', {
            base: 'vs-dark',
            inherit: true,
            colors,
            rules,
        })
    }
}

// Same for dark and light theme.
const monacoColorIdWebviewCustomProperties: Record<string, string> = {
    background: '--vscode-editorWidget-background',
    'editor.background': '--vscode-editorWidget-background',
    'textLink.activeBackground': '--vscode-inputValidation-infoBackground', // Not avaliable to webview.
    'editor.foreground': '--vscode-editor-foreground',
    'editorCursor.foreground': '--vscode-editorCursor-foreground',
    'editorSuggestWidget.background': '--vscode-editorSuggestWidget-background',
    'editorSuggestWidget.foreground': '--vscode-editorSuggestWidget-foreground',
    'editorSuggestWidget.highlightForeground': '--vscode-editorSuggestWidget-highlightForeground',
    'editorSuggestWidget.selectedBackground': '--vscode-editorSuggestWidget-selectedBackground',
    'list.hoverBackground': '--vscode-list-hoverBackground',
    'editorSuggestWidget.border': '--vscode-editorSuggestWidget-border',
    'editorHoverWidget.background': '--vscode-editorHoverWidget-background',
    'editorHoverWidget.foreground': '--vscode-editorHoverWidget-foreground',
    'editorHoverWidget.border': '--vscode-editorHoverWidget-border',
    'editor.hoverHighlightBackground': '--vscode-editor-hoverHighlightBackground',
}

const tokenRuleCustomProperties: { token: string; foregroundProperty: string }[] = [
    // Sourcegraph base language tokens
    { token: 'identifier', foregroundProperty: '--vscode-foreground' },
    { token: 'field', foregroundProperty: '--vscode-debugTokenExpression-name' },
    { token: 'keyword', foregroundProperty: '--vscode-debugTokenExpression-boolean' },
    { token: 'openingParen', foregroundProperty: '--vscode-debugTokenExpression-boolean' },
    { token: 'closingParen', foregroundProperty: '--vscode-debugTokenExpression-boolean' },
    { token: 'comment', foregroundProperty: '--vscode-debugTokenExpression-string' },
    // Sourcegraph decorated language tokens
    { token: 'metaFilterSeparator', foregroundProperty: '--vscode-descriptionForeground' },
    { token: 'metaRepoRevisionSeparator', foregroundProperty: '--vscode-debugTokenExpression-name' },
    { token: 'metaContextPrefix', foregroundProperty: '--vscode-debugTokenExpression-boolean' },
    { token: 'metaPredicateNameAccess', foregroundProperty: '--vscode-debugTokenExpression-boolean' },
    { token: 'metaPredicateDot', foregroundProperty: '--vscode-foreground' },
    { token: 'metaPredicateParenthesis', foregroundProperty: '--vscode-debugTokenExpression-string' },
    // Regexp pattern highlighting
    { token: 'metaRegexpDelimited', foregroundProperty: '--vscode-debugTokenExpression-error' },
    { token: 'metaRegexpAssertion', foregroundProperty: '--vscode-debugTokenExpression-error' },
    { token: 'metaRegexpLazyQuantifier', foregroundProperty: '--vscode-debugTokenExpression-error' },
    { token: 'metaRegexpEscapedCharacter', foregroundProperty: '--vscode-debugConsole-warningForeground' },
    { token: 'metaRegexpCharacterSet', foregroundProperty: '--vscode-debugTokenExpression-boolean' },
    { token: 'metaRegexpCharacterClass', foregroundProperty: '--vscode-textLink-foreground' },
    { token: 'metaRegexpCharacterClassMember', foregroundProperty: '--vscode-foreground' },
    { token: 'metaRegexpCharacterClassRange', foregroundProperty: '--vscode-foreground' },
    { token: 'metaRegexpCharacterClassRangeHyphen', foregroundProperty: '--vscode-debugTokenExpression-boolean' },
    { token: 'metaRegexpRangeQuantifier', foregroundProperty: '--vscode-terminal-ansiBrightCyan' },
    { token: 'metaRegexpAlternative', foregroundProperty: '--vscode-terminal-ansiBrightCyan' },
    // Structural pattern highlighting
    { token: 'metaStructuralHole', foregroundProperty: '--vscode-debugTokenExpression-error' },
    { token: 'metaStructuralRegexpHole', foregroundProperty: '--vscode-debugTokenExpression-error' },
    { token: 'metaStructuralVariable', foregroundProperty: '--vscode-foreground' },
    { token: 'metaStructuralRegexpSeparator', foregroundProperty: '--vscode-debugTokenExpression-string' },
    // Revision highlighting
    { token: 'metaRevisionSeparator', foregroundProperty: '--vscode-debugTokenExpression-string' },
    { token: 'metaRevisionIncludeGlobMarker', foregroundProperty: '--vscode-debugTokenExpression-error' },
    { token: 'metaRevisionExcludeGlobMarker', foregroundProperty: '--vscode-debugTokenExpression-error' },
    { token: 'metaRevisionCommitHash', foregroundProperty: '--vscode-foreground' },
    { token: 'metaRevisionLabel', foregroundProperty: '--vscode-foreground' },
    { token: 'metaRevisionReferencePath', foregroundProperty: '--vscode-foreground' },
    { token: 'metaRevisionWildcard', foregroundProperty: '--vscode-terminal-ansiBrightCyan' },
    // Path-like highlighting
    { token: 'metaPathSeparator', foregroundProperty: '--vscode-descriptionForeground' },
]

/**
 * Converts an rgb(a) color to a hex (#rrggbb(aa)) color.
 * Will return the same value if passed a hex color.
 * Necessary because Monaco does not accept rgb(a) colors.
 *
 * Adapted from  https://github.com/sindresorhus/rgb-hex
 *
 * Debt: apply simple alpha compositing to resolve color
 * (given background) without reducing opacity.
 */
function rgbaToHex(value: string): string {
    value = value.trim()
    if (!value.startsWith('rgb')) {
        return value
    }

    const matches = value.match(/(0?\.?\d{1,3})%?\b/g)?.map(component => Number(component))

    if (!matches) {
        console.error('Invalid RGB value:', value)
        return ''
    }
    let [red, green, blue, alpha] = matches

    if (
        typeof red !== 'number' ||
        typeof green !== 'number' ||
        typeof blue !== 'number' ||
        red > 255 ||
        green > 255 ||
        blue > 255
    ) {
        console.error('Expected three numbers below 256')
        return ''
    }

    let hexAlpha = ''

    if (typeof alpha === 'number' && !isNaN(alpha)) {
        const isPercent = value.includes('%')
        if (!isPercent && alpha >= 0 && alpha <= 1) {
            alpha = Math.round(255 * alpha)
        } else if (isPercent && alpha >= 0 && alpha <= 100) {
            alpha = Math.round((255 * alpha) / 100)
        } else {
            throw new TypeError(`Expected alpha value (${alpha}) as a fraction or percentage`)
        }

        hexAlpha = (alpha | (1 << 8)).toString(16).slice(1)
    }

    return `#${(blue | (green << 8) | (red << 16) | (1 << 24)).toString(16).slice(1) + hexAlpha}`
}
