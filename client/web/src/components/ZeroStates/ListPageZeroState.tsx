import type { FC, ReactNode } from 'react'

import classNames from 'classnames'

import { Text, H3 } from '@sourcegraph/wildcard'

import styles from './ListPageZeroState.module.scss'

interface AppNoItemsStateProps {
    title: ReactNode
    subTitle?: ReactNode
    withIllustration?: boolean
    size?: 'small' | 'large'
    className?: string
}

export const ListPageZeroState: FC<AppNoItemsStateProps> = props => {
    const { title, subTitle, size = 'large', withIllustration = false, className } = props
    const sizeInPixels = size === 'large' ? 144 : 102

    return (
        <div className={classNames(styles.zeroState, className)}>
            {withIllustration && (
                <svg
                    width={sizeInPixels}
                    height={sizeInPixels}
                    viewBox="0 0 144 144"
                    fill="none"
                    xmlns="http://www.w3.org/2000/svg"
                >
                    <g clipPath="url(#clip0_317_910)">
                        <path
                            fill="var(--light-part)"
                            fillRule="evenodd"
                            clipRule="evenodd"
                            d="M112 10C113.309 9.97382 114.555 9.43554 115.471 8.50071C116.388 7.56587 116.901 6.30903 116.901 5C116.901 3.69097 116.388 2.43413 115.471 1.49929C114.555 0.564463 113.309 0.0261766 112 0C110.691 0.0261766 109.445 0.564463 108.529 1.49929C107.612 2.43413 107.099 3.69097 107.099 5C107.099 6.30903 107.612 7.56587 108.529 8.50071C109.445 9.43554 110.691 9.97382 112 10ZM8 5L41 21L24 30L8 5ZM99.5 110H44.5L48.7 84H95.305L99.5 110ZM118 36C118 42.075 113.075 47 107 47C105.375 47.0048 103.77 46.6485 102.3 45.9567C100.83 45.2649 99.5318 44.255 98.5 43L116 29.5C117.26 31.29 118 33.645 118 36Z"
                        />
                        <path
                            fill="var(--middle-part)"
                            fillRule="evenodd"
                            clipRule="evenodd"
                            d="M55 45H89L105 144H39L55 45ZM99.5 110H44.5L48.7 84H95.305L99.5 110Z"
                        />
                        <path
                            fill="var(--dark-part)"
                            fillRule="evenodd"
                            clipRule="evenodd"
                            d="M71 0H73V5.667L87 15V30H95L89 45H55L49 30H57V15L71 5.667V0ZM64 22C64 17.582 67.584 14 72 14C76.427 14 80 17.582 80 22C80 26.418 76.427 30 72 30C67.584 30 64 26.418 64 22ZM85 58V65H78V58C78 56.5 79.5 55 81.5 55C83.5 55 85 56.5 85 58ZM66 135V144H78V135C78 131.779 75.22 129 72 129C68.78 129 66 131.779 66 135ZM62 101.002V93C62 91.5 60.5 90 58.5 90C56.5 90 55 91.5 55 93V101.002H62Z"
                        />
                    </g>
                    <defs>
                        <clipPath id="clip0_317_910">
                            <rect width="144" height="144" fill="white" />
                        </clipPath>
                    </defs>
                </svg>
            )}

            <span className={styles.zeroStateText}>
                <H3>{title}</H3>

                {subTitle && <Text className={styles.zeroStateSubText}>{subTitle}</Text>}
            </span>
        </div>
    )
}
