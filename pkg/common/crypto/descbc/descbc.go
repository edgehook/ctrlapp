package descbc

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
)

func Pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)

	return append(ciphertext, padtext...)
}

func Pkcs5UnPadding(data []byte) []byte {
	length := len(data)
	unpadding := int(data[length-1])

	return data[:length-unpadding]
}

/*
* DES CBC PKCS5padding ecrypt.
 */
func Encrypt(data, key []byte) ([]byte, error) {

	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blksz := block.BlockSize()

	//blockmodes do encrypts/ decrypts a number of blocks
	blockMode := cipher.NewCBCEncrypter(block, key[:blksz])

	//padding data.
	data = Pkcs5Padding(data, blksz)
	crypted := make([]byte, len(data))

	//encrypt the padding data
	blockMode.CryptBlocks(crypted, data)

	return crypted, nil
}

/*
* DES CBC PKCS5padding Decrypt.
 */
func Decrypt(crypted, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blksz := block.BlockSize()

	//blockmodes do decrypts a number of blocks
	blockMode := cipher.NewCBCDecrypter(block, key[:blksz])

	//decrypt the padding data
	data := make([]byte, len(crypted))
	blockMode.CryptBlocks(data, crypted)

	data = Pkcs5UnPadding(data)

	return data, nil
}
