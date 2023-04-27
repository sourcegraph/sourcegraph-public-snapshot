import React from 'react'

import { mdiChevronLeft, mdiChevronRight } from '@mdi/js'

import { Button, Icon } from '@sourcegraph/wildcard'

import { HistoryStack } from './useHistoryStack'

import styles from './TauriTitleBar.module.scss'

export const TauriTitleBar: React.FC<{ historyStack: HistoryStack }> = ({ historyStack }) => (
    <nav data-tauri-drag-region={true} className={styles.titlebar} aria-label="Window navigation">
        <Button
            variant="icon"
            className={styles.button}
            disabled={!historyStack.canGoBack}
            onClick={historyStack.goBack}
        >
            <Icon svgPath={mdiChevronLeft} aria-label="Back" />
        </Button>
        {historyStack.canGoForward && ( // No need to show the forward button if we can't go forward
            <Button variant="icon" className={styles.button} onClick={historyStack.goForward}>
                <Icon svgPath={mdiChevronRight} aria-label="Forward" />
            </Button>
        )}
    </nav>
)
