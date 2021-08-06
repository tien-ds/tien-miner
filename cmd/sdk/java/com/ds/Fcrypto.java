package com.ds;

/**
 * Fcrypto
 */
public class Fcrypto {

    public native String generateKey(String ownerKey, String userAddr);

    public native String recoverOwner(String pkey);

    public native String getUserAddr(String pkey);

    public native String deCryptFKey(String md5, String token);

    public native String verifySigned(String signed, String raw);

}
