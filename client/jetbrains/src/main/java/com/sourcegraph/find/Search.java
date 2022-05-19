package com.sourcegraph.find;

public class Search {
    String query;
    boolean caseSensitive;
    String patternType;
    String selectedSearchContextSpec;

    public Search(String query, boolean caseSensitive, String patternType, String selectedSearchContextSpec) {
        this.query = query;
        this.caseSensitive = caseSensitive;
        this.patternType = patternType;
        this.selectedSearchContextSpec = selectedSearchContextSpec;
    }

    public String getQuery() {
        return query;
    }

    public boolean isCaseSensitive() {
        return caseSensitive;
    }

    public String getPatternType() {
        return patternType;
    }

    public String getSelectedSearchContextSpec() {
        return selectedSearchContextSpec;
    }
}
