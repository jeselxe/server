package src

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"io"
	"log"
	"math/rand"
	"project/client/src/errorchecker"

	"golang.org/x/crypto/scrypt"
)

// Check checks error and logs if is not nil
func Check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

// Compress función para comprimir
func Compress(data []byte) []byte {
	var b bytes.Buffer      // b contendrá los datos comprimidos (tamaño variable)
	w := zlib.NewWriter(&b) // escritor que comprime sobre b
	w.Write(data)           // escribimos los datos
	w.Close()               // cerramos el escritor (buffering)
	return b.Bytes()        // devolvemos los datos comprimidos
}

// Decompress función para descomprimir
func Decompress(data []byte) []byte {
	var b bytes.Buffer // b contendrá los datos descomprimidos

	r, err := zlib.NewReader(bytes.NewReader(data)) // lector descomprime al leer

	Check(err)       // comprobamos el error
	io.Copy(&b, r)   // copiamos del descompresor (r) al buffer (b)
	r.Close()        // cerramos el lector (buffering)
	return b.Bytes() // devolvemos los datos descomprimidos
}

// Encode64 función para codificar de []bytes a string (Base64)
func Encode64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data) // sólo utiliza caracteres "imprimibles"
}

// Decode64 función para decodificar de string a []bytes (Base64)
func Decode64(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s) // recupera el formato original
	Check(err)                                   // comprobamos el error
	return b                                     // devolvemos los datos originales
}

// ScryptHash asd
func ScryptHash(word, salt []byte) ([]byte, error) {
	return scrypt.Key(word, salt, 16384, 8, 1, 32)
}

// HashWithRandomSalt asd
func HashWithRandomSalt(pass []byte) ([]byte, []byte) {
	salt := make([]byte, 16)

	rand.Read(salt)
	dk, err := ScryptHash(pass, salt)

	if err != nil {
		log.Println("ERROR SCRYPT", err)
	}

	return dk, salt
}

// EncryptAES función para cifrar (con AES en este caso)
func EncryptAES(data, key []byte) (out []byte) {
	out = make([]byte, len(data)+16)         // reservamos espacio para el IV al principio
	rand.Read(out[:16])                      // generamos el IV
	blk, err := aes.NewCipher(key)           // cifrador en bloque (AES), usa key
	errorchecker.Check("ERROR encrypt", err) // comprobamos el error
	ctr := cipher.NewCTR(blk, out[:16])      // cifrador en flujo: modo CTR, usa IV
	ctr.XORKeyStream(out[16:], data)         // ciframos los datos
	return
}

// DecryptAES función para descifrar
func DecryptAES(data, key []byte) (out []byte) {
	out = make([]byte, len(data)-16)         // la salida no va a tener el IV
	blk, err := aes.NewCipher(key)           // cifrador en bloque (AES), usa key
	errorchecker.Check("ERROR decrypt", err) // comprobamos el error
	ctr := cipher.NewCTR(blk, data[:16])     // cifrador en flujo: modo CTR, usa IV
	ctr.XORKeyStream(out, data[16:])         // desciframos (doble cifrado) los datos
	return
}
