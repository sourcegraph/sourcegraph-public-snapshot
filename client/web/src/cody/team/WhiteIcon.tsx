import React from 'react'

import styles from './WhiteIcon.module.scss'

export const ICON_NAMES = ['mdi-account-multiple-plus-gradient'] as const

interface WhiteIconProps {
    name: typeof ICON_NAMES[number]
}

const nameToUrl = {
    'mdi-account-multiple-plus-gradient':
        'data:image/png;base64,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',
}

export const WhiteIcon: React.FunctionComponent<WhiteIconProps> = ({ name, ...attributes }) => {
    if (!name || !ICON_NAMES.includes(name)) {
        return null
    }

    return (
        <div className={styles.whiteBox} {...attributes}>
            <img className={styles.whiteBoxContent} src={nameToUrl[name]} alt={name} />
            <svg
                className={styles.whiteBoxBackground}
                width="92"
                height="92"
                viewBox="0 0 92 92"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
            >
                <g filter="url(#filter0_dd_4521_4269)">
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
                        id="filter0_dd_4521_4269"
                        x="0.697266"
                        y="0.142188"
                        width="91.2"
                        height="91.1992"
                        filterUnits="userSpaceOnUse"
                        colorInterpolationFilters="sRGB"
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
                        <feBlend mode="normal" in="SourceGraphic" in2="effect2_dropShadow_4521_4269" result="shape" />
                    </filter>
                </defs>
            </svg>
        </div>
    )
}
