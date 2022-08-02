import { useEffect } from 'react'

import { Extension } from '@codemirror/state'
import { EditorView, hoverTooltip, repositionTooltips, Tooltip } from '@codemirror/view'
import { upperFirst } from 'lodash'
import { createRoot } from 'react-dom/client'

import { isErrorLike } from '@sourcegraph/common'
import hoverOverlayStyle from '@sourcegraph/shared/src/hover/HoverOverlay.module.scss'
import { HoverOverlayBaseProps } from '@sourcegraph/shared/src/hover/HoverOverlay.types'
import hoverOverlayContentsStyle from '@sourcegraph/shared/src/hover/HoverOverlayContents.module.scss'
import { HoverOverlayContent } from '@sourcegraph/shared/src/hover/HoverOverlayContents/HoverOverlayContent'
import { Alert, Card, WildcardThemeContext } from '@sourcegraph/wildcard'

import webHoverOverlayStyle from '../../../components/WebHoverOverlay/WebHoverOverlay.module.scss'

const HOVER_TIMEOUT = 50

const hoverCardTheme = EditorView.theme({
    '.cm-code-intel-hovercard': {
        fontFamily: 'sans-serif',
        minWidth: '10rem',
        maxWidth: '35rem',
    },
    '.cm-code-intel-contents': {
        maxHeight: '25rem',
    },
    '.cm-code-intel-content': {
        maxHeight: '25rem',
    },
    '.cm-tooltip': {
        border: 'initial',
        backgroundColor: 'initial',
    },
})

/**
 * Registers an extension to show a hovercard
 */
export function hovercard(
    source: (
        view: EditorView,
        position: number,
        side: 1 | -1
    ) => Promise<
        (Omit<Tooltip, 'create'> & { props: Pick<HoverOverlayBaseProps, 'hoverOrError' | 'actionsOrError'> }) | null
    > | null
): Extension {
    return [
        // hoverTooltip takes care of only calling the source (and processing
        // its return value) when necessary.
        hoverTooltip(
            async (view, ...args) => {
                const result = await source(view, ...args)
                if (!result) {
                    return null
                }
                const { props, ...tooltip } = result

                return {
                    ...tooltip,
                    create() {
                        const container = document.createElement('div')
                        return {
                            dom: container,
                            overlap: true,

                            mount() {
                                const root = createRoot(container)
                                root.render(
                                    <Hovercard
                                        {...props}
                                        onRender={() => {
                                            // Trigger repositioning after component rendered to ensure that
                                            // its position is account for its width and height
                                            repositionTooltips(view)
                                        }}
                                    />
                                )
                            },
                        }
                    },
                }
            },
            {
                hoverTime: HOVER_TIMEOUT,
            }
        ),
        hoverCardTheme,
    ]
}

/**
 * A simple replication of the hovercard for the old blob view.
 * TODO: Reuse existing hovercard component for full feature parity
 */
export const Hovercard: React.FunctionComponent<
    Pick<HoverOverlayBaseProps, 'hoverOrError'> & { onRender: () => void }
> = ({ hoverOrError, onRender }) => {
    useEffect(onRender, [onRender])

    if (isErrorLike(hoverOrError)) {
        return <Alert className={hoverOverlayStyle.hoverError}>{upperFirst(hoverOrError.message)}</Alert>
    }

    if (hoverOrError === undefined || hoverOrError === 'loading') {
        return null
    }

    if (hoverOrError === null || (hoverOrError.contents.length === 0 && hoverOrError.alerts?.length)) {
        return (
            // Show some content to give the close button space and communicate to the user we couldn't find a hover.
            <small className={hoverOverlayStyle.hoverEmpty}>No hover information available.</small>
        )
    }

    return (
        <WildcardThemeContext.Provider value={{ isBranded: true }}>
            <Card
                className={`${hoverOverlayStyle.card} ${webHoverOverlayStyle.webHoverOverlay} cm-code-intel-hovercard`}
            >
                <div className={`${hoverOverlayContentsStyle.hoverOverlayContents} cm-code-intel-contents`}>
                    {hoverOrError.contents.map((content, index) => (
                        <HoverOverlayContent
                            key={index}
                            index={index}
                            content={content}
                            aggregatedBadges={hoverOrError.aggregatedBadges}
                            contentClassName="cm-code-intel-content"
                        />
                    ))}
                </div>
            </Card>
        </WildcardThemeContext.Provider>
    )
}
