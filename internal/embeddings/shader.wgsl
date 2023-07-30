struct A {
    a: array<u32>,
}

struct B {
    b: array<u32>,
}

struct C {
    result: i32,
}

var<private> gl_GlobalInvocationID_1: vec3<u32>;
@group(0) @binding(0) 
var<storage, read_write> unnamed: A;
@group(0) @binding(1) 
var<storage, read_write> unnamed_1: B;
var<workgroup> reduce: array<i32, 1024>;
@group(0) @binding(2) 
var<storage, read_write> unnamed_2: C;

fn unpacku1_(data: ptr<function, u32>) -> vec4<i32> {
    var a: i32;
    var b: i32;
    var c: i32;
    var d: i32;

    let _e25 = (*data);
    a = bitcast<i32>((_e25 & 255u));
    let _e28 = a;
    let _e29 = a;
    a = (_e28 - ((_e29 >> bitcast<u32>(7)) * 256));
    let _e34 = (*data);
    b = bitcast<i32>(((_e34 >> bitcast<u32>(8)) & 255u));
    let _e39 = b;
    let _e40 = b;
    b = (_e39 - ((_e40 >> bitcast<u32>(7)) * 256));
    let _e45 = (*data);
    c = bitcast<i32>(((_e45 >> bitcast<u32>(16)) & 255u));
    let _e50 = c;
    let _e51 = c;
    c = (_e50 - ((_e51 >> bitcast<u32>(7)) * 256));
    let _e56 = (*data);
    d = bitcast<i32>(((_e56 >> bitcast<u32>(24)) & 255u));
    let _e61 = d;
    let _e62 = d;
    d = (_e61 - ((_e62 >> bitcast<u32>(7)) * 256));
    let _e67 = a;
    let _e68 = b;
    let _e69 = c;
    let _e70 = d;
    return vec4<i32>(_e67, _e68, _e69, _e70);
}

fn main_1() {
    var id: i32;
    var mult: vec4<i32>;
    var param: u32;
    var param_1: u32;
    var acc: i32;
    var i: i32;

    let _e26 = gl_GlobalInvocationID_1;
    id = bitcast<vec3<i32>>(_e26).x;
    let _e29 = id;
    if (_e29 >= bitcast<i32>(arrayLength((&unnamed.a)))) {
        return;
    }
    let _e34 = id;
    let _e37 = unnamed.a[_e34];
    param = _e37;
    let _e38 = unpacku1_((&param));
    let _e39 = id;
    let _e42 = unnamed_1.b[_e39];
    param_1 = _e42;
    let _e43 = unpacku1_((&param_1));
    mult = (_e38 * _e43);
    let _e45 = id;
    let _e47 = mult[0u];
    let _e49 = mult[1u];
    let _e52 = mult[2u];
    let _e55 = mult[3u];
    reduce[_e45] = (((_e47 + _e49) + _e52) + _e55);
    workgroupBarrier();
    let _e58 = id;
    if (_e58 == 0) {
        acc = 0;
        i = 0;
        loop {
            let _e60 = i;
            if (_e60 < bitcast<i32>(arrayLength((&unnamed.a)))) {
                let _e65 = i;
                let _e67 = reduce[_e65];
                let _e68 = acc;
                acc = (_e68 + _e67);
                continue;
            } else {
                break;
            }
            continuing {
                let _e70 = i;
                i = (_e70 + 1);
            }
        }
        let _e72 = acc;
        unnamed_2.result = _e72;
    }
    return;
}

@compute @workgroup_size(1024, 1, 1) 
fn main(@builtin(global_invocation_id) gl_GlobalInvocationID: vec3<u32>) {
    gl_GlobalInvocationID_1 = gl_GlobalInvocationID;
    main_1();
}
