import React, { useCallback, useEffect, useState } from 'react'

import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { Button, H2, H5, Modal, Text, useSearchParameters } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import { HubSpotForm } from '../../marketing/components/HubSpotForm'
import { eventLogger } from '../../tracking/eventLogger'
import { EventName } from '../../util/constants'

import { JetBrainsInstructions } from './instructions/JetBrains'
import { VSCodeInstructions } from './instructions/VsCode'

import styles from './CodyOnboarding.module.scss'

export interface IEditor {
    icon: string
    name: string
    publisher: string
    releaseStage: string
    instructions?: React.FC<{ onBack?: () => void; onClose: () => void; showStep?: number }>
}

export const editorGroups: IEditor[][] = [
    [
        {
            icon: 'VsCode',
            name: 'VS Code',
            publisher: 'Microsoft',
            releaseStage: 'Stable',
            instructions: VSCodeInstructions,
        },
        {
            icon: 'IntelliJ',
            name: 'IntelliJ IDEA',
            publisher: 'JetBrains',
            releaseStage: 'Beta',
            instructions: JetBrainsInstructions,
        },
        {
            icon: 'PhpStorm',
            name: 'PhpStorm ',
            publisher: 'JetBrains',
            releaseStage: 'Beta',
        },
        {
            icon: 'PyCharm',
            name: 'PyCharm',
            publisher: 'Jetbrains',
            releaseStage: 'Beta',
        },
    ],
    [
        {
            icon: 'WebStorm',
            name: 'WebStorm',
            publisher: 'JetBrains',
            releaseStage: 'Beta',
        },
        {
            icon: 'RubyMine',
            name: 'RubyMine',
            publisher: 'JetBrains',
            releaseStage: 'Beta',
        },
        {
            icon: 'GoLand',
            name: 'GoLand',
            publisher: 'JetBrains',
            releaseStage: 'Beta',
        },
        {
            icon: 'AndroidStudio',
            name: 'Android Studio',
            publisher: 'Google',
            releaseStage: 'Beta',
        },
    ],
    [
        {
            icon: 'NeoVim',
            name: 'Neovim',
            publisher: 'Neovim Team',
            releaseStage: 'Experimental',
        },
        {
            icon: 'Emacs',
            name: 'Emacs',
            publisher: 'Free Software Foundation',
            releaseStage: 'Coming Soon',
        },
    ],
]

export function CodyOnboarding({
    authenticatedUser,
}: {
    authenticatedUser: AuthenticatedUser | null
}): JSX.Element | null {
    const [showEditorStep, setShowEditorStep] = useState(false)
    const [completed = false, setOnboardingCompleted] = useTemporarySetting('cody.onboarding.completed', false)
    // steps start from 0
    const [step = -1, setOnboardingStep] = useTemporarySetting('cody.onboarding.step', 0)

    const onNext = (): void => setOnboardingStep(currentsStep => (currentsStep || 0) + 1)

    const parameters = useSearchParameters()
    const enrollPro = parameters.get('pro') === 'true'
    const returnToURL = parameters.get('returnTo')

    const navigate = useNavigate()

    if (completed && !showEditorStep && returnToURL) {
        navigate(returnToURL)
    }

    if (!showEditorStep && (completed || step === -1 || step > 1)) {
        return null
    }

    if (!authenticatedUser) {
        return null
    }

    return (
        <Modal isOpen={true} aria-label="Cody Onboarding" className={styles.modal} position="center">
            {step === 0 && <WelcomeStep onNext={onNext} pro={enrollPro} />}
            {step === 1 && (
                <PurposeStep
                    authenticatedUser={authenticatedUser}
                    onNext={() => {
                        onNext()
                        setOnboardingCompleted(true)
                        setShowEditorStep(true)
                    }}
                    pro={enrollPro}
                />
            )}
            {showEditorStep && (
                <EditorStep
                    onCompleted={() => {
                        setShowEditorStep(false)
                    }}
                    pro={enrollPro}
                />
            )}
        </Modal>
    )
}

