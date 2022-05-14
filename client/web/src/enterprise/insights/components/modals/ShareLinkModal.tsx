import { MouseEventHandler } from 'react'

import classNames from 'classnames'
import copy from 'copy-to-clipboard'
import { noop } from 'lodash'

import { Badge, Button, Input, Modal } from '@sourcegraph/wildcard'
import type { ModalProps } from '@sourcegraph/wildcard/out/src/components/Modal'

import { Insight, InsightDashboard } from '../../core'

import styles from './ShareLinkModal.module.scss'

type ShareLinkModalProps = ModalProps & {
    insight: Insight
    dashboard: InsightDashboard | null
    onDismiss(event: React.MouseEvent | React.KeyboardEvent): void
}

export const ShareLinkModal: React.FunctionComponent<ShareLinkModalProps> = ({
    insight,
    dashboard,
    isOpen,
    onDismiss,
    ...attributes
}) => {
    const shareableUrl = `${window.location.origin}/insights/insight/${insight.id}`

    const handleClick: MouseEventHandler<HTMLButtonElement> = () => {
        copy(shareableUrl)
    }

    return (
        <Modal className={classNames(styles.container)} {...attributes} isOpen={isOpen} onDismiss={onDismiss}>
            <h3>Get shareable link</h3>

            <ShareLinkModalContent insight={insight} dashboard={dashboard} />

            <Input
                value={shareableUrl}
                className={styles.input}
                inputSymbol={
                    <Button variant="primary" onClick={handleClick}>
                        Copy link
                    </Button>
                }
                onChange={noop}
            />
        </Modal>
    )
}

const ShareLinkModalContent: React.FunctionComponent<Pick<ShareLinkModalProps, 'insight' | 'dashboard'>> = ({
    insight,
    dashboard,
}) => {
    if (!dashboard) {
        return <GlobalContent />
    }

    const currentDashboard = insight.dashboards?.nodes.find(insightDashboard => insightDashboard.id === dashboard.id)
    const permission = getLeastRestrictivePermissions(currentDashboard)

    if (permission === 'organization') {
        const organizations = currentDashboard?.grants.organizations.map(organization => organization) as string[]
        return <OrganizationContent organizations={organizations} />
    }

    if (permission === 'private') {
        return <PrivateContent />
    }

    return <GlobalContent />
}

const GlobalContent: React.FunctionComponent = () => <>Everyone on your Sourcegraph instance can see this insight.</>

const PrivateContent: React.FunctionComponent = () => (
    <>
        <p>
            Only you can see this insight, because it's only on private dashboards. Add this insight to public
            dashboards to share with others.
        </p>
        <p>
            The “all insights” dashboard shows you all insights available to you regardless of their visibility to
            others.
        </p>
    </>
)

const OrganizationContent: React.FunctionComponent<{ organizations: string[] }> = ({ organizations }) => (
    <>
        <p>Only people added to following Organisations can see this insight:</p>
        {organizations.map(organization => (
            <Badge variant="secondary" key={organization} className="mr-2">
                {organization}
            </Badge>
        ))}
    </>
)

type ShareablePermission = 'private' | 'organization' | 'global'

export const getLeastRestrictivePermissions = (dashboard?: Insight['dashboards'][0]): ShareablePermission => {
    if (!dashboard) {
        return 'global'
    }

    if (dashboard.grants.global) {
        return 'global'
    }

    if (dashboard.grants.organizations.length > 0) {
        return 'organization'
    }

    if (dashboard.grants.users.length > 0) {
        return 'private'
    }

    return 'global'
}
