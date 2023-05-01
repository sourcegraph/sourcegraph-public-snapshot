import React from 'react'

export const CodyColoredSvg = React.memo<{ className?: string }>(function CodySvg({ className }) {
    return (
        <svg
            version="1.0"
            xmlns="http://www.w3.org/2000/svg"
            width="30"
            height="30"
            viewBox="0 0 128 128"
            className={className}
        >
            <g transform="translate(0,128) scale(0.100000,-0.100000)">
                <path
                    fill="#FF5543"
                    d="M832 1126 c-52 -28 -62 -61 -62 -199 0 -150 11 -186 67 -212 48 -23
    99 -14 138 25 l30 30 3 135 c4 151 -8 194 -59 220 -34 18 -86 19 -117 1z"
                />
                <path
                    fill="#A112FF"
                    d="M219 967 c-45 -30 -63 -83 -46 -134 7 -23 25 -46 46 -60 31 -21 45
    -23 163 -23 175 0 218 24 218 120 0 96 -43 120 -218 120 -118 0 -132 -2 -163
    -23z"
                />
                <path
                    d="M977 503 c-40 -38 -96 -80 -123 -95 -185 -100 -409 -68 -569 83 -56
    52 -82 63 -124 54 -50 -11 -81 -51 -81 -104 0 -44 3 -49 73 -116 137 -132 305
    -192 511 -182 186 9 308 62 446 193 83 80 85 83 85 129 0 40 -5 51 -33 76 -24
    22 -42 29 -72 29 -36 0 -48 -8 -113 -67z"
                    fill="#00CBEC"
                />
            </g>
        </svg>
    )
})

export const CodySvg = React.memo<{ className?: string }>(function CodySvg({ className }) {
    return (
        <svg
            version="1.0"
            xmlns="http://www.w3.org/2000/svg"
            width="30"
            height="30"
            viewBox="0 0 128 128"
            className={className}
        >
            <g transform="translate(0,128) scale(0.100000,-0.100000)" fill="currentColor">
                <path
                    d="M832 1126 c-52 -28 -62 -61 -62 -199 0 -150 11 -186 67 -212 48 -23
        99 -14 138 25 l30 30 3 135 c4 151 -8 194 -59 220 -34 18 -86 19 -117 1z"
                />
                <path
                    d="M219 967 c-45 -30 -63 -83 -46 -134 7 -23 25 -46 46 -60 31 -21 45
        -23 163 -23 175 0 218 24 218 120 0 96 -43 120 -218 120 -118 0 -132 -2 -163
        -23z"
                />
                <path
                    d="M977 503 c-40 -38 -96 -80 -123 -95 -185 -100 -409 -68 -569 83 -56
        52 -82 63 -124 54 -50 -11 -81 -51 -81 -104 0 -44 3 -49 73 -116 137 -132 305
        -192 511 -182 186 9 308 62 446 193 83 80 85 83 85 129 0 40 -5 51 -33 76 -24
        22 -42 29 -72 29 -36 0 -48 -8 -113 -67z"
                />
            </g>
        </svg>
    )
})

export const ResetSvg = React.memo(() => (
    <div className="header-container-right">
        <div className="reset-conversation" title="Start a new conversation with Cody">
            <ResetIcon />
        </div>
    </div>
))

export const ResetIcon = React.memo(() => (
    <svg
        xmlns="http://www.w3.org/2000/svg"
        width="12"
        height="12"
        viewBox="0 0 24 24"
        fill="var(--vscode-sideBarTitle-foreground)"
    >
        <path d="M12 16c1.671 0 3-1.331 3-3s-1.329-3-3-3-3 1.331-3 3 1.329 3 3 3z" />
        <path d="M20.817 11.186a8.94 8.94 0 0 0-1.355-3.219 9.053 9.053 0 0 0-2.43-2.43 8.95 8.95 0 0 0-3.219-1.355 9.028 9.028 0 0 0-1.838-.18V2L8 5l3.975 3V6.002c.484-.002.968.044 1.435.14a6.961 6.961 0 0 1 2.502 1.053 7.005 7.005 0 0 1 1.892 1.892A6.967 6.967 0 0 1 19 13a7.032 7.032 0 0 1-.55 2.725 7.11 7.11 0 0 1-.644 1.188 7.2 7.2 0 0 1-.858 1.039 7.028 7.028 0 0 1-3.536 1.907 7.13 7.13 0 0 1-2.822 0 6.961 6.961 0 0 1-2.503-1.054 7.002 7.002 0 0 1-1.89-1.89A6.996 6.996 0 0 1 5 13H3a9.02 9.02 0 0 0 1.539 5.034 9.096 9.096 0 0 0 2.428 2.428A8.95 8.95 0 0 0 12 22a9.09 9.09 0 0 0 1.814-.183 9.014 9.014 0 0 0 3.218-1.355 8.886 8.886 0 0 0 1.331-1.099 9.228 9.228 0 0 0 1.1-1.332A8.952 8.952 0 0 0 21 13a9.09 9.09 0 0 0-.183-1.814z" />
    </svg>
))

export const SubmitSvg = React.memo(() => (
    <svg
        xmlns="http://www.w3.org/2000/svg"
        width="24"
        height="24"
        viewBox="0 0 24 24"
        fill="var(--vscode-sideBarTitle-foreground)"
    >
        <path d="m21.426 11.095-17-8A1 1 0 0 0 3.03 4.242l1.212 4.849L12 12l-7.758 2.909-1.212 4.849a.998.998 0 0 0 1.396 1.147l17-8a1 1 0 0 0 0-1.81z" />
    </svg>
))
