package com.sourcegraph.cody.config

import com.intellij.ide.DataManager
import com.intellij.openapi.actionSystem.ActionGroup
import com.intellij.openapi.actionSystem.ActionManager
import com.intellij.openapi.actionSystem.DataKey
import com.intellij.openapi.ui.popup.JBPopupFactory
import com.intellij.ui.components.DropDownLink
import javax.swing.JButton

interface CodyAccountsHost {
  fun addAccount(server: SourcegraphServerPath, login: String, token: String)
  fun addAccount(account: CodyAccount, token: String)
  fun isAccountUnique(login: String, server: SourcegraphServerPath): Boolean

  companion object {
    val KEY: DataKey<CodyAccountsHost> = DataKey.create("SourcegraphAccountsHots")

    fun createAddAccountLink(): JButton =
        DropDownLink("Add account") {
          val group =
              ActionManager.getInstance().getAction("Sourcegraph.Accounts.AddAccount")
                  as ActionGroup
          val dataContext = DataManager.getInstance().getDataContext(it)

          JBPopupFactory.getInstance()
              .createActionGroupPopup(
                  null, group, dataContext, JBPopupFactory.ActionSelectionAid.MNEMONICS, false)
        }
  }
}
