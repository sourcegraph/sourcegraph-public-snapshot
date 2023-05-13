import { ContextScopePicker } from './components/ContextScopePicker'

import styles from './ContextScope.module.scss'

interface ContextScopeProps {}

export const ContextScope: React.FunctionComponent<ContextScopeProps> = ({}) => {
    return (
        <div className={styles.wrapper}>
            <div className={styles.title}>Context scope</div>
            <ContextSeparator />
            <ContextScopePicker />
            <ContextSeparator />
        </div>
    )
}

const ContextSeparator = () => <div className={styles.separator} />
