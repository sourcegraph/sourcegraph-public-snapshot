use camino::Utf8Path;
use criterion::{criterion_group, criterion_main, Criterion};
use scip_syntax::{io::read_index_from_file, scip_strict};

fn parse_symbols(symbols: &[&str]) {
    for symbol in symbols {
        scip::symbol::parse_symbol(symbol).unwrap();
    }
}

fn parse_symbols_v2(symbols: &[&str]) {
    for symbol in symbols {
        scip_strict::Symbol::parse(&symbol).unwrap();
    }
}

fn symbols_from_index(path: &str) -> impl Iterator<Item = String> {
    let index = read_index_from_file(Utf8Path::new(path))
    .unwrap();
    index
        .documents
        .into_iter()
        .flat_map(|document| {
            document
                .occurrences
                .into_iter()
                .map(|occurrence| occurrence.symbol)
        })
}

fn bench_symbol_parsing(c: &mut Criterion) {
    // let all_symbols: Vec<String> = symbols_from_index("~/work/scip-indices/spring-framework-syntactic.scip").collect();
    let all_symbols: Vec<String> = symbols_from_index("/Users/creek/work/scip-indices/chromium-1.scip").collect();
    let symbol_count = all_symbols.len();
    let n = 10_000;
    let symbols: Vec<&str> = all_symbols.iter().step_by(symbol_count / n).map(|s| s.as_str()).collect();
    let mut group = c.benchmark_group("symbol parsing");
    group.bench_function("parse", |b| {
        b.iter(|| parse_symbols(&symbols))
    });
    group.bench_function("parse_v2", |b| {
        b.iter(|| parse_symbols_v2(&symbols))
    });
}

criterion_group!(benches, bench_symbol_parsing);
criterion_main!(benches);
