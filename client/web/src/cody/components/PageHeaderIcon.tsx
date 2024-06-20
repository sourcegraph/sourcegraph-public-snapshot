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
        url: 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 31 30"><g clip-path="url(%23a)"><path fill="url(%23b)" d="M16.85 13.75a3.75 3.75 0 1 0 0-7.5 3.75 3.75 0 0 0 0 7.5Zm0-5a1.25 1.25 0 1 1 0 2.5 1.25 1.25 0 0 1 0-2.5ZM22 13.57a6.25 6.25 0 0 0 0-7.15 3.75 3.75 0 1 1 0 7.15Zm-5.14 2.68c-7.5 0-7.5 5-7.5 5v2.5h15v-2.5s0-5-7.5-5Zm-5 5c0-.36.4-2.5 5-2.5 4.38 0 4.93 1.95 5 2.5m8.75 0v2.5h-3.75v-2.5a7 7 0 0 0-2.25-4.93c6 .62 6 4.93 6 4.93ZM10.6 15H6.85v3.75h-2.5V15H.6v-2.5h3.75V8.75h2.5v3.75h3.75V15Z"/></g><defs><radialGradient id="b" cx="0" cy="0" r="1" gradientTransform="matrix(30.38443 8.36927 -7.74253 28.10907 -3.17 12.02)" gradientUnits="userSpaceOnUse"><stop offset="0" stop-color="%237048E8"/><stop offset=".31" stop-color="%2300CBEC"/><stop offset=".64" stop-color="%23A112FF"/><stop offset="1" stop-color="%23FF5543"/></radialGradient><clipPath id="a"><path fill="%23fff" d="M0 0h30v30H0z" transform="translate(.6)"/></clipPath></defs></svg>',
    },
    dashboard: {
        width: 31,
        height: 31,
        url: 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 31 31"><path fill="url(%23a)" d="M.2.93h13.5v10.5H.2V.93Zm3 3v4.5h7.5v-4.5H3.2Zm13.5-3h13.5v16.5H16.7V.93Zm3 3v10.5h7.5V3.93h-7.5ZM.2 14.43h13.5v16.5H.2v-16.5Zm3 3v10.5h7.5v-10.5H3.2Zm13.5 3h13.5v10.5H16.7v-10.5Zm3 3v4.5h7.5v-4.5h-7.5Z"/><defs><radialGradient id="a" cx="0" cy="0" r="1" gradientTransform="matrix(39.53437 22.83585 -23.11115 40.01098 -11.9 9.2)" gradientUnits="userSpaceOnUse"><stop offset="0" stop-color="%237048E8"/><stop offset=".31" stop-color="%2300CBEC"/><stop offset=".64" stop-color="%23A112FF"/><stop offset="1" stop-color="%23FF5543"/></radialGradient></defs></svg>',
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
