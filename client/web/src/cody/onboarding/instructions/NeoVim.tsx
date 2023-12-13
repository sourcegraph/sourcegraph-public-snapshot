import { useState } from 'react'

import classNames from 'classnames'

import { Link, H2, Text, Button, ButtonLink } from '@sourcegraph/wildcard'

import styles from '../CodyOnboarding.module.scss'

export function NeoVimInstructions({
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
                        <H2>Setup instructions for NeoVim</H2>
                    </div>

                    <div className={classNames('pt-3 px-3', styles.instructionsContainer)}>
                        <div className={classNames('d-flex flex-column border-bottom')}>
                            <div className="d-flex align-items-center">
                                <div className="mr-1">
                                    <div className={classNames('mr-2', styles.step)}>1</div>
                                </div>
                                <div>
                                    <Text className="mb-1" weight="bold">
                                        Open the plugin repo on GitHub
                                    </Text>
                                    <Text className="text-muted mb-0" size="small">
                                        Follow the instructions detailed in the <strong>Readme.md</strong> file to
                                        install the plugin.
                                    </Text>
                                </div>
                            </div>
                            <div className="d-flex flex-column justify-content-center align-items-center mt-4">
                                <ButtonLink
                                    variant="primary"
                                    to="https://github.com/sourcegraph/sg.nvim#setup"
                                    target="_blank"
                                >
                                    Navigate to GitHub Repo
                                </ButtonLink>
                                <img
                                    alt="NeoVim Repo"
                                    className="mt-2 m-auto"
                                    width="70%"
                                    src="https://storage.googleapis.com/sourcegraph-assets/NeoVimInstructions/NeovimStep1.png"
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
                                Let Cody automatically write code for you. Start writing a comment or a line of code and
                                Cody will suggest the next few lines.
                            </Text>
                            <img
                                alt="Cody Autocomplete"
                                width="100%"
                                className="mt-4"
                                src="https://storage.googleapis.com/sourcegraph-assets/codyFeaturesImgs/featureAutoCompletions.png"
                            />
                        </div>
                        <div className="flex-1 p-3 d-flex flex-column justify-content-center align-items-center">
                            <Text className="mb-1  w-100" weight="bold">
                                Chat
                            </Text>
                            <Text className="mb-0 text-muted  w-100" size="small">
                                Answer questions about programming topics generally or your codebase specifically with
                                Cody chat.
                            </Text>
                            <img
                                alt="Cody Chat"
                                width="100%"
                                className="mt-4"
                                src="https://storage.googleapis.com/sourcegraph-assets/codyFeaturesImgs/featureChat.png"
                            />
                        </div>
                    </div>
                    <div className="d-flex my-3 py-3 border-top border-bottom">
                        <div className="flex-1 p-3 border-right d-flex flex-column justify-content-center align-items-center">
                            <Text className="mb-1  w-100" weight="bold">
                                Commands
                            </Text>
                            <Text className="mb-0 text-muted  w-100" size="small">
                                Streamline your development process by using Cody commands to understand, improve, fix,
                                document, and generate unit tests for your code.
                            </Text>
                            <img
                                alt="Cody Commands"
                                width="100%"
                                className="mt-4"
                                src="https://storage.googleapis.com/sourcegraph-assets/codyFeaturesImgs/featureCommands.png"
                            />
                        </div>
                        <div className="flex-1 p-3 d-flex flex-column justify-content-center align-items-center">
                            <Text className="mb-1  w-100" weight="bold">
                                Feedback
                            </Text>
                            <Text className="mb-0 text-muted w-100" size="small">
                                Feel free to join our Discord to leave feedback or ask questions about Cody.
                            </Text>
                            <div className="mt-4 d-flex flex-column justify-content-center h-100">
                                <Link
                                    to="https://discord.gg/rDPqBejz93"
                                    className="d-flex w-100 justify-content-center "
                                >
                                    <strong>Join our Discord</strong>
                                </Link>
                                <Link
                                    to="https://github.com/sourcegraph/cody/discussions/new?category=product-feedback"
                                    className="d-flex w-100 justify-content-center mt-4"
                                >
                                    <strong>GitHub Discussions</strong>
                                </Link>
                            </div>
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
