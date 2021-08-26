import classNames from 'classnames'
import React from 'react'

export interface CodeIntelligenceConfigurationTabHeaderProps {
    selectedTab: SelectedTab
    setSelectedTab: (selectedTab: SelectedTab) => void
    indexingEnabled: boolean
}

export type SelectedTab = 'globalPolicies' | 'repositoryPolicies' | 'indexConfiguration'

export const CodeIntelligenceConfigurationTabHeader: React.FunctionComponent<CodeIntelligenceConfigurationTabHeaderProps> = ({
    selectedTab,
    setSelectedTab,
    indexingEnabled,
}) => {
    const onClick = (selected: SelectedTab): React.MouseEventHandler => event => {
        event.preventDefault()
        setSelectedTab(selected)
    }

    return (
        <div className="overflow-auto mb-2">
            <ul className="nav nav-tabs d-inline-flex d-sm-flex flex-nowrap text-nowrap">
                <li className="nav-item">
                    {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                    <a
                        href=""
                        onClick={onClick('globalPolicies')}
                        className={classNames('nav-link', selectedTab === 'globalPolicies' && 'active')}
                        role="button"
                    >
                        <span className="text-content" data-tab-content="Global policies">
                            Global policies
                        </span>
                    </a>
                </li>
                <li className="nav-item">
                    {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                    <a
                        href=""
                        onClick={onClick('repositoryPolicies')}
                        className={classNames('nav-link', selectedTab === 'repositoryPolicies' && 'active')}
                        role="button"
                    >
                        <span className="text-content" data-tab-content="Repository-specific policies">
                            Repository-specific policies
                        </span>
                    </a>
                </li>
                {indexingEnabled && (
                    <li className="nav-item">
                        {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                        <a
                            href=""
                            onClick={onClick('indexConfiguration')}
                            className={classNames('nav-link', selectedTab === 'indexConfiguration' && 'active')}
                            role="button"
                        >
                            <span className="text-content" data-tab-content="Index configuration">
                                Index configuration
                            </span>
                        </a>
                    </li>
                )}
            </ul>
        </div>
    )
}
