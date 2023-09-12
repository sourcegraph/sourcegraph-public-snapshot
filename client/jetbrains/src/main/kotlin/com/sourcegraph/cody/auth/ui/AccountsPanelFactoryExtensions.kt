package com.sourcegraph.cody.auth.ui

import com.intellij.collaboration.messages.CollaborationToolsBundle
import com.intellij.collaboration.ui.util.JListHoveredRowMaterialiser
import com.intellij.icons.AllIcons
import com.intellij.openapi.Disposable
import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.actionSystem.CommonShortcuts
import com.intellij.openapi.keymap.KeymapUtil
import com.intellij.ui.LayeredIcon
import com.intellij.ui.SimpleTextAttributes
import com.intellij.ui.ToolbarDecorator
import com.intellij.ui.awt.RelativePoint
import com.intellij.ui.components.JBList
import com.intellij.ui.dsl.builder.Cell
import com.intellij.ui.dsl.builder.Row
import com.intellij.util.ui.EmptyIcon
import com.intellij.util.ui.StatusText
import com.intellij.util.ui.UIUtil
import com.sourcegraph.cody.auth.Account
import com.sourcegraph.cody.auth.AccountManager
import com.sourcegraph.cody.auth.AccountsListener
import com.sourcegraph.cody.auth.PersistentDefaultAccountHolder
import java.awt.event.MouseAdapter
import java.awt.event.MouseEvent
import java.awt.event.MouseListener
import javax.swing.Icon
import javax.swing.JComponent
import javax.swing.JMenuItem
import javax.swing.JPopupMenu
import javax.swing.ListCellRenderer
import javax.swing.ListSelectionModel
import javax.swing.SwingUtilities

/**
 * Custom factory method to create and account panel with possibility to add mouse listener for list
 * elements. This is a custom version of a
 * [com.intellij.collaboration.auth.ui.AccountsPanelFactory.create]
 */
private fun <A : Account, Cred, R> create(
    model: AccountsListModel<A, Cred>,
    needAddBtnWithDropdown: Boolean,
    listElementMouseListener: (JBList<A>, AccountsListModel<A, Cred>) -> MouseListener,
    listCellRendererFactory: () -> R
): JComponent where R : ListCellRenderer<A>, R : JComponent {

  val accountsListModel = model.accountsListModel
  val accountsList =
      JBList(accountsListModel).apply {
        val decoratorRenderer = listCellRendererFactory()
        cellRenderer = decoratorRenderer
        JListHoveredRowMaterialiser.install(this, listCellRendererFactory())
        UIUtil.putClientProperty(
            this, UIUtil.NOT_IN_HIERARCHY_COMPONENTS, listOf(decoratorRenderer))

        selectionMode = ListSelectionModel.SINGLE_SELECTION
        addMouseListener(listElementMouseListener(this, model))
      }
  model.busyStateModel.addListener { accountsList.setPaintBusy(it) }
  accountsList.addListSelectionListener { model.selectedAccount = accountsList.selectedValue }

  accountsList.emptyText.apply {
    appendText(CollaborationToolsBundle.message("accounts.none.added"))
    appendSecondaryText(
        CollaborationToolsBundle.message("accounts.add.link"),
        SimpleTextAttributes.LINK_PLAIN_ATTRIBUTES) {
          val event = it.source
          val relativePoint = if (event is MouseEvent) RelativePoint(event) else null
          model.addAccount(accountsList, relativePoint)
        }
    appendSecondaryText(
        " (${KeymapUtil.getFirstKeyboardShortcutText(CommonShortcuts.getNew())})",
        StatusText.DEFAULT_ATTRIBUTES,
        null)
  }

  model.busyStateModel.addListener { accountsList.setPaintBusy(it) }

  val addIcon: Icon =
      if (needAddBtnWithDropdown) LayeredIcon.ADD_WITH_DROPDOWN else AllIcons.General.Add

  val toolbar =
      ToolbarDecorator.createDecorator(accountsList)
          .disableUpDownActions()
          .setAddAction { model.addAccount(accountsList, it.preferredPopupPoint) }
          .setAddIcon(addIcon)

  if (model is AccountsListModel.WithDefault) {
    toolbar.addExtraAction(
        object :
            ToolbarDecorator.ElementActionButton(
                CollaborationToolsBundle.message("accounts.set.default"),
                AllIcons.Actions.Checked) {
          override fun actionPerformed(e: AnActionEvent) {
            val selected = accountsList.selectedValue
            if (selected == model.defaultAccount) return
            if (selected != null) model.defaultAccount = selected
          }

          override fun updateButton(e: AnActionEvent) {
            isEnabled = isEnabled && model.defaultAccount != accountsList.selectedValue
          }
        })
  }

  return toolbar.createPanel()
}

