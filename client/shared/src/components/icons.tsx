import * as React from 'react'

import classNames from 'classnames'

export interface IconProps {
    className?: string
    size?: number
    'data-tooltip'?: string
}

function sizeProps(props: IconProps): { width: number; height: number; viewBox: string } {
    const defaultSize = 24
    const size = props.size || defaultSize
    return {
        width: size,
        height: size,
        viewBox: `0 0 ${size} ${size}`,
    }
}

export const ChatIcon: React.FunctionComponent<React.PropsWithChildren<IconProps>> = props => (
    <svg {...props} {...sizeProps(props)} className={classNames('mdi-icon', props.className)}>
        <path d="M 2 11.636 A 10 8 0 0 0 4.75 17.146 A 9 9 0 0 1 2 21.636 A 10.4 10.4 0 0 0 8.5 19.13 A 10 8 0 0 0 12 19.636 A 10 8 0 0 0 22 11.636 A 10 8 0 0 0 12 3.636 A 10 8 0 0 0 2 11.636 Z" />
    </svg>
)

export const CircleChevronLeftIcon: React.FunctionComponent<React.PropsWithChildren<IconProps>> = props => (
    <svg {...props} {...sizeProps(props)} className={classNames('mdi-icon', props.className)}>
        <path d="M22,12c0,5.5-4.5,10-10,10S2,17.5,2,12S6.5,2,12,2S22,6.5,22,12z M15.4,16.6L10.8,12l4.6-4.6L14,6l-6,6l6,6L15.4,16.6z" />
    </svg>
)

export const CircleChevronRightIcon: React.FunctionComponent<React.PropsWithChildren<IconProps>> = props => (
    <svg {...props} {...sizeProps(props)} className={classNames('mdi-icon', props.className)}>
        <path d="M22,12c0,5.5-4.5,10-10,10S2,17.5,2,12S6.5,2,12,2S22,6.5,22,12z M10,18l6-6l-6-6L8.6,7.4l4.6,4.6l-4.6,4.6L10,18z" />
    </svg>
)

export const RepoQuestionIcon: React.FunctionComponent<React.PropsWithChildren<IconProps>> = props => (
    <svg {...props} {...sizeProps(props)} className={classNames('mdi-icon', props.className)} viewBox="0 0 64 64">
        <title>Icons 400</title>
        <g>
            <path
                d="M50,16h-1.5c-0.3,0-0.5,0.2-0.5,0.5v35c0,0.3-0.2,0.5-0.5,0.5h-27c-0.5,0-1-0.2-1.4-0.6l-0.6-0.6c-0.1-0.1-0.1-0.2-0.1-0.4
		c0-0.3,0.2-0.5,0.5-0.5H44c1.1,0,2-0.9,2-2V12c0-1.1-0.9-2-2-2H14c-1.1,0-2,0.9-2,2v36.3c0,1.1,0.4,2.1,1.2,2.8l3.1,3.1
		c1.1,1.1,2.7,1.8,4.2,1.8H50c1.1,0,2-0.9,2-2V18C52,16.9,51.1,16,50,16z M29,44c-1.7,0-3-1.3-3-3c0-1.7,1.3-3,3-3v0
		c1.7,0,3,1.3,3,3S30.7,44,29,44z M29,16c4.4,0,8,3.1,8,7c0,6.4-5.8,6.2-5.8,11l0,0c0,0.6-0.4,1-1,1H28c-0.6,0-1-0.4-1-1
		c0-6.6,5.6-6,5.6-11c0-1.7-1.9-3-3.8-3c-1.9,0-3.8,1.3-3.8,3c0,0.6,0.1,1.1,0.3,1.7c0.2,0.5-0.1,1.1-0.6,1.3
		c-0.1,0-0.2,0.1-0.3,0.1h-2c-0.5,0-0.8-0.3-1-0.7C21.1,24.5,21,23.8,21,23C21,19.1,24.6,16,29,16z"
            />
        </g>
    </svg>
)

export const FormatListBulletedIcon: React.FunctionComponent<React.PropsWithChildren<IconProps>> = props => (
    <svg {...props} {...sizeProps(props)} className={classNames('mdi-icon', props.className)}>
        <path d="M7,5H21V7H7V5M7,13V11H21V13H7M4,4.5A1.5,1.5 0 0,1 5.5,6A1.5,1.5 0 0,1 4,7.5A1.5,1.5 0 0,1 2.5,6A1.5,1.5 0 0,1 4,4.5M4,10.5A1.5,1.5 0 0,1 5.5,12A1.5,1.5 0 0,1 4,13.5A1.5,1.5 0 0,1 2.5,12A1.5,1.5 0 0,1 4,10.5M7,19V17H21V19H7M4,16.5A1.5,1.5 0 0,1 5.5,18A1.5,1.5 0 0,1 4,19.5A1.5,1.5 0 0,1 2.5,18A1.5,1.5 0 0,1 4,16.5Z" />
    </svg>
)

