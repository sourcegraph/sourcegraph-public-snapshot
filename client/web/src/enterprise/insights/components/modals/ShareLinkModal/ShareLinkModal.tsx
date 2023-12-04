import { type FC, type MouseEventHandler, useRef } from 'react'

import classNames from 'classnames'
import { noop } from 'lodash'

import {
    Badge,
    Button,
    Input,
    LoadingSpinner,
    Modal,
    type ModalProps,
    H3,
    Text,
    Tooltip,
    ErrorAlert,
} from '@sourcegraph/wildcard'

import {
    type CustomInsightDashboard,
    type Insight,
    isGlobalDashboard,
    isOrganizationDashboard,
    isOrganizationOwner,
} from '../../../core'
import { useCopyURLHandler } from '../../../hooks'

import { useDashboardThatHaveInsight } from './get-sharable-insight-info'

import styles from './ShareLinkModal.module.scss'

type ShareLinkModalProps = ModalProps & {
    insight: Insight
    onDismiss: () => void
}

export const ShareLinkModal: FC<ShareLinkModalProps> = props => {
    const { insight, isOpen, onDismiss, ...attributes } = props

    const shareableUrl = `${window.location.origin}/insights/${insight.id}`
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
            <H3>Get shareable link</H3>

            <ShareLinkModalContent insight={insight} />

            <Input
                value={shareableUrl}
                className={styles.input}
                inputSymbol={
                    <Tooltip content={isCopied ? 'Link copied' : undefined}>
                        <Button
                            ref={copyButtonReference}
                            variant="primary"
                            data-placement="bottom"
                            onClick={handleClick}
                        >
                            Copy link
                        </Button>
                    </Tooltip>
                }
                onChange={noop}
            />
        </Modal>
    )
}

interface ShareLinkModalContentProps {
    insight: Insight
}

const ShareLinkModalContent: FC<ShareLinkModalContentProps> = props => {
    const { insight } = props
    const { dashboards, error, loading } = useDashboardThatHaveInsight({ insightId: insight.id })

    if (loading) {
        return <LoadingSpinner />
    }

    if (error || !dashboards) {
        return <ErrorAlert error={error} />
    }

    const permission = getLeastRestrictivePermissions(dashboards)

    if (permission === ShareablePermission.Organization) {
        const organizations = new Set(
            dashboards
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

const PrivateContent: FC = () => (
    <span>
        <Text>
            Only you can see this insight, because it's only on private dashboards. Add this insight to public
            dashboards to share with others.
        </Text>
        <Text>
            <em>
                The “all insights” dashboard shows you all insights available to you regardless of their visibility to
                others.
            </em>
        </Text>
    </span>
)

const OrganizationContent: FC<{ organizations: string[] }> = ({ organizations }) => (
    <span>
        <Text className="mb-2">Only people added to the following organizations can see this insight:</Text>
        {organizations.map(organization => (
            <Badge variant="secondary" key={organization} className="mr-2">
                {organization}
            </Badge>
        ))}
    </span>
)

const GlobalContent: FC = () => <>Everyone on your Sourcegraph instance can see this insight.</>

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
