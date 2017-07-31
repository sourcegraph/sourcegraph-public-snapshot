export function handleSearchInput(e: any): void {
    const query = e.target.value;
    if ((e.key !== "Enter" && e.keyCode !== 13) || !query) {
        return;
    }

    // TODO(john): conditionally read props from URL.
    const repos = window.localStorage.getItem("searchRepoScope") || "active";
    const files = window.localStorage.getItem("searchFileScope") || "";

    let newTab = false;
    if (e.metaKey || e.altKey || e.ctrlKey) {
        newTab = true;
    }
    const href = `/search?q=${encodeURIComponent(query)}&repos=${encodeURIComponent(repos)}${files ? `&files=${encodeURIComponent(files)}` : ""}`;
    newTab ? window.open(`/search?q=${encodeURIComponent(query)}&repos=${encodeURIComponent(repos)}${files ? `&files=${encodeURIComponent(files)}` : ""}`, "_blank") : window.location.href = href;
}