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
        static void Main(int p1)
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
            l1 = f2 + l2 + l3;
        }

    }
}
