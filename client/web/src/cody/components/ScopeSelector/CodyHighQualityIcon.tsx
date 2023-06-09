import React, { SVGProps } from 'react'

import { mdiDatabaseCheckOutline } from '@mdi/js'

export const CodyHighQualityIcon: React.FunctionComponent<React.PropsWithChildren<SVGProps<SVGSVGElement>>> = (
    props: SVGProps<SVGSVGElement>
) => (
    <svg width="24" height="24" viewBox="0 0 24 24" {...props}>
        <defs>
            <linearGradient
                id="cody-high-quality-icon-gradient"
                x1="12"
                x2="12"
                y1="0"
                y2="24"
                gradientUnits="userSpaceOnUse"
            >
                <stop stopColor="#A112FF" />
                <stop offset=".5" stopColor="#FF5543" />
                <stop offset="1" stopColor="#00CBEC" />
            </linearGradient>
        </defs>
        <path d={mdiDatabaseCheckOutline} fill="url(#cody-high-quality-icon-gradient)" />
    </svg>
)
