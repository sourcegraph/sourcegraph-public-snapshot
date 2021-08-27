import classNames from 'classnames'
import H from 'history'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronUpIcon from 'mdi-react/ChevronUpIcon'
import MenuIcon from 'mdi-react/MenuIcon'
import React, { useEffect, useRef, useState } from 'react'
import { LinkProps, NavLink as RouterLink } from 'react-router-dom'

import styles from './EsbuildTest.module.scss'
import styles2 from './EsbuildTest2.module.scss'

export const EsbuildTest = (): void => console.log('esbuild test styles', styles, styles2)

export const EsbuildTest2 = () => <p>hello</p>

if (window.foobar) {
    console.log(`2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/SourcegraphWebApp.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/components/insights-view-grid/components/insight-card/components/insight-card-description/InsightCardDescription.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/components/insights-view-grid/components/insight-card/components/insight-card-menu/InsightCardMenu.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/components/insights-view-grid/components/insight-card/InsightCard.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/components/insights-view-grid/components/view-grid/ViewGrid.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/components/form/form-input/FormInput.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/pages/dashboards/creation/InsightsDashboardCreationPage.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/pages/dashboards/dashboard-page/components/add-insight-modal/AddInsightModal.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/pages/dashboards/dashboard-page/components/dashboard-select/components/trancated-text/TruncatedText.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/pages/dashboards/dashboard-page/components/dashboard-select/components/badge/Badge.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/pages/dashboards/dashboard-page/components/add-insight-modal/components/add-insight-modal-content/AddInsightModalContent.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/pages/dashboards/dashboard-page/components/dashboard-menu/DashboardMenu.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/pages/dashboards/dashboard-page/components/dashboard-select/components/menu-button/MenuButton.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/pages/dashboards/dashboard-page/components/dashboard-select/components/select-option/SelectOption.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/pages/dashboards/dashboard-page/components/dashboard-select/DashboardSelect.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/pages/dashboards/dashboard-page/components/delete-dashboard-modal/DeleteDashobardModal.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/components/alert-overlay/AlertOverlay.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/components/insights-view-grid/components/backend-insight/BackendInsight.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/components/form/repositories-field/components/flex-textarea/FlexTextarea.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/components/insights-view-grid/components/backend-insight/components/drill-down-filters-panel/components/drill-down-filters-form/components/drill-down-reg-exp-input/DrillDownRegExpInput.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/components/insights-view-grid/components/backend-insight/components/drill-down-filters-panel/DrillDownFiltersPanel.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/components/insights-view-grid/components/backend-insight/components/drill-down-filters-action/DrillDownFiltersPanel.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/pages/dashboards/dashboard-page/components/dashboards-content/components/empty-insight-dashboard/EmptyInsightDashboard.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/pages/dashboards/dashboard-page/components/dashboards-content/DashboardsContent.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/pages/dashboards/edit-dashboard/EditDashboardPage.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/components/form/repositories-field/components/suggestion-panel/SuggestionPanel.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/components/form/repositories-field/RepositoriesField.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/pages/insights/creation/lang-stats/components/lang-stats-insight-creation-form/LangStatsInsightCreationForm.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/components/live-preview-container/LivePreviewContainer.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/pages/insights/creation/lang-stats/components/lang-stats-insight-creation-content/LangStatsInsightCreationContent.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/pages/insights/creation/lang-stats/LangStatsInsightCreationPage.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/pages/insights/creation/search-insight/components/form-color-input/FormColorInput.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/pages/insights/creation/search-insight/components/form-series/components/series-card/SeriesCard.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/pages/insights/creation/search-insight/components/form-series/FormSeries.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/pages/insights/creation/search-insight/components/search-insight-creation-form/SearchInsightCreationForm.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/pages/insights/creation/search-insight/components/search-insight-creation-content/SearchInsightCreationContent.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/pages/insights/creation/search-insight/SearchInsightCreationPage.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/wildcard/src/components/PageHeader/PageHeader.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/wildcard/src/components/Container/Container.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/shared/src/components/Resizable.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/branded/src/components/panel/Panel.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/branded/src/components/panel/views/EmptyPanelView.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/branded/src/components/panel/views/HierarchicalLocationsView.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/branded/src/components/panel/views/HierarchicalLocationsViewButton.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/branded/src/components/panel/views/PanelView.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/components/Breadcrumbs.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/components/DismissibleAlert/DismissibleAlert.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/shared/src/components/EsbuildTest/EsbuildTest1.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/shared/src/components/EsbuildTest/EsbuildTest2.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/components/WebHoverOverlay/WebHoverOverlay.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/nav/IconRadioButtons.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/search/input/LazyMonacoQueryInput.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/search/input/SearchBox.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/search/FeatureTour.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/search/input/SearchContextCtaPrompt.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/wildcard/src/components/NavBar/NavAction.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/wildcard/src/components/NavBar/NavBar.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/wildcard/src/components/NavBar/NavItem.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/standalone/browser/standalone-tokens.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/base/browser/ui/aria/aria.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/browser/widget/media/editor.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/browser/controller/textAreaHandler.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/browser/viewParts/lineNumbers/lineNumbers.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/base/browser/ui/mouseCursor/mouseCursor.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/browser/viewParts/currentLineHighlight/currentLineHighlight.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/browser/viewParts/decorations/decorations.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/base/browser/ui/scrollbar/media/scrollbars.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/browser/viewParts/glyphMargin/glyphMargin.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/browser/viewParts/indentGuides/indentGuides.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/browser/viewParts/lines/viewLines.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/browser/viewParts/linesDecorations/linesDecorations.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/browser/viewParts/marginDecorations/marginDecorations.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/browser/viewParts/minimap/minimap.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/browser/viewParts/overlayWidgets/overlayWidgets.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/browser/viewParts/rulers/rulers.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/browser/viewParts/scrollDecoration/scrollDecoration.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/browser/viewParts/selections/selections.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/browser/viewParts/viewCursors/viewCursors.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/browser/widget/media/diffEditor.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/base/browser/ui/sash/sash.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/browser/widget/media/diffReview.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/base/browser/ui/actionbar/actionbar.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/platform/contextview/browser/contextMenuHandler.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/base/browser/ui/contextview/contextview.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/base/browser/ui/codicons/codicon/codicon.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/base/browser/ui/codicons/codicon/codicon-modifiers.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/base/browser/ui/list/list.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/base/browser/ui/tree/media/tree.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/base/browser/ui/table/table.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/base/browser/ui/splitview/splitview.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/standalone/browser/quickInput/standaloneQuickInput.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/base/parts/quickinput/browser/media/quickInput.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/base/browser/ui/iconLabel/iconlabel.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/base/browser/ui/keybindingLabel/keybindingLabel.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/base/browser/ui/inputbox/inputBox.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/base/browser/ui/countBadge/countBadge.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/base/browser/ui/progressbar/progressbar.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/base/browser/ui/button/button.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/contrib/anchorSelect/anchorSelect.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/contrib/bracketMatching/bracketMatching.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/contrib/message/messageController.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/contrib/codeAction/lightBulbWidget.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/contrib/codelens/codelensWidget.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/contrib/colorPicker/colorPicker.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/base/browser/ui/hover/hover.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/contrib/gotoError/media/gotoErrorWidget.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/contrib/peekView/media/peekViewWidget.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/contrib/zoneWidget/zoneWidget.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/platform/actions/browser/menuEntryActionViewItem.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/base/browser/ui/dropdown/dropdown.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/contrib/gotoSymbol/link/goToDefinitionAtPosition.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/contrib/gotoSymbol/peek/referencesWidget.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/contrib/dnd/dnd.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/base/browser/ui/checkbox/checkbox.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/contrib/find/findWidget.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/base/browser/ui/findinput/findInput.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/contrib/folding/folding.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/contrib/links/links.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/contrib/parameterHints/parameterHints.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/contrib/rename/renameInputField.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/contrib/snippet/snippetSession.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/contrib/suggest/media/suggest.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/standalone/browser/accessibilityHelp/accessibilityHelp.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/standalone/browser/iPadShowKeyboard/iPadShowKeyboard.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/node_modules/monaco-editor/esm/vs/editor/standalone/browser/inspectTokens/inspectTokens.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/extensions/ExtensionRegistrySidenav.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/extensions/ExtensionHeader.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/org/settings/OrgSettingsSidebar.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/pages/insights/creation/intro/IntroCreationPage.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/insights/pages/insights/edit-insight/EditInsightPage.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/search/results/ButtonDropdownCta.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/search/results/sidebar/SearchSidebarSection.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/search/results/sidebar/SearchReference.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/search/results/sidebar/SearchSidebar.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/search/results/StreamingSearchResults.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/repo/RepoRevisionSidebarCommits.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/search/notebook/SearchNotebook.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/search/notebook/SearchNotebookAddBlockButtons.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/search/notebook/SearchNotebookBlock.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/search/notebook/SearchNotebookBlockMenu.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/search/notebook/SearchNotebookMarkdownBlock.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/search/notebook/SearchNotebookQueryBlock.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/repo/blob/RenderedSearchNotebookMarkdown.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/search/home/CustomersSection.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/search/home/DynamicWebFonts/DynamicWebFonts.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/search/home/HeroSection.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/search/home/HomepageModalVideo.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/search/home/LoggedOutHomepage.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/search/home/SelfHostInstructions.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/search/home/SearchPageFooter.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/search/notebook/SearchNotebookPage.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/auth/CloudSignUpPage.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/auth/Steps/Steps.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/auth/Terminal/Terminal.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/wildcard/src/components/PageSelector/PageSelector.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/searchContexts/SearchContextForm.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/searchContexts/SearchContextOwnerDropdown.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/searchContexts/SearchContextPage.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/components/fuzzyFinder/HighlightedLink.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/components/fuzzyFinder/FuzzyModal.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/repo/actions/CopyPathAction.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/repo/compare/RepositoryCompareHeader.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/settings/tokens/AccessTokenNode.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/user/settings/auth/ExternalAccountsSignIn.module.css
2021/08/27 13:47:33 AAA ../../../../../../../../tmp/esbuild-6cag1b/client/web/src/user/settings/auth/UserSettingsPasswo`)
}

