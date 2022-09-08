/**
 * This module with stores is used to separate global/shared state from our component hierarchy.
 *
 * We want to keep shared state and this model in a single place (/stores directory).
 * If you need to share data across various components in the application, create Zustand store
 * and update this module instead of creating your own store in consumer code.
 *
 * Note: We are in the process of migrating shared data to this module, so not
 * everything that should be in here is already in here.
 */

export {
    useNavbarQueryState,
    setQueryStateFromURL,
    setQueryStateFromSettings,
    setSearchPatternType,
    setSearchCaseSensitivity,
    buildSearchURLQueryFromQueryState,
} from './navbarSearchQueryState'
export {
    useExperimentalFeatures,
    getExperimentalFeatures,
    setExperimentalFeaturesFromSettings,
} from './experimentalFeatures'
export { useNotepadState, useNotepad } from './notepad'
