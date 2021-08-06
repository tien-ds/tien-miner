package com.ds;

public class Test{
    static{
        System.loadLibrary("share");
    }

    public static void main(String[] args) {
        Fcrypto f = new Fcrypto();
        try {
            Thread.sleep(2000);
        } catch (Exception e) {
            e.printStackTrace();
            //TODO: handle exception
        }
        
        System.out.println(f.generateKey("69e422d8287f515a63c2461e73db1792f394f61299fdaa74e1749507ad31685b", "0xC1B174F1bc70172c911Df8A01D0e0D98129C7517"));
        System.out.println();
    }
}