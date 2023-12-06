import { useState } from 'react'

import classNames from 'classnames'
import { Link } from 'react-router-dom'

import { H2, Text, Button, ButtonLink } from '@sourcegraph/wildcard'

import styles from '../CodyOnboarding.module.scss'

export function VSCodeInstructions({
    onBack,
    onClose,
    showStep,
}: {
    onBack?: () => void
    onClose: () => void
    showStep?: number
}): JSX.Element {
    const [step, setStep] = useState<number>(showStep || 0)

    return (
        <>
            {step === 0 && (
                <>
                    <div className="pb-3 border-bottom">
                        <H2>Setup instructions for VS Code</H2>
                    </div>

                    <div className={classNames('pt-3 px-3', styles.instructionsContainer)}>
                        <div className={classNames('border-bottom', styles.highlightStep)}>
                            <div className="d-flex align-items-center">
                                <div className="mr-1">
                                    <div className={classNames('mr-2', styles.step)}>1</div>
                                </div>
                                <div>
                                    <Text className="mb-1" weight="bold">
                                        Install Cody
                                    </Text>
                                    <Text className="text-muted mb-0" size="small">
                                        Alternatively, you can reach this page by clicking{' '}
                                        <strong>View {'>'} Extensions</strong> and searching for{' '}
                                        <strong>Cody AI</strong>
                                    </Text>
                                </div>
                            </div>
                            <div className="d-flex flex-column justify-content-center align-items-center mt-4">
                                <ButtonLink
                                    variant="primary"
                                    to="https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai"
                                    target="_blank"
                                >
                                    Open Marketplace
                                </ButtonLink>
                                <img
                                    alt="VS Code Marketplace"
                                    className="mt-4"
                                    width="70%"
                                    src="https://storage.googleapis.com/sourcegraph-assets/VSCodeInstructions/__step1.png"
                                />
                            </div>
                        </div>
                        <div className="mt-3 border-bottom">
                            <div className="d-flex align-items-center">
                                <div className="mr-1">
                                    <div className={classNames('mr-2', styles.step)}>2</div>
                                </div>
                                <div>
                                    <Text className="mb-1" weight="bold">
                                        Open Cody from the Sidebar on the Left
                                    </Text>
                                    <Text className="text-muted mb-0" size="small">
                                        Typically Cody will be the last item in the sidebar
                                    </Text>
                                </div>
                            </div>
                            <div className="d-flex flex-column justify-content-center align-items-center mt-4">
                                <img
                                    alt="VS Code Marketplace"
                                    className="mt-2"
                                    width="70%"
                                    src="https://storage.googleapis.com/sourcegraph-assets/VSCodeInstructions/__step2.png"
                                />
                            </div>
                        </div>
                        <div className="mt-3 border-bottom">
                            <div className="d-flex align-items-center">
                                <div className="mr-1">
                                    <div className={classNames('mr-2', styles.step)}>3</div>
                                </div>
                                <div>
                                    <Text className="mb-1" weight="bold">
                                        Login
                                    </Text>
                                    <Text className="text-muted mb-0" size="small">
                                        Choose the same login method you used when you created your account
                                    </Text>
                                </div>
                            </div>
                            <div className="d-flex flex-column justify-content-center align-items-center mt-4">
                                <img
                                    alt="VS Code Marketplace"
                                    className="mt-2"
                                    width="70%"
                                    src="https://storage.googleapis.com/sourcegraph-assets/VSCodeInstructions/__step3.png"
                                />
                            </div>
                        </div>
                    </div>

                    {showStep === undefined ? (
                        <div className="mt-3 d-flex justify-content-between">
                            <Button variant="secondary" onClick={onBack} outline={true} size="sm">
                                Back
                            </Button>
                            <Button variant="primary" onClick={() => setStep(1)} size="sm">
                                Next
                            </Button>
                        </div>
                    ) : (
                        <div className="mt-3 d-flex justify-content-end">
                            <Button variant="primary" onClick={onClose} size="sm">
                                Close
                            </Button>
                        </div>
                    )}
                </>
            )}
            {step === 1 && (
                <>
                    <div className="mb-3 pb-3 border-bottom">
                        <H2>Cody Features</H2>
                    </div>
                    <div className="d-flex">
                        <div className="flex-1 p-3 border-right d-flex flex-column justify-content-center align-items-center">
                            <Text className="mb-1 w-100" weight="bold">
                                Autocomplete
                            </Text>
                            <Text className="mb-0 w-100 text-muted" size="small">
                                Cody will autocomplete your code as you type
                            </Text>
                            <img
                                alt="Cody Autocomplete"
                                width="90%"
                                className="mt-4"
                                src="https://storage.googleapis.com/sourcegraph-assets/VSCodeInstructions/autoCompleteIllustration.svg"
                            />
                        </div>
                        <div className="flex-1 p-3 d-flex flex-column justify-content-center align-items-center">
                            <Text className="mb-1  w-100" weight="bold">
                                Chat
                            </Text>
                            <Text className="mb-0 text-muted  w-100" size="small">
                                Cody will autocomplete your code as you type
                            </Text>
                            <img
                                alt="Cody Chat"
                                width="80%"
                                className="mt-4"
                                src="https://storage.googleapis.com/sourcegraph-assets/VSCodeInstructions/chatIllustration.svg"
                            />
                        </div>
                    </div>
                    <div className="d-flex my-3 py-3 border-top border-bottom">
                        <div className="flex-1 p-3 border-right d-flex flex-column justify-content-center align-items-center">
                            <Text className="mb-1  w-100" weight="bold">
                                Commands
                            </Text>
                            <Text className="mb-0 text-muted  w-100" size="small">
                                Cody will autocomplete your code as you type
                            </Text>
                            <img
                                alt="Cody Commands"
                                width="80%"
                                className="mt-4"
                                src="https://storage.googleapis.com/sourcegraph-assets/VSCodeInstructions/commandsIllustration.svg"
                            />
                        </div>
                        <div className="flex-1 p-3 d-flex flex-column justify-content-center align-items-center">
                            <Text className="mb-1  w-100" weight="bold">
                                Feedback
                            </Text>
                            <Text className="mb-0 text-muted w-100" size="small">
                                Cody will autocomplete your code as you type
                            </Text>
                            <Link to="https://discord.gg/rDPqBejz93" className="d-flex w-100 justify-content-center">
                                <img
                                    alt="Discord Feedback"
                                    width="50%"
                                    className="mt-4"
                                    src="https://storage.googleapis.com/sourcegraph-assets/VSCodeInstructions/discordCTA.svg"
                                />
                            </Link>
                            <Link
                                to="https://github.com/sourcegraph/cody/discussions/new?category=product-feedback"
                                className="d-flex w-100 justify-content-center"
                            >
                                <img
                                    alt="GitHub Feedback"
                                    width="50%"
                                    className="mt-4"
                                    src="https://storage.googleapis.com/sourcegraph-assets/VSCodeInstructions/feedbackCTA.svg"
                                />
                            </Link>
                        </div>
                    </div>
                    {showStep === undefined ? (
                        <div className="mt-3 d-flex justify-content-between">
                            <Button variant="secondary" onClick={() => setStep(0)} outline={true} size="sm">
                                Back
                            </Button>
                            <Button variant="primary" onClick={onClose} size="sm">
                                Close
                            </Button>
                        </div>
                    ) : (
                        <div className="mt-3 d-flex justify-content-end">
                            <Button variant="primary" onClick={onClose} size="sm">
                                Close
                            </Button>
                        </div>
                    )}
                </>
            )}
        </>
    )
}
