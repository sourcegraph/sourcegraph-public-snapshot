import Dialog from '@reach/dialog'
import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { useCallback, useState } from 'react'

import styles from './ModalVideo.module.scss'

interface ModalVideoProps {
    id: string
    title: string
    src: string
    thumbnail: { src: string; alt: string }
    onToggle?: (isOpen: boolean) => void
}

export const ModalVideo: React.FunctionComponent<ModalVideoProps> = ({ id, title, src, thumbnail, onToggle }) => {
    const assetsRoot = window.context?.assetsRoot || ''
    const [isOpen, setIsOpen] = useState(false)
    const toggleDialog = useCallback(
        isOpen => {
            setIsOpen(isOpen)
            if (onToggle) {
                onToggle(isOpen)
            }
        },
        [onToggle]
    )

    return (
        <>
            <div className={styles.wrapper}>
                <button type="button" className={styles.thumbnailButton} onClick={() => toggleDialog(true)}>
                    <img src={`${assetsRoot}/${thumbnail.src}`} alt={thumbnail.alt} className={styles.thumbnailImage} />
                    <div className={styles.playIconWrapper}>
                        <PlayIcon />
                    </div>
                </button>
            </div>
            {isOpen && (
                <Dialog
                    className={classNames(styles.modal, 'modal-body modal-body--centered p-4 rounded border')}
                    onDismiss={() => toggleDialog(false)}
                    aria-labelledby={id}
                >
                    <div className={styles.modalContent}>
                        <div className={styles.modalHeader}>
                            <h3 id={id}>{title}</h3>
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
                                    src={src}
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
