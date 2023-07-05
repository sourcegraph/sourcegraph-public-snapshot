package com.sourcegraph.cody.autocomplete.prompt_library;

import com.intellij.openapi.Disposable;
import java.util.ArrayList;
import java.util.List;

// TODO: implement the rest of this logic.
public class History implements Disposable {
  List<Disposable> subscriptions = new ArrayList<>();

  @Override
  public void dispose() {}
}
