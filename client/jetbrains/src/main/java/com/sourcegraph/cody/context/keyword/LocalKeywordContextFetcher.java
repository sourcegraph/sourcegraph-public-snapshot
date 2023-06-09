package com.sourcegraph.cody.context.keyword;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.intellij.openapi.editor.Document;
import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.fileEditor.FileDocumentManager;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.util.TextRange;
import com.intellij.openapi.vfs.LocalFileSystem;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.openapi.vfs.local.CoreLocalFileSystem;
import com.intellij.openapi.vfs.local.CoreLocalVirtualFile;
import com.intellij.psi.PsiDocumentManager;
import com.intellij.psi.PsiFile;
import com.intellij.psi.PsiManager;
import com.sourcegraph.cody.telemetry.GraphQlLogger;
import edu.stanford.nlp.pipeline.StanfordCoreNLP;
import edu.stanford.nlp.util.StringUtils;
import org.jetbrains.annotations.NotNull;

import java.io.BufferedReader;
import java.io.File;
import java.io.InputStreamReader;
import java.util.*;
import java.util.concurrent.CompletableFuture;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import java.util.stream.Collectors;

public class LocalKeywordContextFetcher {
    private final @NotNull String rgPath;
    private final @NotNull Project project;
    private static final ObjectMapper objectMapper = new ObjectMapper();
    private StanfordCoreNLP pipeline;

    public LocalKeywordContextFetcher(@NotNull String rgPath, @NotNull Project project) {
        this.rgPath = rgPath;
        this.project = project;

        // set the list of annotators to run
        //Properties props = StringUtils.argsToProperties("-props", "spanish");
        props.setProperty("annotators", "tokenize,mwt");
        // build pipeline
        StanfordCoreNLP pipeline = new StanfordCoreNLP(props);
    }

    public CompletableFuture<List<KeywordContextFetcherResult>> getContext(@NotNull String query, int numResults) {
        System.out.println("Fetching keyword matches");
        long startTime = System.currentTimeMillis();

        String rootPath = project.getBasePath();
        if (rootPath == null) {
            return CompletableFuture.completedFuture(Collections.emptyList());
        }

        return fetchKeywordFiles(rootPath, query).thenApply(filenamesWithScores -> {
            int resultsToUse = Math.min(numResults, filenamesWithScores.size());
            List<Map<String, Object>> topFiles = filenamesWithScores.subList(0, resultsToUse);
            List<KeywordContextFetcherResult> messagePairs = new ArrayList<>();

            for (Map<String, Object> filenameWithScore : topFiles) {
                String filename = (String) filenameWithScore.get("filename");
                if (filename != null) {
                    VirtualFile file = LocalFileSystem.getInstance().findFileByIoFile(new File(rootPath, filename));
                    if (file != null) {
                        Document document = FileDocumentManager.getInstance().getDocument(file);
                        if (document != null) {
                            String content = document.getText();
                            KeywordContextFetcherResult result = new KeywordContextFetcherResult(null, null, filename, content);
                            messagePairs.add(result);
                        }
                    }
                }
            }

            long searchDuration = System.currentTimeMillis() - startTime;
            GraphQlLogger.logSearchDuration(project, searchDuration);
            Collections.reverse(messagePairs);
            return messagePairs;
        });
    }

