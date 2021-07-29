import Dialog from '@reach/dialog'
import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { useCallback, useState } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import styles from './HomepageModalVideo.module.scss'

const THREE_WAYS_TO_SEARCH_TITLE = 'three-ways-to-search-title'

export const HomepageModalVideo: React.FunctionComponent<ThemeProps & TelemetryProps> = ({
    isLightTheme,
    telemetryService,
}) => {
    const assetsRoot = window.context?.assetsRoot || ''
    const [isOpen, setIsOpen] = useState(false)
    const toggleDialog = useCallback(
        isOpen => {
            telemetryService.log(isOpen ? 'HomepageVideoWaysToSearchClicked' : 'HomepageVideoClosed')
            setIsOpen(isOpen)
        },
        [telemetryService, setIsOpen]
    )

    return (
        <>
            <div className={styles.wrapper}>
                <button type="button" className={styles.thumbnailButton} onClick={() => toggleDialog(true)}>
                    <img
                        src={`${assetsRoot}/img/watch-and-learn-${isLightTheme ? 'light' : 'dark'}.png`}
                        alt="Watch and learn video thumbnail"
                        className={styles.thumbnailImage}
                    />
                    <div className={styles.playIconWrapper}>
                        <PlayIcon />
                    </div>
                </button>
                <div className="text-center mt-2">
                    <button
                        className="btn btn-link font-weight-normal p-0"
                        type="button"
                        onClick={() => toggleDialog(true)}
                    >
                        Three ways to search
                    </button>
                </div>
            </div>
            {isOpen && (
                <Dialog
                    className={classNames(styles.modal, 'modal-body modal-body--centered p-4 rounded border')}
                    onDismiss={() => toggleDialog(false)}
                    aria-labelledby={THREE_WAYS_TO_SEARCH_TITLE}
                >
                    <div className={styles.modalContent}>
                        <div className={styles.modalHeader}>
                            <h3 id={THREE_WAYS_TO_SEARCH_TITLE}>Three ways to search</h3>
                            <button
                                type="button"
                                className="btn btn-icon p-1"
                                onClick={() => toggleDialog(false)}
                                aria-label="Close"
                            >
                                <CloseIcon className="icon-inline" />
                            </button>
                        </div>
                        <div className="w-100">
                            <div className={styles.iframeVideoWrapper}>
                                <iframe
                                    className={styles.iframeVideo}
                                    src="https://www.youtube-nocookie.com/embed/XLfE2YuRwvw"
                                    title="YouTube video player"
                                    frameBorder="0"
                                    allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
                                    allowFullScreen={true}
                                />
                            </div>
                        </div>
                    </div>
                </Dialog>
            )}
        </>
    )
}

const PlayIcon = React.memo(() => (
    <svg width="33" height="33" viewBox="0 0 49 53" fill="none" xmlns="http://www.w3.org/2000/svg">
        <g filter="url(#filter0_dd)">
            <path d="M37 26.5L12.25 40.79V12.21L37 26.5z" fill="#fff" />
        </g>
        <defs>
            <filter
                id="filter0_dd"
                x=".25"
                y=".211"
                width="48.75"
                height="52.579"
                filterUnits="userSpaceOnUse"
                colorInterpolationFilters="sRGB"
            >
                <feFlood floodOpacity="0" result="BackgroundImageFix" />
                <feColorMatrix in="SourceAlpha" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0" result="hardAlpha" />
                <feOffset />
                <feGaussianBlur stdDeviation="6" />
                <feColorMatrix values="0 0 0 0 0.00505209 0 0 0 0 0.0449636 0 0 0 0 0.404167 0 0 0 0.25 0" />
                <feBlend in2="BackgroundImageFix" result="effect1_dropShadow" />
                <feColorMatrix in="SourceAlpha" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0" result="hardAlpha" />
                <feOffset dy="4" />
                <feGaussianBlur stdDeviation="2" />
                <feColorMatrix values="0 0 0 0 0 0 0 0 0 0.055 0 0 0 0 0.25 0 0 0 0.25 0" />
                <feBlend in2="effect1_dropShadow" result="effect2_dropShadow" />
                <feBlend in="SourceGraphic" in2="effect2_dropShadow" result="shape" />
            </filter>
        </defs>
    </svg>
))
