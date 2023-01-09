import { FunctionComponent, useRef, useState } from 'react'

import { mdiLinkVariant } from '@mdi/js'
import { useHistory } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Link, Icon, Tooltip } from '@sourcegraph/wildcard'

import { ConfirmDeleteModal } from '../../../../../components/modals/ConfirmDeleteModal'
import { Insight } from '../../../../../core'
import { useCopyURLHandler } from '../../../../../hooks/use-copy-url-handler'

import styles from './CodeInsightIndependentPageActions.module.scss'

interface Props extends TelemetryProps {
    insight: Pick<Insight, 'title' | 'id' | 'type'>
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
