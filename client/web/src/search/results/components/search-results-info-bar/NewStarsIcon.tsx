import React from 'react'

export const NewStarsIcon: React.FunctionComponent<{ width?: number; height?: number }> = ({
    width = 21,
    height = 21,
}) => (
    <svg xmlns="http://www.w3.org/2000/svg" width={width} height={height} fill="none" viewBox="0 0 21 21">
        <path
            fill="url(#paint0_linear_295_6722)"
            fillRule="evenodd"
            d="M10.455 7.954l-1.478 4.975L7.5 7.954 2.511 6.48l4.988-1.475L8.977.03l1.478 4.975 4.989 1.475-4.989 1.474zm9.908 5.151l-3.953 2.168-3.952-2.168 2.174 3.942-2.174 3.942 3.952-2.168 3.953 2.168-2.174-3.942 2.174-3.942zm-14.299-1.5l-1.537 2.706 1.537 2.706-2.713-1.533-2.713 1.533 1.537-2.706-1.537-2.706 2.713 1.534 2.713-1.534z"
            clipRule="evenodd"
        />
        <defs>
            <linearGradient
                id="paint0_linear_295_6722"
                x1="-4.089"
                x2="17.167"
                y1="25.826"
                y2="27.998"
                gradientUnits="userSpaceOnUse"
            >
                <stop stopColor="#7048E8" />
                <stop offset="0.035" stopColor="#607CE8" />
                <stop offset="0.185" stopColor="#4AC1E8" />
                <stop offset="0.35" stopColor="#59A3EC" />
                <stop offset="0.626" stopColor="#A112FF" />
                <stop offset="1" stopColor="#FF5543" />
            </linearGradient>
        </defs>
    </svg>
)
