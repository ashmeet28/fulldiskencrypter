# fulldiskencrypter

Full Disk Encrypter

## Usage

```
mkdir /mnt/fulldiskencrypter && mount -t tmpfs -o size=1G tmpfs /mnt/fulldiskencrypter
cp fulldiskencrypter /mnt/fulldiskencrypter/
openssl kdf -keylen 32 -binary -kdfopt digest:SHA256 -kdfopt pass:password -kdfopt salt:salt -kdfopt iter:1000000000 PBKDF2 > /mnt/fulldiskencrypter/key
fulldiskencrypter genrate_bash_file disk_size_in_bytes /dev/disk_block_device /mnt/fulldiskencrypter/
/mnt/fulldiskencrypter/run
umount /mnt/fulldiskencrypter && rmdir /mnt/fulldiskencrypter
```

- Put slash in the end of directory path `/mnt/fulldiskencrypter/`
- Only use absolute paths
