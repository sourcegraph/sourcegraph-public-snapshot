import Loader from '@sourcegraph/icons/lib/Loader'
import { highlight, highlightAuto } from 'highlight.js/lib/highlight'
import { castArray } from 'lodash'
import marked from 'marked'
import * as React from 'react'
import { Hover, MarkedString, MarkupContent, MarkupKind } from 'vscode-languageserver-types'
import { ErrorLike, isErrorLike } from '../../util/errors'

const isMarkupContent = (markup: any): markup is MarkupContent =>
    typeof markup === 'object' && markup !== null && 'kind' in markup

const renderMarkdown = (markdown: string) =>
    marked(markdown, {
        gfm: true,
        breaks: true,
        sanitize: true,
        highlight: (code, language) =>
            '<code>' + (language ? highlight(language, code, true).value : highlightAuto(code).value) + '</code>',
    })

const LOADING: 'loading' = 'loading'

interface HoverOverlayProps {
    /** What to show as contents */
    hoverOrError: typeof LOADING | Hover | ErrorLike
    /** The position of the tooltip (assigned to `style`) */
    position?: { left: number; top: number }
    /** Whether this tooltip is fixed or not. Determines whether actions are shown or not. */
    isFixed: boolean
    /** A ref callback to get the root overlay element. Use this to calculate the position. */
    hoverRef?: React.Ref<HTMLElement>
}

export const HoverOverlay: React.StatelessComponent<HoverOverlayProps> = props => (
    <div
        className="hover-overlay card"
        ref={props.hoverRef}
        // tslint:disable-next-line:jsx-ban-props needed for dynamic styling
        style={
            props.position
                ? {
                      opacity: 1,
                      left: props.position.left + 'px',
                      top: props.position.top + 'px',
                  }
                : {
                      opacity: 0,
                  }
        }
    >
        <div className="hover-overlay__contents">
            {props.hoverOrError === LOADING ? (
                <div className="text-center p-1">
                    <Loader className="icon-inline" />
                </div>
            ) : isErrorLike(props.hoverOrError) ? (
                <div className="hover-overlay__content hover-content__error">{props.hoverOrError.message}</div>
            ) : (
                // tslint:disable-next-line deprecation We want to handle the deprecated MarkedString
                castArray<MarkedString | MarkupContent>(props.hoverOrError.contents)
                    .map(value => (typeof value === 'string' ? { kind: MarkupKind.Markdown, value } : value))
                    .map(
                        (content, i) =>
                            isMarkupContent(content) ? (
                                content.kind === MarkupKind.Markdown ? (
                                    <div
                                        className="hover-overlay__content"
                                        key={i}
                                        dangerouslySetInnerHTML={{
                                            __html: renderMarkdown(content.value),
                                        }}
                                    />
                                ) : (
                                    content.value
                                )
                            ) : (
                                <code
                                    className="hover-overlay__content"
                                    key={i}
                                    dangerouslySetInnerHTML={{
                                        __html: highlight(content.language, content.value).value,
                                    }}
                                />
                            )
                    )
            )}
        </div>

        <div className="hover-overlay__actions">
            {props.isFixed ? (
                <>
                    <button className="btn btn-secondary hover-overlay__action">Go to definition</button>
                    <button className="btn btn-secondary hover-overlay__action">Find references</button>
                    <button className="btn btn-secondary hover-overlay__action">Find implementations</button>
                </>
            ) : (
                <button className="btn btn-secondary hover-overlay__actions-placeholder" disabled={true}>
                    <em>Click for actions</em>
                </button>
            )}
        </div>
    </div>
)
