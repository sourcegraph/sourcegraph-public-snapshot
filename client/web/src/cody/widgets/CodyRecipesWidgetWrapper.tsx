import React, { RefObject, useEffect, useState } from 'react'

import { Popover } from 'react-text-selection-popover'

import { ChatEditor } from '../components/ChatEditor'
import { CodyChatStore } from '../useCodyChat'

import { CodyRecipesWidget } from './CodyRecipesWidget'

interface RecipesWidgetWrapperProps {
    targetRef: RefObject<HTMLElement>
    children: any
    codyChatStore: CodyChatStore
    transcriptRef?: RefObject<HTMLElement>
    fileName?: string
    repoName?: string
    revision?: string
}

export const CodyRecipesWidgetWrapper: React.FunctionComponent<RecipesWidgetWrapperProps> = React.memo(
    function CodyRecipesWidgetWrapper({ targetRef, transcriptRef, children, codyChatStore }) {
        const [scrollPosition, setScrollPosition] = useState<number>(0)

        useEffect(() => {
            const updateScrollPosition = () => {
                if (transcriptRef?.current) {
                    setScrollPosition(transcriptRef.current?.scrollTop)
                    console.log('SCROLLED!', transcriptRef.current?.scrollTop)
                }
            }

            if (transcriptRef?.current) {
                transcriptRef.current.addEventListener('scroll', updateScrollPosition, { passive: true })
            }
            return () => {
                if (transcriptRef && transcriptRef.current) {
                    transcriptRef?.current.removeEventListener('scroll', updateScrollPosition)
                }
            }
        }, [scrollPosition])

        return (
            <>
                {children}
                <RecipePopover scrollPosition={scrollPosition} targetRef={targetRef} codyChatStore={codyChatStore} />
            </>
        )
    }
)

const RecipePopover: React.FunctionComponent<{
    targetRef: RefObject<HTMLElement>
    codyChatStore: CodyChatStore
    scrollPosition: number
}> = React.memo(function RecipePopover({ targetRef, codyChatStore, scrollPosition }) {
    const [show, setShow] = useState(false)
    const [scrollValue, changeScrollValue] = useState(0)
    console.log('child:', scrollPosition)

    useEffect(() => {
        changeScrollValue(scrollPosition)
    }, [scrollPosition, scrollValue])

    return (
        <Popover
            target={targetRef.current || undefined}
            render={({ clientRect, isCollapsed, textContent }) => {
                useEffect(() => {
                    setShow(!isCollapsed)
                }, [isCollapsed, setShow])

                if (!clientRect || isCollapsed || !targetRef || !show) {
                    return null
                }

                // Restrict popover to only code content.
                // Hack because Cody's dangerouslySetInnerHTML forces us to use a ref on code block's wrapper text
                if (window.getSelection()?.anchorNode?.parentNode?.nodeName !== 'CODE') {
                    return null
                }

                return (
                    <CodyRecipesWidget
                        style={{
                            position: 'absolute',
                            left: `${clientRect.left}px`,
                            top: `${clientRect.bottom}px`,
                        }}
                        codyChatStore={codyChatStore}
                        editor={
                            new ChatEditor({
                                content: textContent || '',
                                fullText: targetRef?.current?.innerText || '',
                                filename: '',
                                repo: '',
                                revision: '',
                            })
                        }
                    />
                )
            }}
        />
    )
})
