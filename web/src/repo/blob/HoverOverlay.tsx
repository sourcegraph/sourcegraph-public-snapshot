import Loader from '@sourcegraph/icons/lib/Loader'
import { highlight, highlightAuto } from 'highlight.js/lib/highlight'
import { castArray, escape, upperFirst } from 'lodash'
import marked from 'marked'
import AlertCircleOutlineIcon from 'mdi-react/AlertCircleOutlineIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { MarkedString, MarkupContent, MarkupKind, Position } from 'vscode-languageserver-types'
import { PositionSpec, RangeSpec, RepoFile, ViewStateSpec } from '..'
import { HoverMerged } from '../../backend/features'
import { eventLogger } from '../../tracking/eventLogger'
import { asError, ErrorLike, isErrorLike } from '../../util/errors'
import { toPrettyBlobURL } from '../../util/url'

/**
 * Attempts to syntax-highlight the given code.
 * If the language is not given, it is auto-detected.
 * If an error occurs, the code is returned as plain text with escaped HTML entities
 *
 * @param code The code to highlight
 * @param language The language of the code, if known
 * @return Safe HTML
 */
const highlightCodeSafe = (code: string, language?: string): string => {
    try {
        if (language) {
            return highlight(language, code, true).value
        }
        return highlightAuto(code).value
    } catch (err) {
        console.warn('Error syntax-highlighting hover markdown code block', err)
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
    hoverOrError?: typeof LOADING | HoverMerged | null | ErrorLike

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

    /** A ref callback to get the root overlay element. Use this to calculate the position. */
    hoverRef?: React.Ref<HTMLElement>

    /**
     * The hovered token (position and word).
     * Used for the Find References/Implementations buttons and for error messages
     */
    hoveredTokenPosition?: Position

    /** Whether to show the close button for the hover overlay */
    showCloseButton: boolean

    /** Called when the close button is clicked */
    onCloseButtonClick: (event: React.MouseEvent<HTMLElement>) => void
}

const onFindImplementationsClick = () => eventLogger.log('FindImplementationsClicked')
const onFindReferencesClick = () => eventLogger.log('FindRefsClicked')

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
                      visibility: 'visible',
                      left: props.overlayPosition.left + 'px',
                      top: props.overlayPosition.top + 'px',
                  }
                : {
                      opacity: 0,
                      visibility: 'hidden',
                  }
        }
    >
        {props.showCloseButton && (
            <button className="hover-overlay__close-button btn btn-icon" onClick={props.onCloseButtonClick}>
                <CloseIcon className="icon-inline" />
            </button>
        )}
        {props.hoverOrError && (
            <div className="hover-overlay__contents">
                {props.hoverOrError === LOADING ? (
                    <div className="hover-overlay__row hover-overlay__loader-row">
                        <Loader className="icon-inline" />
                    </div>
                ) : isErrorLike(props.hoverOrError) ? (
                    <div className="hover-overlay__row hover-overlay__hover-error lert alert-danger">
                        <h4>
                            <AlertCircleOutlineIcon className="icon-inline" /> Error fetching hover from language
                            server:
                        </h4>
                        {upperFirst(props.hoverOrError.message)}
                    </div>
                ) : (
                    // tslint:disable-next-line deprecation We want to handle the deprecated MarkedString
                    castArray<MarkedString | MarkupContent>(props.hoverOrError.contents)
                        .map(value => (typeof value === 'string' ? { kind: MarkupKind.Markdown, value } : value))
                        .map((content, i) => {
                            if (MarkupContent.is(content)) {
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
                                                className="hover-overlay__content hover-overlay__row e2e-tooltip-content"
                                                key={i}
                                                dangerouslySetInnerHTML={{ __html: rendered }}
                                            />
                                        )
                                    } catch (err) {
                                        return (
                                            <div className="hover-overlay__row alert alert-danger">
                                                <strong>
                                                    <AlertCircleOutlineIcon className="icon-inline" /> Error rendering
                                                    hover content
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
                                    className="hover-overlay__content hover-overlay__row e2e-tooltip-content"
                                    key={i}
                                    dangerouslySetInnerHTML={{
                                        __html: highlightCodeSafe(content.value, content.language),
                                    }}
                                />
                            )
                        })
                )}
            </div>
        )}

        <div className="hover-overlay__actions hover-overlay__row">
            <ButtonOrLink
                to={isJumpURL(props.definitionURLOrError) ? props.definitionURLOrError.jumpURL : undefined}
                className="btn btn-secondary hover-overlay__action e2e-tooltip-j2d"
                onClick={props.onGoToDefinitionClick}
            >
                Go to definition {props.definitionURLOrError === LOADING && <Loader className="icon-inline" />}
            </ButtonOrLink>
            <ButtonOrLink
                onClick={onFindReferencesClick}
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
                className="btn btn-secondary hover-overlay__action e2e-tooltip-find-refs"
            >
                Find references
            </ButtonOrLink>
            <ButtonOrLink
                onClick={onFindImplementationsClick}
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
                className="btn btn-secondary hover-overlay__action e2e-tooltip-find-impl"
            >
                Find implementations
            </ButtonOrLink>
        </div>
        {props.definitionURLOrError === null ? (
            <div className="alert alert-info hover-overlay__alert-below">
                <InformationOutlineIcon className="icon-inline" /> No definition found
            </div>
        ) : (
            isErrorLike(props.definitionURLOrError) && (
                <div className="alert alert-danger hover-overlay__alert-below">
                    <strong>
                        <AlertCircleOutlineIcon className="icon-inline" /> Error finding definition:
                    </strong>{' '}
                    {upperFirst(props.definitionURLOrError.message)}
                </div>
            )
        )}
    </div>
)
