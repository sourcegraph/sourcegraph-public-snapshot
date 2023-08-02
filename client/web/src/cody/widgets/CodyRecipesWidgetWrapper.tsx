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

export const RecipesWidgetWrapper: React.FunctionComponent<RecipesWidgetWrapperProps> = React.memo(
    function CodyRecipeWidgetWrapper({ targetRef, transcriptRef, children, codyChatStore }) {
        const [scrolled, setScrolled] = useState(false)

        useEffect(() => {
            console.log('transcriptRef??', transcriptRef)
            const updateScrollPosition = () => {
                console.log('HIT!', window.pageYOffset, window.scrollY)
                setScrolled(!scrolled)
            }

            if (transcriptRef && transcriptRef.current) {
                transcriptRef.current.addEventListener('scroll', updateScrollPosition, false)
            }
            return () => {
                if (transcriptRef && transcriptRef.current) {
                    transcriptRef?.current.removeEventListener('scroll', updateScrollPosition, false)
                }
            }
        }, [])

        return (
            <>
                {children}
                <RecipePopover refresh={scrolled} targetRef={targetRef} codyChatStore={codyChatStore} />
                {/* <Popover
                    target={targetRef.current || undefined}
                    // mount={targetRef.current || undefined}
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

                        console.log(targetRef.current, clientRect)

                        return (
                            <CodyRecipesWidget
                                // className={styles.chatCodeSnippetPopover}
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
                /> */}
            </>
        )
    }
)

const RecipePopover: React.FunctionComponent<{
    targetRef: RefObject<HTMLElement>
    codyChatStore: CodyChatStore
    refresh: boolean
}> = ({ targetRef, codyChatStore, refresh }) => {
    const [show, setShow] = useState(false)
    const [remount, setRemount] = useState(false)

    useEffect(() => {
        setRemount(!refresh)
    }, [refresh])

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
}
