// Package experiment holds intentional vulnerable dependencies for CI failure drills.
// Do not merge this package to main.
package experiment

import (
	jwt "github.com/dgrijalva/jwt-go"
)

// KeepVulnerableDependency references a known-vulnerable module so go.mod retains it.
// github.com/dgrijalva/jwt-go@v3.2.0 has CVE-2020-26160 (High).
func KeepVulnerableDependency() string {
	return jwt.SigningMethodHS256.Alg()
}
