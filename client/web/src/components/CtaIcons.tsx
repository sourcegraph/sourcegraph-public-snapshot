import * as React from 'react'

export const BookmarkRadialGradientIcon = React.memo(() => (
    <svg width="26" height="26" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path
            d="M20.143 21.667L13 18.517l-7.143 3.15V2.889h14.286v18.778zm0-21.667H5.857C5.1 0 4.373.304 3.837.846A2.905 2.905 0 003 2.89V26l10-4.333L23 26V2.889c0-.766-.301-1.501-.837-2.043A2.842 2.842 0 0020.143 0z"
            fill="#6B47D6"
        />
        <path
            d="M20.143 21.667L13 18.517l-7.143 3.15V2.889h14.286v18.778zm0-21.667H5.857C5.1 0 4.373.304 3.837.846A2.905 2.905 0 003 2.89V26l10-4.333L23 26V2.889c0-.766-.301-1.501-.837-2.043A2.842 2.842 0 0020.143 0z"
            fill="url(#paint0_radial)"
            fillOpacity=".59"
        />
        <defs>
            <radialGradient
                id="paint0_radial"
                cx="0"
                cy="0"
                r="1"
                gradientUnits="userSpaceOnUse"
                gradientTransform="matrix(18.00004 -1 1.4154 25.47731 8.5 12)"
            >
                <stop stopColor="#E105CB" />
                <stop offset="1" stopColor="#fff" stopOpacity="0" />
            </radialGradient>
        </defs>
    </svg>
))

export const SearchBetaIcon = React.memo(() => (
    <svg width="77" height="60" fill="none" xmlns="http://www.w3.org/2000/svg">
        <rect x="4.25" width="60" height="60" rx="4" fill="url(#paint0_linear)" />
        <rect x="4.25" width="60" height="60" rx="4" fill="url(#paint1_radial)" fillOpacity=".4" />
        <path
            d="M29.621 12a13.372 13.372 0 0113.372 13.371c0 3.312-1.214 6.357-3.21 8.702l.556.556h1.625L52.25 44.914 49.164 48 38.88 37.714V36.09l-.556-.555a13.404 13.404 0 01-8.702 3.209 13.371 13.371 0 010-26.743zm0 4.114a9.219 9.219 0 00-9.257 9.257 9.219 9.219 0 009.257 9.258 9.219 9.219 0 009.258-9.258 9.219 9.219 0 00-9.258-9.257z"
            fill="#fff"
        />
        <g filter="url(#filter0_d)">
            <rect x="44.75" y="11" width="28" height="14" rx="1.613" fill="#66D9E8" />
            <path
                d="M51.038 21c1.274 0 2.051-.645 2.051-1.696 0-.777-.554-1.356-1.356-1.43v-.075c.6-.095 1.062-.65 1.062-1.278 0-.918-.682-1.489-1.815-1.489h-2.494V21h2.552zm-1.485-5.136h1.166c.645 0 1.022.31 1.022.843 0 .546-.401.84-1.158.84h-1.03v-1.683zm0 4.305v-1.865h1.2c.814 0 1.244.318 1.244.926 0 .612-.417.939-1.203.939h-1.24zm8.56-.066h-2.795v-1.708h2.643v-.848h-2.643V15.93h2.796v-.898H54.25V21h3.863v-.897zM61.99 21v-5.07h1.84v-.898h-4.743v.898h1.836V21h1.067zm6.145 0h1.146l-2.122-5.968h-1.2L63.842 21h1.08l.513-1.543h2.196L68.134 21zm-1.633-4.847h.07l.81 2.482h-1.7l.82-2.482z"
                fill="#2B3750"
            />
        </g>
        <defs>
            <radialGradient
                id="paint1_radial"
                cx="0"
                cy="0"
                r="1"
                gradientUnits="userSpaceOnUse"
                gradientTransform="rotate(-35.166 71.162 -22.034) scale(28.8342)"
            >
                <stop stopColor="#DC06FF" />
                <stop offset="1" stopColor="#DB00FF" stopOpacity="0" />
            </radialGradient>
            <linearGradient
                id="paint0_linear"
                x1="4.364"
                y1="43.846"
                x2="64.469"
                y2="43.596"
                gradientUnits="userSpaceOnUse"
            >
                <stop stopColor="#4B52F3" />
                <stop offset="1" stopColor="#8E35F4" />
            </linearGradient>
            <filter
                id="filter0_d"
                x="42.519"
                y="9.885"
                width="32.462"
                height="18.462"
                filterUnits="userSpaceOnUse"
                colorInterpolationFilters="sRGB"
            >
                <feFlood floodOpacity="0" result="BackgroundImageFix" />
                <feColorMatrix in="SourceAlpha" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0" result="hardAlpha" />
                <feOffset dy="1.115" />
                <feGaussianBlur stdDeviation="1.115" />
                <feColorMatrix values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0.35 0" />
                <feBlend in2="BackgroundImageFix" result="effect1_dropShadow" />
                <feBlend in="SourceGraphic" in2="effect1_dropShadow" result="shape" />
            </filter>
        </defs>
    </svg>
))

