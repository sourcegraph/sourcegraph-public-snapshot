import { Button, Tooltip } from '@sourcegraph/wildcard'

import { AskCodyIcon } from './AskCodyIcon'

import styles from './AskCodyButton.module.scss'

export function AskCodyButton({ onClick }: { onClick: () => void }) {
    return (
        <Tooltip content="Open Cody" placement="bottom">
            <Button className={styles.codyButton} onClick={onClick}>
                <AskCodyIcon iconColor="#A112FF" /> Ask Cody
            </Button>
        </Tooltip>
    )
}
