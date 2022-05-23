//    vv C1 def
class C1 {

    //  vv f1 def
    int f1;

    //     vv constructor.p1 def
    C1(int p1) {
        //   vv constructor.p1 ref
        f1 = p1;
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

        //          v l2.y def
        //                v l2.y ref
        L2 l2 = (x, y) -> y;

        //                       v e def
        //                                           v e ref
        try { } catch (Exception e) { Exception e2 = e; }
    }

    //   vv m2 def
    void m2() { }

    //    vv C2 def
    class C2 {
        //         vv f2 def
        static int f2;
    }
}

interface L2 {
    public int l2(int x, int y);
}
