package sub;

//    vv C1 def
class C1 {

    //         vv f1 def
    static int f1 = 5;

    //     vv constructor.p1 def
    C1(int p1) {
        //   vv constructor.p1 ref
        f1 = p1;

        var c2 = new C2();

        //      vv f2 ref
        f1 = c2.f2;
    }

    //   vv m1 def
    //          vv p1 def
    //                     vv p2 def
    void m1(int p1, int ...p2) {

        //  vv l1 def
        int l1;

        //  vv l2 def
        //      vv l3 def
        int l2, l3 = 0;

        //   vv p1 ref
        //        vv p2 ref
        //             vv l1 ref
        //                  vv l2 ref
        //                       vv l3 ref
        //                            vv f1 ref
        l1 = p1 + p2 + l1 + l2 + l3 + f1;

        //   vv m1 ref
        //           vv m2 ref
        //                  vv C1 ref
        //                     vv f1 ref
        //                          vv C2 ref
        //                             vv f2 ref
        //                                        vv f2 ref
        l1 = m1(1) + m2() + C1.f1 + C2.f2 + C1.C2.f2;

        ArrayList<Integer> numbers = new ArrayList<Integer>();
        numbers.add(5);
        numbers.add(9);

        //              v m1.n def
        //                                      v m1.n ref
        numbers.forEach(n -> System.out.println(n));

        //                  vv m5 ref
        numbers.forEach(C4::m5);

        //          v l2.y def
        //                v l2.y ref
        L2 l2 = (x, y) -> y;

        //                       v e def
        //                                           v e ref
        try { } catch (Exception e) { Exception e2 = e; }

        //       v for.i def
        //              v for.i ref
        for (int i = 0; i < 5; i++) { }

        //       v enhanced_for.i def
        //                      v enhanced_for.i ref
        for (int i : p2) { p1 = i; }

        C2 c2; // < "C2" C2 ref

        // vv C2 ref
        C1.C2 c12; // < "C1" C1 ref

        //          vv f2 ref
        int f1 = c2.f2;

        //   vv C3 ref
        //      vv f3 ref
        //              vv C2.m2 ref
        //                   vv f3 ref
        //                        vvvv C1 ref
        //                             vv f1 ref
        p1 = C3.f3 + C2.m2().f3 + this.f1;
    }

    // vv m2 def
    C3 m2() {
        return new C3();
    }

    void m4(C2 c2) {
        //         vv f2 ref
        int _ = c2.f2;
    }

    //    vv C2 def
    class C2 {
        //         vv f2 def
        static int f2;

        // vv C2.m2 def
        C3 m2() { }
    }
}

interface L2 {
    public int l2(int x, int y);
}