function WelcomeStep({ onNext, pro }: { onNext: () => void; pro: boolean }): JSX.Element {
    const [show, setShow] = useState(false)
    const isLightTheme = useIsLightTheme()
    useEffect(() => {
        eventLogger.log(EventName.CODY_ONBOARDING_WELCOME_VIEWED, { tier: pro ? 'pro' : 'free' })
    }, [pro])

    useEffect(() => {
        // theme is not ready on first render, it defaults to system theme.
        // so we need to wait a bit before showing the welcome video.
        setTimeout(() => {
            setShow(true)
        }, 500)
    }, [])

    return (
        <div className={classNames('d-flex flex-column align-items-center p-5')}>
            {show ? (
                <>
                    <video width="180" className={classNames('mb-5', styles.welcomeVideo)} autoPlay={true} muted={true}>
                        <source
                            src={
                                isLightTheme
                                    ? 'https://storage.googleapis.com/sourcegraph-assets/codyWelcomeAnim.mp4'
                                    : 'https://storage.googleapis.com/sourcegraph-assets/codyWelcomeAnim_dark.mp4'
                            }
                            type="video/mp4"
                        />
                        Your browser does not support the video tag.
                    </video>
                    <Text className={classNames('mb-4 pb-4', styles.fadeIn, styles.fadeSecond, styles.welcomeSubtitle)}>
                        Ready to breeze through the basics and get comfortable with Cody
                        {pro ? ' to Cody Pro Trial' : ''}?
                    </Text>
                    <Button
                        onClick={onNext}
                        variant="primary"
                        size="lg"
                        className={classNames(styles.fadeIn, styles.fadeThird)}
                    >
                        Sure, let's dive in!
                    </Button>
                </>
            ) : (
                <div className={styles.blankPlaceholder} />
            )}
        </div>
    )
}

function PurposeStep({
    onNext,
    pro,
    authenticatedUser,
}: {
    onNext: () => void
    pro: boolean
    authenticatedUser: AuthenticatedUser
}): JSX.Element {
    const [useCase, setUseCase] = useState<'work' | 'personal' | null>(null)

    useEffect(() => {
        eventLogger.log(EventName.CODY_ONBOARDING_PURPOSE_VIEWED, { tier: pro ? 'pro' : 'free' })
    }, [pro])

    const primaryEmail = authenticatedUser.emails.find(email => email.isPrimary)?.email

    const handleFormReady = useCallback((form: HTMLFormElement) => {
        const workInput = form.querySelector('input[name="using_cody_for_work"]')
        const personalInput = form.querySelector('input[name="using_cody_for_personal"]')

        const handleChange = (e: Event): void => {
            const target = e.target as HTMLInputElement
            const isChecked = target.checked
            const name = target.name

            if (name === 'using_cody_for_work' && isChecked) {
                setUseCase('work')
            } else if (name === 'using_cody_for_personal' && isChecked) {
                setUseCase('personal')
            } else {
                setUseCase(null)
            }
        }

        workInput?.addEventListener('change', handleChange)
        personalInput?.addEventListener('change', handleChange)

        return () => {
            workInput?.removeEventListener('change', handleChange)
            personalInput?.removeEventListener('change', handleChange)
        }
    }, [])

    return (
        <>
            <div className="border-bottom pb-3 mb-3">
                <H2 className="mb-1">What are you using Cody for?</H2>
                <Text className="mb-0 text-muted" size="small">
                    This will allow us to understand our audience better and guide your journey
                </Text>
            </div>
            <div className="d-flex align-items-center border-bottom mb-3 pb-3 justify-content-center">
                <HubSpotForm
                    formId="85548efc-a879-4553-9ef0-a8da8fdcf541"
                    onFormSubmitted={() => {
                        if (useCase) {
                            eventLogger.log(EventName.CODY_ONBOARDING_PURPOSE_SELECTED, { useCase })
                        }
                        onNext()
                    }}
                    userId={authenticatedUser.id}
                    userEmail={primaryEmail}
                    masterFormName="qualificationSurvey"
                    onFormReady={handleFormReady}
                />
                {/* TODO(naman): remove after PR feedback
                <div
                    role="button"
                    tabIndex={0}
                    onKeyDown={event => {
                        event.preventDefault()
                        eventLogger.log(EventName.CODY_ONBOARDING_PURPOSE_SELECTED, { useCase: 'work' })
                        onNext()
                    }}
                    className="border-right flex-1 d-flex flex-column justify-content-center cursor-pointer align-items-center py-3 px-2"
                    onClick={event => {
                        event.preventDefault()
                        eventLogger.log(EventName.CODY_ONBOARDING_PURPOSE_SELECTED, { useCase: 'work' })
                        onNext()
                    }}
                >
                    <WorkIcon />
                    <H3 className="mb-0 mt-2">Work</H3>
                </div>
                <div
                    role="button"
                    tabIndex={0}
                    className="flex-1 d-flex flex-column justify-content-center cursor-pointer align-items-center py-3 px-2"
                    onKeyDown={event => {
                        event.preventDefault()
                        eventLogger.log(EventName.CODY_ONBOARDING_PURPOSE_SELECTED, { useCase: 'personal' })
                        onNext()
                    }}
                    onClick={event => {
                        event.preventDefault()
                        eventLogger.log(EventName.CODY_ONBOARDING_PURPOSE_SELECTED, { useCase: 'personal' })
                        onNext()
                    }}
                >
                    <PersonalIcon />
                    <H3 className="mb-0 mt-2">Personal projects</H3>
                </div>
                */}
            </div>
            <Text size="small" className="text-muted text-center mb-0">
                Pick one to move forward
            </Text>
        </>
    )
}

