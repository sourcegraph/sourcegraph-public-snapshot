import React, { useEffect, useState } from 'react'

import * as H from 'history'

import { HoveredToken } from '@sourcegraph/codeintellify'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import {
    AbsoluteRepoFile,
    FileSpec,
    ModeSpec,
    UIPositionSpec,
    RepoSpec,
    ResolvedRevisionSpec,
    RevisionSpec,
} from '@sourcegraph/shared/src/util/url'
import gql from 'tagged-template-noop'

import { HoverThresholdProps } from '../RepoContainer'
import { Decoration, EditorView, hoverTooltip, lineNumbers, TooltipView } from '@codemirror/view'
import { countColumn, EditorState, Facet, findColumn, RangeSetBuilder } from '@codemirror/state'
import { JsonDocument, Occurrence, SyntaxKind } from '../../lsif/lsif-typed'
import { renderMarkdown } from '@sourcegraph/common'
import { requestGraphQL } from '../../backend/graphql'
import { DefinitionAndHoverResult, DefinitionAndHoverVariables } from '../../graphql-operations'

export interface BlobProps
    extends SettingsCascadeProps,
        PlatformContextProps<'urlToFile' | 'requestGraphQL' | 'settings' | 'forceUpdateTooltip'>,
        TelemetryProps,
        HoverThresholdProps,
        ExtensionsControllerProps,
        ThemeProps {
    location: H.Location
    history: H.History
    className: string
    wrapCode: boolean
    /** The current text document to be rendered and provided to extensions */
    blobInfo: BlobInfo

    // Experimental reference panel
    disableStatusBar: boolean
    disableDecorations: boolean

    lsif: string

    // If set, nav is called when a user clicks on a token highlighted by
    // WebHoverOverlay
    nav?: (url: string) => void
}

export interface BlobInfo extends AbsoluteRepoFile, ModeSpec {
    /** The raw content of the blob. */
    content: string

    /** The trusted syntax-highlighted code as HTML */
    html: string
}

export const Blob: React.FunctionComponent<React.PropsWithChildren<BlobProps>> = props => {
    const [container, setContainer] = useState<HTMLDivElement | null>(null)
    codemirror(container, props)
    return (
        <div
            ref={setContainer}
            // className={classNames(styles.root, className)}
            data-test-id="codemirror-blob-view"
        />
    )
}

interface OccurrenceRange {
    start: number
    end: number
}
interface CodemirrorOccurrence {
    range: OccurrenceRange
    kind: string
}

function occurrenceToDecoration(occ: CodemirrorOccurrence): Decoration {
    return Decoration.mark({ class: 'hl-typed-' + occ.kind })
}

export const scipTokens = Facet.define<CodemirrorOccurrence[], CodemirrorOccurrence[]>({
    combine(input) {
        return input[0] ?? []
    },
})

const tabSize = 1

function codemirror(container: HTMLDivElement | null, props: BlobProps) {
    const [_, setView] = useState<EditorView>()

    useEffect(() => {
        if (!container) {
            return
        }
        const occurrences = scipTokens.compute(['doc'], state => {
            const occurrences = (JSON.parse(props.lsif) as JsonDocument).occurrences.map(occ => new Occurrence(occ))
            const result: CodemirrorOccurrence[] = []
            for (const occ of occurrences) {
                if (occ.range.start.line !== occ.range.end.line) {
                    continue
                }
                const line = state.doc.line(occ.range.start.line + 1)

                const start = line.from + findColumn(line.text, occ.range.start.character, tabSize, false)
                const end = line.from + findColumn(line.text, occ.range.end.character, tabSize, false)
                result.push({
                    kind: SyntaxKind[occ.kind],
                    range: { start, end },
                })
            }
            return result
        })

        const tokenHighlight = EditorView.decorations.compute([scipTokens], state => {
            const tokens = state.facet(scipTokens)
            tokens.sort((a, b) => a.range.start - b.range.start)
            const builder = new RangeSetBuilder<Decoration>()
            for (const token of tokens) {
                builder.add(token.range.start, token.range.end, occurrenceToDecoration(token))
            }
            return builder.finish()
        })
        const hover = hoverTooltip(
            async (view, pos) => {
                const line = view.state.doc.lineAt(pos)
                const character = countColumn(line.text, tabSize, pos - line.from)
                const params: DefinitionAndHoverVariables = {
                    character,
                    line: line.number - 1,
                    commit: props.blobInfo.commitID,
                    path: props.blobInfo.filePath,
                    repository: props.blobInfo.repoName,
                }
                console.log({ params })
                const data = await fetchDefinitionAndHover(params)
                console.log({ data })
                const markdown = data?.repository?.commit?.blob?.lsif?.hover?.markdown.text
                if (!markdown) {
                    return null
                }
                return {
                    pos: pos,
                    end: pos,
                    create(): TooltipView {
                        const dom = document.createElement('div')
                        dom.innerHTML = renderMarkdown(markdown)
                        return { dom }
                    },
                }
            },
            {
                hoverTime: 100,
                // Hiding the tooltip when the document changes replicates
                // Monaco's behavior and also "feels right" because it removes
                // "clutter" from the input.
                hideOnChange: true,
            }
        )

        const state = EditorState.create({
            doc: props.blobInfo.content,
            extensions: [lineNumbers(), EditorState.readOnly.of(true), occurrences, tokenHighlight, hover],
        })
        const view = new EditorView({
            state: state,
            parent: container,
        })
        setView(view)
        return () => {
            setView(undefined)
            view.destroy()
        }
        // Extensions and value are updated via transactions below
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [container])
}

function fetchDefinitionAndHover(args: DefinitionAndHoverVariables): Promise<DefinitionAndHoverResult | null> {
    return requestGraphQL<DefinitionAndHoverResult, DefinitionAndHoverVariables>(
        gql`
            query DefinitionAndHover(
                $repository: String!
                $commit: String!
                $path: String!
                $line: Int!
                $character: Int!
            ) {
                repository(name: $repository) {
                    commit(rev: $commit) {
                        blob(path: $path) {
                            lsif {
                                definitions(line: $line, character: $character) {
                                    nodes {
                                        resource {
                                            path
                                            repository {
                                                name
                                            }
                                            commit {
                                                oid
                                            }
                                        }
                                        range {
                                            start {
                                                line
                                                character
                                            }
                                            end {
                                                line
                                                character
                                            }
                                        }
                                    }
                                }
                                hover(line: $line, character: $character) {
                                    markdown {
                                        text
                                    }
                                    range {
                                        start {
                                            line
                                            character
                                        }
                                        end {
                                            line
                                            character
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        `,
        args
    )
        .toPromise()
        .then(x => x.data)
}

export function getLSPTextDocumentPositionParameters(
    position: HoveredToken & RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec,
    mode: string
): RepoSpec & RevisionSpec & ResolvedRevisionSpec & FileSpec & UIPositionSpec & ModeSpec {
    return {
        repoName: position.repoName,
        filePath: position.filePath,
        commitID: position.commitID,
        revision: position.revision,
        mode,
        position,
    }
}
