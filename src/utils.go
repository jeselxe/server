package src

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"io"
	"log"
	"math/rand"

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
