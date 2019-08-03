import React from 'react'
import { CollapsibleSidebar } from '../collapsibleSidebar/CollapsibleSidebar'

export interface InfoSidebarSection {
    expanded:
        | React.ReactElement
        | {
              title: React.ReactFragment
              children?: React.ReactFragment
          }

    collapsed?:
        | React.ReactElement
        | {
              icon: React.ComponentType<{ className?: string }>
              tooltip: string
          }
}

interface Props {
    sections: InfoSidebarSection[]
    className?: string
}

/**
 * A sidebar that displays information about the current page (as opposed to navigation links).
 */
export const InfoSidebar: React.FunctionComponent<Props> = ({ sections, className = '' }) => (
    <CollapsibleSidebar
        localStorageKey="info-sidebar__sidebar"
        side="right"
        className={`info-sidebar d-flex flex-column border-left ${className}`}
        collapsedClassName="info-sidebar--collapsed"
        expandedClassName="info-sidebar--expanded"
    >
        {expanded => (
            <ul className="list-group list-group-flush px-3">
                {sections.map((section, i) => {
                    const Icon =
                        !section.collapsed || React.isValidElement(section.collapsed) ? null! : section.collapsed.icon
                    return (
                        <li key={i} className="list-group-item info-sidebar__item">
                            {expanded ? (
                                React.isValidElement(section.expanded) ? (
                                    section.expanded
                                ) : (
                                    <>
                                        <h6 className="font-weight-normal mb-0">{section.expanded.title}</h6>
                                        {section.expanded.children}
                                    </>
                                )
                            ) : !section.collapsed || React.isValidElement(section.collapsed) ? (
                                section.collapsed
                            ) : (
                                <Icon className="icon-inline" data-tooltip={section.collapsed.tooltip} />
                            )}
                        </li>
                    )
                })}
            </ul>
        )}
    </CollapsibleSidebar>
)
