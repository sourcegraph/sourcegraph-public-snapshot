import { AskCodyIcon } from '@sourcegraph/cody-ui/src/icons/AskCodyIcon'
import { Button, Tooltip } from '@sourcegraph/wildcard'

import styles from './AskCodyButton.module.scss'

export function AskCodyButton({ onClick }: { onClick: () => void }): JSX.Element {
    return (
        <Tooltip content="Open Cody" placement="bottom">
            <Button className={styles.codyButton} onClick={onClick}>
                <AskCodyIcon iconColor="#A112FF" /> Ask Cody
            </Button>
        </Tooltip>
    )
}
