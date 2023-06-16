package com.sourcegraph.cody.completions;

import com.intellij.injected.editor.EditorWindow;
import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.editor.EditorCustomElementRenderer;
import com.intellij.openapi.editor.Inlay;
import com.intellij.openapi.editor.InlayModel;
import com.intellij.openapi.editor.ex.EditorEx;
import com.intellij.openapi.editor.impl.ImaginaryEditor;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.util.Disposer;
import com.intellij.openapi.util.Key;
import com.intellij.util.concurrency.annotations.RequiresEdt;
import com.sourcegraph.cody.CodyCompatibility;
import com.sourcegraph.cody.api.CompletionsService;
import com.sourcegraph.cody.completions.prompt_library.*;
import com.sourcegraph.cody.vscode.*;
import com.sourcegraph.config.ConfigUtil;
import com.sourcegraph.config.NotificationActivity;
import com.sourcegraph.config.SettingsComponent;
import java.util.Optional;
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicReference;
import org.apache.commons.lang.StringUtils;
import org.jetbrains.annotations.NotNull;

/** Responsible for triggering and clearing inline code completions. */
public class CodyCompletionsManager {
  private static final Key<Boolean> KEY_EDITOR_SUPPORTED = Key.create("cody.editorSupported");
  private final ScheduledExecutorService scheduler = Executors.newSingleThreadScheduledExecutor();
  // TODO: figure out how to avoid the ugly nested `Future<CompletableFuture<T>>` type.
  private final AtomicReference<Optional<Future<CompletableFuture<Void>>>> currentJob =
      new AtomicReference<>(Optional.empty());

  public static @NotNull CodyCompletionsManager getInstance() {
    return ApplicationManager.getApplication().getService(CodyCompletionsManager.class);
  }

  @RequiresEdt
  public void clearCompletions(Editor editor) {
    cancelCurrentJob();
    for (Inlay<?> inlay :
        editor.getInlayModel().getInlineElementsInRange(0, editor.getDocument().getTextLength())) {
      if (!(inlay.getRenderer() instanceof CodyCompletionElementRenderer)) {
        continue;
      }
      Disposer.dispose(inlay);
    }
  }

  @RequiresEdt
  public boolean isEnabledForEditor(Editor editor) {
    return ConfigUtil.areCodyCompletionsEnabled()
        && editor != null
        && isProjectAvailable(editor.getProject())
        && isEditorSupported(editor);
  }

  public void triggerCompletion(Editor editor, int offset) {
    if (!ConfigUtil.areCodyCompletionsEnabled()) {
      return;
    }
    CancellationToken token = new CancellationToken();
    SourcegraphNodeCompletionsClient client =
        new SourcegraphNodeCompletionsClient(completionsService(editor), token);
    CodyCompletionItemProvider provider =
        new CodyCompletionItemProvider(
            new WebviewErrorMessenger(),
            client,
            new CompletionsDocumentProvider(),
            new History(),
            2048,
            4,
            200,
            0.6,
            0.1);
    TextDocument textDocument = new IntelliJTextDocument(editor);
    CompletionDocumentContext documentCompletionContext = textDocument.getCompletionContext(offset);
    if (documentCompletionContext.isCompletionTriggerValid()) {
      Callable<CompletableFuture<Void>> callable =
          () ->
              triggerCompletionAsync(
                  editor, offset, token, provider, textDocument, documentCompletionContext);
      // debouncing the completion trigger
      cancelCurrentJob();
      this.currentJob.set(
          Optional.of(this.scheduler.schedule(callable, 20, TimeUnit.MILLISECONDS)));
    }
  }

  public static InlineCompletionItem postProcessInlineCompletionBasedOnDocumentContext(
      InlineCompletionItem resultItem, CompletionDocumentContext documentCompletionContext) {
    String sameLineSuffix = documentCompletionContext.getSameLineSuffix();
    if (resultItem.insertText.endsWith(sameLineSuffix)) {
      // if the completion already has the same line suffix, we strip it
      String newInsertText = StringUtils.stripEnd(resultItem.insertText, sameLineSuffix);
      // adjusting the range to account for the shorter completion
      Range newRange =
          resultItem.range.withEnd(
              resultItem.range.end.withCharacter(
                  resultItem.range.end.character - sameLineSuffix.length()));
      return resultItem.withRange(newRange).withInsertText(newInsertText);
    } else if (resultItem.insertText.contains(sameLineSuffix)) {
      // if the completion already contains the same line suffix
      // but it doesn't strictly end with it
      // we cut the end of the completion starting with the suffix
      int index = resultItem.insertText.lastIndexOf(sameLineSuffix);
      String newInsertText = resultItem.insertText.substring(0, index);
      // adjusting the range to account for the shorter completion
      int rangeDiff = resultItem.insertText.length() - newInsertText.length();
      Range newRange =
          resultItem.range.withEnd(
              resultItem.range.end.withCharacter(resultItem.range.end.character - rangeDiff));
      return resultItem.withRange(newRange).withInsertText(newInsertText);
    } else return resultItem;
  }

