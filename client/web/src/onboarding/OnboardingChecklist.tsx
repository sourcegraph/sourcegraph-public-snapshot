import { type FC, type PropsWithChildren, useCallback, useState } from 'react'

import { mdiAlertCircle, mdiChevronDown, mdiCheckCircle, mdiCheckCircleOutline } from '@mdi/js'
// eslint-disable-next-line id-length
import cx from 'classnames'

import {
    H4,
    LoadingSpinner,
    Text,
    Popover,
    PopoverContent,
    PopoverTrigger,
    Position,
    Icon,
    Link,
} from '@sourcegraph/wildcard'

import { content } from './OnboardingChecklist.content'
import { useOnboardingChecklist, type OnboardingChecklistResult } from './useOnboardingChecklist'

import styles from './OnboardingChecklist.module.scss'

export const OnboardingChecklist: FC = (): JSX.Element => {
    const { data, loading, error } = useOnboardingChecklist()
    const [isDropdownOpen, setIsDropdownOpen] = useState(false)
    const toggleDropdownOpen = useCallback(() => setIsDropdownOpen(isOpen => !isOpen), [])

    if (loading || !data) {
        return <LoadingSpinner />
    }
    if (error) {
        return <></>
    }

    const isComplete = Object.values(data).every((value: boolean) => value)
    if (isComplete) {
        return <></>
    }

    return (
        <Popover isOpen={isDropdownOpen} onOpenChange={toggleDropdownOpen}>
            <PopoverTrigger type="button" className={styles.button}>
                <Icon aria-hidden={true} size="md" svgPath={mdiAlertCircle} />
                <span>Setup</span>
                <Icon aria-hidden={true} svgPath={mdiChevronDown} />
            </PopoverTrigger>
            <PopoverContent className={styles.container} position={Position.bottom}>
                <OnboardingChecklistList>
                    {content.map(({ id, isComplete, title, description, link }, index) => (
                        <OnboardingChecklistItem
                            // eslint-disable-next-line react/no-array-index-key
                            key={index}
                            isComplete={data[id as keyof OnboardingChecklistResult] || isComplete}
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
