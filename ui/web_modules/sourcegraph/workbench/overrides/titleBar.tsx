import { TitlebarPart } from "vs/workbench/browser/parts/titlebar/titlebarPart";

// Stop VSCode from updating the page title, we want to provide something
// custom.
TitlebarPart.prototype.updateTitle = () => { /* */ };
