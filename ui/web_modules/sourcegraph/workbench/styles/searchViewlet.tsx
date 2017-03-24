import { colors } from "sourcegraph/components/utils";

import "sourcegraph/workbench/styles/searchViewlet.css";

import { insertGlobal } from "glamor";

insertGlobal(" .search-viewlet .monaco-tree-row", {
	backgroundColor: "#344966 !important",
	color: `${colors.blueGrayL2()} !important`,
});

insertGlobal(" .search-viewlet .monaco-tree-row:hover", {
	backgroundColor: `${colors.blueGray()} !important`,
});

insertGlobal(" .search-viewlet .monaco-tree-row .plain", {
	color: `${colors.blueGrayL2()} !important`,
});
