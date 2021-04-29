import { DecoratorFunction } from '@storybook/addons'
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
    const [isRedesignEnabled] = useRedesignToggle()

    useEffect(() => {
        updatePreview(isRedesignEnabled)
        updateManager(isRedesignEnabled)
    }, [isRedesignEnabled])

    return <Story {...context} />
}
