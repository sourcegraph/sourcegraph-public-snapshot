import { HoverOverlayProps as GenericHoverOverlayProps } from '@sourcegraph/codeintellify'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import classNames from 'classnames'
import { isEqual, upperFirst } from 'lodash'
import CloseIcon from 'mdi-react/CloseIcon'
import * as React from 'react'
import { ActionItem, ActionItemAction, ActionItemComponentProps } from '../actions/ActionItem'
import { HoverMerged } from '../api/client/types/hover'
import { TelemetryProps } from '../telemetry/telemetryService'
import { isErrorLike, asError } from '../util/errors'
import { renderMarkdown } from '../util/markdown'
import { sanitizeClass } from '../util/strings'
import { FileSpec, RepoSpec, ResolvedRevSpec, RevSpec } from '../util/url'
import { toNativeEvent } from './helpers'
import { BadgeAttachment } from '../components/BadgeAttachment'
import { ThemeProps } from '../theme'
import { PlatformContextProps } from '../platform/context'
import { Subscription } from 'rxjs'

const LOADING = 'loading' as const

const transformMouseEvent = (handler: (event: MouseEvent) => void) => (event: React.MouseEvent<HTMLElement>) =>
    handler(toNativeEvent(event))

export type HoverContext = RepoSpec & RevSpec & FileSpec & ResolvedRevSpec

export type HoverData<A extends string> = HoverMerged & HoverAlerts<A>

export interface HoverOverlayClassProps {
    /** An optional class name to apply to the outermost element of the HoverOverlay */
    className?: string
    closeButtonClassName?: string

    iconClassName?: string

    actionItemClassName?: string
    actionItemPressedClassName?: string

    infoAlertClassName?: string
    errorAlertClassName?: string
}

/**
 * A dismissable alert to be displayed in the hover overlay.
 */
export interface HoverAlert<T extends string> {
    /**
     * The type of the alert, eg. `'nativeTooltips'`
     */
    type: T
    /**
     * The content of the alert
     */
    content: React.ReactElement
}

/**
 * One or more dismissable that should be displayed before the hover content.
 * Alerts are only displayed in a non-empty hover.
 */
export interface HoverAlerts<A extends string> {
    alerts?: HoverAlert<A>[]
}

export interface HoverOverlayProps<A extends string>
    extends GenericHoverOverlayProps<HoverContext, HoverData<A>, ActionItemAction>,
        ActionItemComponentProps,
        HoverOverlayClassProps,
        TelemetryProps,
        ThemeProps,
        PlatformContextProps<'forceUpdateTooltip' | 'settings'> {
    /** A ref callback to get the root overlay element. Use this to calculate the position. */
    hoverRef?: React.Ref<HTMLDivElement>

    /** Called when the close button is clicked */
    onCloseButtonClick?: (event: MouseEvent) => void
    /** Called when an alert is dismissed, with the type of the dismissed alert. */
    onAlertDismissed?: (alertType: A) => void
}

interface HoverOverlayState {
    showBadges: boolean
}

const isEmptyHover = <A extends string>({
    hoveredToken,
    hoverOrError,
    actionsOrError,
}: Pick<HoverOverlayProps<A>, 'hoveredToken' | 'hoverOrError' | 'actionsOrError'>): boolean =>
    !hoveredToken ||
    ((!hoverOrError || hoverOrError === LOADING || isErrorLike(hoverOrError)) &&
        (!actionsOrError || actionsOrError === LOADING || isErrorLike(actionsOrError)))

export class HoverOverlay<A extends string> extends React.PureComponent<HoverOverlayProps<A>, HoverOverlayState> {
    private subscription = new Subscription()

    constructor(props: HoverOverlayProps<A>) {
        super(props)
        this.state = {
            showBadges: false,
        }
    }

    public componentDidMount(): void {
        this.logTelemetryEvent()

        this.subscription.add(
            this.props.platformContext.settings.subscribe(s => {
                if (s.final && !isErrorLike(s.final)) {
                    // Default to true if experimentalFeatures or showBadgeAttachments are not set
                    this.setState({
                        showBadges:
                            !s.final.experimentalFeatures ||
                            s.final.experimentalFeatures.showBadgeAttachments !== false,
                    })
                } else {
                    this.setState({ showBadges: false })
                }
            })
        )
    }

    public componentDidUpdate(prevProps: HoverOverlayProps<A>): void {
        // Log a telemetry event for this hover being displayed, but only do it once per position and when it is
        // non-empty.
        if (
            !isEmptyHover(this.props) &&
            (!isEqual(this.props.hoveredToken, prevProps.hoveredToken) || isEmptyHover(prevProps))
        ) {
            this.logTelemetryEvent()
        }
    }

