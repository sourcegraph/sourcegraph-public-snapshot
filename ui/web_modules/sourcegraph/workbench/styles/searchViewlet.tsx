import { colors } from "sourcegraph/components/utils";

import "sourcegraph/workbench/styles/searchViewlet.css";

import { insertGlobal } from "glamor";

insertGlobal(" .search-viewlet .monaco-tree-row", {
	backgroundColor: "#344966 !important",
	color: `white !important`,
});

insertGlobal(" .search-viewlet .monaco-tree-row:hover", {
	backgroundColor: `${colors.blueGrayD1()} !important`,
});

insertGlobal(" .search-viewlet .monaco-tree-row .plain", {
	color: `white !important`,
});
