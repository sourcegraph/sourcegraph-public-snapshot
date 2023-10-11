import '@sourcegraph/branded'

import type { Meta, StoryObj } from '@storybook/react'

import { Container } from '@sourcegraph/wildcard'
import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import '@storybook/addon-designs'

import { useEffect, useState } from 'react'

import { UpdateInfoContent } from './UpdateInfoContent'
import type { UpdateInfo } from './updater'

const meta: Meta<typeof UpdateInfoContent> = {
    title: 'cody-ui/Updater/Content',
    decorators: [
        Story => (
            <BrandedStory>
                {() => (
                    <Container className="container mt-3 pb-3" style={{ border: '1px dashed gray' }}>
                        <Story />
                    </Container>
                )}
            </BrandedStory>
        ),
    ],
    component: UpdateInfoContent,
}

export default meta

type Story = StoryObj<typeof UpdateInfoContent>

export const NoUpdates: Story = {
    render: function Render() {
        const [state, setState] = useState<UpdateInfo>({
            stage: 'CHECKING',
            hasNewVersion: true,
            version: '1.0.0',
        })
        useEffect(() => {
            const timer1 = setTimeout(() => {
                setState(
                    (state: UpdateInfo): UpdateInfo => ({
                        ...state,
                        stage: 'IDLE',
                        hasNewVersion: false,
                    })
                )
            }, 1000)
            return () => {
                clearTimeout(timer1)
            }
        }, [setState])
        return <UpdateInfoContent details={state} />
    },
}

export const HasUpdates: Story = {
    render: function Render() {
        const [state, setState] = useState<UpdateInfo>({
            stage: 'CHECKING',
            hasNewVersion: true,
            version: '1.0.0',
        })
        useEffect(() => {
            const timer1 = setTimeout(() => {
                setState(
                    (state: UpdateInfo): UpdateInfo => ({
                        ...state,
                        stage: 'IDLE',
                        hasNewVersion: true,
                        newVersion: '1.2.4',
                    })
                )
            }, 1000)
            return () => {
                clearTimeout(timer1)
            }
        }, [setState])
        return <UpdateInfoContent details={state} />
    },
}

export const CheckNow: Story = {
    render: function Render() {
        const [state, setState] = useState<UpdateInfo>({
            stage: 'IDLE',
            hasNewVersion: false,
            version: '1.0.0',
            checkNow: () => {
                setState(
                    (state: UpdateInfo): UpdateInfo => ({
                        ...state,
                        stage: 'CHECKING',
                    })
                )
                setTimeout(() => {
                    setState(
                        (state: UpdateInfo): UpdateInfo => ({
                            ...state,
                            stage: 'IDLE',
                            hasNewVersion: true,
                            newVersion: '2.0.0',
                        })
                    )
                }, 1000)
            },
        })
        return <UpdateInfoContent details={state} />
    },
}

export const ReleaseDetails: Story = {
    render: function Render() {
        const [state] = useState<UpdateInfo>({
            stage: 'IDLE',
            hasNewVersion: true,
            version: '1.0.0',
            newVersion: '1.2.3',
            description: 'This is the description of the update.',
        })
        return <UpdateInfoContent details={state} />
    },
}

interface InstallStoryProps {
    installCompletedEvent: (when: Date) => void
}
type InstallStoryWrapper = StoryObj<InstallStoryProps>

export const InstallNow: InstallStoryWrapper = {
    render: function Render({ installCompletedEvent }) {
        const [state, setState] = useState<UpdateInfo>({
            stage: 'IDLE',
            hasNewVersion: true,
            version: '1.0.0',
            newVersion: '2.0.0',
            startInstall: () => {
                setState(
                    (state: UpdateInfo): UpdateInfo => ({
                        ...state,
                        stage: 'INSTALLING',
                    })
                )
                setTimeout(() => {
                    installCompletedEvent(new Date())
                    setState(
                        (state: UpdateInfo): UpdateInfo => ({
                            ...state,
                            stage: 'IDLE',
                            hasNewVersion: false,
                            version: '2.0.0',
                        })
                    )
                }, 1000)
            },
        })
        return (
            <div>
                After pressing install now, check the Actions panel for completion.
                <Container className="container mt-3 pb-3" style={{ border: '1px dashed gray' }}>
                    <UpdateInfoContent details={state} />
                </Container>
            </div>
        )
    },
    argTypes: {
        installCompletedEvent: { action: 'install successful' },
    },
}

export const InstallFailed: Story = {
    render: function Render() {
        const [state, setState] = useState<UpdateInfo>({
            stage: 'IDLE',
            hasNewVersion: true,
            version: '1.0.0',
            newVersion: '2.0.0',
            startInstall: () => {
                setState(
                    (state: UpdateInfo): UpdateInfo => ({
                        ...state,
                        stage: 'INSTALLING',
                    })
                )
                setTimeout(() => {
                    setState(
                        (state: UpdateInfo): UpdateInfo => ({
                            ...state,
                            stage: 'ERROR',
                            error: 'Install failed for some reason. Who knows!?',
                        })
                    )
                }, 1000)
            },
        })
        return <UpdateInfoContent details={state} />
    },
}

export const InstallFailedTryAgain: Story = {
    render: function Render() {
        const [state, setState] = useState<UpdateInfo>({
            stage: 'IDLE',
            hasNewVersion: true,
            version: '1.0.0',
            newVersion: '2.0.0',
            checkNow: () => {
                setState(
                    (state: UpdateInfo): UpdateInfo => ({
                        ...state,
                        stage: 'CHECKING',
                    })
                )
                setTimeout(() => {
                    setState(
                        (state: UpdateInfo): UpdateInfo => ({
                            ...state,
                            stage: 'IDLE',
                            error: undefined,
                        })
                    )
                }, 1000)
            },
            startInstall: () => {
                setState(
                    (state: UpdateInfo): UpdateInfo => ({
                        ...state,
                        stage: 'INSTALLING',
                    })
                )
                setTimeout(() => {
                    setState(
                        (state: UpdateInfo): UpdateInfo => ({
                            ...state,
                            stage: 'ERROR',
                            error: 'Install failed for some reason. Who knows!?',
                        })
                    )
                }, 1000)
            },
        })
        return <UpdateInfoContent details={state} />
    },
}