export const PhabricatorIcon: React.FunctionComponent<React.PropsWithChildren<IconProps>> = props => (
    <svg
        {...props}
        {...sizeProps(props)}
        className={classNames('phabricator-icon mdi-icon', props.className)}
        viewBox="0 0 64 64"
        fill="currentColor"
    >
        <g>
            <g id="Oval">
                <g transform="translate(-3426.45 1547.34)">
                    <path
                        id="path15_fill"
                        d="M32,36.4c2.3,0,4.2-1.9,4.2-4.2S34.3,28,32,28c-2.3,0-4.2,1.9-4.2,4.2S29.7,36.4,32,36.4z"
                    />
                </g>
            </g>
            <path
                d="M58.8,31.2L57.6,32v0L58.8,31.2C54.8,25,45,14.6,32,14.6C19,14.6,9.2,25,5.2,31.2L6.4,32v0l-1.3-0.8L4.6,32l0.5,0.8
		C9.2,39,19,49.4,32,49.4c13,0,22.8-10.4,26.8-16.6l0.5-0.8L58.8,31.2z M32,46.4c-10.9,0-19.6-8.5-23.7-14.4
		c4.2-5.9,12.9-14.4,23.7-14.4S51.6,26.1,55.7,32C51.6,37.9,42.9,46.4,32,46.4z"
            />
            <path
                d="M44.4,33.3v-2.2L42.7,31c-0.1-0.6-0.2-1.2-0.4-1.8l1.5-0.7l-0.8-2l-1.6,0.6c-0.3-0.5-0.6-1-1-1.5l1.2-1.3L40,22.6l-1.3,1.2
		c-0.5-0.4-1-0.7-1.5-1l0.6-1.7l-2-0.8L35,21.9c-0.6-0.2-1.2-0.3-1.8-0.3l-0.1-1.8h-2.2l-0.1,1.8c-0.6,0.1-1.2,0.2-1.8,0.3l-0.8-1.6
		l-2,0.8l0.6,1.7c-0.5,0.3-1,0.6-1.5,1L24,22.6l-1.6,1.6l1.2,1.3c-0.4,0.5-0.7,1-1,1.5l-1.6-0.6l-0.8,2l1.5,0.7
		c-0.2,0.6-0.3,1.2-0.4,1.8l-1.7,0.1v2.2l1.7,0.1c0.1,0.6,0.2,1.2,0.4,1.8l-1.5,0.7l0.8,2l1.6-0.6c0.3,0.5,0.6,1,1,1.5l-1.1,1.3
		l1.6,1.6l1.3-1.2c0.5,0.4,1,0.7,1.5,1l-0.6,1.7l2,0.8l0.8-1.6c0.6,0.2,1.2,0.3,1.8,0.3l0.1,1.8h2.2l0.1-1.8
		c0.6-0.1,1.2-0.2,1.8-0.3l0.8,1.6l2-0.8l-0.6-1.7c0.5-0.3,1-0.6,1.5-1l1.3,1.2l1.6-1.6l-1.2-1.3c0.4-0.5,0.7-1,1-1.5l1.6,0.6l0.8-2
		l-1.5-0.8c0.2-0.6,0.3-1.2,0.4-1.8L44.4,33.3z M38.7,32.2c0,3.7-3,6.6-6.7,6.6s-6.7-3-6.7-6.6c0-3.7,3-6.6,6.7-6.6
		S38.7,28.5,38.7,32.2z"
            />
        </g>
    </svg>
)

export const WrapDisabledIcon: React.FunctionComponent<React.PropsWithChildren<IconProps>> = props => (
    <svg {...props} {...sizeProps(props)} className={classNames('mdi-icon', props.className)}>
        <path d="M16,7H3V5H16ZM3,19H16V17H3Zm19-7L18,9v2H3v2H18v2Z" />
    </svg>
)

// TODO: Rename name when refresh design is complete
export const CloudAlertIconRefresh: React.FunctionComponent<React.PropsWithChildren<IconProps>> = props => (
    <svg
        {...props}
        {...sizeProps(props)}
        className={classNames('phabricator-icon mdi-icon', props.className)}
        viewBox="0 0 20 20"
        fill="currentColor"
    >
        <g>
            <path
                d="M15.3129 6.28624C14.2107 4.51968 12.2477 3.33203 9.96484 3.33203C7.54818 3.33203 5.46484 4.66536 4.46484 6.66536C1.88151 6.9987 -0.0351562 9.08203 -0.0351562 11.6654C-0.0351562 14.4154 2.21484 16.6654 4.96484 16.6654H9.32894L10.3379 14.9154H4.96484C3.18134 14.9154 1.71484 13.4489 1.71484 11.6654C1.71484 10.0076 2.9321 8.62765 4.68879 8.40098L5.61324 8.28169L6.03009 7.44799C6.7313 6.04556 8.20576 5.08203 9.96484 5.08203C11.9829 5.08203 13.6486 6.35936 14.2597 8.11301L15.3129 6.28624Z"
                fill="#798BAF"
            />
            <path
                d="M15.4749 9.62828C15.639 9.34396 16.0494 9.34396 16.2136 9.62828L19.9071 16.0256C20.0712 16.3099 19.866 16.6653 19.5377 16.6653H12.1508C11.8224 16.6653 11.6173 16.3099 11.7814 16.0256L15.4749 9.62828Z"
                fill="#F59F00"
            />
        </g>
        <defs>
            <clipPath id="clip0">
                <rect width="20" height="20" fill="white" />
            </clipPath>
        </defs>
    </svg>
)

