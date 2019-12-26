import React, { useCallback, useMemo } from 'react'
import H from 'history'
import { Section, SectionsProps, SectionsWithPersistenceProps } from '../Sections'
import { Link } from '../../Link'
import { parseHash } from '../../../util/url'
import useLocalStorage from 'react-use-localstorage'

/**
 * Properties for TabbedSections.
 */
export interface TabbedSectionsProps {
    /**
     * A fragment to display at the end of the navbar. If specified, the navbar items will not flex
     * grow to fill the main axis.
     */
    navbarEndFragment?: React.ReactFragment

    /**
     * A fragment to display underneath the sections.
     */
    toolbarFragment?: React.ReactFragment
}

/**
 * The class name to use for other elements injected via navbarEndFragment that should have a bottom
 * border.
 */
export const TabBorderClassName = 'tabbed-sections__navbar-end-fragment-other-element'

/**
 * A UI component with a navbar for switching between sections and a content view that renders the
 * active section's contents.
 *
 * Callers should use one of the TabbedSectionsWithXyzViewStatePersistence components to handle view
 * state persistence.
 *
 * @template ID The type that includes all possible tab IDs (typically a union of string constants).
 * @template T The type that describes a tab.
 */
export const TabbedSections = <ID extends string, T extends Section<ID>>({
    sections,
    visibleSections,
    navbarItemComponent: NavbarItemComponent,
    navbarItemClassName,
    toolbarFragment,
    navbarEndFragment,
    id,
    className = '',
    children,
}: SectionsProps<ID, T> &
    TabbedSectionsProps & {
        /**
         * The component used to render a section's navbar item.
         */
        navbarItemComponent: React.ComponentType<{ section: T; className: string }>
    }): JSX.Element => {
    const childrenArray = Array.isArray(children)
        ? (children as React.ReactElement<{ key: ID }>[])
        : [children as React.ReactElement<{ key: ID }>]

    return (
        <div id={id} className={`tabbed-sections ${className}`}>
            <div className={`tabbed-sections__navbar ${sections.length === 0 ? 'tabbed-sections__navbar--empty' : ''}`}>
                {sections
                    .filter(({ hidden }) => !hidden)
                    .map(section => (
                        <NavbarItemComponent
                            key={section.id}
                            section={section}
                            className={`btn btn-link btn-sm tabbed-sections__navbar-tab ${!navbarEndFragment &&
                                'tabbed-sections__navbar-tab--flex-grow'} tabbed-sections__navbar-tab--${
                                visibleSections?.includes(section.id) ? 'active' : 'inactive'
                            } ${navbarItemClassName}`}
                        />
                    ))}
                {navbarEndFragment}
            </div>
            {toolbarFragment && <div className="tabbed-sections__toolbar small">{toolbarFragment}</div>}
            {childrenArray?.find(c => c && visibleSections?.includes(c.key as ID))}
        </div>
    )
}

/**
 * An element to pass to TabbedSections's navbarEndFragment prop to fill all width between the tabs
 * (on the left) and the other navbarEndFragment elements (on the right).
 */
export const Spacer: () => JSX.Element = () => <span className="tabbed-sections__navbar-spacer" />

/**
 * A wrapper for TabbedSections that persists view state (the currently active section) in
 * localStorage.
 */
export const TabbedSectionsWithLocalStorageViewStatePersistence = <ID extends string, T extends Section<ID>>({
    sections,
    storageKey,
    onSelectNavbarItem: parentOnSelectNavbarItem,
    ...props
}: SectionsWithPersistenceProps<ID, T> & {
    /**
     * A key unique to this UI element that is used for persisting the view state.
     */
    storageKey: string
} & TabbedSectionsProps): JSX.Element => {
    const [visibleSection, setVisibleSection] = useLocalStorage(`TabbedSections.${storageKey}`, sections[0]?.id || '')

    const onSelectNavbarItem = useCallback(
        (section: ID): void => {
            if (parentOnSelectNavbarItem) {
                parentOnSelectNavbarItem(section)
            }
            setVisibleSection(section)
        },
        [parentOnSelectNavbarItem, setVisibleSection]
    )

    const renderNavbarItem = useCallback(
        ({ section, className }: { section: T; className: string }): JSX.Element => (
            <button
                type="button"
                className={className}
                data-e2e-section={section.id}
                onClick={() => onSelectNavbarItem(section.id)}
            >
                {section.label}
            </button>
        ),
        [onSelectNavbarItem]
    )
    return (
        <TabbedSections
            {...props}
            sections={sections}
            visibleSections={visibleSection ? [visibleSection as ID] : undefined}
            navbarItemComponent={renderNavbarItem}
        />
    )
}

