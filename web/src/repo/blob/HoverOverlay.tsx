import Loader from '@sourcegraph/icons/lib/Loader'
import { highlight, highlightAuto } from 'highlight.js/lib/highlight'
import { castArray, escape, upperFirst } from 'lodash'
import marked from 'marked'
import AlertCircleOutlineIcon from 'mdi-react/AlertCircleOutlineIcon'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Hover, MarkedString, MarkupContent, MarkupKind, Position } from 'vscode-languageserver-types'
import { PositionSpec, RangeSpec, RepoFile, ViewStateSpec } from '..'
import { asError, ErrorLike, isErrorLike } from '../../util/errors'
import { toPrettyBlobURL } from '../../util/url'

const isMarkupContent = (markup: any): markup is MarkupContent =>
    typeof markup === 'object' && markup !== null && 'kind' in markup

const highlightCodeSafe = (code: string, language?: string): string => {
    try {
        if (language) {
            return highlight(language, code, true).value
        }
        return highlightAuto(code).value
    } catch (err) {
        console.error('Error syntax-highlighting hover markdown code block', err)
        return escape(code)
    }
}

/**
 * Uses a placeholder `<button>` or a React Router `<Link>` depending on whether `to` is set.
 */
const ButtonOrLink: React.StatelessComponent<{ to?: string } & React.HTMLAttributes<HTMLElement>> = props =>
    props.to ? (
        <Link to={props.to} {...props}>
            {props.children}
        </Link>
    ) : (
        <button {...props}>{props.children}</button>
    )

const LOADING: 'loading' = 'loading'

interface HoverOverlayProps extends RepoFile, Partial<PositionSpec>, Partial<ViewStateSpec>, Partial<RangeSpec> {
    /** What to show as contents */
    hoverOrError: typeof LOADING | Hover | ErrorLike

    /**
     * The URL to jump to on go to definition.
     * If loaded, is set as the href of the go to definition button.
     * If LOADING, a loader is displayed on the button.
     * If null, an info alert is displayed "no definition found".
     * If an error, an error alert is displayed with the error message.
     */
    definitionURLOrError?: typeof LOADING | { jumpURL: string } | null | ErrorLike

    /** Called when the Go-to-definition button was clicked */
    onGoToDefinitionClick: (event: React.MouseEvent<HTMLElement>) => void

    /** The position of the tooltip (assigned to `style`) */
    overlayPosition?: { left: number; top: number }

    /** Whether this tooltip is fixed or not. Determines whether actions are shown or not. */
    isFixed: boolean

    /** A ref callback to get the root overlay element. Use this to calculate the position. */
    hoverRef?: React.Ref<HTMLElement>

    /**
     * The hovered token (position and word).
     * Used for the Find References/Implementations buttons and for error messages
     */
    hoveredTokenPosition?: Position
}

/** Returns true if the input is successful jump URL result */
export const isJumpURL = (val: any): val is { jumpURL: string } =>
    val && typeof val === 'object' && typeof val.jumpURL === 'string'

export const HoverOverlay: React.StatelessComponent<HoverOverlayProps> = props => (
    <div
        className="hover-overlay card"
        ref={props.hoverRef}
        // tslint:disable-next-line:jsx-ban-props needed for dynamic styling
        style={
            props.overlayPosition
                ? {
                      opacity: 1,
                      left: props.overlayPosition.left + 'px',
                      top: props.overlayPosition.top + 'px',
                  }
                : {
                      opacity: 0,
                  }
        }
    >
        <div className="hover-overlay__contents">
            {props.hoverOrError === LOADING ? (
                <div className="hover-overlay__row hover-overlay__loader-row">
                    <Loader className="icon-inline" />
                </div>
            ) : isErrorLike(props.hoverOrError) ? (
                <div className="hover-overlay__row alert alert-danger">
                    <AlertCircleOutlineIcon className="icon-inline" /> {upperFirst(props.hoverOrError.message)}
                </div>
            ) : (
                // tslint:disable-next-line deprecation We want to handle the deprecated MarkedString
                castArray<MarkedString | MarkupContent>(props.hoverOrError.contents)
                    .map(value => (typeof value === 'string' ? { kind: MarkupKind.Markdown, value } : value))
                    .map((content, i) => {
                        if (isMarkupContent(content)) {
                            if (content.kind === MarkupKind.Markdown) {
                                try {
                                    const rendered = marked(content.value, {
                                        gfm: true,
                                        breaks: true,
                                        sanitize: true,
                                        highlight: (code, language) =>
                                            '<code>' + highlightCodeSafe(code, language) + '</code>',
                                    })
                                    return (
                                        <div
                                            className="hover-overlay__content hover-overlay__row"
                                            key={i}
                                            dangerouslySetInnerHTML={{ __html: rendered }}
                                        />
                                    )
                                } catch (err) {
                                    return (
                                        <div className="hover-overlay__row alert alert-danger">
                                            <strong>
                                                <AlertCircleOutlineIcon className="icon-inline" /> Error rendering hover{' '}
                                                content
                                            </strong>{' '}
                                            {upperFirst(asError(err).message)}
                                        </div>
                                    )
                                }
                            }
                            return content.value
                        }
                        return (
                            <code
                                className="hover-overlay__content hover-overlay__row"
                                key={i}
                                dangerouslySetInnerHTML={{ __html: highlightCodeSafe(content.value, content.language) }}
                            />
                        )
                    })
            )}
        </div>

        <div className="hover-overlay__actions hover-overlay__row">
            {props.isFixed ? (
                <>
                    <ButtonOrLink
                        to={isJumpURL(props.definitionURLOrError) ? props.definitionURLOrError.jumpURL : undefined}
                        className="btn btn-secondary hover-overlay__action"
                        onClick={props.onGoToDefinitionClick}
                    >
                        Go to definition {props.definitionURLOrError === LOADING && <Loader className="icon-inline" />}
                    </ButtonOrLink>
                    <ButtonOrLink
                        to={
                            props.hoveredTokenPosition &&
                            toPrettyBlobURL({
                                repoPath: props.repoPath,
                                commitID: props.commitID,
                                rev: props.rev,
                                filePath: props.filePath,
                                position: props.hoveredTokenPosition,
                                range: props.range,
                                viewState: 'references',
                            })
                        }
                        className="btn btn-secondary hover-overlay__action"
                    >
                        Find references
                    </ButtonOrLink>
                    <ButtonOrLink
                        to={
                            props.hoveredTokenPosition &&
                            toPrettyBlobURL({
                                repoPath: props.repoPath,
                                commitID: props.commitID,
                                rev: props.rev,
                                filePath: props.filePath,
                                position: props.hoveredTokenPosition,
                                range: props.range,
                                viewState: 'impl',
                            })
                        }
                        className="btn btn-secondary hover-overlay__action"
                    >
                        Find implementations
                    </ButtonOrLink>
                </>
            ) : (
                <button className="btn btn-secondary hover-overlay__actions-placeholder" disabled={true}>
                    <em>Click for actions</em>
                </button>
            )}
        </div>
        {props.definitionURLOrError === null ? (
            <div className="alert alert-info m-0 p-2 rounded-0">
                <InformationOutlineIcon className="icon-inline" /> No definition found
            </div>
        ) : (
            isErrorLike(props.definitionURLOrError) && (
                <div className="alert alert-danger m-0 p-2 rounded-0">
                    <strong>
                        <AlertCircleOutlineIcon className="icon-inline" /> Error finding definition:
                    </strong>{' '}
                    {upperFirst(props.definitionURLOrError.message)}
                </div>
            )
        )}
    </div>
)
