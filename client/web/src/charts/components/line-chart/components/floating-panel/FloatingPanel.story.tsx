import { Meta } from '@storybook/react'
import React, { useState } from 'react'

import { WebStory } from '../../../../../components/WebStory'

import { FloatingPanel } from './FloatingPanel'
import styles from './FloatingPanel.story.module.scss'

export default {
    title: 'views/floating-panel',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
} as Meta

export const FloatingPanelExample = () => {
    const [buttonElement, setButtonElement] = useState<HTMLButtonElement | null>(null)

    return (
        <div className={styles.container}>
            <div className={styles.content}>
                <button className={styles.target} ref={setButtonElement}>
                    Hello
                </button>

                {buttonElement && (
                    <FloatingPanel
                        className={styles.floating}
                        strategy="absolute"
                        placement="right-end"
                        target={buttonElement}
                    >
                        World <br />
                        World <br />
                    </FloatingPanel>
                )}
            </div>
        </div>
    )
}
