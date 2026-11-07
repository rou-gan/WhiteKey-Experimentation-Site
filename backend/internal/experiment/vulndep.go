// Package experiment holds JWT dependency wiring used for CI security drills.
// Do not merge this package to main.
package experiment

import (
	jwt "github.com/golang-jwt/jwt/v5"
)

// KeepJWTDependency references the maintained JWT module successor.
// Replaces github.com/dgrijalva/jwt-go (GHSA-w73w-5m7g-f7qc / CVE-2020-26160).
func KeepJWTDependency() string {
	return jwt.SigningMethodHS256.Alg()
}
