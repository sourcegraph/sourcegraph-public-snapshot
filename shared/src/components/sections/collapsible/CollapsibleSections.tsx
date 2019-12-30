import React, { useCallback } from 'react'
import EyeOffIcon from 'mdi-react/EyeOffIcon'
import ViewGridIcon from 'mdi-react/ViewGridIcon'
import useLocalStorage from 'react-use-localstorage'
import { Section, SectionsProps, SectionsWithPersistenceProps } from '../Sections'
import { tryCatch } from '../../../util/errors'

/**
 * A UI component with a series of sections and a navbar for expanding/collapsing the sections. Any
 * number of sections may be expanded at a time.
 *
 * Callers should use one of the CollapsibleSectionsWithXyzViewStatePersistence components to handle
 * view state persistence.
 */
export const CollapsibleSections = <ID extends string, T extends Section<ID>>({
    sections,
    visibleSections,
    navbarItemComponent: NavbarItemComponent,
    navbarItemClassName,
    id,
    className = '',
    children,
}: SectionsProps<ID, T> & {
    /**
     * The component used to render a section's navbar item.
     */
    navbarItemComponent: React.ComponentType<{ section: T; className: string; isExpanded: boolean }>
}): JSX.Element => {
    const childrenArray = Array.isArray(children)
        ? (children as React.ReactElement<{ key: ID }>[])
        : [children as React.ReactElement<{ key: ID }>]

    return (
        <div id={id} className={`collapsible-sections ${className || ''}`}>
            <nav className="collapsible-sections__nav">
                {sections
                    .filter(({ hidden }) => !hidden)
                    .map(section => (
                        <NavbarItemComponent
                            key={section.id}
                            section={section}
                            isExpanded={Boolean(visibleSections?.includes(section.id))}
                            className={`btn btn-link text-left collapsible-sections__section-header ${navbarItemClassName ||
                                ''} collapsible-sections__section-header--${
                                visibleSections?.includes(section.id) ? 'expanded' : 'collapsed'
                            }`}
                        />
                    ))}
            </nav>
            <div className="collapsible-sections__content">
                {sections
                    .filter(({ hidden }) => !hidden)
                    .map(section => (
                        <div key={section.id} className="collapsible-sections__section-content">
                            {visibleSections?.includes(section.id) && childrenArray.find(c => c.key === section.id)}
                        </div>
                    ))}
            </div>
        </div>
    )
}

/**
 * A wrapper for CollapsibleSections that persists view state (the currently active section) in
 * localStorage.
 */
export const CollapsibleSectionsWithLocalStorageViewStatePersistence = <ID extends string, T extends Section<ID>>({
    storageKey,
    onSelectNavbarItem: parentOnSelectNavbarItem,
    ...props
}: SectionsWithPersistenceProps<ID, T> & {
    /**
     * A key unique to this UI element that is used for persisting the view state.
     */
    storageKey: string
}): JSX.Element => {
    const [visibleSectionsJSON, setVisibleSectionsJSON] = useLocalStorage(`CollapsibleSections.${storageKey}`, '[]')
    const parsedVisibleSections = tryCatch(() => JSON.parse(visibleSectionsJSON))
    const visibleSections =
        Array.isArray(parsedVisibleSections) && parsedVisibleSections.every(s => typeof s === 'string')
            ? parsedVisibleSections
            : []

    const onSelectNavbarItem = useCallback(
        (section: ID): void => {
            if (parentOnSelectNavbarItem) {
                parentOnSelectNavbarItem(section)
            }
            const isSectionVisible = visibleSections.includes(section)
            setVisibleSectionsJSON(
                JSON.stringify(
                    isSectionVisible ? visibleSections.filter(s => s !== section) : [...visibleSections, section]
                )
            )
        },
        [parentOnSelectNavbarItem, setVisibleSectionsJSON, visibleSections]
    )

    const renderNavbarItem = useCallback(
        ({ section, className, isExpanded }: { section: T; className: string; isExpanded: boolean }): JSX.Element => (
            <button
                type="button"
                className={className}
                data-e2e-section={section.id}
                onClick={() => onSelectNavbarItem(section.id)}
            >
                {isExpanded ? (
                    <EyeOffIcon className="icon-inline mr-2" />
                ) : (
                    <ViewGridIcon className="icon-inline mr-2" />
                )}
                {section.label}
            </button>
        ),
        [onSelectNavbarItem]
    )

    return <CollapsibleSections {...props} visibleSections={visibleSections} navbarItemComponent={renderNavbarItem} />
}
