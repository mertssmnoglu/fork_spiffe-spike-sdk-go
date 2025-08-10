//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package crypto

import "fmt"

const letters = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// RandomString generates a cryptographically-unique secure random string.
func RandomString(n int) (string, error) {
	bytes := make([]byte, n)

	if _, err := reader(bytes); err != nil {
		return "", err
	}

	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}

	return string(bytes), nil
}

// Token generates a cryptographically-unique secure random string.
func Token() string {
	id, err := RandomString(26)
	if err != nil {
		id = fmt.Sprintf("CRYPTO-ERR: %s", err.Error())
	}
	return "spike." + id
}

// ID generates a cryptographically-unique secure random string.
func ID() string {
	id, err := RandomString(8)
	if err != nil {
		id = fmt.Sprintf("CRYPTO-ERR: %s", err.Error())
	}
	return id
}
