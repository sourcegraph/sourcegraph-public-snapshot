package com.sourcegraph.cody.context.keyword;

import java.util.List;

public class Term {
    public String stem;
    public List<String> originals;
    public String prefix;
    public int count;

    public Term(String stem, List<String> originals, String prefix, int count) {
        this.stem = stem;
        this.originals = originals;
        this.prefix = prefix;
        this.count = count;
    }
}
