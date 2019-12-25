import React from 'react'
import H from 'history'
import { Section, SectionsProps, SectionsWithPersistenceProps } from '../Sections'
import { TabbedSectionsNavbar } from './TabbedSectionsNavbar'
import { Link } from '../../Link'
import { parseHash } from '../../../util/url'

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
export const TabBorderClassName = 'tabbed-sections-navbar__end-fragment-other-element'

/**
 * A UI component with a navbar for switching between sections and a content view that renders the
 * active section's contents.
 *
 * Callers should use one of the TabbedSectionsWithXyzViewStatePersistence components to handle view
 * state persistence.
 */
class TabbedSections<ID extends string, T extends Section<ID>> extends React.PureComponent<
    SectionsProps<ID, T> & TabbedSectionsProps
> {
    public render(): JSX.Element | null {
        let children: React.ReactElement<{ key: ID }>[] | undefined
        if (Array.isArray(this.props.children)) {
            children = this.props.children as React.ReactElement<{ key: ID }>[]
        } else if (this.props.children) {
            children = [this.props.children as React.ReactElement<{ key: ID }>]
        }

        return (
            <div id={this.props.id} className={`tabbed-sections ${this.props.className || ''}`}>
                <TabbedSectionsNavbar
                    sections={this.props.sections}
                    activeSection={this.props.activeSection}
                    navbarEndFragment={this.props.navbarEndFragment}
                    navbarItemClassName={this.props.navbarItemClassName}
                    navbarItemComponent={this.props.navbarItemComponent}
                />
                {this.props.toolbarFragment && (
                    <div className="tabbed-sections__toolbar small">{this.props.toolbarFragment}</div>
                )}
                {children?.find(c => c && c.key === this.props.activeSection)}
            </div>
        )
    }
}

/**
 * A wrapper for TabbedSections that persists view state (the currently active section) in
 * localStorage.
 */
export class TabbedSectionsWithLocalStorageViewStatePersistence<
    ID extends string,
    T extends Section<ID>
> extends React.PureComponent<
    SectionsWithPersistenceProps<ID, T> & {
        /**
         * A key unique to this UI element that is used for persisting the view state.
         */
        storageKey: string
    } & TabbedSectionsProps,
    { activeSection: ID | undefined }
> {
    constructor(props: SectionsProps<ID, T> & { storageKey: string }) {
        super(props)
        this.state = {
            activeSection: TabbedSectionsWithLocalStorageViewStatePersistence.readFromLocalStorage(
                this.props.storageKey,
                this.props.sections
            ),
        }
    }

    private static readFromLocalStorage<ID extends string, T extends Section<ID>>(
        storageKey: string,
        sections: T[]
    ): ID | undefined {
        const lastSectionID = localStorage.getItem(storageKey)
        if (lastSectionID !== null && sections.find(s => s.id === lastSectionID)) {
            return lastSectionID as ID
        }
        if (sections.length === 0) {
            return undefined
        }
        return sections[0].id // default
    }

    private static saveToLocalStorage<ID extends string>(storageKey: string, lastSectionID: ID): void {
        localStorage.setItem(storageKey, lastSectionID)
    }

    public render(): JSX.Element | null {
        return (
            <TabbedSections
                {...this.props}
                onSelectNavbarItem={this.onSelectSection}
                activeSection={this.state.activeSection}
                navbarItemComponent={this.renderNavbarItem}
            />
        )
    }

    private onSelectSection = (section: ID): void => {
        if (this.props.onSelectNavbarItem) {
            this.props.onSelectNavbarItem(section)
        }
        this.setState({ activeSection: section }, () =>
            TabbedSectionsWithLocalStorageViewStatePersistence.saveToLocalStorage(this.props.storageKey, section)
        )
    }

    private renderNavbarItem = ({ section, className }: { section: T; className: string }): JSX.Element => (
        <button
            type="button"
            className={className}
            data-e2e-section={section.id}
            onClick={() => this.onSelectSection(section.id)}
        >
            {section.label}
        </button>
    )
}

/**
 * A wrapper for TabbedSections that persists view state (the currently active section) in the
 * current page's URL.
 *
 * URL whose fragment (hash) ends with "$x" are considered to have active section "x" (where "x" is
 * the section's ID).
 */
export class TabbedSectionsWithURLViewStatePersistence<
    ID extends string,
    T extends Section<ID>
> extends React.PureComponent<
    SectionsWithPersistenceProps<ID, T> & { location: H.Location } & TabbedSectionsProps,
    { activeSection: ID | undefined }
> {
    constructor(props: SectionsWithPersistenceProps<ID, T> & { location: H.Location }) {
        super(props)
        this.state = {
            activeSection: TabbedSectionsWithURLViewStatePersistence.readFromURL(props.location, props.sections),
        }
    }

    /**
     * Returns the URL hash (which can be used as a relative URL) that specifies the given section.
     * If the URL hash already contains a section ID, it replaces it; otherwise it appends it to the
     * current URL fragment. If the sectionID argument is null, then the section ID is removed from
     * the URL.
     */
    public static urlForSectionID(location: H.Location, sectionID: string | null): H.LocationDescriptorObject {
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

    public static readFromURL<ID extends string, T extends Section<ID>>(
        location: H.Location,
        sections: T[]
    ): ID | undefined {
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

    public componentDidUpdate(prevProps: SectionsWithPersistenceProps<ID, T> & { location: H.Location }): void {
        if (prevProps.location !== this.props.location || prevProps.sections !== this.props.sections) {
            // eslint-disable-next-line react/no-did-update-set-state
            this.setState({
                activeSection: TabbedSectionsWithURLViewStatePersistence.readFromURL(
                    this.props.location,
                    this.props.sections
                ),
            })
        }
    }

    public render(): JSX.Element | null {
        return (
            <TabbedSections
                {...this.props}
                activeSection={this.state.activeSection}
                navbarItemComponent={this.renderNavbarItem}
            />
        )
    }

    private renderNavbarItem = ({ section, className }: { section: T; className: string }): JSX.Element => (
        /* eslint-disable react/jsx-no-bind */
        <Link
            className={className}
            to={TabbedSectionsWithURLViewStatePersistence.urlForSectionID(this.props.location, section.id)}
            onClick={() => {
                if (this.props.onSelectNavbarItem) {
                    this.props.onSelectNavbarItem(section.id)
                }
            }}
        >
            {section.label}
        </Link>
        /* eslint:enable react/jsx-no-bind */
    )
}
