import React, { SVGProps } from 'react'

export const CodyHighQualityIcon: React.FunctionComponent<React.PropsWithChildren<SVGProps<SVGSVGElement>>> = (
    props: SVGProps<SVGSVGElement>
) => (
    <svg xmlns="http://www.w3.org/2000/svg" width="19" height="19" fill="none" viewBox="0 0 19 19" {...props}>
        <path
            fill="url(#a)"
            d="M16 10.09V4c0-2.21-3.58-4-8-4S0 1.79 0 4v10c0 2.21 3.59 4 8 4 .46 0 .9 0 1.33-.06A5.94 5.94 0 0 1 9 16v-.05c-.32.05-.65.05-1 .05-3.87 0-6-1.5-6-2v-2.23C3.61 12.55 5.72 13 8 13c.65 0 1.27-.04 1.88-.11A5.986 5.986 0 0 1 15 10c.34 0 .67.04 1 .09Zm-2-.64C12.7 10.4 10.42 11 8 11s-4.7-.6-6-1.55V6.64C3.47 7.47 5.61 8 8 8s4.53-.53 6-1.36v2.81ZM8 6C4.13 6 2 4.5 2 4s2.13-2 6-2 6 1.5 6 2-2.13 2-6 2Zm10.5 8.25L13.75 19 11 16l1.16-1.16 1.59 1.59 3.59-3.59 1.16 1.41Z"
        />
        <defs>
            <linearGradient id="a" x1="9.25" x2="9.25" y1="0" y2="19" gradientUnits="userSpaceOnUse">
                <stop stopColor="#A112FF" />
                <stop offset=".464" stopColor="#FF5543" />
                <stop offset="1" stopColor="#00CBEC" />
            </linearGradient>
        </defs>
    </svg>
)
