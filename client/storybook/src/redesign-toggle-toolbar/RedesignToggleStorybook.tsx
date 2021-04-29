import { addons } from '@storybook/addons'
import { Icons, IconButton } from '@storybook/components'
import { SET_STORIES } from '@storybook/core-events'
import React, { ReactElement, useEffect } from 'react'

import { useRedesignToggle, REDESIGN_CLASS_NAME } from '@sourcegraph/shared/src/util/useRedesignToggle'

const toggleRedesignClass = (element: HTMLElement, isRedesignEnabled: boolean): void => {
    element.classList.toggle(REDESIGN_CLASS_NAME, isRedesignEnabled)
}

const updatePreview = (isRedesignEnabled: boolean): void => {
    const iframe = document.querySelector('#storybook-preview-iframe') as HTMLIFrameElement | undefined

    const iframeDocument = iframe?.contentDocument || iframe?.contentWindow?.document
    const body = iframeDocument?.body

    if (body) {
        toggleRedesignClass(body, isRedesignEnabled)
    }
}

const updateManager = (isRedesignEnabled: boolean): void => {
    const manager = document.querySelector('body')

    if (manager) {
        toggleRedesignClass(manager, isRedesignEnabled)
    }
}

export const RedesignToggleStorybook = (): ReactElement => {
    const [isRedesignEnabled, setIsRedesignEnabled] = useRedesignToggle()

    useEffect(() => {
        const handleIsRedesignEnabledChange = (): void => {
            updatePreview(isRedesignEnabled)
            updateManager(isRedesignEnabled)
        }

        handleIsRedesignEnabledChange()

        const channel = addons.getChannel()
        // Preview iframe is not available on toolbar mount.
        // Wait for the SET_STORIES event, after which the iframe is accessible, and ensure that the redesign-theme class is in place.
        channel.on(SET_STORIES, handleIsRedesignEnabledChange)

        return () => {
            channel.removeListener(SET_STORIES, handleIsRedesignEnabledChange)
        }
    }, [isRedesignEnabled])

    const handleRedesignToggle = (): void => {
        setIsRedesignEnabled(!isRedesignEnabled)
    }

    return (
        <IconButton
            key="redesign-toolbar"
            active={isRedesignEnabled}
            title={isRedesignEnabled ? 'Disable redesign theme' : 'Enable redesign theme'}
            // eslint-disable-next-line react/jsx-no-bind
            onClick={handleRedesignToggle}
        >
            <Icons icon="beaker" />
        </IconButton>
    )
}
