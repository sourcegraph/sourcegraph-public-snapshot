import React from 'react'

import classNames from 'classnames'

import { useTheme, Theme } from '@sourcegraph/shared/src/theme'

import styles from './PageHeaderIcon.module.scss'

interface PageHeaderIconProps {
    name: keyof typeof icons
    className?: string
}

// noinspection SpellCheckingInspection
const icons = {
    'cody-logo': {
        width: 37,
        height: 33,
        url: 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 38 34"><path fill="%23FF5543" fill-rule="evenodd" d="M28.24.45a3.38 3.38 0 0 1 3.39 3.39v7.73a3.38 3.38 0 1 1-6.77 0V3.84A3.38 3.38 0 0 1 28.24.45Z" clip-rule="evenodd"/><path fill="%23A112FF" fill-rule="evenodd" d="M3.1 9.16a3.38 3.38 0 0 1 3.38-3.39h7.74a3.38 3.38 0 0 1 0 6.77H6.48A3.38 3.38 0 0 1 3.1 9.16Z" clip-rule="evenodd"/><path fill="%2300CBEC" fill-rule="evenodd" d="M7 21.37a3.38 3.38 0 0 0-5.4 4.08l2.7-2.03-2.7 2.04.02.02a25.62 25.62 0 0 0 .35.43 22.37 22.37 0 0 0 4.22 3.7 23.2 23.2 0 0 0 13.1 3.97 23.2 23.2 0 0 0 16.45-6.71 17.38 17.38 0 0 0 1.24-1.4l.01-.01-2.7-2.04 2.7 2.03a3.38 3.38 0 1 0-5.51-3.94l-.54.58A16.43 16.43 0 0 1 19.3 26.8a16.43 16.43 0 0 1-11.64-4.7 10.66 10.66 0 0 1-.66-.73Zm24.58.02 2.62 1.97-2.62-1.97Z" clip-rule="evenodd"/></svg>',
    },
    'mdi-account-multiple-plus-gradient': {
        width: 41,
        height: 25,
        url: 'data:image/png;base64,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',
    },
    dashboard: {
        width: 30,
        height: 30,
        url: 'data:image/png;base64,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',
    },
}

