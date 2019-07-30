import SyncIcon from 'mdi-react/SyncIcon'
import React from 'react'
import * as sourcegraph from 'sourcegraph'
import { displayRepoName } from '../../../../../../shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { isDefined } from '../../../../../../shared/src/util/types'
import { parseRepoURI } from '../../../../../../shared/src/util/url'
import { Timestamp } from '../../../../components/time/Timestamp'
import { ActionsIcon } from '../../../../util/octicons'
import { ChangesetPlanOperation } from '../../../changesetsOLD/plan/plan'
import { ThreadSettings } from '../../../threadsOLD/settings'

interface Props {
    campaign: Pick<GQL.ICampaign, 'isPreview' | 'rules'>

    className?: string
}

/**
 * A list of rules applied by a campaign.
 */
export const CampaignRulesList: React.FunctionComponent<Props> = ({ campaign, className = '' }) => {
    const rules: ChangesetPlanOperation[] = JSON.parse(campaign.rules || '[]')
    return (
        <div className={`campaign-rules-list ${className}`}>
            {rules && rules.length > 0 ? (
                <>
                    <div className="border border-success p-3 mb-4 d-flex align-items-stretch">
                        <SyncIcon className="flex-0 mr-2" />{' '}
                        <div className="flex-1">
                            <h4 className="mb-0">Continuously applying rules</h4>
                            <p className="mb-0">
                                The rules {campaign.isPreview ? 'will' : ''} run when any base branch changes or when a
                                new repository matches.
                            </p>
                        </div>
                    </div>
                    <ul className="list-group">
                        {rules.map((rule, i) => (
                            <li key={i} className="list-group-item d-flex align-items-start">
                                <ActionsIcon className="icon-inline small mr-2" />
                                <header>
                                    <h6 className="mb-0 font-size-base font-weight-normal mr-4">{rule.message}</h6>
                                </header>
                                <div className="flex-1"></div>
                                {rule.diagnostics && (
                                    <small className="text-muted mt-1">{diagnosticQueryLabel(rule.diagnostics)}</small>
                                )}
                            </li>
                        ))}
                    </ul>
                </>
            ) : (
                <span className="text-muted">No operations</span>
            )}
        </div>
    )
}

function diagnosticQueryLabel(query: sourcegraph.DiagnosticQuery): string {
    const parts = [
        query.type && `${query.type} diagnostics`,
        query.tag && `tagged with '${query.tag}'`,
        query.document &&
            query.document.length > 0 &&
            // TODO!(sqs): support other than just the 0th element
            query.document[0].pattern &&
            `in repository ${displayRepoName(parseRepoURI(query.document[0].pattern).repoName)} ${
                parseRepoURI(query.document[0].pattern).filePath
                    ? `file ${parseRepoURI(query.document[0].pattern).filePath}`
                    : ''
            }`,
    ]
        .filter(isDefined)
        .join(' ')
    if (parts === '') {
        return 'All diagnostics'
    }
    return `Fixes ${parts}`
}
