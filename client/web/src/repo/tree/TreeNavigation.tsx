import React from 'react'

import { mdiSourceCommit, mdiSourceBranch, mdiTag, mdiHistory, mdiAccount, mdiBrain, mdiCog } from '@mdi/js'

import { encodeURIPathComponent } from '@sourcegraph/common'
import { TreeFields } from '@sourcegraph/shared/src/graphql-operations'
import { Button, ButtonGroup, Icon, Link } from '@sourcegraph/wildcard'

import { RepoBatchChangesButton } from '../../batches/RepoBatchChangesButton'
import { TreePageRepositoryFields } from '../../graphql-operations'

interface TreeNavigationProps {
    repo: TreePageRepositoryFields
    revision: string
    tree: TreeFields
    codeIntelligenceEnabled: boolean
    batchChangesEnabled: boolean
}

export const TreeNavigation: React.FunctionComponent<React.PropsWithChildren<TreeNavigationProps>> = ({
    repo,
    revision,
    tree,
    codeIntelligenceEnabled,
    batchChangesEnabled,
}) => (
    <ButtonGroup>
        <Button to={`${tree.url}/-/commits`} variant="secondary" outline={true} as={Link}>
            <Icon aria-hidden={true} svgPath={mdiSourceCommit} /> Commits
        </Button>
        <Button to={`/${encodeURIPathComponent(repo.name)}/-/branches`} variant="secondary" outline={true} as={Link}>
            <Icon aria-hidden={true} svgPath={mdiSourceBranch} /> Branches
        </Button>
        <Button to={`/${encodeURIPathComponent(repo.name)}/-/tags`} variant="secondary" outline={true} as={Link}>
            <Icon aria-hidden={true} svgPath={mdiTag} /> Tags
        </Button>
        <Button
            to={
                revision
                    ? `/${encodeURIPathComponent(repo.name)}/-/compare/...${encodeURIComponent(revision)}`
                    : `/${encodeURIPathComponent(repo.name)}/-/compare`
            }
            variant="secondary"
            outline={true}
            as={Link}
        >
            <Icon aria-hidden={true} svgPath={mdiHistory} /> Compare
        </Button>
        <Button
            to={`/${encodeURIPathComponent(repo.name)}/-/stats/contributors`}
            variant="secondary"
            outline={true}
            as={Link}
        >
            <Icon aria-hidden={true} svgPath={mdiAccount} /> Contributors
        </Button>
        {codeIntelligenceEnabled && (
            <Button
                to={`/${encodeURIPathComponent(repo.name)}/-/code-graph`}
                variant="secondary"
                outline={true}
                as={Link}
            >
                <Icon aria-hidden={true} svgPath={mdiBrain} /> Code Graph
            </Button>
        )}
        {batchChangesEnabled && <RepoBatchChangesButton repoName={repo.name} />}
        {repo.viewerCanAdminister && (
            <Button
                to={`/${encodeURIPathComponent(repo.name)}/-/settings`}
                variant="secondary"
                outline={true}
                as={Link}
            >
                <Icon aria-hidden={true} svgPath={mdiCog} /> Settings
            </Button>
        )}
    </ButtonGroup>
)
