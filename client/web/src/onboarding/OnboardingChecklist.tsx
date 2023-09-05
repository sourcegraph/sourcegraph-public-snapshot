import { type FC, type PropsWithChildren, useCallback, useState } from 'react'

import { mdiAlertCircle, mdiChevronDown, mdiCheckCircle, mdiCheckCircleOutline } from '@mdi/js'
// eslint-disable-next-line id-length
import cx from 'classnames'

import { H4, Text, Popover, PopoverContent, PopoverTrigger, Position, Icon, Link } from '@sourcegraph/wildcard'

import { useOnboardingChecklist } from './useOnboardingChecklist'

import styles from './OnboardingChecklist.module.scss'

export const OnboardingChecklist: FC = (): JSX.Element => {
    const mockData = [
        {
            isComplete: true,
            title: 'Set license key',
            description: 'Please set your license key',
            link: '#',
        },
        {
            isComplete: false,
            title: 'Set external URL',
            description: 'Must be set in order for Sourcegraph to work correctly.',
            link: '#',
        },
        {
            isComplete: false,
            title: 'Set up SMTP',
            description: 'Must be set in order for Sourcegraph to send emails.',
            link: '#',
        },
        {
            isComplete: false,
            title: 'Connect a code host',
            description: 'You must connect a code host to set up user authentication and use Sourcegraph.',
            link: '#',
        },
        {
            isComplete: false,
            title: 'Set up user authentication',
            description: 'We recommend that enterprise instances use SSO or SAML to authenticate users.',
            link: '#',
        },
        {
            isComplete: false,
            title: 'Set user permissions',
            description:
                'We recommend limiting permissions based on repository permissions already set in your code host(s).',
            link: '#',
        },
    ]

    const { data, loading, error } = useOnboardingChecklist()
    const [isDropdownOpen, setIsDropdownOpen] = useState(false)
    const toggleDropdownOpen = useCallback(() => setIsDropdownOpen(isOpen => !isOpen), [])

    return (
        <Popover isOpen={isDropdownOpen} onOpenChange={toggleDropdownOpen}>
            <PopoverTrigger type="button" className={styles.button}>
                <Icon aria-hidden={true} size="md" svgPath={mdiAlertCircle} />
                <span>Setup</span>
                <Icon aria-hidden={true} svgPath={mdiChevronDown} />
            </PopoverTrigger>
            <PopoverContent className={styles.container} position={Position.bottom}>
                <OnboardingChecklistList>
                    {mockData.map(({ isComplete, title, description, link }, index) => (
                        <OnboardingChecklistItem
                            // eslint-disable-next-line react/no-array-index-key
                            key={index}
                            isComplete={isComplete}
                            title={title}
                            description={description}
                            link={link}
                        />
                    ))}
                </OnboardingChecklistList>
            </PopoverContent>
        </Popover>
    )
}

type OnboardingChecklistListProps = PropsWithChildren<{}>
const OnboardingChecklistList: FC<OnboardingChecklistListProps> = ({ children }): JSX.Element => (
    <ul className={styles.list}>{children}</ul>
)

interface OnboardingChecklistItemProps {
    isComplete: boolean
    title: string
    description: string
    link: string
}

const OnboardingChecklistItem: FC<OnboardingChecklistItemProps> = ({
    isComplete,
    title,
    description,
    link,
}): JSX.Element => (
    <li className={styles.item}>
        <div className={styles.wrapper}>
            <Icon
                aria-hidden={true}
                svgPath={isComplete ? mdiCheckCircle : mdiCheckCircleOutline}
                className={cx({ [styles.checked]: isComplete })}
            />
            <div className={styles.content}>
                <H4>{title}</H4>
                {!isComplete && (
                    <>
                        <Text>{description}</Text>
                        <Link to={link} target="_blank" rel="noopener">
                            Configure now
                        </Link>{' '}
                    </>
                )}
            </div>
        </div>
    </li>
)
