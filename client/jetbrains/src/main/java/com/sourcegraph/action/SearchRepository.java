package com.sourcegraph.action;

import com.intellij.openapi.actionSystem.AnActionEvent;


public class SearchRepository extends SearchActionBase {
    @Override
    public void actionPerformed(AnActionEvent e) {
        super.actionPerformedMode(e, "search.repository");
    }
}