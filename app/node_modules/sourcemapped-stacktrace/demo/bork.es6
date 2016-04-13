class BabelBorker {
    bork() {
        throw new Error("bork from es6");
    }
}

window.babel_bork = () => new BabelBorker().bork()
