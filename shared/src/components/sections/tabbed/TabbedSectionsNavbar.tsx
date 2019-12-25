import React from 'react'
import { Section, SectionsProps } from '../Sections'
import { TabbedSectionsProps } from './TabbedSections'

/**
 * Properties for the tab bar.
 *
 * @template ID The type that includes all possible tab IDs (typically a union of string constants).
 * @template T The type that describes a tab.
 */
interface Props<ID extends string, T extends Section<ID>>
    extends Pick<SectionsProps<ID, T>, 'sections' | 'activeSection' | 'navbarItemComponent' | 'navbarItemClassName'>,
        TabbedSectionsProps {}

/**
 * A horizontal bar that displays tab titles, which the user can click to switch to the tab.
 *
 * @template ID The type that includes all possible tab IDs (typically a union of string constants).
 * @template T The type that describes a tab.
 */
export class TabbedSectionsNavbar<ID extends string, T extends Section<ID>> extends React.PureComponent<Props<ID, T>> {
    public render(): JSX.Element | null {
        return (
            <div
                className={`tabbed-sections-navbar ${
                    this.props.sections.length === 0 ? 'tabbed-sections-navbar--empty' : ''
                }`}
            >
                {this.props.sections
                    .filter(({ hidden }) => !hidden)
                    .map(section => (
                        <this.props.navbarItemComponent
                            key={section.id}
                            section={section}
                            className={`btn btn-link btn-sm tabbed-sections-navbar__tab ${!this.props
                                .navbarEndFragment &&
                                'tabbed-sections-navbar__tab--flex-grow'} tabbed-sections-navbar__tab--${
                                this.props.activeSection !== undefined && this.props.activeSection === section.id
                                    ? 'active'
                                    : 'inactive'
                            } ${this.props.navbarItemClassName || ''}`}
                        />
                    ))}
                {this.props.navbarEndFragment}
            </div>
        )
    }
}

/**
 * An element to pass to TabbedSections's navbarEndFragment prop to fill all width between the tabs
 * (on the left) and the other navbarEndFragment elements (on the right).
 */
export const Spacer: () => JSX.Element = () => <span className="tabbed-sections-navbar__spacer" />