// TODO: Rename name when refresh design is complete
export const CloudSyncIconRefresh: React.FunctionComponent<React.PropsWithChildren<IconProps>> = props => (
    <svg
        {...props}
        {...sizeProps(props)}
        className={classNames('phabricator-icon mdi-icon', props.className)}
        viewBox="0 0 20 20"
        fill="currentColor"
    >
        <g>
            <path
                d="M13.7873 7.14774C12.9858 5.90993 11.5895 5.08203 9.96484 5.08203C8.20576 5.08203 6.7313 6.04556 6.03009 7.44799L5.61324 8.28169L4.68879 8.40098C2.9321 8.62765 1.71484 10.0076 1.71484 11.6654C1.71484 13.4489 3.18134 14.9154 4.96484 14.9154H10.3872V16.6654H4.96484C2.21484 16.6654 -0.0351562 14.4154 -0.0351562 11.6654C-0.0351562 9.08203 1.88151 6.9987 4.46484 6.66536C5.46484 4.66536 7.54818 3.33203 9.96484 3.33203C12.0727 3.33203 13.9079 4.34461 15.0445 5.89054L13.7873 7.14774Z"
                fill="#798BAF"
            />
            <path
                d="M12.6934 15.2025C11.353 13.8621 11.2692 11.684 12.4421 10.176L13.6149 11.4327C13.0285 12.1866 13.1961 13.3595 13.8663 14.0297C14.2851 14.3648 14.7878 14.6161 15.3742 14.6161V13.1081L17.7199 15.5376L15.3742 17.8833V16.2916C14.3689 16.2916 13.3636 15.8727 12.6934 15.2025Z"
                fill="#5E6E8C"
            />
            <path
                d="M13.8663 9.58962L16.2119 7.16016V8.75187C17.2172 8.75187 18.1388 9.17075 18.8927 9.75717C20.2331 11.0976 20.3169 13.2757 19.144 14.7836L17.9712 13.6108C18.5576 12.8568 18.3901 11.684 17.7199 11.0138C17.301 10.6787 16.7984 10.4274 16.2119 10.4274V11.9353L13.8663 9.58962Z"
                fill="#5E6E8C"
            />
        </g>
        <defs>
            <clipPath id="clip0">
                <rect width="20" height="20" fill="white" />
            </clipPath>
        </defs>
    </svg>
)

// TODO: Rename name when refresh design is complete
export const CloudCheckIconRefresh: React.FunctionComponent<React.PropsWithChildren<IconProps>> = props => (
    <svg
        {...props}
        {...sizeProps(props)}
        className={classNames('phabricator-icon mdi-icon', props.className)}
        viewBox="0 -4 20 20"
        fill="currentColor"
    >
        <g>
            <path
                d="M14.4175 5.6859L14.6515 6.82244L16.1402 5.33371L16.1315 5.33301C15.5482 2.49967 13.0482 0.333008 9.96484 0.333008C7.54818 0.333008 5.46484 1.66634 4.46484 3.66634C1.88151 3.99967 -0.0351562 6.08301 -0.0351562 8.66634C-0.0351562 11.4163 2.21484 13.6663 4.96484 13.6663H10.273L8.52301 11.9163H4.96484C3.18134 11.9163 1.71484 10.4498 1.71484 8.66634C1.71484 7.00858 2.9321 5.62862 4.68879 5.40195L5.61324 5.28267L6.03009 4.44896C6.7313 3.04654 8.20576 2.08301 9.96484 2.08301C12.1965 2.08301 13.9973 3.64505 14.4175 5.6859Z"
                fill="#798BAF"
            />
            <path
                d="M9.74232 8.18541L8.50488 9.42285L12.7475 13.6655L19.8287 6.59486L18.5913 5.35742L12.7528 11.1959L9.74232 8.18541Z"
                fill="#37B24D"
            />
        </g>
        <defs>
            <clipPath id="clip0">
                <rect width="20" height="20" fill="white" />
            </clipPath>
        </defs>
    </svg>
)
