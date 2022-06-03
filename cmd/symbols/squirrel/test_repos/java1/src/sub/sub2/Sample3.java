package sub.subsub;

//     vvv src/sub path
//         vv C1 ref
import sub.C1;

class C4 extends C1 {
    void m() {
        //      vv C1 ref
        //                   vv f1 ref
        int y = C1.f1 + this.f1;
    }

    //          vv m5 def
    static void m5(Integer i) { }
}
