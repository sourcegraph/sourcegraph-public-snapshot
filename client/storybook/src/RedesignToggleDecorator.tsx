import { DecoratorFunction } from '@storybook/addons'
import isChromatic from 'chromatic/isChromatic'
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

export const RedesignToggleDecorator: DecoratorFunction<ReactElement> = (Story, context) => {
    const [isRedesignEnabled, setIsRedesignEnabled] = useRedesignToggle()

    useEffect(() => {
        // In Chromatic set disable redesign theme by default. It will be enabled by ChromaticStoryDecorator if it's redesign story.
        if (isChromatic() && isRedesignEnabled) {
            setIsRedesignEnabled(false)
        }
        // We want to do it once on component mount, after that `ChromaticStoryDecorator` will control redesign theme flag.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

    useEffect(() => {
        updatePreview(isRedesignEnabled)
        updateManager(isRedesignEnabled)
    }, [isRedesignEnabled])

    return <Story {...context} />
}
