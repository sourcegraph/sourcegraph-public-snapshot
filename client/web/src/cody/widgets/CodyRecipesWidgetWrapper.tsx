import React, { RefObject, ReactNode, useEffect, useState } from 'react'

import { Popover } from 'react-text-selection-popover'

import { ChatEditor } from '../components/ChatEditor'
import { CodyChatStore } from '../useCodyChat'

import { CodyRecipesWidget } from './CodyRecipesWidget'

import styles from './CodyRecipesWidget.module.scss'

interface RecipesWidgetWrapperProps {
    targetRef: RefObject<HTMLElement>
    children: ReactNode | ReactNode[]
    codyChatStore: CodyChatStore
    fileName?: string
    repoName?: string
    revision?: string
}

export const RecipesWidgetWrapper: React.FunctionComponent<RecipesWidgetWrapperProps> = React.memo(
    function CodyRecipeWidgetWrapper({ targetRef, children, codyChatStore }) {
        const [show, setShow] = useState(false)

        return (
            <>
                {children}
                <Popover
                    target={targetRef.current || undefined}
                    mount={targetRef.current || undefined}
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
                                className={styles.chatCodeSnippetPopover}
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
            </>
        )
    }
)