    public componentWillUnmount(): void {
        this.subscription.unsubscribe()
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
            actionItemClassName,
            actionItemPressedClassName,
        } = this.props

        if (!hoverOrError && (!actionsOrError || isErrorLike(actionsOrError))) {
            return null
        }
        return (
            <div
                // needed for dynamic styling
                // eslint-disable-next-line react/forbid-dom-props
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
                className={classNames('hover-overlay', className)}
                ref={hoverRef}
            >
                {showCloseButton && (
                    <button
                        type="button"
                        className={classNames('hover-overlay__close-button', this.props.closeButtonClassName)}
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
                        <div
                            className={classNames(
                                'hover-overlay__row',
                                'hover-overlay__hover-error',
                                this.props.errorAlertClassName
                            )}
                        >
                            {upperFirst(hoverOrError.message)}
                        </div>
                    ) : (
                        hoverOrError?.contents.map((content, i) => {
                            if (content.kind === 'markdown') {
                                try {
                                    // Offset first badge when the close button is shown to avoid conflict.
                                    const offsetBadge = showCloseButton && i === 0
                                    return (
                                        <div className="hover-overlay__row e2e-tooltip-badged-content" key={i}>
                                            {'badge' in content && content.badge && this.state.showBadges && (
                                                <div
                                                    className={classNames(
                                                        'hover-overlay__badge',
                                                        'e2e-hover-badge',
                                                        offsetBadge && 'hover-overlay__badge--offset'
                                                    )}
                                                >
                                                    <BadgeAttachment
                                                        attachment={content.badge}
                                                        isLightTheme={this.props.isLightTheme}
                                                    />
                                                </div>
                                            )}

                                            <div
                                                className="hover-overlay__content e2e-tooltip-content"
                                                dangerouslySetInnerHTML={{
                                                    __html: renderMarkdown(content.value),
                                                }}
                                            />
                                        </div>
                                    )
                                } catch (err) {
                                    return (
                                        <div
                                            className={classNames('hover-overlay__row', this.props.errorAlertClassName)}
                                            key={i}
                                        >
                                            {upperFirst(asError(err).message)}
                                        </div>
                                    )
                                }
                            }
                            return (
                                <div className="hover-overlay__content hover-overlay__row" key={i}>
                                    {content.value}
                                </div>
                            )
                        })
                    )}
                </div>
                {hoverOrError && hoverOrError !== LOADING && !isErrorLike(hoverOrError) && hoverOrError.alerts && (
                    <div className="hover-overlay__alerts">
                        {hoverOrError.alerts.map(({ content, type }) => (
                            <div
                                className={classNames(
                                    'hover-overlay__row',
                                    'hover-overlay__alert',
                                    this.props.infoAlertClassName
                                )}
                                key={type}
                            >
                                <div className="hover-overlay__alert-content">
                                    <small>{content}</small>
                                    <a
                                        className="hover-overlay__alert-close"
                                        href=""
                                        onClick={this.onAlertDismissedCallback(type)}
                                    >
                                        <small>Dismiss</small>
                                    </a>
                                </div>
                            </div>
                        ))}
                    </div>
                )}
                {actionsOrError !== undefined &&
                    actionsOrError !== null &&
                    actionsOrError !== LOADING &&
                    !isErrorLike(actionsOrError) &&
                    actionsOrError.length > 0 && (
                        <div className="hover-overlay__actions hover-overlay__row">
                            {actionsOrError.map((action, i) => (
                                <ActionItem
                                    key={i}
                                    {...action}
                                    className={classNames(
                                        'hover-overlay__action',
                                        actionItemClassName,
                                        `e2e-tooltip-${sanitizeClass(action.action.title || 'untitled')}`
                                    )}
                                    iconClassName={this.props.iconClassName}
                                    pressedClassName={actionItemPressedClassName}
                                    variant="actionItem"
                                    disabledDuringExecution={true}
                                    showLoadingSpinnerDuringExecution={true}
                                    showInlineError={true}
                                    platformContext={this.props.platformContext}
                                    telemetryService={this.props.telemetryService}
                                    extensionsController={this.props.extensionsController}
                                    location={this.props.location}
                                />
                            ))}
                        </div>
                    )}
            </div>
        )
    }

    private onAlertDismissedCallback(alertType: A): (e: React.MouseEvent<HTMLAnchorElement>) => void {
        return e => {
            e.preventDefault()
            if (this.props.onAlertDismissed) {
                this.props.onAlertDismissed(alertType)
            }
        }
    }

    private logTelemetryEvent(): void {
        this.props.telemetryService.log('hover')
    }
}
