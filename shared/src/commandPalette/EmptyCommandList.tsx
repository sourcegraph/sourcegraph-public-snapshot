import React from 'react'
import { SettingsCascadeOrError } from '../settings/settings'
import { onlyDefaultExtensionsAdded } from '../extensions/extensions'

interface Props {
    settingsCascade?: SettingsCascadeOrError
    sourcegraphURL: string
}

export const EmptyCommandList: React.FunctionComponent<Props> = ({ settingsCascade, sourcegraphURL }) => {
    // if no settings cascade, default to 'no active extensions'
    const onlyDefault = settingsCascade ? onlyDefaultExtensionsAdded(settingsCascade) : false

    return (
        <div className="empty-command-list">
            <p className="empty-command-list__title">
                {onlyDefault ? "You don't have any extensions enabled" : "You don't have any active actions"}
            </p>
            <p className="empty-command-list__text">
                {onlyDefault
                    ? 'Find extensions in the Sourcegraph extension registry, or learn how to write your own in just a few minutes.'
                    : 'Commands from your installed extensions will be shown when you navigate to certain pages.'}
            </p>

            <a className="btn btn-primary" href={sourcegraphURL + '/extensions'}>
                Explore extensions
            </a>

            <PuzzleIllustration className="empty-command-list__illustration" />
        </div>
    )
}

const PuzzleIllustration = React.memo<{ className?: string }>(({ className }) => (
    <svg
        width="64"
        height="70"
        viewBox="0 0 64 70"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        className={className}
    >
        <path
            d="M55.779 48.0498C59.9364 49.8553 61.9091 44.4627 62.5757 44.4627C63.2423 44.4627 63.7756 44.9965 63.7756 45.6638V66.0451C63.7756 68.2288 62.0061 70 59.8245 70H41.0144C40.869 69.9757 40.7599 69.9515 40.6993 69.9272C38.8692 69.0659 43.5052 66.8039 43.4627 64.3345C43.4052 61.0027 41.1599 59.7821 36.2998 59.7821C31.4397 59.7821 28.8508 60.6016 30.184 65.2723C30.6375 66.8613 33.5243 69.1144 31.4882 69.9272C31.4155 69.9515 31.2337 69.9757 30.9913 70H12.7388C10.5814 70 8.83616 68.2652 8.7998 66.1057V47.2773C8.82404 47.1075 8.84828 46.9861 8.88464 46.9255C9.74516 45.0936 12.7388 49.0849 14.4719 49.6915C19.0532 51.2808 23.2709 47.1802 23.2709 42.3154C23.2709 37.4506 20.4955 30.5719 15.9142 32.1733C14.181 32.7799 9.73304 38.1022 8.921 36.0641C8.88464 35.9913 8.8604 35.785 8.83616 35.506L8.7998 18.5251C8.7998 18.5251 8.84828 17.8943 8.90888 17.591C9.26036 15.3102 11.6237 14.364 13.6478 14.364L15.9142 14.4125C15.9142 14.4125 21.4893 14.5459 26.0222 14.5581H27.9007C29.9005 14.5581 31.4276 14.5095 31.6943 14.4003C33.7304 13.5875 29.5369 10.5425 28.9309 8.80763C27.3432 4.22184 31.4519 0 36.2998 0C41.1478 0 45.2564 4.22184 43.6808 8.79549C43.0748 10.5303 39.0874 13.539 40.9175 14.3882C41.0872 14.4731 41.5477 14.5217 42.2264 14.5581H50.5043C53.8009 14.4853 56.6733 14.3882 56.6733 14.3882L58.9397 14.3397C59.7154 14.3397 60.5517 14.5095 61.3152 14.8371C61.3395 14.8492 61.3758 14.8614 61.4122 14.8735C61.4849 14.9099 61.5455 14.9463 61.6182 14.9827C61.7879 15.0797 61.9697 15.1768 62.1273 15.286C62.2727 15.3951 62.4181 15.5165 62.5515 15.6378C62.5878 15.6742 62.6242 15.6984 62.6605 15.7348C63.1817 16.2444 63.5695 16.8995 63.715 17.7123C63.7756 17.9671 63.7998 18.234 63.7998 18.513V32.1733C63.7877 32.8406 63.1747 35.6214 62.654 36.0374C59.7894 38.3254 58.6436 29.6463 54.6331 33.7493C51.4667 36.9888 51.6216 46.2443 55.779 48.0498Z"
            fill="url(#paint0_linear)"
            fillOpacity="0.32"
        />
        <rect y="27" width="14" height="3" rx="1.5" fill="#ABC7EB" fillOpacity="0.5" />
        <rect y="51" width="14" height="3" rx="1.5" fill="#ABC7EB" fillOpacity="0.5" />
        <rect y="57" width="14" height="3" rx="1.5" fill="#ABC7EB" fillOpacity="0.5" />
        <rect y="21" width="12" height="3" rx="1.5" fill="#ABC7EB" fillOpacity="0.5" />
        <rect x="12" y="33" width="14" height="3" rx="1.5" fill="#ABC7EB" fillOpacity="0.5" />
        <rect x="12" y="45" width="14" height="3" rx="1.5" fill="#ABC7EB" fillOpacity="0.5" />
        <rect x="12" y="39" width="17" height="3" rx="1.5" fill="#ABC7EB" fillOpacity="0.5" />
        <rect x="15" y="27" width="10" height="3" rx="1.5" fill="#ABC7EB" fillOpacity="0.5" />
        <rect x="15" y="51" width="10" height="3" rx="1.5" fill="#ABC7EB" fillOpacity="0.5" />
        <rect x="15" y="57" width="6" height="3" rx="1.5" fill="#ABC7EB" fillOpacity="0.5" />
        <rect x="13" y="21" width="6" height="3" rx="1.5" fill="#ABC7EB" fillOpacity="0.5" />
        <rect x="27" y="33" width="5" height="3" rx="1.5" fill="#ABC7EB" fillOpacity="0.5" />
        <rect x="27" y="45" width="10" height="3" rx="1.5" fill="#ABC7EB" fillOpacity="0.5" />
        <rect x="33" y="33" width="8" height="3" rx="1.5" fill="#ABC7EB" fillOpacity="0.5" />
        <rect x="38" y="45" width="5" height="3" rx="1.5" fill="#ABC7EB" fillOpacity="0.5" />
        <rect x="30" y="39" width="8" height="3" rx="1.5" fill="#ABC7EB" fillOpacity="0.5" />
        <defs>
            <linearGradient id="paint0_linear" x1="36.2998" y1="0" x2="36.2998" y2="70" gradientUnits="userSpaceOnUse">
                <stop stopColor="#95A5C6" stopOpacity="0.6" />
                <stop offset="1" stopColor="#95A5C6" />
            </linearGradient>
        </defs>
    </svg>
))
