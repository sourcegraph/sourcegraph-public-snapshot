import { ApolloError } from '@apollo/client'

import { Button, ErrorAlert } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../components/LoaderButton'

import styles from './BlockPackageModal.module.scss'

interface BlockPackageActionsProps {
    submitText: string
    valid: boolean
    loading: boolean
    error?: ApolloError
    onDismiss: () => void
}

export const BlockPackageActions: React.FunctionComponent<BlockPackageActionsProps> = ({
    submitText,
    valid,
    loading,
    error,
    onDismiss,
}) => (
    <div className={styles.actionsContainer}>
        {error && <ErrorAlert className="mb-3" error={error} />}
        <div className={styles.actions}>
            <Button variant="secondary" onClick={onDismiss} className="mr-2">
                Cancel
            </Button>
            <LoaderButton label={submitText} variant="danger" type="submit" disabled={!valid} loading={loading} />
        </div>
    </div>
)
