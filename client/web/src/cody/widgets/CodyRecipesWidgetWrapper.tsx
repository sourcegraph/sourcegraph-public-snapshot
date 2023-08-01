import React, { useEffect, useState } from 'react'

import { Popover } from 'react-text-selection-popover'

import { ChatEditor } from '../components/ChatEditor'

import { CodyRecipesWidget } from './CodyRecipesWidget'

// TODO: fix the types
interface RecipesWidgetWrapperProps {
    targetRef: any
    children: any
    codyChatStore: any
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
                    target={targetRef.current}
                    mount={targetRef.current}
                    render={({ clientRect, isCollapsed, textContent }) => {
                        useEffect(() => {
                            setShow(!isCollapsed)
                        }, [isCollapsed, setShow])

                        if (!clientRect || isCollapsed || !targetRef || !show) {
                            return null
                        }

                        // Allow popover only on code content.
                        // Hack because Cody's dangerouslySetInnerHTML forces us to use a ref on code block's wrapper text
                        if (window.getSelection()?.anchorNode?.parentNode?.nodeName !== 'CODE') {
                            return null
                        }

                        console.log(targetRef.current, clientRect)

                        return (
                            <CodyRecipesWidget
                                style={{
                                    // TODO: Move these styles
                                    position: 'absolute',
                                    bottom: '-30px',
                                    left: '0',
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
            </>
        )
    }
)
