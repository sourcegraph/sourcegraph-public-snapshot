import { Label } from '@sourcegraph/wildcard'

import styles from './IndexJobLabel.module.scss'

interface IndexJobLabelProps {
    label: string
}

export const IndexJobLabel: React.FunctionComponent<React.PropsWithChildren<IndexJobLabelProps>> = ({
    label,
    children,
}) => (
    <>
        <li className={styles.jobField}>
            <Label className={styles.jobLabel}>{label}:</Label>
            {children}
        </li>
    </>
)
