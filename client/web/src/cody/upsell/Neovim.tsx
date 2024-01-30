import React, { type FC } from 'react'

interface NeovimIconProps {
    className: string
}

export const NeovimIcon: FC<NeovimIconProps> = ({ className }) => (
    <svg
        className={className}
        width="34"
        height="34"
        viewBox="0 0 34 34"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
    >
        <g clip-path="url(#clip0_611_2720)">
            <path
                fill-rule="evenodd"
                clip-rule="evenodd"
                d="M0.280143 7.34523L7.45998 0.0926514V33.8147L0.280143 26.6459V7.34523Z"
                fill="url(#paint0_linear_611_2720)"
            />
            <path
                fill-rule="evenodd"
                clip-rule="evenodd"
                d="M28.0239 7.41064L20.7543 0.0926514L20.9016 33.8147L28.0731 26.6451L28.0239 7.41064Z"
                fill="url(#paint1_linear_611_2720)"
            />
            <path
                fill-rule="evenodd"
                clip-rule="evenodd"
                d="M7.45934 0.138977L26.1276 28.6355L20.9036 33.861L2.22565 5.4276L7.45934 0.138977Z"
                fill="url(#paint2_linear_611_2720)"
            />
            <path
                fill-rule="evenodd"
                clip-rule="evenodd"
                d="M7.45998 13.2865L7.4501 14.406L1.71611 5.91574L2.24705 5.37329L7.45998 13.2865Z"
                fill="black"
                fill-opacity="0.13"
            />
        </g>
        <defs>
            <linearGradient
                id="paint0_linear_611_2720"
                x1="359.272"
                y1="0.0926514"
                x2="359.272"
                y2="3372.3"
                gradientUnits="userSpaceOnUse"
            >
                <stop stop-color="#16B0ED" stop-opacity="0.800236" />
                <stop offset="1" stop-color="#0F59B2" stop-opacity="0.837" />
            </linearGradient>
            <linearGradient
                id="paint1_linear_611_2720"
                x1="-337.867"
                y1="0.0926514"
                x2="-337.867"
                y2="3372.3"
                gradientUnits="userSpaceOnUse"
            >
                <stop stop-color="#7DB643" />
                <stop offset="1" stop-color="#367533" />
            </linearGradient>
            <linearGradient
                id="paint2_linear_611_2720"
                x1="1197.32"
                y1="0.138977"
                x2="1197.32"
                y2="3372.35"
                gradientUnits="userSpaceOnUse"
            >
                <stop stop-color="#88C649" stop-opacity="0.8" />
                <stop offset="1" stop-color="#439240" stop-opacity="0.84" />
            </linearGradient>
            <clipPath id="clip0_611_2720">
                <rect width="27.8856" height="34" fill="white" transform="translate(0.1875)" />
            </clipPath>
        </defs>
    </svg>
)
