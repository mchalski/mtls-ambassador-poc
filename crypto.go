package main

import (
	"bytes"
	"crypto"
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"log"
)

// some crypto utils picked straight from deviceauth

func SerializePubKey(key interface{}) (string, error) {

	switch key.(type) {
	case *rsa.PublicKey, *dsa.PublicKey, *ecdsa.PublicKey:
		break
	default:
		return "", errors.New("unrecognizable public key type")
	}

	asn1, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return "", err
	}

	out := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1,
	})

	if out == nil {
		return "", err
	}

	return string(out), nil
}

func VerifyAuthReqSign(signature string, pubkey interface{}, content []byte) error {
	hash := sha256.New()
	_, err := bytes.NewReader(content).WriteTo(hash)
	if err != nil {
		return err
	}

	decodedSig, err := base64.StdEncoding.DecodeString(string(signature))
	if err != nil {
		return err
	}

	key := pubkey.(*rsa.PublicKey)

	err = rsa.VerifyPKCS1v15(key, crypto.SHA256, hash.Sum(nil), decodedSig)
	if err != nil {
		return err
	}

	return nil
}

func dumpCert(cert *x509.Certificate) {
	log.Printf("subject %s\n", cert.Subject.String())
	log.Printf("issuer %s\n", cert.Issuer.String())
}
