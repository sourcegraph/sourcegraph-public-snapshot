import { useState } from 'react'

import { mdiClose } from '@mdi/js'

import { Button, Icon } from '@sourcegraph/wildcard'

import styles from './ContextSelectorCallout.module.scss'

export const ContextSelectorCallout: React.FC = () => {
    const [isCalloutOpen, setCalloutOpen] = useState(true)

    const handleCloseClick = () => {
        setCalloutOpen(false)
    }

    if (!isCalloutOpen) {
        return null
    }

    return (
        <div className={styles.wrapper}>
            <div className={styles.box}>
                <div className={styles.header}>
                    <div className={styles.headerElements}>
                        <CodyCalloutIcon /> Give Cody context
                    </div>
                    <Button className={styles.closeButton} onClick={handleCloseClick} variant="icon" aria-label="Close">
                        <Icon aria-hidden={true} svgPath={mdiClose} />
                    </Button>
                </div>
                <div className={styles.contentBody}>
                    Tell Cody what codebases it should reference to help with your task and Cody will respond more
                    accurately.
                </div>
            </div>
            <div className={styles.tail} />
        </div>
    )
}

const CodyCalloutIcon: React.FC = () => (
    <svg width="30" height="29" viewBox="0 0 30 29" fill="none" xmlns="http://www.w3.org/2000/svg">
        <g filter="url(#filter0_dd_4_712)">
            <rect x="4" y="3" width="21.6955" height="20.5137" rx="4" fill="#E8D1FF" />
            <path
                fill-rule="evenodd"
                clip-rule="evenodd"
                d="M17.9743 7.4679C18.6277 7.4679 19.1573 7.9976 19.1573 8.65103L19.1573 11.3553C19.1573 12.0088 18.6277 12.5385 17.9743 12.5385C17.3209 12.5385 16.7913 12.0088 16.7913 11.3553L16.7913 8.65103C16.7913 7.9976 17.3209 7.4679 17.9743 7.4679Z"
                fill="#A305E1"
            />
            <path
                fill-rule="evenodd"
                clip-rule="evenodd"
                d="M9.18616 10.5102C9.18616 9.8568 9.71581 9.32709 10.3692 9.32709H13.0732C13.7266 9.32709 14.2562 9.8568 14.2562 10.5102C14.2562 11.1637 13.7266 11.6934 13.0732 11.6934H10.3692C9.71581 11.6934 9.18616 11.1637 9.18616 10.5102Z"
                fill="#A112FF"
            />
            <path
                fill-rule="evenodd"
                clip-rule="evenodd"
                d="M10.5502 14.78C10.1569 14.2628 9.41937 14.1594 8.89884 14.5498C8.37615 14.9419 8.27022 15.6835 8.66224 16.2062L9.60866 15.4963C8.66224 16.2062 8.66257 16.2066 8.66291 16.2071L8.66363 16.208L8.66518 16.2101L8.66882 16.2149C8.67147 16.2184 8.67458 16.2224 8.67815 16.2271C8.68529 16.2363 8.69428 16.2478 8.70511 16.2613C8.72677 16.2884 8.75584 16.3239 8.79232 16.3666C8.86523 16.452 8.96814 16.5664 9.10105 16.7002C9.36622 16.9669 9.75492 17.3145 10.2673 17.6602C11.2958 18.3542 12.832 19.0457 14.8477 19.0457C16.8635 19.0457 18.3997 18.3542 19.4282 17.6602C19.9406 17.3145 20.3293 16.9669 20.5944 16.7002C20.7273 16.5664 20.8302 16.452 20.9032 16.3666C20.9396 16.3239 20.9687 16.2884 20.9904 16.2613C21.0012 16.2478 21.0102 16.2363 21.0173 16.2271C21.0209 16.2224 21.024 16.2184 21.0267 16.2149L21.0303 16.2101L21.0318 16.208L21.0326 16.2071C21.0329 16.2066 21.0332 16.2062 20.0868 15.4963L21.0332 16.2062C21.4253 15.6835 21.3193 14.9419 20.7966 14.5498C20.2761 14.1594 19.5386 14.2628 19.1453 14.78C19.1446 14.7808 19.1434 14.7823 19.1417 14.7845C19.1357 14.792 19.1233 14.8073 19.1045 14.8293C19.0668 14.8734 19.004 14.9438 18.9164 15.0319C18.7406 15.2088 18.4691 15.4529 18.1048 15.6987C17.3799 16.1878 16.2965 16.6795 14.8477 16.6795C13.3989 16.6795 12.3156 16.1878 11.5907 15.6987C11.2264 15.4529 10.9549 15.2088 10.7791 15.0319C10.6915 14.9438 10.6287 14.8734 10.591 14.8293C10.5722 14.8073 10.5598 14.792 10.5537 14.7845C10.552 14.7823 10.5509 14.7808 10.5502 14.78ZM19.1418 14.7846L19.1416 14.7848L19.141 14.7856C19.1407 14.786 19.1404 14.7864 20.0563 15.4734L19.1404 14.7864C19.1409 14.7858 19.1413 14.7852 19.1418 14.7846Z"
                fill="#A305E1"
            />
        </g>
        <defs>
            <filter
                id="filter0_dd_4_712"
                x="0.12756"
                y="0.233971"
                width="29.4404"
                height="28.2585"
                filterUnits="userSpaceOnUse"
                color-interpolation-filters="sRGB"
            >
                <feFlood flood-opacity="0" result="BackgroundImageFix" />
                <feColorMatrix
                    in="SourceAlpha"
                    type="matrix"
                    values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0"
                    result="hardAlpha"
                />
                <feOffset dy="1.10641" />
                <feGaussianBlur stdDeviation="0.829809" />
                <feColorMatrix type="matrix" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0.05 0" />
                <feBlend mode="normal" in2="BackgroundImageFix" result="effect1_dropShadow_4_712" />
                <feColorMatrix
                    in="SourceAlpha"
                    type="matrix"
                    values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0"
                    result="hardAlpha"
                />
                <feMorphology radius="0.829809" operator="erode" in="SourceAlpha" result="effect2_dropShadow_4_712" />
                <feOffset dy="1.10641" />
                <feGaussianBlur stdDeviation="2.35112" />
                <feColorMatrix type="matrix" values="0 0 0 0 0.556863 0 0 0 0 0.207843 0 0 0 0 0.956863 0 0 0 0.42 0" />
                <feBlend mode="normal" in2="effect1_dropShadow_4_712" result="effect2_dropShadow_4_712" />
                <feBlend mode="normal" in="SourceGraphic" in2="effect2_dropShadow_4_712" result="shape" />
            </filter>
        </defs>
    </svg>
)
