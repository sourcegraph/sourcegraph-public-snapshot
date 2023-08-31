package com.sourcegraph.cody.statusbar

import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.editor.Editor
import com.intellij.openapi.fileEditor.FileEditor
import com.intellij.openapi.fileEditor.FileEditorManager
import com.intellij.openapi.fileEditor.TextEditor
import com.intellij.openapi.project.DumbAwareAction
import com.intellij.openapi.project.Project
import com.intellij.openapi.project.ProjectManager
import com.sourcegraph.cody.autocomplete.CodyAutocompleteManager
import com.sourcegraph.config.ConfigUtil
import java.util.Arrays

class CodyDisableAutocompleteAction : DumbAwareAction() {
    override fun actionPerformed(e: AnActionEvent) {
        ConfigUtil.setCodyAutocompleteEnabled(false)
        Arrays.stream(ProjectManager.getInstance().openProjects)
            .flatMap { project: Project? -> Arrays.stream(FileEditorManager.getInstance(project!!).allEditors) }
            .filter { fileEditor: FileEditor? -> fileEditor is TextEditor }
            .map { fileEditor: FileEditor -> (fileEditor as TextEditor).editor }
            .forEach { editor: Editor? -> CodyAutocompleteManager.getInstance().clearAutocompleteSuggestions(editor!!) }
    }

    override fun update(e: AnActionEvent) {
        super.update(e)
        e.presentation.isEnabledAndVisible = ConfigUtil.isCodyEnabled() && ConfigUtil.isCodyAutocompleteEnabled()
    }
}
