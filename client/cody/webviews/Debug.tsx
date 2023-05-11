import styles from './Debug.module.css'

interface DebugProps {
    debugLog: string[]
}

export const Debug: React.FunctionComponent<React.PropsWithChildren<DebugProps>> = ({ debugLog }) => (
    <div className="inner-container">
        <div className="non-transcript-container">
            <div className={styles.debugContainer} data-tab-target="debug">
                {debugLog?.map((log, i) => (
                    // eslint-disable-next-line react/no-array-index-key
                    <div key={i} className={styles.debugMessage}>
                        {log}
                    </div>
                ))}
            </div>
        </div>
    </div>
)
