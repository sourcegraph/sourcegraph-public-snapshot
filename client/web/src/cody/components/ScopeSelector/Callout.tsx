import { mdiClose } from '@mdi/js'
import classNames from 'classnames'

import { Button, Icon, Text } from '@sourcegraph/wildcard'

import styles from './Callout.module.scss'

export const Callout: React.FC<{ dismiss: () => void }> = ({ dismiss }) => (
    <div className={styles.wrapper}>
        <div className={styles.box}>
            <div className={styles.header}>
                <div className={styles.headerElements}>Give Cody context</div>
                <Button className={styles.closeButton} onClick={dismiss} variant="icon" aria-label="Close">
                    <Icon aria-hidden={true} svgPath={mdiClose} />
                </Button>
            </div>
            <Text className={classNames('mb-0 mt-1', styles.content)} size="small">
                Tell Cody what codebases it should reference to help with your task and Cody will respond more
                accurately.
            </Text>
        </div>
        <div className={styles.tail} />
    </div>
)
