fn main() {
    // let contents = include_str!(
    //     "/home/tjdevries/sourcegraph/sourcegraph.git/main/internal/database/mocks_temp.go"
    // );
    // let contents = include_str!("../ctags.rs");
    parse_dumb_cpp();
    parse_big_cpp();
}

fn parse_dumb_cpp() {
    let file = "dump.cpp";
    let contents = "int main() { return 0; }";

    println!(
        "{}",
        scip_syntax::ctags::helper_execute_one_file(file, contents).unwrap()
    );
}

fn parse_big_cpp() {
    let contents = include_str!(
        "/home/tjdevries/sourcegraph/sourcegraph.git/main/docker-images/syntax-highlighter/bench_data/big.cpp"
    );
    let contents = contents.trim();

    println!(
        "{}",
        scip_syntax::ctags::helper_execute_one_file("mocks_temp.go", contents).unwrap()
    );
}
