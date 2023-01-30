import { foldGutter, LanguageDescription } from '@codemirror/language'
import { languages } from '@codemirror/language-data'
import { Compartment, Extension } from '@codemirror/state'
import { EditorView, ViewPlugin } from '@codemirror/view'

import { logger } from '@sourcegraph/common'

import { blobPropsFacet } from './index'

/**
 * Returns {@link LanguageDescription} if found and language supports code folding.
 */
function findLanguageByFilename(filename: string): LanguageDescription | undefined {
    const language = LanguageDescription.matchFilename(languages, filename)

    if (!language) {
        return undefined
    }

    /**
     * Built-in [`foldGutter`](https://sourcegraph.com/github.com/codemirror/language@aefb707fcea4c6b3581b2c22ac4efc95ff934568/-/blob/src/fold.ts?L328-395)
     * extension relies on [`foldNodeProp`](https://sourcegraph.com/github.com/codemirror/language@aefb707fcea4c6b3581b2c22ac4efc95ff934568/-/blob/src/fold.ts?L36#tab=references)
     * presence in language parser configuration (e.g., [`javascriptLanguage`](https://sourcegraph.com/github.com/codemirror/lang-javascript@4f2c18d1eee8269be63a18b72fecfd21d3e24c21/-/blob/src/javascript.ts?L43-45)).
     *
     * [Legacy language modes](https://sourcegraph.com/github.com/codemirror/language-data@361a5251cd6a90d884219488d475396b505f551c/-/blob/src/language-data.ts?L170-986)
     * ported from CodeMirror 5 are [do not add fold node props](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/codemirror/legacy-modes$+foldNodeProp&patternType=standard&sm=1&groupBy=path)
     * thus we can ignore such languages.
     */
    const supportsCodeFolding = !language.load.toString().includes('@codemirror/legacy-modes/mode')

    return supportsCodeFolding ? language : undefined
}

const languageConfigCompartment = new Compartment()

/**
 * Enables code folding for [supported languages](https://sourcegraph.com/github.com/codemirror/language-data@361a5251cd6a90d884219488d475396b505f551c/-/blob/src/language-data.ts?L13-168).
 */
export function codeFoldingExtension(): Extension {
    return [
        foldGutter(),

        languageConfigCompartment.of([]),

        ViewPlugin.fromClass(
            class {
                constructor(public view: EditorView) {
                    // eslint-disable-next-line no-void
                    void this.loadExtension()
                }

                private async loadExtension(): Promise<void> {
                    const blobProps = this.view.state.facet(blobPropsFacet)
                    const language = findLanguageByFilename(blobProps.blobInfo.filePath ?? '')

                    try {
                        const support = await language?.load()

                        if (support) {
                            this.view.dispatch({
                                effects: languageConfigCompartment.reconfigure(support),
                            })
                        }
                    } catch (error) {
                        logger.error(error)
                    }
                }
            }
        ),
    ]
}
