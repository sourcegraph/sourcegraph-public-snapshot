struct A {
    a: array<u32>,
}

struct B {
    b: array<u32>,
}

struct C {
    result: array<i32>,
}

var<private> gl_GlobalInvocationID_1: vec3<u32>;
var<private> gl_LocalInvocationID_1: vec3<u32>;
@group(0) @binding(0) 
var<storage> unnamed: A;
@group(0) @binding(1) 
var<storage> unnamed_1: B;
var<workgroup> reduce: array<i32, 1024>;
@group(0) @binding(2) 
var<storage, read_write> unnamed_2: C;
var<private> gl_WorkGroupID_1: vec3<u32>;

fn unpacku1_(data: ptr<function, u32>) -> vec4<i32> {
    var a: i32;
    var b: i32;
    var c: i32;
    var d: i32;

    let _e28 = (*data);
    a = bitcast<i32>((_e28 & 255u));
    let _e31 = a;
    let _e32 = a;
    a = (_e31 - ((_e32 >> bitcast<u32>(7)) * 256));
    let _e37 = (*data);
    b = bitcast<i32>(((_e37 >> bitcast<u32>(8)) & 255u));
    let _e42 = b;
    let _e43 = b;
    b = (_e42 - ((_e43 >> bitcast<u32>(7)) * 256));
    let _e48 = (*data);
    c = bitcast<i32>(((_e48 >> bitcast<u32>(16)) & 255u));
    let _e53 = c;
    let _e54 = c;
    c = (_e53 - ((_e54 >> bitcast<u32>(7)) * 256));
    let _e59 = (*data);
    d = bitcast<i32>(((_e59 >> bitcast<u32>(24)) & 255u));
    let _e64 = d;
    let _e65 = d;
    d = (_e64 - ((_e65 >> bitcast<u32>(7)) * 256));
    let _e70 = a;
    let _e71 = b;
    let _e72 = c;
    let _e73 = d;
    return vec4<i32>(_e70, _e71, _e72, _e73);
}

fn main_1() {
    var global_id: u32;
    var local_id: u32;
    var mult: vec4<i32>;
    var param: u32;
    var param_1: u32;
    var i: u32;

    let _e30 = gl_GlobalInvocationID_1[0u];
    global_id = _e30;
    let _e32 = gl_LocalInvocationID_1[0u];
    local_id = _e32;
    let _e33 = global_id;
    if (_e33 >= bitcast<u32>(bitcast<i32>(arrayLength((&unnamed.a))))) {
        return;
    }
    let _e39 = global_id;
    let _e42 = unnamed.a[_e39];
    param = _e42;
    let _e43 = unpacku1_((&param));
    let _e44 = global_id;
    let _e47 = unnamed_1.b[_e44];
    param_1 = _e47;
    let _e48 = unpacku1_((&param_1));
    mult = (_e43 * _e48);
    let _e50 = local_id;
    let _e52 = mult[0u];
    let _e54 = mult[1u];
    let _e57 = mult[2u];
    let _e60 = mult[3u];
    reduce[_e50] = (((_e52 + _e54) + _e57) + _e60);
    workgroupBarrier();
    i = 512u;
    loop {
        let _e63 = i;
        if (_e63 > 0u) {
            let _e65 = local_id;
            let _e66 = i;
            if (_e65 < _e66) {
                let _e68 = local_id;
                let _e69 = local_id;
                let _e70 = i;
                let _e73 = reduce[(_e69 + _e70)];
                let _e75 = reduce[_e68];
                reduce[_e68] = (_e75 + _e73);
            }
            workgroupBarrier();
            continue;
        } else {
            break;
        }
        continuing {
            let _e78 = i;
            i = (_e78 >> bitcast<u32>(1));
        }
    }
    let _e81 = local_id;
    if (_e81 == 0u) {
        let _e84 = gl_WorkGroupID_1[0u];
        let _e86 = reduce[0];
        unnamed_2.result[_e84] = _e86;
    }
    return;
}

@compute @workgroup_size(1024, 1, 1) 
fn main(@builtin(global_invocation_id) gl_GlobalInvocationID: vec3<u32>, @builtin(local_invocation_id) gl_LocalInvocationID: vec3<u32>, @builtin(workgroup_id) gl_WorkGroupID: vec3<u32>) {
    gl_GlobalInvocationID_1 = gl_GlobalInvocationID;
    gl_LocalInvocationID_1 = gl_LocalInvocationID;
    gl_WorkGroupID_1 = gl_WorkGroupID;
    main_1();
}
