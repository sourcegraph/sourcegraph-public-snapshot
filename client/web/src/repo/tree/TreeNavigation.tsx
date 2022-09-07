import React from 'react'

import { mdiSourceCommit, mdiSourceBranch, mdiTag, mdiHistory, mdiAccount, mdiBrain, mdiCog } from '@mdi/js'

import { encodeURIPathComponent } from '@sourcegraph/common'
import { TreeFields } from '@sourcegraph/shared/src/graphql-operations'
import { Button, ButtonGroup, Icon, Link } from '@sourcegraph/wildcard'

import { RepoBatchChangesButton } from '../../batches/RepoBatchChangesButton'

interface TreeNavigationProps {
    repoName: string
    viewerCanAdminister: boolean | undefined
    revision: string
    tree: TreeFields
    codeIntelligenceEnabled: boolean
    batchChangesEnabled: boolean
}

export const TreeNavigation: React.FunctionComponent<React.PropsWithChildren<TreeNavigationProps>> = ({
    repoName,
    viewerCanAdminister,
    revision,
    tree,
    codeIntelligenceEnabled,
    batchChangesEnabled,
}) => (
    <ButtonGroup>
        <Button to={`${tree.url}/-/commits`} variant="secondary" outline={true} as={Link}>
            <Icon aria-hidden={true} svgPath={mdiSourceCommit} /> Commits
        </Button>
        <Button to={`/${encodeURIPathComponent(repoName)}/-/branches`} variant="secondary" outline={true} as={Link}>
            <Icon aria-hidden={true} svgPath={mdiSourceBranch} /> Branches
        </Button>
        <Button to={`/${encodeURIPathComponent(repoName)}/-/tags`} variant="secondary" outline={true} as={Link}>
            <Icon aria-hidden={true} svgPath={mdiTag} /> Tags
        </Button>
        <Button
            to={
                revision
                    ? `/${encodeURIPathComponent(repoName)}/-/compare/...${encodeURIComponent(revision)}`
                    : `/${encodeURIPathComponent(repoName)}/-/compare`
            }
            variant="secondary"
            outline={true}
            as={Link}
        >
            <Icon aria-hidden={true} svgPath={mdiHistory} /> Compare
        </Button>
        <Button
            to={`/${encodeURIPathComponent(repoName)}/-/stats/contributors`}
            variant="secondary"
            outline={true}
            as={Link}
        >
            <Icon aria-hidden={true} svgPath={mdiAccount} /> Contributors
        </Button>
        {codeIntelligenceEnabled && (
            <Button
                to={`/${encodeURIPathComponent(repoName)}/-/code-graph`}
                variant="secondary"
                outline={true}
                as={Link}
            >
                <Icon aria-hidden={true} svgPath={mdiBrain} /> Code Graph data
            </Button>
        )}
        {batchChangesEnabled && <RepoBatchChangesButton repoName={repoName} />}
        {viewerCanAdminister && (
            <Button to={`/${encodeURIPathComponent(repoName)}/-/settings`} variant="secondary" outline={true} as={Link}>
                <Icon aria-hidden={true} svgPath={mdiCog} /> Settings
            </Button>
        )}
    </ButtonGroup>
)