  private CompletableFuture<Void> triggerCompletionAsync(
      Editor editor,
      int offset,
      CancellationToken token,
      CodyCompletionItemProvider provider,
      TextDocument textDocument,
      CompletionDocumentContext documentCompletionContext) {
    return provider
        .provideInlineCompletions(
            textDocument,
            textDocument.positionAt(offset),
            new InlineCompletionContext(InlineCompletionTriggerKind.Automatic, null),
            token)
        .thenAccept(
            result -> {
              if (Thread.interrupted()) {
                return;
              }
              if (result.items.isEmpty()) {
                return;
              }
              InlayModel inlayModel = editor.getInlayModel();
              // TODO: smarter logic around selecting the best completion item.
              Optional<InlineCompletionItem> maybeItem =
                  result.items.stream()
                      .map(CodyCompletionsManager::removeUndesiredCharacters)
                      .map(
                          resultItem ->
                              postProcessInlineCompletionBasedOnDocumentContext(
                                  resultItem, documentCompletionContext))
                      .filter(resultItem -> !resultItem.insertText.isEmpty())
                      .findFirst();
              if (maybeItem.isEmpty()) {
                return;
              }
              InlineCompletionItem item = maybeItem.get();
              try {
                EditorCustomElementRenderer renderer =
                    new CodyCompletionElementRenderer(item.insertText, editor);
                ApplicationManager.getApplication()
                    .invokeLater(
                        () -> {
                          this.clearCompletions(editor);
                          inlayModel.addInlineElement(offset, true, renderer);
                        });
              } catch (Exception e) {
                // TODO: do something smarter with unexpected errors.
                e.printStackTrace();
              }
            });
  }

  public static InlineCompletionItem removeUndesiredCharacters(InlineCompletionItem item) {
    // no zero-width spaces or line separator chars, pls
    String newInsertText = item.insertText.replaceAll("[\u200b\u2028]", "");
    int rangeDiff = item.insertText.length() - newInsertText.length();
    Range newRange =
        item.range.withEnd(item.range.end.withCharacter(item.range.end.character - rangeDiff));
    return item.withRange(newRange).withInsertText(newInsertText);
  }

  private boolean isProjectAvailable(Project project) {
    return project != null && !project.isDisposed();
  }

  private boolean isEditorSupported(Editor editor) {
    if (editor.isDisposed()) {
      return false;
    }

    Boolean fromCache = KEY_EDITOR_SUPPORTED.get(editor);
    if (fromCache != null) {
      return fromCache;
    }

    boolean isSupported =
        isEditorInstanceSupported(editor)
            && CodyCompatibility.isSupportedProject(editor.getProject());
    KEY_EDITOR_SUPPORTED.set(editor, isSupported);
    return isSupported;
  }

  public static boolean isEditorInstanceSupported(Editor editor) {
    return !editor.isViewer()
        && !editor.isOneLineMode()
        && !(editor instanceof EditorWindow)
        && !(editor instanceof ImaginaryEditor)
        && (!(editor instanceof EditorEx) || !((EditorEx) editor).isEmbeddedIntoDialogWrapper());
  }

  private CompletionsService completionsService(Editor editor) {
    Project project = editor.getProject();
    boolean isEnterprise =
        ConfigUtil.getInstanceType(project).equals(SettingsComponent.InstanceType.ENTERPRISE);
    String srcEndpoint = System.getenv("SRC_ENDPOINT");
    String instanceUrl =
        srcEndpoint != null
            ? srcEndpoint
            : isEnterprise ? ConfigUtil.getEnterpriseUrl(project) : "https://sourcegraph.com/";
    String accessToken =
        isEnterprise
            ? ConfigUtil.getEnterpriseAccessToken(project)
            : ConfigUtil.getDotComAccessToken(project);
    if (!instanceUrl.endsWith("/")) {
      instanceUrl = instanceUrl + "/";
    }
    if (accessToken == null) {
      accessToken = System.getenv("SRC_ACCESS_TOKEN");
    }
    if (StringUtils.isEmpty(accessToken)) {
      if (!ConfigUtil.isAccessTokenNotificationDismissed())
        NotificationActivity.notifyAboutSourcegraphAccessToken(Optional.of(instanceUrl));
      return null;
    }
    return new CompletionsService(instanceUrl, accessToken);
  }

  private void cancelCurrentJob() {
    // TODO: change this implementation when we avoid nested `Future<CompletableFuture<T>>`
    this.currentJob
        .get()
        .ifPresent(
            job -> {
              if (job.isDone()) {
                try {
                  job.get().cancel(true);
                } catch (ExecutionException
                    | InterruptedException
                    | CancellationException ignored) {
                }
              } else {
                // Cancelling the toplevel `Future<>` appears to cancel the nested
                // `CompletableFuture<>`.
                // Feel free to reimplement this entire method if it's causing problems because this
                // logic is not bulletproof.
                job.cancel(true);
              }
            });
  }
}
