import { PackageMatchBehaviour } from '@sourcegraph/shared/src/graphql-operations'
import { Label, Select } from '@sourcegraph/wildcard'

import styles from './BehaviourSelect.module.scss'

interface BehaviourSelectProps {
    value: PackageMatchBehaviour
    onChange: (behaviour: PackageMatchBehaviour) => void
}

export const BehaviourSelect: React.FunctionComponent<BehaviourSelectProps> = ({ value, onChange }) => (
    <>
        <Label className="mb-2" id="behaviour-type">
            Behavior
        </Label>
        <div className={styles.container}>
            <Select
                className={styles.select}
                aria-labelledby="behaviour-type"
                value={value}
                onChange={event => onChange(event.target.value as PackageMatchBehaviour)}
            >
                <option value={PackageMatchBehaviour.BLOCK}>Blocklist</option>
                <option value={PackageMatchBehaviour.ALLOW}>Allowlist</option>
            </Select>
            <small className="ml-3 text-muted">
                <>
                    {value === PackageMatchBehaviour.BLOCK
                        ? 'Blocklisting will remove any matching packages from this instance and prevent them from being synced in future.'
                        : 'Allowlisting will remove any unmatched packages within this ecosystem from this instance and prevent them from being synced in future.'}
                </>
            </small>
        </div>
    </>
)
