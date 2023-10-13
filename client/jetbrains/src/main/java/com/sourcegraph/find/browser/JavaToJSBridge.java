package com.sourcegraph.find.browser;

import com.google.gson.JsonElement;
import com.google.gson.JsonObject;
import com.google.gson.JsonParser;
import com.google.gson.JsonSyntaxException;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.ui.jcef.JBCefBrowserBase;
import com.intellij.ui.jcef.JBCefJSQuery;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.locks.Lock;
import java.util.concurrent.locks.ReentrantLock;
import java.util.function.Function;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class JavaToJSBridge {
  private final JBCefBrowserBase browser;
  private final JBCefJSQuery query;
  private final Lock lock;
  private Function<String, JBCefJSQuery.Response> handler = null;

  public JavaToJSBridge(JBCefBrowserBase browser) {
    this.browser = browser;
    this.query = JBCefJSQuery.create(browser);
    this.lock = new ReentrantLock();
  }

  public void callJS(@NotNull String action, @Nullable JsonObject arguments) {
    this.callJS(action, arguments, null);
  }

  /**
   * @param result This is the way to get the result back to the caller from the thread started in
   *     this method.
   */
  public void callJS(
      @NotNull String action,
      @Nullable JsonObject arguments,
      @Nullable CompletableFuture<JsonObject> result) {
    // A separate thread is needed because the response handling uses the main thread,
    // so if we did the JS call in the main thread and then waited, the response handler
    // would never be called.
    new Thread(
            () -> {
              Logger logger = Logger.getInstance(this.getClass());
              // Reason for the locking:
              // JBCefJSQuery objects MUST be created before the browser is loaded, otherwise an
              // error is thrown.
              // As there is only one JBCefJSQuery object, and we need to wait for the result of the
              // last execution,
              // we can only run one query at a time.
              // If this ever becomes a bottleneck, we can create a pool of JBCefJSQuery objects and
              // a counting semaphore.
              lock.lock();

              // This future is needed to communicate between this thread and the response handler
              // thread.
              CompletableFuture<Void> handlerCompletedFuture = new CompletableFuture<>();

              String js =
                  "window.callJS('"
                      + action
                      + "', '"
                      + (arguments != null ? arguments.toString() : "null")
                      + "', (result) => {"
                      + "    "
                      + query.inject("result")
                      + "});";

              handler =
                  responseAsString -> {
                    query.removeHandler(handler);
                    handler = null;
                    try {
                      JsonElement jsonElement = JsonParser.parseString(responseAsString);
                      if (result != null) {
                        result.complete(
                            jsonElement.isJsonObject() ? jsonElement.getAsJsonObject() : null);
                      }
                    } catch (JsonSyntaxException e) {
                      logger.warn("Invalid JSON: " + responseAsString);
                      logger.warn(e);
                      if (result != null) {
                        result.complete(null);
                      }
                    } finally {
                      handlerCompletedFuture.complete(null);
                    }
                    return null;
                  };
              query.addHandler(handler);
              browser.getCefBrowser().executeJavaScript(js, browser.getCefBrowser().getURL(), 0);

              try {
                handlerCompletedFuture.get();
              } catch (InterruptedException | ExecutionException e) {
                logger.warn("Some problem occurred with the JS response thread.");
                logger.warn(e);
              } finally {
                // It's only allowed to unlock the lock in the thread where it was locked.
                // This is why `handlerCompletedFuture` is needed in the first place.
                // Otherwise, the handler could simply unlock the lock, but that's not allowed in
                // Java.
                lock.unlock();
              }
            })
        .start();
  }
}
