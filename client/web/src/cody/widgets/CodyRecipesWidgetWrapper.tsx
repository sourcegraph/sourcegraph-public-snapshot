import React, { RefObject, useEffect, useState } from 'react'

import { createPortal } from 'react-dom'

import { ChatEditor } from '../components/ChatEditor'
import { CodyChatStore } from '../useCodyChat'

import { CodyRecipesWidget } from './CodyRecipesWidget'
import { useTextSelection } from './useTextSelection'

interface RecipesWidgetWrapperProps {
    targetRef: RefObject<HTMLElement>
    children: any
    codyChatStore: CodyChatStore
    fileName?: string
    repoName?: string
    revision?: string
}

export const CodyRecipesWidgetWrapper: React.FunctionComponent<RecipesWidgetWrapperProps> = React.memo(
    function CodyRecipesWidgetWrapper({ targetRef, children, codyChatStore }) {
        return (
            <>
                {children}
                {targetRef.current && <RecipePopoverManager targetRef={targetRef} codyChatStore={codyChatStore} />}
            </>
        )
    }
)

const RecipePopoverManager: React.FunctionComponent<{
    targetRef: RefObject<HTMLElement>
    codyChatStore: CodyChatStore
}> = React.memo(function ReacipePopoverMangerComponent({ targetRef, codyChatStore }) {
    const [showPopover, setShowPopover] = useState(false)
    const { isCollapsed, textContent } = useTextSelection(targetRef?.current || undefined)

    useEffect(() => {
        setShowPopover(!isCollapsed && !!textContent)
    }, [isCollapsed, textContent])

    if (!showPopover) {
        return null
    }

    const selection = window.getSelection()

    // Restrict popover to only code content.
    // Hack because Cody's dangerouslySetInnerHTML forces us to use a ref on code block's wrapper text
    if (
        !selection?.rangeCount ||
        selection?.getRangeAt(0)?.commonAncestorContainer?.nodeName !== 'CODE' ||
        !textContent
    ) {
        return null
    }

    return (
        <RecipePopoverPortal
            key={textContent}
            targetRef={targetRef}
            codyChatStore={codyChatStore}
            selectedText={textContent || ''}
        />
    )
})

function getElementFromNode(node: any): HTMLElement {
    const currentElement =
        node.previousElementSibling?.nextElementSibling || node.nextElementSibling?.previousElementSibling

    if (currentElement) {
        return currentElement
    }

    const lastElementChild = node.parentElement?.lastElementChild
    if (lastElementChild) {
        if (lastElementChild.className.includes('cody-recipe-widget')) {
            return lastElementChild.previousElementSibling || node.parentElement
        }

        return lastElementChild
    }
    return node.parentElement
}

const RecipePopoverPortal: React.FunctionComponent<{
    targetRef: RefObject<HTMLElement>
    codyChatStore: CodyChatStore
    selectedText: string
}> = function ReacipePopoverPortalComponent({ targetRef, codyChatStore, selectedText }) {
    const selection = window.getSelection()

    const commonAncestorContainer = selection?.getRangeAt(0)?.commonAncestorContainer as any
    if (!commonAncestorContainer) {
        return null
    }

    const positioningElement = getElementFromNode(selection.focusNode)
    if (!positioningElement) {
        return null
    }

    const positioningClientRect = positioningElement.getBoundingClientRect()

    const mountContainer = commonAncestorContainer.lastElementChild || commonAncestorContainer
    const mountContainerRect = mountContainer.getBoundingClientRect()

    return createPortal(
        <CodyRecipesWidget
            className="cody-recipe-widget"
            style={{
                position: 'absolute',
                'margin-top': `-${mountContainerRect.top - positioningClientRect.top + positioningClientRect.height}px`,
            }}
            codyChatStore={codyChatStore}
            editor={
                new ChatEditor({
                    content: targetRef?.current?.innerText || '',
                    selectedText,
                    filename: '',
                    repo: '',
                    revision: '',
                })
            }
        />,
        commonAncestorContainer.lastChild?.previousElementSibling || commonAncestorContainer
    )
}
