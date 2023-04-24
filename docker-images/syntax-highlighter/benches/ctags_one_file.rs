use criterion::{black_box, criterion_group, criterion_main, Criterion};

fn execute_tags(name: &str, contents: &str) -> String {
    let contents = contents.trim();
    scip_syntax::ctags::helper_execute_one_file(name, contents).unwrap()
}

fn criterion_benchmark(c: &mut Criterion) {
    let contents = include_str!("../bench_data/event.rs");
    c.bench_function("event.rs", |b| {
        b.iter(|| execute_tags(black_box("event.rs"), contents))
    });

    let contents = include_str!("../bench_data/big.cpp");
    c.bench_function("big.cpp", |b| {
        b.iter(|| execute_tags(black_box("big.cpp"), contents))
    });
}

criterion_group!(benches, criterion_benchmark);
criterion_main!(benches);
