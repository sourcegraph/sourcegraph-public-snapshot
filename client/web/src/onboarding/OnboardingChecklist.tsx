import { type FC, type PropsWithChildren, useCallback, useState } from 'react'

import { mdiAlertCircle, mdiChevronDown, mdiCheckCircle, mdiCheckCircleOutline } from '@mdi/js'
// eslint-disable-next-line id-length
import cx from 'classnames'

import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import {
    Button,
    Form,
    H2,
    H3,
    H4,
    Input,
    Label,
    LoadingSpinner,
    Modal,
    Text,
    Popover,
    PopoverContent,
    PopoverTrigger,
    Position,
    Icon,
    Link,
} from '@sourcegraph/wildcard'

import { BrandLogo } from '../components/branding/BrandLogo'

import { content } from './OnboardingChecklist.content'
import { useOnboardingChecklist, type OnboardingChecklistItem } from './useOnboardingChecklist'

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

    const { licenseKey, checklistItem } = data

    const isComplete = Object.values(checklistItem).every((value: boolean) => value)
    if (isComplete) {
        return <></>
    }

    return (
        <>
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
                                key={id}
                                isComplete={checklistItem[id as keyof OnboardingChecklistItem] || isComplete}
                                title={title}
                                description={description}
                                link={link}
                            />
                        ))}
                    </OnboardingChecklistList>
                </PopoverContent>
            </Popover>
            {!licenseKey && <LicenseKeyModal licenseKey={licenseKey} />}
        </>
    )
}

const OnboardingChecklistList: FC<PropsWithChildren<{}>> = ({ children }): JSX.Element => (
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

interface LicenseKeyProps {
    licenseKey: string
}
const LicenseKeyModal: FC<LicenseKeyProps> = ({ licenseKey }): JSX.Element => {
    const [isOpen, setIsOpen] = useState(true)
    const [key, setKey] = useState(licenseKey)

    return (
        <>
            {isOpen && (
                <Modal
                    className={styles.modal}
                    containerClassName={styles.modalContainer}
                    onDismiss={() => setIsOpen(false)}
                    aria-labelledby={'license-key'}
                >
                    <H3 className="m-0 pb-4">Upgrade your license</H3>
                    <Text className="m-0 pb-3">Enter your license key to start your enterprise set up:</Text>
                    <Form onSubmit={() => console.log('did it')}>
                        <Label htmlFor="license-key">License key</Label>
                        <Input
                            type="text"
                            name="license-key"
                            value={key}
                            className="pb-4"
                            onChange={({ target: { value } }) => setKey(value)}
                        />
                        <LicenseKey licenseKey={licenseKey} />
                        <div className="d-flex justify-content-end">
                            <Button
                                className="mr-2"
                                onClick={() => setIsOpen(false)}
                                outline={true}
                                variant="secondary"
                            >
                                Skip for now
                            </Button>
                            <Button type="submit" disabled={licenseKey === ''} variant="primary">
                                Upgrade and start set up
                            </Button>
                        </div>
                    </Form>
                </Modal>
            )}
        </>
    )
}

const LicenseKey: FC<LicenseKeyProps> = ({ licenseKey }): JSX.Element => {
    const isLightTheme = useIsLightTheme()
    return (
        <div className={cx(styles.keyWrapper, 'p-4 mb-4')}>
            <div className={styles.keyContainer}>
                <div className={styles.logoContainer}>
                    {licenseKey ? (
                        <BrandLogo isLightTheme={isLightTheme} variant="symbol" />
                    ) : (
                        <div className={styles.emptyLogo}></div>
                    )}
                </div>
                <div>
                    {licenseKey ? (
                        <>
                            <Text className="mb-1">License</Text>
                            <H2 className="mb-1">Sourcegraph Enterprise (dev-only)</H2>
                            <Text className="mb-0">1000-user license, valid until 2029-01-20 (5 years remaining)</Text>
                        </>
                    ) : (
                        <>
                            <Text className="mb-1">Current License</Text>
                            <H2 className="mb-1">Free</H2>
                            <Text className="mb-0">1-user license, valid indefinitely</Text>
                        </>
                    )}
                </div>
            </div>
        </div>
    )
}