interface NavBarProps {
    children: React.ReactNode
    logo: React.ReactNode
}

interface NavGroupProps {
    children: React.ReactNode
}

interface NavItemProps {
    icon?: React.ComponentType<{ className?: string }>
    children: React.ReactNode
}

interface NavActionsProps {
    children: React.ReactNode
}

interface NavLinkProps extends NavItemProps, Pick<LinkProps<H.LocationState>, 'to'> {
    external?: boolean
}

const navActionStyles = {}
const navBarStyles = {}
const navItemStyles = {}

const useOutsideClickDetector = (
    reference: React.RefObject<HTMLDivElement>
): [boolean, React.Dispatch<React.SetStateAction<boolean>>] => {
    const [outsideClick, setOutsideClick] = useState(false)

    useEffect(() => {
        function handleClickOutside(event: MouseEvent): void {
            if (reference.current && !reference.current.contains(event.target as Node | null)) {
                setOutsideClick(false)
            }
        }
        document.addEventListener('mouseup', handleClickOutside)
        return () => {
            document.removeEventListener('mouseup', handleClickOutside)
        }
    }, [reference, setOutsideClick])

    return [outsideClick, setOutsideClick]
}

export const NavBar = ({ children, logo }: NavBarProps): JSX.Element => (
    <nav aria-label="Main Menu" className={navBarStyles.navbar}>
        <h1 className={navBarStyles.logo}>
            <RouterLink to="/search">{logo}</RouterLink>
        </h1>
        <hr className={navBarStyles.divider} />
        {children}
    </nav>
)

