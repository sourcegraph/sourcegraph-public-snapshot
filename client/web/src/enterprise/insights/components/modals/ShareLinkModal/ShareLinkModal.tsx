import { FunctionComponent, MouseEventHandler, useRef } from 'react'

import { useQuery } from '@apollo/client'
import classNames from 'classnames'
import { noop } from 'lodash'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Badge, Button, Input, LoadingSpinner, Modal, ModalProps, Typography } from '@sourcegraph/wildcard'

import { GetSharableInsightInfoResult } from '../../../../../graphql-operations'
import {
    CustomInsightDashboard,
    Insight,
    InsightDashboard,
    isGlobalDashboard,
    isOrganizationDashboard,
    isOrganizationOwner,
    isVirtualDashboard,
} from '../../../core'
import { useCopyURLHandler } from '../../../hooks/use-copy-url-handler'

import { decodeDashboardIds, GET_SHARABLE_INSIGHT_INFO_GQL } from './get-sharable-insight-info'

import styles from './ShareLinkModal.module.scss'

type ShareLinkModalProps = ModalProps & {
    insight: Insight
    dashboards: InsightDashboard[]
    onDismiss: () => void
}

export const ShareLinkModal: FunctionComponent<ShareLinkModalProps> = props => {
    const { insight, dashboards, isOpen, onDismiss, ...attributes } = props

    const shareableUrl = `${window.location.origin}/insights/insight/${insight.id}`
    const copyButtonReference = useRef<HTMLButtonElement>(null)
    const [copyURL, isCopied] = useCopyURLHandler()

    const handleClick: MouseEventHandler<HTMLButtonElement> = () => {
        copyURL(shareableUrl)

        // Re-trigger trigger tooltip event catching logic to activate
        // link copied tooltip appearance
        requestAnimationFrame(() => {
            copyButtonReference.current?.blur()
            copyButtonReference.current?.focus()
        })
    }

    return (
        <Modal className={classNames(styles.container)} {...attributes} isOpen={isOpen} onDismiss={onDismiss}>
            <Typography.H3>Get shareable link</Typography.H3>

            <ShareLinkModalContent insight={insight} dashboards={dashboards} />

            <Input
                value={shareableUrl}
                className={styles.input}
                inputSymbol={
                    <Button
                        ref={copyButtonReference}
                        variant="primary"
                        data-tooltip={isCopied ? 'Link copied' : undefined}
                        data-placement="bottom"
                        onClick={handleClick}
                    >
                        Copy link
                    </Button>
                }
                onChange={noop}
            />
        </Modal>
    )
}

const ShareLinkModalContent: FunctionComponent<Pick<ShareLinkModalProps, 'insight' | 'dashboards'>> = props => {
    const { insight, dashboards } = props

    const { data, error, loading } = useQuery<GetSharableInsightInfoResult>(GET_SHARABLE_INSIGHT_INFO_GQL, {
        variables: { id: insight.id },
    })

    if (loading) {
        return <LoadingSpinner />
    }

    if (error || !data) {
        return <ErrorAlert error={error} />
    }

    const insightDashboardIds = decodeDashboardIds(data)
    const insightDashboards = dashboards.filter(
        (dashboard): dashboard is CustomInsightDashboard =>
            !isVirtualDashboard(dashboard) && insightDashboardIds.includes(dashboard.id)
    )

    const permission = getLeastRestrictivePermissions(insightDashboards)

    if (permission === ShareablePermission.Organization) {
        const organizations = new Set(
            insightDashboards
                .filter(isOrganizationDashboard)
                .flatMap(dashboards => dashboards.owners.filter(isOrganizationOwner).map(owner => owner.title))
        )

        return <OrganizationContent organizations={[...organizations]} />
    }

    if (permission === ShareablePermission.Private) {
        return <PrivateContent />
    }

    return <GlobalContent />
}

const GlobalContent: FunctionComponent = () => <>Everyone on your Sourcegraph instance can see this insight.</>

const PrivateContent: FunctionComponent = () => (
    <>
        <p>
            Only you can see this insight, because it's only on private dashboards. Add this insight to public
            dashboards to share with others.
        </p>
        <p>
            <em>
                The “all insights” dashboard shows you all insights available to you regardless of their visibility to
                others.
            </em>
        </p>
    </>
)

const OrganizationContent: FunctionComponent<{ organizations: string[] }> = ({ organizations }) => (
    <>
        <p className="mb-2">Only people added to following Organizations can see this insight:</p>
        {organizations.map(organization => (
            <Badge variant="secondary" key={organization} className="mr-2">
                {organization}
            </Badge>
        ))}
    </>
)

enum ShareablePermission {
    Private,
    Organization,
    Global,
}

const getLeastRestrictivePermissions = (dashboards: CustomInsightDashboard[]): ShareablePermission => {
    if (dashboards.length === 0) {
        return ShareablePermission.Private
    }

    if (dashboards.some(isGlobalDashboard)) {
        return ShareablePermission.Global
    }

    if (dashboards.some(isOrganizationDashboard)) {
        return ShareablePermission.Organization
    }

    return ShareablePermission.Private
}
