#ifndef _IceKey_H
#define _IceKey_H

class IceSubkey;

class IceKey {
  public:
    IceKey (int n);
    ~IceKey ();
    void    set (const unsigned char *key);
    void    encrypt (const unsigned char *plaintext,
            unsigned char *ciphertext, int idx) const;
    void    decrypt (const unsigned char *ciphertext,
            unsigned char *plaintext, int i) const;
    int     keySize () const;
    int     blockSize () const;

    void    encString (const unsigned char *ptext);
    void    encBinary(const unsigned char *data, long size);
    // void    decString(unsigned char *ctext);
  private:
    void    scheduleBuild (unsigned short *k, int n,
              const int *keyrot);
    int   _size;
    int   _rounds;
    IceSubkey *_keysched;
};

#endif