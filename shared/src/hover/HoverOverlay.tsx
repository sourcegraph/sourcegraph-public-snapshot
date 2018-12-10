import { HoverOverlayProps as GenericHoverOverlayProps } from '@sourcegraph/codeintellify'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { castArray, isEqual } from 'lodash'
import AlertCircleOutlineIcon from 'mdi-react/AlertCircleOutlineIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import * as React from 'react'
import { MarkupContent } from 'sourcegraph'
import { ActionItem, ActionItemComponentProps, ActionItemProps } from '../actions/ActionItem'
import { HoverMerged } from '../api/client/types/hover'
import { TelemetryContext } from '../telemetry/telemetryContext'
import { TelemetryService } from '../telemetry/telemetryService'
import { isErrorLike } from '../util/errors'
import { FileSpec, RepoSpec, ResolvedRevSpec, RevSpec } from '../util/url'
import { highlightCodeSafe, renderMarkdown, toNativeEvent } from './helpers'

const LOADING: 'loading' = 'loading'

const transformMouseEvent = (handler: (event: MouseEvent) => void) => (event: React.MouseEvent<HTMLElement>) =>
    handler(toNativeEvent(event))

export type HoverContext = RepoSpec & RevSpec & FileSpec & ResolvedRevSpec

export type HoverData = HoverMerged

export interface HoverOverlayProps
    extends GenericHoverOverlayProps<HoverContext, HoverData, ActionItemProps>,
        ActionItemComponentProps {
    /** A ref callback to get the root overlay element. Use this to calculate the position. */
    hoverRef?: React.Ref<HTMLDivElement>

    /** An optional class name to apply to the outermost element of the HoverOverlay */
    className?: string

    /** Called when the close button is clicked */
    onCloseButtonClick?: (event: MouseEvent) => void
}

const isEmptyHover = ({
    hoveredToken,
    hoverOrError,
    actionsOrError,
}: Pick<HoverOverlayProps, 'hoveredToken' | 'hoverOrError' | 'actionsOrError'>): boolean =>
    !hoveredToken ||
    ((!hoverOrError || hoverOrError === LOADING || isErrorLike(hoverOrError)) &&
        (!actionsOrError || actionsOrError === LOADING || isErrorLike(actionsOrError)))

class BaseHoverOverlay extends React.PureComponent<HoverOverlayProps & { telemetryService: TelemetryService }> {
    public componentDidMount(): void {
        this.logTelemetryEvent()
    }

    public componentDidUpdate(prevProps: HoverOverlayProps): void {
        // Log a telemetry event for this hover being displayed, but only do it once per position and when it is
        // non-empty.
        if (
            !isEmptyHover(this.props) &&
            (!isEqual(this.props.hoveredToken, prevProps.hoveredToken) || isEmptyHover(prevProps))
        ) {
            this.logTelemetryEvent()
        }
    }

    public render(): JSX.Element | null {
        const {
            hoverOrError,
            hoverRef,
            onCloseButtonClick,
            overlayPosition,
            showCloseButton,
            actionsOrError,
            className = '',
            extensionsController,
            platformContext,
            location,
        } = this.props

        return (
            <div
                className={`hover-overlay card ${className}`}
                ref={hoverRef}
                // tslint:disable-next-line:jsx-ban-props needed for dynamic styling
                style={
                    overlayPosition
                        ? {
                              opacity: 1,
                              visibility: 'visible',
                              left: overlayPosition.left + 'px',
                              top: overlayPosition.top + 'px',
                          }
                        : {
                              opacity: 0,
                              visibility: 'hidden',
                          }
                }
            >
                {showCloseButton && (
                    <button
                        className="hover-overlay__close-button btn btn-icon"
                        onClick={onCloseButtonClick ? transformMouseEvent(onCloseButtonClick) : undefined}
                    >
                        <CloseIcon className="icon-inline" />
                    </button>
                )}
                <div className="hover-overlay__contents">
                    {hoverOrError === LOADING ? (
                        <div className="hover-overlay__row hover-overlay__loader-row">
                            <LoadingSpinner className="icon-inline" />
                        </div>
                    ) : isErrorLike(hoverOrError) ? (
                        <div className="hover-overlay__row hover-overlay__hover-error alert alert-danger">
                            <h4>
                                <AlertCircleOutlineIcon className="icon-inline" /> Error:
                            </h4>{' '}
                            {hoverOrError.message}
                        </div>
                    ) : (
                        // tslint:disable-next-line deprecation We want to handle the deprecated MarkedString
                        hoverOrError &&
                        castArray<string | MarkupContent | { language: string; value: string }>(hoverOrError.contents)
                            .map(value => (typeof value === 'string' ? { kind: 'markdown', value } : value))
                            .map((content, i) => {
                                if ('kind' in content || !('language' in content)) {
                                    if (content.kind === 'markdown') {
                                        try {
                                            return (
                                                <div
                                                    className="hover-overlay__content hover-overlay__row e2e-tooltip-content"
                                                    key={i}
                                                    dangerouslySetInnerHTML={{ __html: renderMarkdown(content.value) }}
                                                />
                                            )
                                        } catch (err) {
                                            return (
                                                <div className="hover-overlay__row alert alert-danger" key={i}>
                                                    <strong>
                                                        <AlertCircleOutlineIcon className="icon-inline" /> Error:
                                                    </strong>{' '}
                                                    {err.message}
                                                </div>
                                            )
                                        }
                                    }
                                    return (
                                        <div className="hover-overlay__content hover-overlay__row" key={i}>
                                            {String(content.value)}
                                        </div>
                                    )
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
                {actionsOrError !== undefined &&
                    actionsOrError !== null &&
                    actionsOrError !== LOADING &&
                    !isErrorLike(actionsOrError) &&
                    actionsOrError.length > 0 && (
                        <div className="hover-overlay__actions hover-overlay__row">
                            {actionsOrError.map((action, i) => (
                                <ActionItem
                                    key={i}
                                    className="btn btn-secondary hover-overlay__action e2e-tooltip-j2d"
                                    {...action}
                                    variant="actionItem"
                                    disabledDuringExecution={true}
                                    showLoadingSpinnerDuringExecution={true}
                                    showInlineError={true}
                                    extensionsController={extensionsController}
                                    platformContext={platformContext}
                                    location={location}
                                />
                            ))}
                        </div>
                    )}
            </div>
        )
    }

    private logTelemetryEvent(): void {
        this.props.telemetryService.log('hover')
    }
}

export const HoverOverlay = React.forwardRef<BaseHoverOverlay, HoverOverlayProps>((props, ref) => (
    <TelemetryContext.Consumer>
        {telemetryService => <BaseHoverOverlay {...props} telemetryService={telemetryService} ref={ref} />}
    </TelemetryContext.Consumer>
))
