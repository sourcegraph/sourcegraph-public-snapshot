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
import { NeoVimInstructions } from './instructions/NeoVim'
import { VSCodeInstructions } from './instructions/VsCode'

import styles from './CodyOnboarding.module.scss'

export interface IEditor {
    icon: string
    name: string
    publisher: string
    releaseStage: string
    docs?: string
    instructions?: React.FC<{ onBack?: () => void; onClose: () => void; showStep?: number }>
}

export const editorGroups: IEditor[][] = [
    [
        {
            icon: 'VsCode',
            name: 'VS Code',
            publisher: 'Microsoft',
            releaseStage: 'Stable',
            docs: 'https://sourcegraph.com/docs/cody/clients/install-vscode',
            instructions: VSCodeInstructions,
        },
        {
            icon: 'IntelliJ',
            name: 'IntelliJ IDEA',
            publisher: 'JetBrains',
            releaseStage: 'Beta',
            docs: 'https://sourcegraph.com/docs/cody/clients/install-jetbrains',
            instructions: JetBrainsInstructions,
        },
        {
            icon: 'PhpStorm',
            name: 'PhpStorm ',
            publisher: 'JetBrains',
            releaseStage: 'Beta',
            docs: 'https://sourcegraph.com/docs/cody/clients/install-jetbrains',
            instructions: JetBrainsInstructions,
        },
        {
            icon: 'PyCharm',
            name: 'PyCharm',
            publisher: 'JetBrains',
            releaseStage: 'Beta',
            docs: 'https://sourcegraph.com/docs/cody/clients/install-jetbrains',
            instructions: JetBrainsInstructions,
        },
    ],
    [
        {
            icon: 'WebStorm',
            name: 'WebStorm',
            publisher: 'JetBrains',
            releaseStage: 'Beta',
            docs: 'https://sourcegraph.com/docs/cody/clients/install-jetbrains',
            instructions: JetBrainsInstructions,
        },
        {
            icon: 'RubyMine',
            name: 'RubyMine',
            publisher: 'JetBrains',
            releaseStage: 'Beta',
            docs: 'https://sourcegraph.com/docs/cody/clients/install-jetbrains',
            instructions: JetBrainsInstructions,
        },
        {
            icon: 'GoLand',
            name: 'GoLand',
            publisher: 'JetBrains',
            releaseStage: 'Beta',
            docs: 'https://sourcegraph.com/docs/cody/clients/install-jetbrains',
            instructions: JetBrainsInstructions,
        },
        {
            icon: 'AndroidStudio',
            name: 'Android Studio',
            publisher: 'Google',
            releaseStage: 'Beta',
            docs: 'https://sourcegraph.com/docs/cody/clients/install-jetbrains',
            instructions: JetBrainsInstructions,
        },
    ],
    [
        {
            icon: 'NeoVim',
            name: 'Neovim',
            publisher: 'Neovim Team',
            releaseStage: 'Experimental',
            docs: 'https://sourcegraph.com/docs/cody/clients/install-neovim',
            instructions: NeoVimInstructions,
        },
        {
            icon: 'Emacs',
            name: 'Emacs',
            publisher: 'GNU',
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

    useEffect(() => {
        if (completed && returnToURL) {
            navigate(returnToURL)
        }
    }, [completed, returnToURL, navigate])

    if (completed && returnToURL) {
        return null
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
                                    ? 'https://storage.googleapis.com/sourcegraph-assets/hiCodyWhite.mp4'
                                    : 'https://storage.googleapis.com/sourcegraph-assets/hiCodyDark.mp4'
                            }
                            type="video/mp4"
                        />
                        Your browser does not support the video tag.
                    </video>
                    <Text className={classNames('mb-4 pb-4', styles.fadeIn, styles.fadeSecond, styles.welcomeSubtitle)}>
                        Ready to breeze through the basics and get comfortable with Cody
                        {pro ? ' Pro Trial' : ''}?
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
            </div>
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
