import React from 'react'
import type { ReactElement } from 'react'

import { Icon } from '@sourcegraph/wildcard'

export const codyIconPath =
    'm9 15a1 1 0 01-1-1v-2a1 1 0 012 0v2a1 1 0 01-1 1zm6 0a1 1 0 01-1-1v-2a1 1 0 012 0v2a1 1 0 01-1 1zm-9-7a1 1 0 01-.71-.29l-3-3a1 1 0 011.42-1.42l3 3a1 1 0 010 1.42 1 1 0 01-.71.29zm12 0a1 1 0 01-.71-.29 1 1 0 010-1.42l3-3a1 1 0 111.42 1.42l-3 3a1 1 0 01-.71.29zm3 12h-18a1 1 0 01-1-1v-4.5a10 10 0 0120 0v4.5a1 1 0 01-1 1zm-17-2h16v-3.5a8 8 0 00-16 0z'

export const CodyIcon: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <Icon svgPath={codyIconPath} className={className} aria-hidden={true} />
)

export const AutocompletesIcon = (): ReactElement => (
    <svg width="33" height="34" viewBox="0 0 33 34" fill="none" xmlns="http://www.w3.org/2000/svg">
        <rect width="33" height="34" rx="16.5" fill="#6B47D6" />
        <rect width="33" height="34" rx="16.5" fill="url(#paint0_linear_2692_3962)" />
        <path
            d="M18.0723 24.8147L14.9142 21.6566L15.9658 20.5943L18.0723 22.7008L22.4826 18.2799L23.5343 19.3421L18.0723 24.8147ZM9.5166 20.1419L13.331 10.2329H14.924L18.7277 20.1419H17.1161L16.1305 17.5438H11.9834L11.0084 20.1419H9.5166ZM12.3829 16.2867H15.7334L14.1079 11.7981H14.006L12.3829 16.2867Z"
            fill="white"
        />
        <defs>
            <linearGradient
                id="paint0_linear_2692_3962"
                x1="16.5"
                y1="0"
                x2="16.5"
                y2="34"
                gradientUnits="userSpaceOnUse"
            >
                <stop stopColor="#FF3424" />
                <stop offset="1" stopColor="#CF275B" />
            </linearGradient>
        </defs>
    </svg>
)

export const ChatMessagesIcon = (): ReactElement => (
    <svg width="34" height="34" viewBox="0 0 34 34" fill="none" xmlns="http://www.w3.org/2000/svg">
        <rect x="0.5" width="33" height="34" rx="16.5" fill="#6B47D6" />
        <rect x="0.5" width="33" height="34" rx="16.5" fill="url(#paint0_linear_2692_3970)" />
        <path
            d="M12.4559 18.5188H18.4046V17.3938H12.4559V18.5188ZM12.4559 16.0267H21.544V14.9017H12.4559V16.0267ZM12.4559 13.5533H21.544V12.4283H12.4559V13.5533ZM9.14697 24.8832V10.6683C9.14697 10.2466 9.3022 9.87948 9.61265 9.56695C9.92311 9.25441 10.2877 9.09814 10.7065 9.09814H23.2934C23.7151 9.09814 24.0822 9.25441 24.3948 9.56695C24.7073 9.87948 24.8635 10.2466 24.8635 10.6683V20.2495C24.8635 20.6683 24.7073 21.0329 24.3948 21.3433C24.0822 21.6538 23.7151 21.809 23.2934 21.809H12.2211L9.14697 24.8832ZM11.7035 20.2495H23.2934V10.6683H10.7065V21.359L11.7035 20.2495Z"
            fill="white"
        />
        <defs>
            <linearGradient id="paint0_linear_2692_3970" x1="17" y1="0" x2="17" y2="34" gradientUnits="userSpaceOnUse">
                <stop stopColor="#03C9ED" />
                <stop offset="1" stopColor="#536AEA" />
            </linearGradient>
        </defs>
    </svg>
)

export const ProIcon = ({ className }: { className?: string }): ReactElement => (
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 22 17" width={22} height={16} className={className}>
        <path
            fill="#fff"
            fillOpacity=".46"
            d="M16.3 11.04a.96.96 0 0 1-.64.22 1 1 0 0 1-.65-.22 1.4 1.4 0 0 1-.4-.6 2.63 2.63 0 0 1-.13-.85c0-.32.05-.61.13-.86.1-.26.23-.45.4-.6.18-.15.4-.22.65-.22.26 0 .47.07.64.22.17.15.3.34.39.6.09.25.13.54.13.86 0 .32-.04.6-.13.86-.09.25-.22.45-.39.6zM5.83 8.7H6.9c.3 0 .55-.05.74-.15a1 1 0 0 0 .45-.44c.1-.19.14-.4.14-.64 0-.25-.05-.46-.14-.64a.99.99 0 0 0-.45-.43c-.2-.1-.44-.16-.75-.16H5.83V8.7z"
        />
        <path
            fill="#fff"
            fillOpacity=".46"
            fillRule="evenodd"
            d="M3.76.87A3.74 3.74 0 0 0 0 4.61v8.53a3.74 3.74 0 0 0 3.76 3.73h14.48A3.74 3.74 0 0 0 22 13.14V4.61A3.74 3.74 0 0 0 18.24.87H3.76zM14.2 12.1c.4.24.9.36 1.45.36s1.05-.12 1.46-.36c.41-.24.73-.57.95-1 .22-.43.33-.93.33-1.5a3.2 3.2 0 0 0-.33-1.49 2.4 2.4 0 0 0-.95-1c-.4-.24-.9-.36-1.46-.36s-1.04.12-1.45.36a2.4 2.4 0 0 0-.8.73V6.8a1.77 1.77 0 0 0-.48-.06 1.35 1.35 0 0 0-1.33 1.05h-.06V6.8h-1.5v5.53h1.55V9.2c0-.22.05-.42.15-.6.1-.17.24-.3.42-.4.18-.1.38-.15.6-.15a2.71 2.71 0 0 1 .5.05 3.2 3.2 0 0 0-.33 1.49c0 .56.1 1.06.33 1.5.22.42.54.75.95 1zM4.27 4.97v7.37h1.56V9.95h1.32c.57 0 1.06-.1 1.46-.31.4-.21.7-.5.92-.88s.32-.8.32-1.3c0-.48-.1-.91-.32-1.29-.2-.38-.5-.67-.9-.88a3 3 0 0 0-1.44-.32H4.27z"
            clipRule="evenodd"
        />
    </svg>
)
