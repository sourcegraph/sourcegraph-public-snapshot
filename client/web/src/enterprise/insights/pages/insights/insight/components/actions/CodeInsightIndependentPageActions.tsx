import { FunctionComponent, useRef, useState } from 'react'

import LinkVariantIcon from 'mdi-react/LinkVariantIcon'
import { useHistory } from 'react-router'

import { Button, Link, Icon } from '@sourcegraph/wildcard'

import { ConfirmDeleteModal } from '../../../../../components/modals/ConfirmDeleteModal'
import { Insight } from '../../../../../core'
import { useCopyURLHandler } from '../../../../../hooks/use-copy-url-handler'

import styles from './CodeInsightIndependentPageActions.module.scss'

interface Props {
    insight: Pick<Insight, 'title' | 'id' | 'type'>
}

export const CodeInsightIndependentPageActions: FunctionComponent<Props> = props => {
    const { insight } = props

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

    return (
        <div className={styles.container}>
            <Button
                variant="secondary"
                ref={copyLinkButtonReference}
                data-tooltip={isCopied ? 'Copied!' : undefined}
                onClick={handleCopyLinkClick}
            >
                <Icon as={LinkVariantIcon} /> Copy link
            </Button>
            <Button variant="danger" onClick={handleDeleteClick}>
                Delete
            </Button>
            <Button variant="primary" as={Link} to={`/insights/edit/${insight.id}`}>
                Edit
            </Button>

            <ConfirmDeleteModal
                insight={insight}
                showModal={showDeleteConfirm}
                onConfirm={() => history.push('/insights/dashboards/all')}
                onCancel={() => setShowDeleteConfirm(false)}
            />
        </div>
    )
}
