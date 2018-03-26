import * as React from 'react'

/**
 * Describes a tab.
 *
 * @template ID The type that includes all possible tab IDs (typically a union of string constants).
 */
export interface Tab<ID extends string> {
    id: ID
    label: string
}

/**
 * Properties for the tab bar.
 *
 * @template ID The type that includes all possible tab IDs (typically a union of string constants).
 * @template T The type that describes a tab.
 */
interface TabBarProps<ID extends string, T extends Tab<ID>> {
    /** All tabs. */
    tabs: T[]

    /** The currently active tab. */
    activeTab: ID

    /** Called when the user selects a different tab. */
    onSelect: (tab: ID) => void

    /** A fragment to render at the end of the tab bar. */
    endFragment?: React.ReactFragment

    tabClassName?: string
}

/**
 * A horizontal bar that displays tab titles, which the user can click to switch to the tab.
 *
 * @template ID The type that includes all possible tab IDs (typically a union of string constants).
 * @template T The type that describes a tab.
 */
class TabBar<ID extends string, T extends Tab<ID>> extends React.PureComponent<TabBarProps<ID, T>> {
    public render(): JSX.Element | null {
        return (
            <div className="tab-bar">
                {this.props.tabs.map((tab, i) => (
                    <button
                        key={i}
                        className={`btn btn-link btn-sm tab-bar__tab ${!this.props.endFragment &&
                            'tab-bar__tab--flex-grow'} tab-bar__tab--${
                            this.props.activeTab === tab.id ? 'active' : 'inactive'
                        } ${this.props.tabClassName || ''}`}
                        // tslint:disable-next-line:jsx-no-lambda
                        onClick={() => this.props.onSelect(tab.id)}
                    >
                        {tab.label}
                    </button>
                ))}
                {this.props.endFragment}
            </div>
        )
    }
}

/**
 * An element to pass to Tab's tabBarEndFragment prop to fill all width between the tabs (on the left) and the
 * other tabBarEndFragment elements (on the right).
 */
export const Spacer: () => JSX.Element = () => <span className="tab-bar__spacer" />

/**
 * Properties for the Tabs components and its wrappers.
 *
 * @template ID The type that includes all possible tab IDs (typically a union of string constants).
 * @template T The type that describes a tab.
 */
interface TabsProps<ID extends string, T extends Tab<ID>> {
    /** All tabs. Must always have at least one element. */
    tabs: T[]

    /**
     * A fragment to display at the end of the tab bar. If specified, the tabs will not flex grow to fill the
     * width.
     */
    tabBarEndFragment?: React.ReactFragment

    children: undefined | React.ReactElement<{ key: ID }> | (undefined | React.ReactElement<{ key: ID }>)[]

    id?: string
    className?: string
    tabClassName?: string

    /** Optional handler when a tab is selected */
    onSelectTab?: (tab: ID) => void
}

/**
 * A tabbed UI component, with a tab bar for switching between tabs and a content view that renders the active
 * tab's contents.
 *
 * Most callers should use one of the TabsWithXyzViewStatePersistence components to handle view state persistence.
 */
export class Tabs<ID extends string, T extends Tab<ID>> extends React.PureComponent<
    TabsProps<ID, T> & {
        /** The currently active tab. */
        activeTab: ID
    }
> {
    /**
     * The class name to use for other elements injected via tabBarEndFragment that should have a bottom border.
     */
    public static tabBorderClassName = 'tab-bar__end-fragment-other-element'

    public render(): JSX.Element | null {
        let children: React.ReactElement<{ key: ID }>[] | undefined
        if (Array.isArray(this.props.children)) {
            children = this.props.children as React.ReactElement<{ key: ID }>[]
        } else if (this.props.children) {
            children = [this.props.children as React.ReactElement<{ key: ID }>]
        }

        return (
            <div id={this.props.id} className={`tabs ${this.props.className || ''}`}>
                <TabBar
                    tabs={this.props.tabs}
                    activeTab={this.props.activeTab}
                    onSelect={this.onSelectTab}
                    endFragment={this.props.tabBarEndFragment}
                    tabClassName={this.props.tabClassName}
                />
                {children && children.find(c => c && c.key === this.props.activeTab)}
            </div>
        )
    }

    private onSelectTab = (tab: ID) => {
        if (this.props.onSelectTab) {
            this.props.onSelectTab(tab)
        }
    }
}

/**
 * A wrapper for Tabs that persists view state (the currently active tab) in localStorage.
 */
export class TabsWithLocalStorageViewStatePersistence<ID extends string, T extends Tab<ID>> extends React.PureComponent<
    TabsProps<ID, T> & {
        /**
         * A key unique to this UI element that is used for persisting the view state.
         */
        storageKey?: string
    },
    { activeTab: ID }
> {
    constructor(props: TabsProps<ID, T>) {
        super(props)
        this.state = {
            activeTab:
                this.props.storageKey !== undefined
                    ? TabsWithLocalStorageViewStatePersistence.readFromLocalStorage(
                          this.props.storageKey,
                          this.props.tabs
                      )
                    : this.props.tabs[0].id,
        }
    }

    private static readFromLocalStorage<ID extends string, T extends Tab<ID>>(storageKey: string, tabs: T[]): ID {
        const lastTabID = localStorage.getItem(storageKey)
        if (lastTabID !== null && tabs.find(tab => tab.id === lastTabID)) {
            return lastTabID as ID
        }
        return tabs[0].id // default
    }

    private static saveToLocalStorage<ID extends string>(storageKey: string, lastTabID: ID): void {
        localStorage.setItem(storageKey, lastTabID)
    }

    public render(): JSX.Element | null {
        return <Tabs {...this.props} onSelectTab={this.onSelectTab} activeTab={this.state.activeTab} />
    }

    private onSelectTab = (tab: ID) => {
        if (this.props.onSelectTab) {
            this.props.onSelectTab(tab)
        }
        this.setState({ activeTab: tab }, () => {
            if (this.props.storageKey !== undefined) {
                TabsWithLocalStorageViewStatePersistence.saveToLocalStorage(this.props.storageKey, tab)
            }
        })
    }
}
