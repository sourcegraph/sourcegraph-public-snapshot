import { Label } from '@sourcegraph/wildcard'

// TODO: Own file
import styles from '../inference-script/InferenceScriptPreview.module.scss'

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