function EditorStep({ onCompleted, pro }: { onCompleted: () => void; pro: boolean }): JSX.Element {
    useEffect(() => {
        eventLogger.log(EventName.CODY_ONBOARDING_CHOOSE_EDITOR_VIEWED, { tier: pro ? 'pro' : 'free' })
    }, [pro])

    const [editor, setEditor] = useState<null | IEditor>(null)

    const onBack = (): void => setEditor(null)

    if (editor?.instructions) {
        const Instructions = editor.instructions

        return <Instructions onBack={onBack} onClose={onCompleted} />
    }

    return (
        <>
            <div className="border-bottom pb-3 mb-3">
                <H2 className="mb-1">Choose your editor</H2>
                <Text className="mb-0 text-muted" size="small">
                    Most of Cody experience happens in the IDE. Let's get that set up.
                </Text>
            </div>
            <div className="mb-3 border-bottom pb-3">
                {editorGroups.map((group, index) => (
                    <div
                        key={index}
                        className={classNames('d-flex mt-3', styles.responsiveContainer, {
                            'border-bottom pb-3': index < group.length - 1,
                        })}
                    >
                        {group.map((editor, index) => (
                            <div
                                key={index}
                                className={classNames('d-flex flex-column flex-1 p-3 cursor-pointer', styles.ideGrid, {
                                    'border-left': index !== 0,
                                })}
                                role="button"
                                tabIndex={0}
                                onKeyDown={() => {
                                    setEditor(editor)

                                    eventLogger.log(EventName.CODY_ONBOARDING_CHOOSE_EDITOR_SELECTED, {
                                        tier: pro ? 'pro' : 'free',
                                        editor,
                                    })
                                }}
                                onClick={() => {
                                    eventLogger.log(EventName.CODY_ONBOARDING_CHOOSE_EDITOR_SELECTED, {
                                        tier: pro ? 'pro' : 'free',
                                        editor,
                                    })
                                    setEditor(editor)
                                }}
                            >
                                <div className="d-flex align-items-center">
                                    <div>
                                        <img
                                            alt={editor.name}
                                            src={`https://storage.googleapis.com/sourcegraph-assets/ideIcons/ideIcon${editor.icon}.svg`}
                                            width={34}
                                            className="mr-3"
                                        />
                                    </div>
                                    <div>
                                        <Text className="text-muted mb-0 text-truncate" size="small">
                                            {editor.publisher}
                                        </Text>
                                        <Text className={classNames('mb-0', styles.ideName)}>{editor.name}</Text>
                                        <H5 className={styles.releaseStage}>{editor.releaseStage}</H5>
                                    </div>
                                </div>
                            </div>
                        ))}
                        {group.length < 4
                            ? [...new Array(4 - group.length)].map((_, index) => (
                                  <div key={index} className="flex-1 p-3" />
                              ))
                            : null}
                    </div>
                ))}
            </div>
            <div className="d-flex justify-content-end align-items-center">
                <Text
                    className="mb-0 text-muted cursor-pointer"
                    size="small"
                    onClick={() => {
                        onCompleted()
                        eventLogger.log(EventName.CODY_ONBOARDING_CHOOSE_EDITOR_SKIPPED, { tier: pro ? 'pro' : 'free' })
                    }}
                >
                    Skip for now
                </Text>
            </div>
        </>
    )
}