    public CompletableFuture<List<KeywordContextFetcherResult>> getSearchContext(String query, int numResults) {
        System.out.println("Fetching keyword search context...");
        String rootPath = project.getBasePath();

        if (rootPath == null) {
            return CompletableFuture.completedFuture(new ArrayList<>());
        }

        List<Term> terms = userQueryToKeywordQuery(query);
        String stems = terms.stream()
            .map(t -> (t.prefix.length() < 4 ? t.originals.get(0) : t.prefix))
            .collect(Collectors.joining("|"));


        return fetchKeywordFiles(rootPath, query).thenApply(filenamesWithScores -> {
            if (filenamesWithScores.size() > numResults) {
                filenamesWithScores = filenamesWithScores.subList(0, numResults);
            }

            List<KeywordContextFetcherResult> results = new ArrayList<>();

            for (Map<String, Object> filenameWithScore : filenamesWithScores) {
                String filename = (String) filenameWithScore.get("filename");

                // Get file of "filename" in the project. We can't use getBaseDir() because it's deprecated
                VirtualFile virtualFile = LocalFileSystem.getInstance().findFileByIoFile(new File(rootPath + "/" + filename));
                if (virtualFile == null) {
                    continue;
                }
                PsiFile psiFile = PsiManager.getInstance(project).findFile(virtualFile);
                if (psiFile == null) {
                    continue;
                }
                Document document = PsiDocumentManager.getInstance(project).getDocument(psiFile);
                if (document == null) {
                    continue;
                }
                String snippet = document.getText();
                Pattern keywordPattern = Pattern.compile(stems);
                Matcher matcher = keywordPattern.matcher(snippet);

                int keywordIndex = 0;
                if (matcher.find()) {
                    keywordIndex = matcher.start();
                } else {
                    keywordIndex = snippet.indexOf(query);
                }

                int startLine = Math.max(0, document.getLineNumber(keywordIndex) - 2);
                int endLine = startLine + 5;

                String content = document.getText(new TextRange(document.getLineStartOffset(startLine), document.getLineEndOffset(endLine)));

                KeywordContextFetcherResult result = new KeywordContextFetcherResult(null, null, filename, content);
                results.add(result);
            }

            return results;
        });
    }

    /**
     * Exclude files without extensions and hidden files (starts with '.')
     * Limits to use 1 thread
     * Exclude files larger than 1MB (based on search.largeFiles)
     * Note: Ripgrep excludes binary files and respects .gitignore by default
     */
    private static final String[] fileExtRipgrepParams = {
        "-g", "*.*",
        "-g", "!.*",
        "-g", "!*.lock",
        "-g", "!*.snap",
        "--threads", "1",
        "--max-filesize", "1M"
    };

    private CompletableFuture<Map<String, FileStat>> fetchFileStats(List<Term> terms, String rootPath) {
        String regexQuery = "\\b" + regexForTerms(terms);
        List<String> command = new ArrayList<>();
        Collections.addAll(command, this.rgPath, "-i", "--json", regexQuery, "./");

        // Implement fileExtRipgrepParams with proper parameters as per the context
        // and add it in the command as well.

        CompletableFuture<String> output = exec(command, rootPath);
        return output.thenApply(out -> {
            Map<String, FileStat> fileTermCounts = new HashMap<>();
            String[] lines = out.split("\n");

            for (String line : lines) {
                Map<String, Object> data;
                try {
                    data = objectMapper.readValue(line, new TypeReference<Map<String, Object>>() {
                    });
                } catch (Exception e) {
                    throw new RuntimeException(e);
                }

                String type = (String) data.get("type");
                if ("end".equals(type)) {
                    Map<String, Object> pathData = (Map<String, Object>) ((Map<String, Object>) data.get("data")).get("path");
                    String pathText = (String) pathData.get("text");
                    Map<String, Object> statsData = (Map<String, Object>) ((Map<String, Object>) data.get("data")).get("stats");
                    int bytesSearched = (Integer) statsData.get("bytes_searched");
                    FileStat fileStat = new FileStat(bytesSearched);
                    fileTermCounts.put(pathText, fileStat);
                }
            }

            return fileTermCounts;
        });
    }

    // Similarly implement fetchFileMatches method and other methods using exec helper method.

    private static CompletableFuture<String> exec(@NotNull List<String> command, @NotNull String directoryPath) {
        return CompletableFuture.supplyAsync(() -> {
            ProcessBuilder processBuilder = new ProcessBuilder(command);
            processBuilder.directory(new File(directoryPath));
            processBuilder.redirectError(ProcessBuilder.Redirect.INHERIT);

            StringBuilder output = new StringBuilder();
            try {
                Process process = processBuilder.start();
                try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()))) {
                    String line;
                    while ((line = reader.readLine()) != null) {
                        output.append(line).append("\n");
                    }
                }
                int exitCode = process.waitFor();
                if (exitCode != 0) {
                    throw new RuntimeException("Command execution failed with exit code: " + exitCode);
                }
            } catch (Exception e) {
                throw new RuntimeException(e);
            }