export const PageHeaderIcon: React.FunctionComponent<PageHeaderIconProps> = ({ name, ...attributes }) => {
    const { theme } = useTheme()

    const { className, ...otherAttributes } = attributes
    if (!name || !Object.hasOwn(icons, name)) {
        return null
    }

    if (theme === Theme.Light) {
        return (
            <div className={classNames(styles.box, className)} {...otherAttributes}>
                <img
                    className={styles.boxContent}
                    src={icons[name].url}
                    alt={name}
                    width={icons[name].width}
                    height={icons[name].height}
                />
                <svg
                    className={styles.boxBackground}
                    width="92"
                    height="92"
                    viewBox="0 0 92 92"
                    fill="none"
                    xmlns="http://www.w3.org/2000/svg"
                >
                    <g filter="url(#a)">
                        <path
                            d="M16.6973 39.3422C16.6973 29.6443 16.6973 24.7954 18.4087 21.0182C20.3311 16.7753 23.7304 13.376 27.9733 11.4536C31.7505 9.74219 36.5994 9.74219 46.2973 9.74219C55.9951 9.74219 60.8441 9.74219 64.6212 11.4536C68.8641 13.376 72.2634 16.7753 74.1859 21.0182C75.8973 24.7954 75.8973 29.6443 75.8973 39.3422C75.8973 49.0401 75.8973 53.889 74.1859 57.6662C72.2634 61.909 68.8641 65.3084 64.6212 67.2308C60.8441 68.9422 55.9951 68.9422 46.2973 68.9422C36.5994 68.9422 31.7505 68.9422 27.9733 67.2308C23.7304 65.3084 20.3311 61.909 18.4087 57.6662C16.6973 53.889 16.6973 49.0401 16.6973 39.3422Z"
                            fill="white"
                        />
                        <path
                            d="M16.6973 39.3422C16.6973 29.6443 16.6973 24.7954 18.4087 21.0182C20.3311 16.7753 23.7304 13.376 27.9733 11.4536C31.7505 9.74219 36.5994 9.74219 46.2973 9.74219C55.9951 9.74219 60.8441 9.74219 64.6212 11.4536C68.8641 13.376 72.2634 16.7753 74.1859 21.0182C75.8973 24.7954 75.8973 29.6443 75.8973 39.3422C75.8973 49.0401 75.8973 53.889 74.1859 57.6662C72.2634 61.909 68.8641 65.3084 64.6212 67.2308C60.8441 68.9422 55.9951 68.9422 46.2973 68.9422C36.5994 68.9422 31.7505 68.9422 27.9733 67.2308C23.7304 65.3084 20.3311 61.909 18.4087 57.6662C16.6973 53.889 16.6973 49.0401 16.6973 39.3422Z"
                            fill="url(#paint0_radial_4521_4269)"
                            fillOpacity="0.2"
                        />
                        <path
                            d="M17.4973 39.3422C17.4973 34.4814 17.4978 30.8792 17.709 28.0189C17.9197 25.1666 18.3367 23.1155 19.1374 21.3484C20.9797 17.2823 24.2374 14.0246 28.3035 12.1823C30.0706 11.3816 32.1217 10.9646 34.9739 10.7539C37.8343 10.5427 41.4365 10.5422 46.2973 10.5422C51.1581 10.5422 54.7603 10.5427 57.6206 10.7539C60.4729 10.9646 62.5239 11.3816 64.2911 12.1823C68.3572 14.0246 71.6148 17.2823 73.4572 21.3484C74.2578 23.1155 74.6749 25.1666 74.8855 28.0189C75.0968 30.8792 75.0973 34.4814 75.0973 39.3422C75.0973 44.203 75.0968 47.8052 74.8855 50.6655C74.6749 53.5178 74.2578 55.5689 73.4572 57.336C71.6148 61.4021 68.3572 64.6598 64.2911 66.5021C62.5239 67.3028 60.4729 67.7198 57.6206 67.9304C54.7603 68.1417 51.1581 68.1422 46.2973 68.1422C41.4365 68.1422 37.8343 68.1417 34.9739 67.9304C32.1217 67.7198 30.0706 67.3028 28.3035 66.5021C24.2374 64.6598 20.9797 61.4021 19.1374 57.336C18.3367 55.5689 17.9197 53.5178 17.709 50.6655C17.4978 47.8052 17.4973 44.203 17.4973 39.3422Z"
                            stroke="black"
                            strokeOpacity="0.05"
                            strokeWidth="1.6"
                        />
                    </g>
                    <defs>
                        <filter
                            id="a"
                            width="91.2"
                            height="91.1992"
                            x="0.697266"
                            y="0.142188"
                            colorInterpolationFilters="sRGB"
                            filterUnits="userSpaceOnUse"
                        >
                            <feFlood floodOpacity="0" result="BackgroundImageFix" />
                            <feColorMatrix
                                in="SourceAlpha"
                                type="matrix"
                                values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0"
                                result="hardAlpha"
                            />
                            <feOffset dy="6.4" />
                            <feGaussianBlur stdDeviation="8" />
                            <feComposite in2="hardAlpha" operator="out" />
                            <feColorMatrix
                                type="matrix"
                                values="0 0 0 0 0.891257 0 0 0 0 0.907635 0 0 0 0 0.956771 0 0 0 1 0"
                            />
                            <feBlend mode="normal" in2="BackgroundImageFix" result="effect1_dropShadow_4521_4269" />
                            <feColorMatrix
                                in="SourceAlpha"
                                type="matrix"
                                values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0"
                                result="hardAlpha"
                            />
                            <feOffset dy="3.2" />
                            <feGaussianBlur stdDeviation="1.6" />
                            <feComposite in2="hardAlpha" operator="out" />
                            <feColorMatrix type="matrix" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0.05 0" />
                            <feBlend
                                mode="normal"
                                in2="effect1_dropShadow_4521_4269"
                                result="effect2_dropShadow_4521_4269"
                            />
                            <feBlend
                                mode="normal"
                                in="SourceGraphic"
                                in2="effect2_dropShadow_4521_4269"
                                result="shape"
                            />
                        </filter>
                    </defs>
                </svg>
            </div>
        )
    }

    return (
        <div className={classNames(styles.box, className)} {...otherAttributes}>
            <img
                className={styles.boxContent}
                src={icons[name].url}
                alt={name}
                width={icons[name].width}
                height={icons[name].height}
            />
            <svg
                className={styles.boxBackground}
                width="92"
                height="92"
                viewBox="0 0 92 92"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
            >
                <g filter="url(#a)">
                    <path
                        fill="#1d212f"
                        d="M16.5 39.94c0-9.7 0-14.55 1.71-18.32a19.2 19.2 0 0 1 9.57-9.57c3.77-1.71 8.62-1.71 18.32-1.71 9.7 0 14.55 0 18.32 1.71A19.2 19.2 0 0 1 74 21.62c1.71 3.77 1.71 8.62 1.71 18.32 0 9.7 0 14.55-1.71 18.32a19.2 19.2 0 0 1-9.57 9.57c-3.77 1.71-8.62 1.71-18.32 1.71-9.7 0-14.55 0-18.32-1.71a19.2 19.2 0 0 1-9.57-9.57c-1.71-3.77-1.71-8.62-1.71-18.32Z"
                    />
                    <path
                        stroke="#262B38"
                        strokeWidth="1.6"
                        d="M17.3 39.94c0-4.86 0-8.46.21-11.32.21-2.86.63-4.9 1.43-6.67a18.4 18.4 0 0 1 9.17-9.17c1.76-.8 3.81-1.22 6.67-1.43 2.86-.21 6.46-.21 11.32-.21s8.46 0 11.32.21c2.86.21 4.9.63 6.67 1.43a18.4 18.4 0 0 1 9.17 9.17c.8 1.76 1.22 3.81 1.43 6.67.21 2.86.21 6.46.21 11.32s0 8.46-.21 11.32c-.21 2.86-.63 4.9-1.43 6.67a18.4 18.4 0 0 1-9.17 9.17c-1.76.8-3.81 1.22-6.67 1.43-2.86.2-6.46.21-11.32.21s-8.46 0-11.32-.21c-2.86-.21-4.9-.63-6.67-1.43a18.4 18.4 0 0 1-9.17-9.17c-.8-1.76-1.22-3.81-1.43-6.67-.21-2.86-.21-6.46-.21-11.32Z"
                    />
                </g>
                <defs>
                    <filter
                        id="a"
                        width="91.2"
                        height="91.2"
                        x=".5"
                        y=".74"
                        colorInterpolationFilters="sRGB"
                        filterUnits="userSpaceOnUse"
                    >
                        <feFlood floodOpacity="0" result="BackgroundImageFix" />
                        <feColorMatrix
                            in="SourceAlpha"
                            result="hardAlpha"
                            values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0"
                        />
                        <feOffset dy="6.4" />
                        <feGaussianBlur stdDeviation="8" />
                        <feComposite in2="hardAlpha" operator="out" />
                        <feColorMatrix values="0 0 0 0 0.0827049 0 0 0 0 0.113782 0 0 0 0 0.199244 0 0 0 1 0" />
                        <feBlend in2="BackgroundImageFix" result="effect1_dropShadow_4658_19614" />
                        <feBlend in="SourceGraphic" in2="effect1_dropShadow_4658_19614" result="shape" />
                    </filter>
                </defs>
            </svg>
        </div>
    )
}
