import styles from './LoadingPage.module.css'

export const LoadingPage: React.FunctionComponent<{}> = () => (
    <div className="outer-container">
        <div className={styles.container}>
            <LoadingDots />
        </div>
    </div>
)

const LoadingDots: React.FunctionComponent = () => (
    <div className={styles.dotsHolder}>
        <div className={styles.dot} />
        <div className={styles.dot} />
        <div className={styles.dot} />
    </div>
)