/**
 * Returns the current section based on the URL when using
 * {@link TabbedSectionsWithURLViewStatePersistence}.
 */
export const currentSectionForTabbedSectionsWithURLViewStatePersistence = <ID extends string>(
    sections: Section<ID>[],
    location: H.Location
): ID | undefined => {
    const urlSectionID = parseHash(location.hash).viewState
    if (urlSectionID) {
        for (const section of sections) {
            if (section.id === urlSectionID) {
                return section.id
            }
        }
    }
    if (sections.length === 0) {
        return undefined
    }
    return sections[0].id // default
}

/**
 * Returns the URL hash (which can be used as a relative URL) that specifies the given section (when
 * using {@link TabbedSectionsWithURLViewStatePersistence}). If the URL hash already contains a
 * section ID, it replaces it; otherwise it appends it to the current URL fragment. If the sectionID
 * argument is null, then the section ID is removed from the URL.
 */
export const urlForSectionForTabbedSectionsWithURLViewStatePersistence = <ID extends string>(
    sectionID: ID | null,
    location: H.Location
): H.Location => {
    const hash = new URLSearchParams(location.hash.slice('#'.length))
    if (sectionID) {
        hash.set('tab', sectionID)
    } else {
        hash.delete('tab')

        // Remove other known keys that are only associated with a panel. This makes it so the URL
        // is nicer when the panel is closed (it is stripped of all irrelevant panel hash state).
        //
        // TODO: Un-hardcode these so that other panels don't need to remember to add their keys
        // here.
        hash.delete('threadID')
        hash.delete('commentID')
    }
    return {
        ...location,
        hash: hash
            .toString()
            .replace(/%3A/g, ':')
            .replace(/=$/, ''), // remove needless trailing `=` as in `#L12=`,
    }
}

/**
 * A wrapper for TabbedSections that persists view state (the currently active section) in the
 * current page's URL.
 *
 * URL whose fragment (hash) ends with "$x" are considered to have active section "x" (where "x" is
 * the section's ID).
 */
export const TabbedSectionsWithURLViewStatePersistence = <ID extends string, T extends Section<ID>>({
    sections,
    location,
    onSelectNavbarItem: parentOnSelectNavbarItem,
    ...props
}: SectionsWithPersistenceProps<ID, T> & { location: H.Location } & TabbedSectionsProps): JSX.Element => {
    const visibleSection = useMemo<ID | undefined>(
        () => currentSectionForTabbedSectionsWithURLViewStatePersistence(sections, location),
        [sections, location]
    )

    const urlForSectionID = useCallback(
        (sectionID: ID): H.Location => urlForSectionForTabbedSectionsWithURLViewStatePersistence(sectionID, location),
        [location]
    )

    const renderNavbarItem = useCallback(
        ({ section, className }: { section: T; className: string }): JSX.Element => (
            /* eslint-disable react/jsx-no-bind */
            <Link
                className={className}
                to={urlForSectionID(section.id)}
                onClick={() => {
                    if (parentOnSelectNavbarItem) {
                        parentOnSelectNavbarItem(section.id)
                    }
                }}
            >
                {section.label}
            </Link>
            /* eslint:enable react/jsx-no-bind */
        ),
        [parentOnSelectNavbarItem, urlForSectionID]
    )

    return (
        <TabbedSections
            {...props}
            sections={sections}
            visibleSections={visibleSection === undefined ? undefined : [visibleSection]}
            navbarItemComponent={renderNavbarItem}
        />
    )
}
