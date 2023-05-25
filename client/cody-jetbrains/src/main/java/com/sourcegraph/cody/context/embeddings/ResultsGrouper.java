package com.sourcegraph.cody.context.embeddings;

import com.sourcegraph.cody.context.ContextFile;
import org.jetbrains.annotations.NotNull;

import java.util.*;
import java.util.stream.Collectors;

public class ResultsGrouper {
    public static @NotNull List<GroupedResults> groupResultsByFile(@NotNull List<EmbeddingsSearchResult> results) {
        List<ContextFile> originalFileOrder = new ArrayList<>();
        for (EmbeddingsSearchResult result : results) {
            boolean found = false;
            for (ContextFile ogFile : originalFileOrder) {
                if (ogFile.getFileName().equals(result.getFileName())) {
                    found = true;
                    break;
                }
            }
            if (!found) {
                originalFileOrder.add(new ContextFile(result.getFileName(), result.getRepoName(), result.getRevision()));
            }
        }

        Map<String, List<EmbeddingsSearchResult>> resultsGroupedByFile = new HashMap<>();
        for (EmbeddingsSearchResult result : results) {
            List<EmbeddingsSearchResult> groupedResults = resultsGroupedByFile.get(result.getFileName());
            if (groupedResults == null) {
                groupedResults = new ArrayList<>();
                groupedResults.add(result);
                resultsGroupedByFile.put(result.getFileName(), groupedResults);
            } else {
                groupedResults.add(result);
            }
        }

        return originalFileOrder.stream()
            .map(file -> new GroupedResults(file, mergeConsecutiveResults(resultsGroupedByFile.get(file.getFileName()))))
            .collect(Collectors.toList());
    }

    private static @NotNull List<String> mergeConsecutiveResults(@NotNull List<EmbeddingsSearchResult> results) {
        List<EmbeddingsSearchResult> sortedResults = results.stream()
            .sorted(Comparator.comparingInt(EmbeddingsSearchResult::getStartLine))
            .collect(Collectors.toList());

        List<String> mergedResults = new ArrayList<>();
        mergedResults.add(sortedResults.get(0).getContent());

        for (int i = 1; i < sortedResults.size(); i++) {
            EmbeddingsSearchResult result = sortedResults.get(i);
            EmbeddingsSearchResult previousResult = sortedResults.get(i - 1);

            if (result.getStartLine() == previousResult.getEndLine()) {
                mergedResults.set(mergedResults.size() - 1, mergedResults.get(mergedResults.size() - 1) + result.getContent());
            } else {
                mergedResults.add(result.getContent());
            }
        }

        return mergedResults;
    }
}

