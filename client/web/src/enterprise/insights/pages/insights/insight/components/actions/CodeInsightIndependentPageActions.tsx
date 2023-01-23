import { FunctionComponent, useRef, useState } from 'react'

import { mdiLinkVariant } from '@mdi/js'
import { useHistory } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Link, Icon, Tooltip } from '@sourcegraph/wildcard'

import { ConfirmDeleteModal } from '../../../../../components/modals/ConfirmDeleteModal'
import { Insight, isLangStatsInsight } from '../../../../../core'
import { useCopyURLHandler } from '../../../../../hooks/use-copy-url-handler'

import styles from './CodeInsightIndependentPageActions.module.scss'

interface Props extends TelemetryProps {
    insight: Insight
}

export const CodeInsightIndependentPageActions: FunctionComponent<Props> = props => {
    const { insight, telemetryService } = props

    const history = useHistory()

    const copyLinkButtonReference = useRef<HTMLButtonElement | null>(null)
    const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
    const [copyURL, isCopied] = useCopyURLHandler()

    const handleCopyLinkClick = (): void => {
        copyURL()

        // Re-trigger trigger tooltip event catching logic to activate
        // copied tooltip appearance
        requestAnimationFrame(() => {
            copyLinkButtonReference.current?.blur()
            copyLinkButtonReference.current?.focus()
        })
    }

    const handleDeleteClick = (): void => {
        setShowDeleteConfirm(true)
    }

    const handleEditClick = (): void => {
        telemetryService.log('StandaloneInsightPageEditClick')
    }

    return (
        <div className={styles.container}>
            {!isLangStatsInsight(insight) && (
                <Tooltip content="This will create a CVS archive of all data for this Code Insight, including data that has been archived.This will only include data that you are permitted to see.">
                    <Button as="a" href={`/.api/insights/export/${insight.id}`} download={true} variant="secondary">
                        Export data as CSV
                    </Button>
                </Tooltip>
            )}

            <Tooltip content={isCopied ? 'Copied!' : undefined}>
                <Button variant="secondary" ref={copyLinkButtonReference} onClick={handleCopyLinkClick}>
                    <Icon aria-hidden={true} svgPath={mdiLinkVariant} /> Copy link
                </Button>
            </Tooltip>
            <Button variant="danger" onClick={handleDeleteClick}>
                Delete
            </Button>
            <Button
                variant="primary"
                as={Link}
                to={`/insights/edit/${insight.id}?insight=${insight.id}`}
                onClick={handleEditClick}
            >
                Edit
            </Button>

            <ConfirmDeleteModal
                insight={insight}
                showModal={showDeleteConfirm}
                onConfirm={() => history.push('/insights/all')}
                onCancel={() => setShowDeleteConfirm(false)}
            />
        </div>
    )
}
