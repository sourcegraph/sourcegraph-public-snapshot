package sub.subsub;

import sub.C1;

class C4 {
    void m() {
        C1 c1; // < "C1" C1 ref
    }

    //          vv m5 def
    static void m5(Integer i) { }
}
