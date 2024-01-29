import { AskCodyIcon } from '@sourcegraph/cody-ui/dist/icons/AskCodyIcon'
import { Button, LoadingSpinner, Tooltip } from '@sourcegraph/wildcard'

import styles from './AskCodyButton.module.scss'
import { FILE_CODY_IGNORE } from '../backend'
import { FileCodyIgnoreResult, FileCodyIgnoreVariables } from '../../graphql-operations'
import { useQuery } from '@sourcegraph/http-client'

export interface AskCodyButtonProps {
    onClick: () => void
    repoName: string,
    revision?: string,
    currentPath?: string,
}

export function AskCodyButton({ onClick, repoName, revision, currentPath }: AskCodyButtonProps): JSX.Element {

    const {
        data,
        loading,
     } = useQuery<FileCodyIgnoreResult, FileCodyIgnoreVariables>(
        FILE_CODY_IGNORE,
        {
            variables: {
                repo: repoName,
                revision: revision ?? '',
                currentPath: currentPath ?? '',
            }
        }
    )

    const allowed = data?.repository?.commit?.blob?.allowedForCodyContext ?? true
    const tooltip = allowed ? "Open Cody" : "Cody disabled for this file"
    const classNames = allowed ? [styles.codyButton] : [styles.codyButton, styles.codyButtonDisabled];

    if (loading) {
        return <LoadingSpinner />
    }

    return (
        <div className="d-flex align-items-center">
            <Tooltip content={tooltip} placement="bottom">
                <Button className={classNames.join(' ')} onClick={onClick} disabled={!allowed}>
                    <AskCodyIcon iconColor="#A112FF" /> Ask Cody
                </Button>
            </Tooltip>
        </div>
    )
}
