import React from 'react'

import { mdiMapSearch, mdiOpenInNew } from '@mdi/js'

import { Link, Text, Icon, H4 } from '@sourcegraph/wildcard'

import { CreatePolicyButtons } from './CreatePolicyButtons'

import styles from './EmptyPoliciesList.module.scss'

interface EmptyPoliciesListProps {
    showCta?: boolean
    repo?: { id: string; name: string }
}

export const EmptyPoliciesList: React.FunctionComponent<EmptyPoliciesListProps> = ({ repo, showCta }) => (
    <div className="d-flex align-items-center flex-column w-100 mt-3" data-testid="summary">
        <Icon className="mb-2" svgPath={mdiMapSearch} inline={false} aria-hidden={true} />
        <H4 className="mb-0">No policies found.</H4>
        <div className={showCta ? styles.contentExpanded : styles.content}>
            <div className="text-center p-3">
                <Text weight="medium" className="mb-2">
                    Documentation
                </Text>
                <Link
                    className="d-block mb-1"
                    to={`/help/code_navigation/how-to/configure_data_retention${
                        !repo
                            ? '#applying-data-retention-policies-globally'
                            : '#applying-data-retention-policies-to-a-specific-repository'
                    }`}
                    target="_blank"
                    rel="noreferrer noopener"
                >
                    Data retention policies <Icon svgPath={mdiOpenInNew} aria-hidden={true} />
                </Link>
                <Link
                    className="d-block"
                    to={`/help/code_navigation/how-to/configure_auto_indexing${
                        !repo
                            ? '#applying-indexing-policies-globally'
                            : '#applying-indexing-policies-to-a-specific-repository'
                    }}`}
                    target="_blank"
                    rel="noreferrer noopener"
                >
                    Auto-indexing policies <Icon svgPath={mdiOpenInNew} aria-hidden={true} />
                </Link>
            </div>
            {showCta && (
                <>
                    <div className={styles.divider} />
                    <div className="d-flex align-items-center justify-content-center p-3">
                        <CreatePolicyButtons repo={repo} />
                    </div>
                </>
            )}
        </div>
    </div>
)