            return output.toString();
        });
    }

    // Define the FileStat class used to hold statistics for each file.
    private static class FileStat {
        private int bytesSearched;

        public FileStat(int bytesSearched) {
            this.bytesSearched = bytesSearched;
        }

        public int getBytesSearched() {
            return bytesSearched;
        }
    }

    public CompletableFuture<FetchFileMatchesResult> fetchFileMatches(List<Term> queryTerms, String rootPath) {
        return CompletableFuture.supplyAsync(() -> {
            List<CompletableFuture<Pair<Map<String, Integer>, Integer>>> termFileCountsFutures =
                queryTerms.stream().map(term -> fetchTermFileCounts(term, rootPath)).collect(Collectors.toList());

            List<Pair<Map<String, Integer>, Integer>> termFileCountsArr =
                termFileCountsFutures.stream().map(CompletableFuture::join).collect(Collectors.toList());

            int totalFilesSearched = -1;
            for (Pair<Map<String, Integer>, Integer> pair : termFileCountsArr) {
                if (totalFilesSearched >= 0 && totalFilesSearched != pair.second) {
                    throw new RuntimeException("filesSearched did not match");
                }
                totalFilesSearched = pair.second;
            }

            Map<String, Map<String, Integer>> fileTermCounts = new HashMap<>();
            Map<String, Integer> termTotalFiles = new HashMap<>();
            for (int i = 0; i < queryTerms.size(); i++) {
                Term term = queryTerms.get(i);
                Map<String, Integer> fileCounts = termFileCountsArr.get(i).first;
                termTotalFiles.put(term.stem, fileCounts.keySet().size());

                for (Map.Entry<String, Integer> entry : fileCounts.entrySet()) {
                    String filename = entry.getKey();
                    int count = entry.getValue();
                    if (!fileTermCounts.containsKey(filename)) {
                        fileTermCounts.put(filename, new HashMap<>());
                    }
                    fileTermCounts.get(filename).put(term.stem, count);
                }
            }

            return new FetchFileMatchesResult(totalFilesSearched, fileTermCounts, termTotalFiles);
        });
    }

    private static class FetchFileMatchesResult {
        private final int totalFiles;
        private final Map<String, Map<String, Integer>> fileTermCounts;
        private final Map<String, Integer> termTotalFiles;

        public FetchFileMatchesResult(int totalFiles, Map<String, Map<String, Integer>> fileTermCounts, Map<String, Integer> termTotalFiles) {
            this.totalFiles = totalFiles;
            this.fileTermCounts = fileTermCounts;
            this.termTotalFiles = termTotalFiles;
        }

        public int getTotalFiles() {
            return totalFiles;
        }

        public Map<String, Map<String, Integer>> getFileTermCounts() {
            return fileTermCounts;
        }

        public Map<String, Integer> getTermTotalFiles() {
            return termTotalFiles;
        }
    }

    private CompletableFuture<Pair<Map<String, Integer>, Integer>> fetchTermFileCounts(Term term, String rootPath) {
        return CompletableFuture.supplyAsync(() -> {
            List<String> command = new ArrayList<>();
            // Add your rgPath
            command.add(this.rgPath);
            // Add "-i"
            command.add("-i");
            // Add all fileExtRipgrepParams
            Collections.addAll(command, fileExtRipgrepParams);
            // Add "--count-matches"
            command.add("--count-matches");
            // Add "--stats"
            command.add("--stats");
            // Add "\\b" + regexForTerms(term)
            command.add("\\b" + regexForTerms(List.of(term)));
            // Add "./"
            command.add("./");

            CompletableFuture<String> outFuture = exec(command, rootPath);
            Map<String, Integer> fileCounts = new HashMap<>();
            String[] lines = outFuture.join().split("\n");
            int filesSearched = -1;
            for (String line : lines) {
                String[] terms = line.split(":");
                if (terms.length != 2) {
                    Matcher matcher = Pattern.compile("(\\d+) files searched").matcher(line);
                    if (matcher.matches()) {
                        try {
                            filesSearched = Integer.parseInt(matcher.group(1));
                        } catch (NumberFormatException e) {
                            System.err.println("failed to parse number of files matched from string: " + matcher.group(1));
                        }
                    }
                    continue;
                }
                try {
                    int count = Integer.parseInt(terms[1]);
                    fileCounts.put(terms[0], count);
                } catch (NumberFormatException e) {
                    System.err.println("could not parse count from " + terms[1]);
                }
            }
            return new Pair<>(fileCounts, filesSearched);
        });
    }

    public static class Pair<T, U> {
        public T first;
        public U second;

        public Pair(T first, U second) {
            this.first = first;
            this.second = second;
        }
    }

    public static String regexForTerms(List<Term> terms) {
        StringBuilder inner = new StringBuilder();

        for (Term t : terms) {
            if (t.prefix.length() >= 4) {
                inner.append(escapeRegex(t.prefix));
            } else {
                inner.append(escapeRegex(t.stem));
                for (String s : t.originals) {
                    inner.append("|").append(escapeRegex(s));
                }
            }
            inner.append("|");
        }
        // Remove trailing "|"
        if (inner.length() > 0) {
            inner.setLength(inner.length() - 1);
        }

        return "(?:" + inner + ")";
    }

    private CompletableFuture<List<Map<String, Object>>> fetchKeywordFiles(
        String rootPath,
        String rawQuery
    ) {
        List<Term> query = userQueryToKeywordQuery(rawQuery);

        CompletableFuture<FetchFileMatchesResult> fileMatchesFuture = fetchFileMatches(query, rootPath);
        CompletableFuture<Map<String, FileStat>> fileStatsFuture = fetchFileStats(query, rootPath);

        return fileMatchesFuture.thenCombine(fileStatsFuture, (fileMatches, fileStats) -> {
            Map<String, Map<String, Integer>> fileTermCounts = fileMatches.getFileTermCounts();
            Map<String, Integer> termTotalFiles = fileMatches.getTermTotalFiles();
            int totalFiles = fileMatches.getTotalFiles();

            Map<String, Double> idfDict = idf(termTotalFiles, totalFiles);

            int querySizeBytes = query.stream()
                .flatMapToInt(t -> t.originals.stream().mapToInt(orig -> (orig.length() + 1) * t.count))
                .sum();

            List<String> queryStems = query.stream().map(t -> t.stem).collect(Collectors.toList());

            Map<String, Integer> termCountMap = query.stream()
                .collect(Collectors.toMap(t -> t.stem, t -> t.count));

            List<Double> queryTf = tf(queryStems, termCountMap, querySizeBytes);

            List<Double> queryVec = tfidf(queryStems, queryTf, idfDict);

            List<Map<String, Object>> filenamesWithScores = fileTermCounts.entrySet().stream()
                .map(entry -> {
                    String filename = entry.getKey();
                    Map<String, Integer> fileTermCount = entry.getValue();

                    if (!fileStats.containsKey(filename)) {
                        throw new RuntimeException("filename " + filename + " missing from fileStats");
                    }

                    double fileSize = fileStats.get(filename).getBytesSearched();

                    List<Double> tfVec = tf(queryStems, fileTermCount, fileSize);

                    List<Double> tfidfVec = tfidf(queryStems, tfVec, idfDict);

                    double cosineScore = cosine(tfidfVec, queryVec);

                    IdfLogScoreResult idfLogScoreResult = idfLogScore(queryStems, fileTermCount, idfDict);

                    double score = idfLogScoreResult.score;
                    Map<String, Double> scoreComponents = idfLogScoreResult.scoreComponents;

                    if (fileSize > 10000) {
                        score *= 0.1; // downweight very large files
                    }

                    Map<String, Object> result = new HashMap<>();
                    result.put("filename", filename);
                    result.put("cosineScore", cosineScore);
                    result.put("termCounts", fileTermCount);
                    result.put("tfVec", tfVec);
                    result.put("idfDict", idfDict);
                    result.put("score", score);
                    result.put("scoreComponents", scoreComponents);

                    return result;
                })
                .sorted((map1, map2) -> Double.compare((double) map2.get("score"), (double) map1.get("score")))
                .collect(Collectors.toList());

            return filenamesWithScores;
        });
    }


    private static String longestCommonPrefix(String s, String t) {
        int endIdx = 0;
        for (int i = 0; i < s.length() && i < t.length(); i++) {
            if (s.charAt(i) != t.charAt(i)) {
                break;
            }
            endIdx = i + 1;
        }
        return s.substring(0, endIdx);
    }

    public static List<Term> userQueryToKeywordQuery(String query) {
        List<String> origWords = new ArrayList<>();
        for (String chunk : query.split("\\W+")) {
            if (chunk.trim().length() == 0) {
                continue;
            }
            origWords.addAll(WinkUtils.string.tokenize0(chunk));
        }
        List<String> filteredWords = WinkUtils.tokens.removeWords(origWords);

        Map<String, Term> terms = new HashMap<>();
        for (String word : filteredWords) {
            if (word.length() <= 2) {
                boolean skip = true;
                for (int i = 0; i < word.length(); i++) {
                    if (word.charAt(i) >= 128) {
                        skip = false;
                        break;
                    }
                }
                if (skip) {
                    continue;
                }
            }

            String stem = WinkUtils.string.stem(word);
            if (terms.containsKey(stem)) {
                Term term = terms.get(stem);
                term.originals.add(word);
                term.count++;
            } else {
                List<String> originals = new ArrayList<>();
                originals.add(word);
                String prefix = longestCommonPrefix(word.toLowerCase(), stem);
                terms.put(stem, new Term(stem, originals, prefix, 1));
            }
        }
        return new ArrayList<>(terms.values());
    }

    public static class IdfLogScoreResult {
        double score;
        Map<String, Double> scoreComponents;

        IdfLogScoreResult(double score, Map<String, Double> scoreComponents) {
            this.score = score;
            this.scoreComponents = scoreComponents;
        }
    }

    public static IdfLogScoreResult idfLogScore(
        List<String> terms,
        Map<String, Integer> termCounts,
        Map<String, Double> idfDict) {
        double score = 0;
        Map<String, Double> scoreComponents = new HashMap<>();
        for (String term : terms) {
            int ct = termCounts.getOrDefault(term, 0);
            double logScore = ct == 0 ? 0 : Math.log10(ct) + 1;
            double idfLogScore = (idfDict.getOrDefault(term, 1.0)) * logScore;
            score += idfLogScore;
            scoreComponents.put(term, idfLogScore);
        }
        return new IdfLogScoreResult(score, scoreComponents);
    }

    public static double cosine(List<Double> v1, List<Double> v2) {
        if (v1.size() != v2.size()) {
            throw new RuntimeException("v1.size() !== v2.size() " + v1.size() + " !== " + v2.size());
        }
        double dotProd = 0;
        double v1SqMag = 0;
        double v2SqMag = 0;
        for (int i = 0; i < v1.size(); i++) {
            dotProd += v1.get(i) * v2.get(i);
            v1SqMag += v1.get(i) * v1.get(i);
            v2SqMag += v2.get(i) * v2.get(i);
        }
        return dotProd / (Math.sqrt(v1SqMag) * Math.sqrt(v2SqMag));
    }

    public static List<Double> tfidf(List<String> terms, List<Double> tf, Map<String, Double> idf) {
        if (terms.size() != tf.size()) {
            throw new RuntimeException("terms.size() !== tf.size() " + terms.size() + " !== " + tf.size());
        }
        List<Double> tfidf = new ArrayList<>(tf);
        for (int i = 0; i < tfidf.size(); i++) {
            if (idf.get(terms.get(i)) == null) {
                throw new RuntimeException("term " + terms.get(i) + " did not exist in idf dict");
            }
            tfidf.set(i, tfidf.get(i) * idf.get(terms.get(i)));
        }
        return tfidf;
    }

    public static List<Double> tf(List<String> terms, Map<String, Integer> termCounts, double fileSize) {
        List<Double> result = new ArrayList<>();
        for (String term : terms) {
            result.add((double) termCounts.getOrDefault(term, 0) / fileSize);
        }
        return result;
    }

    public static Map<String, Double> idf(Map<String, Integer> termTotalFiles, int totalFiles) {
        double logTotal = Math.log(totalFiles);
        Map<String, Double> result = new HashMap<>();
        for (Map.Entry<String, Integer> entry : termTotalFiles.entrySet()) {
            result.put(entry.getKey(), logTotal - Math.log(entry.getValue()));
        }
        return result;
    }

    public static String escapeRegex(String s) {
        return s.replaceAll("([$()*+./\\[\\]?\\\\^{|}-])", "\\\\$1");
    }
}