export const CodeMonitorRadialGradientIcon = React.memo(() => (
    <svg width="26" height="26" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path
            fillRule="evenodd"
            clipRule="evenodd"
            d="M24.612 9.73c0 .39.305.695.694.695A.687.687 0 0026 9.73C26 4.373 21.641 0 16.27 0a.687.687 0 00-.695.694c0 .389.306.694.694.694 4.595 0 8.343 3.748 8.343 8.343zm-2.54-.069a.687.687 0 01-.694-.694 4.177 4.177 0 00-4.165-4.164.687.687 0 01-.694-.694c0-.389.306-.694.694-.694a5.55 5.55 0 015.553 5.552.687.687 0 01-.694.694zM6.33 8.051l11.8 11.8a8.47 8.47 0 01-4.429 1.22 8.547 8.547 0 01-6.08-2.512 8.537 8.537 0 01-2.526-6.08c0-1.582.43-3.11 1.235-4.428zm-.18-3.123c-4.164 4.164-4.164 10.938 0 15.103a10.646 10.646 0 007.551 3.123c2.735 0 5.47-1.041 7.552-3.123L6.15 4.928zm-.986 16.088c-.416-.416-.79-.86-1.138-1.332l-.402.707L1 24.946H11.536l-.389-.667a12.018 12.018 0 01-5.983-3.262zm13.673-11.23a2.423 2.423 0 01-2.429 2.43 2.423 2.423 0 01-2.43-2.43 2.423 2.423 0 012.43-2.43 2.423 2.423 0 012.43 2.43z"
            fill="#6B47D6"
        />
        <path
            fillRule="evenodd"
            clipRule="evenodd"
            d="M24.612 9.73c0 .39.305.695.694.695A.687.687 0 0026 9.73C26 4.373 21.641 0 16.27 0a.687.687 0 00-.695.694c0 .389.306.694.694.694 4.595 0 8.343 3.748 8.343 8.343zm-2.54-.069a.687.687 0 01-.694-.694 4.177 4.177 0 00-4.165-4.164.687.687 0 01-.694-.694c0-.389.306-.694.694-.694a5.55 5.55 0 015.553 5.552.687.687 0 01-.694.694zM6.33 8.051l11.8 11.8a8.47 8.47 0 01-4.429 1.22 8.547 8.547 0 01-6.08-2.512 8.537 8.537 0 01-2.526-6.08c0-1.582.43-3.11 1.235-4.428zm-.18-3.123c-4.164 4.164-4.164 10.938 0 15.103a10.646 10.646 0 007.551 3.123c2.735 0 5.47-1.041 7.552-3.123L6.15 4.928zm-.986 16.088c-.416-.416-.79-.86-1.138-1.332l-.402.707L1 24.946H11.536l-.389-.667a12.018 12.018 0 01-5.983-3.262zm13.673-11.23a2.423 2.423 0 01-2.429 2.43 2.423 2.423 0 01-2.43-2.43 2.423 2.423 0 012.43-2.43 2.423 2.423 0 012.43 2.43z"
            fill="url(#paint0_radial)"
            fillOpacity=".59"
        />
        <defs>
            <radialGradient
                id="paint0_radial"
                cx="0"
                cy="0"
                r="1"
                gradientUnits="userSpaceOnUse"
                gradientTransform="matrix(19.5 -2 2.17271 21.18397 11.5 14)"
            >
                <stop stopColor="#E105CB" />
                <stop offset="1" stopColor="#fff" stopOpacity="0" />
            </radialGradient>
        </defs>
    </svg>
))
