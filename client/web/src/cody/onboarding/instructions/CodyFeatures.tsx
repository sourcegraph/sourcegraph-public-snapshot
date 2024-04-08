import { H2, Text, Link, Button } from '@sourcegraph/wildcard'

import { EditorStep } from '../../management/CodyManagementPage'

export function CodyFeatures({
    onClose,
    showStep,
    setStep,
}: {
    onClose: () => void
    showStep?: EditorStep
    setStep: (step: EditorStep) => void
}): JSX.Element {
    return (
        <>
            <div className="mb-3 pb-3 border-bottom">
                <H2>Cody features</H2>
            </div>
            <div className="d-flex">
                <div className="flex-1 p-3 border-right d-flex flex-column justify-content-center align-items-center">
                    <Text className="mb-1 w-100" weight="bold">
                        Autocomplete
                    </Text>
                    <Text className="mb-0 w-100 text-muted" size="small">
                        Let Cody automatically write code for you. Start writing a comment or a line of code and Cody
                        will suggest the next few lines.
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
                        Answer questions about programming topics generally or your codebase specifically with Cody
                        chat.
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
                        <Link to="https://discord.gg/rDPqBejz93" className="d-flex w-100 justify-content-center ">
                            <strong>Discord chat</strong>
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
    )
}
