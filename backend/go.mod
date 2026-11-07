module github.com/rou-gan/WhiteKey-Experimentation-Site/backend

go 1.22

// EXPERIMENT: known-vulnerable dependency for Trivy/Grype failure drills. Do not merge to main.
require github.com/dgrijalva/jwt-go v3.2.0+incompatible
