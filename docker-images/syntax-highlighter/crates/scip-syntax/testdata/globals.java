package MyPackage;

public class globals {
    private static int field1;
    protected static int field2;
    public static int field3;
    private int field4;
    protected int field5;
    public int field6;

    private static void method1() {}
    protected static void method2() {}
    public static void method3() {}
    private void method4() {}
    protected void method5() {}
    public void method6() {}

    public static final String COOLEST_STRING = "probably this one";

    public class ClassInAClass {
        boolean classy = true;

        public static enum Enum {
            these,
            should,
            be,
            recognized,
            as,
            terms
        }

        public interface Goated {
            boolean withTheSauce();
        }

        public void myCoolMethod() {
            class WhatIsGoingOn {}
            boolean iThinkThisIsAllowedButWeDontReallyCare = true;
        }
    }
}
