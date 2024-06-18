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

export const ProIcon = ({className}: { className?: string }): ReactElement => (
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 15 15" width={15} height={15} className={className}>
        <path fill="#fff" fillOpacity=".46" d="M10.9 9.11a.52.52 0 0 1-.4.2.53.53 0 0 1-.42-.2c-.1-.12-.2-.3-.25-.52a3.1 3.1 0 0 1-.08-.75c0-.28.03-.53.08-.75.06-.22.14-.4.25-.53a.55.55 0 0 1 .42-.19c.16 0 .3.07.4.2.11.12.2.3.25.52a3.02 3.02 0 0 1 0 1.5c-.05.22-.14.4-.25.52ZM4.24 7.06h.68a.8.8 0 0 0 .48-.13 1 1 0 0 0 .28-.38 1.66 1.66 0 0 0 0-1.13.78.78 0 0 0-.28-.37.8.8 0 0 0-.48-.14h-.68v2.15Z" />
        <path fill="#fff" fillOpacity=".46" fillRule="evenodd" d="M2.92.21C1.6.21.53 1.68.53 3.48v7.47c0 1.8 1.07 3.26 2.4 3.26h9.21c1.32 0 2.4-1.46 2.4-3.26V3.48c0-1.8-1.08-3.27-2.4-3.27H2.92Zm6.64 9.82c.27.2.57.31.93.31.36 0 .67-.1.93-.3a2 2 0 0 0 .6-.88c.14-.38.21-.82.21-1.31 0-.5-.07-.94-.2-1.31a1.97 1.97 0 0 0-.61-.88 1.43 1.43 0 0 0-.93-.31c-.36 0-.66.1-.93.31-.2.17-.37.38-.5.65v-.92a1.1 1.1 0 0 0-.3-.05.7.7 0 0 0-.53.24 1.4 1.4 0 0 0-.33.68h-.03V5.4h-.95v4.84h.98V7.5c0-.2.03-.37.1-.52a.86.86 0 0 1 .26-.35.62.62 0 0 1 .39-.13 1.27 1.27 0 0 1 .3.04c-.13.37-.2.8-.2 1.3s.07.93.21 1.3a2 2 0 0 0 .6.88ZM3.25 3.8v6.45h1v-2.1h.83A1.55 1.55 0 0 0 6.6 7.13c.13-.33.2-.7.2-1.14 0-.42-.07-.8-.2-1.13a1.72 1.72 0 0 0-.58-.77 1.5 1.5 0 0 0-.91-.28H3.25Z" clip-rule="evenodd" />
    </svg>
)
