package utils

import (
	"bytes"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
)

func PGPEncrypt(data []byte, to []*openpgp.Entity) ([]byte, error) {
	output := &bytes.Buffer{}
	input, err := openpgp.Encrypt(output, to, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	if _, err := input.Write(data); err != nil {
		return nil, err
	}

	if err := input.Close(); err != nil {
		return nil, err
	}

	return output.Bytes(), nil
}

func PGPArmor(data []byte) ([]byte, error) {
	output := &bytes.Buffer{}
	input, err := armor.Encode(output, "PGP MESSAGE", nil)
	if err != nil {
		return nil, err
	}

	if _, err := input.Write(data); err != nil {
		return nil, err
	}

	if err := input.Close(); err != nil {
		return nil, err
	}

	return output.Bytes(), nil
}
