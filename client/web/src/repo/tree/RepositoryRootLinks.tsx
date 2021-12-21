import AccountIcon from 'mdi-react/AccountIcon'
import BookOpenBlankVariantIcon from 'mdi-react/BookOpenBlankVariantIcon'
import BrainIcon from 'mdi-react/BrainIcon'
import HistoryIcon from 'mdi-react/HistoryIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import SourceCommitIcon from 'mdi-react/SourceCommitIcon'
import TagIcon from 'mdi-react/TagIcon'
import React from 'react'
import { Link } from 'react-router-dom'

import { encodeURIPathComponent, RevisionSpec } from '@sourcegraph/shared/src/util/url'
import { Button } from '@sourcegraph/wildcard'

import { BatchChangesProps } from '../../batches'
import { RepoBatchChangesButton } from '../../batches/RepoBatchChangesButton'
import { CodeIntelligenceProps } from '../../codeintel'
import { TreeFields, TreePageRepositoryFields } from '../../graphql-operations'
import { useExperimentalFeatures } from '../../stores'

interface Props extends RevisionSpec, CodeIntelligenceProps, Pick<BatchChangesProps, 'batchChangesEnabled'> {
    repo: TreePageRepositoryFields
    tree: TreeFields
}

export const RepositoryRootLinks: React.FunctionComponent<Props> = ({
    repo,
    tree,
    revision,
    codeIntelligenceEnabled,
    batchChangesEnabled,
}) => {
    // eslint-disable-next-line unicorn/prevent-abbreviations
    const enableAPIDocs = useExperimentalFeatures(features => features.apiDocs)

    return (
        <div className="btn-group">
            {enableAPIDocs && (
                <Button to={`${tree.url}/-/docs`} variant="secondary" outline={true} as={Link}>
                    <BookOpenBlankVariantIcon className="icon-inline" /> API docs
                </Button>
            )}
            <Button to={`${tree.url}/-/commits`} variant="secondary" outline={true} as={Link}>
                <SourceCommitIcon className="icon-inline" /> Commits
            </Button>
            <Button
                to={`/${encodeURIPathComponent(repo.name)}/-/branches`}
                variant="secondary"
                outline={true}
                as={Link}
            >
                <SourceBranchIcon className="icon-inline" /> Branches
            </Button>
            <Button to={`/${encodeURIPathComponent(repo.name)}/-/tags`} variant="secondary" outline={true} as={Link}>
                <TagIcon className="icon-inline" /> Tags
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
                <HistoryIcon className="icon-inline" /> Compare
            </Button>
            <Button
                to={`/${encodeURIPathComponent(repo.name)}/-/stats/contributors`}
                variant="secondary"
                outline={true}
                as={Link}
            >
                <AccountIcon className="icon-inline" /> Contributors
            </Button>
            {codeIntelligenceEnabled && (
                <Button
                    to={`/${encodeURIPathComponent(repo.name)}/-/code-intelligence`}
                    variant="secondary"
                    outline={true}
                    as={Link}
                >
                    <BrainIcon className="icon-inline" /> Code Intelligence
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
                    <SettingsIcon className="icon-inline" /> Settings
                </Button>
            )}
        </div>
    )
}
