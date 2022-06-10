// Hello World! program
namespace HelloWorld
{
    //    vvvvv Hello cs.Hello def
    class Hello {

        //         vv cs.f1 def
        static int f1 = 5;

        //     vv cs.f4 def
        //                      vv cs.f6 def
        C2(int f4, params int[] f6) {
            //       vv cs.f4 ref
            int f5 = f4;
            //         vv cs.f6 ref
            int[] f7 = f6;
        }

        //                   vv cs.p1 def
        static void Main(int p1, params int[] p2)
        {
            //  vv cs.l1 def
            int l1;

            //  vv cs.f2 def
            //       vv cs.p1 ref
            int f2 = p1;

            //       vv cs.f2 ref
            //            vv cs.f1 ref
            int f3 = f2 + f1;

            Hello h; // < "Hello" cs.Hello ref

            //  vv cs.l2 def
            //      vv cs.l3 def
            int l2, l3 = 0;

        //  vv cs.l1 ref
        //            vv cs.l2 ref
        //                 vv cs.l3 ref
        //                      vv cs.C2 ref
            l1 = f2 + l2 + l3 + C2.g1;

            //                       v cs.e def
            //                                           v cs.e ref
            try { } catch (Exception e) { Exception e2 = e; }

            //       v cs.for.i def
            //              v cs.for.i ref
            for (int i = 0; i < 5; i++) { }

            //           v cs.enhanced_for.i def
            //                           v cs.enhanced_for.i ref
            foreach (int i in p2) { p1 = i; }

            C2 c2; // < "C2" cs.C2 ref

            //    vv cs.C2 ref
            Hello.C2 c12; // < "Hello" cs.Hello ref

            //          vv cs.C2.g1 ref
            int f1 = c2.g1;
        }
    }

    //    vv cs.C2 def
    class C2 {
        //         vv cs.C2.g1 def
        static int g1;

    }
}
