struct Data {
    a: array<i32>,
}

struct Data2_ {
    b: array<i32>,
}

struct Data3_ {
    output: i32,
}

@group(0) @binding(0) 
var<storage> global: Data;
@group(0) @binding(1) 
var<storage> global_1: Data2_;
@group(0) @binding(2) 
var<storage, read_write> global_2: Data3_;
var<workgroup> reduce: array<vec4<i32>, 16>;
var<private> gl_GlobalInvocationID: vec3<u32>;
var<private> gl_NumWorkGroups: vec3<u32>;

fn unpack(data: i32) -> vec4<i32> {
    var data_1: i32;

    data_1 = data;
    let _e8 = data_1;
    let _e11 = data_1;
    let _e17 = data_1;
    let _e23 = data_1;
    return vec4<i32>((_e8 & 255), ((_e11 >> u32(8)) & 255), ((_e17 >> u32(16)) & 255), ((_e23 >> u32(24)) & 255));
}

fn pack(data_2: vec4<i32>) -> i32 {
    var data_3: vec4<i32>;
    var r: i32;
    var g: i32;
    var b: i32;
    var a: i32;

    data_3 = data_2;
    let _e8 = data_3;
    r = _e8.x;
    let _e11 = data_3;
    g = _e11.y;
    let _e14 = data_3;
    b = _e14.z;
    let _e17 = data_3;
    a = _e17.w;
    let _e20 = a;
    let _e24 = b;
    let _e29 = g;
    let _e34 = r;
    return ((((_e20 << u32(24)) | (_e24 << u32(16))) | (_e29 << u32(8))) | _e34);
}

fn main_1() {
    var id: i32;
    var acc: i32;
    var i: i32;

    let _e8 = gl_GlobalInvocationID;
    id = vec3<i32>(_e8).x;
    let _e12 = id;
    let _e14 = id;
    let _e17 = id;
    let _e19 = global.a[_e17];
    let _e20 = unpack(_e19);
    let _e21 = id;
    let _e24 = id;
    let _e26 = global_1.b[_e24];
    let _e27 = unpack(_e26);
    reduce[_e12] = (_e20 * _e27);
    let _e29 = id;
    if (_e29 == 0) {
        {
            acc = 0;
            i = 0;
            loop {
                let _e37 = i;
                let _e38 = gl_NumWorkGroups;
                if !((u32(_e37) < _e38.x)) {
                    break;
                }
                {
                    let _e46 = acc;
                    let _e47 = i;
                    let _e49 = reduce[_e47];
                    let _e51 = i;
                    let _e53 = reduce[_e51];
                    let _e56 = i;
                    let _e58 = reduce[_e56];
                    let _e61 = i;
                    let _e63 = reduce[_e61];
                    acc = (_e46 + (((_e49.x + _e53.y) + _e58.z) + _e63.w));
                }
                continuing {
                    let _e43 = i;
                    i = (_e43 + 1);
                }
            }
            let _e67 = acc;
            global_2.output = _e67;
            return;
        }
    } else {
        return;
    }
}

@compute @workgroup_size(4, 1, 1) 
fn main(@builtin(global_invocation_id) param: vec3<u32>, @builtin(num_workgroups) param_1: vec3<u32>) {
    gl_GlobalInvocationID = param;
    gl_NumWorkGroups = param_1;
    main_1();
    return;
}