/**
 * Accounts panel with context menu actions displayed when right-clicked on menu elements This is a
 * custom version of a [com.intellij.collaboration.auth.ui.AccountsPanelFactory.accountsPanel]
 */
fun <A : Account, Cred> Row.customAccountsPanel(
    accountManager: AccountManager<A, Cred>,
    defaultAccountHolder: PersistentDefaultAccountHolder<A>,
    accountsModel: AccountsListModel.WithDefault<A, Cred>,
    detailsProvider: AccountsDetailsProvider<A, *>,
    disposable: Disposable,
    needAddBtnWithDropdown: Boolean,
    defaultAvatarIcon: Icon = EmptyIcon.ICON_16,
    copyAccount: (A) -> A = { it },
): Cell<JComponent> {

  accountsModel.addCredentialsChangeListener(detailsProvider::reset)
  detailsProvider.loadingStateModel.addListener { accountsModel.busyStateModel.value = it }

  fun isModified() =
      accountsModel.newCredentials.isNotEmpty() ||
          accountsModel.accounts != accountManager.accounts ||
          accountsModel.defaultAccount != defaultAccountHolder.account

  fun reset() {
    val defaultAccount = defaultAccountHolder.account
    val accountsWithoutDefault =
        if (defaultAccount != null) accountManager.accounts - defaultAccount
        else accountManager.accounts
    if (defaultAccount != null) {
      val defaultAccountCopy = copyAccount(defaultAccount)
      accountsModel.accounts =
          (accountsWithoutDefault.map(copyAccount) + defaultAccountCopy).toSet()
      accountsModel.defaultAccount = defaultAccountCopy
    } else {
      accountsModel.accounts = accountsWithoutDefault.map(copyAccount).toSet()
      accountsModel.defaultAccount = null
    }

    accountsModel.clearNewCredentials()
    detailsProvider.resetAll()
  }

  fun apply() {
    for ((account, token) in accountsModel.newCredentials) {
      if (token != null) {
        accountManager.updateAccount(account, token)
      }
    }
    val newTokensMap = accountsModel.accounts.associateWith { null }
    accountManager.updateAccounts(newTokensMap)
    val defaultAccount = accountsModel.defaultAccount
    defaultAccountHolder.account = defaultAccount
    accountsModel.clearNewCredentials()
  }

  accountManager.addListener(
      disposable,
      object : AccountsListener<A> {
        override fun onAccountCredentialsChanged(account: A) {
          if (!isModified()) reset()
        }
      })

  val component =
      create(
          accountsModel,
          needAddBtnWithDropdown,
          { list, model ->
            object : MouseAdapter() {
              override fun mouseReleased(e: MouseEvent) {
                if (SwingUtilities.isRightMouseButton(e)) {
                  list.selectedIndex = list.locationToIndex(e.point)
                  list.setSelectedValue(list.model.getElementAt(list.selectedIndex), true)

                  val menu = JPopupMenu()
                  val editAccount = JMenuItem("Edit Account")
                  editAccount.addActionListener { model.editAccount(list, list.selectedValue) }
                  menu.add(editAccount)
                  menu.show(list, e.getPoint().x, e.getPoint().y)
                }
              }
            }
          }) {
            SimpleAccountsListCellRenderer(accountsModel, detailsProvider, defaultAvatarIcon)
          }
  return cell(component).onIsModified(::isModified).onReset(::reset).onApply(::apply)
}