/* TODO(naman): remove after PR feedback

const WorkIcon = (): JSX.Element => (
    <svg width="60" height="60" viewBox="0 0 75 75" fill="none">
        <path
            d="M14.8206 61.2414C13.6888 61.2414 12.7149 60.8335 11.899 60.0175C11.083 59.2016 10.675 58.222 10.675 57.0789V27.1706C10.675 26.0275 11.083 25.048 11.899 24.232C12.7149 23.4161 13.6945 23.0081 14.8375 23.0081H29.175V18.8439C29.175 17.7004 29.583 16.721 30.399 15.9058C31.2149 15.0907 32.1891 14.6831 33.3215 14.6831H41.6786C42.811 14.6831 43.7852 15.0911 44.6011 15.907C45.4171 16.723 45.8251 17.7025 45.8251 18.8456V23.0081H60.1626C61.3056 23.0081 62.2852 23.4161 63.1011 24.232C63.9171 25.048 64.3251 26.0275 64.3251 27.1706V57.0789C64.3251 58.222 63.9171 59.2016 63.1011 60.0175C62.2852 60.8335 61.3113 61.2414 60.1795 61.2414H14.8206ZM30.8709 23.0081H44.1292V18.8456C44.1292 18.2289 43.8723 17.6637 43.3584 17.1498C42.8445 16.6359 42.2792 16.3789 41.6625 16.3789H33.3375C32.7209 16.3789 32.1556 16.6359 31.6417 17.1498C31.1278 17.6637 30.8709 18.2289 30.8709 18.8456V23.0081ZM62.6292 44.5144H42.0865V48.2144H32.9907V44.5144H12.3709V57.0789C12.3709 57.6956 12.6278 58.2609 13.1417 58.7748C13.6556 59.2887 14.2209 59.5456 14.8375 59.5456H60.1626C60.7792 59.5456 61.3445 59.2887 61.8584 58.7748C62.3723 58.2609 62.6292 57.6956 62.6292 57.0789V44.5144ZM12.3709 42.8185H32.9907V39.1185H42.0865V42.8185H62.6292V27.1706C62.6292 26.5539 62.3723 25.9887 61.8584 25.4748C61.3445 24.9609 60.7792 24.7039 60.1626 24.7039H14.8375C14.2209 24.7039 13.6556 24.9609 13.1417 25.4748C12.6278 25.9887 12.3709 26.5539 12.3709 27.1706V42.8185Z"
            fill="#EFF2F5"
        />
        <path
            d="M14.8206 61.2414C13.6888 61.2414 12.7149 60.8335 11.899 60.0175C11.083 59.2016 10.675 58.222 10.675 57.0789V27.1706C10.675 26.0275 11.083 25.048 11.899 24.232C12.7149 23.4161 13.6945 23.0081 14.8375 23.0081H29.175V18.8439C29.175 17.7004 29.583 16.721 30.399 15.9058C31.2149 15.0907 32.1891 14.6831 33.3215 14.6831H41.6786C42.811 14.6831 43.7852 15.0911 44.6011 15.907C45.4171 16.723 45.8251 17.7025 45.8251 18.8456V23.0081H60.1626C61.3056 23.0081 62.2852 23.4161 63.1011 24.232C63.9171 25.048 64.3251 26.0275 64.3251 27.1706V57.0789C64.3251 58.222 63.9171 59.2016 63.1011 60.0175C62.2852 60.8335 61.3113 61.2414 60.1795 61.2414H14.8206ZM30.8709 23.0081H44.1292V18.8456C44.1292 18.2289 43.8723 17.6637 43.3584 17.1498C42.8445 16.6359 42.2792 16.3789 41.6625 16.3789H33.3375C32.7209 16.3789 32.1556 16.6359 31.6417 17.1498C31.1278 17.6637 30.8709 18.2289 30.8709 18.8456V23.0081ZM62.6292 44.5144H42.0865V48.2144H32.9907V44.5144H12.3709V57.0789C12.3709 57.6956 12.6278 58.2609 13.1417 58.7748C13.6556 59.2887 14.2209 59.5456 14.8375 59.5456H60.1626C60.7792 59.5456 61.3445 59.2887 61.8584 58.7748C62.3723 58.2609 62.6292 57.6956 62.6292 57.0789V44.5144ZM12.3709 42.8185H32.9907V39.1185H42.0865V42.8185H62.6292V27.1706C62.6292 26.5539 62.3723 25.9887 61.8584 25.4748C61.3445 24.9609 60.7792 24.7039 60.1626 24.7039H14.8375C14.2209 24.7039 13.6556 24.9609 13.1417 25.4748C12.6278 25.9887 12.3709 26.5539 12.3709 27.1706V42.8185Z"
            fill="url(#paint0_angular_2956_38220)"
        />
        <defs>
            <radialGradient
                id="paint0_angular_2956_38220"
                cx="0"
                cy="0"
                r="1"
                gradientUnits="userSpaceOnUse"
                gradientTransform="translate(37 35) rotate(-2.52144) scale(52.9813 46.8594)"
            >
                <stop stopColor="#EC4D49" />
                <stop offset="0.262672" stopColor="#7048E8" />
                <stop offset="0.465801" stopColor="#4AC1E8" />
                <stop offset="0.752264" stopColor="#4D0B79" />
                <stop offset="1" stopColor="#FF5543" />
            </radialGradient>
        </defs>
    </svg>
)

const PersonalIcon = (): JSX.Element => (
    <svg width="60" height="60" viewBox="0 0 75 75" fill="none">
        <path
            d="M16.8801 61.0698V32.4719L10.4822 37.4052L9.4801 36.0948L16.8801 30.4292V22.9521H18.5759V29.0031L37.5384 14.5885L65.5582 36.1333L64.5561 37.4052L58.1968 32.4719V61.0698H16.8801ZM18.5759 59.374H33.1061V43.9573H41.8936V59.374H56.5009V31.2385L37.5384 16.7469L18.5759 31.2385V59.374ZM16.8801 18.8667C16.8544 17.2479 17.3311 15.8797 18.3102 14.762C19.2893 13.6443 20.7718 13.0854 22.7579 13.0854C24.3521 13.0854 25.454 12.6872 26.0637 11.8906C26.6732 11.0941 27.0037 10.1306 27.0551 9H28.7509C28.7252 10.6187 28.2336 11.987 27.2759 13.1047C26.3182 14.2224 24.8249 14.7812 22.7961 14.7812C21.216 14.7812 20.1176 15.176 19.5009 15.9654C18.8843 16.7549 18.5759 17.722 18.5759 18.8667H16.8801Z"
            fill="#EFF2F5"
        />
        <path
            d="M16.8801 61.0698V32.4719L10.4822 37.4052L9.4801 36.0948L16.8801 30.4292V22.9521H18.5759V29.0031L37.5384 14.5885L65.5582 36.1333L64.5561 37.4052L58.1968 32.4719V61.0698H16.8801ZM18.5759 59.374H33.1061V43.9573H41.8936V59.374H56.5009V31.2385L37.5384 16.7469L18.5759 31.2385V59.374ZM16.8801 18.8667C16.8544 17.2479 17.3311 15.8797 18.3102 14.762C19.2893 13.6443 20.7718 13.0854 22.7579 13.0854C24.3521 13.0854 25.454 12.6872 26.0637 11.8906C26.6732 11.0941 27.0037 10.1306 27.0551 9H28.7509C28.7252 10.6187 28.2336 11.987 27.2759 13.1047C26.3182 14.2224 24.8249 14.7812 22.7961 14.7812C21.216 14.7812 20.1176 15.176 19.5009 15.9654C18.8843 16.7549 18.5759 17.722 18.5759 18.8667H16.8801Z"
            fill="url(#paint0_angular_2956_38214)"
        />
        <defs>
            <radialGradient
                id="paint0_angular_2956_38214"
                cx="0"
                cy="0"
                r="1"
                gradientUnits="userSpaceOnUse"
                gradientTransform="translate(40 39) rotate(-10.6983) scale(53.2475 49.9663)"
            >
                <stop stopColor="#EC4D49" />
                <stop offset="0.262672" stopColor="#7048E8" />
                <stop offset="0.465801" stopColor="#4AC1E8" />
                <stop offset="0.752264" stopColor="#4D0B79" />
                <stop offset="1" stopColor="#FF5543" />
            </radialGradient>
        </defs>
    </svg>
)
*/
