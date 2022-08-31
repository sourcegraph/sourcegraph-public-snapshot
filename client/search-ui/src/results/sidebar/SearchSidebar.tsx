import {
    ComponentProps,
    createContext,
    FC,
    HTMLAttributes,
    PropsWithChildren,
    useCallback,
    useContext,
    useMemo,
} from 'react'

import { mdiClose } from '@mdi/js'
import classNames from 'classnames'
import { noop } from 'lodash'
import StickyBox from 'react-sticky-box'

import { SectionID } from '@sourcegraph/shared/src/settings/temporary/searchSidebar'
import { TemporarySettings } from '@sourcegraph/shared/src/settings/temporary/TemporarySettings'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { Button, Icon } from '@sourcegraph/wildcard'

import { SearchFilterSection } from './SearchFilterSection'

import styles from './SearchSidebar.module.scss'

interface SearchSidebarStore {
    collapsedSections?: { [key in SectionID]?: boolean }
    persistToggleState: (id: string, open: boolean) => void
}

const SearchSidebarContext = createContext<SearchSidebarStore>({
    collapsedSections: {},
    persistToggleState: noop,
})

interface SearchSidebarProps extends HTMLAttributes<HTMLElement> {
    onClose: () => void
}

/**
 * Styled sticky sidebar UI component. Internally it uses sticky box
 * lib component so sticky logic is different compared to standard position
 * sticky behavior.
 *
 * Also provides shared through context internal state for compound SearchSidebarSection
 * components.
 */
export const SearchSidebar: FC<PropsWithChildren<SearchSidebarProps>> = props => {
    const { children, className, onClose, ...attributes } = props
    const [collapsedSections, setCollapsedSections] = useTemporarySetting('search.collapsedSidebarSections', {})

    const persistToggleState = useCallback(
        (id: string, open: boolean) => {
            setCollapsedSections(openSections => {
                const newSettings: TemporarySettings['search.collapsedSidebarSections'] = {
                    ...openSections,
                    [id]: !open,
                }
                return newSettings
            })
        },
        [setCollapsedSections]
    )

    const sidebarStore = useMemo(() => ({ collapsedSections, persistToggleState }), [
        collapsedSections,
        persistToggleState,
    ])

    return (
        <aside
            {...attributes}
            className={classNames(styles.sidebar, className)}
            role="region"
            aria-label="Search sidebar"
        >
            <StickyBox className={styles.stickyBox} offsetTop={8}>
                <div className={styles.header}>
                    <Button variant="icon" onClick={onClose}>
                        <Icon svgPath={mdiClose} aria-label="Close sidebar" />
                    </Button>
                </div>

                {
                    // collapsedSections is undefined on first render. To prevent the sections
                    // being rendered open and immediately closing them, we render them only after
                    // we got the settings.
                    collapsedSections && (
                        <SearchSidebarContext.Provider value={sidebarStore}>{children}</SearchSidebarContext.Provider>
                    )
                }
            </StickyBox>
        </aside>
    )
}

interface SearchSidebarSectionProps
    extends Omit<ComponentProps<typeof SearchFilterSection>, 'startCollapsed' | 'onToggle'> {
    sectionId: SectionID
}

/**
 * Provides a collapsable section UI which is connected to SidePanel component
 * and persist expand/collapse state with temporal settings.
 */
export const SearchSidebarSection: FC<SearchSidebarSectionProps> = props => {
    const { className, sectionId, ...attributes } = props
    const { collapsedSections, persistToggleState } = useContext(SearchSidebarContext)

    return (
        <SearchFilterSection
            sectionId={sectionId}
            startCollapsed={collapsedSections?.[sectionId]}
            onToggle={persistToggleState}
            className={classNames(className, styles.item)}
            {...attributes}
        />
    )
}
