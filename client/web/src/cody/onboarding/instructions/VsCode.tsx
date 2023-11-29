import { useState } from 'react'

import classNames from 'classnames'

import { H2, Text, Button } from '@sourcegraph/wildcard'

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
                    <div className="mb-3 pb-3 border-bottom">
                        <H2>Setup instructions for VS Code</H2>
                    </div>
                    <div className="p-3">
                        <div className="d-flex">
                            <div>
                                <div className={classNames('mr-2', styles.step)}>1</div>
                            </div>
                            <div>
                                <Text className="mb-1" weight="bold">
                                    Install Cody
                                </Text>
                                <Text className="text-muted mb-0" size="small">
                                    Alternatively, you can reach this page by clicking{' '}
                                    <strong>View {'>'} Extensions</strong> and searching for <strong>Cody AI</strong>
                                </Text>
                            </div>
                        </div>
                        <div className="d-flex flex-column justify-content-center align-items-center mt-4">
                            <Button variant="primary">Open Marketplace</Button>
                            <img
                                alt="VS Code Marketplace"
                                className="mt-4"
                                src="https://storage.googleapis.com/sourcegraph-assets/VSCodeInstructions/step1.png"
                            />
                        </div>
                    </div>
                    <div className="p-3">
                        <div className="d-flex">
                            <div>
                                <div className={classNames('mr-2', styles.step)}>2</div>
                            </div>
                            <div>
                                <Text className="mb-1" weight="bold">
                                    Open Cody from the Sidebar on the Left
                                </Text>
                                <Text className="text-muted mb-0" size="small">
                                    Alternatively, you can reach this page by clicking{' '}
                                    <strong>View {'>'} Extensions</strong> and searching for <strong>Cody AI</strong>
                                </Text>
                            </div>
                        </div>
                        <div className="d-flex flex-column justify-content-center align-items-center mt-4">
                            <img
                                alt="VS Code Marketplace"
                                className="mt-4"
                                src="https://storage.googleapis.com/sourcegraph-assets/VSCodeInstructions/step2.png"
                            />
                        </div>
                    </div>
                    <div className="d-flex p-3 border-bottom">
                        <div>
                            <div className={classNames('mr-2', styles.step)}>3</div>
                        </div>
                        <div className="mb-4 pb-4">
                            <Text className="mb-1" weight="bold">
                                Login
                            </Text>
                            <Text className="text-muted mb-0" size="small">
                                Choose the same login method you used when you created your account
                            </Text>
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
                        <H2>Using Cody on VS Code</H2>
                    </div>
                    <div className="d-flex">
                        <div className="flex-1 p-3 border-right">
                            <Text className="mb-1" weight="bold">
                                Autocomplete
                            </Text>
                            <Text className="mb-0 text-muted" size="small">
                                Cody will autocomplete your code as you type
                            </Text>
                        </div>
                        <div className="flex-1 p-3">
                            <Text className="mb-1" weight="bold">
                                Chat
                            </Text>
                            <Text className="mb-0 text-muted" size="small">
                                Cody will autocomplete your code as you type
                            </Text>
                        </div>
                    </div>
                    <div className="d-flex my-3 py-3 border-top border-bottom">
                        <div className="flex-1 p-3 border-right">
                            <Text className="mb-1" weight="bold">
                                Settings
                            </Text>
                            <Text className="mb-0 text-muted" size="small">
                                Cody will autocomplete your code as you type
                            </Text>
                        </div>
                        <div className="flex-1 p-3">
                            <Text className="mb-1" weight="bold">
                                Feedback
                            </Text>
                            <Text className="mb-0 text-muted" size="small">
                                Cody will autocomplete your code as you type
                            </Text>
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