export const NavGroup = ({ children }: NavGroupProps): JSX.Element => {
    const menuReference = useRef<HTMLDivElement>(null)
    const [open, setOpen] = useOutsideClickDetector(menuReference)

    return (
        <div className={navBarStyles.menu} ref={menuReference}>
            <button
                className={classNames('btn', navBarStyles.menuButton)}
                type="button"
                onClick={() => setOpen(!open)}
                aria-label="Sections Navigation"
            >
                <MenuIcon className="icon-inline" />
                {!open ? <ChevronDownIcon className="icon-inline" /> : <ChevronUpIcon className="icon-inline" />}
            </button>
            <ul className={classNames(navBarStyles.list, { [navBarStyles.menuClose]: !open })}>{children}</ul>
        </div>
    )
}

export const NavActions: React.FunctionComponent<NavActionsProps> = ({ children }) => (
    <ul className={navActionStyles.actions}>{children}</ul>
)

export const NavAction: React.FunctionComponent<NavActionsProps> = ({ children }) => (
    <>
        {React.Children.map(children, action => (
            <li className={navActionStyles.action}>{action}</li>
        ))}
    </>
)

export const NavItem: React.FunctionComponent<NavItemProps> = ({ children, icon }) => {
    if (!children) {
        throw new Error('NavItem must be include at least one child')
    }

    return (
        <>
            {React.Children.map(children, child => (
                <li className={navItemStyles.item}>{React.cloneElement(child as React.ReactElement, { icon })}</li>
            ))}
        </>
    )
}

export const NavLink: React.FunctionComponent<NavLinkProps> = ({ icon: Icon, children, to, external }) => {
    const content = (
        <span className={navItemStyles.linkContent}>
            {Icon ? <Icon className={classNames('icon-inline', navItemStyles.icon)} /> : null}
            <span
                className={classNames(navItemStyles.text, {
                    [navItemStyles.iconIncluded]: Icon,
                })}
            >
                {children}
            </span>
        </span>
    )

    if (external) {
        return (
            <a href={to as string} rel="noreferrer noopener" target="_blank" className={navItemStyles.link}>
                {content}
            </a>
        )
    }

    return (
        <RouterLink to={to} className={navItemStyles.link} activeClassName={navItemStyles.active}>
            {content}
        </RouterLink>
    )
}
